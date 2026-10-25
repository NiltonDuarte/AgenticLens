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
	cleanPath = strings.TrimPrefix(cleanPath, "./")
	cleanPath = strings.TrimPrefix(cleanPath, "/")
	baseName := filepath.Base(cleanPath)
	ext := strings.ToLower(filepath.Ext(cleanPath))

	// 1. Workspace directory check - classify immediately as CategoryIgnored
	if cfg != nil && strings.TrimSpace(cfg.WorkspacesDir) != "" {
		wsDir := strings.Trim(strings.TrimSpace(cfg.WorkspacesDir), "/")
		if cleanPath == wsDir || strings.HasPrefix(cleanPath, wsDir+"/") || strings.Contains(cleanPath, "/"+wsDir+"/") {
			return CategoryIgnored
		}
	}

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

	// 2. Specific lockfiles and dotfiles ignored
	if baseName == "go.sum" || baseName == "package-lock.json" || baseName == "pnpm-lock.yaml" || baseName == "yarn.lock" {
		return CategoryIgnored
	}
	if strings.HasPrefix(baseName, ".") {
		return CategoryIgnored
	}

	// 2. Test files detection
	isTestDir := strings.HasPrefix(cleanPath, "testdata/") ||
		strings.Contains(cleanPath, "/testdata/") ||
		strings.HasPrefix(cleanPath, "test/") ||
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
