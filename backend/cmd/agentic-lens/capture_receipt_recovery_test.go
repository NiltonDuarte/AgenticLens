package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

// Receipt recovery uses actual source extraction and publication for L0. L1 is
// extracted from a newer source, then staged without ever successfully committing
// its receipt: this preserves both real crash windows rather than undoing a
// completed publication to simulate one.
type receiptRecoveryFixture struct {
	f  captureGenerationFixture
	db string
}

func newReceiptRecoveryFixture(t *testing.T) receiptRecoveryFixture {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	r := receiptRecoveryFixture{f: newCaptureGenerationFixture(t), db: setupIsolatedEmptyDB(t)}
	db, err := sql.Open("sqlite", r.db)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO session(id,agent,model,time_created,time_updated)VALUES('receipt-root','coder','model',1760000000000,1760000001000);
INSERT INTO message(id,session_id,time_created,time_updated,data)VALUES('receipt-message','receipt-root',1760000000100,1760000000800,'{"role":"assistant","modelID":"model","time":{"created":1760000000100,"completed":1760000000800},"tokens":{"input":12,"output":3}}');
INSERT INTO part(id,session_id,message_id,time_created,time_updated,data)VALUES
('receipt-skill','receipt-root','receipt-message',1760000000200,1760000000300,'{"type":"tool","tool":"skill","state":{"status":"completed","input":{"name":"receipt-skill"}}}'),
('receipt-task','receipt-root','receipt-message',1760000000400,1760000000500,'{"type":"tool","tool":"task","state":{"status":"completed","input":{"subagent_type":"tester","description":"initial source work"},"output":"{\"session_id\":\"unavailable-child\"}"}}');`)
	if closeErr := db.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err != nil {
		t.Fatal(err)
	}
	setOpenCodeFixtureProject(t, r.db, r.f.project)
	return r
}

func (r receiptRecoveryFixture) run(t *testing.T, extra ...string) ([]byte, error) {
	t.Helper()
	args := r.f.args(t, "--source", "opencode", "--db-path", r.db)
	cmd := exec.Command(cliBinaryPath, append(args, extra...)...)
	cmd.Dir = r.f.project
	return cmd.CombinedOutput()
}

func (r receiptRecoveryFixture) open(t *testing.T) *capture.Store {
	t.Helper()
	s, err := capture.OpenProject(r.f.state, r.f.project)
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

func receiptFacts(l domain.AgenticFactLedger) capture.AuxiliaryFacts {
	return capture.AuxiliaryFacts{Skills: l.SkillFacts(), TaskLinks: l.TaskLinkFacts()}
}

func receiptLegacy(l domain.AgenticFactLedger, transitional bool) domain.AgenticFactLedger {
	l.Skills, l.TaskLinks = l.SkillFacts(), l.TaskLinkFacts()
	l.Sessions = append([]domain.SessionFact(nil), l.Sessions...)
	for i := range l.Sessions {
		l.Sessions[i].Skills, l.Sessions[i].TaskLinks = nil, nil
		l.Sessions[i].AuxiliaryExtractionVersion = 0
	}
	for i := range l.Skills {
		owner := l.Skills[i].SessionID
		l.Skills[i].FactProvenance = domain.FactProvenance{}
		l.Skills[i].Outcome, l.Skills[i].Evidence = "", ""
		if transitional {
			l.Skills[i].SessionID = owner
		}
	}
	for i := range l.TaskLinks {
		owner := l.TaskLinks[i].SessionID
		l.TaskLinks[i].FactProvenance = domain.FactProvenance{}
		if transitional {
			l.TaskLinks[i].SessionID = owner
		}
	}
	version := 5
	l.SchemaVersion = &version
	return l
}

func (r receiptRecoveryFixture) initial(t *testing.T, representation string) domain.AgenticFactLedger {
	t.Helper()
	seed := r.f
	if representation != "canonical" {
		seed.output = filepath.Join(r.f.project, "source-seed.json")
	}
	first := seed.generate(t, "--source", "opencode", "--session-id", "receipt-root", "--db-path", r.db)
	if len(first.SkillFacts()) != 1 || len(first.TaskLinkFacts()) != 1 {
		t.Fatal("source-backed positive control lacks skill/delegation evidence")
	}
	if representation == "canonical" {
		return first
	}
	legacy := receiptLegacy(first, representation == "transitional")
	wire := adapters.AgenticFactLedgerDTOFromDomain(legacy)
	// Preserve the historical root representation while using the real wire
	// codec for nullable costs. The current writer intentionally normalizes a
	// transitional known owner, so reverse only that structural projection here.
	wire.Skills, wire.TaskLinks = legacy.Skills, adapters.TaskLinkFactDTOsFromDomain(legacy.TaskLinks)
	for i := range wire.Sessions {
		wire.Sessions[i].Skills, wire.Sessions[i].TaskLinks = nil, nil
	}
	body, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(r.f.output, body, 0600); err != nil {
		t.Fatal(err)
	}
	// This models the supported historical writer using facts derived from the
	// real source. It establishes no historical destination association.
	if err = r.open(t).RecordAuxiliaryReceipt(context.Background(), "TEST-100", receiptFacts(legacy)); err != nil {
		t.Fatal(err)
	}
	source := readCaptureFile(t, r.db)
	if err = os.Remove(r.db); err != nil {
		t.Fatal(err)
	}
	if out, err := r.run(t); err != nil {
		t.Fatalf("independently proved legacy migration failed: %v %s", err, out)
	}
	if err = os.WriteFile(r.db, source, 0600); err != nil {
		t.Fatal(err)
	}
	return r.f.ledger(t)
}

func (r receiptRecoveryFixture) next(t *testing.T, prior domain.AgenticFactLedger, representation string, empty bool) domain.AgenticFactLedger {
	t.Helper()
	db, err := sql.Open("sqlite", r.db)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`UPDATE session SET time_updated=1760000011000 WHERE id='receipt-root';
UPDATE message SET time_updated=1760000010800,data=replace(data,'1760000000800','1760000010800') WHERE id='receipt-message';
UPDATE part SET time_updated=1760000010300,data='{"type":"tool","tool":"skill","state":{"status":"error","input":{"name":"receipt-skill"},"error":"recorded failure"}}' WHERE id='receipt-skill';
UPDATE part SET time_updated=1760000010500,data=replace(data,'initial source work','new source work') WHERE id='receipt-task';`)
	if err == nil && empty {
		_, err = db.Exec(`DELETE FROM part WHERE session_id='receipt-root'`)
	}
	if closeErr := db.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err != nil {
		t.Fatal(err)
	}
	source, err := opencode.OpenDB(r.db)
	if err != nil {
		t.Fatal(err)
	}
	owners, _, _, err := opencode.CaptureSessionSnapshot(context.Background(), source.DB(), "receipt-root", prior.Sessions[0].User, 1<<20)
	if closeErr := source.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err != nil {
		t.Fatal(err)
	}
	prior.Sessions, prior.Skills, prior.TaskLinks = owners, nil, nil
	if representation != "canonical" {
		prior = receiptLegacy(prior, representation == "transitional")
	}
	if !empty && (len(prior.SkillFacts()) != 1 || prior.SkillFacts()[0].Succeeded() || prior.TaskLinkFacts()[0].Description != "new source work") {
		t.Fatal("new source control did not change exact old auxiliary proof")
	}
	if empty && len(prior.SkillFacts())+len(prior.TaskLinkFacts()) != 0 {
		t.Fatal("authoritative empty source retained old facts")
	}
	return prior
}

func TestCaptureReceiptRecoveryPreservesBothCommitWindows(t *testing.T) {
	for _, representation := range []string{"canonical", "legacy"} {
		for _, committed := range []bool{false, true} {
			for _, empty := range []bool{false, true} {
				name := fmt.Sprintf("%s/committed=%t/empty=%t", representation, committed, empty)
				t.Run(name, func(t *testing.T) {
					r := newReceiptRecoveryFixture(t)
					old := r.initial(t, representation)
					oldBytes := readCaptureFile(t, r.f.output)
					next := r.next(t, old, representation, empty)
					s := r.open(t)
					lock, err := capture.AcquirePublication(context.Background(), r.f.output)
					if err != nil {
						t.Fatal(err)
					}
					if err = adapters.BeginPublication(r.f.output, []string{r.f.output}); err != nil {
						t.Fatal(err)
					}
					if err = adapters.SaveFactLedger(r.f.output, next); err != nil {
						t.Fatal(err)
					}
					pending, err := s.StagePublicationAuxiliary(context.Background(), "TEST-100", r.f.output, receiptFacts(next))
					if err != nil || pending.ID == "" {
						t.Fatalf("stage true pending generation: %+v %v", pending, err)
					}
					if committed {
						if err = adapters.CommitPublication(r.f.output, []string{r.f.output}); err != nil {
							t.Fatal(err)
						}
					}
					if err = lock.Close(); err != nil {
						t.Fatal(err)
					}
					if err = os.Remove(r.db); err != nil {
						t.Fatal(err)
					}
					want := next
					if !committed {
						// Source resolution deliberately fails after real CLI file
						// recovery, allowing an exact byte rollback oracle before
						// a subsequent valid publication updates metadata.
						out, err := r.run(t, "--source", "unsupported")
						if err == nil || !strings.Contains(string(out), "Recovered interrupted publication:") || !bytes.Equal(oldBytes, readCaptureFile(t, r.f.output)) {
							t.Fatalf("validated rollback did not restore exact L0: %v %s", err, out)
						}
						want = old
					}
					for pass := 0; pass < 2; pass++ {
						if out, err := r.run(t); err != nil {
							t.Fatalf("source-unavailable recovery pass %d: %v %s", pass, err, out)
						}
						if got := receiptFacts(r.f.ledger(t)); !reflect.DeepEqual(got, receiptFacts(want)) {
							t.Fatalf("wrong generation proof/occurrence facts after recovery: %+v want %+v", got, receiptFacts(want))
						}
					}
					if _, err := os.Stat(r.f.output + ".publication-pending"); !os.IsNotExist(err) {
						t.Fatalf("recovery left marker: %v", err)
					}
				})
			}
		}
	}
}

func TestCaptureReceiptMigrationRequiresExistingProof(t *testing.T) {
	for _, representation := range []string{"legacy", "transitional"} {
		t.Run(representation, func(t *testing.T) {
			r := newReceiptRecoveryFixture(t)
			initial := r.initial(t, representation)
			if err := os.Remove(r.db); err != nil {
				t.Fatal(err)
			}
			if out, err := r.run(t); err != nil {
				t.Fatalf("normalized trusted receipt lost source-unavailable proof: %v %s", err, out)
			}
			if !reflect.DeepEqual(receiptFacts(initial), receiptFacts(r.f.ledger(t))) {
				t.Fatal("metadata migration changed accepted source facts")
			}
		})
	}
	t.Run("already lost old proof stays unavailable", func(t *testing.T) {
		r := newReceiptRecoveryFixture(t)
		old := r.initial(t, "canonical")
		oldBytes := readCaptureFile(t, r.f.output)
		next := r.next(t, old, "canonical", false)
		other := filepath.Join(r.f.project, "unproved.json")
		if err := os.WriteFile(other, oldBytes, 0600); err != nil {
			t.Fatal(err)
		}
		// A legacy writer has already overwritten R0. No source/snapshot or
		// committed proof of old facts remains for this unbound destination.
		if err := r.open(t).RecordAuxiliaryReceipt(context.Background(), "TEST-100", receiptFacts(next)); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(r.db); err != nil {
			t.Fatal(err)
		}
		r.f.output = other
		if out, err := r.run(t); err == nil || !strings.Contains(string(out), "Baseline authorization:") {
			t.Fatalf("untrusted old ledger fabricated missing R0: %v %s", err, out)
		}
		if !bytes.Equal(oldBytes, readCaptureFile(t, other)) {
			t.Fatal("failed metadata migration changed stored artifact")
		}
		lock, err := capture.AcquirePublication(context.Background(), other)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := lock.Close(); err != nil {
				t.Error(err)
			}
		}()
		if facts, bound, err := r.open(t).SelectPublicationAuxiliary(context.Background(), "TEST-100", other); err != nil || bound || len(facts) == 0 {
			t.Fatalf("failed authorization invented destination metadata: %+v %t %v", facts, bound, err)
		}
	})
}

func TestCaptureReceiptFirstPublicationRollbackPreservesAbsence(t *testing.T) {
	r := newReceiptRecoveryFixture(t)
	seed := r.f
	seed.output = filepath.Join(r.f.project, "source-seed.json")
	first := seed.generate(t, "--source", "opencode", "--session-id", "receipt-root", "--db-path", r.db)
	seedBytes := readCaptureFile(t, seed.output)
	lock, err := capture.AcquirePublication(context.Background(), r.f.output)
	if err != nil {
		t.Fatal(err)
	}
	if err = adapters.BeginPublication(r.f.output, []string{r.f.output}); err != nil {
		t.Fatal(err)
	}
	if err = adapters.SaveFactLedger(r.f.output, first); err != nil {
		t.Fatal(err)
	}
	if _, err = r.open(t).StagePublicationAuxiliary(context.Background(), "TEST-100", r.f.output, receiptFacts(first)); err != nil {
		t.Fatal(err)
	}
	if err = lock.Close(); err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(r.db); err != nil {
		t.Fatal(err)
	}
	if out, err := r.run(t, "--source", "unsupported"); err == nil || !strings.Contains(string(out), "Recovered interrupted publication:") {
		t.Fatalf("first publication did not roll back: %v %s", err, out)
	}
	if _, err := os.Stat(r.f.output); !os.IsNotExist(err) {
		t.Fatalf("rollback invented an accepted empty ledger: %v", err)
	}
	if out, err := r.run(t); err == nil {
		t.Fatalf("unavailable source invented a first accepted ledger: %s", out)
	}
	if _, err := os.Stat(r.f.output); !os.IsNotExist(err) {
		t.Fatalf("retry published a ledger without source facts: %v", err)
	}
	if !bytes.Equal(seedBytes, readCaptureFile(t, seed.output)) {
		t.Fatal("first destination rollback changed independent destination")
	}
}

func TestCaptureReceiptRebindRequiresIndependentSnapshotProof(t *testing.T) {
	r := newReceiptRecoveryFixture(t)
	old := r.initial(t, "canonical")
	next := r.next(t, old, "canonical", false)
	if err := adapters.SaveFactLedger(r.f.output, next); err != nil {
		t.Fatal(err)
	}
	unmatchedBytes := readCaptureFile(t, r.f.output)
	s := r.open(t)
	// An unscoped legacy receipt exactly matching the new artifact is still
	// insufficient once this destination is bound to a different generation.
	if err := s.RecordAuxiliaryReceipt(context.Background(), "TEST-100", receiptFacts(next)); err != nil {
		t.Fatal(err)
	}
	source := readCaptureFile(t, r.db)
	if err := os.Remove(r.db); err != nil {
		t.Fatal(err)
	}
	if out, err := r.run(t); err == nil || !strings.Contains(string(out), "Baseline authorization:") {
		t.Fatalf("bound mismatch borrowed legacy receipt without independent proof: %v %s", err, out)
	}
	if !bytes.Equal(unmatchedBytes, readCaptureFile(t, r.f.output)) {
		t.Fatal("failed receipt rebind changed the artifact")
	}
	if err := os.WriteFile(r.db, source, 0600); err != nil {
		t.Fatal(err)
	}
	refreshed, err := s.Reconcile(context.Background(), "TEST-100", old.Sessions[0].User, 20, capture.DefaultByteBudget)
	if err != nil || len(refreshed.RefreshedSessions) != 1 {
		t.Fatalf("source-backed independent snapshot not established: %+v %v", refreshed, err)
	}
	if !reflect.DeepEqual(receiptFacts(domain.AgenticFactLedger{Sessions: refreshed.RefreshedSessions}), receiptFacts(next)) {
		t.Fatal("independent snapshot does not prove the changed exact rows")
	}
	if err = os.Remove(r.db); err != nil {
		t.Fatal(err)
	}
	if out, err := r.run(t); err != nil {
		t.Fatalf("independently proved explicit rebind failed with source unavailable: %v %s", err, out)
	}
	if !reflect.DeepEqual(receiptFacts(r.f.ledger(t)), receiptFacts(next)) {
		t.Fatal("explicit rebind altered proved source facts")
	}
}
