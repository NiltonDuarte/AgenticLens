package opencode_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

func TestOpenCodeMetadataOnlyModelIdentity(t *testing.T) {
	for _, family := range []string{"v1", "v2"} {
		for _, tc := range []struct {
			name  string
			value *string
			want  string
		}{
			{name: "null"}, {name: "empty", value: modelString("")}, {name: "blank", value: modelString("  ")},
			{name: "observed_default", value: modelString("default"), want: "default"},
			{name: "observed_unknown", value: modelString("unknown"), want: "unknown"},
			{name: "observed_named", value: modelString("native-model"), want: "native-model"},
			{name: "observed_object", value: modelString(`{"id":"native-model"}`), want: "native-model"},
		} {
			for _, messages := range []int{0, 2} {
				t.Run(fmt.Sprintf("%s/%s/messages%d", family, tc.name, messages), func(t *testing.T) {
					db, table := resolutionDB(t, family)
					resolutionSession(t, db, table, "owner", "", "model identity", 1000)
					if _, err := db.Exec(`UPDATE `+table+` SET model=? WHERE id='owner'`, tc.value); err != nil {
						t.Fatal(err)
					}
					for i := 0; i < messages; i++ {
						if family == "v1" {
							if _, err := db.Exec(`INSERT INTO message VALUES(?, 'owner',1500,1500,?)`, fmt.Sprint(i), `{"role":"user"}`); err != nil {
								t.Fatal(err)
							}
						} else {
							if _, err := db.Exec(`INSERT INTO session_message VALUES(?,'owner','user',?,1500,1500,'{}')`, fmt.Sprint(i), i); err != nil {
								t.Fatal(err)
							}
						}
					}
					var source domain.SourceEvents
					sessions, _, _, err := opencode.ExtractSessionsStrictWithSourceAt(db, []string{"owner"}, "tester", nil, &source, time.Unix(20, 0))
					if err != nil {
						t.Fatal(err)
					}
					if len(sessions) != 1 || len(sessions[0].Invocations) != 1 {
						t.Fatalf("metadata-backed aggregate lost: %+v", sessions)
					}
					inv := sessions[0].Invocations[0]
					if inv.Model != tc.want || inv.Turns != int64(messages) || inv.UserIterations != int64(messages) || inv.Compactions != 0 || inv.ActiveDurationSeconds != 0 || inv.Tokens != (domain.InvocationTokens{}) || inv.CostUSD != 0 {
						t.Fatalf("identity or zero-token counters changed: %+v", inv)
					}
					if len(source.CheckpointEvidence) != 1 || source.CheckpointEvidence[0].Model != tc.want {
						t.Fatalf("checkpoint identity changed: %+v", source.CheckpointEvidence)
					}
					persisted := saveReloadSourceOwners(t, sessions)
					if persisted[0].Invocations[0].Model != tc.want {
						t.Fatalf("ledger identity changed: %+v", persisted)
					}
				})
			}
		}
	}
}

func modelString(value string) *string { return &value }

func TestSchemaV1AssistantMissingModelAndObservedNames(t *testing.T) {
	for _, tc := range []struct{ name, member, want string }{
		{"absent", "", ""}, {"null", `,"modelID":null`, ""}, {"blank", `,"modelID":"  "`, ""},
		{"default", `,"modelID":"default"`, "default"}, {"unknown", `,"modelID":"unknown"`, "unknown"}, {"named", `,"modelID":"native-model"`, "native-model"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, table := resolutionDB(t, "v1")
			resolutionSession(t, db, table, "owner", "", "model identity", 1000)
			payload := `{"role":"assistant"` + tc.member + `,"time":{"created":2000,"completed":5000},"tokens":{"input":7,"output":3,"cache":{"read":2},"reasoning":1},"cost":0.25}`
			if _, err := db.Exec(`INSERT INTO message VALUES('m','owner',2000,5000,?)`, payload); err != nil {
				t.Fatal(err)
			}
			sessions, _, _, err := opencode.ExtractSessionsStrict(db, []string{"owner"}, "tester", nil)
			if err != nil {
				t.Fatal(err)
			}
			inv := sessions[0].Invocations[0]
			if len(sessions[0].Invocations) != 1 || inv.Model != tc.want || inv.Turns != 1 || inv.Tokens.Input != 7 || inv.Tokens.Output != 3 || inv.Tokens.CacheRead != 2 || inv.Tokens.Reasoning != 1 || inv.CostUSD != 0.25 || inv.ActiveDurationSeconds != 3 {
				t.Fatalf("assistant model/counters changed: %+v", sessions)
			}
		})
	}
}

func TestSchemaV1MissingModelDoesNotCollideWithObservedDefault(t *testing.T) {
	db, table := resolutionDB(t, "v1")
	resolutionSession(t, db, table, "owner", "", "mixed identity", 1000)
	for i, member := range []string{"", `,"modelID":"default"`, `,"modelID":"unknown"`} {
		payload := fmt.Sprintf(`{"role":"assistant"%s,"time":{"created":%d,"completed":%d},"tokens":{"input":7},"cost":0.25}`, member, 2000+i*4000, 5000+i*4000)
		if _, err := db.Exec(`INSERT INTO message VALUES(?,'owner',?,?,?)`, fmt.Sprint(i), 2000+i*4000, 5000+i*4000, payload); err != nil {
			t.Fatal(err)
		}
	}
	sessions, _, _, err := opencode.ExtractSessionsStrict(db, []string{"owner"}, "tester", nil)
	if err != nil {
		t.Fatal(err)
	}
	invs := sessions[0].Invocations
	if len(invs) != 3 {
		t.Fatalf("distinct model identities collapsed: %+v", invs)
	}
	for i, want := range []string{"", "default", "unknown"} {
		if invs[i].Model != want || invs[i].Turns != 1 || invs[i].Tokens.Input != 7 || invs[i].CostUSD != 0.25 || invs[i].ActiveDurationSeconds != 3 {
			t.Fatalf("mixed row changed: %+v", invs)
		}
	}
}

func TestSchemaV2AssistantStillRequiresNativeModelID(t *testing.T) {
	for _, model := range []string{`{}`, `{"model":null}`, `{"model":{}}`, `{"model":{"id":null}}`, `{"model":{"id":""}}`} {
		t.Run(model, func(t *testing.T) {
			db, table := resolutionDB(t, "v2")
			resolutionSession(t, db, table, "owner", "", "required assistant model", 1000)
			if _, err := db.Exec(`INSERT INTO session_message VALUES('m','owner','assistant',1,2000,2000,?)`, model); err != nil {
				t.Fatal(err)
			}
			if _, _, _, err := opencode.ExtractSessionsStrict(db, []string{"owner"}, "tester", nil); err == nil || !strings.Contains(err.Error(), "missing model.id") {
				t.Fatalf("v2 required model validation lost: %v", err)
			}
		})
	}
}
