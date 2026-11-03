package capture

import (
	"encoding/json"
	"testing"
)

func TestCodexInstalledHookWireBudgets(t *testing.T) {
	for _, platform := range []struct {
		name    string
		windows bool
		budget  int
	}{{"posix", false, 2}, {"windows", true, 5}} {
		for _, event := range []string{"SessionStart", "Stop"} {
			t.Run(platform.name+"/"+event, func(t *testing.T) {
				commandWindows := ""
				if platform.windows {
					commandWindows = "powershell.exe -File wrapper.ps1"
				}
				raw, err := codexHookGroup(event, "wrapper --al-offline", commandWindows, platform.windows)
				if err != nil {
					t.Fatal(err)
				}
				var decoded struct {
					Hooks []struct {
						Timeout        int    `json:"timeout"`
						WindowsCommand string `json:"commandWindows"`
						Limit          int    `json:"additionalContextLimit"`
					} `json:"hooks"`
				}
				if err := json.Unmarshal(raw, &decoded); err != nil {
					t.Fatal(err)
				}
				if len(decoded.Hooks) != 1 || decoded.Hooks[0].Timeout != platform.budget || decoded.Hooks[0].WindowsCommand != commandWindows {
					t.Fatalf("wrong installed wire contract: %s", raw)
				}
				expectedLimit := 0
				if event == "SessionStart" {
					expectedLimit = 1000
				}
				if decoded.Hooks[0].Limit != expectedLimit {
					t.Fatalf("binding feedback limit changed: %s", raw)
				}
			})
		}
	}
}
