package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestGenerateConfigYAML_Default(t *testing.T) {
	t.Parallel()

	defaultCfg := testProjectConfig()
	defaultCfg.RoleMappings = domain.RoleMappings{}
	defaultCfg.RoleOrder = []string{}
	data, err := adaptersconfig.GenerateConfigYAML(defaultCfg)
	if err != nil {
		t.Fatalf("GenerateConfigYAML failed: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "src_folders:") {
		t.Error("generated YAML must not contain workspaces.src_folders")
	}

	// Check top-level comments and keys
	expectedSubstrings := []string{
		"# AgenticLens Global Configuration",
		"project_name: AgenticLens",
		"ticket:",
		"  prefixes:",
		"    - prefix: AL",
		"      ticket_type: chore",
		"      ticket_type: docs",
		"  case_insensitive: true",
		"workspaces:",
		"database:",
		"  path: ~/.local/share/opencode/opencode.db",
		"  query_timeout_ms: 5000",
		"  busy_timeout_ms: 5000",
		"  read_only: true",
		"live:",
		"  active_threshold_seconds: 300",
		"  poll_interval_ms: 2000",
		"  max_trace_buffer_messages: 500",
		"anomaly_thresholds:",
		"  stall_threshold_seconds: 120",
		"  error_rate_threshold: 0.25",
		"  context_exhaustion_pct: 0.85",
		"  context_ceiling_warning_tokens: 160000",
		"  context_ceiling_critical_tokens: 200000",
		"code_metrics:",
		"  source_roots:",
		"    - web",
		"    - backend",
		"  include: []",
		"  exclude: []",
		"  test_patterns: []",
		"  prod_patterns: []",
		"role_mappings: {}",
		"role_order: []",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(content, sub) {
			t.Errorf("expected generated YAML to contain %q", sub)
		}
	}

	unexpectedSubstrings := []string{
		"healthy_ceiling_tokens:",
		"orchestrator: orchestrator",
		"planner: planner",
		"- orchestrator",
	}
	for _, sub := range unexpectedSubstrings {
		if strings.Contains(content, sub) {
			t.Errorf("expected generated YAML to NOT contain %q", sub)
		}
	}
}

func TestGenerateConfigYAML_CustomRoles(t *testing.T) {
	t.Parallel()

	cfg := testProjectConfig()
	data, err := adaptersconfig.GenerateConfigYAML(cfg)
	if err != nil {
		t.Fatalf("GenerateConfigYAML failed: %v", err)
	}

	content := string(data)
	expectedSubstrings := []string{
		"role_mappings:",
		"  coder: coder",
		"  orchestrator: orchestrator",
		"role_order:",
		"  - orchestrator",
		"  - planner",
		"  - coder",
		"  - tester",
		"  - reviewer",
	}
	for _, sub := range expectedSubstrings {
		if !strings.Contains(content, sub) {
			t.Errorf("expected generated YAML to contain %q", sub)
		}
	}
}

func TestGenerateConfigYAML_RoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("with populated roles", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgPath := filepath.Join(tempDir, ".agenticlens-config.yaml")

		defaultCfg := testProjectConfig()
		data, err := adaptersconfig.GenerateConfigYAML(defaultCfg)
		if err != nil {
			t.Fatalf("GenerateConfigYAML failed: %v", err)
		}

		if err := os.WriteFile(cfgPath, data, 0644); err != nil {
			t.Fatalf("writing temp config file: %v", err)
		}

		loaded, err := adaptersconfig.LoadFromFile(cfgPath)
		if err != nil {
			t.Fatalf("LoadFromFile failed on generated config: %v", err)
		}

		if loaded.ProjectName != defaultCfg.ProjectName {
			t.Errorf("expected project_name %q, got %q", defaultCfg.ProjectName, loaded.ProjectName)
		}
		if len(loaded.Ticket.Prefixes) != len(defaultCfg.Ticket.Prefixes) {
			t.Fatalf("expected %d prefixes, got %d", len(defaultCfg.Ticket.Prefixes), len(loaded.Ticket.Prefixes))
		}
		if len(loaded.CodeMetrics.SourceRoots) != len(defaultCfg.CodeMetrics.SourceRoots) || loaded.CodeMetrics.SourceRoots[0] != defaultCfg.CodeMetrics.SourceRoots[0] {
			t.Errorf("expected canonical source roots %v, got %v", defaultCfg.CodeMetrics.SourceRoots, loaded.CodeMetrics.SourceRoots)
		}
		if loaded.Workspaces.Dir != defaultCfg.Workspaces.Dir {
			t.Errorf("expected workspaces.dir %q, got %q", defaultCfg.Workspaces.Dir, loaded.Workspaces.Dir)
		}
		if len(loaded.Workspaces.Artifacts) != len(defaultCfg.Workspaces.Artifacts) {
			t.Errorf("expected %d workspace artifacts, got %d", len(defaultCfg.Workspaces.Artifacts), len(loaded.Workspaces.Artifacts))
		}
		if loaded.Database.Path != defaultCfg.Database.Path {
			t.Errorf("expected database.path %q, got %q", defaultCfg.Database.Path, loaded.Database.Path)
		}
		if loaded.Live.ActiveThresholdSeconds != defaultCfg.Live.ActiveThresholdSeconds {
			t.Errorf("expected live.active_threshold_seconds %d, got %d", defaultCfg.Live.ActiveThresholdSeconds, loaded.Live.ActiveThresholdSeconds)
		}
		if len(loaded.RoleMappings) != len(defaultCfg.RoleMappings) {
			t.Errorf("expected %d role mappings, got %d", len(defaultCfg.RoleMappings), len(loaded.RoleMappings))
		}
		if len(loaded.RoleOrder) != len(defaultCfg.RoleOrder) {
			t.Errorf("expected %d role order, got %d", len(defaultCfg.RoleOrder), len(loaded.RoleOrder))
		}
		if loaded.AnomalyThresholds.ContextCeilingCriticalTokens != defaultCfg.AnomalyThresholds.ContextCeilingCriticalTokens {
			t.Errorf("expected context_ceiling_critical_tokens %d, got %d", defaultCfg.AnomalyThresholds.ContextCeilingCriticalTokens, loaded.AnomalyThresholds.ContextCeilingCriticalTokens)
		}
		if loaded.AnomalyThresholds.ContextCeilingWarningTokens != defaultCfg.AnomalyThresholds.ContextCeilingWarningTokens {
			t.Errorf("expected context_ceiling_warning_tokens %d, got %d", defaultCfg.AnomalyThresholds.ContextCeilingWarningTokens, loaded.AnomalyThresholds.ContextCeilingWarningTokens)
		}
	})

	t.Run("with empty roles preserves empty collections round trip", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgPath := filepath.Join(tempDir, ".agenticlens-config.yaml")

		emptyCfg := domain.DefaultProjectConfig()
		emptyCfg.ProjectName = "EmptyRolesProject"
		emptyCfg.Ticket.Prefixes = []domain.TicketPrefix{{Prefix: "EMPTY"}}
		emptyCfg.CodeMetrics.SourceRoots = []string{"src"}

		data, err := adaptersconfig.GenerateConfigYAML(emptyCfg)
		if err != nil {
			t.Fatalf("GenerateConfigYAML failed: %v", err)
		}

		if err := os.WriteFile(cfgPath, data, 0644); err != nil {
			t.Fatalf("writing temp config file: %v", err)
		}

		loaded, err := adaptersconfig.LoadFromFile(cfgPath)
		if err != nil {
			t.Fatalf("LoadFromFile failed on generated config: %v", err)
		}

		if len(loaded.RoleMappings) != 0 || loaded.RoleMappings == nil {
			t.Errorf("expected empty non-nil RoleMappings, got %v", loaded.RoleMappings)
		}
		if len(loaded.RoleOrder) != 0 || loaded.RoleOrder == nil {
			t.Errorf("expected empty non-nil RoleOrder, got %v", loaded.RoleOrder)
		}
	})
}

func TestGenerateDefaultYAML_Overrides(t *testing.T) {
	t.Parallel()

	customPrefixes := []domain.TicketPrefix{
		{Prefix: "MYAPP", Description: "Primary development tickets", TicketType: "feat"},
		{Prefix: "GEN", Description: "General tasks", TicketType: "chore"},
	}
	customSrc := []string{"services", "frontend"}

	data, err := adaptersconfig.GenerateDefaultYAML("CustomProject", "custom_workspaces", "/var/data/opencode.db", customSrc, customPrefixes)
	if err != nil {
		t.Fatalf("GenerateDefaultYAML failed: %v", err)
	}

	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}

	loaded, err := adaptersconfig.LoadFromFile(cfgPath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if loaded.ProjectName != "CustomProject" {
		t.Errorf("expected project_name CustomProject, got %s", loaded.ProjectName)
	}
	if loaded.Workspaces.Dir != "custom_workspaces" {
		t.Errorf("expected workspaces.dir custom_workspaces, got %s", loaded.Workspaces.Dir)
	}
	if loaded.Database.Path != "/var/data/opencode.db" {
		t.Errorf("expected database.path /var/data/opencode.db, got %s", loaded.Database.Path)
	}
	if len(loaded.CodeMetrics.SourceRoots) != 2 || loaded.CodeMetrics.SourceRoots[0] != "services" {
		t.Errorf("unexpected code_metrics.source_roots: %v", loaded.CodeMetrics.SourceRoots)
	}
	if len(loaded.Ticket.Prefixes) != 2 || loaded.Ticket.Prefixes[0].Prefix != "MYAPP" {
		t.Errorf("unexpected prefixes: %v", loaded.Ticket.Prefixes)
	}
	if len(loaded.RoleMappings) != 0 {
		t.Errorf("expected empty RoleMappings, got %v", loaded.RoleMappings)
	}
	if len(loaded.RoleOrder) != 0 {
		t.Errorf("expected empty RoleOrder, got %v", loaded.RoleOrder)
	}
	if len(loaded.CodeMetrics.TestPatterns) != len(adaptersconfig.DefaultTestPatterns) {
		t.Errorf("expected %d default test_patterns, got %d", len(adaptersconfig.DefaultTestPatterns), len(loaded.CodeMetrics.TestPatterns))
	}
	if len(loaded.CodeMetrics.ProdPatterns) != len(adaptersconfig.DefaultProdPatterns) {
		t.Errorf("expected %d default prod_patterns, got %d", len(adaptersconfig.DefaultProdPatterns), len(loaded.CodeMetrics.ProdPatterns))
	}

	content := string(data)
	if !strings.Contains(content, "role_mappings: {}") {
		t.Errorf("expected role_mappings: {} in generated default YAML, got:\n%s", content)
	}
	if !strings.Contains(content, "role_order: []") {
		t.Errorf("expected role_order: [] in generated default YAML, got:\n%s", content)
	}
}

func TestGenerateDefaultConfigYAML(t *testing.T) {
	t.Parallel()

	prefixes := []domain.TicketPrefix{
		{Prefix: "DEMO", Description: "Demo prefix", TicketType: "feat"},
	}
	srcFolders := []string{"src", "pkg"}

	data1, err1 := adaptersconfig.GenerateDefaultYAML("DemoProject", "workspaces", "/var/db/demo.db", srcFolders, prefixes)
	if err1 != nil {
		t.Fatalf("GenerateDefaultYAML failed: %v", err1)
	}

	data2, err2 := adaptersconfig.GenerateDefaultConfigYAML("DemoProject", "workspaces", "/var/db/demo.db", srcFolders, prefixes)
	if err2 != nil {
		t.Fatalf("GenerateDefaultConfigYAML failed: %v", err2)
	}

	if string(data1) != string(data2) {
		t.Errorf("GenerateDefaultConfigYAML output does not match GenerateDefaultYAML output.\nGot:\n%s\nExpected:\n%s", string(data2), string(data1))
	}
}

func TestWriteConfigFile_CollisionAndForce(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, ".agenticlens-config.yaml")

	data1 := []byte("project_name: InitialProject\n")
	data2 := []byte("project_name: OverwrittenProject\n")

	// First write should succeed
	if err := adaptersconfig.WriteConfigFile(targetPath, data1, false); err != nil {
		t.Fatalf("first WriteConfigFile failed: %v", err)
	}

	// Second write without force should fail with ErrFileExists
	err := adaptersconfig.WriteConfigFile(targetPath, data2, false)
	if err == nil {
		t.Fatal("expected error on duplicate write without force, got nil")
	}
	if !errors.Is(err, adaptersconfig.ErrFileExists) {
		t.Errorf("expected ErrFileExists, got %v", err)
	}

	var fileExistsErr *adaptersconfig.FileExistsError
	if !errors.As(err, &fileExistsErr) {
		t.Errorf("expected *FileExistsError type, got %T", err)
	}

	// Verify original file was not overwritten
	readBack, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if !strings.Contains(string(readBack), "InitialProject") {
		t.Errorf("expected original content, got %s", string(readBack))
	}

	// Third write with force should succeed
	if err := adaptersconfig.WriteConfigFile(targetPath, data2, true); err != nil {
		t.Fatalf("WriteConfigFile with force failed: %v", err)
	}

	// Verify file was updated
	readBackAfterForce, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("reading file after force: %v", err)
	}
	if !strings.Contains(string(readBackAfterForce), "OverwrittenProject") {
		t.Errorf("expected overwritten content, got %s", string(readBackAfterForce))
	}
}

func TestWriteConfigFile_PermissionHandlingAndInvalidPath(t *testing.T) {
	t.Parallel()

	// Writing to a path under a non-creatable or read-only directory
	tempDir := t.TempDir()
	readOnlyDir := filepath.Join(tempDir, "readonly")
	if err := os.MkdirAll(readOnlyDir, 0555); err != nil {
		t.Fatalf("creating readonly dir: %v", err)
	}
	defer os.Chmod(readOnlyDir, 0755) // Ensure cleanup succeeds

	targetPath := filepath.Join(readOnlyDir, "config.yaml")
	data := []byte("project_name: PermTest\n")

	// In non-root environments, creating a temp file inside a read-only dir should return an error
	if os.Geteuid() != 0 {
		err := adaptersconfig.WriteConfigFile(targetPath, data, false)
		if err == nil {
			t.Errorf("expected permission error when writing to readonly dir, got nil")
		}
	}
}

func TestGenerateDefaultYAML_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		projectName  string
		wsDir        string
		dbPath       string
		srcFolders   []string
		prefixes     []domain.TicketPrefix
		checkStrings []string
	}{
		{
			name:        "spaces in project name",
			projectName: "Enterprise Core Platform",
			wsDir:       "my workspaces/dir",
			dbPath:      "/path with spaces/opencode.db",
			srcFolders:  []string{"services/app one", "frontend/ui module"},
			prefixes: []domain.TicketPrefix{
				{Prefix: "ECP", Description: "Enterprise Core Platform", TicketType: "feat"},
			},
			checkStrings: []string{
				"project_name: Enterprise Core Platform",
				"dir: my workspaces/dir",
				"path: /path with spaces/opencode.db",
				"- services/app one",
				"- frontend/ui module",
				"prefix: ECP",
			},
		},
		{
			name:        "custom prefixes with special chars",
			projectName: "Custom-App_V2",
			wsDir:       "workspaces",
			dbPath:      "/tmp/db.sqlite",
			srcFolders:  []string{"pkg"},
			prefixes: []domain.TicketPrefix{
				{Prefix: "CUSTOM_SYS", Description: "Subsystem tickets", TicketType: "refactor"},
				{Prefix: "BUG-FIX", Description: "Bug fixes", TicketType: "fix"},
			},
			checkStrings: []string{
				"project_name: Custom-App_V2",
				"prefix: CUSTOM_SYS",
				"description: Subsystem tickets",
				"ticket_type: refactor",
				"prefix: BUG-FIX",
				"ticket_type: fix",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			yamlData, err := adaptersconfig.GenerateDefaultYAML(tc.projectName, tc.wsDir, tc.dbPath, tc.srcFolders, tc.prefixes)
			if err != nil {
				t.Fatalf("GenerateDefaultYAML failed: %v", err)
			}
			content := string(yamlData)
			for _, check := range tc.checkStrings {
				if !strings.Contains(content, check) {
					t.Errorf("expected generated YAML to contain %q, but was not found in:\n%s", check, content)
				}
			}

			// Validate unmarshaling back into domain.ProjectConfig
			tempDir := t.TempDir()
			cfgPath := filepath.Join(tempDir, "test-config.yaml")
			if err := os.WriteFile(cfgPath, yamlData, 0644); err != nil {
				t.Fatalf("writing temp config: %v", err)
			}
			loaded, err := adaptersconfig.LoadFromFile(cfgPath)
			if err != nil {
				t.Fatalf("LoadFromFile failed on generated YAML: %v", err)
			}
			if loaded.ProjectName != tc.projectName {
				t.Errorf("expected loaded ProjectName %q, got %q", tc.projectName, loaded.ProjectName)
			}
		})
	}
}

func TestGenerateDefaultYAML_ValidationErrors(t *testing.T) {
	t.Parallel()

	validPrefixes := []domain.TicketPrefix{{Prefix: "AL"}}
	validSrc := []string{"src"}

	t.Run("empty project name returns error", func(t *testing.T) {
		_, err := adaptersconfig.GenerateDefaultYAML("   ", "workspaces", "", validSrc, validPrefixes)
		if err == nil {
			t.Fatalf("expected error for empty project name, got nil")
		}
	})

	t.Run("empty source folders returns error", func(t *testing.T) {
		_, err := adaptersconfig.GenerateDefaultYAML("TestProj", "workspaces", "", nil, validPrefixes)
		if err == nil {
			t.Fatalf("expected error for empty source folders, got nil")
		}
		_, err = adaptersconfig.GenerateDefaultYAML("TestProj", "workspaces", "", []string{" ", ""}, validPrefixes)
		if err == nil {
			t.Fatalf("expected error for blank source folders, got nil")
		}
	})

	t.Run("empty ticket prefixes returns error", func(t *testing.T) {
		_, err := adaptersconfig.GenerateDefaultYAML("TestProj", "workspaces", "", validSrc, nil)
		if err == nil {
			t.Fatalf("expected error for empty ticket prefixes, got nil")
		}
		_, err = adaptersconfig.GenerateDefaultYAML("TestProj", "workspaces", "", validSrc, []domain.TicketPrefix{{Prefix: " "}})
		if err == nil {
			t.Fatalf("expected error for blank ticket prefixes, got nil")
		}
	})
}

func TestGenerateConfigYAML_ValidationErrors(t *testing.T) {
	t.Parallel()

	t.Run("empty project name returns error", func(t *testing.T) {
		cfg := testProjectConfig()
		cfg.ProjectName = ""
		_, err := adaptersconfig.GenerateConfigYAML(cfg)
		if err == nil {
			t.Fatalf("expected error for empty project name, got nil")
		}
	})

	t.Run("empty prefixes returns error", func(t *testing.T) {
		cfg := testProjectConfig()
		cfg.Ticket.Prefixes = []domain.TicketPrefix{}
		_, err := adaptersconfig.GenerateConfigYAML(cfg)
		if err == nil {
			t.Fatalf("expected error for empty prefixes, got nil")
		}
	})

	t.Run("empty canonical source roots returns error", func(t *testing.T) {
		cfg := testProjectConfig()
		cfg.CodeMetrics.SourceRoots = []string{}
		_, err := adaptersconfig.GenerateConfigYAML(cfg)
		if err == nil {
			t.Fatalf("expected error for empty code_metrics.source_roots, got nil")
		}
	})

	t.Run("empty artifacts serializes empty list", func(t *testing.T) {
		cfg := testProjectConfig()
		cfg.Workspaces.Artifacts = []domain.WorkspaceArtifactConfig{}
		data, err := adaptersconfig.GenerateConfigYAML(cfg)
		if err != nil {
			t.Fatalf("GenerateConfigYAML failed: %v", err)
		}
		content := string(data)
		if !strings.Contains(content, "artifacts: []") {
			t.Errorf("expected 'artifacts: []' in YAML, got:\n%s", content)
		}
	})
}

func TestGenerateConfigYAML_WorkspaceArtifacts(t *testing.T) {
	t.Parallel()

	cfg := testProjectConfig()
	cfg.Workspaces.Artifacts = []domain.WorkspaceArtifactConfig{
		{
			Path:        "DOC.md",
			Title:       "Custom Documentation",
			Description: "Custom documentation guide",
		},
	}

	data, err := adaptersconfig.GenerateConfigYAML(cfg)
	if err != nil {
		t.Fatalf("GenerateConfigYAML failed: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "- path: DOC.md") || !strings.Contains(content, "title: Custom Documentation") {
		t.Errorf("expected artifacts in yaml, got: %s", content)
	}

	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, ".agenticlens-config.yaml")
	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	loaded, err := adaptersconfig.LoadFromFile(cfgPath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}
	if len(loaded.Workspaces.Artifacts) != 1 || loaded.Workspaces.Artifacts[0].Path != "DOC.md" {
		t.Errorf("expected 1 artifact DOC.md, got: %+v", loaded.Workspaces.Artifacts)
	}
	if loaded.Workspaces.Artifacts[0].Title != "Custom Documentation" {
		t.Errorf("expected title 'Custom Documentation', got %q", loaded.Workspaces.Artifacts[0].Title)
	}
}
