package config

import (
	"fmt"

	"agenticlens/internal/domain"
	"gopkg.in/yaml.v3"
)

// decodeGenerationConfig checks the YAML node types (also for JSON input) so
// quoted booleans/integers and explicit null never silently become defaults.
func decodeGenerationConfig(data []byte) (domain.GenerationConfig, error) {
	config := domain.GenerationConfig{Metrics: domain.DefaultGenerationPolicy(), Feedback: domain.DefaultGenerationFeedbackConfig()}
	policy := &config.Metrics
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return config, fmt.Errorf("decoding generation configuration: %w", err)
	}
	if len(document.Content) == 0 {
		return config, nil
	}
	root := document.Content[0]
	if root.Kind != yaml.MappingNode {
		return config, fmt.Errorf("configuration must be a mapping")
	}
	for i := 0; i < len(root.Content); i += 2 {
		if root.Content[i].Value != "generation" {
			continue
		}
		generation := root.Content[i+1]
		if generation.Kind != yaml.MappingNode {
			return config, fmt.Errorf("generation must be a mapping")
		}
		for j := 0; j < len(generation.Content); j += 2 {
			if generation.Content[j].Value == "feedback" {
				feedback := generation.Content[j+1]
				if feedback.Kind != yaml.MappingNode {
					return config, fmt.Errorf("generation.feedback must be a mapping")
				}
				for k := 0; k < len(feedback.Content); k += 2 {
					name, value := feedback.Content[k].Value, feedback.Content[k+1]
					if name != "provisional_threshold_seconds" {
						return config, fmt.Errorf("unknown generation.feedback.%s; expected provisional_threshold_seconds", name)
					}
					if value.Kind != yaml.ScalarNode || value.Tag != "!!int" {
						return config, fmt.Errorf("generation.feedback.provisional_threshold_seconds must be a nonnegative integer")
					}
					var seconds int64
					if err := value.Decode(&seconds); err != nil {
						return config, fmt.Errorf("generation.feedback.provisional_threshold_seconds must be a representable nonnegative integer: %w", err)
					}
					if seconds < 0 {
						return config, fmt.Errorf("generation.feedback.provisional_threshold_seconds must be a nonnegative integer")
					}
					config.Feedback.ProvisionalThresholdSeconds = seconds
				}
				continue
			}
			if generation.Content[j].Value != "metrics" {
				return config, fmt.Errorf("unknown generation setting %q; expected metrics or feedback", generation.Content[j].Value)
			}
			metrics := generation.Content[j+1]
			if metrics.Kind != yaml.MappingNode {
				return config, fmt.Errorf("generation.metrics must be a mapping")
			}
			for k := 0; k < len(metrics.Content); k += 2 {
				name, value := metrics.Content[k].Value, metrics.Content[k+1]
				if name != "skills" && name != "delegations" && name != "anomalies" {
					return config, fmt.Errorf("unknown generation.metrics.%s; expected skills, delegations or anomalies", name)
				}
				if value.Kind != yaml.ScalarNode || value.Tag != "!!bool" {
					return config, fmt.Errorf("generation.metrics.%s must be a boolean (true or false)", name)
				}
				var enabled bool
				if err := value.Decode(&enabled); err != nil {
					return config, fmt.Errorf("generation.metrics.%s must be a boolean: %w", name, err)
				}
				switch name {
				case "skills":
					policy.Skills = enabled
				case "delegations":
					policy.Delegations = enabled
				case "anomalies":
					policy.Anomalies = enabled
				}
			}
		}
	}
	return config, nil
}
