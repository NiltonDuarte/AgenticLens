package capture

import (
	"encoding/json"
	"testing"
)

func TestCodexInstalledHookWireBudgets(t *testing.T) {
	for _, event := range []string{"SessionStart", "UserPromptSubmit", "SubagentStart", "SubagentStop", "Stop", "SessionEnd"} {
		t.Run(event, func(t *testing.T) {
			raw, err := codexHookGroup(event, "wrapper --al-offline")
			if err != nil {
				t.Fatal(err)
			}
			var decoded struct {
				Hooks []struct {
					Timeout int    `json:"timeout"`
					Command string `json:"command"`
					Limit   int    `json:"additionalContextLimit"`
				} `json:"hooks"`
			}
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatal(err)
			}
			if len(decoded.Hooks) != 1 || decoded.Hooks[0].Timeout != 2 || decoded.Hooks[0].Command != "wrapper --al-offline" {
				t.Fatalf("wrong installed wire contract: %s", raw)
			}
			expectedLimit := 0
			if event == "SessionStart" || event == "UserPromptSubmit" || event == "SubagentStart" {
				expectedLimit = 1000
			}
			if decoded.Hooks[0].Limit != expectedLimit {
				t.Fatalf("binding feedback limit changed: %s", raw)
			}
		})
	}
}
