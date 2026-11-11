package config_test

import (
	"agenticlens/internal/adapters/config"
	"strings"
	"testing"
)

func TestGenerationFeedbackRequiresExplicitNonnegativeValue(t *testing.T) {
	for _, value := range []string{"0", "1", "2147483647", "<missing>", "null", "-1", "1.5", "1.0", "'0'", "[]", "{}"} {
		t.Run(value, func(t *testing.T) {
			cfg, err := loadBody(t, replaceSetting(t, fixtureSharedBody(t), "generation.feedback.provisional_threshold_seconds", value))
			if err == nil {
				err = config.ValidateFeedbackPolicy(cfg)
			}
			valid := value == "0" || value == "1" || value == "2147483647"
			if valid != (err == nil) {
				t.Fatalf("value %s error %v", value, err)
			}
			if !valid && err != nil && !strings.Contains(err.Error(), "generation.feedback") {
				t.Fatalf("missing setting guidance: %v", err)
			}
		})
	}
}

func TestSharedSummaryRequiresOnlyEnabledDestination(t *testing.T) {
	base := fixtureSharedBody(t)
	for _, value := range []string{"<missing>", "null", "'false'", "0"} {
		cfg, err := loadBody(t, replaceSetting(t, base, "generation.summary.enabled", value))
		if err == nil {
			err = config.ValidateSummaryPolicy(cfg)
		}
		if err == nil {
			t.Fatalf("invalid enabled value %s accepted", value)
		}
	}
	for _, value := range []string{"<missing>", "''", "'  '", "../METRICS.md", "nested/METRICS.md", "agentic_metrics.json", "'.'", "'..'", "'bad\\\\name.md'"} {
		t.Run(value, func(t *testing.T) {
			cfg, err := loadBody(t, replaceSetting(t, base, "generation.summary.filename", value))
			if err == nil {
				err = config.ValidateSummaryPolicy(cfg)
			}
			if err == nil {
				t.Fatalf("unsafe configured destination %s accepted", value)
			}
		})
	}
	body := replaceSetting(t, base, "generation.summary.enabled", "false")
	body = replaceSetting(t, body, "generation.summary.filename", "<missing>")
	cfg, err := loadBody(t, body)
	if err != nil || config.ValidateSummaryPolicy(cfg) != nil {
		t.Fatalf("disabled summary required filename: %v", err)
	}
}
