package main_test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFIX19GenerateRejectsCopiesWithoutChangingAcceptedPublication(t *testing.T) {
	for _, variant := range []string{"identical-jsonl", "identical-gzip", "divergent-jsonl", "warning-gzip"} {
		t.Run(variant, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			path := f.write(t, 17)
			writeManualStateOnlyConfig(t, f.project)
			body := readCaptureFile(t, path)
			for i := 0; i < 2; i++ {
				body = append(body, []byte(fmt.Sprintf(`{"timestamp":"2026-01-01T00:00:03Z","type":"response_item","payload":{"type":"function_call","call_id":"skill-%d","name":"load_skill","arguments":"{\"name\":\"repeated\"}"}}
{"timestamp":"2026-01-01T00:00:04Z","type":"response_item","payload":{"type":"function_call","call_id":"task-%d","name":"spawn_agent","arguments":"{\"agent_type\":\"tester\"}"}}
`, i, i))...)
			}
			if err := os.WriteFile(path, body, 0600); err != nil {
				t.Fatal(err)
			}
			run := func() ([]byte, error) {
				return directOwnedGenerate(t, f.project, f.store, filepath.Join(f.stateBase, "manual"), "root", f.output)
			}
			for repeat := 0; repeat < 2; repeat++ {
				if out, err := run(); err != nil {
					t.Fatalf("valid repeated source: %v %s", err, out)
				}
				ledger := f.ledger(t)
				if len(ledger.SkillFacts()) != 2 || len(ledger.DelegationFacts()) != 2 || len(ledger.InvocationFacts()) != 1 {
					t.Fatalf("real multiplicity changed: %+v", ledger.Sessions)
				}
			}
			accepted := readCaptureFile(t, f.output)
			if strings.HasPrefix(variant, "divergent") {
				body = bytes.Replace(body, []byte(`"input_tokens":17`), []byte(`"input_tokens":99`), 1)
			}
			if strings.HasPrefix(variant, "warning") {
				body = append(body, []byte("{malformed\n")...)
			}
			suffix := ".jsonl"
			if strings.HasSuffix(variant, "gzip") {
				suffix += ".gz"
			}
			copyPath := filepath.Join(f.store, "archived_sessions", "copy"+suffix)
			if err := os.MkdirAll(filepath.Dir(copyPath), 0700); err != nil {
				t.Fatal(err)
			}
			if strings.HasSuffix(copyPath, ".gz") {
				file, err := os.Create(copyPath)
				if err != nil {
					t.Fatal(err)
				}
				gz := gzip.NewWriter(file)
				if _, err = gz.Write(body); err != nil {
					t.Fatal(err)
				}
				if err = gz.Close(); err != nil {
					t.Fatal(err)
				}
				if err = file.Close(); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(copyPath, body, 0600); err != nil {
				t.Fatal(err)
			}
			for repeat := 0; repeat < 2; repeat++ {
				out, err := run()
				if err == nil || !strings.Contains(string(out), "ambiguous") {
					t.Fatalf("duplicate accepted without reason: %v %s", err, out)
				}
				if !bytes.Equal(accepted, readCaptureFile(t, f.output)) {
					t.Fatal("failed refresh changed accepted ledger")
				}
			}
		})
	}
}
