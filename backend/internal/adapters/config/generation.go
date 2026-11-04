package config

import (
	"agenticlens/internal/domain"
	"fmt"
	"gopkg.in/yaml.v3"
)

// decodeGenerationPolicy checks the YAML node types (also for JSON input) so
// quoted booleans and explicit null never silently become defaults.
func decodeGenerationPolicy(data []byte) (domain.GenerationPolicy, error) {
	policy := domain.DefaultGenerationPolicy()
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return policy, fmt.Errorf("decoding generation configuration: %w", err)
	}
	if len(document.Content) == 0 {
		return policy, nil
	}
	root := document.Content[0]
	if root.Kind != yaml.MappingNode {
		return policy, fmt.Errorf("configuration must be a mapping")
	}
	for i := 0; i < len(root.Content); i += 2 {
		if root.Content[i].Value != "generation" {
			continue
		}
		generation := root.Content[i+1]
		if generation.Kind != yaml.MappingNode {
			return policy, fmt.Errorf("generation must be a mapping")
		}
		for j := 0; j < len(generation.Content); j += 2 {
			if generation.Content[j].Value != "metrics" {
				return policy, fmt.Errorf("unknown generation setting %q; expected metrics", generation.Content[j].Value)
			}
			metrics := generation.Content[j+1]
			if metrics.Kind != yaml.MappingNode {
				return policy, fmt.Errorf("generation.metrics must be a mapping")
			}
			for k := 0; k < len(metrics.Content); k += 2 {
				name, value := metrics.Content[k].Value, metrics.Content[k+1]
				if name != "skills" && name != "delegations" && name != "anomalies" {
					return policy, fmt.Errorf("unknown generation.metrics.%s; expected skills, delegations or anomalies", name)
				}
				if value.Kind != yaml.ScalarNode || value.Tag != "!!bool" {
					return policy, fmt.Errorf("generation.metrics.%s must be a boolean (true or false)", name)
				}
				var enabled bool
				if err := value.Decode(&enabled); err != nil {
					return policy, fmt.Errorf("generation.metrics.%s must be a boolean: %w", name, err)
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
	return policy, nil
}
