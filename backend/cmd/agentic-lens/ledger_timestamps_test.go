package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestAutomaticGenerationRetainsLedgerTimestampsUntilFactsChange(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	writeSnapshot := func(completed string) string {
		t.Helper()
		path := f.write(t, 10)
		body := readCaptureFile(t, path)
		body = fmt.Appendf(body, `{"timestamp":"2026-01-01T00:00:02Z","type":"event_msg","payload":{"type":"task_started","turn_id":"one"}}
{"timestamp":%q,"type":"event_msg","payload":{"type":"task_complete","turn_id":"one"}}
`, completed)
		if err := os.WriteFile(path, body, 0600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	const firstInvocationCompletion = "2026-01-01T00:00:03Z"
	f.bind(t, writeSnapshot(firstInvocationCompletion))
	summary := filepath.Join(f.project, "delivery.md")
	const authored = "# Report\n\nAuthored context.\n"
	if err := os.WriteFile(summary, []byte(authored), 0600); err != nil {
		t.Fatal(err)
	}
	first := f.generate(t, "--append", summary)
	assertCaptureRootInput(t, first, 10)
	if first.UpdatedAt == "" || first.CompletedAt == nil || *first.CompletedAt == "" {
		t.Fatalf("first publication must assign both timestamps: updated_at=%q, completed_at=%v", first.UpdatedAt, first.CompletedAt)
	}
	// Seed an earlier pair without waiting for the wall clock. The source's
	// invocation timestamps and all other facts remain as generated.
	const oldUpdated, oldCompleted = "2026-01-02T00:00:00+00:00", "2026-01-02T00:00:01+00:00"
	var stored map[string]json.RawMessage
	if err := json.Unmarshal(readCaptureFile(t, f.output), &stored); err != nil {
		t.Fatal(err)
	}
	stored["updated_at"] = json.RawMessage(`"` + oldUpdated + `"`)
	stored["completed_at"] = json.RawMessage(`"` + oldCompleted + `"`)
	// Format the accepted snapshot with a different object-key order and
	// indentation to also check that representation alone is a no-op.
	baseline, err := json.MarshalIndent(stored, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f.output, baseline, 0600); err != nil {
		t.Fatal(err)
	}
	ledgerMTime := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(f.output, ledgerMTime, ledgerMTime); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(f.output)
	if err != nil {
		t.Fatal(err)
	}
	first = f.ledger(t)
	if first.UpdatedAt != oldUpdated || first.CompletedAt == nil || *first.CompletedAt != oldCompleted {
		t.Fatalf("fixture failed to seed old ledger timestamps: updated_at=%q, completed_at=%v", first.UpdatedAt, first.CompletedAt)
	}
	assertUnchangedLedger := func(stage string) {
		t.Helper()
		if !bytes.Equal(readCaptureFile(t, f.output), baseline) {
			after := f.ledger(t)
			t.Errorf("%s rewrote the ledger: updated_at %q -> %q, completed_at %q -> %s", stage, oldUpdated, after.UpdatedAt, oldCompleted, rootClockJSON(after.CompletedAt))
		}
		after, err := os.Stat(f.output)
		if err != nil {
			t.Fatal(err)
		}
		if !os.SameFile(before, after) || !after.ModTime().Equal(before.ModTime()) {
			t.Errorf("%s replaced or wrote the ledger file", stage)
		}
	}

	malformedSummary := []byte(authored + "\n<!-- agentic-lens:metrics:start -->\nUnterminated metrics.\n")
	if err := os.WriteFile(summary, malformedSummary, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cliBinaryPath, f.args(t, "--append", summary)...)
	cmd.Dir = f.project
	if out, err := cmd.CombinedOutput(); err == nil || !bytes.Contains(out, []byte("unterminated generated metrics block")) {
		t.Fatalf("expected the unchanged refresh to reach the malformed summary: %v %s", err, out)
	}
	assertUnchangedLedger("failed summary refresh")
	if !bytes.Equal(readCaptureFile(t, summary), malformedSummary) {
		t.Error("failed publication changed the existing summary")
	}
	if _, err := os.Stat(f.output + ".publication-pending"); !os.IsNotExist(err) {
		t.Fatalf("handled failure left a pending publication: %v", err)
	}

	// Summary repair must still run when the ledger facts are unchanged.
	if err := os.WriteFile(summary, []byte(authored), 0600); err != nil {
		t.Fatal(err)
	}
	f.generate(t, "--append", summary)
	assertUnchangedLedger("successful summary retry")
	rendered := readCaptureFile(t, summary)
	if !bytes.Contains(rendered, []byte(authored)) || !bytes.Contains(rendered, []byte("<!-- agentic-lens:metrics:start -->")) {
		t.Error("an unchanged ledger prevented its summary from being repaired")
	}

	// A sub-second move keeps rounded duration and token counts identical but
	// changes the native invocation completion, which is a persisted fact.
	const changedInvocationCompletion = "2026-01-01T00:00:03.100Z"
	writeSnapshot(changedInvocationCompletion)
	changed := f.generate(t, "--append", summary)
	assertCaptureRootInput(t, changed, 10)
	oldInvocation, newInvocation := first.Sessions[0].Invocations[0], changed.Sessions[0].Invocations[0]
	if oldInvocation.CompletedAt == nil || *oldInvocation.CompletedAt != firstInvocationCompletion || newInvocation.CompletedAt == nil || *newInvocation.CompletedAt != changedInvocationCompletion || oldInvocation.ActiveDurationSeconds != newInvocation.ActiveDurationSeconds {
		t.Fatalf("expected a nested completion change with unchanged rounded duration, got before=%+v and after=%+v", oldInvocation, newInvocation)
	}
	oldInvocation.CompletedAt, newInvocation.CompletedAt = nil, nil
	if !reflect.DeepEqual(oldInvocation, newInvocation) {
		t.Fatalf("fixture changed invocation facts other than completion: before=%+v and after=%+v", oldInvocation, newInvocation)
	}
	if changed.UpdatedAt <= oldUpdated || changed.CompletedAt == nil || *changed.CompletedAt <= oldCompleted {
		t.Fatalf("changed source facts must advance both timestamps: updated_at=%q, completed_at=%v", changed.UpdatedAt, changed.CompletedAt)
	}
}

func rootClockJSON(value *string) string {
	if value == nil {
		return "null"
	}
	return fmt.Sprintf("%q", *value)
}
