package capture

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestFIX19DiscoveryCandidateAllowanceSurvivesSecondRefreshClass(t *testing.T) {
	ctx := context.Background()
	cfg, err := adaptersconfig.LoadFromFile("../config/testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	store, err := Open(t.TempDir(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	}()
	if err := store.SetSourceBudgets(8, 1024*1024); err != nil {
		t.Fatal(err)
	}
	var identities []domain.CaptureIdentity
	var paths []string
	write := func(path, id string, tokens int) {
		t.Helper()
		body := fmt.Sprintf(`{"type":"session_meta","payload":{"id":%q}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"fixture","turn_id":"one"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":%q,"turn_id":"one","usage":{"input_tokens":%d,"output_tokens":2}}}
`, id, id, tokens)
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"first", "second"} {
		home := t.TempDir()
		dir := filepath.Join(home, "sessions")
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, id+".jsonl")
		write(path, id, 10)
		identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: home, SessionID: id}
		row, err := store.Register(ctx, identity, path, "SessionStart")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = store.Bind(ctx, identity, "TEST-100", row.Request); err != nil {
			t.Fatal(err)
		}
		identities = append(identities, identity)
		paths = append(paths, path)
	}
	baseline, err := store.Reconcile(ctx, "TEST-100", "tester", 8, 1024*1024, fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if len(baseline.RefreshedSessions) != 2 {
		t.Fatalf("baseline did not extract actual owners: %+v", baseline)
	}
	rows, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var acceptedSecond []domain.SessionFact
	for _, row := range rows {
		if row.Identity.SessionID == "second" {
			acceptedSecond = row.Snapshot
		}
	}
	for i, path := range paths {
		write(path, identities[i].SessionID, 20)
	}
	if err := store.SetSourceBudgets(1, 1024*1024); err != nil {
		t.Fatal(err)
	}
	budget := &workBudget{sessions: 8, bytes: 1024 * 1024}
	result := ReconcileResult{StartedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	// The scheduler can call this class again after recovery makes another owner
	// eligible. Keep its operation budget and attempted set, while changing only
	// that owner's eligibility between the two real extraction passes.
	attempted := map[string]bool{IdentityKey(identities[1]): true}
	archiveProjection := map[string]domain.CaptureRegistration{}
	_, firstAccepted, err := store.refreshTicketOwners(ctx, "TEST-100", "tester", budget, &result, attempted, archiveProjection, fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if !firstAccepted[IdentityKey(identities[0])] {
		t.Fatal("first one-candidate store was not refreshed")
	}
	delete(attempted, IdentityKey(identities[1]))
	rows, secondAccepted, err := store.refreshTicketOwners(ctx, "TEST-100", "tester", budget, &result, attempted, archiveProjection, fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if secondAccepted[IdentityKey(identities[1])] {
		t.Fatal("second refresh class restarted the operation's exhausted discovery candidate allowance")
	}
	found := false
	for _, row := range rows {
		if row.Identity.SessionID == "second" {
			found = true
			if row.State != domain.CapturePending || !strings.Contains(strings.ToLower(row.Detail), "budget") || !reflect.DeepEqual(row.Snapshot, acceptedSecond) {
				t.Fatalf("exhausted second owner changed accepted facts or lost pending reason: %+v", row)
			}
		}
	}
	if !found {
		t.Fatal("second owner disappeared")
	}
}
