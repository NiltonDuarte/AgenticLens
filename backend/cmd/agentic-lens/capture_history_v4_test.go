package main_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

// Regression discovered by reviewer-data / gpt-6.1-sol / high: an accepted v4
// historical owner must use the same cost representation mapping as a current
// persisted owner, without compacting its supported evidence.
func TestCaptureCommittedV4HistoryNormalizesLikeCurrentUnavailableSource(t *testing.T) {
	for _, kind := range []string{"absent amount and evidence", "null amount missing evidence", "zero missing evidence", "nonzero missing evidence", "null evidence"} {
		t.Run(kind, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("XDG_CACHE_HOME", t.TempDir())
			a, b, _ := newCommittedHistoryFixture(t)
			original := adapters.AgenticFactLedgerDTOFromDomain(a.ledger(t))
			version := 4
			original.SchemaVersion = &version
			tool := "bash"
			count := int64(3)
			original.Sessions[0].Events = []adapters.SessionEventFactDTO{{EventID: "legacy-event", Timestamp: "2026-01-01T00:00:01Z", Role: "coder", EventType: "tool_call", ToolName: &tool, Status: "error", ContextTokens: 11}}
			original.Sessions[0].Anomalies = []adapters.AnomalyDTO{
				{RuleID: string(domain.RuleToolRetryStreak), Severity: string(domain.SeverityWarning), Context: adapters.AnomalyContextDTO{ToolName: &tool, Count: &count}, Evidence: []adapters.AnomalyEvidenceRefDTO{{SessionID: "prior-root", EventID: "legacy-event"}}},
				{RuleID: string(domain.RuleContextSaturationSpike), Severity: string(domain.SeverityCritical), Context: adapters.AnomalyContextDTO{ContextTokens: &count}},
			}
			original.TaskLinks = []adapters.TaskLinkFactDTO{{SubagentType: "tester", Description: "committed descriptive legacy task", ChildSession: "descriptive-child", Timestamp: "2026-01-01T00:00:01Z"}}
			encoded, err := json.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			var wire struct {
				SchemaVersion int                          `json:"schema_version"`
				Sessions      []map[string]json.RawMessage `json:"sessions"`
				TaskLinks     []map[string]json.RawMessage `json:"task_links"`
			}
			if err = json.Unmarshal(encoded, &wire); err != nil {
				t.Fatal(err)
			}
			// Modify only the legacy cost members, retaining all real fixture metadata.
			var envelope map[string]json.RawMessage
			if err = json.Unmarshal(encoded, &envelope); err != nil {
				t.Fatal(err)
			}
			for _, session := range wire.Sessions {
				var invocations []map[string]json.RawMessage
				if err = json.Unmarshal(session["invocations"], &invocations); err != nil {
					t.Fatal(err)
				}
				for _, invocation := range invocations {
					setLegacyHistoryCost(invocation, kind)
				}
				session["invocations"], err = json.Marshal(invocations)
				if err != nil {
					t.Fatal(err)
				}
			}
			for _, task := range wire.TaskLinks {
				setLegacyHistoryCost(task, kind)
			}
			envelope["sessions"], err = json.Marshal(wire.Sessions)
			if err != nil {
				t.Fatal(err)
			}
			envelope["task_links"], err = json.Marshal(wire.TaskLinks)
			if err != nil {
				t.Fatal(err)
			}
			legacy, err := json.Marshal(envelope)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(a.output, legacy, 0600); err != nil {
				t.Fatal(err)
			}
			accepted, err := adapters.LoadExistingFactLedger(a.output)
			if err != nil {
				t.Fatal(err)
			}
			historyGit(t, a.project, "add", "ledger.json")
			historyGit(t, a.project, "commit", "-qm", "supported v4 committed history")
			historyGit(t, b.project, "merge", "--ff-only", "history-a")
			pinned := historyGit(t, b.project, "rev-parse", "HEAD")
			if err = os.RemoveAll(a.store); err != nil {
				t.Fatal(err)
			}
			a.generate(t)
			current, err := adapters.LoadExistingFactLedger(a.output)
			if err != nil {
				t.Fatal(err)
			}
			assertLegacyHistoryCost(t, kind, current.PersistedSessions[0].Invocations[0].CostUSD, current.PersistedSessions[0].Invocations[0].CostEvidence)
			if current.SchemaVersion == nil || *current.SchemaVersion != 5 {
				t.Fatal("supported current v4 did not publish schema5")
			}
			afterControl := foreignHistoryState(t, a)
			// Native current-only generation must not be used as a source for foreign
			// authorization: its output is uncommitted and HEAD still contains v4.
			for pass := 0; pass < 2; pass++ {
				b.generate(t)
				published, err := adapters.LoadExistingFactLedger(b.output)
				if err != nil {
					t.Fatal(err)
				}
				if published.SchemaVersion == nil || *published.SchemaVersion != 5 {
					t.Fatal("foreign v4 did not publish schema5")
				}
				var retained *adapters.SessionFactDTO
				for i := range published.PersistedSessions {
					if published.PersistedSessions[i].ID == "prior-root" {
						retained = &published.PersistedSessions[i]
					}
				}
				if retained == nil {
					t.Fatal("legacy historical owner missing")
				}
				assertLegacyHistoryCost(t, kind, retained.Invocations[0].CostUSD, retained.Invocations[0].CostEvidence)
				if !reflect.DeepEqual(retained.Invocations[0].CostUSD, current.PersistedSessions[0].Invocations[0].CostUSD) || !reflect.DeepEqual(retained.Invocations[0].CostEvidence, current.PersistedSessions[0].Invocations[0].CostEvidence) {
					t.Fatal("foreign cost mapping differs from current persisted-data path")
				}
				expected := accepted.PersistedSessions[0]
				expected.Invocations = append([]adapters.InvocationFactDTO{}, expected.Invocations...)
				expected.Invocations[0].CostUSD = retained.Invocations[0].CostUSD
				expected.Invocations[0].CostEvidence = retained.Invocations[0].CostEvidence
				expectedBytes, err := json.Marshal(expected)
				if err != nil {
					t.Fatal(err)
				}
				retainedBytes, err := json.Marshal(retained)
				if err != nil {
					t.Fatal(err)
				}
				if string(expectedBytes) != string(retainedBytes) {
					t.Fatalf("pass %d changed non-cost historical facts/evidence/order", pass)
				}
				if len(published.TaskLinks) != 1 || published.TaskLinks[0].ChildSession != "descriptive-child" {
					t.Fatal("legacy descriptive task changed")
				}
				if !reflect.DeepEqual(published.TaskLinks, current.TaskLinks) {
					t.Fatal("legacy task cost normalization differs from current path")
				}
				if foreignHistoryState(t, a) != afterControl {
					t.Fatal("history normalization changed foreign operational metadata")
				}
				if historyGit(t, b.project, "rev-parse", "HEAD") != pinned {
					t.Fatal("test lost pinned v4 authority")
				}
			}
			// Representation normalization grants no authority to change the cost value.
			var modified adapters.AgenticFactLedgerDTO
			if err = json.Unmarshal(readCaptureFile(t, b.output), &modified); err != nil {
				t.Fatal(err)
			}
			value := float64(99)
			for i := range modified.Sessions {
				if modified.Sessions[i].ID == "prior-root" {
					modified.Sessions[i].Invocations[0].CostUSD = &value
					modified.Sessions[i].Invocations[0].CostEvidence = &domain.CostEvidence{State: domain.CostObserved, KnownSubtotalUSD: &value}
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
				t.Fatalf("changed legacy owner cost accepted: %s", out)
			}
			if err = adapters.CheckPublication(b.output); err != nil {
				t.Fatalf("rejected cost edit staged publication: %v", err)
			}
			if !reflect.DeepEqual(changed, readCaptureFile(t, b.output)) {
				t.Fatal("rejected cost edit replaced baseline")
			}
		})
	}
}

func setLegacyHistoryCost(row map[string]json.RawMessage, kind string) {
	delete(row, "cost_evidence")
	switch kind {
	case "absent amount and evidence":
		delete(row, "cost_usd")
	case "zero missing evidence":
		row["cost_usd"] = json.RawMessage(`0`)
	case "nonzero missing evidence":
		row["cost_usd"] = json.RawMessage(`1.25`)
	case "null evidence":
		row["cost_usd"] = json.RawMessage(`null`)
		row["cost_evidence"] = json.RawMessage(`null`)
	default:
		row["cost_usd"] = json.RawMessage(`null`)
	}
}
func assertLegacyHistoryCost(t *testing.T, kind string, amount *float64, evidence *domain.CostEvidence) {
	t.Helper()
	if evidence == nil {
		t.Fatal("legacy cost evidence remained absent")
	}
	switch kind {
	case "zero missing evidence", "nonzero missing evidence":
		expected := float64(0)
		if kind == "nonzero missing evidence" {
			expected = 1.25
		}
		if amount == nil || *amount != expected || evidence.State != domain.CostObserved || evidence.KnownSubtotalUSD == nil || *evidence.KnownSubtotalUSD != expected {
			t.Fatalf("legacy numeric cost semantics changed: amount=%v evidence=%+v", amount, evidence)
		}
	default:
		if amount != nil || evidence.State != domain.CostUnavailable || evidence.KnownSubtotalUSD != nil {
			t.Fatalf("missing/null legacy cost became observed or known zero: amount=%v evidence=%+v", amount, evidence)
		}
	}
}
