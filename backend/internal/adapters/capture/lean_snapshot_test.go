package capture_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

func TestRegistryRetiredCachedSnapshotsRecoverFromUnchangedNativeSource(t *testing.T) {
	for _, tc := range []struct{ field, value string }{
		{"events", "[]"}, {"events_are_partial", "false"}, {"task_links", `[{"subagent_type":"tester","timestamp":"2026-01-01T00:00:01Z","child_reference":"opaque-cache-only"}]`}, {"agent_path", `"/root/cache-only"`}, {"skills", `[{"name":"cache-only","caller_role":"coder","timestamp":"2026-01-01T00:00:01Z","success":true}]`},
	} {
		t.Run(tc.field, func(t *testing.T) {
			ctx := context.Background()
			dir := t.TempDir()
			store := openStore(t, dir)
			sourceHome := t.TempDir()
			if err := os.MkdirAll(filepath.Join(sourceHome, "sessions"), 0700); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(sourceHome, "sessions", "owner.jsonl")
			native := writeRollout(t, path, "owner", "", 10)
			native = append(native, []byte(`{"timestamp":"2026-01-01T00:00:03Z","type":"response_item","payload":{"type":"function_call","call_id":"native-call","name":"load_skill","arguments":"{\"name\":\"native-skill\"}"}}`+"\n")...)
			if err := os.WriteFile(path, native, 0600); err != nil {
				t.Fatal(err)
			}
			bound := registerBound(t, store, sourceHome, path, "owner", "TEST-100")
			initial, err := store.Reconcile(ctx, "TEST-100", "tester", 20, 1000000, fixtureGenerationPolicy())
			if err != nil {
				t.Fatal(err)
			}
			if len(initial.Registrations) != 1 || len(initial.Sessions) != 1 || len(initial.Sessions[0].Skills) != 1 {
				t.Fatalf("native current control did not capture: %+v", initial)
			}
			accepted := initial.Registrations[0]
			db, err := sql.Open("sqlite", filepath.Join(dir, "registry.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			readRow := func() []byte {
				t.Helper()
				var raw []byte
				if err := db.QueryRow(`SELECT record FROM registrations WHERE session_id='owner'`).Scan(&raw); err != nil {
					t.Fatal(err)
				}
				return raw
			}
			var wire map[string]json.RawMessage
			if err = json.Unmarshal(readRow(), &wire); err != nil {
				t.Fatal(err)
			}
			var snapshot []map[string]json.RawMessage
			if err = json.Unmarshal(wire["snapshot"], &snapshot); err != nil {
				t.Fatal(err)
			}
			snapshot[0][tc.field] = json.RawMessage(tc.value)
			wire["snapshot"], err = json.Marshal(snapshot)
			if err != nil {
				t.Fatal(err)
			}
			opaque := append([]byte(nil), wire["snapshot"]...)
			bad, err := json.Marshal(wire)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = db.Exec(`UPDATE registrations SET record=? WHERE session_id='owner'`, bad); err != nil {
				t.Fatal(err)
			}
			rows, err := store.List(ctx)
			if err != nil {
				t.Fatalf("metadata discovery blocked native recovery: %v", err)
			}
			if len(rows) != 1 || !rows[0].SnapshotNeedsRefresh || len(rows[0].Snapshot) != 0 || rows[0].TicketID != bound.TicketID || rows[0].Request != bound.Request || rows[0].Identity != bound.Identity || rows[0].SourceFingerprint != accepted.SourceFingerprint || rows[0].UpdatedAt != accepted.UpdatedAt || !strings.Contains(rows[0].Detail, "re-extract") {
				t.Fatalf("old body became current facts or lost binding/checkpoint: %+v", rows)
			}
			if string(readRow()) != string(bad) {
				t.Fatal("metadata read rewrote opaque snapshot")
			}
			otherPath := filepath.Join(sourceHome, "sessions", "other.jsonl")
			writeRollout(t, otherPath, "other", "", 7)
			registerBound(t, store, sourceHome, otherPath, "other", "TEST-200")
			other, err := store.Reconcile(ctx, "TEST-200", "tester", 20, 1000000, fixtureGenerationPolicy())
			if err != nil {
				t.Fatalf("old unrelated snapshot blocked independent ticket: %v", err)
			}
			if len(other.Sessions) != 1 || other.Sessions[0].ID != "other" {
				t.Fatalf("old owner contaminated another ticket: %+v", other.Sessions)
			}
			if string(readRow()) != string(bad) {
				t.Fatal("unrelated ticket rewrote old snapshot row")
			}
			registered, err := store.Register(ctx, bound.Identity, path, "UserPromptSubmit")
			if err != nil {
				t.Fatalf("fresh callback could not reach recovery: %v", err)
			}
			if registered.TicketID != bound.TicketID || registered.Request != bound.Request || !registered.SnapshotNeedsRefresh || len(registered.Snapshot) != 0 {
				t.Fatalf("callback erased binding or silently accepted old facts: %+v", registered)
			}
			var afterCallback map[string]json.RawMessage
			if err = json.Unmarshal(readRow(), &afterCallback); err != nil {
				t.Fatal(err)
			}
			if string(afterCallback["snapshot"]) != string(opaque) {
				t.Fatal("callback discarded old opaque snapshot before source recovery")
			}
			recovered, err := store.Reconcile(ctx, "TEST-100", "tester", 20, 1000000, fixtureGenerationPolicy())
			if err != nil {
				t.Fatal(err)
			}
			if len(recovered.RefreshedSessions) != 1 || len(recovered.Sessions) != 1 || !reflect.DeepEqual(recovered.Sessions[0].Skills, initial.Sessions[0].Skills) || len(recovered.Registrations) != 1 || recovered.Registrations[0].SnapshotNeedsRefresh || recovered.Registrations[0].State != domain.CaptureCurrent || recovered.Registrations[0].UpdatedAt != accepted.UpdatedAt || recovered.Registrations[0].SourceFingerprint != accepted.SourceFingerprint {
				t.Fatalf("unchanged complete native source could not replace old cache: %+v", recovered)
			}
			var final map[string]json.RawMessage
			if err = json.Unmarshal(readRow(), &final); err != nil {
				t.Fatal(err)
			}
			for _, retired := range []string{`"events"`, `"events_are_partial"`, `"task_links"`, `"agent_path"`, `"timestamp"`, `"success"`, `opaque-cache-only`, `cache-only`} {
				if strings.Contains(string(final["snapshot"]), retired) {
					t.Fatalf("retired cached payload republished: %s", final["snapshot"])
				}
			}
			sourceAfter, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(sourceAfter) != string(native) {
				t.Fatal("cache repair rewrote native source")
			}
		})
	}
}
