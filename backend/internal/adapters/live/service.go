package live

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

var (
	// ErrSessionNotFound is returned when a session ID does not exist in the database.
	ErrSessionNotFound = errors.New("session not found")
	// ErrInvalidSessionID is returned when a session ID is blank or contains path traversal characters.
	ErrInvalidSessionID = errors.New("invalid session id")
)

// LiveEventType represents an SSE event kind.
type LiveEventType string

const (
	LiveEventSessionTree   LiveEventType = "session_tree"
	LiveEventTraceUpdate   LiveEventType = "trace_update"
	LiveEventAnomalyAlert LiveEventType = "anomaly_alert"
)

// LiveEvent represents an event pushed to real-time stream subscribers.
type LiveEvent struct {
	Type LiveEventType `json:"type"`
	Data []byte        `json:"data"`
}

// LiveService coordinates SQLite polling, tree reconstruction, and real-time streaming.
type LiveService struct {
	hook            *SQLiteHook
	buffer          *TraceBuffer
	cfg             domain.ProjectConfig
	subscribers     map[chan LiveEvent]struct{}
	subMu           sync.RWMutex
	stopCh          chan struct{}
	closeOnce       sync.Once
	lastTreeJSON    []byte
	lastTreeMu      sync.Mutex
	seenAnomalies   map[string]bool
	seenAnomaliesMu sync.Mutex
}

// NewLiveService constructs a new LiveService instance.
func NewLiveService(hook *SQLiteHook, cfg domain.ProjectConfig) *LiveService {
	maxBuffer := cfg.Live.MaxTraceBufferMessages
	if maxBuffer <= 0 {
		maxBuffer = 500
	}

	return &LiveService{
		hook:          hook,
		buffer:        NewTraceBuffer(maxBuffer),
		cfg:           cfg,
		subscribers:   make(map[chan LiveEvent]struct{}),
		stopCh:        make(chan struct{}),
		seenAnomalies: make(map[string]bool),
	}
}

// Buffer returns the underlying transient in-memory trace buffer.
func (s *LiveService) Buffer() *TraceBuffer {
	return s.buffer
}

// Hook returns the underlying read-only SQLite hook.
func (s *LiveService) Hook() *SQLiteHook {
	return s.hook
}

// Close gracefully stops the background polling loop and closes the SQLite hook.
func (s *LiveService) Close() error {
	var closeErr error
	s.closeOnce.Do(func() {
		close(s.stopCh)
		if s.hook != nil {
			closeErr = s.hook.Close()
		}
	})
	return closeErr
}

// Subscribe registers a new subscriber channel for receiving live events.
// It returns the channel and an unregister callback to prevent leaks.
func (s *LiveService) Subscribe() (<-chan LiveEvent, func()) {
	ch := make(chan LiveEvent, 32)
	s.subMu.Lock()
	s.subscribers[ch] = struct{}{}
	s.subMu.Unlock()

	cleanup := func() {
		s.subMu.Lock()
		delete(s.subscribers, ch)
		close(ch)
		s.subMu.Unlock()
	}

	return ch, cleanup
}

func (s *LiveService) broadcast(event LiveEvent) {
	s.subMu.RLock()
	defer s.subMu.RUnlock()

	for ch := range s.subscribers {
		select {
		case ch <- event:
		default:
			// Subscriber buffer full; skip to keep broadcaster non-blocking
		}
	}
}

// Start begins the background polling loop on live.poll_interval_ms.
func (s *LiveService) Start(ctx context.Context) {
	intervalMs := s.cfg.Live.PollIntervalMs
	if intervalMs <= 0 {
		intervalMs = 2000
	}
	interval := time.Duration(intervalMs) * time.Millisecond

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stopCh:
				return
			case <-ticker.C:
				s.pollAndBroadcast(ctx)
			}
		}
	}()
}

func (s *LiveService) pollAndBroadcast(ctx context.Context) {
	tree, err := s.GetSessionTree(ctx)
	if err != nil {
		return
	}

	data, err := json.Marshal(tree)
	if err != nil {
		return
	}

	s.lastTreeMu.Lock()
	changed := !bytes.Equal(s.lastTreeJSON, data)
	if changed {
		s.lastTreeJSON = data
	}
	s.lastTreeMu.Unlock()

	s.checkAndBroadcastAnomalies(tree.ActiveAnomalies)

	if changed {
		s.broadcast(LiveEvent{
			Type: LiveEventSessionTree,
			Data: data,
		})
	}
}

func (s *LiveService) checkAndBroadcastAnomalies(anomalies []domain.LiveAnomalyAlert) {
	s.seenAnomaliesMu.Lock()
	defer s.seenAnomaliesMu.Unlock()

	currentKeys := make(map[string]bool, len(anomalies))
	for _, alert := range anomalies {
		key := fmt.Sprintf("%s:%s:%s:%s", alert.SessionID, alert.RuleName, alert.Severity, alert.TriggerEvent)
		currentKeys[key] = true
		if !s.seenAnomalies[key] {
			s.seenAnomalies[key] = true
			alertData, err := json.Marshal(alert)
			if err == nil {
				s.broadcast(LiveEvent{
					Type: LiveEventAnomalyAlert,
					Data: alertData,
				})
			}
		}
	}

	for key := range s.seenAnomalies {
		if !currentKeys[key] {
			delete(s.seenAnomalies, key)
		}
	}
}

// GetSessionTree queries active sessions, runs title ticket inference, and constructs the session tree.
// Zero Git branch calls, zero workspace scans.
func (s *LiveService) GetSessionTree(ctx context.Context) (domain.LiveSessionTree, error) {
	if s.hook == nil {
		return domain.LiveSessionTree{
			RootSessions:        []*domain.LiveSessionNode{},
			LastRefreshedAt:     "",
			TotalActiveSessions: 0,
			ActiveAnomalies:     []domain.LiveAnomalyAlert{},
		}, nil
	}

	activeThresholdSec := int64(s.cfg.Live.ActiveThresholdSeconds)
	if activeThresholdSec <= 0 {
		activeThresholdSec = 300
	}
	nowMs := time.Now().UnixMilli()

	records, err := s.hook.FetchActiveSessions(activeThresholdSec, nowMs)
	if err != nil {
		return domain.LiveSessionTree{}, fmt.Errorf("fetching active sessions: %w", err)
	}

	stallThresholdSec := int64(s.cfg.Live.AnomalyThresholds.StallThresholdSeconds)
	if stallThresholdSec <= 0 {
		stallThresholdSec = 120
	}

	nodes := make([]*domain.LiveSessionNode, 0, len(records))
	for _, rec := range records {
		role := s.cfg.RoleMappings.ResolveRole(rec.Agent)
		modelID := resolveModelName(rec.Model)

		// Infer ticket ID strictly from session title
		var guessedID *string
		var confidence float64
		if rec.Title != "" {
			guessedID, confidence = domain.InferTicketID(rec.Title, s.cfg.Ticket)
		}

		// Calculate execution status
		lastActive := rec.TimeUpdated
		if lastActive <= 0 {
			lastActive = rec.TimeCreated
		}
		status := "completed"
		if lastActive > 0 {
			elapsedSec := (nowMs - lastActive) / 1000
			if elapsedSec < stallThresholdSec {
				status = "running"
			} else if elapsedSec < activeThresholdSec {
				status = "stalled"
			}
		}

		sessionAnomalies := []domain.LiveAnomalyAlert{}
		trace, err := s.GetSessionTrace(ctx, rec.ID)
		if err == nil && len(trace.Anomalies) > 0 {
			sessionAnomalies = trace.Anomalies
		}

		node := &domain.LiveSessionNode{
			ID:               rec.ID,
			ParentID:         rec.ParentID,
			Agent:            rec.Agent,
			Role:             role,
			Model:            modelID,
			Title:            rec.Title,
			TimeCreated:      formatEpochMilli(rec.TimeCreated),
			TimeUpdated:      formatEpochMilli(rec.TimeUpdated),
			Status:           status,
			GuessedTicketID:  guessedID,
			TicketConfidence: confidence,
			Anomalies:        sessionAnomalies,
			SubSessions:      []*domain.LiveSessionNode{},
		}
		nodes = append(nodes, node)
	}

	roots := domain.BuildSessionTree(nodes)
	totalActive := countTotalNodes(roots)
	refreshedAt := formatEpochMilli(nowMs)

	allActiveAnomalies := make([]domain.LiveAnomalyAlert, 0)
	for _, n := range nodes {
		allActiveAnomalies = append(allActiveAnomalies, n.Anomalies...)
	}
	domain.SortLiveAnomalies(allActiveAnomalies)

	return domain.LiveSessionTree{
		RootSessions:        roots,
		LastRefreshedAt:     refreshedAt,
		TotalActiveSessions: totalActive,
		ActiveAnomalies:     allActiveAnomalies,
	}, nil
}

// GetSessionTrace extracts messages, reasoning, and tool executions for a given session.
func (s *LiveService) GetSessionTrace(ctx context.Context, sessionID string) (domain.LiveTraceView, error) {
	if err := validateSessionID(sessionID); err != nil {
		return domain.LiveTraceView{}, fmt.Errorf("%w: %s", ErrInvalidSessionID, err.Error())
	}

	if s.hook == nil {
		return domain.LiveTraceView{}, ErrSessionNotFound
	}

	exists, err := s.hook.SessionExists(sessionID)
	if err != nil {
		return domain.LiveTraceView{}, fmt.Errorf("verifying session existence: %w", err)
	}
	if !exists {
		return domain.LiveTraceView{}, fmt.Errorf("%w: %s", ErrSessionNotFound, sessionID)
	}

	rawMsgs, rawParts, err := s.hook.FetchSessionTrace(sessionID)
	if err != nil {
		return domain.LiveTraceView{}, fmt.Errorf("fetching session trace for %s: %w", sessionID, err)
	}

	// Group parts by message ID
	partsByMsg := make(map[string][]RawPartRecord)
	for _, p := range rawParts {
		partsByMsg[p.MessageID] = append(partsByMsg[p.MessageID], p)
	}

	messages := make([]domain.LiveMessageFact, 0, len(rawMsgs))
	var totalTools, failedTools int
	var recentErrors []domain.LivePartFact
	consecutiveErrors := 0

	for _, rm := range rawMsgs {
		role := "assistant"
		var contentStr string
		var reasoningStr *string
		var inputTokens, outputTokens, cacheReadTokens, reasoningTokens int64

		if strings.TrimSpace(rm.Data) != "" {
			var msgData struct {
				Role             *string         `json:"role"`
				Content          json.RawMessage `json:"content"`
				Reasoning        *string         `json:"reasoning"`
				ReasoningContent *string         `json:"reasoning_content"`
				Tokens           *struct {
					Input     *int64 `json:"input"`
					Output    *int64 `json:"output"`
					Cache     *struct {
						Read *int64 `json:"read"`
					} `json:"cache"`
					Reasoning *int64 `json:"reasoning"`
				} `json:"tokens"`
			}
			if err := json.Unmarshal([]byte(rm.Data), &msgData); err == nil {
				if msgData.Role != nil && strings.TrimSpace(*msgData.Role) != "" {
					role = strings.TrimSpace(*msgData.Role)
				}
				contentStr = extractMessageContent(msgData.Content)
				if msgData.ReasoningContent != nil && strings.TrimSpace(*msgData.ReasoningContent) != "" {
					reasoningStr = msgData.ReasoningContent
				} else if msgData.Reasoning != nil && strings.TrimSpace(*msgData.Reasoning) != "" {
					reasoningStr = msgData.Reasoning
				}
				if msgData.Tokens != nil {
					if msgData.Tokens.Input != nil {
						inputTokens = *msgData.Tokens.Input
					}
					if msgData.Tokens.Output != nil {
						outputTokens = *msgData.Tokens.Output
					}
					if msgData.Tokens.Cache != nil && msgData.Tokens.Cache.Read != nil {
						cacheReadTokens = *msgData.Tokens.Cache.Read
					}
					if msgData.Tokens.Reasoning != nil {
						reasoningTokens = *msgData.Tokens.Reasoning
					}
				}
			}
		}

		partRecords := partsByMsg[rm.ID]
		parts := make([]domain.LivePartFact, 0, len(partRecords))

		for _, rp := range partRecords {
			partFact := parsePartRecord(rp, &consecutiveErrors)
			parts = append(parts, partFact)

			if partFact.Type == "tool" {
				totalTools++
				if partFact.IsError {
					failedTools++
					recentErrors = append(recentErrors, partFact)
				}
			}

			// If message lacks reasoning but part has reasoning type, inherit reasoning
			if reasoningStr == nil && partFact.Type == "reasoning" && partFact.OutputSnippet != "" {
				res := partFact.OutputSnippet
				reasoningStr = &res
			}
		}

		contextTokens := inputTokens + cacheReadTokens
		totalTokens := inputTokens + outputTokens + cacheReadTokens

		messages = append(messages, domain.LiveMessageFact{
			ID:               rm.ID,
			Role:             role,
			Content:          contentStr,
			Timestamp:        formatEpochMilli(rm.TimeCreated),
			ReasoningContent: reasoningStr,
			ContextTokens:    contextTokens,
			InputTokens:      inputTokens,
			OutputTokens:     outputTokens,
			CacheReadTokens:  cacheReadTokens,
			ReasoningTokens:  reasoningTokens,
			TotalTokens:      totalTokens,
			Parts:            parts,
		})
	}

	// Update transient in-memory ring buffer
	s.buffer.SetMessages(sessionID, messages)

	// Order recent errors most recent first
	for i, j := 0, len(recentErrors)-1; i < j; i, j = i+1, j-1 {
		recentErrors[i], recentErrors[j] = recentErrors[j], recentErrors[i]
	}

	trace := domain.LiveTraceView{
		SessionID:       sessionID,
		Messages:        messages,
		TotalToolCalls:  totalTools,
		FailedToolCalls: failedTools,
		RecentErrors:    recentErrors,
		Anomalies:       []domain.LiveAnomalyAlert{},
	}
	trace.Anomalies = domain.EvaluateLiveAnomalies(nil, &trace)
	if trace.Anomalies == nil {
		trace.Anomalies = []domain.LiveAnomalyAlert{}
	}

	return trace, nil
}

func parsePartRecord(rp RawPartRecord, consecutiveErrors *int) domain.LivePartFact {
	partType := "tool"
	var toolName *string
	var inputSnippet, outputSnippet, status string

	if strings.TrimSpace(rp.Data) != "" {
		var pData struct {
			Type  *string `json:"type"`
			Tool  *string `json:"tool"`
			Text  *string `json:"text"`
			State *struct {
				Status *string         `json:"status"`
				Input  json.RawMessage `json:"input"`
				Output json.RawMessage `json:"output"`
			} `json:"state"`
		}
		if err := json.Unmarshal([]byte(rp.Data), &pData); err == nil {
			if pData.Type != nil && strings.TrimSpace(*pData.Type) != "" {
				partType = strings.TrimSpace(*pData.Type)
			}
			if pData.Tool != nil && strings.TrimSpace(*pData.Tool) != "" {
				toolName = pData.Tool
			}
			if pData.State != nil {
				if pData.State.Status != nil {
					status = strings.TrimSpace(*pData.State.Status)
				}
				inputSnippet = rawMessageToString(pData.State.Input)
				outputSnippet = rawMessageToString(pData.State.Output)
			}
			if outputSnippet == "" && pData.Text != nil {
				outputSnippet = strings.TrimSpace(*pData.Text)
			}
		}
	}

	if status == "" {
		status = "completed"
	}

	var isError bool
	var severity string
	var errorMessage *string

	if partType == "tool" {
		tName := ""
		if toolName != nil {
			tName = *toolName
		}
		isError, severity, errorMessage = domain.ClassifyToolFailure(tName, status, outputSnippet, *consecutiveErrors)
		if isError {
			*consecutiveErrors++
			if status == "completed" {
				status = "error"
			}
		} else {
			*consecutiveErrors = 0
		}
	} else {
		severity = string(domain.SeverityNone)
	}

	// Bound snippets
	if len(inputSnippet) > 500 {
		inputSnippet = inputSnippet[:500]
	}
	if len(outputSnippet) > 1000 {
		outputSnippet = outputSnippet[:1000]
	}

	return domain.LivePartFact{
		ID:            rp.ID,
		Type:          partType,
		ToolName:      toolName,
		InputSnippet:  inputSnippet,
		OutputSnippet: outputSnippet,
		Status:        status,
		IsError:       isError,
		ErrorMessage:  errorMessage,
		Severity:      severity,
	}
}

func validateSessionID(id string) error {
	clean := strings.TrimSpace(id)
	if clean == "" {
		return errors.New("empty session id")
	}
	if len(id) != len(clean) {
		return fmt.Errorf("session id contains whitespace: %q", id)
	}
	lower := strings.ToLower(clean)
	if strings.Contains(clean, "..") || strings.Contains(clean, "/") || strings.Contains(clean, "\\") ||
		strings.Contains(lower, "%2e") || strings.Contains(lower, "%2f") || strings.Contains(lower, "%5c") {
		return fmt.Errorf("session id contains path traversal characters: %s", clean)
	}
	for _, r := range clean {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' && r != '-' && r != ':' && r != '.' {
			return fmt.Errorf("session id contains invalid character %q: %s", r, clean)
		}
	}
	return nil
}

func rawMessageToString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return strings.TrimSpace(str)
	}
	return strings.TrimSpace(string(raw))
}

func extractMessageContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var sb strings.Builder
		for _, b := range blocks {
			if b.Text != "" {
				if sb.Len() > 0 {
					sb.WriteString("\n")
				}
				sb.WriteString(b.Text)
			}
		}
		return sb.String()
	}
	return string(raw)
}

func resolveModelName(rawModel string) string {
	trimmed := strings.TrimSpace(rawModel)
	if trimmed == "" {
		return "default"
	}
	if strings.HasPrefix(trimmed, "{") {
		var m struct {
			ID *string `json:"id"`
		}
		if err := json.Unmarshal([]byte(trimmed), &m); err == nil && m.ID != nil && strings.TrimSpace(*m.ID) != "" {
			return strings.TrimSpace(*m.ID)
		}
	}
	return trimmed
}

func countTotalNodes(roots []*domain.LiveSessionNode) int {
	count := 0
	visited := make(map[string]bool)
	var countNode func(n *domain.LiveSessionNode)
	countNode = func(n *domain.LiveSessionNode) {
		if n == nil || visited[n.ID] {
			return
		}
		visited[n.ID] = true
		count++
		for _, sub := range n.SubSessions {
			countNode(sub)
		}
	}
	for _, r := range roots {
		countNode(r)
	}
	return count
}

func formatEpochMilli(ms int64) string {
	if ms <= 0 {
		return ""
	}
	t := time.UnixMilli(ms).UTC()
	return t.Format("2006-01-02T15:04:05.000000+00:00")
}

// HydrateWorkspaceEvents dynamically hydrates full execution events from the local SQLite database.
// If the database is missing, sessions are not found, or an error occurs, it gracefully falls back
// to the committed ledger facts. Zero transient data is persisted to disk.
func (s *LiveService) HydrateWorkspaceEvents(ctx context.Context, ticketID string, ledger domain.AgenticMetricsLedger) (domain.WorkspaceEventsResponse, error) {
	fallbackResp := domain.WorkspaceEventsResponse{
		TicketID:   ticketID,
		IsHydrated: false,
		Events:     domain.SortEvents(ledger.Facts.Events),
		DataSource: domain.DataSourceCommittedLedger,
	}

	if s == nil || s.hook == nil || s.hook.DB() == nil {
		return fallbackResp, nil
	}

	// 1. Collect session IDs from ledger invocations and events
	sessionIDMap := make(map[string]bool)
	for _, inv := range ledger.Facts.Invocations {
		trimmed := strings.TrimSpace(inv.ID)
		if trimmed != "" {
			sessionIDMap[trimmed] = true
		}
	}
	for _, ev := range ledger.Facts.Events {
		trimmed := strings.TrimSpace(ev.SessionID)
		if trimmed != "" {
			sessionIDMap[trimmed] = true
		}
	}

	// If no session IDs in ledger, attempt resolution via OpenCode resolver
	if len(sessionIDMap) == 0 {
		resolvedID, err := opencode.ResolveActiveRootSession(s.hook.DB(), "", ticketID, "")
		if err == nil && resolvedID != nil && strings.TrimSpace(*resolvedID) != "" {
			sessionIDMap[strings.TrimSpace(*resolvedID)] = true
		}
	}

	if len(sessionIDMap) == 0 {
		return fallbackResp, nil
	}

	// 2. Discover child sub-sessions to gather complete execution tree
	var initialIDs []string
	for sid := range sessionIDMap {
		initialIDs = append(initialIDs, sid)
	}

	frontier := initialIDs
	for len(frontier) > 0 {
		children, err := s.hook.ResolveChildSessionIDs(frontier)
		if err != nil {
			break
		}
		var nextFrontier []string
		for _, cid := range children {
			if !sessionIDMap[cid] {
				sessionIDMap[cid] = true
				nextFrontier = append(nextFrontier, cid)
			}
		}
		frontier = nextFrontier
	}

	// 3. Verify existence in SQLite
	var verifiedSessions []string
	for sid := range sessionIDMap {
		exists, err := s.hook.SessionExists(sid)
		if err == nil && exists {
			verifiedSessions = append(verifiedSessions, sid)
		}
	}

	if len(verifiedSessions) == 0 {
		return fallbackResp, nil
	}

	// 4. Map session IDs to roles
	sessionRoleMap := make(map[string]string)
	for _, inv := range ledger.Facts.Invocations {
		if inv.ID != "" && inv.Role != "" {
			sessionRoleMap[inv.ID] = inv.Role
		}
	}
	for _, sid := range verifiedSessions {
		if _, exists := sessionRoleMap[sid]; !exists {
			var agent, parentID sql.NullString
			if err := s.hook.DB().QueryRow("SELECT agent, parent_id FROM session WHERE id = ?", sid).Scan(&agent, &parentID); err == nil {
				if agent.Valid && strings.TrimSpace(agent.String) != "" {
					sessionRoleMap[sid] = string(s.cfg.RoleMappings.ResolveRole(strings.TrimSpace(agent.String)))
				} else if !parentID.Valid || strings.TrimSpace(parentID.String) == "" {
					sessionRoleMap[sid] = "orchestrator"
				} else {
					sessionRoleMap[sid] = "unassigned"
				}
			} else {
				sessionRoleMap[sid] = "coder"
			}
		}
	}

	// 5. Fetch all messages and parts ordered by time_created ASC
	rawMsgs, rawParts, err := s.hook.FetchSessionTraceMulti(verifiedSessions)
	if err != nil {
		return fallbackResp, nil
	}

	if len(rawMsgs) == 0 && len(rawParts) == 0 {
		return fallbackResp, nil
	}

	// Group parts by message ID
	partsByMsg := make(map[string][]RawPartRecord)
	orphanParts := make(map[string][]RawPartRecord)
	for _, p := range rawParts {
		if p.MessageID != "" {
			partsByMsg[p.MessageID] = append(partsByMsg[p.MessageID], p)
		} else {
			orphanParts[p.SessionID] = append(orphanParts[p.SessionID], p)
		}
	}

	var events []domain.EventFact
	sessionTurns := make(map[string]int64)
	lastUserPromptPerSession := make(map[string]string)

	// Process messages per session
	for _, rm := range rawMsgs {
		role := sessionRoleMap[rm.SessionID]
		if role == "" {
			role = "coder"
		}

		var msgRole string
		var contentStr string
		var reasoningStr *string
		var ctxTokens int64
		var msgDurMs *int64

		if strings.TrimSpace(rm.Data) != "" {
			var rawMD struct {
				Role             *string         `json:"role"`
				Content          json.RawMessage `json:"content"`
				Reasoning        *string         `json:"reasoning"`
				ReasoningContent *string         `json:"reasoning_content"`
				Tokens           *struct {
					Input  *int64 `json:"input"`
					Output *int64 `json:"output"`
					Cache  *struct {
						Read *int64 `json:"read"`
					} `json:"cache"`
					Reasoning *int64 `json:"reasoning"`
				} `json:"tokens"`
				Time *struct {
					Created   *float64 `json:"created"`
					Completed *float64 `json:"completed"`
				} `json:"time"`
			}
			if err := json.Unmarshal([]byte(rm.Data), &rawMD); err == nil {
				if rawMD.Role != nil {
					msgRole = strings.TrimSpace(*rawMD.Role)
				}
				contentStr = extractMessageContent(rawMD.Content)
				if rawMD.ReasoningContent != nil && strings.TrimSpace(*rawMD.ReasoningContent) != "" {
					reasoningStr = rawMD.ReasoningContent
				} else if rawMD.Reasoning != nil && strings.TrimSpace(*rawMD.Reasoning) != "" {
					reasoningStr = rawMD.Reasoning
				}
				if rawMD.Tokens != nil {
					var inp, cache int64
					if rawMD.Tokens.Input != nil {
						inp = *rawMD.Tokens.Input
					}
					if rawMD.Tokens.Cache != nil && rawMD.Tokens.Cache.Read != nil {
						cache = *rawMD.Tokens.Cache.Read
					}
					ctxTokens = inp + cache
				}
				if rawMD.Time != nil && rawMD.Time.Created != nil && rawMD.Time.Completed != nil {
					dur := int64(*rawMD.Time.Completed) - int64(*rawMD.Time.Created)
					if dur >= 0 {
						msgDurMs = &dur
					}
				}
			}
		}

		if msgRole == "user" {
			if contentStr != "" {
				lastUserPromptPerSession[rm.SessionID] = contentStr
			}
			// User messages provide context for assistant turns
			continue
		}

		if msgDurMs == nil && rm.TimeUpdated >= rm.TimeCreated && rm.TimeCreated > 0 {
			dur := rm.TimeUpdated - rm.TimeCreated
			msgDurMs = &dur
		}

		sessionTurns[rm.SessionID]++
		turn := sessionTurns[rm.SessionID]

		ts := formatEpochMilli(rm.TimeCreated)

		// Check parts for this message
		partRecords := partsByMsg[rm.ID]

		// Inherit reasoning from part if not already on message
		if reasoningStr == nil {
			for _, pr := range partRecords {
				var pCheck struct {
					Type *string `json:"type"`
					Text *string `json:"text"`
				}
				if json.Unmarshal([]byte(pr.Data), &pCheck) == nil {
					if pCheck.Type != nil && *pCheck.Type == "reasoning" && pCheck.Text != nil && *pCheck.Text != "" {
						t := strings.TrimSpace(*pCheck.Text)
						reasoningStr = &t
						break
					}
				}
			}
		}

		var promptContentPtr *string
		if userPrompt, ok := lastUserPromptPerSession[rm.SessionID]; ok && userPrompt != "" {
			promptContentPtr = &userPrompt
			delete(lastUserPromptPerSession, rm.SessionID)
		} else if contentStr != "" {
			promptContentPtr = &contentStr
		}

		// Append turn_message event
		events = append(events, domain.EventFact{
			Timestamp:        ts,
			SessionID:        rm.SessionID,
			Turn:             turn,
			Role:             role,
			EventType:        domain.EventTypeTurnMessage,
			ToolName:         nil,
			Status:           "completed",
			DurationMs:       msgDurMs,
			ContextTokens:    ctxTokens,
			PromptContent:    promptContentPtr,
			ReasoningContent: reasoningStr,
		})

		// Append parts for this message
		consecutiveErrors := 0
		for _, pr := range partRecords {
			partEvent := convertPartToEventFact(pr, turn, role, ctxTokens, &consecutiveErrors)
			if partEvent != nil {
				events = append(events, *partEvent)
			}
		}
	}

	// Append any orphan parts that weren't tied to a specific message ID
	for sid, parts := range orphanParts {
		role := sessionRoleMap[sid]
		if role == "" {
			role = "coder"
		}
		turn := sessionTurns[sid]
		if turn == 0 {
			turn = 1
		}
		consecutiveErrors := 0
		for _, pr := range parts {
			partEvent := convertPartToEventFact(pr, turn, role, 0, &consecutiveErrors)
			if partEvent != nil {
				events = append(events, *partEvent)
			}
		}
	}

	if len(events) == 0 {
		return fallbackResp, nil
	}

	return domain.WorkspaceEventsResponse{
		TicketID:   ticketID,
		IsHydrated: true,
		Events:     domain.SortEvents(events),
		DataSource: domain.DataSourceLocalSQLite,
	}, nil
}

func convertPartToEventFact(pr RawPartRecord, turn int64, role string, ctxTokens int64, consecutiveErrors *int) *domain.EventFact {
	ts := formatEpochMilli(pr.TimeCreated)

	var pData struct {
		Type  *string `json:"type"`
		Tool  *string `json:"tool"`
		Text  *string `json:"text"`
		State *struct {
			Status *string         `json:"status"`
			Input  json.RawMessage `json:"input"`
			Output json.RawMessage `json:"output"`
		} `json:"state"`
	}

	partType := "tool"
	var toolName *string
	var inputStr, outputStr, status string

	if strings.TrimSpace(pr.Data) != "" {
		if err := json.Unmarshal([]byte(pr.Data), &pData); err == nil {
			if pData.Type != nil && strings.TrimSpace(*pData.Type) != "" {
				partType = strings.TrimSpace(*pData.Type)
			}
			if pData.Tool != nil && strings.TrimSpace(*pData.Tool) != "" {
				toolName = pData.Tool
			}
			if pData.State != nil {
				if pData.State.Status != nil {
					status = strings.TrimSpace(*pData.State.Status)
				}
				inputStr = rawMessageToString(pData.State.Input)
				outputStr = rawMessageToString(pData.State.Output)
			}
			if outputStr == "" && pData.Text != nil {
				outputStr = strings.TrimSpace(*pData.Text)
			}
		}
	}

	if status == "" {
		status = "completed"
	}

	var durMs *int64
	if pr.TimeUpdated >= pr.TimeCreated && pr.TimeCreated > 0 {
		dur := pr.TimeUpdated - pr.TimeCreated
		durMs = &dur
	}

	if partType == "compaction" {
		return &domain.EventFact{
			Timestamp:     ts,
			SessionID:     pr.SessionID,
			Turn:          turn,
			Role:          role,
			EventType:     domain.EventTypeCompaction,
			ToolName:      nil,
			Status:        "completed",
			DurationMs:    durMs,
			ContextTokens: ctxTokens,
		}
	}

	if partType == "tool" {
		tName := "unknown"
		if toolName != nil {
			tName = *toolName
		}
		isError, _, errorMessage := domain.ClassifyToolFailure(tName, status, outputStr, *consecutiveErrors)
		if isError {
			*consecutiveErrors++
			status = "error"
		} else {
			*consecutiveErrors = 0
		}

		var inputPtr *string
		if inputStr != "" {
			inputPtr = &inputStr
		}
		var outputPtr *string
		if outputStr != "" {
			outputPtr = &outputStr
		}

		return &domain.EventFact{
			Timestamp:     ts,
			SessionID:     pr.SessionID,
			Turn:          turn,
			Role:          role,
			EventType:     domain.EventTypeToolCall,
			ToolName:      toolName,
			Status:        status,
			ErrorMessage:  errorMessage,
			DurationMs:    durMs,
			ContextTokens: ctxTokens,
			ToolInput:     inputPtr,
			ToolOutput:    outputPtr,
		}
	}

	return nil
}
