package http_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"agenticlens/internal/adapters/config"
	adaptershttp "agenticlens/internal/adapters/http"
	"gopkg.in/yaml.v3"
)

func TestProjectConfigHTTPPreservesPollingPolicyPresence(t *testing.T) {
	original, err := os.ReadFile("../config/testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, value string
		want        *int
	}{
		{"configured-nonpreset", "333", intPointer(333)},
		{"absent-field", "", nil},
		{"source-free-absent-live", "<absent-live>", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			var doc yaml.Node
			if err := yaml.Unmarshal(original, &doc); err != nil {
				t.Fatal(err)
			}
			mapping := doc.Content[0]
			for i := 0; i < len(mapping.Content); i += 2 {
				if mapping.Content[i].Value != "live" {
					continue
				}
				if tc.value == "<absent-live>" {
					mapping.Content = append(mapping.Content[:i], mapping.Content[i+2:]...)
					break
				}
				liveNode := mapping.Content[i+1]
				for j := 0; j < len(liveNode.Content); j += 2 {
					if liveNode.Content[j].Value != "poll_interval_ms" {
						continue
					}
					if tc.value == "" {
						liveNode.Content = append(liveNode.Content[:j], liveNode.Content[j+2:]...)
					} else {
						liveNode.Content[j+1].Value = tc.value
					}
					break
				}
				break
			}
			policy, err := yaml.Marshal(&doc)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), policy, 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := config.LoadShared(root)
			if err != nil {
				t.Fatalf("unused polling policy blocked shared read: %v", err)
			}
			workspace := filepath.Join(root, "records")
			ticketPath := filepath.Join(workspace, "T-02", "agentic_metrics.json")
			if err := os.MkdirAll(filepath.Dir(ticketPath), 0700); err != nil {
				t.Fatal(err)
			}
			ledger := []byte(sampleTestLedgerJSON("T-02", "Portable recorded ticket", 12, 0.25, 1))
			if err := os.WriteFile(ticketPath, ledger, 0600); err != nil {
				t.Fatal(err)
			}
			server := adaptershttp.NewServerWithConfig(workspace, fstest.MapFS{"index.html": {Data: []byte("dashboard")}}, cfg)
			ts := httptest.NewServer(server.Handler())
			defer ts.Close()
			response, err := ts.Client().Get(ts.URL + "/api/config")
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusOK {
				t.Fatalf("config status %d", response.StatusCode)
			}
			var wire map[string]json.RawMessage
			if err := json.NewDecoder(response.Body).Decode(&wire); err != nil {
				t.Fatal(err)
			}
			raw, present := wire["live_poll_interval_ms"]
			if present != (tc.want != nil) {
				t.Fatalf("poll presence changed: field=%s present=%t want=%v", raw, present, tc.want)
			}
			if present {
				var value int
				if err := json.Unmarshal(raw, &value); err != nil {
					t.Fatal(err)
				}
				if value != *tc.want {
					t.Fatalf("configured cadence replaced: got=%d want=%d", value, *tc.want)
				}
			}
			// Neither missing nor unusable unused live policy may hide saved history.
			response, err = ts.Client().Get(ts.URL + "/api/workspaces/T-02")
			if err != nil {
				t.Fatal(err)
			}
			body, readErr := io.ReadAll(response.Body)
			if err := response.Body.Close(); err != nil {
				t.Fatal(err)
			}
			if readErr != nil || response.StatusCode != http.StatusOK || !bytes.Contains(body, []byte(`"ticket_id":"T-02"`)) {
				t.Fatalf("persisted history unavailable: status=%d err=%v body=%s", response.StatusCode, readErr, body)
			}
			if after, err := os.ReadFile(ticketPath); err != nil || !bytes.Equal(after, ledger) {
				t.Fatalf("read changed portable ledger: %v", err)
			}
			if _, err := os.Stat(filepath.Join(root, ".agenticlens-config.local.yaml")); !os.IsNotExist(err) {
				t.Fatalf("unused local policy accessed/created: %v", err)
			}
		})
	}
	// Invalid supplied YAML values remain errors, never an invented cadence.
	for _, value := range []string{"0", "-7", "null", "'333'", "1.5"} {
		root := t.TempDir()
		body := strings.Replace(string(original), "poll_interval_ms: 333", "poll_interval_ms: "+value, 1)
		if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := config.LoadShared(root); err == nil || !strings.Contains(err.Error(), "live.poll_interval_ms") {
			t.Fatalf("invalid supplied cadence accepted: value=%s err=%v", value, err)
		}
	}
}

func TestConfigHTTPDoesNotHideExplicitZeroPollingValue(t *testing.T) {
	cfg := httpFixturePolicy()
	cfg.Live.PollIntervalMs = 0
	cfg.Settings = map[string]bool{"live.poll_interval_ms": true}
	response := httptest.NewRecorder()
	adaptershttp.ConfigHandler(cfg, "records")(response, httptest.NewRequest(http.MethodGet, "/api/config", nil))
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &wire); err != nil {
		t.Fatal(err)
	}
	if string(wire["live_poll_interval_ms"]) != "0" {
		t.Fatalf("projection replaced/omitted explicit value: %s", response.Body.String())
	}
}

func intPointer(value int) *int { return &value }
