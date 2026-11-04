package main_test

import (
	"agenticlens/internal/adapters"
	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/domain"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeGenerationPolicy(t *testing.T, f captureGenerationFixture, base []byte, policy domain.GenerationPolicy) {
	t.Helper()
	config := fmt.Sprintf("\ngeneration:\n  metrics:\n    skills: %t\n    delegations: %t\n    anomalies: %t\n", policy.Skills, policy.Delegations, policy.Anomalies)
	if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.yaml"), append(append([]byte{}, base...), []byte(config)...), 0600); err != nil {
		t.Fatal(err)
	}
}
func policyGenerate(t *testing.T, f captureGenerationFixture, summary string, sessions ...string) ([]byte, error) {
	t.Helper()
	session := "root"
	if len(sessions) > 0 {
		session = sessions[0]
	}
	cmd := exec.Command(cliBinaryPath, "generate", "--source", "codex", "--codex-home", f.store, "--capture-dir", f.state, "--session-id", session, "--ticket-id", "TEST-100", "--ticket-title", "Optional output", "--ticket-type", "feat", "--json-output", f.output, "--append", summary, "--anomalies")
	cmd.Dir = f.project
	return cmd.CombinedOutput()
}
func assertWrittenPolicy(t *testing.T, path string, policy domain.GenerationPolicy, owners int) {
	t.Helper()
	var root map[string]json.RawMessage
	if err := json.Unmarshal(readCaptureFile(t, path), &root); err != nil {
		t.Fatal(err)
	}
	var marker domain.GenerationPolicy
	if err := json.Unmarshal(root["generation_policy"], &marker); err != nil || marker != policy {
		t.Fatalf("wrong written policy: %+v %v", marker, err)
	}
	if _, exists := root["anomalies"]; exists != policy.Anomalies {
		t.Fatalf("root anomalies presence %t, expected %t", exists, policy.Anomalies)
	}
	var sessions []map[string]json.RawMessage
	if err := json.Unmarshal(root["sessions"], &sessions); err != nil {
		t.Fatal(err)
	}
	if len(sessions) != owners {
		t.Fatalf("policy removed session owners: got %d want %d", len(sessions), owners)
	}
	for _, session := range sessions {
		for name, enabled := range map[string]bool{"skills": policy.Skills, "delegations": policy.Delegations, "anomalies": policy.Anomalies} {
			_, exists := session[name]
			if !enabled && exists {
				t.Fatalf("disabled %s still written for %s: %s", name, session["id"], session[name])
			}
			if enabled && !exists {
				t.Fatalf("nonempty enabled %s disappeared for %s", name, session["id"])
			}
		}
	}
}

// Unlike the fixture's permissive json.Unmarshal helper, exercise canonical
// readers and the source-free API before accepting a regeneration result.
func canonicalPolicyLedger(t *testing.T, f captureGenerationFixture) domain.AgenticFactLedger {
	t.Helper()
	raw := readCaptureFile(t, f.output)
	if err := adapters.ValidateRuntimeFactLedger(raw); err != nil {
		t.Fatalf("successfully generated ledger rejected by runtime validator: %v", err)
	}
	if _, err := adapters.LoadExistingFactLedger(f.output); err != nil {
		t.Fatalf("successfully generated ledger cannot be reloaded: %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/workspaces/TEST-100", nil)
	request.SetPathValue("ticketId", "TEST-100")
	recorder := httptest.NewRecorder()
	adaptershttp.WorkspaceDetailHandler(filepath.Join(f.project, "records")).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("successfully generated ledger unavailable through API: %d %s", recorder.Code, recorder.Body.String())
	}
	facts := f.ledger(t)
	var detail struct {
		Ledger struct {
			Facts struct {
				Sessions []struct {
					ID           string          `json:"id"`
					Availability string          `json:"anomaly_availability"`
					Anomalies    json.RawMessage `json:"anomalies"`
				} `json:"sessions"`
			} `json:"facts"`
			Totals struct {
				ContextHealth string `json:"context_health"`
			} `json:"totals"`
		} `json:"ledger"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &detail); err != nil {
		t.Fatal(err)
	}
	expectedAvailability := map[string]string{}
	unavailable := false
	for _, owner := range facts.Sessions {
		expectedAvailability[owner.ID] = string(owner.AnomalyAvailability)
		unavailable = unavailable || owner.AnomalyAvailability == domain.AnomalyUnavailable
	}
	for _, owner := range detail.Ledger.Facts.Sessions {
		if owner.Availability != expectedAvailability[owner.ID] {
			t.Fatalf("API changed availability for %s", owner.ID)
		}
		if owner.Availability == "unavailable" && len(owner.Anomalies) > 0 {
			t.Fatalf("API fabricated measured anomaly state for %s: %s", owner.ID, owner.Anomalies)
		}
	}
	if unavailable && facts.EffectiveGenerationPolicy().Anomalies && detail.Ledger.Totals.ContextHealth != "UNAVAILABLE" {
		t.Fatalf("partially unavailable analysis claimed health %s", detail.Ledger.Totals.ContextHealth)
	}
	return facts
}

func TestGenerateOptionalMetricsOmissionRetainedOwnersAndSummary(t *testing.T) {
	for mask := 0; mask < 8; mask++ {
		t.Run(fmt.Sprintf("policy_%03b", mask), func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			f.output = filepath.Join(f.project, "records", "TEST-100", "agentic_metrics.json")
			source := f.write(t, 300000)
			occurrences := `{"timestamp":"2026-01-01T00:00:03Z","type":"response_item","payload":{"type":"function_call","call_id":"skill-call","name":"load_skill","arguments":"{\"name\":\"inspect-source\"}"}}
{"timestamp":"2026-01-01T00:00:04Z","type":"response_item","payload":{"type":"function_call","call_id":"spawn-call","name":"spawn_agent","arguments":"{\"agent_type\":\"tester\",\"task_name\":\"review-source\"}"}}
`
			if err := os.WriteFile(source, append(readCaptureFile(t, source), []byte(occurrences)...), 0600); err != nil {
				t.Fatal(err)
			}
			base := readCaptureFile(t, filepath.Join(f.project, ".agenticlens-config.yaml"))
			summary := filepath.Join(f.project, "SUMMARY.md")
			prose := "# Human summary\n\nKeep this user-authored paragraph.\n"
			if err := os.WriteFile(summary, []byte(prose), 0600); err != nil {
				t.Fatal(err)
			}
			initialPolicy := domain.GenerationPolicy{Skills: mask&1 != 0, Delegations: mask&2 != 0, Anomalies: mask&4 != 0}
			writeGenerationPolicy(t, f, base, initialPolicy)
			if out, err := policyGenerate(t, f, summary); err != nil {
				t.Fatalf("configured first creation failed: %v %s", err, out)
			}
			assertWrittenPolicy(t, f.output, initialPolicy, 1)
			writeGenerationPolicy(t, f, base, domain.DefaultGenerationPolicy())
			if out, err := policyGenerate(t, f, summary); err != nil {
				t.Fatalf("default creation failed: %v %s", err, out)
			}
			baseline := f.ledger(t)
			if len(baseline.SkillFacts()) == 0 || len(baseline.DelegationFacts()) == 0 || len(baseline.Sessions[0].Anomalies) == 0 {
				t.Fatalf("fixture lacks nonempty optional control facts: %+v", baseline)
			}
			root := baseline.Sessions[0]
			for _, id := range []string{"unavailable-same-source", "unavailable-child", "unavailable-other-source"} {
				owner := root
				owner.ID = id
				owner.Skills = append(append([]domain.SkillFact{}, root.Skills...), domain.SkillFact{Name: "retained-" + id}, domain.SkillFact{Name: "retained-" + id})
				owner.Delegations = append(append([]domain.DelegationFact{}, root.Delegations...), domain.DelegationFact{Description: "retained-" + id}, domain.DelegationFact{Description: "retained-" + id})
				toolName, count := "event-only-"+id, int64(3)
				owner.Anomalies = append(append([]domain.Anomaly{}, root.Anomalies...), domain.Anomaly{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning, Context: domain.AnomalyContext{ToolName: &toolName, Count: &count}})
				if id == "unavailable-child" {
					parent := "root"
					owner.ParentID = &parent
				}
				if id == "unavailable-other-source" {
					owner.Source = domain.SourceOpenCode
				}
				baseline.Sessions = append(baseline.Sessions, owner)
			}
			if err := adapters.SaveFactLedger(f.output, baseline); err != nil {
				t.Fatal(err)
			}
			baselineByID := map[string]domain.SessionFact{}
			for _, owner := range canonicalPolicyLedger(t, f).Sessions {
				baselineByID[owner.ID] = owner
			}
			policy := domain.GenerationPolicy{Skills: mask&1 != 0, Delegations: mask&2 != 0, Anomalies: mask&4 != 0}
			writeGenerationPolicy(t, f, base, policy)
			out, err := policyGenerate(t, f, summary)
			if err != nil {
				t.Fatalf("configured generation failed: %v %s", err, out)
			}
			assertWrittenPolicy(t, f.output, policy, 4)
			got := f.ledger(t)
			for _, owner := range got.Sessions {
				expected := baselineByID[owner.ID]
				if owner.Source != expected.Source {
					t.Fatalf("source attribution changed for %s", owner.ID)
				}
				if policy.Skills && !reflect.DeepEqual(owner.Skills, expected.Skills) {
					t.Fatalf("enabled skills/multiplicity changed for %s: got %+v want %+v", owner.ID, owner.Skills, expected.Skills)
				}
				if policy.Delegations && !reflect.DeepEqual(owner.Delegations, expected.Delegations) {
					t.Fatalf("enabled delegations/multiplicity changed for %s: got %+v want %+v", owner.ID, owner.Delegations, expected.Delegations)
				}
				if policy.Anomalies && owner.ID != "root" && !reflect.DeepEqual(owner.Anomalies, expected.Anomalies) {
					t.Fatalf("enabled retained anomalies changed for %s: got %+v want %+v", owner.ID, owner.Anomalies, expected.Anomalies)
				}
				if !reflect.DeepEqual(owner.Invocations, root.Invocations) || !reflect.DeepEqual(owner.Tools, root.Tools) {
					t.Fatalf("policy changed ordinary execution facts of %s", owner.ID)
				}
				if owner.ID == "unavailable-child" && (owner.ParentID == nil || *owner.ParentID != "root") {
					t.Fatal("delegation suppression removed child linkage")
				}
			}
			markdown := string(readCaptureFile(t, summary))
			if !strings.Contains(markdown, prose) {
				t.Fatal("managed summary replacement damaged user prose")
			}
			if !policy.Anomalies {
				for _, stale := range []string{"| Anomalies |", "HEALTHY", "CRITICAL", "Anomaly Breakdown"} {
					if strings.Contains(markdown, stale) {
						t.Fatalf("disabled summary retained %q", stale)
					}
				}
				if !strings.Contains(strings.ToLower(string(out)), "disabled") || strings.Contains(string(out), "No anomalies detected") {
					t.Fatalf("disabled diagnostic implied measured analysis: %s", out)
				}
			}
			first := f.ledger(t)
			if out, err = policyGenerate(t, f, summary); err != nil {
				t.Fatalf("repeat generation: %v %s", err, out)
			}
			assertWrittenPolicy(t, f.output, policy, 4)
			second := f.ledger(t)
			if !reflect.DeepEqual(first.Sessions, second.Sessions) || !reflect.DeepEqual(first.GenerationPolicy, second.GenerationPolicy) {
				t.Fatal("unchanged repeat changed owners/occurrences")
			}
			writeGenerationPolicy(t, f, base, domain.DefaultGenerationPolicy())
			if out, err = policyGenerate(t, f, summary); err != nil {
				t.Fatalf("re-enable generation: %v %s", err, out)
			}
			enabled := canonicalPolicyLedger(t, f)
			recoveredMarkdown := string(readCaptureFile(t, summary))
			if !strings.Contains(recoveredMarkdown, prose) || !strings.Contains(recoveredMarkdown, "### Anomalies & Health Flags") || strings.Contains(recoveredMarkdown, "Optional anomaly analysis: disabled") {
				t.Fatalf("re-enabled managed summary stale or damaged: %s", recoveredMarkdown)
			}
			if !strings.Contains(recoveredMarkdown, "CONTEXT_SATURATION_SPIKE") {
				t.Fatalf("re-enabled summary lost source-backed anomaly: %s", recoveredMarkdown)
			}
			for _, owner := range enabled.Sessions {
				if owner.ID == "root" && (len(owner.Skills) != 1 || len(owner.Delegations) != 1 || len(owner.Anomalies) == 0) {
					t.Fatalf("source-backed categories not recovered: %+v", owner)
				}
				if owner.ID == "root" && owner.AnomalyAvailability != "" {
					t.Fatal("fresh source-backed root marked unavailable")
				}
				if owner.ID != "root" && !policy.Anomalies {
					if owner.AnomalyAvailability != domain.AnomalyUnavailable || owner.Anomalies != nil {
						t.Fatalf("unavailable historical analysis became measured-empty: %+v", owner)
					}
					for _, anomaly := range owner.Anomalies {
						if anomaly.RuleID == domain.RuleToolRetryStreak && anomaly.Context.ToolName != nil && strings.HasPrefix(*anomaly.Context.ToolName, "event-only-") {
							t.Fatalf("fabricated unavailable event-only summary for %s", owner.ID)
						}
					}
				}
				if owner.ID != "root" && (!policy.Skills && len(owner.Skills) > 0 || !policy.Delegations && len(owner.Delegations) > 0) {
					t.Fatalf("unavailable omitted facts fabricated: %+v", owner)
				}
			}
			if out, err = policyGenerate(t, f, summary); err != nil {
				t.Fatalf("subsequent generate cannot read re-enabled ledger: %v %s", err, out)
			}
			canonicalPolicyLedger(t, f)
			if mask == 0 {
				// This owner was genuinely absent until this point. Once its
				// native source becomes readable, only that owner's marker clears.
				restoredID := "unavailable-same-source"
				restoredPath := filepath.Join(f.store, "sessions", restoredID+".jsonl")
				if _, err := os.Stat(restoredPath); !os.IsNotExist(err) {
					t.Fatalf("restoration fixture source was not unavailable: %v", err)
				}
				contents := strings.ReplaceAll(string(readCaptureFile(t, source)), `"id":"root"`, `"id":"`+restoredID+`"`)
				contents = strings.ReplaceAll(contents, `"thread_id":"root"`, `"thread_id":"`+restoredID+`"`)
				if err := os.WriteFile(restoredPath, []byte(contents), 0600); err != nil {
					t.Fatal(err)
				}
				if out, err = policyGenerate(t, f, summary, restoredID); err != nil {
					t.Fatalf("restored native source refresh failed: %v %s", err, out)
				}
				restored := canonicalPolicyLedger(t, f)
				for _, owner := range restored.Sessions {
					if owner.ID == restoredID && (owner.AnomalyAvailability != "" || len(owner.Anomalies) == 0 || len(owner.Skills) != 1 || len(owner.Delegations) != 1) {
						t.Fatalf("complete source refresh failed to restore recorded state: %+v", owner)
					}
					if (owner.ID == "unavailable-child" || owner.ID == "unavailable-other-source") && owner.AnomalyAvailability != domain.AnomalyUnavailable {
						t.Fatalf("restored another source changed unavailable owner's state: %+v", owner)
					}
				}
			}
		})
	}
}
func TestGenerateInvalidPolicyLeavesLedgerAndSummaryUntouched(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.write(t, 17)
	summary := filepath.Join(f.project, "SUMMARY.md")
	if err := os.WriteFile(summary, []byte("Human prose\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := policyGenerate(t, f, summary); err != nil {
		t.Fatalf("control: %v %s", err, out)
	}
	ledgerBefore, summaryBefore := readCaptureFile(t, f.output), readCaptureFile(t, summary)
	configPath := filepath.Join(f.project, ".agenticlens-config.yaml")
	config := append(readCaptureFile(t, configPath), []byte("\ngeneration:\n  metrics:\n    delegations: 'false'\n")...)
	if err := os.WriteFile(configPath, config, 0600); err != nil {
		t.Fatal(err)
	}
	out, err := policyGenerate(t, f, summary)
	if err == nil || !strings.Contains(string(out), "generation.metrics.delegations") {
		t.Fatalf("invalid boolean did not fail with setting path: %v %s", err, out)
	}
	if !reflect.DeepEqual(ledgerBefore, readCaptureFile(t, f.output)) || !reflect.DeepEqual(summaryBefore, readCaptureFile(t, summary)) {
		t.Fatal("configuration failure changed accepted output")
	}
}

func TestRegisteredGenerateHonorsDisabledAnomalies(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	source := f.write(t, 300000)
	f.bind(t, source)
	control := f.generate(t)
	if len(control.Sessions[0].Anomalies) == 0 {
		t.Fatal("registered control had no anomalies to suppress")
	}
	base := readCaptureFile(t, filepath.Join(f.project, ".agenticlens-config.yaml"))
	writeGenerationPolicy(t, f, base, domain.GenerationPolicy{Skills: true, Delegations: true})
	got := f.generate(t)
	if len(got.Sessions) != 1 || len(got.Sessions[0].Invocations) != 1 || got.Sessions[0].Invocations[0].Tokens.Input != 300000 {
		t.Fatalf("registered policy dropped core facts: %+v", got)
	}
	raw := readCaptureFile(t, f.output)
	var written struct {
		Anomalies json.RawMessage `json:"anomalies"`
		Sessions  []struct {
			Anomalies json.RawMessage `json:"anomalies"`
		} `json:"sessions"`
	}
	if err := json.Unmarshal(raw, &written); err != nil {
		t.Fatal(err)
	}
	if len(written.Anomalies) != 0 || len(written.Sessions[0].Anomalies) != 0 {
		t.Fatalf("registered path restored disabled anomaly facts: %s", raw)
	}
}

func TestOpenCodeGenerationPolicyAcrossSchemaFamilies(t *testing.T) {
	for _, family := range []string{"v1", "v2"} {
		t.Run(family, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			dbPath := filepath.Join(t.TempDir(), "source.db")
			db, err := sql.Open("sqlite", dbPath)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := db.Close(); err != nil {
					t.Error(err)
				}
			})
			var schema string
			if family == "v1" {
				schema = `CREATE TABLE session(id TEXT PRIMARY KEY,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER,directory TEXT);
CREATE TABLE message(id TEXT PRIMARY KEY,session_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);
CREATE TABLE part(id TEXT PRIMARY KEY,session_id TEXT,message_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);`
			} else {
				schema = `CREATE TABLE session_v2(id TEXT PRIMARY KEY,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER,directory TEXT);
CREATE TABLE session_message(id TEXT PRIMARY KEY,session_id TEXT,type TEXT,seq INTEGER,time_created INTEGER,time_updated INTEGER,data TEXT);`
			}
			if _, err := db.Exec(schema); err != nil {
				t.Fatal(err)
			}
			table := "session"
			if family == "v2" {
				table = "session_v2"
			}
			if _, err := db.Exec("INSERT INTO "+table+" VALUES('root',NULL,'coder','source-model','Optional output',1793231970000,1793231971000,?)", f.project); err != nil {
				t.Fatal(err)
			}
			skill := `{"type":"tool","id":"skill-call","name":"skill","state":{"status":"completed","input":{"name":"inspect-source"},"content":[{"type":"text","text":"loaded"}],"time":{"created":1793231970100,"completed":1793231970200}}}`
			task := `{"type":"tool","id":"task-call","name":"task","state":{"status":"completed","input":{"subagent_type":"tester","description":"review-source"},"content":[{"type":"text","text":"done"}],"time":{"created":1793231970300,"completed":1793231970400}}}`
			if family == "v1" {
				payload := `{"role":"assistant","modelID":"source-model","time":{"created":1793231970000,"completed":1793231971000},"cost":0.1,"tokens":{"input":300000,"output":2}}`
				if _, err := db.Exec(`INSERT INTO message VALUES('m','root',1793231970000,1793231971000,?)`, payload); err != nil {
					t.Fatal(err)
				}
				for index, part := range []string{skill, task} {
					// V1 carries the native tool name on tool rather than name.
					part = strings.Replace(part, `"name":"skill"`, `"tool":"skill"`, 1)
					part = strings.Replace(part, `"name":"task"`, `"tool":"task"`, 1)
					if _, err := db.Exec(`INSERT INTO part VALUES(?,'root','m',1793231970100,1793231970400,?)`, fmt.Sprint(index), part); err != nil {
						t.Fatal(err)
					}
				}
			} else {
				payload := `{"time":{"created":1793231970000,"completed":1793231971000},"agent":"coder","model":{"id":"source-model"},"cost":0.1,"tokens":{"input":300000,"output":2},"content":[` + skill + "," + task + `]}`
				if _, err := db.Exec(`INSERT INTO session_message VALUES('m','root','assistant',1,1793231970000,1793231971000,?)`, payload); err != nil {
					t.Fatal(err)
				}
			}
			generate := func() {
				t.Helper()
				cmd := exec.Command(cliBinaryPath, "generate", "--source", "opencode", "--db-path", dbPath, "--capture-dir", f.state, "--session-id", "root", "--ticket-id", "TEST-100", "--ticket-title", "Optional output", "--ticket-type", "feat", "--json-output", f.output)
				cmd.Dir = f.project
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("%s generation: %v %s", family, err, out)
				}
			}
			generate()
			control := f.ledger(t)
			if len(control.SkillFacts()) != 1 || len(control.DelegationFacts()) != 1 || len(control.Sessions[0].Anomalies) == 0 {
				t.Fatalf("%s control lacks optional source facts: %+v", family, control)
			}
			base := readCaptureFile(t, filepath.Join(f.project, ".agenticlens-config.yaml"))
			writeGenerationPolicy(t, f, base, domain.GenerationPolicy{})
			generate()
			assertWrittenPolicy(t, f.output, domain.GenerationPolicy{}, 1)
			got := f.ledger(t)
			if !reflect.DeepEqual(control.Sessions[0].Invocations, got.Sessions[0].Invocations) || !reflect.DeepEqual(control.Sessions[0].Tools, got.Sessions[0].Tools) {
				t.Fatalf("%s suppression changed ordinary facts", family)
			}
		})
	}
}
