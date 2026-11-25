package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"agenticlens/internal/adapters"
	adaptershttp "agenticlens/internal/adapters/http"
)

func TestHistoricalProgressionWorkspaceContract(t *testing.T) {
	root := httpWorkspaceRoot(t)
	fixtures := []struct {
		id, kind, completed string
		invocations         []string
	}{
		{"AL-41", "feat", "2000-01-08T12:00:00Z", []string{
			"1999-12-31T23:59:59Z", "2000-01-01T00:00:00Z", "2000-01-08T23:59:59Z",
		}},
		{"AL-42", "fix", "2000-02-15T12:00:00Z", []string{
			"2000-02-07T23:59:59Z", "2000-02-15T16:00:00Z",
		}},
	}
	for _, fixture := range fixtures {
		dir := filepath.Join(root, fixture.id)
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
		var ledger adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(sampleTestLedgerJSON(fixture.id, fixture.id, 60, 1, 1)), &ledger); err != nil {
			t.Fatal(err)
		}
		ledger.TicketType = &fixture.kind
		ledger.StartedAt = &fixture.completed
		ledger.CompletedAt = &fixture.completed
		ledger.UpdatedAt = fixture.completed
		session := &ledger.Sessions[0]
		seed := session.Invocations[0]
		session.Invocations = nil
		for _, date := range fixture.invocations {
			invocation := seed
			invocation.CompletedAt = &date
			session.Invocations = append(session.Invocations, invocation)
		}
		raw, err := json.Marshal(ledger)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "agentic_metrics.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}

	handler := adaptershttp.WorkspacesHandler(root)
	cases := []struct {
		query string
		ids   []string
	}{
		{"", []string{"AL-41", "AL-42"}},
		{"?prefix=AL", []string{"AL-41", "AL-42"}},
		{"?ticket_type=feature&prefix=AL", []string{"AL-41"}},
		{"?window=7d&prefix=AL", []string{}},
	}
	var baseline map[string]map[string]json.RawMessage
	for _, tc := range cases {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/workspaces"+tc.query, nil))
		if recorder.Code != http.StatusOK || recorder.Header().Get("X-AgenticLens-Workspace-Diagnostics") != "" {
			t.Fatalf("query %q: status=%d diagnostics=%q body=%s", tc.query, recorder.Code, recorder.Header().Get("X-AgenticLens-Workspace-Diagnostics"), recorder.Body.String())
		}
		var rows []map[string]json.RawMessage
		if err := json.Unmarshal(recorder.Body.Bytes(), &rows); err != nil {
			t.Fatal(err)
		}
		ids := make([]string, 0, len(rows))
		byID := make(map[string]map[string]json.RawMessage, len(rows))
		for _, row := range rows {
			var id string
			if err := json.Unmarshal(row["ticket_id"], &id); err != nil {
				t.Fatal(err)
			}
			ids = append(ids, id)
			// Substitute only the fixture's temporary workspace path.
			row["workspace_path"] = json.RawMessage(`"workspace/` + id + `"`)
			byID[id] = row
		}
		slices.Sort(ids)
		if !slices.Equal(ids, tc.ids) {
			t.Fatalf("query %q: ids=%v want %v", tc.query, ids, tc.ids)
		}
		if tc.query == "" {
			baseline = byID
			continue
		}
		for _, id := range ids {
			if !reflect.DeepEqual(byID[id], baseline[id]) {
				t.Fatalf("query %q changed the history of %s", tc.query, id)
			}
		}
	}

	path := filepath.Join("..", "..", "..", "..", "web", "src", "test", "fixtures", "progression-workspaces-wire.json")
	raw, err := os.ReadFile(path)
	var expected map[string]map[string]json.RawMessage
	if err == nil {
		err = json.Unmarshal(raw, &expected)
	}
	if err != nil || !reflect.DeepEqual(expected, baseline) {
		wire, marshalErr := json.Marshal(baseline)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		t.Fatalf("shared workspace wire fixture differs from HTTP response: %v\nWIRE=%s", err, wire)
	}
}
