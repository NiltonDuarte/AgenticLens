package opencode_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

func TestOpenCodeCheckpointPartialLifecycleRemainsQualified(t *testing.T) {
	const start int64 = 1793231970000
	for _, family := range []string{"v1", "v2"} {
		for _, tc := range []struct {
			name            string
			partial, future bool
			age             int64
		}{{"complete", false, false, 180000}, {"partial", true, false, 180000}, {"future_completion_short", false, true, 179999}, {"future_completion_long", false, true, 180000}, {"future_superseding_record", false, false, 180000}, {"future_superseding_short", false, false, 179999}} {
			partial := tc.partial
			t.Run(fmt.Sprintf("%s/%s", family, tc.name), func(t *testing.T) {
				var db *sql.DB
				var err error
				if family == "v1" {
					db = setupTestDB(t)
				} else {
					db, err = sql.Open("sqlite", ":memory:")
					if err != nil {
						t.Fatal(err)
					}
					createSchemaV2TestDB(t, db)
					if _, err = db.Exec(`DELETE FROM session_message;DELETE FROM session_v2`); err != nil {
						t.Fatal(err)
					}
				}
				defer db.Close()
				if family == "v1" {
					if _, err = db.Exec(`INSERT INTO session(id,agent,model,title,time_created,time_updated) VALUES('owner','coder','m','fixture',?,?);`, start-1000, start); err != nil {
						t.Fatal(err)
					}
					if partial {
						if _, err = db.Exec(`INSERT INTO message VALUES('earlier','owner',?,?,?)`, start-1000, start-1000, `{"role":"assistant","modelID":"m","tokens":{"input":7}}`); err != nil {
							t.Fatal(err)
						}
					}
					payload := fmt.Sprintf(`{"role":"assistant","modelID":"m","time":{"created":%d},"tokens":{"input":17}}`, start)
					if _, err = db.Exec(`INSERT INTO message VALUES('open','owner',?,?,?)`, start, start, payload); err != nil {
						t.Fatal(err)
					}
					if _, err = db.Exec(`INSERT INTO part VALUES('part','owner','open',?,?,?)`, start, start, `{"type":"tool","tool":"bash","state":{"status":"running"}}`); err != nil {
						t.Fatal(err)
					}
				} else {
					if _, err = db.Exec(`INSERT INTO session_v2 VALUES('owner',NULL,'coder','m','fixture',?,?)`, start-1000, start); err != nil {
						t.Fatal(err)
					}
					if partial {
						payload := fmt.Sprintf(`{"model":{"id":"m"},"time":{"created":%d},"tokens":{"input":7},"content":[]}`, start-1000)
						if _, err = db.Exec(`INSERT INTO session_message VALUES('earlier','owner','assistant',1,?,?,?)`, start-1000, start-1000, payload); err != nil {
							t.Fatal(err)
						}
					}
					payload := fmt.Sprintf(`{"model":{"id":"m"},"time":{"created":%d},"tokens":{"input":17},"content":[{"type":"tool","id":"call","name":"bash","state":{"status":"running"}}]}`, start)
					if _, err = db.Exec(`INSERT INTO session_message VALUES('open','owner','assistant',2,?,?,?)`, start, start, payload); err != nil {
						t.Fatal(err)
					}
				}
				if tc.future {
					if family == "v1" {
						payload := fmt.Sprintf(`{"role":"assistant","modelID":"m","time":{"created":%d,"completed":%d},"tokens":{"input":17}}`, start, start+200000)
						if _, err = db.Exec(`UPDATE message SET data=? WHERE id='open'`, payload); err != nil {
							t.Fatal(err)
						}
						if _, err = db.Exec(`DELETE FROM part`); err != nil {
							t.Fatal(err)
						}
					} else {
						payload := fmt.Sprintf(`{"model":{"id":"m"},"time":{"created":%d,"completed":%d},"tokens":{"input":17},"content":[]}`, start, start+200000)
						if _, err = db.Exec(`UPDATE session_message SET data=? WHERE id='open'`, payload); err != nil {
							t.Fatal(err)
						}
					}
				}
				if tc.name == "future_superseding_record" || tc.name == "future_superseding_short" {
					if family == "v1" {
						_, err = db.Exec(`INSERT INTO message VALUES('later','owner',?,?,?)`, start+200000, start+200000, `{"role":"user"}`)
					} else {
						_, err = db.Exec(`INSERT INTO session_message VALUES('later','owner','user',3,?,?,?)`, start+200000, start+200000, `{"text":"later"}`)
					}
					if err != nil {
						t.Fatal(err)
					}
				}
				cutoff := time.UnixMilli(start + tc.age)
				var source domain.SourceEvents
				sessions, _, _, err := opencode.ExtractSessionsStrictWithSourceAt(db, []string{"owner"}, "tester", nil, &source, cutoff)
				if err != nil {
					t.Fatal(err)
				}
				if len(sessions) != 1 || len(sessions[0].Invocations) != 1 || sessions[0].Invocations[0].ActiveDurationSeconds != 180 || sessions[0].Invocations[0].ProvisionalSeconds != 180 || sessions[0].Invocations[0].Tokens.Input != 17+7*int64(boolInt(partial)) {
					t.Fatalf("native facts changed: %+v", sessions)
				}
				got := domain.SelectCheckpointFeedback(sessions, source.CheckpointEvidence, domain.CheckpointCaller{}, cutoff, 180)
				counts := map[domain.CheckpointCase]int{}
				for _, item := range got {
					counts[item.Case]++
					if item.SessionID != "owner" || item.Model != "m" || item.InvocationRow != 1 {
						t.Fatalf("grouping lost owner/row/model: %+v", item)
					}
				}
				wantExact, wantQualified := 1, boolInt(partial)
				if tc.future || tc.name == "future_superseding_record" || tc.name == "future_superseding_short" {
					wantExact = 0
					wantQualified = 0
					if tc.age >= 180000 {
						wantQualified = 1
					}
				}
				if counts[domain.CheckpointUnknownCaller] != wantExact || counts[domain.CheckpointUnavailableInterval] != wantQualified || len(got) != wantExact+wantQualified {
					t.Fatalf("partial lifecycle was treated as complete interval proof: %+v", got)
				}
			})
		}
	}
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func TestOpenCodeV2CheckpointIdleBoundaryPreservesNumericFacts(t *testing.T) {
	const start int64 = 1793231970000
	for _, tc := range []struct {
		name    string
		column  bool
		idleSQL string
		age     int64
		want    domain.CheckpointCase
	}{
		{"absent_optional_column", false, "", 180000, domain.CheckpointUnknownCaller},
		{"active_null", true, "NULL", 180000, domain.CheckpointUnknownCaller},
		{"old_idle_before_new_interval", true, fmt.Sprint(start - 1), 180000, domain.CheckpointUnknownCaller},
		{"idle_equal_start", true, fmt.Sprint(start), 180000, domain.CheckpointUnavailableInterval},
		{"idle_after_start_before_cutoff", true, fmt.Sprint(start + 1000), 180000, domain.CheckpointUnavailableInterval},
		{"idle_equal_cutoff", true, fmt.Sprint(start + 180000), 180000, domain.CheckpointUnavailableInterval},
		{"idle_after_cutoff", true, fmt.Sprint(start + 200000), 180000, domain.CheckpointUnavailableInterval},
		{"malformed_text_idle", true, "'unsupported'", 180000, domain.CheckpointUnavailableInterval},
		{"fractional_idle", true, fmt.Sprintf("%d.5", start), 180000, domain.CheckpointUnavailableInterval},
		{"zero_idle", true, "0", 180000, domain.CheckpointUnavailableInterval},
		{"negative_idle", true, "-1", 180000, domain.CheckpointUnavailableInterval},
		{"known_short_native_idle", true, fmt.Sprint(start + 1000), 179999, ""},
		{"known_short_unusable_idle", true, "'unsupported'", 179999, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := sql.Open("sqlite", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			createSchemaV2TestDB(t, db)
			if _, err = db.Exec(`DELETE FROM session_message;DELETE FROM session_v2;INSERT INTO session_v2 VALUES('owner',NULL,'coder','m','fixture',?,?)`, start-1000, start); err != nil {
				t.Fatal(err)
			}
			if tc.column {
				if _, err = db.Exec(`ALTER TABLE session_v2 ADD COLUMN time_idle INTEGER`); err != nil {
					t.Fatal(err)
				}
				if _, err = db.Exec(`UPDATE session_v2 SET time_idle=` + tc.idleSQL); err != nil {
					t.Fatal(err)
				}
			}
			payload := fmt.Sprintf(`{"model":{"id":"m"},"time":{"created":%d},"tokens":{"input":17},"content":[{"type":"tool","id":"call","name":"bash","state":{"status":"running"}}]}`, start)
			if _, err = db.Exec(`INSERT INTO session_message VALUES('open','owner','assistant',1,?,?,?)`, start, start, payload); err != nil {
				t.Fatal(err)
			}
			cutoff := time.UnixMilli(start + tc.age)
			var source domain.SourceEvents
			sessions, _, _, err := opencode.ExtractSessionsStrictWithSourceAt(db, []string{"owner"}, "tester", nil, &source, cutoff)
			if err != nil {
				t.Fatalf("optional idle metadata blocked supported accounting: %v", err)
			}
			if len(sessions) != 1 || len(sessions[0].Invocations) != 1 {
				t.Fatalf("owner/model changed: %+v", sessions)
			}
			inv := sessions[0].Invocations[0]
			if inv.ActiveDurationSeconds != 180 || inv.ProvisionalSeconds != 180 || inv.Tokens.Input != 17 {
				t.Fatalf("idle qualification changed numeric facts: %+v", inv)
			}
			got := domain.SelectCheckpointFeedback(sessions, source.CheckpointEvidence, domain.CheckpointCaller{}, cutoff, 180)
			if tc.want == "" {
				if len(got) != 0 {
					t.Fatalf("known short interval fell back to feedback: %+v", got)
				}
				return
			}
			if len(got) != 1 || got[0].Case != tc.want || got[0].Source != domain.SourceOpenCode || got[0].SessionID != "owner" || got[0].Model != "m" || got[0].InvocationRow != 1 || got[0].ProvisionalSeconds != 180 {
				t.Fatalf("native idle currentness: got %+v want %s", got, tc.want)
			}
		})
	}
}
