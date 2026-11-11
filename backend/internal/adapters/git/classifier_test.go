package git_test

import (
	"strings"
	"testing"

	"agenticlens/internal/adapters/git"
	"agenticlens/internal/domain"
)

func classification(prodInclude, prodExclude, testInclude, testExclude []string) *git.ClassificationConfig {
	return &git.ClassificationConfig{
		BaseBranch: "baseline",
		Production: domain.CodeCategoryConfig{Include: append([]string{}, prodInclude...), Exclude: append([]string{}, prodExclude...)},
		Tests:      domain.CodeCategoryConfig{Include: append([]string{}, testInclude...), Exclude: append([]string{}, testExclude...)},
	}
}

func TestClassifyFileIndependentCategories(t *testing.T) {
	t.Parallel()
	cfg := classification(
		[]string{"backend/**/*.go", "runtime/**/*.custom", "build.recipe"},
		[]string{"**/*_test.go", "**/testdata/**"},
		[]string{"**/*_test.go", "**/testdata/**", "checks/**/samples/**"},
		[]string{"vendor/**"},
	)
	for _, tc := range []struct {
		path string
		want git.FileCategory
	}{
		{"backend/main.go", git.CategoryProd},
		{"backend/internal/nested/main.go", git.CategoryProd},
		{"backend/main_test.go", git.CategoryTest},
		{"backend/testdata/input.go", git.CategoryTest},
		{"./backend/testdata/nested/input.jsonl", git.CategoryTest},
		{"testdata/input.json", git.CategoryTest},
		{"checks/samples/data.json", git.CategoryTest},
		{"checks/unit/samples/nested/data.json", git.CategoryTest},
		{"checks/unit/samples_backup/data.json", git.CategoryOther},
		{"other/checks/samples/data.json", git.CategoryOther},
		{"runtime/main.custom", git.CategoryProd},
		{"runtime/nested/main.custom", git.CategoryProd},
		{"runtime_backup/main.custom", git.CategoryOther},
		{"other/runtime/main.custom", git.CategoryOther},
		{"build.recipe", git.CategoryProd},
		{"vendor/pkg/main_test.go", git.CategoryOther},
		{"application/unclassified.go", git.CategoryOther},
		{"backend/testdata_backup/input.json", git.CategoryOther},
		{"docs/README.md", git.CategoryOther},
	} {
		t.Run(tc.path, func(t *testing.T) {
			got, err := git.ClassifyFile(tc.path, cfg)
			if err != nil || got != tc.want {
				t.Fatalf("ClassifyFile(%q) = %q, %v; want %q", tc.path, got, err, tc.want)
			}
		})
	}
}

func TestClassifyFileExplicitPolicyHasNoImplicitIgnores(t *testing.T) {
	t.Parallel()
	files := []string{"tickets/CASE-1/helper.go", "workspaces/file.go", "nested/custom_ws/data.csv", "go.sum", "package-lock.json", "pnpm-lock.yaml", "yarn.lock", ".gitignore", ".env", "Makefile", "README.md", "scripts/tool.py"}
	for _, file := range files {
		for _, tc := range []struct {
			cfg  *git.ClassificationConfig
			want git.FileCategory
		}{
			{classification(nil, nil, nil, nil), git.CategoryOther},
			{classification([]string{"**"}, nil, nil, nil), git.CategoryProd},
			{classification(nil, nil, []string{"**"}, nil), git.CategoryTest},
		} {
			got, err := git.ClassifyFile(file, tc.cfg)
			if err != nil || got != tc.want {
				t.Errorf("%q: got %q, %v; want %q", file, got, err, tc.want)
			}
		}
	}
}

func TestClassifyFileRejectsOverlapWithoutCategoryPriority(t *testing.T) {
	t.Parallel()
	cfg := classification([]string{"**/*.go"}, nil, []string{"**/*_test.go"}, nil)
	if _, err := git.ClassifyFile("backend/main_test.go", cfg); err == nil || !strings.Contains(err.Error(), "backend/main_test.go") || !strings.Contains(err.Error(), "both production and tests") {
		t.Fatalf("overlap must identify the path and both categories: %v", err)
	}
	cfg.Production.Exclude = []string{"**/*_test.go"}
	if got, err := git.ClassifyFile("backend/main_test.go", cfg); err != nil || got != git.CategoryTest {
		t.Fatalf("explicit production exclusion: got %s, %v", got, err)
	}
	cfg.Production.Exclude = []string{}
	cfg.Tests.Exclude = []string{"**/*_test.go"}
	if got, err := git.ClassifyFile("backend/main_test.go", cfg); err != nil || got != git.CategoryProd {
		t.Fatalf("explicit test exclusion: got %s, %v", got, err)
	}
}

func TestClassificationRequiresEveryValidList(t *testing.T) {
	t.Parallel()
	for _, field := range []string{"production.include", "production.exclude", "tests.include", "tests.exclude"} {
		for _, patterns := range [][]string{nil, {""}, {" \t "}, {"[broken"}} {
			cfg := classification(nil, nil, nil, nil)
			switch field {
			case "production.include":
				cfg.Production.Include = patterns
			case "production.exclude":
				cfg.Production.Exclude = patterns
			case "tests.include":
				cfg.Tests.Include = patterns
			case "tests.exclude":
				cfg.Tests.Exclude = patterns
			}
			if err := git.ValidateClassificationConfig(cfg); err == nil || !strings.Contains(err.Error(), "code_metrics."+field) {
				t.Errorf("%s %q: expected exact-setting error, got %v", field, patterns, err)
			}
		}
	}
	if _, err := git.ClassifyFile("src/main.go", nil); err == nil {
		t.Fatal("missing policy must fail rather than invent classification")
	}
	if err := git.ValidateClassificationConfig(classification(nil, nil, nil, nil)); err != nil {
		t.Fatalf("explicit empty lists are valid: %v", err)
	}
}

func TestClassifyFilePreservesFilenameWhitespace(t *testing.T) {
	t.Parallel()
	name := "runtime/ trailing space.custom "
	cfg := classification([]string{name}, nil, nil, nil)
	if got, err := git.ClassifyFile(name, cfg); err != nil || got != git.CategoryProd {
		t.Fatalf("literal whitespace path changed: %s, %v", got, err)
	}
	if got, err := git.ClassifyFile(strings.TrimSpace(name), cfg); err != nil || got != git.CategoryOther {
		t.Fatalf("trimmed filename unexpectedly matched: %s, %v", got, err)
	}
}
