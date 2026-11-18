package backfill

import (
	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
	"compress/gzip"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestFIX19BackfillRetainsAmbiguousAndUnavailableOwnersWhileRefreshingHealthy(t *testing.T) {
	for _, variant := range []string{"identical-jsonl", "divergent-gzip", "warning-gzip"} {
		t.Run(variant, func(t *testing.T) {
			project := newBackfillProject(t)
			ticket := "FIX-19"
			preserved := []domain.SessionFact{
				{ID: "root", Source: domain.SourceCodex, User: "developer-a", Tools: []domain.SessionToolFact{{Name: "accepted", Calls: 7}}, Skills: []domain.SkillFact{{Name: "repeat"}, {Name: "repeat"}}, Delegations: []domain.DelegationFact{{ChildSession: "unavailable-child"}, {ChildSession: "unavailable-child"}}},
				{ID: "missing", Source: domain.SourceCodex, User: "developer-a", Tools: []domain.SessionToolFact{{Name: "missing", Calls: 9}}},
				{ID: "opencode", Source: domain.SourceOpenCode, User: "developer-b", Tools: []domain.SessionToolFact{{Name: "opencode", Calls: 11}}},
			}
			sessions := append(append([]domain.SessionFact{}, preserved...), domain.SessionFact{ID: "healthy", Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "stale", Calls: 2}}})
			schema := domain.CurrentSchemaVersion
			createTestWorkspace(t, project, ticket, domain.AgenticFactLedger{SchemaVersion: &schema, TicketID: &ticket, Sessions: sessions})
			ledgerPath := filepath.Join(project, ticket, "agentic_metrics.json")
			before, err := adapters.LoadExistingFactLedger(ledgerPath)
			if err != nil {
				t.Fatal(err)
			}
			home := t.TempDir()
			active := filepath.Join(home, "sessions")
			if err = os.MkdirAll(active, 0700); err != nil {
				t.Fatal(err)
			}
			path := writeBackfillRollout(t, active, "root", "", "coder", "model", map[string]int{"duplicate": 1}, "")
			writeBackfillRollout(t, active, "healthy", "", "coder", "model", map[string]int{"fresh": 3}, "")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(variant, "divergent") {
				data = []byte(strings.ReplaceAll(string(data), "duplicate", "conflicting"))
			}
			if strings.HasPrefix(variant, "warning") {
				data = append(data, []byte("{malformed\n")...)
			}
			archive := filepath.Join(home, "archived_sessions")
			if err = os.MkdirAll(archive, 0700); err != nil {
				t.Fatal(err)
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
			} else if err = os.WriteFile(filepath.Join(archive, "copy.jsonl"), data, 0600); err != nil {
				t.Fatal(err)
			}
			for repeat := 0; repeat < 2; repeat++ {
				report, err := NewEngine().Execute(BackfillOptions{Project: project, TicketID: ticket, Source: "codex", CodexHome: home, RecomputeDB: true, Force: true})
				if err != nil {
					t.Fatal(err)
				}
				if len(report.Results) != 1 || !report.Results[0].RefreshIncomplete || !strings.Contains(strings.Join(report.Results[0].Diagnostics, " "), "ambiguous") {
					t.Fatalf("ambiguity lost: %+v", report.Results)
				}
				after, err := adapters.LoadExistingFactLedger(ledgerPath)
				if err != nil {
					t.Fatal(err)
				}
				if len(after.Sessions) != 4 {
					t.Fatalf("owners changed: %+v", after.Sessions)
				}
				for _, want := range before.Sessions {
					var got *domain.SessionFact
					for i := range after.Sessions {
						if after.Sessions[i].ID == want.ID {
							got = &after.Sessions[i]
						}
					}
					if got == nil {
						t.Fatalf("missing owner %s", want.ID)
					}
					if want.ID == "healthy" {
						if len(got.Tools) != 1 || got.Tools[0].Name != "fresh" || got.Tools[0].Calls != 3 {
							t.Fatalf("healthy owner hidden/doubled: %+v", got)
						}
					} else if !reflect.DeepEqual(*got, want) {
						t.Fatalf("accepted owner %s changed: got=%+v want=%+v", want.ID, *got, want)
					}
				}
			}
		})
	}
}
