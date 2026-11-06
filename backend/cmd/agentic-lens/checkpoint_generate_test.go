package main_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCheckpointGenerationSuccessAndPublicationFailureBoundaries(t *testing.T) {
	for _, mode := range []string{"default", "disabled", "custom_threshold", "stderr_unwritable", "publication_failure"} {
		t.Run(mode, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			timestamp := time.Now().UTC().Add(-10 * time.Minute).Format(time.RFC3339Nano)
			source := fmt.Sprintf(`{"timestamp":%q,"type":"session_meta","payload":{"id":"root","cwd":%q}}
{"timestamp":%q,"type":"turn_context","payload":{"model":"source-model","turn_id":"open"}}
{"timestamp":%q,"type":"event_msg","payload":{"type":"task_started","turn_id":"open"}}
`, timestamp, f.project, timestamp, timestamp)
			if err := os.WriteFile(filepath.Join(f.store, "sessions", "root.jsonl"), []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			if mode == "disabled" || mode == "custom_threshold" {
				threshold := 0
				if mode == "custom_threshold" {
					threshold = 86400
				}
				path := filepath.Join(f.project, ".agenticlens-config.yaml")
				raw := readCaptureFile(t, path)
				raw = append(raw, []byte(fmt.Sprintf("\ngeneration:\n  feedback:\n    provisional_threshold_seconds: %d\n", threshold))...)
				if err := os.WriteFile(path, raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			summary := filepath.Join(f.project, "SUMMARY.md")
			if mode == "publication_failure" {
				if err := os.Mkdir(summary, 0700); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command(cliBinaryPath, "generate", "--source", "codex", "--codex-home", f.store, "--capture-dir", f.state, "--session-id", "root", "--ticket-id", "TEST-100", "--ticket-title", "Checkpoint fixture", "--ticket-type", "feat", "--json-output", f.output, "--append", summary)
			cmd.Dir = f.project
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if mode == "stderr_unwritable" {
				path := filepath.Join(t.TempDir(), "readonly-stderr")
				if err := os.WriteFile(path, nil, 0600); err != nil {
					t.Fatal(err)
				}
				file, err := os.Open(path)
				if err != nil {
					t.Fatal(err)
				}
				defer file.Close()
				if n, err := file.Write([]byte("descriptor preflight")); err == nil || n != 0 {
					t.Fatalf("stderr descriptor unexpectedly writable: n=%d err=%v", n, err)
				}
				cmd.Stderr = file
			}
			err := cmd.Run()
			if strings.Contains(stdout.String(), "AgenticLens checkpoint:") {
				t.Fatalf("stderr feedback leaked to stdout: %s", stdout.String())
			}
			if mode == "publication_failure" {
				if err == nil || strings.Contains(stderr.String(), "AgenticLens checkpoint:") {
					t.Fatalf("failed publication emitted checkpoint success: %v %s", err, stderr.String())
				}
				return
			}
			if err != nil {
				t.Fatalf("feedback changed successful publication outcome: %v %s", err, stderr.String())
			}
			ledger := f.ledger(t)
			if len(ledger.Sessions) != 1 || ledger.Sessions[0].Invocations[0].ProvisionalSeconds < 600 {
				t.Fatalf("feedback changed supported facts: %+v", ledger)
			}
			if _, err := os.Stat(f.output + ".publication-pending"); !os.IsNotExist(err) {
				t.Fatalf("success left publication incomplete: %v", err)
			}
			if mode == "default" {
				if strings.Count(stderr.String(), "AgenticLens checkpoint:") != 1 || !strings.Contains(stderr.String(), "invoking caller is unverified") || strings.Contains(stderr.String(), "automatic next-turn") {
					t.Fatalf("false native caller/continuation: %s", stderr.String())
				}
			} else if strings.Contains(stderr.String(), "AgenticLens checkpoint:") {
				t.Fatalf("disabled/subthreshold feedback: %s", stderr.String())
			}
			for _, path := range []string{f.output, summary} {
				if strings.Contains(string(readCaptureFile(t, path)), "AgenticLens checkpoint:") {
					t.Fatalf("feedback persisted in %s", path)
				}
			}
		})
	}
}
