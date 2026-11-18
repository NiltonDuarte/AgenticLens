package live_test

import (
	"agenticlens/internal/adapters/live"
	"agenticlens/internal/domain"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestFIX19HydrationRejectsAmbiguousChildAndKeepsUnambiguousParent(t *testing.T) {
	for _, variant := range []string{"identical-jsonl", "divergent-gzip", "warning-gzip"} {
		t.Run(variant, func(t *testing.T) {
			home := t.TempDir()
			active := filepath.Join(home, "sessions")
			archive := filepath.Join(home, "archived_sessions")
			for _, dir := range []string{active, archive} {
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
			}
			source := func(id, parent string) []byte {
				return []byte(fmt.Sprintf(`{"type":"session_meta","payload":{"id":%q,"parent_thread_id":%q}}
{"timestamp":"2026-01-01T00:00:01Z","type":"response_item","payload":{"type":"function_call","call_id":"one","name":"load_skill","arguments":"{\"name\":\"repeat\"}"}}
{"timestamp":"2026-01-01T00:00:01Z","type":"response_item","payload":{"type":"function_call","call_id":"two","name":"load_skill","arguments":"{\"name\":\"repeat\"}"}}
`, id, parent))
			}
			for _, id := range []string{"root", "child"} {
				parent := ""
				if id == "child" {
					parent = "root"
				}
				if err := os.WriteFile(filepath.Join(active, id+".jsonl"), source(id, parent), 0600); err != nil {
					t.Fatal(err)
				}
			}
			data := source("child", "root")
			if strings.HasPrefix(variant, "divergent") {
				data = source("child", "other")
			}
			if strings.HasPrefix(variant, "warning") {
				data = append(data, []byte("{malformed\n")...)
			}
			if strings.HasSuffix(variant, "gzip") {
				file, err := os.Create(filepath.Join(archive, "copy.jsonl.gz"))
				if err != nil {
					t.Fatal(err)
				}
				gz := gzip.NewWriter(file)
				if _, err = gz.Write(data); err != nil {
					t.Fatal(err)
				}
				if err = gz.Close(); err != nil {
					t.Fatal(err)
				}
				if err = file.Close(); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(filepath.Join(archive, "copy.jsonl"), data, 0600); err != nil {
				t.Fatal(err)
			}
			parent := "root"
			facts := domain.AgenticFactLedger{Sessions: []domain.SessionFact{{ID: "root", Source: domain.SourceCodex, User: "tester"}, {ID: "child", ParentID: &parent, Source: domain.SourceCodex, User: "tester", Skills: []domain.SkillFact{{Name: "stored"}, {Name: "stored"}}}}}
			before, err := json.Marshal(facts)
			if err != nil {
				t.Fatal(err)
			}
			service := live.NewLiveServiceWithCodexHome(nil, testProjectConfig(), home)
			defer service.Close()
			var first domain.WorkspaceEventsResponse
			for repeat := 0; repeat < 2; repeat++ {
				got, err := service.HydrateWorkspaceEvents(context.Background(), "TEST-1", domain.AgenticMetricsLedger{Facts: facts})
				if err != nil {
					t.Fatal(err)
				}
				if got.Availability != domain.TracePartial || !reflect.DeepEqual(got.Coverage.CompleteSessionIDs, []string{"root"}) || len(got.Events) != 2 || !strings.Contains(strings.Join(got.Coverage.Diagnostics, " "), "ambiguous") {
					t.Fatalf("ambiguous child certified or healthy parent hidden: %+v", got)
				}
				for _, event := range got.Events {
					if event.SessionID != "root" {
						t.Fatalf("ambiguous source event exposed: %+v", event)
					}
				}
				if repeat == 0 {
					first = got
				} else if !reflect.DeepEqual(first, got) {
					t.Fatal("unchanged hydration changed output")
				}
			}
			after, err := json.Marshal(facts)
			if err != nil {
				t.Fatal(err)
			}
			if string(after) != string(before) {
				t.Fatal("hydration changed accepted facts")
			}
		})
	}
}
