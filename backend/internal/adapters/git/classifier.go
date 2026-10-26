package git

import (
	"path/filepath"
	"strings"

	"agenticlens/internal/adapters/config"
)

// ClassificationConfig holds user-configured file glob patterns and workspace layout options.
type ClassificationConfig struct {
	Include       []string `json:"include"`
	Exclude       []string `json:"exclude"`
	TestPatterns  []string `json:"test_patterns"`
	ProdPatterns  []string `json:"prod_patterns"`
	BaseBranch    string   `json:"base_branch,omitempty"`
	SrcFolders    []string `json:"src_folders,omitempty"`
	WorkspacesDir string   `json:"workspaces_dir,omitempty"`
}

// LoadConfig attempts to load ClassificationConfig from repository configuration.
func LoadConfig(repoRoot string) (*ClassificationConfig, error) {
	cfg, err := config.Load(repoRoot)
	if err != nil {
		return nil, err
	}

	return &ClassificationConfig{
		Include:       cfg.CodeMetrics.Include,
		Exclude:       cfg.CodeMetrics.Exclude,
		TestPatterns:  cfg.CodeMetrics.TestPatterns,
		ProdPatterns:  cfg.CodeMetrics.ProdPatterns,
		BaseBranch:    cfg.CodeMetrics.BaseBranch,
		SrcFolders:    cfg.Workspaces.SrcFolders,
		WorkspacesDir: cfg.Workspaces.Dir,
	}, nil
}

// FileCategory represents the category of a modified file.
type FileCategory string

const (
	CategoryIgnored FileCategory = "ignored"
	CategoryTest    FileCategory = "test"
	CategoryProd    FileCategory = "prod"
	CategoryOther   FileCategory = "other"
)

// DefaultTestPatterns defines standard fallback glob patterns for identifying test code.
var DefaultTestPatterns = config.DefaultTestPatterns

// DefaultProdPatterns defines standard fallback glob patterns for identifying production runtime code.
var DefaultProdPatterns = config.DefaultProdPatterns

func hasPatterns(patterns []string) bool {
	for _, p := range patterns {
		if strings.TrimSpace(p) != "" {
			return true
		}
	}
	return false
}

// EffectiveTestPatterns returns configured test patterns if present, or standard default fallbacks.
func EffectiveTestPatterns(cfg *ClassificationConfig) []string {
	if cfg != nil && hasPatterns(cfg.TestPatterns) {
		return cfg.TestPatterns
	}
	return DefaultTestPatterns
}

// EffectiveProdPatterns returns configured production patterns if present, or standard default fallbacks.
func EffectiveProdPatterns(cfg *ClassificationConfig) []string {
	if cfg != nil && hasPatterns(cfg.ProdPatterns) {
		return cfg.ProdPatterns
	}
	return DefaultProdPatterns
}

// Patterns containing a slash are relative to the repository root. A ** path
// component matches zero or more directories; patterns without slashes match basenames.
func matchPattern(pattern, path string) bool {
	pattern = strings.TrimPrefix(filepath.ToSlash(strings.TrimSpace(pattern)), "./")
	path = filepath.ToSlash(path)
	if !strings.Contains(pattern, "/") {
		matched, _ := filepath.Match(pattern, filepath.Base(path))
		return matched
	}
	return matchSegments(strings.Split(pattern, "/"), strings.Split(path, "/"))
}

func matchSegments(pattern, path []string) bool {
	if len(pattern) == 0 {
		return len(path) == 0
	}
	if pattern[0] == "**" {
		return matchSegments(pattern[1:], path) || (len(path) > 0 && matchSegments(pattern, path[1:]))
	}
	if len(path) == 0 {
		return false
	}
	matched, _ := filepath.Match(pattern[0], path[0])
	return matched && matchSegments(pattern[1:], path[1:])
}

// ClassifyFile classifies a repository file path as ignored, test, prod, or other.
func ClassifyFile(rawPath string, cfg *ClassificationConfig) FileCategory {
	cleanPath := filepath.ToSlash(strings.TrimSpace(rawPath))
	cleanPath = strings.TrimPrefix(cleanPath, "./")
	cleanPath = strings.TrimPrefix(cleanPath, "/")
	baseName := filepath.Base(cleanPath)

	// 1. Workspace directory check - classify immediately as CategoryIgnored
	if cfg != nil && strings.TrimSpace(cfg.WorkspacesDir) != "" {
		wsDir := strings.Trim(strings.TrimSpace(cfg.WorkspacesDir), "/")
		if cleanPath == wsDir || strings.HasPrefix(cleanPath, wsDir+"/") || strings.Contains(cleanPath, "/"+wsDir+"/") {
			return CategoryIgnored
		}
	}

	// 2. Explicit exclusions
	if cfg != nil {
		for _, p := range cfg.Exclude {
			if matchPattern(p, cleanPath) {
				return CategoryIgnored
			}
		}
	}

	// 3. Explicit test patterns (take precedence over production rules)
	if cfg != nil && hasPatterns(cfg.TestPatterns) {
		for _, p := range cfg.TestPatterns {
			if matchPattern(p, cleanPath) {
				return CategoryTest
			}
		}
	}

	// 4. Explicit production patterns
	if cfg != nil && hasPatterns(cfg.ProdPatterns) {
		for _, p := range cfg.ProdPatterns {
			if matchPattern(p, cleanPath) {
				return CategoryProd
			}
		}
	}

	// 5. Explicit includes
	if cfg != nil {
		for _, p := range cfg.Include {
			if matchPattern(p, cleanPath) {
				return CategoryProd
			}
		}
	}

	// 6. Specific lockfiles and dotfiles ignored
	if baseName == "go.sum" || baseName == "package-lock.json" || baseName == "pnpm-lock.yaml" || baseName == "yarn.lock" {
		return CategoryIgnored
	}
	if strings.HasPrefix(baseName, ".") {
		return CategoryIgnored
	}

	// 7. Fallback test patterns when test patterns are empty or unconfigured
	if cfg == nil || !hasPatterns(cfg.TestPatterns) {
		for _, p := range DefaultTestPatterns {
			if matchPattern(p, cleanPath) {
				return CategoryTest
			}
		}
	}

	// 8. Fallback production patterns when production patterns are empty or unconfigured
	if cfg == nil || !hasPatterns(cfg.ProdPatterns) {
		for _, p := range DefaultProdPatterns {
			if matchPattern(p, cleanPath) {
				return CategoryProd
			}
		}
	}

	return CategoryOther
}
