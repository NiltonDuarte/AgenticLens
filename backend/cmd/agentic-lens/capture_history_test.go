package main_test

import (
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
	"agenticlens/internal/domain"
)

func historyGit(t *testing.T, project string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = project
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}
func historyBind(t *testing.T, f captureGenerationFixture, id string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(f.store, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(f.store, "sessions", id+".jsonl")
	body := fmt.Sprintf("{\"timestamp\":\"2026-01-01T00:00:00Z\",\"type\":\"session_meta\",\"payload\":{\"id\":%q,\"cwd\":%q,\"timestamp\":\"2026-01-01T00:00:00Z\"}}\n{\"timestamp\":\"2026-01-01T00:00:01Z\",\"type\":\"turn_context\",\"payload\":{\"model\":\"history-model\",\"turn_id\":\"one\"}}\n{\"timestamp\":\"2026-01-01T00:00:02Z\",\"type\":\"token_usage_record\",\"payload\":{\"thread_id\":%q,\"turn_id\":\"one\",\"usage\":{\"input_tokens\":17,\"output_tokens\":2}}}\n", id, f.project, id)
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: id}
	row, err := s.Register(context.Background(), identity, path, "SessionStart")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Bind(context.Background(), identity, "TEST-100", row.Request); err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
}
func newCommittedHistoryFixture(t *testing.T) (captureGenerationFixture, captureGenerationFixture, string) {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	base := newCaptureGenerationFixture(t)
	historyGit(t, base.project, "init", "-qb", "main")
	historyGit(t, base.project, "config", "user.name", "Fixture")
	historyGit(t, base.project, "config", "user.email", "fixture@example.invalid")
	historyGit(t, base.project, "add", ".agenticlens-config.yaml")
	historyGit(t, base.project, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "initial")
	a := base
	a.project = filepath.Join(t.TempDir(), "worktree a")
	a.output = filepath.Join(a.project, "ledger.json")
	historyGit(t, base.project, "worktree", "add", "-qb", "history-a", a.project)
	historyBind(t, a, "prior-root")
	a.generate(t)
	historyGit(t, a.project, "add", "ledger.json")
	historyGit(t, a.project, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "published history")
	b := base
	b.project = filepath.Join(t.TempDir(), "worktree b")
	b.output = filepath.Join(b.project, "ledger.json")
	b.store = t.TempDir()
	historyGit(t, base.project, "worktree", "add", "-qb", "history-b", b.project, "history-a")
	historyBind(t, b, "current-root")
	return a, b, base.project
}

// modelUnboundLegacyReceipt represents a scope written before destination-bound
// receipt metadata existed. It is fixture setup, never an operational recovery.
func modelUnboundLegacyReceipt(t *testing.T, f captureGenerationFixture) {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(f.state, "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("DELETE FROM capture_auxiliary_publications WHERE worktree_id=(SELECT id FROM worktrees WHERE directory=?) AND ticket=? AND destination=?", f.project, "TEST-100", f.output); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
}

func foreignHistoryState(t *testing.T, f captureGenerationFixture) string {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(f.state, "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	var id int64
	if err = db.QueryRow("SELECT id FROM worktrees WHERE directory=?", f.project).Scan(&id); err != nil {
		t.Fatal(err)
	}
	queries := []string{
		"SELECT quote(record) FROM registrations WHERE worktree_id=? ORDER BY identity",
		"SELECT quote(parent)||quote(path) FROM discovery_cursor WHERE worktree_id=? ORDER BY parent",
		"SELECT quote(kind)||quote(parent)||quote(key)||quote(more) FROM child_cursor WHERE worktree_id=? ORDER BY kind,parent",
		"SELECT quote(generation)||quote(acknowledged)||quote(requested_at) FROM publication_request WHERE worktree_id=?",
	}
	var result strings.Builder
	for _, query := range queries {
		rows, err := db.Query(query, id)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var row string
			if err = rows.Scan(&row); err != nil {
				t.Fatal(err)
			}
			result.WriteString(row)
		}
		if err = rows.Err(); err != nil {
			t.Fatal(err)
		}
		if err = rows.Close(); err != nil {
			t.Fatal(err)
		}
	}
	return result.String()
}
func TestCaptureCommittedHistorySurvivesDeletedWorktreeAndRepeatedOwnPublications(t *testing.T) {
	a, b, repo := newCommittedHistoryFixture(t)
	prior := a.ledger(t).Sessions[0]
	before := foreignHistoryState(t, a)
	historyGit(t, repo, "worktree", "remove", "--force", a.project)
	if err := os.RemoveAll(a.store); err != nil {
		t.Fatal(err)
	}
	for pass := 0; pass < 2; pass++ {
		ledger := b.generate(t, "--append", filepath.Join(b.project, "SUMMARY.md"))
		if len(ledger.Sessions) != 2+pass {
			t.Fatalf("pass %d: sessions %+v", pass, ledger.Sessions)
		}
		var retained *domain.SessionFact
		for i := range ledger.Sessions {
			if ledger.Sessions[i].ID == prior.ID {
				retained = &ledger.Sessions[i]
			}
		}
		if retained == nil || !reflect.DeepEqual(*retained, prior) {
			t.Fatalf("history changed: %+v", retained)
		}
		if after := foreignHistoryState(t, a); after != before {
			t.Fatal("foreign registration/cursor/ack bytes changed")
		}
		if pass == 0 {
			historyBind(t, b, "current-second")
		}
	}
	summary := string(readCaptureFile(t, filepath.Join(b.project, "SUMMARY.md")))
	if !strings.Contains(summary, "Retained committed history: 1") || strings.Contains(summary, a.store) {
		t.Fatalf("history summary invalid: %s", summary)
	}
}
func TestCaptureCommittedHistoryRejectsUncommittedForeignChanges(t *testing.T) {
	for _, kind := range []string{"tokens", "model", "parent", "user", "source", "missing committed path", "symlink destination", "fresh collision", "uncommitted foreign addition"} {
		t.Run(kind, func(t *testing.T) {
			a, b, _ := newCommittedHistoryFixture(t)
			original := b.ledger(t)
			switch kind {
			case "tokens":
				original.Sessions[0].Invocations[0].Tokens.Input++
			case "model":
				original.Sessions[0].Invocations[0].Model = "changed"
			case "parent":
				parent := "invented"
				original.Sessions[0].ParentID = &parent
			case "user":
				original.Sessions[0].User = "changed"
			case "source":
				original.Sessions[0].Source = domain.SourceOpenCode
			case "missing committed path":
				historyGit(t, b.project, "rm", "ledger.json")
				historyGit(t, b.project, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "remove history")
			case "symlink destination":
				if err := os.Rename(b.output, b.output+".target"); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("ledger.json.target", b.output); err != nil {
					t.Fatal(err)
				}
			case "fresh collision":
				historyBind(t, b, "prior-root")
			case "uncommitted foreign addition":
				historyBind(t, a, "uncommitted-prior")
				for _, owner := range a.generate(t).Sessions {
					if owner.ID == "uncommitted-prior" {
						original.Sessions = append(original.Sessions, owner)
					}
				}
			}
			if kind != "symlink destination" {
				body, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(original))
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(b.output, body, 0600); err != nil {
					t.Fatal(err)
				}
			}
			before := foreignHistoryState(t, a)
			accepted := readCaptureFile(t, b.output)
			cmd := exec.Command(cliBinaryPath, b.args(t)...)
			cmd.Dir = b.project
			if out, err := cmd.CombinedOutput(); err == nil {
				t.Fatalf("unauthorized %s publication accepted: %s", kind, out)
			}
			if !reflect.DeepEqual(accepted, readCaptureFile(t, b.output)) {
				t.Fatal("failed authorization changed ledger")
			}
			if foreignHistoryState(t, a) != before {
				t.Fatal("failed authorization mutated foreign state")
			}
		})
	}
}

func TestCaptureManualAuxiliaryPublicationSurvivesUnavailableSource(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	f := newCaptureGenerationFixture(t)
	path := setupIsolatedEmptyDB(t)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO session(id,agent,model,time_created,time_updated)VALUES('manual-root','coder','model',1760000000000,1760000001000);
 INSERT INTO message(id,session_id,time_created,time_updated,data)VALUES('m','manual-root',1760000000100,1760000000800,'{"role":"assistant","modelID":"model","time":{"created":1760000000100,"completed":1760000000800},"tokens":{"input":12,"output":3}}');
 INSERT INTO part(id,session_id,message_id,time_created,time_updated,data)VALUES
 ('skill','manual-root','m',1760000000200,1760000000300,'{"type":"tool","tool":"skill","state":{"status":"completed","input":{"name":"fixture-skill"}}}'),
 ('task','manual-root','m',1760000000400,1760000000500,'{"type":"tool","tool":"task","state":{"status":"completed","input":{"subagent_type":"tester","description":"manual work"},"output":"{\"session_id\":\"descriptive-child\"}"}}');`)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	setOpenCodeFixtureProject(t, path, f.project)
	first := f.generate(t, "--source", "opencode", "--session-id", "manual-root", "--db-path", path)
	if len(first.Sessions) != 1 || first.Sessions[0].ID != "manual-root" || first.Sessions[0].Source != domain.SourceOpenCode {
		t.Fatalf("source owner changed: %+v", first.Sessions)
	}
	if len(first.Sessions[0].Skills) != 1 || len(first.Sessions[0].TaskLinks) != 1 || len(first.Skills) != 0 || len(first.TaskLinks) != 0 {
		t.Fatalf("source facts missing from canonical owner or duplicated at root: %+v", first)
	}
	skills, taskLinks := first.SkillFacts(), first.TaskLinkFacts()
	if len(skills) != 1 || skills[0].Name != "fixture-skill" || skills[0].SessionID != "manual-root" || skills[0].Source != domain.SourceOpenCode || skills[0].SourceEventID != "opencode:part:skill" || skills[0].SourceMessageID != "m" || skills[0].Outcome != "succeeded" {
		t.Fatalf("source skill evidence changed: %+v", skills)
	}
	if len(taskLinks) != 1 || taskLinks[0].Description != "manual work" || taskLinks[0].SessionID != "manual-root" || taskLinks[0].Source != domain.SourceOpenCode || taskLinks[0].SourceEventID != "opencode:part:task" || taskLinks[0].SourceMessageID != "m" || taskLinks[0].ChildSession != "descriptive-child" {
		t.Fatalf("source delegation evidence changed: %+v", taskLinks)
	}
	if err = os.Remove(path); err != nil {
		t.Fatal(err)
	}
	second := f.generate(t, "--source", "opencode", "--db-path", path)
	withoutGit := exec.Command(cliBinaryPath, f.args(t, "--source", "opencode", "--db-path", path)...)
	withoutGit.Dir = f.project
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "PATH=") {
			withoutGit.Env = append(withoutGit.Env, entry)
		}
	}
	withoutGit.Env = append(withoutGit.Env, "PATH="+t.TempDir())
	if out, err := withoutGit.CombinedOutput(); err != nil {
		t.Fatalf("current-only stored data gained optional Git dependency: %v %s", err, out)
	}
	withoutGitLedger := f.ledger(t)
	if !reflect.DeepEqual(first.SkillFacts(), withoutGitLedger.SkillFacts()) || !reflect.DeepEqual(first.TaskLinkFacts(), withoutGitLedger.TaskLinkFacts()) {
		t.Fatal("own manual auxiliary rows lost without source or Git")
	}
	if !reflect.DeepEqual(first.SkillFacts(), second.SkillFacts()) || !reflect.DeepEqual(first.TaskLinkFacts(), second.TaskLinkFacts()) {
		t.Fatal("own manual auxiliary rows lost without hydration")
	}
	if len(second.Sessions) != 1 || !reflect.DeepEqual(first.Sessions[0].Skills, second.Sessions[0].Skills) || !reflect.DeepEqual(first.Sessions[0].TaskLinks, second.Sessions[0].TaskLinks) {
		t.Fatal("canonical owner auxiliary rows changed without hydration")
	}
	if len(second.Sessions) != 1 || second.Sessions[0].ParentID != nil {
		t.Fatal("descriptive child link created native ancestry")
	}
	// A receipt attests exact rows and their multiplicity, not merely the fact
	// that this worktree once published some auxiliary data.
	accepted := readCaptureFile(t, f.output)
	for _, edit := range []string{"skill name", "task description", "skill multiplicity", "task multiplicity", "skill event identity", "skill message identity", "skill outcome", "task invocation identity", "task child identity", "owner source"} {
		candidate := f.ledger(t)
		switch edit {
		case "skill name":
			candidate.Sessions[0].Skills[0].Name = "unproven-skill"
		case "task description":
			candidate.Sessions[0].TaskLinks[0].Description = "unproven description"
		case "skill multiplicity":
			candidate.Sessions[0].Skills = append(candidate.Sessions[0].Skills, candidate.Sessions[0].Skills[0])
		case "task multiplicity":
			candidate.Sessions[0].TaskLinks = append(candidate.Sessions[0].TaskLinks, candidate.Sessions[0].TaskLinks[0])
		case "skill event identity":
			candidate.Sessions[0].Skills[0].SourceEventID = "unproven-event"
		case "skill message identity":
			candidate.Sessions[0].Skills[0].SourceMessageID = "unproven-message"
		case "skill outcome":
			candidate.Sessions[0].Skills[0].Outcome = "failed"
		case "task invocation identity":
			candidate.Sessions[0].TaskLinks[0].SourceInvocationID = "unproven-call"
		case "task child identity":
			candidate.Sessions[0].TaskLinks[0].ChildSession = "unproven-child"
		case "owner source":
			candidate.Sessions[0].Source = domain.SourceCodex
		}
		body, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(candidate))
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(f.output, body, 0600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(cliBinaryPath, f.args(t, "--source", "opencode", "--db-path", path)...)
		cmd.Dir = f.project
		if out, err := cmd.CombinedOutput(); err == nil {
			t.Fatalf("unproven auxiliary %s accepted: %s", edit, out)
		}
		if !reflect.DeepEqual(readCaptureFile(t, f.output), body) {
			t.Fatal("failed auxiliary authorization changed baseline")
		}
		if err = os.WriteFile(f.output, accepted, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCaptureMergedCommittedLedgerRetainsBothWorktrees(t *testing.T) {
	a, b, repo := newCommittedHistoryFixture(t)
	published := b.generate(t)
	historyGit(t, b.project, "add", "ledger.json")
	historyGit(t, b.project, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "union of published worktrees")
	historyGit(t, a.project, "merge", "--ff-only", "history-b")
	c := b
	c.project = filepath.Join(t.TempDir(), "merged worktree")
	c.output = filepath.Join(c.project, "ledger.json")
	c.store = t.TempDir()
	historyGit(t, repo, "worktree", "add", "-qb", "history-c", c.project, "history-a")
	historyBind(t, c, "third-root")
	beforeA, beforeB := foreignHistoryState(t, a), foreignHistoryState(t, b)
	result := c.generate(t)
	if len(result.Sessions) != 3 {
		t.Fatalf("merged histories lost: %+v", result.Sessions)
	}
	for _, prior := range published.Sessions {
		found := false
		for _, retained := range result.Sessions {
			if retained.ID == prior.ID {
				found = true
				if !reflect.DeepEqual(retained, prior) {
					t.Fatalf("merged committed fact changed: %+v", retained)
				}
			}
		}
		if !found {
			t.Fatalf("merged owner %s missing", prior.ID)
		}
	}
	if foreignHistoryState(t, a) != beforeA || foreignHistoryState(t, b) != beforeB {
		t.Fatal("merged publication changed foreign scopes")
	}
}

func TestCaptureCommittedHistoryPreservesNativeParent(t *testing.T) {
	a, b, _ := newCommittedHistoryFixture(t)
	historyBind(t, a, "prior-child")
	path := filepath.Join(a.store, "sessions", "prior-child.jsonl")
	body := string(readCaptureFile(t, path))
	body = strings.Replace(body, `"payload":{"id":"prior-child",`, `"payload":{"source":{"subagent":{"thread_spawn":{"parent_thread_id":"prior-root","depth":1}}},"id":"prior-child",`, 1)
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	before := a.generate(t)
	historyGit(t, a.project, "add", "ledger.json")
	historyGit(t, a.project, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "native child publication")
	historyGit(t, b.project, "merge", "--ff-only", "history-a")
	if err := os.RemoveAll(a.store); err != nil {
		t.Fatal(err)
	}
	result := b.generate(t)
	if result.SchemaVersion == nil || *result.SchemaVersion != domain.CurrentSchemaVersion {
		t.Fatal("history changed canonical schema")
	}
	found := false
	for _, owner := range result.Sessions {
		if owner.ID == "prior-child" {
			found = true
			if owner.ParentID == nil || *owner.ParentID != "prior-root" {
				t.Fatalf("native parent lost: %+v", owner)
			}
			for _, prior := range before.Sessions {
				if prior.ID == owner.ID && !reflect.DeepEqual(prior, owner) {
					t.Fatal("retained child value changed")
				}
			}
		}
	}
	if !found {
		t.Fatal("native child missing from retained history")
	}
}

func TestCaptureCommittedHistoryPreservesSupportedStoredEvidence(t *testing.T) {
	a, b, _ := newCommittedHistoryFixture(t)
	old := a.ledger(t)
	dto := adapters.AgenticFactLedgerDTOFromDomain(old)
	tool := "bash"
	dto.Sessions[0].Events = []adapters.SessionEventFactDTO{{EventID: "stored-event", Timestamp: "2026-01-01T00:00:01Z", Role: "coder", EventType: "tool_call", ToolName: &tool, Status: "error", ContextTokens: 3}}
	count := int64(3)
	dto.Sessions[0].Anomalies = []adapters.AnomalyDTO{
		{RuleID: string(domain.RuleToolRetryStreak), Severity: string(domain.SeverityWarning), Context: adapters.AnomalyContextDTO{ToolName: &tool, Count: &count}, Evidence: []adapters.AnomalyEvidenceRefDTO{{SessionID: "prior-root", EventID: "stored-event"}}},
		{RuleID: string(domain.RuleContextSaturationSpike), Severity: string(domain.SeverityCritical), Context: adapters.AnomalyContextDTO{ContextTokens: &count}},
	}
	zero := float64(0)
	knownZero := dto.Sessions[0].Invocations[0]
	knownZero.Model = "known-zero-model"
	knownZero.CostUSD = &zero
	knownZero.CostEvidence = &domain.CostEvidence{State: domain.CostObserved, KnownSubtotalUSD: &zero}
	dto.Sessions[0].Invocations = append(dto.Sessions[0].Invocations, knownZero)
	body, err := json.Marshal(dto)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(a.output, body, 0600); err != nil {
		t.Fatal(err)
	}
	accepted, err := adapters.LoadExistingFactLedger(a.output)
	if err != nil {
		t.Fatal(err)
	}
	historyGit(t, a.project, "add", "ledger.json")
	historyGit(t, a.project, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "supported stored evidence")
	historyGit(t, b.project, "merge", "--ff-only", "history-a")
	if err = os.RemoveAll(a.store); err != nil {
		t.Fatal(err)
	}
	for pass := 0; pass < 2; pass++ {
		b.generate(t)
		persisted, err := adapters.LoadExistingFactLedger(b.output)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, owner := range persisted.PersistedSessions {
			if owner.ID == "prior-root" {
				found = true
				if !reflect.DeepEqual(owner, accepted.PersistedSessions[0]) {
					t.Fatalf("pass %d changed committed sparse events/anomalies/evidence/null-cost values: before=%+v after=%+v", pass, accepted.PersistedSessions[0], owner)
				}
			}
		}
		if !found {
			t.Fatal("historical evidence owner missing")
		}
	}
	var modified adapters.AgenticFactLedgerDTO
	if err = json.Unmarshal(readCaptureFile(t, b.output), &modified); err != nil {
		t.Fatal(err)
	}
	for i := range modified.Sessions {
		if modified.Sessions[i].ID == "prior-root" {
			modified.Sessions[i].Anomalies[0].Evidence[0].EventID = "unproven-evidence"
		}
	}
	changed, err := json.Marshal(modified)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(b.output, changed, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cliBinaryPath, b.args(t)...)
	cmd.Dir = b.project
	if out, err := cmd.CombinedOutput(); err == nil {
		t.Fatalf("foreign anomaly evidence edit accepted: %s", out)
	}
	if !reflect.DeepEqual(readCaptureFile(t, b.output), changed) {
		t.Fatal("rejected evidence edit changed baseline")
	}
}

func TestCaptureCommittedHistoryRejectsRawBodiesBeforeReplacement(t *testing.T) {
	for _, field := range []string{"prompt_content", "reasoning_content", "tool_input", "tool_output"} {
		t.Run(field, func(t *testing.T) {
			a, b, _ := newCommittedHistoryFixture(t)
			dto := adapters.AgenticFactLedgerDTOFromDomain(a.ledger(t))
			dto.Sessions[0].Events = []adapters.SessionEventFactDTO{{EventID: "unsafe", Timestamp: "2026-01-01T00:00:01Z", Role: "coder", EventType: "tool_call", Status: "completed"}}
			body, err := json.Marshal(dto)
			if err != nil {
				t.Fatal(err)
			}
			body = []byte(strings.Replace(string(body), `"event_id":"unsafe"`, `"`+field+`":"PRIVATE_BODY_SENTINEL","event_id":"unsafe"`, 1))
			if err = os.WriteFile(a.output, body, 0600); err != nil {
				t.Fatal(err)
			}
			historyGit(t, a.project, "add", "ledger.json")
			historyGit(t, a.project, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "incompatible historical representation")
			historyGit(t, b.project, "merge", "--ff-only", "history-a")
			// The working baseline is supported and contains no raw bodies. Only
			// the immutable committed authority has the incompatible representation.
			supported, err := json.Marshal(dto)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(b.output, supported, 0600); err != nil {
				t.Fatal(err)
			}
			accepted := readCaptureFile(t, b.output)
			cmd := exec.Command(cliBinaryPath, b.args(t)...)
			cmd.Dir = b.project
			out, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("raw historical %s body accepted: %s", field, out)
			}
			if err = adapters.CheckPublication(b.output); err != nil {
				t.Fatalf("incompatible historical representation staged publication: %v", err)
			}
			if strings.Contains(string(out), "PRIVATE_BODY_SENTINEL") {
				t.Fatal("raw body leaked in recovery diagnostic")
			}
			if !reflect.DeepEqual(readCaptureFile(t, b.output), accepted) {
				t.Fatal("privacy rejection changed accepted baseline")
			}
		})
	}
}
