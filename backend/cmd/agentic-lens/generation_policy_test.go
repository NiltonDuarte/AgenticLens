package main_test

import (
	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/domain"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
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
	if _, exists := root["generation_policy"]; exists {
		t.Fatal("persisted generation policy reintroduced")
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
		if _, exists := session["anomaly_availability"]; exists {
			t.Fatal("persisted availability reintroduced")
		}
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

// Exercise canonical readers and the source-free API before accepting a
// regeneration; nil collections must remain absent throughout both boundaries.
func canonicalPolicyLedger(t *testing.T, f captureGenerationFixture) domain.AgenticFactLedger {
	t.Helper()
	raw := readCaptureFile(t, f.output)
	if err := adapters.ValidateRuntimeFactLedger(raw); err != nil {
		t.Fatalf("generated ledger rejected canonically: %v", err)
	}
	if _, err := adapters.LoadExistingFactLedger(f.output); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/workspaces/TEST-100", nil)
	request.SetPathValue("ticketId", "TEST-100")
	recorder := httptest.NewRecorder()
	adaptershttp.WorkspaceDetailHandler(filepath.Join(f.project, "records")).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("generated ledger unavailable through API: %d %s", recorder.Code, recorder.Body.String())
	}
	var detail struct {
		Ledger struct {
			Facts struct {
				GenerationPolicy json.RawMessage `json:"generation_policy"`
				Anomalies        json.RawMessage `json:"anomalies"`
				Sessions         []struct {
					ID           string          `json:"id"`
					Skills       json.RawMessage `json:"skills"`
					Delegations  json.RawMessage `json:"delegations"`
					Anomalies    json.RawMessage `json:"anomalies"`
					Availability json.RawMessage `json:"anomaly_availability"`
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
	if len(detail.Ledger.Facts.GenerationPolicy) != 0 {
		t.Fatal("API emitted obsolete policy")
	}
	facts := f.ledger(t)
	owners := map[string]domain.SessionFact{}
	incomplete := facts.Anomalies == nil
	for _, owner := range facts.Sessions {
		owners[owner.ID] = owner
		incomplete = incomplete || owner.Anomalies == nil
	}
	for _, owner := range detail.Ledger.Facts.Sessions {
		expected := owners[owner.ID]
		if len(owner.Availability) != 0 {
			t.Fatal("API emitted obsolete availability")
		}
		if (len(owner.Skills) > 0) != (expected.Skills != nil) || (len(owner.Delegations) > 0) != (expected.Delegations != nil) || (len(owner.Anomalies) > 0) != (expected.Anomalies != nil) {
			t.Fatalf("API changed absent/recorded shape for %s", owner.ID)
		}
	}
	if incomplete && detail.Ledger.Totals.ContextHealth != "UNAVAILABLE" {
		t.Fatalf("incomplete recorded analysis claimed health %s", detail.Ledger.Totals.ContextHealth)
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
			if !reflect.DeepEqual(first.Sessions, second.Sessions) {
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
				if owner.ID == "root" && owner.Anomalies == nil {
					t.Fatal("fresh source-backed root marked unavailable")
				}
				if owner.ID != "root" && !policy.Anomalies {
					if owner.Anomalies != nil {
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
					if owner.ID == restoredID && (len(owner.Anomalies) == 0 || len(owner.Skills) != 1 || len(owner.Delegations) != 1) {
						t.Fatalf("complete source refresh failed to restore recorded state: %+v", owner)
					}
					if (owner.ID == "unavailable-child" || owner.ID == "unavailable-other-source") && owner.Anomalies != nil {
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

func assertRecordedEmptyGeneration(t *testing.T, path string) {
	t.Helper()
	raw := readCaptureFile(t, path)
	if err := adapters.ValidateRuntimeFactLedger(raw); err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Anomalies json.RawMessage              `json:"anomalies"`
		Sessions  []map[string]json.RawMessage `json:"sessions"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	if string(wire.Anomalies) != "[]" {
		t.Fatalf("complete empty root analysis not recorded: %s", wire.Anomalies)
	}
	for _, owner := range wire.Sessions {
		for _, key := range []string{"skills", "delegations", "anomalies"} {
			if string(owner[key]) != "[]" {
				t.Fatalf("complete empty source %s not recorded: %s", key, owner[key])
			}
		}
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
	writeGenerationPolicy(t, f, base, domain.GenerationPolicy{})
	f.generate(t)
	assertWrittenPolicy(t, f.output, domain.GenerationPolicy{}, 1)
	f.write(t, 17)
	writeGenerationPolicy(t, f, base, domain.DefaultGenerationPolicy())
	f.generate(t)
	assertRecordedEmptyGeneration(t, f.output)
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
			// Re-enable against a complete supported source with no optional
			// occurrences; all three collections must now be explicit [].
			if family == "v1" {
				if _, err := db.Exec(`DELETE FROM part`); err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`UPDATE message SET data=replace(data,'300000','17')`); err != nil {
					t.Fatal(err)
				}
			} else {
				if _, err := db.Exec(`UPDATE session_message SET data=?`, `{"time":{"created":1793231970000,"completed":1793231971000},"agent":"coder","model":{"id":"source-model"},"cost":0.1,"tokens":{"input":17,"output":2},"content":[]}`); err != nil {
					t.Fatal(err)
				}
			}
			writeGenerationPolicy(t, f, base, domain.DefaultGenerationPolicy())
			generate()
			assertRecordedEmptyGeneration(t, f.output)
		})
	}
}

func TestGenerateOmissionPreservesActiveDurationAndDescendants(t *testing.T) {
	for _, family := range []string{"codex", "v1", "v2"} {
		for _, selection := range []string{"explicit", "registered"} {
			t.Run(family+"/"+selection, func(t *testing.T) {
				f := newCaptureGenerationFixture(t)
				base := readCaptureFile(t, filepath.Join(f.project, ".agenticlens-config.yaml"))
				writeGenerationPolicy(t, f, base, domain.GenerationPolicy{})
				const startMS int64 = 1793231970000
				dbPath := filepath.Join(t.TempDir(), "active-source.db")
				var db *sql.DB
				var table string
				source := domain.SourceOpenCode
				storePath := dbPath
				if family == "codex" {
					source = domain.SourceCodex
					storePath = f.store
				} else {
					var err error
					db, err = sql.Open("sqlite", dbPath)
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() {
						if err := db.Close(); err != nil {
							t.Error(err)
						}
					})
					schema := `CREATE TABLE session(id TEXT PRIMARY KEY,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER,directory TEXT);
CREATE TABLE message(id TEXT PRIMARY KEY,session_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);
CREATE TABLE part(id TEXT PRIMARY KEY,session_id TEXT,message_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);`
					table = "session"
					if family == "v2" {
						table = "session_v2"
						schema = `CREATE TABLE session_v2(id TEXT PRIMARY KEY,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER,directory TEXT);
CREATE TABLE session_message(id TEXT PRIMARY KEY,session_id TEXT,type TEXT,seq INTEGER,time_created INTEGER,time_updated INTEGER,data TEXT);`
					}
					if _, err := db.Exec(schema); err != nil {
						t.Fatal(err)
					}
				}
				for _, id := range []string{"root", "child"} {
					var parent *string
					if id == "child" {
						root := "root"
						parent = &root
					}
					if family == "codex" {
						parentSource := `"cli"`
						if id == "child" {
							parentSource = `{"subagent":{"thread_spawn":{"parent_thread_id":"root"}}}`
						}
						timestamp := time.UnixMilli(startMS).UTC().Format(time.RFC3339Nano)
						rollout := fmt.Sprintf(`{"timestamp":%q,"type":"session_meta","payload":{"id":%q,"cwd":%q,"source":%s}}
{"timestamp":%q,"type":"turn_context","payload":{"model":"source-model","turn_id":"active"}}
{"timestamp":%q,"type":"event_msg","payload":{"type":"task_started","turn_id":"active"}}
{"timestamp":%q,"type":"token_usage_record","payload":{"thread_id":%q,"turn_id":"active","usage":{"input_tokens":17,"output_tokens":2}}}
`, timestamp, id, f.project, parentSource, timestamp, timestamp, timestamp, id)
						if err := os.WriteFile(filepath.Join(f.store, "sessions", id+".jsonl"), []byte(rollout), 0600); err != nil {
							t.Fatal(err)
						}
						continue
					}
					if _, err := db.Exec("INSERT INTO "+table+" VALUES(?,?,'coder','source-model','Active optional output',?,?,?)", id, parent, startMS, startMS+1000, f.project); err != nil {
						t.Fatal(err)
					}
					if family == "v1" {
						payload := fmt.Sprintf(`{"role":"assistant","modelID":"source-model","time":{"created":%d},"tokens":{"input":17,"output":2}}`, startMS)
						if _, err := db.Exec(`INSERT INTO message VALUES(?,?,?,?,?)`, "message-"+id, id, startMS, startMS+1000, payload); err != nil {
							t.Fatal(err)
						}
						part := fmt.Sprintf(`{"type":"tool","tool":"bash","state":{"status":"running","time":{"start":%d}}}`, startMS)
						if _, err := db.Exec(`INSERT INTO part VALUES(?,?,?,?,?,?)`, "part-"+id, id, "message-"+id, startMS, startMS+1000, part); err != nil {
							t.Fatal(err)
						}
					} else {
						payload := fmt.Sprintf(`{"agent":"coder","model":{"id":"source-model"},"time":{"created":%d},"tokens":{"input":17,"output":2},"content":[{"type":"tool","id":"call","name":"bash","state":{"status":"running"}}]}`, startMS)
						if _, err := db.Exec(`INSERT INTO session_message VALUES(?,?,'assistant',1,?,?,?)`, "message-"+id, id, startMS, startMS+1000, payload); err != nil {
							t.Fatal(err)
						}
					}
				}
				args := []string{"generate", "--capture-dir", f.state, "--ticket-id", "TEST-100", "--ticket-title", "Active optional output", "--ticket-type", "feat", "--json-output", f.output}
				if family == "codex" {
					args = append(args, "--codex-home", f.store)
				} else {
					args = append(args, "--db-path", dbPath)
				}
				if selection == "explicit" {
					args = append(args, "--source", string(source), "--session-id", "root")
				} else {
					store, err := capture.OpenProject(f.state, f.project)
					if err != nil {
						t.Fatal(err)
					}
					identity := domain.CaptureIdentity{Source: source, Store: storePath, SessionID: "root"}
					transcript := ""
					if family == "codex" {
						transcript = filepath.Join(f.store, "sessions", "root.jsonl")
					}
					registration, registerErr := store.Register(context.Background(), identity, transcript, "session.created")
					if registerErr == nil {
						_, registerErr = store.Bind(context.Background(), identity, "TEST-100", registration.Request)
					}
					if registerErr == nil && family == "codex" {
						registerErr = store.QueueChild(context.Background(), f.store, filepath.Join(f.store, "sessions", "child.jsonl"), "root")
					}
					closeErr := store.Close()
					if registerErr != nil || closeErr != nil {
						t.Fatalf("register root: %v; close: %v", registerErr, closeErr)
					}
				}
				before := time.Now().UTC()
				cmd := exec.Command(cliBinaryPath, args...)
				cmd.Dir = f.project
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("generate: %v %s", err, out)
				}
				after := time.Now().UTC()
				assertWrittenPolicy(t, f.output, domain.GenerationPolicy{}, 2)
				if err := adapters.ValidateRuntimeFactLedger(readCaptureFile(t, f.output)); err != nil {
					t.Fatal(err)
				}
				var operationSeconds int64
				for _, owner := range f.ledger(t).Sessions {
					if owner.Source != source || len(owner.Invocations) != 1 || (owner.ID != "root" && owner.ID != "child") {
						t.Fatalf("optional output changed native owners: %+v", owner)
					}
					if owner.ID == "child" && (owner.ParentID == nil || *owner.ParentID != "root") {
						t.Fatal("delegation omission removed source-proven child linkage")
					}
					inv := owner.Invocations[0]
					if inv.CompletedAt != nil || inv.Tokens.Input != 17 {
						t.Fatalf("omission lost active numeric facts: %+v", inv)
					}
					lower := int64(math.Round(float64(before.UnixMilli()-startMS) / 1000))
					upper := int64(math.Round(float64(after.UnixMilli()-startMS) / 1000))
					if inv.ActiveDurationSeconds < lower || inv.ActiveDurationSeconds > upper || inv.ProvisionalSeconds != inv.ActiveDurationSeconds || inv.ProvisionalSeconds <= 0 {
						t.Fatalf("optional policy changed numeric duration: %+v bounds %d..%d", inv, lower, upper)
					}
					if operationSeconds != 0 && operationSeconds != inv.ProvisionalSeconds {
						t.Fatalf("one operation used inconsistent owner durations: %d and %d", operationSeconds, inv.ProvisionalSeconds)
					}
					operationSeconds = inv.ProvisionalSeconds
				}

				// Complete one exact source while its child is unavailable.
				// Both registered and explicit generation must retain the child's
				// recorded provisional contribution without advancing it.
				prior := f.ledger(t).Sessions
				var retained domain.SessionFact
				for _, owner := range prior {
					if owner.ID == "child" {
						retained = owner
					}
				}
				if family == "codex" {
					if err := os.Remove(filepath.Join(f.store, "sessions", "child.jsonl")); err != nil {
						t.Fatal(err)
					}
					path := filepath.Join(f.store, "sessions", "root.jsonl")
					body := readCaptureFile(t, path)
					terminal := fmt.Sprintf(`{"timestamp":%q,"type":"event_msg","payload":{"type":"task_complete","turn_id":"active"}}`+"\n", time.UnixMilli(startMS+30000).UTC().Format(time.RFC3339Nano))
					if err := os.WriteFile(path, append(body, []byte(terminal)...), 0600); err != nil {
						t.Fatal(err)
					}
				} else if family == "v1" {
					if _, err := db.Exec(`DELETE FROM part WHERE session_id='child'; DELETE FROM message WHERE session_id='child'; DELETE FROM session WHERE id='child'`); err != nil {
						t.Fatal(err)
					}
					payload := fmt.Sprintf(`{"role":"assistant","modelID":"source-model","time":{"created":%d,"completed":%d},"tokens":{"input":17,"output":2}}`, startMS, startMS+30000)
					if _, err := db.Exec(`UPDATE message SET data=?,time_updated=? WHERE session_id='root'`, payload, startMS+30000); err != nil {
						t.Fatal(err)
					}
					if _, err := db.Exec(`UPDATE part SET data='{"type":"tool","tool":"bash","state":{"status":"completed"}}' WHERE session_id='root'`); err != nil {
						t.Fatal(err)
					}
				} else {
					if _, err := db.Exec(`DELETE FROM session_message WHERE session_id='child'; DELETE FROM session_v2 WHERE id='child'`); err != nil {
						t.Fatal(err)
					}
					payload := fmt.Sprintf(`{"agent":"coder","model":{"id":"source-model"},"time":{"created":%d,"completed":%d},"tokens":{"input":17,"output":2},"content":[]}`, startMS, startMS+30000)
					if _, err := db.Exec(`UPDATE session_message SET data=?,time_updated=? WHERE session_id='root'`, payload, startMS+30000); err != nil {
						t.Fatal(err)
					}
				}
				var completed []domain.SessionFact
				for pass := 0; pass < 2; pass++ {
					cmd := exec.Command(cliBinaryPath, args...)
					cmd.Dir = f.project
					if out, err := cmd.CombinedOutput(); err != nil {
						t.Fatalf("completion with unavailable child: %v %s", err, out)
					}
					assertWrittenPolicy(t, f.output, domain.GenerationPolicy{}, 2)
					body := readCaptureFile(t, f.output)
					if strings.Contains(string(body), `"duration_evidence"`) || strings.Contains(string(body), `"duration_reasoning"`) {
						t.Fatal("generation emitted retired duration metadata")
					}
					owners := f.ledger(t).Sessions
					if len(owners) != 2 {
						t.Fatalf("completion lost/duplicated owners: %+v", owners)
					}
					for _, owner := range owners {
						if owner.ID == "child" {
							if !reflect.DeepEqual(owner, retained) {
								t.Fatalf("unavailable child's recorded numeric history changed: before=%+v after=%+v", retained, owner)
							}
						} else if owner.ID == "root" {
							if len(owner.Invocations) != 1 || owner.Invocations[0].ActiveDurationSeconds != 30 || owner.Invocations[0].ProvisionalSeconds != 0 || owner.Invocations[0].CompletedAt == nil {
								t.Fatalf("completion did not replace provisional snapshot: %+v", owner)
							}
						} else {
							t.Fatalf("unexpected owner: %+v", owner)
						}
					}
					if pass == 1 && !reflect.DeepEqual(owners, completed) {
						t.Fatal("unchanged completed refresh accumulated snapshots")
					}
					completed = owners
				}
			})
		}
	}
}
