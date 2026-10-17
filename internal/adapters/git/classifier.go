package git

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ClassificationConfig holds user-configured file glob patterns.
type ClassificationConfig struct {
	Include      []string `json:"include"`
	Exclude      []string `json:"exclude"`
	TestPatterns []string `json:"test_patterns"`
	ProdPatterns []string `json:"prod_patterns"`
}

type rootConfig struct {
	WorkspacesDir string                `json:"workspaces_dir"`
	CodeMetrics   *ClassificationConfig `json:"code_metrics"`
}

// LoadConfig attempts to load ClassificationConfig from .agenticlens/config.json.
func LoadConfig(repoRoot string) (*ClassificationConfig, error) {
	searchDir := repoRoot
	if searchDir == "" {
		cwd, err := os.Getwd()
		if err == nil {
			searchDir = cwd
		}
	}

	curr := searchDir
	for curr != "" {
		cfgPath := filepath.Join(curr, ".agenticlens", "config.json")
		if fi, err := os.Stat(cfgPath); err == nil && !fi.IsDir() {
			data, err := os.ReadFile(cfgPath)
			if err != nil {
				return nil, err
			}
			var cfg rootConfig
			if err := json.Unmarshal(data, &cfg); err != nil {
				return nil, err
			}
			return cfg.CodeMetrics, nil
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	return nil, nil
}

// FileCategory represents the category of a modified file.
type FileCategory string

const (
	CategoryIgnored FileCategory = "ignored"
	CategoryTest    FileCategory = "test"
	CategoryProd    FileCategory = "prod"
	CategoryOther   FileCategory = "other"
)

func matchPattern(pattern, path string) bool {
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)

	if pattern == path {
		return true
	}

	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")
			hasPrefix := prefix == "" || strings.HasPrefix(path, prefix)
			hasSuffix := suffix == "" || strings.HasSuffix(path, suffix) || matchPattern(suffix, filepath.Base(path))
			if hasPrefix && hasSuffix {
				return true
			}
		}
	}

	matched, err := filepath.Match(pattern, path)
	if err == nil && matched {
		return true
	}

	baseMatched, err := filepath.Match(pattern, filepath.Base(path))
	if err == nil && baseMatched {
		return true
	}

	return false
}

// ClassifyFile classifies a repository file path as ignored, test, prod, or other.
func ClassifyFile(rawPath string, cfg *ClassificationConfig) FileCategory {
	cleanPath := filepath.ToSlash(strings.TrimSpace(rawPath))
	baseName := filepath.Base(cleanPath)
	ext := strings.ToLower(filepath.Ext(cleanPath))

	if cfg != nil {
		for _, p := range cfg.Exclude {
			if matchPattern(p, cleanPath) {
				return CategoryIgnored
			}
		}
		for _, p := range cfg.TestPatterns {
			if matchPattern(p, cleanPath) {
				return CategoryTest
			}
		}
		for _, p := range cfg.ProdPatterns {
			if matchPattern(p, cleanPath) {
				return CategoryProd
			}
		}
		for _, p := range cfg.Include {
			if matchPattern(p, cleanPath) {
				return CategoryProd
			}
		}
	}

	// 1. Universal ignore rules
	if strings.HasPrefix(cleanPath, "workspaces/") || strings.Contains(cleanPath, "/workspaces/") {
		return CategoryIgnored
	}
	if ext == ".md" || ext == ".json" || ext == ".lock" {
		return CategoryIgnored
	}
	if baseName == "go.sum" || baseName == "package-lock.json" || baseName == "pnpm-lock.yaml" || baseName == "yarn.lock" {
		return CategoryIgnored
	}
	if strings.HasPrefix(baseName, ".") {
		return CategoryIgnored
	}

	// 2. Test files detection
	isTestDir := strings.HasPrefix(cleanPath, "test/") ||
		strings.HasPrefix(cleanPath, "tests/") ||
		strings.HasPrefix(cleanPath, "__tests__/") ||
		strings.Contains(cleanPath, "/test/") ||
		strings.Contains(cleanPath, "/tests/") ||
		strings.Contains(cleanPath, "/__tests__/")

	isTestFile := strings.HasSuffix(baseName, "_test.go") ||
		strings.HasSuffix(baseName, ".test.ts") ||
		strings.HasSuffix(baseName, ".test.tsx") ||
		strings.HasSuffix(baseName, ".spec.ts") ||
		strings.HasSuffix(baseName, ".spec.tsx") ||
		strings.HasSuffix(baseName, ".test.js") ||
		strings.HasSuffix(baseName, ".test.jsx") ||
		strings.HasSuffix(baseName, ".spec.js") ||
		strings.HasSuffix(baseName, ".spec.jsx") ||
		strings.HasPrefix(baseName, "test_") && ext == ".py" ||
		strings.HasSuffix(baseName, "_test.py")

	if isTestDir || isTestFile {
		return CategoryTest
	}

	// 3. Production files detection
	if baseName == "Makefile" {
		return CategoryProd
	}
	prodExtensions := map[string]struct{}{
		".go":   {},
		".ts":   {},
		".tsx":  {},
		".py":   {},
		".js":   {},
		".jsx":  {},
		".html": {},
		".css":  {},
		".scss": {},
		".sql":  {},
		".sh":   {},
	}
	if _, ok := prodExtensions[ext]; ok {
		return CategoryProd
	}

	return CategoryOther
}
