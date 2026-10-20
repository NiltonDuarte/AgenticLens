package http_test

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/adapters/live"
	"agenticlens/internal/domain"
)

type mockLiveService struct {
	tree         domain.LiveSessionTree
	trace        domain.LiveTraceView
	treeErr      error
	traceErr     error
	events       chan live.LiveEvent
	unsubMu      sync.Mutex
	unsubscribed bool
}

func (m *mockLiveService) GetSessionTree(ctx context.Context) (domain.LiveSessionTree, error) {
	if m.treeErr != nil {
		return domain.LiveSessionTree{}, m.treeErr
	}
	return m.tree, nil
}

func (m *mockLiveService) GetSessionTrace(ctx context.Context, sessionID string) (domain.LiveTraceView, error) {
	if m.traceErr != nil {
		return domain.LiveTraceView{}, m.traceErr
	}
	if sessionID != m.trace.SessionID {
		return domain.LiveTraceView{}, live.ErrSessionNotFound
	}
	return m.trace, nil
}

func (m *mockLiveService) Subscribe() (<-chan live.LiveEvent, func()) {
	cleanup := func() {
		m.unsubMu.Lock()
		m.unsubscribed = true
		m.unsubMu.Unlock()
	}
	return m.events, cleanup
}

func (m *mockLiveService) isUnsubscribed() bool {
	m.unsubMu.Lock()
	defer m.unsubMu.Unlock()
	return m.unsubscribed
}

func setupLiveTestServer(mock *mockLiveService) *httptest.Server {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html><body>Dashboard</body></html>")},
	}
	var svc adaptershttp.LiveServiceInterface
	if mock != nil {
		svc = mock
	}
	server := adaptershttp.NewServer("workspaces", mockFS, svc)
	return httptest.NewServer(server.Handler())
}

func TestLiveSessionsHandler_Success(t *testing.T) {
	guessed := "AL-17"
	mock := &mockLiveService{
		tree: domain.LiveSessionTree{
			RootSessions: []*domain.LiveSessionNode{
				{
					ID:               "ses_root_1",
					Agent:            "orchestrator",
					Role:             domain.RoleOrchestrator,
					Model:            "gemini-3.8-flash",
					Title:            "Task AL-17: implement live telemetry",
					TimeCreated:      "2026-10-19T10:00:00Z",
					TimeUpdated:      "2026-10-19T10:05:00Z",
					Status:           "running",
					GuessedTicketID:  &guessed,
					TicketConfidence: 1.0,
					SubSessions:      []*domain.LiveSessionNode{},
				},
			},
			LastRefreshedAt:     "2026-10-19T10:05:01Z",
			TotalActiveSessions: 1,
		},
		events: make(chan live.LiveEvent, 1),
	}

	ts := setupLiveTestServer(mock)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/live/sessions")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("expected application/json, got %s", contentType)
	}

	var tree domain.LiveSessionTree
	if err := json.NewDecoder(resp.Body).Decode(&tree); err != nil {
		t.Fatalf("failed decoding json: %v", err)
	}

	if tree.TotalActiveSessions != 1 {
		t.Errorf("expected 1 active session, got %d", tree.TotalActiveSessions)
	}
	if len(tree.RootSessions) != 1 {
		t.Fatalf("expected 1 root session, got %d", len(tree.RootSessions))
	}
	if tree.RootSessions[0].GuessedTicketID == nil || *tree.RootSessions[0].GuessedTicketID != "AL-17" {
		t.Errorf("expected guessed ticket AL-17, got %v", tree.RootSessions[0].GuessedTicketID)
	}
}

func TestLiveSessionsHandler_NilService(t *testing.T) {
	ts := setupLiveTestServer(nil)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/live/sessions")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for nil service fallback, got %d", resp.StatusCode)
	}

	var tree domain.LiveSessionTree
	if err := json.NewDecoder(resp.Body).Decode(&tree); err != nil {
		t.Fatalf("failed decoding json: %v", err)
	}
	if len(tree.RootSessions) != 0 {
		t.Errorf("expected empty root sessions, got %d", len(tree.RootSessions))
	}
}

func TestLiveTraceHandler_SuccessAndNotFound(t *testing.T) {
	toolName := "bash"
	errMsg := "exit status 1"
	mock := &mockLiveService{
		trace: domain.LiveTraceView{
			SessionID: "ses_123",
			Messages: []domain.LiveMessageFact{
				{
					ID:        "msg_1",
					Role:      "assistant",
					Content:   "running command",
					Timestamp: "2026-10-19T10:00:00Z",
					Parts: []domain.LivePartFact{
						{
							ID:           "part_1",
							Type:         "tool",
							ToolName:     &toolName,
							InputSnippet: "go test",
							Status:       "error",
							IsError:      true,
							ErrorMessage: &errMsg,
							Severity:     "medium",
						},
					},
				},
			},
			TotalToolCalls:  1,
			FailedToolCalls: 1,
			RecentErrors: []domain.LivePartFact{
				{
					ID:           "part_1",
					Type:         "tool",
					ToolName:     &toolName,
					InputSnippet: "go test",
					Status:       "error",
					IsError:      true,
					ErrorMessage: &errMsg,
					Severity:     "medium",
				},
			},
		},
		events: make(chan live.LiveEvent, 1),
	}

	ts := setupLiveTestServer(mock)
	defer ts.Close()

	// 1. Success case
	resp, err := http.Get(ts.URL + "/api/live/sessions/ses_123/trace")
	if err != nil {
		t.Fatalf("trace request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var trace domain.LiveTraceView
	if err := json.NewDecoder(resp.Body).Decode(&trace); err != nil {
		t.Fatalf("failed decoding trace view: %v", err)
	}
	if trace.SessionID != "ses_123" {
		t.Errorf("expected session ses_123, got %s", trace.SessionID)
	}
	if trace.FailedToolCalls != 1 {
		t.Errorf("expected 1 failed tool call, got %d", trace.FailedToolCalls)
	}

	// 2. Not found case
	resp404, err := http.Get(ts.URL + "/api/live/sessions/ses_non_existent/trace")
	if err != nil {
		t.Fatalf("missing trace request failed: %v", err)
	}
	defer resp404.Body.Close()

	if resp404.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for missing session, got %d", resp404.StatusCode)
	}

	var errResp adaptershttp.ErrorResponse
	if err := json.NewDecoder(resp404.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed decoding 404 json error: %v", err)
	}
	if errResp.Error != "session_not_found" {
		t.Errorf("expected error code session_not_found, got %s", errResp.Error)
	}

	// 3. Invalid characters / path traversal
	resp400, err := http.Get(ts.URL + "/api/live/sessions/..%2Fevil/trace")
	if err != nil {
		t.Fatalf("traversal request failed: %v", err)
	}
	defer resp400.Body.Close()

	if resp400.StatusCode != http.StatusBadRequest && resp400.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 400 or 404 for path traversal attempt, got %d", resp400.StatusCode)
	}
}

func TestLiveStreamHandler_SSEProtocol(t *testing.T) {
	eventsCh := make(chan live.LiveEvent, 5)
	mock := &mockLiveService{
		tree: domain.LiveSessionTree{
			RootSessions:        []*domain.LiveSessionNode{},
			LastRefreshedAt:     "2026-10-19T10:00:00Z",
			TotalActiveSessions: 0,
		},
		events: eventsCh,
	}

	ts := setupLiveTestServer(mock)
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", ts.URL+"/api/live/stream", nil)
	if err != nil {
		t.Fatalf("creating stream request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("stream connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for stream, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/event-stream") {
		t.Errorf("expected text/event-stream, got %s", contentType)
	}

	reader := bufio.NewReader(resp.Body)

	// Read initial session_tree event
	initialEvent, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("reading event type: %v", err)
	}
	if !strings.Contains(initialEvent, "event: session_tree") {
		t.Errorf("expected event: session_tree, got %s", initialEvent)
	}

	// Read data line
	initialData, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("reading event data: %v", err)
	}
	if !strings.Contains(initialData, "data:") {
		t.Errorf("expected data line, got %s", initialData)
	}

	// Push a live event to the broadcaster
	eventsCh <- live.LiveEvent{
		Type: live.LiveEventTraceUpdate,
		Data: []byte(`{"session_id":"s1"}`),
	}

	// Read streamed event
	for {
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			t.Fatalf("reading streamed event line: %v", readErr)
		}
		if strings.Contains(line, "event: trace_update") {
			break
		}
	}

	// Cancel context and verify clean disconnection
	cancel()
	time.Sleep(50 * time.Millisecond)

	if !mock.isUnsubscribed() {
		t.Error("expected subscriber to unsubscribe upon client disconnection")
	}
}

func TestLiveTraceHandler_AdversarialPathTraversal(t *testing.T) {
	mock := &mockLiveService{
		trace: domain.LiveTraceView{SessionID: "ses_valid"},
	}
	ts := setupLiveTestServer(mock)
	defer ts.Close()

	traversalPayloads := []string{
		"../etc/passwd",
		"..%2Fetc%2Fpasswd",
		"%2e%2e%2froot",
		"%2e%2e/root",
		"..%2froot",
		"something/../../secret",
		"sub/folder",
	}

	for _, payload := range traversalPayloads {
		t.Run("traversal_"+payload, func(t *testing.T) {
			url := ts.URL + "/api/live/sessions/" + payload + "/trace"
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
				t.Errorf("expected 400 or 404 for traversal payload %q, got status %d", payload, resp.StatusCode)
			}

			if resp.StatusCode == http.StatusBadRequest {
				var errResp adaptershttp.ErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
					if errResp.Error != "invalid_session_id" {
						t.Errorf("expected error code invalid_session_id, got %q", errResp.Error)
					}
				}
			}
		})
	}
}

func TestLiveStreamHandler_MultipleConcurrentSubscribers(t *testing.T) {
	eventsCh := make(chan live.LiveEvent, 10)
	mock := &mockLiveService{
		tree: domain.LiveSessionTree{
			RootSessions:        []*domain.LiveSessionNode{},
			LastRefreshedAt:     "2026-10-19T10:00:00Z",
			TotalActiveSessions: 0,
		},
		events: eventsCh,
	}

	ts := setupLiveTestServer(mock)
	defer ts.Close()

	const clientCount = 5
	var wg sync.WaitGroup
	wg.Add(clientCount)

	for i := 0; i < clientCount; i++ {
		go func(clientIdx int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, "GET", ts.URL+"/api/live/stream", nil)
			if err != nil {
				t.Errorf("client %d creating request: %v", clientIdx, err)
				return
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Errorf("client %d stream request: %v", clientIdx, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("client %d expected 200 OK, got %d", clientIdx, resp.StatusCode)
				return
			}

			reader := bufio.NewReader(resp.Body)
			line, err := reader.ReadString('\n')
			if err != nil {
				t.Errorf("client %d reading initial line: %v", clientIdx, err)
				return
			}
			if !strings.Contains(line, "event: session_tree") {
				t.Errorf("client %d expected session_tree event, got %q", clientIdx, line)
			}
		}(i)
	}

	wg.Wait()
}
