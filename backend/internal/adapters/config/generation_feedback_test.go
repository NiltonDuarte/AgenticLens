package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	config "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestGenerationFeedbackThresholdConfigAndWriter(t *testing.T) {
	for _, format := range []string{"yaml", "json"} {
		for _, tc := range []struct {
			name, value string
			want        int64
			valid       bool
		}{
			{"absent", "", 180, true}, {"disabled", "0", 0, true}, {"custom", "17", 17, true}, {"large", "9223372036854775807", 9223372036854775807, true},
			{"negative", "-1", 0, false}, {"fractional", "1.5", 0, false}, {"float_integer", "180.0", 0, false}, {"null", "null", 0, false}, {"quoted", `"180"`, 0, false}, {"overflow", "9223372036854775808", 0, false}, {"boolean", "true", 0, false}, {"sequence", "[]", 0, false}, {"mapping", "{}", 0, false},
		} {
			t.Run(format+"/"+tc.name, func(t *testing.T) {
				root := t.TempDir()
				path := filepath.Join(root, ".agenticlens-config."+format)
				generation := ""
				if tc.value != "" {
					generation = "generation:\n  metrics:\n    skills: false\n  feedback:\n    provisional_threshold_seconds: " + tc.value + "\n"
				}
				input := "project_name: Feedback\n" + generation + testValidSuffixYAML
				if format == "json" {
					generation = ""
					if tc.value != "" {
						generation = fmt.Sprintf(`,"generation":{"metrics":{"skills":false},"feedback":{"provisional_threshold_seconds":%s}}`, tc.value)
					}
					input = `{"project_name":"Feedback"` + generation + testValidSuffixJSON
				}
				if err := os.WriteFile(path, []byte(input), 0600); err != nil {
					t.Fatal(err)
				}
				cfg, err := config.Load(root)
				if !tc.valid {
					if err == nil || !strings.Contains(err.Error(), "generation.feedback.provisional_threshold_seconds") {
						t.Fatalf("invalid threshold lacked setting path: %v", err)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if cfg.Generation.Feedback.ProvisionalThresholdSeconds != tc.want {
					t.Fatalf("threshold %d want %d", cfg.Generation.Feedback.ProvisionalThresholdSeconds, tc.want)
				}
				wantMetrics := domain.DefaultGenerationPolicy()
				if tc.value != "" {
					wantMetrics.Skills = false
				}
				if cfg.Generation.Metrics != wantMetrics {
					t.Fatalf("feedback changed optional metric defaults: %+v", cfg.Generation.Metrics)
				}
				rendered, err := config.GenerateConfigYAML(cfg)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), rendered, 0600); err != nil {
					t.Fatal(err)
				}
				loaded, err := config.Load(root)
				if err != nil || loaded.Generation != cfg.Generation {
					t.Fatalf("writer lost threshold/default/metrics: %+v %v", loaded.Generation, err)
				}
			})
		}
	}
	for _, settings := range []string{"feedback: null", "feedback: []", "feedback:\n    typo: 180", "typo: {}"} {
		t.Run("strict_"+settings, func(t *testing.T) {
			root := t.TempDir()
			input := "project_name: Feedback\ngeneration:\n  " + settings + "\n" + testValidSuffixYAML
			if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), []byte(input), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := config.Load(root); err == nil {
				t.Fatalf("invalid configuration accepted: %s", input)
			}
		})
	}
}
