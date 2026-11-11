package live

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"agenticlens/internal/adapters/codex"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

var (
	ErrSourceUnavailable = errors.New("live OpenCode source unavailable; configure an accessible telemetry.sources.opencode.database_path in the selected local YAML")
	// ErrSessionNotFound is returned when a session ID does not exist in the database.
	ErrSessionNotFound = errors.New("session not found")
	// ErrInvalidSessionID is returned when a session ID is blank or contains path traversal characters.
	ErrInvalidSessionID = errors.New("invalid session id")
)

// LiveEventType represents an SSE event kind.
type LiveEventType string

const (
	LiveEventSessionTree  LiveEventType = "session_tree"
	LiveEventTraceUpdate  LiveEventType = "trace_update"
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
	cfg             domain.ProjectConfig
	codexHome       string
	subscribers     map[chan LiveEvent]struct{}
	subMu           sync.RWMutex
	stopCh          chan struct{}
	closeOnce       sync.Once
	lastTreeJSON    []byte
	lastTreeMu      sync.Mutex
	lastTraces      map[string]*domain.LiveTraceView
	traceMu         sync.Mutex
	traceRevision   SourceRevision
	traceCache      map[string]*traceSnapshot
	discovered      map[string]struct{}
	selectedLookup  string
	lastTraceMu     sync.Mutex
	seenAnomalies   map[string]bool
	seenAnomaliesMu sync.Mutex
}

// NewLiveService constructs a new LiveService instance.
func NewLiveService(hook *SQLiteHook, cfg domain.ProjectConfig) *LiveService {
	return NewLiveServiceWithCodexHome(hook, cfg, cfg.Telemetry.Sources.CodexHome)
}

// NewLiveServiceWithCodexHome constructs a hydrator with an explicit local Codex data root.
func NewLiveServiceWithCodexHome(hook *SQLiteHook, cfg domain.ProjectConfig, codexHome string) *LiveService {

	return &LiveService{
		hook:          hook,
		codexHome:     codexHome,
		cfg:           cfg,
		subscribers:   make(map[chan LiveEvent]struct{}),
		stopCh:        make(chan struct{}),
		lastTraces:    make(map[string]*domain.LiveTraceView),
		traceCache:    make(map[string]*traceSnapshot),
		discovered:    make(map[string]struct{}),
		seenAnomalies: make(map[string]bool),
	}
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
func (s *LiveService) Start(ctx context.Context) error {
	if err := errors.Join(adaptersconfig.ValidatePollingPolicy(s.cfg), adaptersconfig.ValidateActivityPolicy(s.cfg), adaptersconfig.ValidateAnomalyPolicy(s.cfg)); err != nil {
		return err
	}
	intervalMs := s.cfg.Live.PollIntervalMs
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
	return nil
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

	compareTree := tree
	compareTree.LastRefreshedAt = ""
	compareData, err := json.Marshal(compareTree)
	if err != nil {
		return
	}

	s.lastTreeMu.Lock()
	changed := !bytes.Equal(s.lastTreeJSON, compareData)
	if changed {
		s.lastTreeJSON = compareData
	}
	s.lastTreeMu.Unlock()

	s.checkAndBroadcastAnomalies(tree.ActiveAnomalies)

	if changed {
		s.broadcast(LiveEvent{
			Type: LiveEventSessionTree,
			Data: data,
		})
	}
	s.broadcastTraceUpdates(ctx, tree)
}

func (s *LiveService) broadcastTraceUpdates(ctx context.Context, tree domain.LiveSessionTree) {
	current := make(map[string]struct{})
	var visit func([]*domain.LiveSessionNode)
	visit = func(nodes []*domain.LiveSessionNode) {
		for _, node := range nodes {
			if node == nil || strings.TrimSpace(node.ID) == "" {
				continue
			}
			current[node.ID] = struct{}{}
			snapshot, err := s.sessionTraceSnapshot(ctx, node.ID)
			if err == nil {
				trace := &snapshot.Trace
				s.lastTraceMu.Lock()
				previous := s.lastTraces[node.ID]
				changed := previous != trace && (previous == nil || !domain.LiveTracesEqual(*previous, *trace))
				if !changed {
					s.lastTraces[node.ID] = trace
				}
				s.lastTraceMu.Unlock()
				if changed {
					data, marshalErr := json.Marshal(trace)
					if marshalErr == nil {
						s.lastTraceMu.Lock()
						s.lastTraces[node.ID] = trace
						s.lastTraceMu.Unlock()
						s.broadcast(LiveEvent{Type: LiveEventTraceUpdate, Data: data})
					}
				}
			}
			visit(node.SubSessions)
		}
	}
	visit(tree.RootSessions)

	s.lastTraceMu.Lock()
	for sessionID := range s.lastTraces {
		if _, active := current[sessionID]; !active {
			delete(s.lastTraces, sessionID)
		}
	}
	s.lastTraceMu.Unlock()
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

// CheckAvailability distinguishes an unconfigured/unavailable source from a
// selected read-only database containing no active sessions.
func (s *LiveService) CheckAvailability() error {
	if s.hook == nil {
		return ErrSourceUnavailable
	}
	return nil
}

// GetSessionTree queries active sessions, runs title ticket inference, and constructs the session tree.
// Zero Git branch calls, zero workspace scans.
func (s *LiveService) GetSessionTree(ctx context.Context) (domain.LiveSessionTree, error) {
	if err := s.CheckAvailability(); err != nil {
		return domain.LiveSessionTree{}, err
	}

	if err := adaptersconfig.ValidateActivityPolicy(s.cfg); err != nil {
		return domain.LiveSessionTree{}, err
	}
	activeThresholdSec := int64(s.cfg.Live.ActiveThresholdSeconds)
	nowMs := time.Now().UnixMilli()

	records, err := s.hook.FetchActiveSessions(activeThresholdSec, nowMs)
	if err != nil {
		return domain.LiveSessionTree{}, fmt.Errorf("fetching active sessions: %w", err)
	}

	s.traceMu.Lock()
	s.discovered = make(map[string]struct{}, len(records))
	for _, rec := range records {
		s.discovered[rec.ID] = struct{}{}
	}
	for id := range s.traceCache {
		if _, exists := s.discovered[id]; !exists && id != s.selectedLookup {
			delete(s.traceCache, id)
		}
	}
	s.traceMu.Unlock()

	stallThresholdSec := int64(s.cfg.AnomalyThresholds.StallThresholdSeconds)

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
		snapshot, err := s.sessionTraceSnapshot(ctx, rec.ID)
		if err == nil && len(snapshot.Trace.Anomalies) > 0 {
			sessionAnomalies = snapshot.Trace.Anomalies
		}
		if err == nil {
			persona := snapshot.Persona.Resolve(rec.Agent)
			rec.Agent = string(persona)
			role = s.cfg.RoleMappings.ResolveRole(rec.Agent)
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

// GetSessionTrace returns a fresh source snapshot, reusing unchanged read-only data.
func (s *LiveService) GetSessionTrace(ctx context.Context, sessionID string) (domain.LiveTraceView, error) {
	trace, err := s.sessionTraceSnapshot(ctx, sessionID)
	if err != nil {
		return domain.LiveTraceView{}, err
	}
	return trace.Trace, nil
}

// traceSnapshot binds persona evidence to the exact trace read that produced it.
// Entries are immutable after construction and survive cache invalidation for
// readers already holding a snapshot.
type traceSnapshot struct {
	Trace   domain.LiveTraceView
	Persona domain.PersonaEvidence
}

func (s *LiveService) sessionTraceSnapshot(ctx context.Context, sessionID string) (*traceSnapshot, error) {
	if err := validateSessionID(sessionID); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidSessionID, err.Error())
	}
	if s.hook == nil {
		return nil, ErrSourceUnavailable
	}
	s.traceMu.Lock()
	defer s.traceMu.Unlock()
	revision, err := s.hook.SourceRevision(ctx)
	if err != nil {
		s.traceRevision = SourceRevision{}
		return nil, err
	}
	if revision != s.traceRevision {
		s.traceCache = make(map[string]*traceSnapshot)
		s.traceRevision = revision
	}
	if _, visible := s.discovered[sessionID]; !visible {
		if s.selectedLookup != sessionID {
			if _, discovered := s.discovered[s.selectedLookup]; !discovered {
				delete(s.traceCache, s.selectedLookup)
			}
			s.selectedLookup = sessionID
		}
	}
	if cached := s.traceCache[sessionID]; cached != nil {
		return cached, nil
	}
	trace, err := s.readSessionTrace(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	after, err := s.hook.SourceRevision(ctx)
	if err != nil {
		s.traceRevision = SourceRevision{}
		return nil, err
	}
	if after == revision {
		s.traceCache[sessionID] = &trace
	} else {
		s.traceRevision = SourceRevision{}
	}
	return &trace, nil
}

func (s *LiveService) readSessionTrace(ctx context.Context, sessionID string) (traceSnapshot, error) {
	exists, err := s.hook.SessionExists(sessionID)
	if err != nil {
		return traceSnapshot{}, fmt.Errorf("verifying session existence: %w", err)
	}
	if !exists {
		return traceSnapshot{}, fmt.Errorf("%w: %s", ErrSessionNotFound, sessionID)
	}

	rawMsgs, rawParts, err := s.hook.FetchSessionTrace(sessionID)
	if err != nil {
		return traceSnapshot{}, fmt.Errorf("fetching session trace for %s: %w", sessionID, err)
	}
	childIDs, childErr := s.hook.ResolveChildSessionIDs([]string{sessionID})
	knownChildren := make(map[string]struct{}, len(childIDs))
	if childErr == nil {
		for _, childID := range childIDs {
			knownChildren[childID] = struct{}{}
		}
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
		role := domain.AgentRole("assistant")
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
					Input  *int64 `json:"input"`
					Output *int64 `json:"output"`
					Cache  *struct {
						Read *int64 `json:"read"`
					} `json:"cache"`
					Reasoning *int64 `json:"reasoning"`
				} `json:"tokens"`
			}
			if err := json.Unmarshal([]byte(rm.Data), &msgData); err == nil {
				if msgData.Role != nil && strings.TrimSpace(*msgData.Role) != "" {
					role = domain.AgentRole(strings.TrimSpace(*msgData.Role))
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
			if partFact.SubagentType != "" {
				if childID := explicitOutputSessionID(partFact.OutputSnippet); childID != "" {
					if _, exists := knownChildren[childID]; exists {
						partFact.ChildSessionID = childID
					}
				}
			}
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

	// Order recent errors most recent first
	for i, j := 0, len(recentErrors)-1; i < j; i, j = i+1, j-1 {
		recentErrors[i], recentErrors[j] = recentErrors[j], recentErrors[i]
	}

	relationships := domain.LiveRelationships{State: domain.LiveRelationshipsComplete}
	diagnostics := []domain.LiveTraceDiagnostic{}
	if childErr != nil {
		relationships.State = domain.LiveRelationshipsUnavailable
		diagnostics = append(diagnostics, domain.LiveTraceDiagnostic{Code: domain.DiagnosticChildRelationshipLookupFailed, Message: fmt.Sprintf("Native child relationships could not be checked: %v", childErr)})
	}
	trace := domain.LiveTraceView{
		Relationships: relationships, Diagnostics: diagnostics,
		SessionID:       sessionID,
		Messages:        messages,
		TotalToolCalls:  totalTools,
		FailedToolCalls: failedTools,
		RecentErrors:    recentErrors,
		Anomalies:       []domain.LiveAnomalyAlert{},
	}
	if err := adaptersconfig.ValidateAnomalyPolicy(s.cfg); err != nil {
		return traceSnapshot{}, err
	}
	anomalyOpts := domain.LiveAnomalyOptions{
		Thresholds: s.cfg.AnomalyThresholds,
	}
	trace.Anomalies = domain.EvaluateLiveAnomalies(nil, &trace, anomalyOpts)
	if trace.Anomalies == nil {
		trace.Anomalies = []domain.LiveAnomalyAlert{}
	}

	return traceSnapshot{Trace: trace, Persona: assistantPersonaEvidence(rawMsgs)}, nil
}

func assistantPersonaEvidence(messages []RawMessageRecord) domain.PersonaEvidence {
	var evidence domain.PersonaEvidence
	for _, message := range messages {
		var metadata struct {
			Role  string `json:"role"`
			Agent string `json:"agent"`
		}
		if err := json.Unmarshal([]byte(message.Data), &metadata); err == nil && metadata.Role == "assistant" {
			evidence = evidence.WithCandidate(metadata.Agent)
		}
	}
	return evidence
}

func parsePartRecord(rp RawPartRecord, consecutiveErrors *int) domain.LivePartFact {
	partType := "unknown"
	var toolName *string
	var inputSnippet, outputSnippet, status string
	var skillName, subagentType string
	var process domain.ToolProcessEvidence

	if strings.TrimSpace(rp.Data) != "" {
		var pData opencode.SchemaV1PartData
		if err := json.Unmarshal([]byte(rp.Data), &pData); err == nil {
			if pData.Type != nil && strings.TrimSpace(*pData.Type) != "" {
				partType = strings.TrimSpace(*pData.Type)
			}
			if pData.Tool != nil && strings.TrimSpace(*pData.Tool) != "" {
				toolName = pData.Tool
			}
			if pData.State != nil {
				process = pData.ProcessEvidence()
				if pData.State.Status != nil {
					status = strings.TrimSpace(*pData.State.Status)
				}
				inputSnippet = rawMessageToString(pData.State.Input)
				outputSnippet = rawMessageToString(pData.State.Output)
				if outputSnippet == "" && pData.State.Error != nil {
					outputSnippet = opencode.SchemaV1RawValue(pData.State.Error)
				}
				var input map[string]json.RawMessage
				if json.Unmarshal(pData.State.Input, &input) == nil {
					if toolName != nil && strings.TrimSpace(*toolName) == "skill" {
						var name string
						if json.Unmarshal(input["name"], &name) == nil {
							skillName = strings.TrimSpace(name)
						}
					}
					if toolName != nil && strings.TrimSpace(*toolName) == "task" {
						var parsedSubagentType string
						if json.Unmarshal(input["subagent_type"], &parsedSubagentType) == nil {
							subagentType = strings.TrimSpace(parsedSubagentType)
						}
					}
				}
			}
			if outputSnippet == "" && pData.Text != nil {
				outputSnippet = *pData.Text
			}
		}
	}

	if status == "" {
		status = "unknown"
	}

	var isError bool
	var severity domain.FailureSeverity
	var errorMessage *string

	if partType == "tool" {
		tName := ""
		if toolName != nil {
			tName = *toolName
		}
		classification := domain.ClassifyToolExecution(tName, status, outputSnippet, *consecutiveErrors, process)
		isError, severity, errorMessage = classification.IsError, classification.Severity, classification.ErrorMessage
		status = classification.Status
		if isError {
			*consecutiveErrors++
			status = "error"
		} else {
			*consecutiveErrors = 0
		}
	} else {
		severity = domain.SeverityNone
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
		SkillName:     skillName,
		SubagentType:  subagentType,
	}
}

func explicitOutputSessionID(output string) string {
	var fields map[string]json.RawMessage
	if json.Unmarshal([]byte(output), &fields) != nil {
		return ""
	}
	for _, key := range []string{"child_session_id", "session_id", "sessionID"} {
		var value string
		if json.Unmarshal(fields[key], &value) == nil && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
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
	return opencode.SchemaV1RawValue(raw)
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

// HydrateWorkspaceEvents returns request-local source evidence. Ledger facts
// are never mutated or used to synthesize trace events.
func (s *LiveService) HydrateWorkspaceEvents(ctx context.Context, ticketID string, ledger domain.AgenticMetricsLedger) (domain.WorkspaceEventsResponse, error) {
	response := domain.WorkspaceEventsResponse{TicketID: ticketID, Events: []domain.EventFact{}, Anomalies: ledger.Facts.AnomalyFacts(), Availability: domain.TraceUnavailable, DataSource: domain.DataSourceUnavailable, Coverage: domain.TraceCoverage{RequiredSessionIDs: []string{}, CompleteSessionIDs: []string{}, Diagnostics: []string{}}}
	var transcripts []codex.Transcript
	openCodeObserved, codexObserved := false, false
	for _, owner := range ledger.Facts.Sessions {
		response.Coverage.RequiredSessionIDs = append(response.Coverage.RequiredSessionIDs, owner.ID)
		if err := ctx.Err(); err != nil {
			response.Coverage.Diagnostics = append(response.Coverage.Diagnostics, owner.ID+": "+err.Error())
			continue
		}
		switch owner.Source {
		case domain.SourceOpenCode:
			if s == nil || s.hook == nil || s.hook.DB() == nil {
				response.Coverage.Diagnostics = append(response.Coverage.Diagnostics, owner.ID+": OpenCode source unavailable")
				continue
			}
			events, err := opencode.ExtractSessionSourceEvents(s.hook.DB(), owner.ID, owner.User)
			if err == nil || len(events) > 0 {
				openCodeObserved = true
			}
			if err != nil {
				response.Coverage.Diagnostics = append(response.Coverage.Diagnostics, owner.ID+": "+err.Error())
				response.Events = append(response.Events, events...)
				continue
			}
			response.Events = append(response.Events, events...)
			response.Coverage.CompleteSessionIDs = append(response.Coverage.CompleteSessionIDs, owner.ID)
		case domain.SourceCodex:
			if s == nil || strings.TrimSpace(s.codexHome) == "" {
				response.Coverage.Diagnostics = append(response.Coverage.Diagnostics, owner.ID+": Codex source unavailable")
				continue
			}
			paths, err := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: s.codexHome, SessionID: owner.ID})
			if err != nil || len(paths) == 0 {
				response.Coverage.Diagnostics = append(response.Coverage.Diagnostics, owner.ID+": Codex native session unavailable")
				continue
			}
			found := false
			for _, path := range paths {
				transcript, parseErr := codex.ParseRolloutPath(path)
				if parseErr != nil {
					response.Coverage.Diagnostics = append(response.Coverage.Diagnostics, owner.ID+": "+parseErr.Error())
					continue
				}
				id := transcript.ThreadID
				if id == "" {
					id = transcript.SessionID
				}
				if id != owner.ID {
					continue
				}
				if transcript.ParentThread != "" && (owner.ParentID == nil || *owner.ParentID != transcript.ParentThread) {
					response.Coverage.Diagnostics = append(response.Coverage.Diagnostics, owner.ID+": source parent contradicts recorded owner")
					continue
				}
				transcripts = append(transcripts, transcript)
				if len(transcript.Warnings) == 0 || len(transcript.Events) > 0 {
					codexObserved = true
				}
				found = true
				if len(transcript.Warnings) == 0 {
					response.Coverage.CompleteSessionIDs = append(response.Coverage.CompleteSessionIDs, owner.ID)
				} else {
					response.Coverage.Diagnostics = append(response.Coverage.Diagnostics, owner.ID+": incomplete Codex source records")
				}
				break
			}
			if !found {
				response.Coverage.Diagnostics = append(response.Coverage.Diagnostics, owner.ID+": exact Codex native session unavailable")
			}
		default:
			response.Coverage.Diagnostics = append(response.Coverage.Diagnostics, owner.ID+": unsupported source")
		}
	}
	if len(transcripts) > 0 {
		combined, err := codex.CombineTranscripts(transcripts)
		if err != nil {
			response.Coverage.Diagnostics = append(response.Coverage.Diagnostics, err.Error())
		} else {
			response.Events = append(response.Events, combined.Events...)
		}
	}
	if len(response.Coverage.RequiredSessionIDs) > 0 && len(response.Coverage.CompleteSessionIDs) == len(response.Coverage.RequiredSessionIDs) && len(response.Coverage.Diagnostics) == 0 {
		response.Availability = domain.TraceComplete
	} else if len(response.Events) > 0 || len(response.Coverage.CompleteSessionIDs) > 0 {
		response.Availability = domain.TracePartial
	}
	switch {
	case openCodeObserved && codexObserved:
		response.DataSource = domain.DataSourceMixed
	case openCodeObserved:
		response.DataSource = domain.DataSourceLocalSQLite
	case codexObserved:
		response.DataSource = domain.DataSourceLocalCodex
	}
	response.IsHydrated = response.Availability != domain.TraceUnavailable
	response.Events = domain.SortEvents(response.Events)
	source := domain.SourceEvents{Events: response.Events, CompleteSessionIDs: response.Coverage.CompleteSessionIDs}
	var thresholds []domain.AnomalyThresholdsConfig
	if s != nil {
		thresholds = append(thresholds, s.cfg.AnomalyThresholds)
	}
	response.Anomalies = hydrateAnomalyAssociations(ledger.Facts, source, thresholds...)
	if err := response.Validate(); err != nil {
		return domain.WorkspaceEventsResponse{}, err
	}
	return response, nil
}
func hydrateAnomalyAssociations(facts domain.AgenticFactLedger, source domain.SourceEvents, thresholds ...domain.AnomalyThresholdsConfig) []domain.Anomaly {
	anomalies, err := domain.DetectAnomaliesWithEvidence(facts, source, thresholds...)
	if err != nil {
		return domain.RefreshAnomalySummariesFromSource(facts, source, thresholds...)
	}
	return anomalies
}
