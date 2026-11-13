package adapters

import (
	"agenticlens/internal/adapters/capture"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

const formatFixture = `{"schema_version":5,"ticket_id":"T-1","sessions":[{"tools":[{"name":"read","calls":9007199254740993,"unknown":null}],"invocations":[{"tokens":{"input":1e+09,"output":1.2300},"cost_evidence":{"future":{"x":" braces {} \\\" \\n "}},"cost_usd":null,"evidence":{"other":true}}],"anomalies":[{"context":{"x":[1,2]},"message":"a  b"}]}],"unknown":{"tokens":{"input":3},"tools":[{"x":4}]},"anomalies":[{"context":{"value":-0.00E-2}}]}`

func TestFormatLedgerJSONLexemesAndLayout(t *testing.T) {
	formatted, err := FormatLedgerJSON([]byte(formatFixture))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{`{"name": "read", "calls": 9007199254740993, "unknown": null}`, `"tokens": {"input": 1e+09, "output": 1.2300}`, `"evidence": {"other": true}`, `"cost_evidence": {"future": {"x":`, `{"context": {"x": [1, 2]}, "message": "a  b"}`, `{"context": {"value": -0.00E-2}}`} {
		if !bytes.Contains(formatted, []byte(line)) {
			t.Errorf("missing compact layout %s\n%s", line, formatted)
		}
	}
	if !bytes.Contains(formatted, []byte("\"unknown\": {\n    \"tokens\": {\n")) {
		t.Fatalf("unrelated tokens compacted: %s", formatted)
	}
	var originalCompact, formattedCompact bytes.Buffer
	if err := json.Compact(&originalCompact, []byte(formatFixture)); err != nil {
		t.Fatal(err)
	}
	if err := json.Compact(&formattedCompact, formatted); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(originalCompact.Bytes(), formattedCompact.Bytes()) {
		t.Fatal("changed JSON lexemes/order")
	}
	again, err := FormatLedgerJSON(formatted)
	if err != nil || !bytes.Equal(again, formatted) {
		t.Fatalf("not idempotent: %v", err)
	}
	if !bytes.HasSuffix(formatted, []byte("\n")) || bytes.HasSuffix(formatted, []byte("\n\n")) {
		t.Fatal("incorrect final newline")
	}
	legacy, err := FormatLedgerJSON([]byte(`{"ticket_id":"L","invocations":[{"tokens":{"x":2},"evidence":{"x":null}}],"tools":[{"x":1}]}`))
	if err != nil || !bytes.Contains(legacy, []byte(`"tokens": {"x": 2}`)) || !bytes.Contains(legacy, []byte("\"tools\": [\n    {\"x\": 1}")) {
		t.Fatalf("legacy layout: %s %v", legacy, err)
	}
}
func TestFormatLedgerJSONRejectsAmbiguousInputs(t *testing.T) {
	for _, input := range []string{`{`, `{} `, `[]`, `{"ticket_id":"X","sessions":{},"schema_version":6}`, `{"ticket_id":"X","sessions":[],"schema_version":999}`, `{"ticket_id":"X","sessions":[{"tools":[null]}]}`, `{"ticket_id":"X","sessions":[],"future":{"x":1,"\u0078":2}}`, `{"ticket_id":"X","invocations":[{"tokens":[]}]}`, `{"ticket_id":"X","sessions":[{"invocations":[{"cost_evidence":1}]}]}`, `{"ticket_id":"X","sessions":[],"anomalies":[false]}`} {
		t.Run(input, func(t *testing.T) {
			if _, err := FormatLedgerJSON([]byte(input)); err == nil {
				t.Fatal("accepted unsupported or ambiguous ledger")
			}
		})
	}
}
func putFormatLedger(t *testing.T, root, id, body string) string {
	t.Helper()
	writePublicationTestPolicy(t, root)
	path := filepath.Join(root, id, "agentic_metrics.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0640); err != nil {
		t.Fatal(err)
	}
	return path
}
func TestFormatLedgerFilesSelectionAndPreflight(t *testing.T) {
	root := t.TempDir()
	path := putFormatLedger(t, root, "T-1", formatFixture)
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []LedgerFormatMode{LedgerFormatDryRun, LedgerFormatCheck} {
		results, err := FormatLedgerFiles(context.Background(), LedgerFormatOptions{Project: root, TicketIDs: []string{"T-1", "T-1"}, Mode: mode})
		if err != nil || len(results) != 1 || !results[0].Changed || results[0].Written {
			t.Fatalf("preview: %+v %v", results, err)
		}
		after, err := os.Stat(path)
		if err != nil || !after.ModTime().Equal(before.ModTime()) {
			t.Fatalf("preview touched file: %v", err)
		}
	}
	broken := putFormatLedger(t, root, "T-2", "{")
	if _, err := FormatLedgerFiles(context.Background(), LedgerFormatOptions{Project: root, All: true}); err == nil {
		t.Fatal("expected preflight failure")
	}
	body, err := os.ReadFile(path)
	if err != nil || string(body) != formatFixture {
		t.Fatal("preflight replaced valid sibling")
	}
	if err := os.Remove(broken); err != nil {
		t.Fatal(err)
	}
	results, err := FormatLedgerFiles(context.Background(), LedgerFormatOptions{Project: root, All: true})
	if err != nil || !results[0].Written {
		t.Fatalf("write: %+v %v", results, err)
	}
	after, err := os.Stat(path)
	if err != nil || after.Mode().Perm() != 0640 {
		t.Fatalf("permissions: %v %v", after, err)
	}
	mark := time.Unix(123456789, 0)
	if err := os.Chtimes(path, mark, mark); err != nil {
		t.Fatal(err)
	}
	results, err = FormatLedgerFiles(context.Background(), LedgerFormatOptions{Project: root, All: true})
	if err != nil || results[0].Changed {
		t.Fatalf("idempotent: %+v %v", results, err)
	}
	after, err = os.Stat(path)
	if err != nil || !after.ModTime().Equal(mark) {
		t.Fatal("unchanged rewritten")
	}
	for _, ids := range [][]string{{"missing"}, {"../escape"}, {""}} {
		if _, err := FormatLedgerFiles(context.Background(), LedgerFormatOptions{Project: root, TicketIDs: ids}); err == nil {
			t.Errorf("accepted %q", ids)
		}
	}
}
func TestFormatLedgerFilesAliasAndJournalSafety(t *testing.T) {
	root := t.TempDir()
	path := putFormatLedger(t, root, "T-1", formatFixture)
	alias := filepath.Join(root, "T-2")
	if err := os.Symlink(filepath.Dir(path), alias); err != nil {
		t.Fatal(err)
	}
	results, err := FormatLedgerFiles(context.Background(), LedgerFormatOptions{Project: root, TicketIDs: []string{"T-1", "T-1"}})
	if err != nil || len(results) != 1 {
		t.Fatalf("alias %+v %v", results, err)
	}
	outside := putFormatLedger(t, t.TempDir(), "outside", formatFixture)
	if err := os.Symlink(filepath.Dir(outside), filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	if _, err := FormatLedgerFiles(context.Background(), LedgerFormatOptions{Project: root, TicketIDs: []string{"escape"}}); err == nil || !strings.Contains(err.Error(), "outside") {
		t.Fatalf("escape: %v", err)
	}
	if err := beginTestPublication(t, path, []string{path}); err != nil {
		t.Fatal(err)
	}
	if err := capture.PublicationOwner(path).Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := FormatLedgerFiles(context.Background(), LedgerFormatOptions{Project: root, TicketIDs: []string{"T-1"}}); err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("pending publication: %v", err)
	}
}

// waitingFormatContext signals only when the real publication lock has observed
// contention. This makes the newer-baseline test independent of scheduler timing.
type waitingFormatContext struct {
	context.Context
	waiting chan struct{}
	once    sync.Once
}

func (c *waitingFormatContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.waiting) })
	return c.Context.Done()
}
func TestFormatLedgerFilesWaitsForPublisherBeforeReading(t *testing.T) {
	root := t.TempDir()
	path := putFormatLedger(t, root, "T-1", formatFixture)
	owner, err := capture.AcquirePublication(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	locked := true
	defer func() {
		if locked {
			if err := owner.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	deadline, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx := &waitingFormatContext{Context: deadline, waiting: make(chan struct{})}
	type outcome struct {
		results []LedgerFormatResult
		err     error
	}
	finished := make(chan outcome, 1)
	go func() {
		results, err := FormatLedgerFiles(ctx, LedgerFormatOptions{Project: root, TicketIDs: []string{"T-1"}})
		finished <- outcome{results, err}
	}()
	select {
	case <-ctx.waiting:
	case result := <-finished:
		t.Fatalf("formatter bypassed publication owner: %+v", result)
	case <-deadline.Done():
		t.Fatal("formatter did not attempt publication ownership")
	}
	newer := `{"ticket_id":"T-1","sessions":[],"unknown":{"newer":9007199254740995}}`
	if err := os.WriteFile(path, []byte(newer), 0640); err != nil {
		t.Fatal(err)
	}
	if err := owner.Close(); err != nil {
		t.Fatal(err)
	}
	locked = false
	select {
	case result := <-finished:
		if result.err != nil || len(result.results) != 1 || !result.results[0].Written {
			t.Fatalf("format after publisher: %+v", result)
		}
	case <-deadline.Done():
		t.Fatal("formatter did not finish after publisher release")
	}
	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, actual); err != nil {
		t.Fatal(err)
	}
	if compact.String() != newer {
		t.Fatalf("formatter overwrote newer facts: %s", actual)
	}
}
func TestFormatLedgerFilesRetiredSidecarAndSelectionIsolation(t *testing.T) {
	root := t.TempDir()
	selected := putFormatLedger(t, root, "T-1", formatFixture)
	unselectedBody := strings.Replace(formatFixture, `"ticket_id":"T-1"`, `"ticket_id":"T-2"`, 1)
	unselected := putFormatLedger(t, root, "T-2", unselectedBody)
	retired := filepath.Join(filepath.Dir(selected), "capture_report.json")
	summary := filepath.Join(filepath.Dir(selected), "SUMMARY.md")
	for _, path := range []string{retired, summary} {
		if err := os.WriteFile(path, []byte("historical evidence remains verbatim"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	results, err := FormatLedgerFiles(context.Background(), LedgerFormatOptions{Project: root, TicketIDs: []string{"T-1"}})
	if err != nil || len(results) != 1 || !results[0].Written {
		t.Fatalf("selected format: %+v %v", results, err)
	}
	for _, path := range []string{retired, summary, unselected} {
		actual, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		want := "historical evidence remains verbatim"
		if path == unselected {
			want = unselectedBody
		}
		if string(actual) != want {
			t.Errorf("unselected artifact changed: %s", path)
		}
	}
}
func TestFormatLedgerJSONUnknownPathNamesRemainExpanded(t *testing.T) {
	input := `{"ticket_id":"T","sessions":[],"sessions/[]/invocations":[{"tokens":{"x":1}}],"invocations":[{"tokens":{"x":2}}]}`
	actual, err := FormatLedgerJSON([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(actual, []byte("\"tokens\": {\n        \"x\": 1")) || !bytes.Contains(actual, []byte(`"tokens": {"x": 2}`)) {
		t.Fatalf("unknown key collided with structural path: %s", actual)
	}
}

func TestFormatLedgerFilesHardLinkRefusalPreservesBothTargets(t *testing.T) {
	root := t.TempDir()
	first := putFormatLedger(t, root, "T-1", formatFixture)
	second := filepath.Join(root, "T-2", "agentic_metrics.json")
	if err := os.MkdirAll(filepath.Dir(second), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(first, second); err != nil {
		t.Fatal(err)
	}
	results, err := FormatLedgerFiles(context.Background(), LedgerFormatOptions{Project: root, All: true})
	if err == nil || !strings.Contains(err.Error(), "hard-linked") || len(results) != 2 {
		t.Fatalf("hard link outcome: %+v %v", results, err)
	}
	for _, path := range []string{first, second} {
		body, err := os.ReadFile(path)
		if err != nil || string(body) != formatFixture {
			t.Fatalf("hard link rewritten: %s %v", path, err)
		}
	}
}

func TestFormatLedgerFilesReportsPartialLatePublicationFailure(t *testing.T) {
	root := t.TempDir()
	paths := []string{
		putFormatLedger(t, root, "A", formatFixture),
		putFormatLedger(t, root, "B", formatFixture),
		putFormatLedger(t, root, "C", formatFixture),
	}
	owner, err := capture.AcquirePublication(context.Background(), paths[2])
	if err != nil {
		t.Fatal(err)
	}
	locked := true
	defer func() {
		if locked {
			if err := owner.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	deadline, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx := &waitingFormatContext{Context: deadline, waiting: make(chan struct{})}
	type outcome struct {
		results []LedgerFormatResult
		err     error
	}
	finished := make(chan outcome, 1)
	go func() {
		results, err := FormatLedgerFiles(ctx, LedgerFormatOptions{Project: root, All: true})
		finished <- outcome{results, err}
	}()
	select {
	case <-ctx.waiting: // Sorted ownership proves A/B already passed their preflight.
	case result := <-finished:
		t.Fatalf("formatter bypassed held C ownership: %+v", result)
	case <-deadline.Done():
		t.Fatal("formatter did not reach C ownership")
	}
	preserved := filepath.Join(t.TempDir(), "preserved-B")
	if err := os.Rename(filepath.Dir(paths[1]), preserved); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Dir(paths[1]), []byte("staging blocked by regular file"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := owner.Close(); err != nil {
		t.Fatal(err)
	}
	locked = false
	var actual outcome
	select {
	case actual = <-finished:
	case <-deadline.Done():
		t.Fatal("formatter did not finish after C ownership released")
	}
	if actual.err == nil || !strings.Contains(actual.err.Error(), paths[1]) || len(actual.results) != 3 {
		t.Fatalf("missing late publication failure: %+v %v", actual.results, actual.err)
	}
	for i, result := range actual.results {
		if result.Path != paths[i] || !result.Changed {
			t.Fatalf("nondeterministic result %d: %+v", i, result)
		}
		if i == 1 {
			if result.Error == nil || result.Written {
				t.Fatalf("failed B reported as written: %+v", result)
			}
			continue
		}
		if result.Error != nil || !result.Written {
			t.Fatalf("successful sibling %d lost: %+v", i, result)
		}
		body, err := os.ReadFile(result.Path)
		if err != nil || string(body) == formatFixture || !json.Valid(body) {
			t.Fatalf("sibling %d not published: %s %v", i, body, err)
		}
	}
	body, err := os.ReadFile(filepath.Join(preserved, "agentic_metrics.json"))
	if err != nil || string(body) != formatFixture {
		t.Fatalf("failed B original damaged: %s %v", body, err)
	}
}
