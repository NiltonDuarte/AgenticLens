package adapters

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"agenticlens/internal/domain"
)

func TestImmutableLedgerDecoderSharesPersistedBoundary(t *testing.T) {
	ticket, title := "TEST-1", "Immutable history"
	version := domain.CurrentSchemaVersion
	ledger := domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title, Sessions: []domain.SessionFact{{ID: "prior", Source: domain.SourceCodex, User: "recorded-user", Invocations: []domain.InvocationFact{{Role: "coder", Model: "recorded-model", Tokens: domain.InvocationTokens{Input: 17}}}}}}
	path := filepath.Join(t.TempDir(), "ledger.json")
	if err := SaveFactLedger(path, ledger); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := LoadExistingFactLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	immutable, err := DecodeExistingFactLedger(body)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(persisted, immutable) {
		t.Fatal("committed blob uses a different ledger boundary")
	}
	var raw map[string]json.RawMessage
	if err = json.Unmarshal(body, &raw); err != nil {
		t.Fatal(err)
	}
	raw["schema_version"] = json.RawMessage(`999`)
	unsupported, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = DecodeExistingFactLedger(unsupported); err == nil {
		t.Fatal("unsupported committed schema accepted")
	}
	if _, err = DecodeExistingFactLedger([]byte(`{"schema_version":6,"sessions":[{"id":17}]}`)); err == nil {
		t.Fatal("malformed committed ledger accepted")
	}
}

func TestHistoryCostNormalizationKeepsExplicitEvidenceAndRejectsInvalidCurrentFacts(t *testing.T) {
	zero, subtotal := float64(0), float64(2.5)
	for _, schema := range []int{domain.CurrentSchemaVersion} {
		for _, evidence := range []domain.CostEvidence{{State: domain.CostUnavailable}, {State: domain.CostObserved, KnownSubtotalUSD: &zero}, {State: domain.CostPartial, KnownSubtotalUSD: &subtotal}} {
			amount := (*float64)(nil)
			if evidence.State == domain.CostObserved {
				amount = &zero
			}
			original := SessionFactDTO{ID: "history", Source: "codex", User: "fixture", Invocations: []InvocationFactDTO{{Model: "recorded", CostUSD: amount, CostEvidence: &evidence}}, Anomalies: []AnomalyDTO{{RuleID: "TOOL_RETRY_STREAK", Evidence: []AnomalyEvidenceRefDTO{{SessionID: "history", EventID: "e"}}}}}
			before, err := json.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			normalized, err := NormalizeHistorySessionForPublication(original, schema)
			if err != nil {
				t.Fatalf("schema%d explicit%s: %v", schema, evidence.State, err)
			}
			after, err := json.Marshal(normalized)
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) {
				t.Fatalf("schema%d rewrote explicit %s cost or supported evidence", schema, evidence.State)
			}
			inputAfter, err := json.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(inputAfter) {
				t.Fatal("normalization mutated the authorized input")
			}
		}
	}
	invalidCurrent := SessionFactDTO{ID: "history", Invocations: []InvocationFactDTO{{CostUSD: &zero}}}
	if _, err := NormalizeHistorySessionForPublication(invalidCurrent, domain.CurrentSchemaVersion); err == nil {
		t.Fatal("current invalid missing evidence was reinterpreted as legacy known zero")
	}
	for _, legacy := range []int{4, 5} {
		if _, err := NormalizeHistorySessionForPublication(invalidCurrent, legacy); err == nil {
			t.Fatalf("accepted retired historical schema %d", legacy)
		}
	}
	if _, err := NormalizeHistorySessionForPublication(invalidCurrent, 99); err == nil {
		t.Fatal("unknown historical schema normalized")
	}
}
