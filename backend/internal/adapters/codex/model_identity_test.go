package codex_test

import (
	"fmt"
	"strings"
	"testing"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/domain"
)

func TestParseRolloutMissingAndObservedModelIdentity(t *testing.T) {
	for _, model := range []string{"", "  ", "default", "unknown", "native-model"} {
		for _, active := range []bool{false, true} {
			t.Run(fmt.Sprintf("%q/active%t", model, active), func(t *testing.T) {
				records := []string{`{"timestamp":"2026-10-25T09:00:00Z","type":"session_meta","payload":{"id":"owner"}}`}
				if model != "" {
					records = append(records, fmt.Sprintf(`{"timestamp":"2026-10-25T09:00:00Z","type":"turn_context","payload":{"turn_id":"turn-1","model":%q}}`, model))
				}
				if active {
					records = append(records,
						`{"timestamp":"2026-10-25T09:00:00Z","type":"response_item","payload":{"type":"message","role":"user","content":[]}}`,
						`{"timestamp":"2026-10-25T09:00:00Z","type":"event_msg","payload":{"type":"task_started","turn_id":"turn-1"}}`,
						`{"timestamp":"2026-10-25T09:00:01Z","type":"token_usage_record","payload":{"turn_id":"turn-1","model":"  ","usage":{"input_tokens":10,"cached_input_tokens":4,"output_tokens":2}}}`,
						`{"timestamp":"2026-10-25T09:00:02Z","type":"compacted","payload":{}}`,
						`{"timestamp":"2026-10-25T09:00:03Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"turn-1"}}`)
				}
				transcript, err := codex.ParseRollout(strings.NewReader(strings.Join(records, "\n")))
				if err != nil {
					t.Fatal(err)
				}
				if len(transcript.Invocations) != 1 {
					t.Fatalf("metadata-backed aggregate lost: %+v", transcript.Invocations)
				}
				inv := transcript.Invocations[0]
				want := strings.TrimSpace(model)
				if inv.Model != want || inv.CostEvidence.State != domain.CostUnavailable || inv.CostUSD != 0 || len(transcript.CheckpointEvidence) != 1 || transcript.CheckpointEvidence[0].Model != want {
					t.Fatalf("model/cost evidence changed: %+v", transcript)
				}
				if active {
					if inv.Turns != 1 || inv.UserIterations != 1 || inv.Compactions != 1 || inv.ActiveDurationSeconds != 3 || inv.Tokens.Input != 6 || inv.Tokens.CacheRead != 4 || inv.Tokens.Output != 2 {
						t.Fatalf("active counters changed: %+v", inv)
					}
				} else if inv.Turns != 0 || inv.UserIterations != 0 || inv.Compactions != 0 || inv.ActiveDurationSeconds != 0 || inv.Tokens != (domain.InvocationTokens{}) {
					t.Fatalf("zero-token aggregate changed: %+v", inv)
				}
			})
		}
	}
}

func TestParseRolloutObservedNamesAreNotProvisional(t *testing.T) {
	for _, model := range []string{"default", "unknown"} {
		t.Run(model, func(t *testing.T) {
			records := []string{
				fmt.Sprintf(`{"timestamp":"2026-10-25T09:00:00Z","type":"turn_context","payload":{"model":%q}}`, model),
				`{"timestamp":"2026-10-25T09:00:00Z","type":"event_msg","payload":{"type":"task_started","turn_id":"turn-1"}}`,
				`{"timestamp":"2026-10-25T09:00:01Z","type":"turn_context","payload":{"turn_id":"turn-1","model":"later-model"}}`,
				`{"timestamp":"2026-10-25T09:00:02Z","type":"turn_context","payload":{"turn_id":"turn-1","model":"  "}}`,
				`{"timestamp":"2026-10-25T09:00:03Z","type":"token_usage_record","payload":{"model":"later-model","turn_id":"turn-1","usage":{"input_tokens":7}}}`,
				`{"timestamp":"2026-10-25T09:00:04Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"turn-1"}}`,
			}
			transcript, err := codex.ParseRollout(strings.NewReader(strings.Join(records, "\n")))
			if err != nil {
				t.Fatal(err)
			}
			if transcript.Model != "later-model" || len(transcript.Invocations) != 2 {
				t.Fatalf("observed name or blank context changed grouping: %+v", transcript)
			}
			rows := map[string]domain.InvocationFact{}
			for _, inv := range transcript.Invocations {
				rows[inv.Model] = inv
			}
			if rows[model].Turns != 1 || rows[model].ActiveDurationSeconds != 4 || rows[model].Tokens.Input != 0 || rows["later-model"].Tokens.Input != 7 || rows["later-model"].ActiveDurationSeconds != 0 {
				t.Fatalf("known model reassigned as provisional: %+v", rows)
			}
		})
	}
}

func TestParseRolloutLateModelPreservesEarlierUsageAndDuration(t *testing.T) {
	records := []string{
		`{"timestamp":"2026-10-25T09:00:00Z","type":"event_msg","payload":{"type":"task_started","turn_id":"turn-1"}}`,
		`{"timestamp":"2026-10-25T09:00:01Z","type":"token_usage_record","payload":{"turn_id":"turn-1","usage":{"input_tokens":7}}}`,
		`{"timestamp":"2026-10-25T09:00:02Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"turn-1"}}`,
		`{"timestamp":"2026-10-25T09:00:03Z","type":"turn_context","payload":{"turn_id":"turn-1","model":"default"}}`,
		`{"timestamp":"2026-10-25T09:00:04Z","type":"token_usage_record","payload":{"turn_id":"turn-1","usage":{"input_tokens":11}}}`,
	}
	transcript, err := codex.ParseRollout(strings.NewReader(strings.Join(records, "\n")))
	if err != nil {
		t.Fatal(err)
	}
	if len(transcript.Invocations) != 2 {
		t.Fatalf("pre-context usage bucket lost: %+v", transcript.Invocations)
	}
	unknown, named := transcript.Invocations[0], transcript.Invocations[1]
	if unknown.Model != "" || unknown.Tokens.Input != 7 || unknown.ActiveDurationSeconds != 2 || unknown.Turns != 0 || named.Model != "default" || named.Tokens.Input != 11 || named.ActiveDurationSeconds != 0 || named.Turns != 1 || transcript.Usage.Input != 18 || transcript.ActiveSeconds != 2 {
		t.Fatalf("existing late association semantics changed: %+v", transcript)
	}
}

func TestParseRolloutUnknownModelTurnLookupUsesPresence(t *testing.T) {
	records := []string{
		`{"timestamp":"2026-10-25T09:00:00Z","type":"event_msg","payload":{"type":"task_started","turn_id":"first"}}`,
		`{"timestamp":"2026-10-25T09:00:01Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"first"}}`,
		`{"timestamp":"2026-10-25T09:00:02Z","type":"event_msg","payload":{"type":"task_started","turn_id":"second"}}`,
		`{"timestamp":"2026-10-25T09:00:02Z","type":"event_msg","payload":{"type":"exec_command_begin","turn_id":"first","call_id":"command"}}`,
		`{"timestamp":"2026-10-25T09:00:03Z","type":"event_msg","payload":{"type":"exec_command_end","turn_id":"first","call_id":"command","exit_code":0}}`,
		`{"timestamp":"2026-10-25T09:00:05Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"second"}}`,
	}
	transcript, err := codex.ParseRollout(strings.NewReader(strings.Join(records, "\n")))
	if err != nil {
		t.Fatal(err)
	}
	if len(transcript.Invocations) != 1 || transcript.Invocations[0].Model != "" || transcript.Invocations[0].Turns != 2 || transcript.Invocations[0].ActiveDurationSeconds != 4 || len(transcript.Events) != 1 || transcript.Events[0].Turn != 1 || transcript.ToolCalls != 1 {
		t.Fatalf("empty model confused with absent turn: %+v", transcript)
	}
}
