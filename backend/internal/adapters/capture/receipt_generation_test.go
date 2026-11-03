package capture

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

func receiptGenerationStore(t *testing.T) *Store {
	t.Helper()
	s, err := openRegistry(t.TempDir(), "ReceiptFixture", t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	return s
}

func receiptGenerationFacts(name string) AuxiliaryFacts {
	return AuxiliaryFacts{
		Skills:    []domain.SkillFact{{FactProvenance: domain.FactProvenance{SessionID: "owner", Source: domain.SourceCodex, SourceInvocationID: "read"}, Name: name, Outcome: "unavailable"}},
		TaskLinks: []domain.TaskLinkFact{{FactProvenance: domain.FactProvenance{SessionID: "owner", Source: domain.SourceCodex, SourceInvocationID: "spawn"}, Description: name, ChildSession: "unavailable-child", CostEvidence: domain.UnavailableCost()}},
	}
}

func receiptGenerationWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
}

func receiptGenerationLock(t *testing.T, path string) {
	t.Helper()
	lock, err := AcquirePublication(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := lock.Close(); err != nil {
			t.Error(err)
		}
	})
}

func receiptGenerationRecord(t *testing.T, s *Store, ticket, path string) []byte {
	t.Helper()
	var body []byte
	if err := s.db.QueryRow("SELECT record FROM capture_auxiliary_publications WHERE worktree_id=? AND ticket=? AND destination=?", s.worktreeID, ticket, path).Scan(&body); err != nil {
		t.Fatal(err)
	}
	return body
}

func TestReceiptGenerationSelectsOnlySurvivingArtifact(t *testing.T) {
	for _, committed := range []bool{false, true} {
		t.Run(map[bool]string{false: "rollback", true: "committed before finalize"}[committed], func(t *testing.T) {
			s := receiptGenerationStore(t)
			path := filepath.Join(s.project, "ledger.json")
			lock, err := AcquirePublication(context.Background(), path)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := lock.Close(); err != nil {
					t.Error(err)
				}
			})
			old, next := receiptGenerationFacts("old"), receiptGenerationFacts("new")
			receiptGenerationWrite(t, path, "artifact L0")
			// Migration is a trusted-caller boundary. Actual CLI tests separately
			// prove its prerequisite authorization rather than inferring proof
			// from these synthetic adapter artifact bytes.
			active, err := s.MigratePublicationAuxiliary(context.Background(), "TEST-1", path, old)
			if err != nil {
				t.Fatal(err)
			}
			receiptGenerationWrite(t, path+".publication-pending", "caller-owned file journal")
			receiptGenerationWrite(t, path, "artifact L1")
			pending, err := s.StagePublicationAuxiliary(context.Background(), "TEST-1", path, next)
			if err != nil || pending.ID == active.ID {
				t.Fatalf("pending generation not distinct: %+v %v", pending, err)
			}
			if _, _, err = s.SelectPublicationAuxiliary(context.Background(), "TEST-1", path); err == nil {
				t.Fatal("selected proof before validated file recovery")
			}
			if err = s.FinalizePublicationAuxiliary(context.Background(), "TEST-1", path, pending.ID); err == nil {
				t.Fatal("finalized while file journal remains")
			}
			if !committed {
				receiptGenerationWrite(t, path, "artifact L0")
			}
			if err = os.Remove(path + ".publication-pending"); err != nil {
				t.Fatal(err)
			}
			want := old
			if committed {
				want = next
			}
			for pass := 0; pass < 2; pass++ {
				facts, bound, err := s.SelectPublicationAuxiliary(context.Background(), "TEST-1", path)
				if err != nil || !bound || !reflect.DeepEqual(facts, []AuxiliaryFacts{want}) {
					t.Fatalf("recovery exposed wrong or unioned proof: %+v %t %v", facts, bound, err)
				}
			}
			if committed {
				if err = s.FinalizePublicationAuxiliary(context.Background(), "TEST-1", path, pending.ID); err != nil {
					t.Fatalf("finalize not idempotent after startup promotion: %v", err)
				}
			} else if err = s.FinalizePublicationAuxiliary(context.Background(), "TEST-1", path, pending.ID); err == nil {
				t.Fatal("rolled-back pending generation remained authorized")
			}
		})
	}
}

func TestReceiptGenerationEmptyAbsentAndIdenticalArtifacts(t *testing.T) {
	s := receiptGenerationStore(t)
	path := filepath.Join(s.project, "first.json")
	receiptGenerationLock(t, path)
	receiptGenerationWrite(t, path, "first unfinished artifact")
	receiptGenerationWrite(t, path+".publication-pending", "caller-owned file journal")
	if _, err := s.StagePublicationAuxiliary(context.Background(), "TEST-1", path, AuxiliaryFacts{}); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path + ".publication-pending"); err != nil {
		t.Fatal(err)
	}
	for pass := 0; pass < 2; pass++ {
		facts, bound, err := s.SelectPublicationAuxiliary(context.Background(), "TEST-1", path)
		if err != nil || !bound || len(facts) != 0 {
			t.Fatalf("first rollback invented an accepted empty ledger: %+v %t %v", facts, bound, err)
		}
	}
	receiptGenerationWrite(t, path, "accepted empty-fact artifact")
	receiptGenerationWrite(t, path+".publication-pending", "caller-owned file journal")
	first, err := s.StagePublicationAuxiliary(context.Background(), "TEST-1", path, AuxiliaryFacts{})
	if err != nil {
		t.Fatal(err)
	}
	again, err := s.StagePublicationAuxiliary(context.Background(), "TEST-1", path, AuxiliaryFacts{})
	if err != nil || first.ID != again.ID {
		t.Fatalf("identical pending artifact not stable: %+v %+v %v", first, again, err)
	}
	if _, err = s.StagePublicationAuxiliary(context.Background(), "TEST-1", path, receiptGenerationFacts("fabricated")); err == nil {
		t.Fatal("identical artifact accepted different proof")
	}
	if err = os.Remove(path + ".publication-pending"); err != nil {
		t.Fatal(err)
	}
	if err = s.FinalizePublicationAuxiliary(context.Background(), "TEST-1", path, first.ID); err != nil {
		t.Fatal(err)
	}
	facts, bound, err := s.SelectPublicationAuxiliary(context.Background(), "TEST-1", path)
	if err != nil || !bound || !reflect.DeepEqual(facts, []AuxiliaryFacts{{}}) {
		t.Fatalf("empty accepted generation lost: %+v %t %v", facts, bound, err)
	}
}

func TestReceiptGenerationScopesAndLegacyFallback(t *testing.T) {
	s := receiptGenerationStore(t)
	ctx := context.Background()
	paths := []string{filepath.Join(s.project, "one.json"), filepath.Join(s.project, "two.json")}
	var generations []AuxiliaryGeneration
	for i, path := range paths {
		receiptGenerationLock(t, path)
		receiptGenerationWrite(t, path, "same artifact bytes")
		generation, err := s.MigratePublicationAuxiliary(ctx, "TEST-1", path, receiptGenerationFacts(path))
		if err != nil {
			t.Fatal(err)
		}
		generations = append(generations, generation)
		if i == 1 {
			facts, _, err := s.SelectPublicationAuxiliary(ctx, "TEST-1", paths[0])
			if err != nil || !reflect.DeepEqual(facts, []AuxiliaryFacts{receiptGenerationFacts(paths[0])}) {
				t.Fatal("second destination replaced first destination proof")
			}
		}
	}
	before := receiptGenerationRecord(t, s, "TEST-1", paths[1])
	if err := s.FinalizePublicationAuxiliary(ctx, "TEST-1", paths[1], generations[0].ID); err == nil {
		t.Fatal("generation identity crossed destination scope")
	}
	if got := receiptGenerationRecord(t, s, "TEST-1", paths[1]); !bytes.Equal(got, before) {
		t.Fatal("wrong generation finalization mutated another destination")
	}
	receiptGenerationWrite(t, paths[0]+".publication-pending", "caller-owned file journal")
	receiptGenerationWrite(t, paths[0], "new first destination artifact")
	if _, err := s.StagePublicationAuxiliary(ctx, "TEST-1", paths[0], receiptGenerationFacts("new")); err != nil {
		t.Fatal(err)
	}
	if got := receiptGenerationRecord(t, s, "TEST-1", paths[1]); !bytes.Equal(got, before) {
		t.Fatal("staging first destination changed second destination metadata")
	}
	if err := s.RecordAuxiliaryReceipt(ctx, "TEST-1", receiptGenerationFacts("legacy")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(paths[0] + ".publication-pending"); err != nil {
		t.Fatal(err)
	}
	receiptGenerationWrite(t, paths[0], "neither trusted generation")
	if _, bound, err := s.SelectPublicationAuxiliary(ctx, "TEST-1", paths[0]); err == nil || !bound {
		t.Fatal("bound digest mismatch fell back to ticket-wide legacy proof")
	}
	for _, wrong := range []string{"TEST-2", " TEST-1 "} {
		facts, bound, err := s.SelectPublicationAuxiliary(ctx, wrong, paths[1])
		if wrong == "TEST-2" && (err != nil || bound || len(facts) != 0) {
			t.Fatalf("receipt crossed ticket scope: %+v %t %v", facts, bound, err)
		}
		if wrong != "TEST-2" && err == nil {
			t.Fatal("noncanonical ticket alias accepted")
		}
	}
	peer, err := openRegistry(s.directory, "ReceiptFixture", t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := peer.Close(); err != nil {
			t.Error(err)
		}
	}()
	peerPath := filepath.Join(peer.project, "one.json")
	receiptGenerationLock(t, peerPath)
	receiptGenerationWrite(t, peerPath, "same artifact bytes")
	if facts, bound, err := peer.SelectPublicationAuxiliary(ctx, "TEST-1", peerPath); err != nil || bound || len(facts) != 0 {
		t.Fatalf("receipt crossed worktree scope: %+v %t %v", facts, bound, err)
	}
	if _, _, err := peer.SelectPublicationAuxiliary(ctx, "TEST-1", paths[1]); err == nil {
		t.Fatal("foreign physical destination accepted")
	}
}

func TestReceiptGenerationMalformedMetadataFailsClosed(t *testing.T) {
	for _, edit := range []string{"unknown version", "missing pending", "duplicate version", "unknown field", "invalid generation", "malformed JSON"} {
		t.Run(edit, func(t *testing.T) {
			s := receiptGenerationStore(t)
			path := filepath.Join(s.project, "ledger.json")
			receiptGenerationLock(t, path)
			receiptGenerationWrite(t, path, "accepted artifact")
			generation, err := s.MigratePublicationAuxiliary(context.Background(), "TEST-1", path, receiptGenerationFacts("old"))
			if err != nil {
				t.Fatal(err)
			}
			body := receiptGenerationRecord(t, s, "TEST-1", path)
			switch edit {
			case "unknown version":
				body = bytes.Replace(body, []byte(`"version":1`), []byte(`"version":999`), 1)
			case "missing pending":
				body = bytes.Replace(body, []byte(`,"pending":null`), nil, 1)
			case "duplicate version":
				body = bytes.Replace(body, []byte(`"version":1`), []byte(`"version":1,"version":1`), 1)
			case "unknown field":
				body = append([]byte(`{"unexpected":true,`), body[1:]...)
			case "invalid generation":
				body = bytes.Replace(body, []byte(generation.ID), []byte("not-a-generation"), 1)
			case "malformed JSON":
				body = []byte("{")
			}
			if _, err = s.db.Exec("UPDATE capture_auxiliary_publications SET record=? WHERE worktree_id=? AND ticket=? AND destination=?", body, s.worktreeID, "TEST-1", path); err != nil {
				t.Fatal(err)
			}
			if err = s.RecordAuxiliaryReceipt(context.Background(), "TEST-1", receiptGenerationFacts("old")); err != nil {
				t.Fatal(err)
			}
			if _, _, err = s.SelectPublicationAuxiliary(context.Background(), "TEST-1", path); err == nil {
				t.Fatal("invalid bound metadata silently used legacy proof")
			}
			if _, err = s.MigratePublicationAuxiliary(context.Background(), "TEST-1", path, receiptGenerationFacts("old")); err == nil {
				t.Fatal("migration overwrote invalid bound metadata")
			}
			if err = s.FinalizePublicationAuxiliary(context.Background(), "TEST-1", path, generation.ID); err == nil {
				t.Fatal("finalization overwrote invalid metadata")
			}
			receiptGenerationWrite(t, path+".publication-pending", "caller-owned file journal")
			if _, err = s.StagePublicationAuxiliary(context.Background(), "TEST-1", path, receiptGenerationFacts("new")); err == nil {
				t.Fatal("staging overwrote invalid metadata")
			}
			if err = os.Remove(path + ".publication-pending"); err != nil {
				t.Fatal(err)
			}
			if got := receiptGenerationRecord(t, s, "TEST-1", path); !bytes.Equal(got, body) {
				t.Fatal("metadata rejection mutated retained proof")
			}
		})
	}
}

func TestReceiptGenerationRespectsRelocationFence(t *testing.T) {
	s := receiptGenerationStore(t)
	path := filepath.Join(s.project, "ledger.json")
	receiptGenerationLock(t, path)
	receiptGenerationWrite(t, path, "accepted artifact")
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = reserveRelocationScopeTx(tx, s.worktreeID, "receipt-fence"); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if _, err = s.MigratePublicationAuxiliary(context.Background(), "TEST-1", path, receiptGenerationFacts("old")); err == nil || !strings.Contains(err.Error(), "relocation") {
		t.Fatalf("receipt migration bypassed pending relocation fence: %v", err)
	}
	receiptGenerationWrite(t, path+".publication-pending", "caller-owned file journal")
	if _, err = s.StagePublicationAuxiliary(context.Background(), "TEST-1", path, receiptGenerationFacts("new")); err == nil || !strings.Contains(err.Error(), "relocation") {
		t.Fatalf("receipt stage bypassed pending relocation fence: %v", err)
	}
	var count int
	if err = s.db.QueryRow("SELECT count(*) FROM capture_auxiliary_publications WHERE worktree_id=?", s.worktreeID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("fenced operation left receipt state: %d %v", count, err)
	}
}
