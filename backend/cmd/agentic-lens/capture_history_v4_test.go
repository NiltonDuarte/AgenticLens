package main_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Older persisted history is now an explicit migration input, never an implicit
// normalization step in generate. Cost conversion remains covered by migration.
func TestCaptureLegacyHistoryRequiresExplicitMigrationWithoutRewriting(t *testing.T) {
	for _, version := range []int{1, 4, 5} {
		t.Run(string(rune('0'+version)), func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			path := f.write(t, 10)
			f.bind(t, path)
			f.generate(t)
			var body map[string]json.RawMessage
			if err := json.Unmarshal(readCaptureFile(t, f.output), &body); err != nil {
				t.Fatal(err)
			}
			encodedVersion, err := json.Marshal(version)
			if err != nil {
				t.Fatal(err)
			}
			body["schema_version"] = encodedVersion
			legacy, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(f.output, legacy, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(cliBinaryPath, f.args(t)...)
			cmd.Dir = f.project
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(strings.ToLower(string(out)), "migration") {
				t.Fatalf("old history accepted without explicit migration: %v %s", err, out)
			}
			if string(readCaptureFile(t, f.output)) != string(legacy) {
				t.Fatal("failed old-schema read rewrote ledger")
			}
		})
	}
}
