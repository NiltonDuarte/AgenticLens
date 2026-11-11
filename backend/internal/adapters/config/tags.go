package config

import (
	"fmt"
	"strings"

	"agenticlens/internal/domain"
	"gopkg.in/yaml.v3"
)

func decodeTags(node *yaml.Node) (map[string]string, error) {
	if node == nil {
		return nil, nil
	}
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("tags must be a string mapping")
	}
	result := make(map[string]string, len(node.Content)/2)
	for i := 0; i < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]
		if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
			return nil, fmt.Errorf("tags require string keys and values")
		}
		k, v := strings.TrimSpace(key.Value), strings.TrimSpace(value.Value)
		if _, exists := result[k]; exists {
			return nil, fmt.Errorf("duplicate normalized tag key %q", k)
		}
		result[k] = v
	}
	return domain.NormalizeTags(result)
}
