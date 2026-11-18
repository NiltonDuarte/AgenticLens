package capture_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

func TestFIX19ReconcileAccountsExactDiscoveryWithinSharedOperationBudget(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	home := t.TempDir()
	dir := filepath.Join(home, "sessions")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	ids := []string{"a", "b", "c"}
	for _, id := range ids {
		path := filepath.Join(dir, id+".jsonl")
		writeRollout(t, path, id, "", 10)
		registerBound(t, s, home, path, id, "TEST-100")
	}
	baseline, err := s.Reconcile(ctx, "TEST-100", "tester", 8, fixtureByteBudget, fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if len(baseline.RefreshedSessions) != 3 {
		t.Fatalf("baseline did not refresh fixture owners: %+v", baseline)
	}
	prior := baseline.Sessions
	for _, id := range ids {
		writeRollout(t, filepath.Join(dir, id+".jsonl"), id, "", 20)
	}
	padding := []byte(fmt.Sprintf("{\"type\":\"session_meta\",\"payload\":{\"id\":\"unrelated\",\"padding\":%q}}\n", strings.Repeat("x", 700*1024)))
	if err := os.WriteFile(filepath.Join(dir, "z-unrelated.jsonl"), padding, 0600); err != nil {
		t.Fatal(err)
	}
	const allowance int64 = 1024 * 1024
	if err := s.SetSourceBudgets(8, allowance); err != nil {
		t.Fatal(err)
	}
	for repeat := 0; repeat < 2; repeat++ {
		result, err := s.Reconcile(ctx, "TEST-100", "tester", 8, allowance, fixtureGenerationPolicy())
		if err != nil {
			t.Fatal(err)
		}
		// Proving exact uniqueness requires reading the unrelated store candidate as
		// well as registered paths. Byte accounting cannot report only tiny owners.
		if result.SourceBytes < int64(len(padding)) || result.SourceBytes > allowance {
			t.Fatalf("exact discovery missing from shared source accounting: reported=%d unrelated=%d allowance=%d refreshed=%d", result.SourceBytes, len(padding), allowance, len(result.RefreshedSessions))
		}
		complete := map[string]bool{}
		for _, id := range result.SourceEvents.CompleteSessionIDs {
			complete[id] = true
		}
		refreshed := map[string]bool{}
		for _, owner := range result.RefreshedSessions {
			refreshed[owner.ID] = true
		}
		if !reflect.DeepEqual(complete, refreshed) {
			t.Fatalf("complete certification differs from accepted refresh: complete=%v refreshed=%v", complete, refreshed)
		}
		if len(result.Sessions) != len(prior) {
			t.Fatalf("budget exhaustion dropped owners: %+v", result.Sessions)
		}
		for _, before := range prior {
			var got *domain.SessionFact
			for i := range result.Sessions {
				if result.Sessions[i].ID == before.ID {
					got = &result.Sessions[i]
				}
			}
			if got == nil {
				t.Fatalf("missing accepted owner %s", before.ID)
			}
			if !refreshed[before.ID] && !reflect.DeepEqual(*got, before) {
				t.Fatalf("unrefreshed owner %s changed: got=%+v want=%+v", before.ID, *got, before)
			}
			if !refreshed[before.ID] {
				found := false
				for _, row := range result.Registrations {
					if row.Identity.SessionID == before.ID {
						found = row.State == domain.CapturePending && strings.Contains(strings.ToLower(row.Detail), "budget")
					}
				}
				if !found {
					t.Fatalf("unrefreshed source not explicitly pending budget: %+v", result.Registrations)
				}
			}
		}
		prior = result.Sessions
	}
}
