package git

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

// ClassificationConfig contains the explicit repository-relative policies used
// for committed code measurements. Categories are evaluated independently.
type ClassificationConfig struct {
	BaseBranch string                    `json:"base_branch"`
	Production domain.CodeCategoryConfig `json:"production"`
	Tests      domain.CodeCategoryConfig `json:"tests"`
}

// LoadConfig loads the selected project's shared classification policy.
func LoadConfig(repoRoot string) (*ClassificationConfig, error) {
	cfg, err := config.Load(repoRoot)
	if err != nil {
		return nil, err
	}
	classification := &ClassificationConfig{
		BaseBranch: cfg.CodeMetrics.BaseBranch,
		Production: cfg.CodeMetrics.Production,
		Tests:      cfg.CodeMetrics.Tests,
	}
	if err := ValidateClassificationConfig(classification); err != nil {
		return nil, err
	}
	return classification, nil
}

type FileCategory string

const (
	CategoryTest  FileCategory = "test"
	CategoryProd  FileCategory = "prod"
	CategoryOther FileCategory = "other"
)

// ValidateClassificationConfig distinguishes an explicit empty list from an
// absent list and validates every configured glob before measuring any paths.
func ValidateClassificationConfig(cfg *ClassificationConfig) error {
	if cfg == nil {
		return fmt.Errorf("code_metrics classification is required: configure production.include/exclude and tests.include/exclude")
	}
	for _, list := range []struct {
		name     string
		patterns []string
	}{
		{"production.include", cfg.Production.Include},
		{"production.exclude", cfg.Production.Exclude},
		{"tests.include", cfg.Tests.Include},
		{"tests.exclude", cfg.Tests.Exclude},
	} {
		if list.patterns == nil {
			return fmt.Errorf("code_metrics.%s is required: configure a string list, using [] for an intentional empty list", list.name)
		}
		for i, pattern := range list.patterns {
			if strings.TrimSpace(pattern) == "" {
				return fmt.Errorf("code_metrics.%s[%d] must be a nonblank glob", list.name, i)
			}
			if _, err := path.Match(pattern, ""); err != nil {
				return fmt.Errorf("code_metrics.%s[%d] contains invalid glob %q: %w", list.name, i, pattern, err)
			}
		}
	}
	return nil
}

// Patterns containing a slash are repository-relative. A ** path component
// matches zero or more directories; slashless patterns match basenames.
func matchPattern(pattern, filePath string) (bool, error) {
	pattern = strings.TrimPrefix(pattern, "./")
	if !strings.Contains(pattern, "/") {
		return path.Match(pattern, path.Base(filePath))
	}
	return matchSegments(strings.Split(pattern, "/"), strings.Split(filePath, "/"))
}

func matchSegments(pattern, filePath []string) (bool, error) {
	if len(pattern) == 0 {
		return len(filePath) == 0, nil
	}
	if pattern[0] == "**" {
		matched, err := matchSegments(pattern[1:], filePath)
		if err != nil || matched || len(filePath) == 0 {
			return matched, err
		}
		return matchSegments(pattern, filePath[1:])
	}
	if len(filePath) == 0 {
		return false, nil
	}
	matched, err := path.Match(pattern[0], filePath[0])
	if err != nil || !matched {
		return matched, err
	}
	return matchSegments(pattern[1:], filePath[1:])
}

func categoryMatches(filePath string, cfg domain.CodeCategoryConfig) (bool, error) {
	included := false
	for _, pattern := range cfg.Include {
		matched, err := matchPattern(pattern, filePath)
		if err != nil {
			return false, err
		}
		if matched {
			included = true
			break
		}
	}
	if !included {
		return false, nil
	}
	for _, pattern := range cfg.Exclude {
		matched, err := matchPattern(pattern, filePath)
		if err != nil {
			return false, err
		}
		if matched {
			return false, nil
		}
	}
	return true, nil
}

// ClassifyFile applies only the explicitly configured category policies. It
// preserves filename whitespace and does not infer roots, extensions or priority.
func ClassifyFile(rawPath string, cfg *ClassificationConfig) (FileCategory, error) {
	if err := ValidateClassificationConfig(cfg); err != nil {
		return CategoryOther, err
	}
	return classifyFile(rawPath, cfg)
}

func classifyFile(rawPath string, cfg *ClassificationConfig) (FileCategory, error) {
	filePath := strings.TrimPrefix(filepath.ToSlash(rawPath), "./")
	production, err := categoryMatches(filePath, cfg.Production)
	if err != nil {
		return CategoryOther, fmt.Errorf("matching code_metrics.production policy for %q: %w", rawPath, err)
	}
	tests, err := categoryMatches(filePath, cfg.Tests)
	if err != nil {
		return CategoryOther, fmt.Errorf("matching code_metrics.tests policy for %q: %w", rawPath, err)
	}
	switch {
	case production && tests:
		return CategoryOther, fmt.Errorf("code_metrics classification conflict for %q: both production and tests match; adjust their include/exclude lists so the path matches at most one category", rawPath)
	case production:
		return CategoryProd, nil
	case tests:
		return CategoryTest, nil
	default:
		return CategoryOther, nil
	}
}
