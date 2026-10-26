package git_test

import (
	"testing"

	"agenticlens/internal/adapters/git"
)

func TestClassifyFile_WorkspacesDir(t *testing.T) {
	t.Run("custom workspaces directory is classified as CategoryIgnored", func(t *testing.T) {
		cfgTickets := &git.ClassificationConfig{
			WorkspacesDir: "tickets",
		}
		if cat := git.ClassifyFile("tickets/AL-39/CONTRACT.md", cfgTickets); cat != git.CategoryIgnored {
			t.Errorf("expected CategoryIgnored for tickets/AL-39/CONTRACT.md, got %v", cat)
		}
		if cat := git.ClassifyFile("tickets/AL-39/helper.go", cfgTickets); cat != git.CategoryIgnored {
			t.Errorf("expected CategoryIgnored for tickets/AL-39/helper.go, got %v", cat)
		}
		if cat := git.ClassifyFile("./tickets/AL-39/helper.go", cfgTickets); cat != git.CategoryIgnored {
			t.Errorf("expected CategoryIgnored for ./tickets/AL-39/helper.go, got %v", cat)
		}

		cfgCustomWs := &git.ClassificationConfig{
			WorkspacesDir: "custom_workspaces",
		}
		if cat := git.ClassifyFile("custom_workspaces/AL-39/PLAN.md", cfgCustomWs); cat != git.CategoryIgnored {
			t.Errorf("expected CategoryIgnored for custom_workspaces/AL-39/PLAN.md, got %v", cat)
		}
		if cat := git.ClassifyFile("custom_workspaces/AL-39/main.go", cfgCustomWs); cat != git.CategoryIgnored {
			t.Errorf("expected CategoryIgnored for custom_workspaces/AL-39/main.go, got %v", cat)
		}
	})

	t.Run("workspaces directory with whitespace and slashes is normalized", func(t *testing.T) {
		cfg := &git.ClassificationConfig{
			WorkspacesDir: "  /tickets/  ",
		}
		if cat := git.ClassifyFile("tickets/AL-39/data.json", cfg); cat != git.CategoryIgnored {
			t.Errorf("expected CategoryIgnored for trimmed workspaces dir, got %v", cat)
		}
		if cat := git.ClassifyFile("tickets/AL-39/feature.go", cfg); cat != git.CategoryIgnored {
			t.Errorf("expected CategoryIgnored for trimmed workspaces dir, got %v", cat)
		}
	})

	t.Run("workspaces directory without configuration falls through", func(t *testing.T) {
		category := git.ClassifyFile("workspaces/AL-39/CONTRACT.md", nil)
		if category != git.CategoryOther {
			t.Errorf("expected CategoryOther for unconfigured workspaces markdown, got %v", category)
		}

		categoryCode := git.ClassifyFile("workspaces/AL-39/main.go", nil)
		if categoryCode != git.CategoryProd {
			t.Errorf("expected CategoryProd for code inside unconfigured workspaces, got %v", categoryCode)
		}

		cfgWs := &git.ClassificationConfig{
			WorkspacesDir: "workspaces",
		}
		if cat := git.ClassifyFile("workspaces/AL-39/CONTRACT.md", cfgWs); cat != git.CategoryIgnored {
			t.Errorf("expected CategoryIgnored for configured workspaces dir, got %v", cat)
		}
		if cat := git.ClassifyFile("workspaces/AL-39/main.go", cfgWs); cat != git.CategoryIgnored {
			t.Errorf("expected CategoryIgnored for code inside configured workspaces dir, got %v", cat)
		}
	})

	t.Run("nested workspaces directory path is ignored", func(t *testing.T) {
		cfg := &git.ClassificationConfig{
			WorkspacesDir: "custom_ws",
		}
		category := git.ClassifyFile("nested/path/custom_ws/TICK-1/data.csv", cfg)
		if category != git.CategoryIgnored {
			t.Errorf("expected CategoryIgnored for nested workspaces path, got %v", category)
		}
	})

	t.Run("boundary paths starting with workspace name must not be falsely ignored", func(t *testing.T) {
		cfg := &git.ClassificationConfig{
			WorkspacesDir: "tickets",
		}
		// File named tickets_service.go is not inside tickets/ folder
		if cat := git.ClassifyFile("tickets_service.go", cfg); cat == git.CategoryIgnored {
			t.Errorf("prohibition violation: tickets_service.go falsely classified as ignored")
		}
		// File named tickets-component.tsx is not inside tickets/ folder
		if cat := git.ClassifyFile("tickets-component.tsx", cfg); cat == git.CategoryIgnored {
			t.Errorf("prohibition violation: tickets-component.tsx falsely classified as ignored")
		}
		// File named workspaces_archive.go with configured workspaces
		cfgWs := &git.ClassificationConfig{
			WorkspacesDir: "workspaces",
		}
		if cat := git.ClassifyFile("workspaces_archive.go", cfgWs); cat == git.CategoryIgnored {
			t.Errorf("prohibition violation: workspaces_archive.go falsely classified as ignored")
		}
	})
}

func TestClassifyFile_TestAndProd(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		cfg      *git.ClassificationConfig
		expected git.FileCategory
	}{
		{
			name:     "go production file",
			path:     "backend/internal/domain/config.go",
			cfg:      &git.ClassificationConfig{ProdPatterns: []string{"**/*.go", "**/*.tsx"}},
			expected: git.CategoryProd,
		},
		{
			name:     "go test file",
			path:     "backend/internal/domain/config_test.go",
			cfg:      &git.ClassificationConfig{TestPatterns: []string{"**/*_test.go", "**/*.test.tsx"}},
			expected: git.CategoryTest,
		},
		{
			name:     "typescript production file",
			path:     "web/src/components/MetricCard.tsx",
			cfg:      &git.ClassificationConfig{ProdPatterns: []string{"**/*.go", "**/*.tsx"}},
			expected: git.CategoryProd,
		},
		{
			name:     "typescript test file",
			path:     "web/src/components/MetricCard.test.tsx",
			cfg:      &git.ClassificationConfig{TestPatterns: []string{"**/*_test.go", "**/*.test.tsx"}},
			expected: git.CategoryTest,
		},
		{
			name:     "markdown doc is other when unconfigured",
			path:     "docs/README.md",
			cfg:      nil,
			expected: git.CategoryOther,
		},
		{
			name: "markdown doc is ignored when explicitly excluded",
			path: "docs/README.md",
			cfg: &git.ClassificationConfig{
				Exclude: []string{"docs/**"},
			},
			expected: git.CategoryIgnored,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := git.ClassifyFile(tc.path, tc.cfg)
			if actual != tc.expected {
				t.Errorf("ClassifyFile(%q) = %v, expected %v", tc.path, actual, tc.expected)
			}
		})
	}
}

func TestClassifyFile_AdversarialBoundaries(t *testing.T) {
	t.Parallel()

	t.Run("unconfigured workspaces directory classifies prod and markdown separately", func(t *testing.T) {
		t.Parallel()
		if cat := git.ClassifyFile("workspaces/file.go", nil); cat != git.CategoryProd {
			t.Errorf("expected CategoryProd for workspaces/file.go with nil config, got %v", cat)
		}
		if cat := git.ClassifyFile("workspaces/spec.md", nil); cat != git.CategoryOther {
			t.Errorf("expected CategoryOther for workspaces/spec.md with nil config, got %v", cat)
		}
	})

	t.Run("configured workspaces directory classifies both as CategoryIgnored", func(t *testing.T) {
		t.Parallel()
		cfg := &git.ClassificationConfig{WorkspacesDir: "workspaces"}
		if cat := git.ClassifyFile("workspaces/file.go", cfg); cat != git.CategoryIgnored {
			t.Errorf("expected CategoryIgnored for workspaces/file.go with configured WorkspacesDir, got %v", cat)
		}
		if cat := git.ClassifyFile("workspaces/spec.md", cfg); cat != git.CategoryIgnored {
			t.Errorf("expected CategoryIgnored for workspaces/spec.md with configured WorkspacesDir, got %v", cat)
		}
	})

	t.Run("removal of universal md json lock ignore rules classifies them as CategoryOther", func(t *testing.T) {
		t.Parallel()
		files := []string{"test.md", "data.json", "tool.lock"}
		for _, f := range files {
			if cat := git.ClassifyFile(f, nil); cat != git.CategoryOther {
				t.Errorf("expected CategoryOther for %s with nil config, got %v", f, cat)
			}
		}
	})

	t.Run("specific lockfiles return CategoryIgnored", func(t *testing.T) {
		t.Parallel()
		lockfiles := []string{"go.sum", "package-lock.json", "pnpm-lock.yaml", "yarn.lock"}
		for _, lf := range lockfiles {
			if cat := git.ClassifyFile(lf, nil); cat != git.CategoryIgnored {
				t.Errorf("expected CategoryIgnored for %s with nil config, got %v", lf, cat)
			}
		}
	})

	t.Run("dotfiles return CategoryIgnored", func(t *testing.T) {
		t.Parallel()
		dotfiles := []string{".gitignore", ".env"}
		for _, df := range dotfiles {
			if cat := git.ClassifyFile(df, nil); cat != git.CategoryIgnored {
				t.Errorf("expected CategoryIgnored for %s with nil config, got %v", df, cat)
			}
		}
	})
}

func TestClassifyFileTestdataFixtures(t *testing.T) {
	for _, path := range []string{"testdata/input.json", "src/parser/testdata/rollout.jsonl", "./src/testdata/nested/expected.txt", "src/testdata/example.go"} {
		if got := git.ClassifyFile(path, &git.ClassificationConfig{TestPatterns: []string{"**/testdata/**"}}); got != git.CategoryTest {
			t.Errorf("fixture %q: got %s, want test", path, got)
		}
	}
	for _, path := range []string{"src/testdata_backup/input.json", "src/mytestdata/input.json"} {
		if got := git.ClassifyFile(path, nil); got != git.CategoryOther {
			t.Errorf("non-fixture %q: got %s, want other", path, got)
		}
	}
	path := "src/testdata/input.json"
	for _, tc := range []struct {
		cfg  git.ClassificationConfig
		want git.FileCategory
	}{
		{git.ClassificationConfig{Exclude: []string{path}}, git.CategoryIgnored},
		{git.ClassificationConfig{ProdPatterns: []string{path}}, git.CategoryProd},
		{git.ClassificationConfig{WorkspacesDir: "src/testdata"}, git.CategoryIgnored},
	} {
		if got := git.ClassifyFile(path, &tc.cfg); got != tc.want {
			t.Errorf("configuration precedence: got %s, want %s", got, tc.want)
		}
	}
}

func TestClassificationFallbackRules(t *testing.T) {
	testCases := []struct {
		path     string
		expected git.FileCategory
	}{
		{"testdata/input.json", git.CategoryTest},
		{"__tests__/example.ts", git.CategoryTest},
		{"src/app_test.go", git.CategoryTest},
		{"web/app.test.tsx", git.CategoryTest},
		{"src/flow.spec.ts", git.CategoryTest},
		{"src/flow.spec.tsx", git.CategoryTest},
		{"backend/util_test.py", git.CategoryTest},
		{"backend/test_util.py", git.CategoryTest},
		{"src/app.go", git.CategoryProd},
		{"web/app.tsx", git.CategoryProd},
		{"web/app.ts", git.CategoryProd},
		{"web/styles.css", git.CategoryProd},
		{"web/index.html", git.CategoryProd},
		{"scripts/tool.py", git.CategoryProd},
		{"scripts/tool.js", git.CategoryProd},
		{"Makefile", git.CategoryOther},
		{"README.md", git.CategoryOther},
		{"docs/guide.txt", git.CategoryOther},
	}

	configs := []*git.ClassificationConfig{
		nil,
		{},
		{TestPatterns: []string{}, ProdPatterns: []string{}},
		{TestPatterns: []string{"", "  "}, ProdPatterns: []string{"", "  "}},
	}

	for _, cfg := range configs {
		for _, tc := range testCases {
			got := git.ClassifyFile(tc.path, cfg)
			if got != tc.expected {
				t.Errorf("unconfigured %q classified as %s, want %s (cfg: %+v)", tc.path, got, tc.expected, cfg)
			}
		}
	}

	t.Run("explicit test patterns override fallbacks", func(t *testing.T) {
		cfg := &git.ClassificationConfig{
			TestPatterns: []string{"custom/tests/**/*.go"},
		}
		// Standard _test.go is not matched by custom pattern, so falls through to prod fallback
		if got := git.ClassifyFile("src/app_test.go", cfg); got != git.CategoryProd {
			t.Errorf("expected CategoryProd when standard test not in explicit test patterns, got %s", got)
		}
		// Custom test pattern matches
		if got := git.ClassifyFile("custom/tests/helper.go", cfg); got != git.CategoryTest {
			t.Errorf("expected CategoryTest for custom test pattern, got %s", got)
		}
	})

	t.Run("explicit prod patterns override fallbacks", func(t *testing.T) {
		cfg := &git.ClassificationConfig{
			ProdPatterns: []string{"runtime/**/*.custom"},
		}
		// Standard .go file is not in explicit prod patterns, so falls through to other
		if got := git.ClassifyFile("src/app.go", cfg); got != git.CategoryOther {
			t.Errorf("expected CategoryOther when standard code not in explicit prod patterns, got %s", got)
		}
		// Custom prod pattern matches
		if got := git.ClassifyFile("runtime/main.custom", cfg); got != git.CategoryProd {
			t.Errorf("expected CategoryProd for custom prod pattern, got %s", got)
		}
	})
}

func TestConfiguredPathBoundariesAndMonitoring(t *testing.T) {
	cfg := &git.ClassificationConfig{
		SrcFolders:   []string{"application"},
		TestPatterns: []string{"checks/**/samples/**", "application/**/*_check.custom"},
		ProdPatterns: []string{"runtime/**/*.custom"},
		Include:      []string{"build.recipe"},
	}
	for _, tc := range []struct {
		path      string
		category  git.FileCategory
		monitored bool
	}{
		{"checks/samples/data.json", git.CategoryTest, true},
		{"checks/unit/samples/nested/data.json", git.CategoryTest, true},
		{"checks/unit/samples_backup/data.json", git.CategoryOther, false},
		{"other/checks/samples/data.json", git.CategoryOther, false},
		{"application/unit/flow_check.custom", git.CategoryTest, true},
		{"runtime/main.custom", git.CategoryProd, true},
		{"runtime/nested/main.custom", git.CategoryProd, true},
		{"runtime_backup/main.custom", git.CategoryOther, false},
		{"other/runtime/main.custom", git.CategoryOther, false},
		{"application/unclassified.txt", git.CategoryOther, true},
		{"build.recipe", git.CategoryProd, true},
	} {
		if got := git.ClassifyFile(tc.path, cfg); got != tc.category {
			t.Errorf("%s: category %s, want %s", tc.path, got, tc.category)
		}
		if got := git.IsMonitoredSourcePath(tc.path, cfg.SrcFolders, cfg); got != tc.monitored {
			t.Errorf("%s: monitored %v, want %v", tc.path, got, tc.monitored)
		}
	}
}
