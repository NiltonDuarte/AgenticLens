package live_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/adapters/live"
	"agenticlens/internal/domain"
)

// This fixture is also consumed by the real frontend API validator and view.
// Its only normalized field is the source-independent HTTP refresh timestamp.
func TestNativeLiveHTTPWireMatchesFrontendFixture(t *testing.T) {
	path := createTestDBFile(t)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	const stamp int64 = 4102444800000
	if _, err := db.Exec(`INSERT INTO session(id,agent,model,title,time_created,time_updated) VALUES('wire-root','builder','wire-model','T-02 wire projection',?,?);
 INSERT INTO message(id,session_id,time_created,time_updated,data) VALUES('wire-message','wire-root',?,?,'{"role":"assistant","agent":"builder","model":{"modelID":"wire-model"},"tokens":{"input":17,"output":3}}');`, stamp, stamp, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	hook, err := live.OpenReadOnlyHook(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	cfg := testProjectConfig()
	cfg.Ticket = domain.TicketConfig{Prefixes: []domain.TicketPrefix{{Prefix: "T-", Description: "Wire fixture"}}, CaseInsensitive: false}
	cfg.RoleMappings = domain.RoleMappings{"builder": "planner", "planner": "other"}
	service := live.NewLiveService(hook, cfg)
	defer service.Close()
	recorder := httptest.NewRecorder()
	adaptershttp.LiveSessionsHandler(service)(recorder, httptest.NewRequest(http.MethodGet, "/api/live/sessions", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("live HTTP: %d %s", recorder.Code, recorder.Body.String())
	}
	var tree domain.LiveSessionTree
	if err := json.Unmarshal(recorder.Body.Bytes(), &tree); err != nil {
		t.Fatal(err)
	}
	if len(tree.RootSessions) != 1 {
		t.Fatalf("native root missing: %+v", tree)
	}
	root := tree.RootSessions[0]
	if root.Agent != "builder" || root.Role != domain.AgentRole("planner") || root.Model != "wire-model" || root.GuessedTicketID == nil || *root.GuessedTicketID != "T-02" {
		t.Fatalf("source persona must map once, preserving identity: %+v", root)
	}
	if tree.LastRefreshedAt == "" {
		t.Fatal("available source has no refresh observation")
	}
	actual := bytes.Replace(recorder.Body.Bytes(), []byte(tree.LastRefreshedAt), []byte("2100-01-01T00:00:00.000Z"), 1)
	var actualCompact bytes.Buffer
	if err := json.Compact(&actualCompact, actual); err != nil {
		t.Fatal(err)
	}
	fixturePath := filepath.Join("..", "..", "..", "..", "web", "src", "test", "fixtures", "live-session-wire.json")
	expected, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("shared frontend fixture: %v; producer JSON=%s", err, actualCompact.Bytes())
	}
	var expectedCompact bytes.Buffer
	if err := json.Compact(&expectedCompact, expected); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actualCompact.Bytes(), expectedCompact.Bytes()) {
		t.Fatalf("native HTTP differs from shared frontend fixture: actual=%s expected=%s", actualCompact.Bytes(), expectedCompact.Bytes())
	}
}

func TestLiveUnavailableDiffersFromAvailableEmptySource(t *testing.T) {
	unavailable := live.NewLiveService(nil, testProjectConfig())
	defer unavailable.Close()
	path := createTestDBFile(t)
	hook, err := live.OpenReadOnlyHook(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	available := live.NewLiveService(hook, testProjectConfig())
	defer available.Close()
	for _, tc := range []struct {
		name    string
		service *live.LiveService
		status  int
	}{
		{"unavailable", unavailable, http.StatusServiceUnavailable},
		{"available-empty", available, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			adaptershttp.LiveSessionsHandler(tc.service)(recorder, httptest.NewRequest(http.MethodGet, "/api/live/sessions", nil))
			if recorder.Code != tc.status {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			if tc.status == http.StatusOK {
				var tree domain.LiveSessionTree
				if err := json.Unmarshal(recorder.Body.Bytes(), &tree); err != nil {
					t.Fatal(err)
				}
				if tree.TotalActiveSessions != 0 || len(tree.RootSessions) != 0 || tree.LastRefreshedAt == "" {
					t.Fatalf("available empty snapshot invalid: %+v", tree)
				}
			} else if !bytes.Contains(recorder.Body.Bytes(), []byte("source_unavailable")) {
				t.Fatal(recorder.Body.String())
			}
		})
	}
	recorder := httptest.NewRecorder()
	adaptershttp.LiveStreamHandler(unavailable)(recorder, httptest.NewRequest(http.MethodGet, "/api/live/stream", nil))
	if recorder.Code != http.StatusServiceUnavailable || recorder.Header().Get("Content-Type") == "text/event-stream" {
		t.Fatalf("unavailable source advertised connected SSE: %d %s", recorder.Code, recorder.Body.String())
	}
}
