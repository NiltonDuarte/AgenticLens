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
	LiveEventSessionTree LiveEventType = "session_tree"
	LiveEventTraceUpdate LiveEventType = "trace_update"
)

// LiveEvent represents an event pushed to real-time stream subscribers.
type LiveEvent struct {
	Type LiveEventType `json:"type"`
	Data []byte        `json:"data"`
}

// LiveService coordinates SQLite polling, tree reconstruction, and real-time streaming.
type LiveService struct {
	hook         *SQLiteHook
	buffer       *TraceBuffer
	cfg          domain.ProjectConfig
	subscribers  map[chan LiveEvent]struct{}
	subMu        sync.RWMutex
	stopCh       chan struct{}
	closeOnce    sync.Once
	lastTreeJSON []byte
	lastTreeMu   sync.Mutex
}

// NewLiveService constructs a new LiveService instance.
func NewLiveService(hook *SQLiteHook, cfg domain.ProjectConfig) *LiveService {
	maxBuffer := cfg.Live.MaxTraceBufferMessages
	if maxBuffer <= 0 {
		maxBuffer = 500
	}

	return &LiveService{
		hook:        hook,
		buffer:      NewTraceBuffer(maxBuffer),
		cfg:         cfg,
		subscribers: make(map[chan LiveEvent]struct{}),
		stopCh:      make(chan struct{}),
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

	if changed {
		s.broadcast(LiveEvent{
			Type: LiveEventSessionTree,
			Data: data,
		})
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
			SubSessions:      []*domain.LiveSessionNode{},
		}
		nodes = append(nodes, node)
	}

	roots := domain.BuildSessionTree(nodes)
	totalActive := countTotalNodes(roots)
	refreshedAt := formatEpochMilli(nowMs)

	return domain.LiveSessionTree{
		RootSessions:        roots,
		LastRefreshedAt:     refreshedAt,
		TotalActiveSessions: totalActive,
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

		if strings.TrimSpace(rm.Data) != "" {
			var msgData struct {
				Role             *string         `json:"role"`
				Content          json.RawMessage `json:"content"`
				Reasoning        *string         `json:"reasoning"`
				ReasoningContent *string         `json:"reasoning_content"`
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

		messages = append(messages, domain.LiveMessageFact{
			ID:               rm.ID,
			Role:             role,
			Content:          contentStr,
			Timestamp:        formatEpochMilli(rm.TimeCreated),
			ReasoningContent: reasoningStr,
			Parts:            parts,
		})
	}

	// Update transient in-memory ring buffer
	s.buffer.SetMessages(sessionID, messages)

	// Order recent errors most recent first
	for i, j := 0, len(recentErrors)-1; i < j; i, j = i+1, j-1 {
		recentErrors[i], recentErrors[j] = recentErrors[j], recentErrors[i]
	}

	return domain.LiveTraceView{
		SessionID:       sessionID,
		Messages:        messages,
		TotalToolCalls:  totalTools,
		FailedToolCalls: failedTools,
		RecentErrors:    recentErrors,
	}, nil
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
	lower := strings.ToLower(clean)
	if strings.Contains(clean, "..") || strings.Contains(clean, "/") || strings.Contains(clean, "\\") ||
		strings.Contains(lower, "%2e") || strings.Contains(lower, "%2f") || strings.Contains(lower, "%5c") {
		return fmt.Errorf("session id contains path traversal characters: %s", clean)
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
