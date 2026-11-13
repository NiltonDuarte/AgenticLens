package config

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"

	"agenticlens/internal/domain"
	"gopkg.in/yaml.v3"
)

func configurationNodes(data []byte) (map[string]*yaml.Node, error) {
	var doc yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&doc); err != nil {
		return nil, err
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("configuration must contain exactly one YAML document")
	}
	if len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("configuration must be a mapping")
	}
	if err := validateCaptureDocument(&doc); err != nil {
		return nil, err
	}
	nodes := make(map[string]*yaml.Node)
	var visit func(string, *yaml.Node)
	visit = func(prefix string, node *yaml.Node) {
		if prefix != "" {
			nodes[prefix] = node
		}
		if node.Kind == yaml.MappingNode {
			for i := 0; i < len(node.Content); i += 2 {
				key := configurationKeySegment(node.Content[i].Value)
				if prefix != "" {
					key = prefix + "." + key
				}
				visit(key, node.Content[i+1])
			}
		}
	}
	visit("", doc.Content[0])
	return nodes, nil
}

// Keep literal mapping-key punctuation distinct from structural separators.
// A root key "ticket.case_insensitive" cannot stand in for nested YAML policy.
func configurationKeySegment(key string) string {
	return strings.NewReplacer(`\`, `\\`, `.`, `\.`).Replace(key)
}

func requireNodeType(n *yaml.Node, key string, kind yaml.Kind, tag string) error {
	if n == nil {
		return fmt.Errorf("%s must be explicitly configured", key)
	}
	if n.Kind != kind || (tag != "" && n.Tag != tag) {
		return fmt.Errorf("%s has an invalid type; expected %s", key, tag)
	}
	return nil
}

func validateSharedNodes(nodes map[string]*yaml.Node) error {
	if nodes["agenticlens.state_dir"] != nil {
		return fmt.Errorf("agenticlens.state_dir belongs only in local YAML")
	}
	if err := validateWorkspaceNodes(nodes["workspaces"]); err != nil {
		return err
	}
	for _, key := range []string{"ticket", "workspaces", "database", "live", "anomaly_thresholds", "code_metrics", "code_metrics.production", "code_metrics.tests", "generation", "generation.metrics", "generation.feedback", "generation.summary", "role_mappings"} {
		if n := nodes[key]; n != nil {
			if err := requireNodeType(n, key, yaml.MappingNode, ""); err != nil {
				return err
			}
		}
	}
	for _, key := range []string{"project_name", "code_metrics.base_branch", "generation.summary.filename"} {
		if n := nodes[key]; n != nil {
			if err := requireNodeType(n, key, yaml.ScalarNode, "!!str"); err != nil {
				return err
			}
		}
	}
	for _, key := range []string{"ticket.case_insensitive", "generation.metrics.skills", "generation.metrics.delegations", "generation.metrics.anomalies", "generation.summary.enabled"} {
		if n := nodes[key]; n != nil {
			if err := requireNodeType(n, key, yaml.ScalarNode, "!!bool"); err != nil {
				return err
			}
		}
	}
	for _, key := range []string{"ticket.prefixes", "workspaces.artifacts", "role_order", "code_metrics.production.include", "code_metrics.production.exclude", "code_metrics.tests.include", "code_metrics.tests.exclude"} {
		n := nodes[key]
		if n == nil {
			continue
		}
		if err := requireNodeType(n, key, yaml.SequenceNode, ""); err != nil {
			return err
		}
		for i, item := range n.Content {
			if key == "ticket.prefixes" || key == "workspaces.artifacts" {
				if err := requireNodeType(item, fmt.Sprintf("%s[%d]", key, i), yaml.MappingNode, ""); err != nil {
					return err
				}
				for j := 0; j < len(item.Content); j += 2 {
					field := item.Content[j].Value
					if key == "ticket.prefixes" && field != "prefix" && field != "description" && field != "ticket_type" {
						return fmt.Errorf("ticket.prefixes[%d] contains unsupported field %q", i, field)
					}
					if key == "workspaces.artifacts" && field != "path" && field != "title" && field != "description" {
						return fmt.Errorf("workspaces.artifacts[%d] contains unsupported field %q", i, field)
					}
					if err := requireNodeType(item.Content[j+1], fmt.Sprintf("%s[%d].%s", key, i, item.Content[j].Value), yaml.ScalarNode, "!!str"); err != nil {
						return err
					}
				}
			} else if err := requireNodeType(item, fmt.Sprintf("%s[%d]", key, i), yaml.ScalarNode, "!!str"); err != nil {
				return err
			}
		}
	}
	if n := nodes["role_mappings"]; n != nil {
		for i := 0; i < len(n.Content); i += 2 {
			if err := requireNodeType(n.Content[i+1], "role_mappings."+n.Content[i].Value, yaml.ScalarNode, "!!str"); err != nil {
				return err
			}
		}
	}
	for _, key := range []string{"database.busy_timeout_ms", "live.active_threshold_seconds", "live.poll_interval_ms", "generation.feedback.provisional_threshold_seconds", "anomaly_thresholds.stall_threshold_seconds", "anomaly_thresholds.error_rate_min_calls", "anomaly_thresholds.consecutive_errors_streak_warning", "anomaly_thresholds.consecutive_errors_streak_critical", "anomaly_thresholds.context_ceiling_warning_tokens", "anomaly_thresholds.context_ceiling_critical_tokens", "anomaly_thresholds.turn_usage_warning_tokens", "anomaly_thresholds.turn_usage_critical_tokens", "anomaly_thresholds.context_growth_warning_tokens", "anomaly_thresholds.context_growth_critical_tokens", "anomaly_thresholds.tool_thrashing_min_failures", "anomaly_thresholds.tool_thrashing_critical_failures"} {
		if n := nodes[key]; n != nil {
			if err := requireNodeType(n, key, yaml.ScalarNode, "!!int"); err != nil {
				return err
			}
			var value int64
			if err := n.Decode(&value); err != nil {
				return fmt.Errorf("%s must be a representable integer: %w", key, err)
			}
			if value < 0 || (value == 0 && key != "database.busy_timeout_ms" && key != "generation.feedback.provisional_threshold_seconds") {
				return fmt.Errorf("%s must be %s", key, map[bool]string{true: "a nonnegative integer", false: "a positive integer"}[key == "database.busy_timeout_ms" || key == "generation.feedback.provisional_threshold_seconds"])
			}
		}
	}
	for _, key := range []string{"anomaly_thresholds.error_rate_threshold", "anomaly_thresholds.error_rate_critical_threshold"} {
		if n := nodes[key]; n != nil {
			if n.Kind != yaml.ScalarNode || (n.Tag != "!!float" && n.Tag != "!!int") {
				return fmt.Errorf("%s must be a finite ratio greater than 0 and at most 1", key)
			}
			var value float64
			if err := n.Decode(&value); err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 || value > 1 {
				return fmt.Errorf("%s must be a finite ratio greater than 0 and at most 1", key)
			}
		}
	}
	return nil
}

func ValidateGenerationPolicy(cfg domain.ProjectConfig) error {
	return cfg.RequireSettings("generation.metrics.skills", "generation.metrics.delegations", "generation.metrics.anomalies")
}
func ValidateFeedbackPolicy(cfg domain.ProjectConfig) error {
	if err := cfg.RequireSettings("generation.feedback.provisional_threshold_seconds"); err != nil {
		return err
	}
	if cfg.Generation.Feedback.ProvisionalThresholdSeconds < 0 {
		return fmt.Errorf("generation.feedback.provisional_threshold_seconds must be a nonnegative integer")
	}
	return nil
}
func ValidateSummaryPolicy(cfg domain.ProjectConfig) error {
	if err := cfg.RequireSettings("generation.summary.enabled"); err != nil {
		return err
	}
	if !cfg.Generation.Summary.Enabled {
		return nil
	}
	if err := cfg.RequireSettings("generation.summary.filename"); err != nil {
		return err
	}
	return ValidateSummaryFilename(cfg.Generation.Summary.Filename)
}

func ValidateSummaryFilename(name string) error {
	if name == "" || strings.TrimSpace(name) != name || name == "." || name == ".." || filepath.Base(name) != name || strings.ContainsAny(name, "/\\\x00\r\n") || strings.EqualFold(name, "agentic_metrics.json") {
		return fmt.Errorf("generation.summary.filename must be a safe nonblank basename distinct from the ledger filename")
	}
	return nil
}
func ValidateDatabasePolicy(cfg domain.ProjectConfig) error {
	if err := cfg.RequireSettings("database.busy_timeout_ms"); err != nil {
		return err
	}
	if cfg.Database.BusyTimeoutMs < 0 {
		return fmt.Errorf("database.busy_timeout_ms must be a nonnegative integer")
	}
	return nil
}
func ValidateActivityPolicy(cfg domain.ProjectConfig) error {
	if err := cfg.RequireSettings("live.active_threshold_seconds", "anomaly_thresholds.stall_threshold_seconds"); err != nil {
		return err
	}
	if cfg.Live.ActiveThresholdSeconds <= 0 || cfg.AnomalyThresholds.StallThresholdSeconds <= 0 {
		return fmt.Errorf("live.active_threshold_seconds and anomaly_thresholds.stall_threshold_seconds must be positive integers")
	}
	return nil
}
func ValidatePollingPolicy(cfg domain.ProjectConfig) error {
	if err := cfg.RequireSettings("live.poll_interval_ms"); err != nil {
		return err
	}
	if cfg.Live.PollIntervalMs <= 0 {
		return fmt.Errorf("live.poll_interval_ms must be a positive integer")
	}
	return nil
}
func ValidateAnomalyPolicy(cfg domain.ProjectConfig) error {
	keys := []string{"error_rate_min_calls", "error_rate_threshold", "error_rate_critical_threshold", "context_ceiling_warning_tokens", "context_ceiling_critical_tokens", "consecutive_errors_streak_warning", "consecutive_errors_streak_critical", "tool_thrashing_min_failures", "tool_thrashing_critical_failures", "turn_usage_warning_tokens", "turn_usage_critical_tokens", "context_growth_warning_tokens", "context_growth_critical_tokens"}
	for _, key := range keys {
		if err := cfg.RequireSettings("anomaly_thresholds." + key); err != nil {
			return err
		}
	}
	if err := validateErrorRatePolicy(cfg.AnomalyThresholds); err != nil {
		return err
	}
	t := cfg.AnomalyThresholds
	for _, p := range []struct {
		key               string
		warning, critical int64
	}{{"context_ceiling", t.ContextCeilingWarningTokens, t.ContextCeilingCriticalTokens}, {"consecutive_errors_streak", int64(t.ConsecutiveErrorsStreakWarning), int64(t.ConsecutiveErrorsStreakCritical)}, {"tool_thrashing", int64(t.ToolThrashingMinFailures), int64(t.ToolThrashingCriticalFailures)}, {"turn_usage", t.TurnUsageWarningTokens, t.TurnUsageCriticalTokens}, {"context_growth", t.ContextGrowthWarningTokens, t.ContextGrowthCriticalTokens}} {
		if p.warning <= 0 || p.critical < p.warning {
			return fmt.Errorf("anomaly_thresholds.%s requires explicit positive warning/critical values, warning <= critical", p.key)
		}
	}
	return nil
}

func readConfigurationComments(data []byte) map[string]domain.ConfigComment {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil || len(doc.Content) == 0 {
		return nil
	}
	comments := map[string]domain.ConfigComment{}
	comments["$document"] = domain.ConfigComment{Head: doc.HeadComment, Line: doc.LineComment, Foot: doc.FootComment}
	walkConfigurationComments(doc.Content[0], "", func(path string, key, value *yaml.Node) {
		c := domain.ConfigComment{Head: value.HeadComment, Line: value.LineComment, Foot: value.FootComment}
		if key != nil {
			c.KeyHead = key.HeadComment
			c.KeyLine = key.LineComment
			c.KeyFoot = key.FootComment
		}
		comments[path] = c
	})
	return comments
}
func applyConfigurationComments(root *yaml.Node, comments map[string]domain.ConfigComment) {
	if document, ok := comments["$document"]; ok {
		if document.Head != "" {
			root.HeadComment = document.Head
		}
		if document.Foot != "" {
			root.FootComment = document.Foot
		}
		if document.Line != "" {
			root.LineComment = document.Line
		}
	}
	walkConfigurationComments(root, "", func(path string, key, value *yaml.Node) {
		c, ok := comments[path]
		if !ok {
			return
		}
		if c.Head != "" {
			value.HeadComment = c.Head
		}
		if c.Line != "" {
			value.LineComment = c.Line
		}
		if c.Foot != "" {
			value.FootComment = c.Foot
		}
		if key != nil {
			if c.KeyHead != "" {
				key.HeadComment = c.KeyHead
			}
			if c.KeyLine != "" {
				key.LineComment = c.KeyLine
			}
			if c.KeyFoot != "" {
				key.FootComment = c.KeyFoot
			}
		}
	})
}
func walkConfigurationComments(root *yaml.Node, path string, visit func(string, *yaml.Node, *yaml.Node)) {
	if path == "" {
		visit(path, nil, root)
	}
	switch root.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(root.Content); i += 2 {
			key, value := root.Content[i], root.Content[i+1]
			child := configurationKeySegment(key.Value)
			if path != "" {
				child = path + "." + child
			}
			visit(child, key, value)
			walkConfigurationComments(value, child, visit)
		}
	case yaml.SequenceNode:
		for i, item := range root.Content {
			child := fmt.Sprintf("%s[%d]", path, i)
			visit(child, nil, item)
			walkConfigurationComments(item, child, visit)
		}
	}
}

// ValidateContextPolicy guards fresh derived saturation/health calculations.
// Shared-only reads may instead project unavailable policy without a denominator.
func ValidateContextPolicy(cfg domain.ProjectConfig) error {
	if err := cfg.RequireSettings("anomaly_thresholds.context_ceiling_warning_tokens", "anomaly_thresholds.context_ceiling_critical_tokens"); err != nil {
		return err
	}
	warn, critical := cfg.AnomalyThresholds.ContextCeilingWarningTokens, cfg.AnomalyThresholds.ContextCeilingCriticalTokens
	if warn <= 0 || critical < warn {
		return fmt.Errorf("anomaly_thresholds context ceilings require explicit positive token counts, warning <= critical")
	}
	return nil
}
