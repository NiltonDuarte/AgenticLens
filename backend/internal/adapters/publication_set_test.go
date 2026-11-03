package adapters

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// A restart must restore every prior member after each visible-write boundary,
// including an absent first-generation member and a second recovery interruption.
func TestPublicationSetInterruptedRecovery(t *testing.T) {
	for _, writes := range []int{0, 1, 2, 3} {
		t.Run(string(rune('0'+writes)), func(t *testing.T) {
			dir := t.TempDir()
			paths := []string{filepath.Join(dir, "ledger.json"), filepath.Join(dir, "capture_report.json"), filepath.Join(dir, "SUMMARY.md")}
			for _, p := range paths[:2] {
				if err := os.WriteFile(p, []byte("prior "+filepath.Base(p)), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := BeginPublication(paths[0], paths); err != nil {
				t.Fatal(err)
			}
			for _, p := range paths[:writes] {
				if err := os.WriteFile(p, []byte("replacement"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := readPublishedLedger(paths[0]); err == nil {
				t.Fatal("interrupted publication exposed as successful read")
			}
			// Model process death/restart by discarding all in-memory transaction state.
			if recovered, err := RecoverPublication(paths[0], paths); !recovered || err != nil {
				t.Fatalf("recovery %v %v", recovered, err)
			}
			for _, p := range paths[:2] {
				b, err := os.ReadFile(p)
				if err != nil || string(b) != "prior "+filepath.Base(p) {
					t.Fatalf("prior artifact lost %s: %s %v", p, b, err)
				}
			}
			if _, err := os.Stat(paths[2]); !os.IsNotExist(err) {
				t.Fatalf("new summary survived rollback: %v", err)
			}
			if recovered, err := RecoverPublication(paths[0], paths); recovered || err != nil {
				t.Fatalf("recovery not idempotent %v %v", recovered, err)
			}
			if err := BeginPublication(paths[0], paths); err != nil {
				t.Fatal(err)
			}
			for _, p := range paths {
				if err := os.WriteFile(p, []byte("accepted"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := CommitPublication(paths[0], paths); err != nil {
				t.Fatal(err)
			}
			if b, err := readPublishedLedger(paths[0]); err != nil || string(b) != "accepted" {
				t.Fatalf("commit %s %v", b, err)
			}
		})
	}
}

func TestPublicationRecoveryFailureRetainsPriorGoodJournal(t *testing.T) {
	dir := t.TempDir()
	ledger := filepath.Join(dir, "ledger")
	report := filepath.Join(dir, "report")
	if err := os.WriteFile(ledger, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := BeginPublication(ledger, []string{ledger, report}); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(report, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(report, "obstruction"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := RecoverPublication(ledger, []string{ledger, report}); err == nil {
		t.Fatal("expected interrupted recovery")
	}
	if err := CheckPublication(ledger); err == nil {
		t.Fatal("failed recovery marked successful")
	}
	if err := os.RemoveAll(report); err != nil {
		t.Fatal(err)
	}
	if _, err := RecoverPublication(ledger, []string{ledger, report}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(ledger)
	if err != nil || string(b) != "original" {
		t.Fatalf("lost prior good %s %v", b, err)
	}
}

// Missing leaves must already be distinct physically, including a nearest
// ancestor alias several not-yet-created directories above the final artifact.
func TestPublicationRejectsMissingAliasDestinations(t *testing.T) {
	for _, nested := range []bool{false, true} {
		t.Run(map[bool]string{false: "parent", true: "nearest-ancestor"}[nested], func(t *testing.T) {
			dir := t.TempDir()
			real := filepath.Join(dir, "real")
			if err := os.Mkdir(real, 0700); err != nil {
				t.Fatal(err)
			}
			alias := filepath.Join(dir, "alias")
			if err := os.Symlink(real, alias); err != nil {
				t.Fatal(err)
			}
			leaf := "first.json"
			if nested {
				leaf = filepath.Join("missing", "deeper", leaf)
			}
			ledger := filepath.Join(real, leaf)
			if err := BeginPublication(ledger, []string{ledger, filepath.Join(alias, leaf)}); err == nil {
				t.Fatal("overlapping first publication accepted")
			}
			if _, err := os.Stat(journalPath(ledger)); !os.IsNotExist(err) {
				t.Fatalf("journal created before distinctness validation: %v", err)
			}
			if _, err := os.Stat(ledger); !os.IsNotExist(err) {
				t.Fatalf("first artifact created: %v", err)
			}
		})
	}
}

func TestPublicationRejectsExistingPhysicalAliases(t *testing.T) {
	for _, kind := range []string{"symlink", "hardlink", "dangling-leaf"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			ledger, alias := filepath.Join(dir, "ledger"), filepath.Join(dir, "alias")
			if kind != "dangling-leaf" {
				if err := os.WriteFile(ledger, []byte("prior"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			var err error
			if kind == "hardlink" {
				err = os.Link(ledger, alias)
			} else {
				err = os.Symlink(ledger, alias)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := BeginPublication(ledger, []string{ledger, alias}); err == nil {
				t.Fatal("physical aliases accepted")
			}
			if kind != "dangling-leaf" {
				b, err := os.ReadFile(ledger)
				if err != nil || string(b) != "prior" {
					t.Fatalf("prior lost: %q %v", b, err)
				}
			}
		})
	}
}

func TestPublicationUntrustedJournalCannotMutateAnyMember(t *testing.T) {
	cases := []string{"outside-write", "outside-delete", "inside-write", "inside-delete", "outside-parent-write", "outside-parent-delete", "same-project-wrong-ledger", "relocated", "copied", "legacy", "version", "count", "missing-member", "duplicate", "alias-parent", "missing-presence", "missing-data", "malformed", "invalid-base64", "null-presence", "relative-member", "too-many-members", "absent-with-data", "unknown-field", "duplicate-key", "duplicate-artifact-key", "wrong-case-key", "trailing", "no-authorization", "authorization-without-ledger", "authorization-alias", "journal-member"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			project, outside := filepath.Join(root, "project"), filepath.Join(root, "outside")
			for _, dir := range []string{project, outside} {
				if err := os.Mkdir(dir, 0700); err != nil {
					t.Fatal(err)
				}
			}
			ledger, report := filepath.Join(project, "ledger"), filepath.Join(project, "report")
			victim := filepath.Join(outside, "victim")
			if name == "inside-write" || name == "inside-delete" {
				victim = filepath.Join(project, "unrelated")
			}
			for _, path := range []string{ledger, report, victim} {
				if err := os.WriteFile(path, []byte("prior"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			paths := []string{ledger, report}
			if err := BeginPublication(ledger, paths); err != nil {
				t.Fatal(err)
			}
			original, err := os.ReadFile(journalPath(ledger))
			if err != nil {
				t.Fatal(err)
			}
			var j publicationJournal
			if err := json.Unmarshal(original, &j); err != nil {
				t.Fatal(err)
			}
			// An earlier legitimate member would be visibly restored if validation
			// were interleaved with mutation of a later malicious member.
			for _, path := range paths {
				if err := os.WriteFile(path, []byte("current"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			recoveryLedger := ledger
			allowed := paths
			switch name {
			case "outside-write", "inside-write":
				j.Artifacts[1].Path = victim
			case "outside-delete", "inside-delete":
				j.Artifacts[1] = priorArtifact{Path: victim}
			case "outside-parent-write", "outside-parent-delete":
				alias := filepath.Join(project, "escaped-parent")
				if err := os.Symlink(outside, alias); err != nil {
					t.Fatal(err)
				}
				j.Artifacts[1].Path = filepath.Join(alias, "victim")
				if name == "outside-parent-delete" {
					j.Artifacts[1].Present = false
					j.Artifacts[1].Data = nil
				}
			case "same-project-wrong-ledger":
				j.Ledger = report
			case "relocated", "copied":
				recoveryLedger = filepath.Join(project, "another-ledger")
				if name == "relocated" {
					recoveryLedger = filepath.Join(outside, "ledger")
				}
				if err := os.WriteFile(recoveryLedger, []byte("current"), 0600); err != nil {
					t.Fatal(err)
				}
				allowed = []string{recoveryLedger, report}
			case "legacy":
				j.Version = 0
				j.Ledger = ""
				j.MemberCount = 0
			case "version":
				j.Version = 2
			case "count":
				j.MemberCount = 3
			case "missing-member":
				j.Artifacts = j.Artifacts[:1]
			case "duplicate":
				j.Artifacts[1] = j.Artifacts[0]
			case "alias-parent":
				alias := filepath.Join(root, "alias")
				if err := os.Symlink(project, alias); err != nil {
					t.Fatal(err)
				}
				j.Artifacts[1].Path = filepath.Join(alias, "report")
			case "relative-member":
				j.Artifacts[1].Path = "relative"
			case "too-many-members":
				j.MemberCount = 4
				j.Artifacts = append(j.Artifacts, j.Artifacts...)
			case "absent-with-data":
				j.Artifacts[1].Present = false
			case "no-authorization":
				allowed = nil
			case "authorization-without-ledger":
				allowed = []string{report}
			case "authorization-alias":
				allowed = []string{ledger, ledger}
			case "journal-member":
				j.Artifacts[1].Path = journalPath(ledger)
			}
			body, err := json.Marshal(j)
			if err != nil {
				t.Fatal(err)
			}
			if name == "malformed" {
				body = []byte(`{"version":`)
			}
			if name == "invalid-base64" {
				body = bytes.Replace(body, []byte(`"data":"cHJpb3I="`), []byte(`"data":"***"`), 1)
			}
			if name == "null-presence" {
				body = bytes.Replace(body, []byte(`"present":true`), []byte(`"present":null`), 1)
			}
			if name == "missing-presence" {
				body = bytes.Replace(body, []byte(`"present":true,`), nil, 1)
			}
			if name == "missing-data" {
				body = bytes.Replace(body, []byte(`,"data":"cHJpb3I="`), nil, 1)
			}
			if name == "unknown-field" {
				body = append([]byte(`{"surprise":true,`), body[1:]...)
			}
			if name == "duplicate-key" {
				body = append([]byte(`{"version":1,`), body[1:]...)
			}
			if name == "duplicate-artifact-key" {
				body = bytes.Replace(body, []byte(`"present":true,`), []byte(`"present":false,"present":true,`), 1)
			}
			if name == "wrong-case-key" {
				body = bytes.Replace(body, []byte(`"version"`), []byte(`"Version"`), 1)
			}
			if name == "trailing" {
				body = append(body, []byte(` {}`)...)
			}
			marker := journalPath(recoveryLedger)
			if err := os.WriteFile(marker, body, 0600); err != nil {
				t.Fatal(err)
			}
			recovered, err := RecoverPublication(recoveryLedger, allowed)
			if !recovered || err == nil {
				t.Fatalf("untrusted journal accepted: %v %v", recovered, err)
			}
			// The compatibility retry must preserve the same full prevalidation:
			// recognizing a retired report never authorizes a journal's other paths.
			recovered, err = RecoverPublicationWithLegacyReport(filepath.Dir(recoveryLedger), recoveryLedger, allowed)
			if !recovered || err == nil {
				t.Fatalf("legacy recovery retry accepted untrusted journal: %v %v", recovered, err)
			}
			for _, path := range paths {
				b, err := os.ReadFile(path)
				if err != nil || string(b) != "current" {
					t.Fatalf("earlier member mutated: %s %q %v", path, b, err)
				}
			}
			b, err := os.ReadFile(victim)
			if err != nil || string(b) != "prior" {
				t.Fatalf("outside victim mutated: %q %v", b, err)
			}
			b, err = os.ReadFile(marker)
			if err != nil || !bytes.Equal(b, body) {
				t.Fatalf("rejected journal changed: %v", err)
			}
		})
	}
}

func TestPublicationMissingSetRecoveryThroughParentAlias(t *testing.T) {
	dir := t.TempDir()
	real, alias := filepath.Join(dir, "real"), filepath.Join(dir, "alias")
	if err := os.Mkdir(real, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, alias); err != nil {
		t.Fatal(err)
	}
	ledger, report := filepath.Join(alias, "ledger"), filepath.Join(real, "report")
	paths := []string{ledger, report}
	if err := BeginPublication(ledger, paths); err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		if err := os.WriteFile(p, []byte("interrupted first generation"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if recovered, err := RecoverPublication(filepath.Join(real, "ledger"), []string{filepath.Join(real, "ledger"), report}); !recovered || err != nil {
		t.Fatalf("recovery: %v %v", recovered, err)
	}
	for _, p := range paths {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatalf("new member survived: %s %v", p, err)
		}
	}
}

// A real restoration error after the first member was restored must retain the
// prior-good bytes and permit an idempotent full retry after the obstruction ends.
func TestPublicationPartialRecoveryCanRetry(t *testing.T) {
	dir := t.TempDir()
	ledger, reportDir := filepath.Join(dir, "ledger"), filepath.Join(dir, "reports")
	report := filepath.Join(reportDir, "report")
	if err := os.Mkdir(reportDir, 0700); err != nil {
		t.Fatal(err)
	}
	paths := []string{ledger, report}
	for _, path := range paths {
		if err := os.WriteFile(path, []byte("prior"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := BeginPublication(ledger, paths); err != nil {
		t.Fatal(err)
	}
	journal, err := os.ReadFile(journalPath(ledger))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ledger, []byte("interrupted"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(report); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(reportDir); err != nil {
		t.Fatal(err)
	}
	if recovered, err := RecoverPublication(ledger, paths); !recovered || err == nil {
		t.Fatalf("expected restore failure: %v %v", recovered, err)
	}
	b, err := os.ReadFile(ledger)
	if err != nil || string(b) != "prior" {
		t.Fatalf("first restore did not execute: %q %v", b, err)
	}
	b, err = os.ReadFile(journalPath(ledger))
	if err != nil || !bytes.Equal(b, journal) {
		t.Fatalf("failure lost journal: %v", err)
	}
	if err := os.Mkdir(reportDir, 0700); err != nil {
		t.Fatal(err)
	}
	if recovered, err := RecoverPublication(ledger, paths); !recovered || err != nil {
		t.Fatalf("retry: %v %v", recovered, err)
	}
	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err != nil || string(b) != "prior" {
			t.Fatalf("retry lost member: %q %v", b, err)
		}
	}
	if recovered, err := RecoverPublication(ledger, paths); recovered || err != nil {
		t.Fatalf("retry not idempotent: %v %v", recovered, err)
	}
}

func TestPublicationLegacyRecoveryRejectsPhysicalReportEscapeBeforeMutation(t *testing.T) {
	for _, present := range []bool{true, false} {
		t.Run(map[bool]string{true: "restore", false: "remove"}[present], func(t *testing.T) {
			project := t.TempDir()
			ledger := filepath.Join(project, "ledger.json")
			victim := filepath.Join(t.TempDir(), "outside")
			if err := os.WriteFile(ledger, []byte("accepted ledger"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(victim, []byte("accepted outside"), 0600); err != nil {
				t.Fatal(err)
			}
			report := filepath.Join(project, "capture_report.json")
			if err := os.Symlink(victim, report); err != nil {
				t.Fatal(err)
			}
			data := []byte(nil)
			if present {
				data = []byte("forged replacement")
			}
			journal := publicationJournal{Version: 1, Ledger: ledger, MemberCount: 2, Artifacts: []priorArtifact{{Path: ledger, Present: true, Data: []byte("forged ledger")}, {Path: victim, Present: present, Data: data}}}
			body, err := json.Marshal(journal)
			if err != nil {
				t.Fatal(err)
			}
			marker := journalPath(ledger)
			if err = os.WriteFile(marker, body, 0600); err != nil {
				t.Fatal(err)
			}
			if recovered, err := RecoverPublicationWithLegacyReport(project, ledger, []string{ledger}); !recovered || err == nil {
				t.Fatalf("outside legacy authority accepted: %v %v", recovered, err)
			}
			for path, want := range map[string][]byte{ledger: []byte("accepted ledger"), victim: []byte("accepted outside"), marker: body} {
				if got, err := os.ReadFile(path); err != nil || !bytes.Equal(got, want) {
					t.Fatalf("unsafe journal mutated %s: %q %v", path, got, err)
				}
			}
		})
	}
}
