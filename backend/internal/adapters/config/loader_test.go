package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

const (
	testValidSuffixYAML = "\nticket:\n  prefixes:\n    - prefix: AL\nworkspaces:\n  src_folders:\n    - src\n"
	testValidSuffixJSON = `,"ticket":{"prefixes":[{"prefix":"AL"}]},"workspaces":{"src_folders":["src"]}}`
)

func testProjectConfig() domain.ProjectConfig {
	return domain.ProjectConfig{
		ProjectName: "AgenticLens",
		Ticket: domain.TicketConfig{
			Prefixes: []domain.TicketPrefix{
				{
					Prefix:      "AL",
					Description: "Core AgenticLens development features, bugfixes, and technical initiatives",
				},
				{
					Prefix:      "GEN",
					Description: "General multi-agent engineering tasks and ad-hoc untracked work",
					TicketType:  "chore",
				},
				{
					Prefix:      "PO",
					Description: "Product Owner specifications, requirements definitions, and roadmap breakdowns",
					TicketType:  "docs",
				},
			},
			CaseInsensitive: true,
			CustomPatterns:  []string{},
		},
		Workspaces: domain.WorkspacesConfig{
			Dir:        "workspaces",
			SrcFolders: []string{"web", "backend"},
			Artifacts:  []domain.WorkspaceArtifactConfig{},
		},
		Database: domain.DatabaseConfig{
			Path:           "~/.local/share/opencode/opencode.db",
			QueryTimeoutMs: 5000,
			BusyTimeoutMs:  5000,
			ReadOnly:       true,
		},
		Live: domain.LiveConfig{
			ActiveThresholdSeconds: 300,
			PollIntervalMs:         2000,
			MaxTraceBufferMessages: 500,
			AnomalyThresholds: domain.AnomalyThresholdsConfig{
				StallThresholdSeconds: 120,
				ErrorRateThreshold:    0.25,
				ContextExhaustionPct:  0.85,
			},
		},
		CodeMetrics: domain.CodeMetricsConfig{
			Include:      []string{},
			Exclude:      []string{},
			TestPatterns: []string{},
			ProdPatterns: []string{},
		},
		RoleMappings: domain.RoleMappings{
			"orchestrator": "orchestrator",
			"build":        "orchestrator",
			"planner":      "planner",
			"coder":        "coder",
			"tester":       "tester",
			"reviewer":     "reviewer",
		},
		RoleOrder: []string{
			"orchestrator",
			"planner",
			"coder",
			"tester",
			"reviewer",
		},
	}
}

func TestLoad_YAMLConfigurationParsing(t *testing.T) {
	tempDir := t.TempDir()
	cfgDir := filepath.Join(tempDir, ".agenticlens")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}

	yamlContent := `
project_name: CustomProject
ticket:
  prefixes:
    - prefix: CUSTOM
      description: Custom initiatives
    - prefix: TICK
      description: Ticket trackers
  case_insensitive: false
  custom_patterns:
    - PAT-(\d+)
workspaces:
  dir: custom_workspaces
  src_folders:
    - frontend
    - services
database:
  path: /tmp/test.db
  query_timeout_ms: 8000
  busy_timeout_ms: 9000
  read_only: false
live:
  active_threshold_seconds: 400
  poll_interval_ms: 3000
  max_trace_buffer_messages: 800
  anomaly_thresholds:
    stall_threshold_seconds: 150
    error_rate_threshold: 0.35
    context_exhaustion_pct: 0.90
code_metrics:
  include:
    - "services/**"
  exclude:
    - "legacy/**"
  test_patterns:
    - "*_spec.go"
  prod_patterns:
    - "*.go"
role_mappings:
  architect: planner
  qa: tester
role_order:
  - tester
  - planner
  - coder
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.ProjectName != "CustomProject" {
		t.Errorf("expected CustomProject, got %s", cfg.ProjectName)
	}
	if len(cfg.Ticket.Prefixes) != 2 || cfg.Ticket.Prefixes[0].Prefix != "CUSTOM" || cfg.Ticket.Prefixes[1].Prefix != "TICK" {
		t.Errorf("unexpected ticket prefixes: %v", cfg.Ticket.Prefixes)
	}
	if cfg.Ticket.Prefixes[0].Description != "Custom initiatives" {
		t.Errorf("expected description 'Custom initiatives', got %s", cfg.Ticket.Prefixes[0].Description)
	}
	if cfg.Ticket.Prefixes[1].Description != "Ticket trackers" {
		t.Errorf("expected description 'Ticket trackers', got %s", cfg.Ticket.Prefixes[1].Description)
	}
	names := cfg.Ticket.PrefixNames()
	if len(names) != 2 || names[0] != "CUSTOM" || names[1] != "TICK" {
		t.Errorf("unexpected PrefixNames: %v", names)
	}
	if cfg.Ticket.CaseInsensitive {
		t.Errorf("expected CaseInsensitive to be false")
	}
	if len(cfg.Ticket.CustomPatterns) != 1 || cfg.Ticket.CustomPatterns[0] != `PAT-(\d+)` {
		t.Errorf("unexpected custom patterns: %v", cfg.Ticket.CustomPatterns)
	}
	if cfg.Workspaces.Dir != "custom_workspaces" {
		t.Errorf("expected custom_workspaces, got %s", cfg.Workspaces.Dir)
	}
	if len(cfg.Workspaces.SrcFolders) != 2 || cfg.Workspaces.SrcFolders[0] != "frontend" || cfg.Workspaces.SrcFolders[1] != "services" {
		t.Errorf("unexpected src_folders: %v", cfg.Workspaces.SrcFolders)
	}
	if cfg.Database.Path != "/tmp/test.db" {
		t.Errorf("unexpected Database.Path: %s", cfg.Database.Path)
	}
	if cfg.Database.QueryTimeoutMs != 8000 || cfg.Database.BusyTimeoutMs != 9000 || cfg.Database.ReadOnly {
		t.Errorf("unexpected database config: %+v", cfg.Database)
	}
	if cfg.Live.ActiveThresholdSeconds != 400 || cfg.Live.PollIntervalMs != 3000 || cfg.Live.MaxTraceBufferMessages != 800 {
		t.Errorf("unexpected live config: %+v", cfg.Live)
	}
	if cfg.Live.AnomalyThresholds.StallThresholdSeconds != 150 || cfg.Live.AnomalyThresholds.ErrorRateThreshold != 0.35 || cfg.Live.AnomalyThresholds.ContextExhaustionPct != 0.90 {
		t.Errorf("unexpected anomaly thresholds: %+v", cfg.Live.AnomalyThresholds)
	}
	if len(cfg.CodeMetrics.Include) != 1 || cfg.CodeMetrics.Include[0] != "services/**" {
		t.Errorf("unexpected code metrics include: %v", cfg.CodeMetrics.Include)
	}
	if cfg.RoleMappings.ResolveRole("architect") != domain.RolePlanner {
		t.Errorf("expected RolePlanner for architect, got %s", cfg.RoleMappings.ResolveRole("architect"))
	}
	if cfg.RoleMappings.ResolveRole("qa") != domain.RoleTester {
		t.Errorf("expected RoleTester for qa, got %s", cfg.RoleMappings.ResolveRole("qa"))
	}
	if len(cfg.RoleOrder) != 3 || cfg.RoleOrder[0] != "tester" || cfg.RoleOrder[1] != "planner" || cfg.RoleOrder[2] != "coder" {
		t.Errorf("unexpected RoleOrder: %v", cfg.RoleOrder)
	}
}

func TestLoad_FilePriorityResolution(t *testing.T) {
	t.Run("yaml takes precedence over json", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		yamlContent := `project_name: FromYAML` + testValidSuffixYAML
		jsonContent := `{"project_name": "FromJSON"` + testValidSuffixJSON

		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(jsonContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "FromYAML" {
			t.Errorf("expected FromYAML, got %s", cfg.ProjectName)
		}
	})

	t.Run("yaml takes precedence over yml", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		yamlContent := `project_name: FromYAML` + testValidSuffixYAML
		ymlContent := `project_name: FromYML` + testValidSuffixYAML

		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yml"), []byte(ymlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "FromYAML" {
			t.Errorf("expected FromYAML, got %s", cfg.ProjectName)
		}
	})

	t.Run("yml takes precedence over json", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		ymlContent := `project_name: FromYML` + testValidSuffixYAML
		jsonContent := `{"project_name": "FromJSON"` + testValidSuffixJSON

		if err := os.WriteFile(filepath.Join(cfgDir, "config.yml"), []byte(ymlContent), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(jsonContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "FromYML" {
			t.Errorf("expected FromYML, got %s", cfg.ProjectName)
		}
	})

	t.Run("json used when yaml absent", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		jsonContent := `{"project_name": "FromJSON"` + testValidSuffixJSON
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(jsonContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "FromJSON" {
			t.Errorf("expected FromJSON, got %s", cfg.ProjectName)
		}
	})
}

func TestLoad_LegacyJSONRejected(t *testing.T) {
	tempDir := t.TempDir()
	cfgDir := filepath.Join(tempDir, ".agenticlens")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}

	legacyJSON := `{
		"project_name": "LegacyProject",
		"ticket_prefix": "LEG",
		"workspaces_dir": "custom_workspaces_dir",
		"src_folders": ["api", "ui"]
	}`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(legacyJSON), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := config.Load(tempDir)
	if err == nil {
		t.Fatalf("expected error loading legacy json without canonical sections, got nil")
	}
	if !errors.Is(err, config.ErrInvalidConfig) {
		t.Errorf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestLoad_MissingConfigurationReturnsErrConfigNotFound(t *testing.T) {
	tempDir := t.TempDir()

	_, err := config.Load(tempDir)
	if err == nil {
		t.Fatalf("expected error on missing config, got nil")
	}
	if !errors.Is(err, config.ErrConfigNotFound) {
		t.Errorf("expected ErrConfigNotFound, got %v", err)
	}
}

func TestLoad_PartialConfigurationMerging(t *testing.T) {
	tempDir := t.TempDir()
	cfgDir := filepath.Join(tempDir, ".agenticlens")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}

	partialYAML := `
project_name: PartialLens
ticket:
  prefixes:
    - prefix: AL
workspaces:
  src_folders:
    - src
database:
  path: /custom/path/db.sqlite
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(partialYAML), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ProjectName != "PartialLens" {
		t.Errorf("expected PartialLens, got %s", cfg.ProjectName)
	}
	if cfg.Database.Path != "/custom/path/db.sqlite" {
		t.Errorf("expected /custom/path/db.sqlite, got %s", cfg.Database.Path)
	}

	// Unspecified workspaces.dir retains empty string without silent defaulting
	if cfg.Workspaces.Dir != "" {
		t.Errorf("expected empty workspaces dir when omitted, got %s", cfg.Workspaces.Dir)
	}
	if len(cfg.Workspaces.Artifacts) != 0 {
		t.Errorf("expected empty artifacts, got %v", cfg.Workspaces.Artifacts)
	}
	if cfg.Database.QueryTimeoutMs != 5000 {
		t.Errorf("expected default query timeout 5000, got %d", cfg.Database.QueryTimeoutMs)
	}
	if cfg.Live.ActiveThresholdSeconds != 300 {
		t.Errorf("expected default active threshold 300, got %d", cfg.Live.ActiveThresholdSeconds)
	}
	if len(cfg.RoleMappings) != 0 || cfg.RoleMappings == nil {
		t.Errorf("expected empty non-nil RoleMappings, got %v", cfg.RoleMappings)
	}
	if len(cfg.RoleOrder) != 0 || cfg.RoleOrder == nil {
		t.Errorf("expected empty non-nil RoleOrder, got %v", cfg.RoleOrder)
	}
}

func TestLoad_MalformedFileErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	cfgDir := filepath.Join(tempDir, ".agenticlens")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}

	malformedYAML := `
project_name: [invalid yaml unclosed
  test: {
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(malformedYAML), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := config.Load(tempDir)
	if err == nil {
		t.Errorf("expected error loading malformed yaml, got nil")
	}

	malformedJSON := `{"project_name": "broken", invalid}`
	_ = os.Remove(filepath.Join(cfgDir, "config.yaml"))
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(malformedJSON), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = config.Load(tempDir)
	if err == nil {
		t.Errorf("expected error loading malformed json, got nil")
	}
}

func TestFindConfigFile_UpwardTraversal(t *testing.T) {
	tempDir := t.TempDir()
	cfgDir := filepath.Join(tempDir, ".agenticlens")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("project_name: Root"), 0644); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(tempDir, "sub1", "sub2", "sub3")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	found, err := config.FindConfigFile(subDir)
	if err != nil {
		t.Fatalf("unexpected error finding config file: %v", err)
	}
	expected := filepath.Join(cfgDir, "config.yaml")
	if found != expected {
		t.Errorf("expected %s, got %s", expected, found)
	}
}

func TestLoad_AdversarialDiscoveryAndFallbackOrder(t *testing.T) {
	t.Run("all three formats present selects yaml", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("project_name: FromYAML"+testValidSuffixYAML), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yml"), []byte("project_name: FromYML"+testValidSuffixYAML), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"project_name": "FromJSON"`+testValidSuffixJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "FromYAML" {
			t.Errorf("expected FromYAML, got %s", cfg.ProjectName)
		}
	})

	t.Run("yml and json present selects yml", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(cfgDir, "config.yml"), []byte("project_name: FromYML"+testValidSuffixYAML), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"project_name": "FromJSON"`+testValidSuffixJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "FromYML" {
			t.Errorf("expected FromYML, got %s", cfg.ProjectName)
		}
	})
}

func TestLoad_AdversarialErrorWrapping(t *testing.T) {
	tempDir := t.TempDir()
	cfgDir := filepath.Join(tempDir, ".agenticlens")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Run("malformed yaml wrapped error", func(t *testing.T) {
		yamlPath := filepath.Join(cfgDir, "config.yaml")
		if err := os.WriteFile(yamlPath, []byte("project_name: [unclosed"), 0644); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(yamlPath)

		_, err := config.LoadFromFile(yamlPath)
		if err == nil {
			t.Fatal("expected error on malformed yaml, got nil")
		}
		if !containsSubstr(err.Error(), "decoding yaml configuration") {
			t.Errorf("expected descriptive wrapped error mentioning yaml decoding, got %v", err)
		}
	})

	t.Run("malformed yml wrapped error", func(t *testing.T) {
		ymlPath := filepath.Join(cfgDir, "config.yml")
		if err := os.WriteFile(ymlPath, []byte("project_name: [unclosed"), 0644); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(ymlPath)

		_, err := config.LoadFromFile(ymlPath)
		if err == nil {
			t.Fatal("expected error on malformed yml, got nil")
		}
		if !containsSubstr(err.Error(), "decoding yaml configuration") {
			t.Errorf("expected descriptive wrapped error mentioning yaml decoding, got %v", err)
		}
	})

	t.Run("malformed json wrapped error", func(t *testing.T) {
		jsonPath := filepath.Join(cfgDir, "config.json")
		if err := os.WriteFile(jsonPath, []byte(`{"project_name": broken}`), 0644); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(jsonPath)

		_, err := config.LoadFromFile(jsonPath)
		if err == nil {
			t.Fatal("expected error on malformed json, got nil")
		}
		if !containsSubstr(err.Error(), "decoding json configuration") {
			t.Errorf("expected descriptive wrapped error mentioning json decoding, got %v", err)
		}
	})

	t.Run("nonexistent file wrapped read error", func(t *testing.T) {
		nonexistent := filepath.Join(tempDir, "does_not_exist.yaml")
		_, err := config.LoadFromFile(nonexistent)
		if err == nil {
			t.Fatal("expected error for nonexistent file, got nil")
		}
		if !containsSubstr(err.Error(), "reading configuration file") {
			t.Errorf("expected descriptive wrapped error mentioning reading file, got %v", err)
		}
	})
}

func TestLoad_AdversarialPartialConfigs(t *testing.T) {
	t.Run("zero byte empty file returns validation error", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error on empty file, got nil")
		}
		if !errors.Is(err, config.ErrInvalidConfig) {
			t.Errorf("expected ErrInvalidConfig, got %v", err)
		}
	})

	t.Run("empty object returns validation error", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("{}\n"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error on empty object, got nil")
		}
		if !errors.Is(err, config.ErrInvalidConfig) {
			t.Errorf("expected ErrInvalidConfig, got %v", err)
		}
	})

	t.Run("missing project_name returns validation error", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}
		yamlContent := "ticket:\n  prefixes:\n    - prefix: AL\nworkspaces:\n  src_folders:\n    - src\n"
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error for missing project_name, got nil")
		}
		if !errors.Is(err, config.ErrInvalidConfig) {
			t.Errorf("expected ErrInvalidConfig, got %v", err)
		}
	})

	t.Run("missing prefixes returns validation error", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}
		yamlContent := "project_name: TestProject\nworkspaces:\n  src_folders:\n    - src\n"
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error for missing prefixes, got nil")
		}
		if !errors.Is(err, config.ErrInvalidConfig) {
			t.Errorf("expected ErrInvalidConfig, got %v", err)
		}
	})

	t.Run("missing src_folders returns validation error", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}
		yamlContent := "project_name: TestProject\nticket:\n  prefixes:\n    - prefix: AL\n"
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error for missing src_folders, got nil")
		}
		if !errors.Is(err, config.ErrInvalidConfig) {
			t.Errorf("expected ErrInvalidConfig, got %v", err)
		}
	})

	t.Run("nested single field overrides retain other defaults", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		partialContent := `
project_name: CustomApp
ticket:
  case_insensitive: false
  prefixes:
    - prefix: AL
workspaces:
  dir: custom_workspaces_only
  src_folders:
    - src
database:
  read_only: false
live:
  poll_interval_ms: 1234
  anomaly_thresholds:
    stall_threshold_seconds: 45
`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(partialContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verified overridden fields
		if cfg.Ticket.CaseInsensitive {
			t.Errorf("expected CaseInsensitive to be false")
		}
		if cfg.Workspaces.Dir != "custom_workspaces_only" {
			t.Errorf("expected custom_workspaces_only, got %s", cfg.Workspaces.Dir)
		}
		if cfg.Database.ReadOnly {
			t.Errorf("expected ReadOnly to be false")
		}
		if cfg.Live.PollIntervalMs != 1234 {
			t.Errorf("expected PollIntervalMs 1234, got %d", cfg.Live.PollIntervalMs)
		}
		if cfg.Live.AnomalyThresholds.StallThresholdSeconds != 45 {
			t.Errorf("expected StallThresholdSeconds 45, got %d", cfg.Live.AnomalyThresholds.StallThresholdSeconds)
		}

		// Explicitly configured fields match
		if len(cfg.Ticket.Prefixes) != 1 || cfg.Ticket.Prefixes[0].Prefix != "AL" {
			t.Errorf("unexpected ticket prefixes: %v", cfg.Ticket.Prefixes)
		}
		if len(cfg.Workspaces.SrcFolders) != 1 || cfg.Workspaces.SrcFolders[0] != "src" {
			t.Errorf("unexpected src_folders: %v", cfg.Workspaces.SrcFolders)
		}
		if len(cfg.Workspaces.Artifacts) != 0 {
			t.Errorf("expected zero default artifacts, got: %v", cfg.Workspaces.Artifacts)
		}
		if cfg.Database.Path != "~/.local/share/opencode/opencode.db" || cfg.Database.QueryTimeoutMs != 5000 || cfg.Database.BusyTimeoutMs != 5000 {
			t.Errorf("unexpected database defaults: %+v", cfg.Database)
		}
		if cfg.Live.ActiveThresholdSeconds != 300 || cfg.Live.MaxTraceBufferMessages != 500 {
			t.Errorf("unexpected live defaults: %+v", cfg.Live)
		}
		if cfg.Live.AnomalyThresholds.ErrorRateThreshold != 0.25 || cfg.Live.AnomalyThresholds.ContextExhaustionPct != 0.85 {
			t.Errorf("unexpected anomaly threshold defaults: %+v", cfg.Live.AnomalyThresholds)
		}
	})
}

func TestLoad_AdversarialLegacyJSONTranslationAndPriority(t *testing.T) {
	t.Run("legacy keys translate and empty values fail required validation", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		legacyEmptyJSON := `{
			"ticket_prefix": "   ",
			"workspaces_dir": "",
			"src_folders": []
		}`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(legacyEmptyJSON), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error on empty legacy config, got nil")
		}
		if !errors.Is(err, config.ErrInvalidConfig) {
			t.Errorf("expected ErrInvalidConfig, got %v", err)
		}
	})

	t.Run("modern nested keys supersede legacy root keys", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		mixedJSON := `{
			"project_name": "ModernSupersede",
			"ticket_prefix": "LEGACY",
			"workspaces_dir": "legacy_workspaces",
			"src_folders": ["legacy_src"],
			"ticket": {
				"prefixes": ["MODERN"]
			},
			"workspaces": {
				"dir": "modern_workspaces",
				"src_folders": ["modern_web", "modern_backend"]
			}
		}`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(mixedJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Ticket.Prefixes) != 1 || cfg.Ticket.Prefixes[0].Prefix != "MODERN" {
			t.Errorf("expected MODERN ticket prefix, got %v", cfg.Ticket.Prefixes)
		}
		if cfg.Workspaces.Dir != "modern_workspaces" {
			t.Errorf("expected modern_workspaces, got %s", cfg.Workspaces.Dir)
		}
		if len(cfg.Workspaces.SrcFolders) != 2 || cfg.Workspaces.SrcFolders[0] != "modern_web" {
			t.Errorf("expected modern src folders, got %v", cfg.Workspaces.SrcFolders)
		}
	})
}

func TestLoad_AdversarialNonexistentPathsResilience(t *testing.T) {
	t.Run("nonexistent directory returns ErrConfigNotFound", func(t *testing.T) {
		nonexistentDir := filepath.Join(os.TempDir(), "agenticlens_nonexistent_test_dir_12345")
		_ = os.RemoveAll(nonexistentDir)

		_, err := config.Load(nonexistentDir)
		if err == nil {
			t.Fatalf("expected error on nonexistent directory, got nil")
		}
		if !errors.Is(err, config.ErrConfigNotFound) {
			t.Errorf("expected ErrConfigNotFound, got %v", err)
		}
	})

	t.Run("FindConfigFile on nonexistent directory returns empty string and nil error", func(t *testing.T) {
		nonexistentDir := filepath.Join(os.TempDir(), "agenticlens_nonexistent_test_dir_67890")
		_ = os.RemoveAll(nonexistentDir)

		path, err := config.FindConfigFile(nonexistentDir)
		if err != nil {
			t.Fatalf("expected nil error for FindConfigFile on nonexistent dir, got %v", err)
		}
		if path != "" {
			t.Errorf("expected empty string path for nonexistent directory, got %s", path)
		}
	})
}

func containsSubstr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || filepath.Base(s) == substr || (len(substr) > 0 && searchSubstr(s, substr)))
}

func searchSubstr(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func assertConfigMatchesDefaults(t *testing.T, actual, expected domain.ProjectConfig) {
	t.Helper()
	if actual.ProjectName != expected.ProjectName {
		t.Errorf("ProjectName mismatch: got %s, expected %s", actual.ProjectName, expected.ProjectName)
	}
	if actual.Workspaces.Dir != expected.Workspaces.Dir {
		t.Errorf("Workspaces.Dir mismatch: got %s, expected %s", actual.Workspaces.Dir, expected.Workspaces.Dir)
	}
	if len(actual.Ticket.Prefixes) != len(expected.Ticket.Prefixes) {
		t.Errorf("Ticket.Prefixes length mismatch: got %v, expected %v", actual.Ticket.Prefixes, expected.Ticket.Prefixes)
	}
	for i := 0; i < len(actual.Ticket.Prefixes) && i < len(expected.Ticket.Prefixes); i++ {
		if actual.Ticket.Prefixes[i] != expected.Ticket.Prefixes[i] {
			t.Errorf("Ticket.Prefixes[%d] mismatch: got %+v, expected %+v", i, actual.Ticket.Prefixes[i], expected.Ticket.Prefixes[i])
		}
	}
	if actual.Database.QueryTimeoutMs != expected.Database.QueryTimeoutMs {
		t.Errorf("Database.QueryTimeoutMs mismatch: got %d, expected %d", actual.Database.QueryTimeoutMs, expected.Database.QueryTimeoutMs)
	}
	if actual.Live.ActiveThresholdSeconds != expected.Live.ActiveThresholdSeconds {
		t.Errorf("Live.ActiveThresholdSeconds mismatch: got %d, expected %d", actual.Live.ActiveThresholdSeconds, expected.Live.ActiveThresholdSeconds)
	}
	if actual.Live.AnomalyThresholds.StallThresholdSeconds != expected.Live.AnomalyThresholds.StallThresholdSeconds {
		t.Errorf("StallThresholdSeconds mismatch: got %d, expected %d", actual.Live.AnomalyThresholds.StallThresholdSeconds, expected.Live.AnomalyThresholds.StallThresholdSeconds)
	}
}

func TestLoad_FlexiblePrefixParsing(t *testing.T) {
	t.Run("YAML structured prefix objects", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		yamlContent := `
project_name: PrefixApp
ticket:
  prefixes:
    - prefix: AL
      description: Core features
    - prefix: GEN
      description: General tasks
    - prefix: PO
      description: Product owner
workspaces:
  src_folders:
    - src
`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Ticket.Prefixes) != 3 {
			t.Fatalf("expected 3 prefixes, got %d", len(cfg.Ticket.Prefixes))
		}
		if cfg.Ticket.Prefixes[0].Prefix != "AL" || cfg.Ticket.Prefixes[0].Description != "Core features" {
			t.Errorf("unexpected AL prefix: %+v", cfg.Ticket.Prefixes[0])
		}
		if cfg.Ticket.Prefixes[1].Prefix != "GEN" || cfg.Ticket.Prefixes[1].Description != "General tasks" {
			t.Errorf("unexpected GEN prefix: %+v", cfg.Ticket.Prefixes[1])
		}
		if cfg.Ticket.Prefixes[2].Prefix != "PO" || cfg.Ticket.Prefixes[2].Description != "Product owner" {
			t.Errorf("unexpected PO prefix: %+v", cfg.Ticket.Prefixes[2])
		}
		names := cfg.Ticket.PrefixNames()
		if len(names) != 3 || names[0] != "AL" || names[1] != "GEN" || names[2] != "PO" {
			t.Errorf("unexpected PrefixNames: %v", names)
		}
	})

	t.Run("YAML plain string scalars", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		yamlContent := `
project_name: PrefixApp
ticket:
  prefixes:
    - AL
    - GEN
    - PO
workspaces:
  src_folders:
    - src
`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Ticket.Prefixes) != 3 {
			t.Fatalf("expected 3 prefixes, got %d", len(cfg.Ticket.Prefixes))
		}
		for i, name := range []string{"AL", "GEN", "PO"} {
			if cfg.Ticket.Prefixes[i].Prefix != name {
				t.Errorf("expected prefix[%d] %s, got %s", i, name, cfg.Ticket.Prefixes[i].Prefix)
			}
			if cfg.Ticket.Prefixes[i].Description != "" {
				t.Errorf("expected empty description for scalar, got %q", cfg.Ticket.Prefixes[i].Description)
			}
		}
		names := cfg.Ticket.PrefixNames()
		if len(names) != 3 || names[0] != "AL" || names[1] != "GEN" || names[2] != "PO" {
			t.Errorf("unexpected PrefixNames: %v", names)
		}
	})

	t.Run("YAML mixed prefix formats", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		yamlContent := `
project_name: PrefixApp
ticket:
  prefixes:
    - prefix: AL
      description: Core features
    - GEN
    - prefix: PO
      description: Product owner
workspaces:
  src_folders:
    - src
`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Ticket.Prefixes) != 3 {
			t.Fatalf("expected 3 prefixes, got %d", len(cfg.Ticket.Prefixes))
		}
		if cfg.Ticket.Prefixes[0].Prefix != "AL" || cfg.Ticket.Prefixes[0].Description != "Core features" {
			t.Errorf("unexpected AL prefix: %+v", cfg.Ticket.Prefixes[0])
		}
		if cfg.Ticket.Prefixes[1].Prefix != "GEN" || cfg.Ticket.Prefixes[1].Description != "" {
			t.Errorf("unexpected GEN prefix: %+v", cfg.Ticket.Prefixes[1])
		}
		if cfg.Ticket.Prefixes[2].Prefix != "PO" || cfg.Ticket.Prefixes[2].Description != "Product owner" {
			t.Errorf("unexpected PO prefix: %+v", cfg.Ticket.Prefixes[2])
		}
	})

	t.Run("JSON structured and string prefix compatibility", func(t *testing.T) {
		tempDir := t.TempDir()
		structuredPath := filepath.Join(tempDir, "structured.json")
		scalarPath := filepath.Join(tempDir, "scalar.json")

		structuredJSON := `{
			"project_name": "PrefixApp",
			"workspaces": {"src_folders": ["src"]},
			"ticket": {
				"prefixes": [
					{"prefix": "AL", "description": "Core features"},
					{"prefix": "PO", "description": "Product Owner"}
				]
			}
		}`
		scalarJSON := `{
			"project_name": "PrefixApp",
			"workspaces": {"src_folders": ["src"]},
			"ticket": {
				"prefixes": ["AL", "PO"]
			}
		}`

		if err := os.WriteFile(structuredPath, []byte(structuredJSON), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(scalarPath, []byte(scalarJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cfgStruct, err := config.LoadFromFile(structuredPath)
		if err != nil {
			t.Fatalf("unexpected error loading structured json: %v", err)
		}
		cfgScalar, err := config.LoadFromFile(scalarPath)
		if err != nil {
			t.Fatalf("unexpected error loading scalar json: %v", err)
		}

		namesStruct := cfgStruct.Ticket.PrefixNames()
		namesScalar := cfgScalar.Ticket.PrefixNames()

		if len(namesStruct) != 2 || namesStruct[0] != "AL" || namesStruct[1] != "PO" {
			t.Errorf("unexpected structured names: %v", namesStruct)
		}
		if len(namesScalar) != 2 || namesScalar[0] != "AL" || namesScalar[1] != "PO" {
			t.Errorf("unexpected scalar names: %v", namesScalar)
		}
		if cfgStruct.Ticket.Prefixes[0].Description != "Core features" {
			t.Errorf("expected description 'Core features', got %q", cfgStruct.Ticket.Prefixes[0].Description)
		}
		if cfgScalar.Ticket.Prefixes[0].Description != "" {
			t.Errorf("expected empty description for scalar, got %q", cfgScalar.Ticket.Prefixes[0].Description)
		}
	})

	t.Run("legacy flat JSON with PO prefix rejected", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		legacyJSON := `{"project_name": "LegacyPO", "workspaces": {"src_folders": ["src"]}, "ticket_prefix": "PO"}`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(legacyJSON), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error for legacy flat ticket_prefix without canonical section, got nil")
		}
		if !errors.Is(err, config.ErrInvalidConfig) {
			t.Errorf("expected ErrInvalidConfig, got %v", err)
		}
	})

	t.Run("priority after config.json deletion", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		yamlContent := `project_name: ModernOnly` + testValidSuffixYAML
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "ModernOnly" {
			t.Errorf("expected ModernOnly, got %s", cfg.ProjectName)
		}
	})

	t.Run("invalid prefix format handling in YAML and JSON", func(t *testing.T) {
		tempDir := t.TempDir()

		// YAML with integer prefix
		yamlInt := `ticket: { prefixes: [123] }`
		yamlIntPath := filepath.Join(tempDir, "invalid_int.yaml")
		if err := os.WriteFile(yamlIntPath, []byte(yamlInt), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := config.LoadFromFile(yamlIntPath)
		if err == nil {
			t.Fatal("expected error on integer prefix in YAML, got nil")
		}

		// YAML with bool prefix
		yamlBool := `ticket: { prefixes: [true] }`
		yamlBoolPath := filepath.Join(tempDir, "invalid_bool.yaml")
		if err := os.WriteFile(yamlBoolPath, []byte(yamlBool), 0644); err != nil {
			t.Fatal(err)
		}
		_, err = config.LoadFromFile(yamlBoolPath)
		if err == nil {
			t.Fatal("expected error on boolean prefix in YAML, got nil")
		}

		// YAML with nested array prefix
		yamlArray := `ticket: { prefixes: [[nested]] }`
		yamlArrayPath := filepath.Join(tempDir, "invalid_array.yaml")
		if err := os.WriteFile(yamlArrayPath, []byte(yamlArray), 0644); err != nil {
			t.Fatal(err)
		}
		_, err = config.LoadFromFile(yamlArrayPath)
		if err == nil {
			t.Fatal("expected error on nested array prefix in YAML, got nil")
		}

		// JSON with integer prefix
		jsonInt := `{"ticket": {"prefixes": [123]}}`
		jsonIntPath := filepath.Join(tempDir, "invalid_int.json")
		if err := os.WriteFile(jsonIntPath, []byte(jsonInt), 0644); err != nil {
			t.Fatal(err)
		}
		_, err = config.LoadFromFile(jsonIntPath)
		if err == nil {
			t.Fatal("expected error on integer prefix in JSON, got nil")
		}

		// JSON with boolean prefix
		jsonBool := `{"ticket": {"prefixes": [false]}}`
		jsonBoolPath := filepath.Join(tempDir, "invalid_bool.json")
		if err := os.WriteFile(jsonBoolPath, []byte(jsonBool), 0644); err != nil {
			t.Fatal(err)
		}
		_, err = config.LoadFromFile(jsonBoolPath)
		if err == nil {
			t.Fatal("expected error on boolean prefix in JSON, got nil")
		}
	})
}

func TestLoad_TicketTypePrefixParsing(t *testing.T) {
	t.Parallel()

	t.Run("YAML structured prefix with ticket_type", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		yamlContent := `
project_name: TicketTypeApp
ticket:
  prefixes:
    - prefix: AL
      description: Core features
    - prefix: GEN
      description: General tasks
      ticket_type: chore
    - prefix: PO
      description: Product owner
      ticket_type: docs
    - prefix: FIX
      ticket_type: " fix "
workspaces:
  src_folders:
    - src
`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Ticket.Prefixes) != 4 {
			t.Fatalf("expected 4 prefixes, got %d", len(cfg.Ticket.Prefixes))
		}
		if cfg.Ticket.Prefixes[0].Prefix != "AL" || cfg.Ticket.Prefixes[0].TicketType != "" {
			t.Errorf("expected AL with empty ticket_type, got %+v", cfg.Ticket.Prefixes[0])
		}
		if cfg.Ticket.Prefixes[1].Prefix != "GEN" || cfg.Ticket.Prefixes[1].TicketType != "chore" {
			t.Errorf("expected GEN with ticket_type chore, got %+v", cfg.Ticket.Prefixes[1])
		}
		if cfg.Ticket.Prefixes[2].Prefix != "PO" || cfg.Ticket.Prefixes[2].TicketType != "docs" {
			t.Errorf("expected PO with ticket_type docs, got %+v", cfg.Ticket.Prefixes[2])
		}
		if cfg.Ticket.Prefixes[3].Prefix != "FIX" || cfg.Ticket.Prefixes[3].TicketType != "fix" {
			t.Errorf("expected FIX with trimmed ticket_type fix, got %+v", cfg.Ticket.Prefixes[3])
		}
	})

	t.Run("JSON structured prefix with ticket_type", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		jsonPath := filepath.Join(tempDir, "config.json")

		jsonContent := `{
			"project_name": "TicketTypeApp",
			"workspaces": {"src_folders": ["src"]},
			"ticket": {
				"prefixes": [
					{"prefix": "AL", "description": "Core"},
					{"prefix": "GEN", "description": "Tasks", "ticket_type": "chore"},
					{"prefix": "PO", "ticket_type": "docs"},
					"SCALAR"
				]
			}
		}`
		if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.LoadFromFile(jsonPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Ticket.Prefixes) != 4 {
			t.Fatalf("expected 4 prefixes, got %d", len(cfg.Ticket.Prefixes))
		}
		if cfg.Ticket.Prefixes[0].Prefix != "AL" || cfg.Ticket.Prefixes[0].TicketType != "" {
			t.Errorf("expected AL with empty ticket_type, got %+v", cfg.Ticket.Prefixes[0])
		}
		if cfg.Ticket.Prefixes[1].Prefix != "GEN" || cfg.Ticket.Prefixes[1].TicketType != "chore" {
			t.Errorf("expected GEN with ticket_type chore, got %+v", cfg.Ticket.Prefixes[1])
		}
		if cfg.Ticket.Prefixes[2].Prefix != "PO" || cfg.Ticket.Prefixes[2].TicketType != "docs" {
			t.Errorf("expected PO with ticket_type docs, got %+v", cfg.Ticket.Prefixes[2])
		}
		if cfg.Ticket.Prefixes[3].Prefix != "SCALAR" || cfg.Ticket.Prefixes[3].TicketType != "" {
			t.Errorf("expected SCALAR with empty ticket_type, got %+v", cfg.Ticket.Prefixes[3])
		}
	})
}

func TestLoad_RoleOrderParsingAndFallback(t *testing.T) {
	t.Parallel()

	defaultOrder := domain.DefaultProjectConfig().RoleOrder

	t.Run("role_order parsed accurately from JSON", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		jsonContent := `{
			"project_name": "JSONRoleOrderProject",
			"workspaces": {"src_folders": ["src"]},
			"ticket": {"prefixes": [{"prefix": "AL"}]},
			"role_order": ["tester", "researcher", "developer", "coder"]
		}`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(jsonContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error loading config: %v", err)
		}

		expected := []string{"tester", "researcher", "developer", "coder"}
		if len(cfg.RoleOrder) != len(expected) {
			t.Fatalf("expected RoleOrder len %d, got %d: %v", len(expected), len(cfg.RoleOrder), cfg.RoleOrder)
		}
		for i, r := range expected {
			if cfg.RoleOrder[i] != r {
				t.Errorf("expected RoleOrder[%d] %q, got %q", i, r, cfg.RoleOrder[i])
			}
		}
	})

	t.Run("role_order parsed accurately from YAML", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		yamlContent := `
project_name: YAMLRoleOrderProject
ticket:
  prefixes:
    - prefix: AL
workspaces:
  src_folders:
    - src
role_order:
  - lead-qa
  - architect
  - researcher
`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error loading config: %v", err)
		}

		expected := []string{"lead-qa", "architect", "researcher"}
		if len(cfg.RoleOrder) != len(expected) {
			t.Fatalf("expected RoleOrder len %d, got %d: %v", len(expected), len(cfg.RoleOrder), cfg.RoleOrder)
		}
		for i, r := range expected {
			if cfg.RoleOrder[i] != r {
				t.Errorf("expected RoleOrder[%d] %q, got %q", i, r, cfg.RoleOrder[i])
			}
		}
	})

	t.Run("fallback to DefaultProjectConfig when role_order omitted from YAML", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		yamlContent := `project_name: OmittedRoleOrderYAML` + testValidSuffixYAML
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error loading config: %v", err)
		}

		if len(cfg.RoleOrder) != 0 || cfg.RoleOrder == nil {
			t.Fatalf("expected empty non-nil RoleOrder, got %v", cfg.RoleOrder)
		}
		if len(cfg.RoleOrder) != len(defaultOrder) {
			t.Fatalf("expected default RoleOrder len %d, got %d: %v", len(defaultOrder), len(cfg.RoleOrder), cfg.RoleOrder)
		}
	})

	t.Run("fallback to DefaultProjectConfig when role_order omitted from JSON", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		jsonContent := `{"project_name": "OmittedRoleOrderJSON"` + testValidSuffixJSON
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(jsonContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error loading config: %v", err)
		}

		if len(cfg.RoleOrder) != 0 || cfg.RoleOrder == nil {
			t.Fatalf("expected empty non-nil RoleOrder, got %v", cfg.RoleOrder)
		}
		if len(cfg.RoleOrder) != len(defaultOrder) {
			t.Fatalf("expected default RoleOrder len %d, got %d: %v", len(defaultOrder), len(cfg.RoleOrder), cfg.RoleOrder)
		}
	})

	t.Run("returns ErrConfigNotFound when config file completely absent", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()

		_, err := config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error loading config from empty dir, got nil")
		}
		if !errors.Is(err, config.ErrConfigNotFound) {
			t.Errorf("expected ErrConfigNotFound, got: %v", err)
		}
	})

	t.Run("role_order sanitization trims whitespace lowercases and filters empty in YAML", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		yamlContent := `
project_name: SanitizedYAML
ticket:
  prefixes:
    - prefix: AL
workspaces:
  src_folders:
    - src
role_order:
  - "  TESTER  "
  - "   "
  - ""
  - "  RESEARCHER\t"
  - "DeVeLoPeR"
`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []string{"tester", "researcher", "developer"}
		if len(cfg.RoleOrder) != len(expected) {
			t.Fatalf("expected len %d, got %d: %v", len(expected), len(cfg.RoleOrder), cfg.RoleOrder)
		}
		for i, r := range expected {
			if cfg.RoleOrder[i] != r {
				t.Errorf("at index %d: expected %q, got %q", i, r, cfg.RoleOrder[i])
			}
		}
	})

	t.Run("role_order sanitization trims whitespace lowercases and filters empty in JSON", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		jsonContent := `{
			"project_name": "SanitizedJSON",
			"workspaces": {"src_folders": ["src"]},
			"ticket": {"prefixes": [{"prefix": "AL"}]},
			"role_order": ["  LEAD-QA  ", "  \n  ", "", "ArChItEcT"]
		}`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(jsonContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []string{"lead-qa", "architect"}
		if len(cfg.RoleOrder) != len(expected) {
			t.Fatalf("expected len %d, got %d: %v", len(expected), len(cfg.RoleOrder), cfg.RoleOrder)
		}
		for i, r := range expected {
			if cfg.RoleOrder[i] != r {
				t.Errorf("at index %d: expected %q, got %q", i, r, cfg.RoleOrder[i])
			}
		}
	})

	t.Run("role_order containing only whitespace items collapses to empty non-nil slice", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		yamlContent := `
project_name: WhitespaceRoles
ticket:
  prefixes:
    - prefix: AL
workspaces:
  src_folders:
    - src
role_order:
  - "   "
  - "\t"
  - "\n"
`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.RoleOrder) != 0 || cfg.RoleOrder == nil {
			t.Fatalf("expected empty non-nil RoleOrder, got %v", cfg.RoleOrder)
		}
	})
}

func TestLoad_RootConfigurationParsing(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	yamlContent := `
project_name: RootProject
ticket:
  prefixes:
    - prefix: ROOT
      description: Root initiatives
      ticket_type: feat
    - prefix: TASK
      description: Task trackers
      ticket_type: chore
  case_insensitive: true
  custom_patterns:
    - ROOT-(\d+)
workspaces:
  dir: root_workspaces
  src_folders:
    - web
    - backend
database:
  path: /tmp/root.db
  query_timeout_ms: 6000
  busy_timeout_ms: 7000
  read_only: true
live:
  active_threshold_seconds: 250
  poll_interval_ms: 1500
  max_trace_buffer_messages: 600
  anomaly_thresholds:
    stall_threshold_seconds: 90
    error_rate_threshold: 0.20
    context_exhaustion_pct: 0.80
code_metrics:
  include:
    - "backend/**"
  exclude:
    - "vendor/**"
  test_patterns:
    - "*_test.go"
  prod_patterns:
    - "*.go"
role_mappings:
  architect: planner
  qa: tester
role_order:
  - orchestrator
  - planner
  - coder
`
	if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("unexpected error loading root config: %v", err)
	}

	if cfg.ProjectName != "RootProject" {
		t.Errorf("expected RootProject, got %s", cfg.ProjectName)
	}
	if len(cfg.Ticket.Prefixes) != 2 || cfg.Ticket.Prefixes[0].Prefix != "ROOT" || cfg.Ticket.Prefixes[1].Prefix != "TASK" {
		t.Errorf("unexpected ticket prefixes: %v", cfg.Ticket.Prefixes)
	}
	if cfg.Ticket.Prefixes[0].TicketType != "feat" || cfg.Ticket.Prefixes[1].TicketType != "chore" {
		t.Errorf("unexpected ticket types: %v", cfg.Ticket.Prefixes)
	}
	if !cfg.Ticket.CaseInsensitive {
		t.Errorf("expected CaseInsensitive to be true")
	}
	if len(cfg.Ticket.CustomPatterns) != 1 || cfg.Ticket.CustomPatterns[0] != `ROOT-(\d+)` {
		t.Errorf("unexpected custom patterns: %v", cfg.Ticket.CustomPatterns)
	}
	if cfg.Workspaces.Dir != "root_workspaces" {
		t.Errorf("expected root_workspaces, got %s", cfg.Workspaces.Dir)
	}
	if len(cfg.Workspaces.SrcFolders) != 2 || cfg.Workspaces.SrcFolders[0] != "web" || cfg.Workspaces.SrcFolders[1] != "backend" {
		t.Errorf("unexpected src_folders: %v", cfg.Workspaces.SrcFolders)
	}
	if cfg.Database.Path != "/tmp/root.db" {
		t.Errorf("unexpected Database.Path: %s", cfg.Database.Path)
	}
	if cfg.Database.QueryTimeoutMs != 6000 || cfg.Database.BusyTimeoutMs != 7000 || !cfg.Database.ReadOnly {
		t.Errorf("unexpected database config: %+v", cfg.Database)
	}
	if cfg.Live.ActiveThresholdSeconds != 250 || cfg.Live.PollIntervalMs != 1500 || cfg.Live.MaxTraceBufferMessages != 600 {
		t.Errorf("unexpected live config: %+v", cfg.Live)
	}
	if cfg.Live.AnomalyThresholds.StallThresholdSeconds != 90 || cfg.Live.AnomalyThresholds.ErrorRateThreshold != 0.20 || cfg.Live.AnomalyThresholds.ContextExhaustionPct != 0.80 {
		t.Errorf("unexpected anomaly thresholds: %+v", cfg.Live.AnomalyThresholds)
	}
	if len(cfg.CodeMetrics.Include) != 1 || cfg.CodeMetrics.Include[0] != "backend/**" {
		t.Errorf("unexpected code metrics include: %v", cfg.CodeMetrics.Include)
	}
	if cfg.RoleMappings.ResolveRole("architect") != domain.RolePlanner {
		t.Errorf("expected RolePlanner for architect, got %s", cfg.RoleMappings.ResolveRole("architect"))
	}
	if len(cfg.RoleOrder) != 3 || cfg.RoleOrder[0] != "orchestrator" || cfg.RoleOrder[1] != "planner" || cfg.RoleOrder[2] != "coder" {
		t.Errorf("unexpected RoleOrder: %v", cfg.RoleOrder)
	}
}

func TestLoad_RootFilePriorityAndPrecedence(t *testing.T) {
	t.Parallel()

	t.Run("root yaml takes precedence over root yml and root json", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()

		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte("project_name: FromRootYAML"+testValidSuffixYAML), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yml"), []byte("project_name: FromRootYML"+testValidSuffixYAML), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.json"), []byte(`{"project_name": "FromRootJSON"`+testValidSuffixJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "FromRootYAML" {
			t.Errorf("expected FromRootYAML, got %s", cfg.ProjectName)
		}
	})

	t.Run("root yml takes precedence over root json", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()

		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yml"), []byte("project_name: FromRootYML"+testValidSuffixYAML), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.json"), []byte(`{"project_name": "FromRootJSON"`+testValidSuffixJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "FromRootYML" {
			t.Errorf("expected FromRootYML, got %s", cfg.ProjectName)
		}
	})

	t.Run("root json selected when root yaml and yml absent", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()

		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.json"), []byte(`{"project_name": "FromRootJSON"`+testValidSuffixJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "FromRootJSON" {
			t.Errorf("expected FromRootJSON, got %s", cfg.ProjectName)
		}
	})

	t.Run("root yaml takes precedence over legacy subfolder yaml", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		legacyDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(legacyDir, 0755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte("project_name: FromRootYAML"+testValidSuffixYAML), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(legacyDir, "config.yaml"), []byte("project_name: FromLegacyYAML"+testValidSuffixYAML), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "FromRootYAML" {
			t.Errorf("expected FromRootYAML over legacy YAML, got %s", cfg.ProjectName)
		}
	})

	t.Run("root json takes precedence over legacy subfolder yaml", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		legacyDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(legacyDir, 0755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.json"), []byte(`{"project_name": "FromRootJSON"`+testValidSuffixJSON), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(legacyDir, "config.yaml"), []byte("project_name: FromLegacyYAML"+testValidSuffixYAML), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "FromRootJSON" {
			t.Errorf("expected FromRootJSON over legacy YAML, got %s", cfg.ProjectName)
		}
	})

	t.Run("legacy fallback when root config files absent", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		legacyDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(legacyDir, 0755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(legacyDir, "config.yaml"), []byte("project_name: LegacyFallbackYAML"+testValidSuffixYAML), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "LegacyFallbackYAML" {
			t.Errorf("expected LegacyFallbackYAML, got %s", cfg.ProjectName)
		}
	})
}

func TestFindConfigFile_RootUpwardTraversal(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	expectedPath := filepath.Join(tempDir, ".agenticlens-config.yaml")
	if err := os.WriteFile(expectedPath, []byte("project_name: RootTraversal"), 0644); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(tempDir, "sub1", "sub2", "sub3")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	found, err := config.FindConfigFile(subDir)
	if err != nil {
		t.Fatalf("unexpected error finding config file: %v", err)
	}
	if found != expectedPath {
		t.Errorf("expected %s, got %s", expectedPath, found)
	}
}

func TestLoad_StrictSixTierCandidatePriorityElimination(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	legacyDir := filepath.Join(tempDir, ".agenticlens")
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		t.Fatalf("failed creating legacy directory: %v", err)
	}

	c1 := filepath.Join(tempDir, ".agenticlens-config.yaml")
	c2 := filepath.Join(tempDir, ".agenticlens-config.yml")
	c3 := filepath.Join(tempDir, ".agenticlens-config.json")
	c4 := filepath.Join(legacyDir, "config.yaml")
	c5 := filepath.Join(legacyDir, "config.yml")
	c6 := filepath.Join(legacyDir, "config.json")

	if err := os.WriteFile(c1, []byte("project_name: Tier1_RootYAML"+testValidSuffixYAML), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c2, []byte("project_name: Tier2_RootYML"+testValidSuffixYAML), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c3, []byte(`{"project_name": "Tier3_RootJSON"`+testValidSuffixJSON), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c4, []byte("project_name: Tier4_LegacyYAML"+testValidSuffixYAML), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c5, []byte("project_name: Tier5_LegacyYML"+testValidSuffixYAML), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c6, []byte(`{"project_name": "Tier6_LegacyJSON"`+testValidSuffixJSON), 0644); err != nil {
		t.Fatal(err)
	}

	tiers := []struct {
		name         string
		expectedPath string
		expectedName string
		toDelete     string
	}{
		{
			name:         "Tier 1: Root YAML",
			expectedPath: c1,
			expectedName: "Tier1_RootYAML",
			toDelete:     c1,
		},
		{
			name:         "Tier 2: Root YML",
			expectedPath: c2,
			expectedName: "Tier2_RootYML",
			toDelete:     c2,
		},
		{
			name:         "Tier 3: Root JSON",
			expectedPath: c3,
			expectedName: "Tier3_RootJSON",
			toDelete:     c3,
		},
		{
			name:         "Tier 4: Legacy YAML",
			expectedPath: c4,
			expectedName: "Tier4_LegacyYAML",
			toDelete:     c4,
		},
		{
			name:         "Tier 5: Legacy YML",
			expectedPath: c5,
			expectedName: "Tier5_LegacyYML",
			toDelete:     c5,
		},
		{
			name:         "Tier 6: Legacy JSON",
			expectedPath: c6,
			expectedName: "Tier6_LegacyJSON",
			toDelete:     c6,
		},
	}

	for _, tc := range tiers {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			found, err := config.FindConfigFile(tempDir)
			if err != nil {
				t.Fatalf("unexpected error finding config: %v", err)
			}
			if found != tc.expectedPath {
				t.Errorf("expected path %s, got %s", tc.expectedPath, found)
			}

			cfg, err := config.Load(tempDir)
			if err != nil {
				t.Fatalf("unexpected error loading config: %v", err)
			}
			if cfg.ProjectName != tc.expectedName {
				t.Errorf("expected project name %s, got %s", tc.expectedName, cfg.ProjectName)
			}

			if err := os.Remove(tc.toDelete); err != nil {
				t.Fatalf("failed removing %s: %v", tc.toDelete, err)
			}
		})
	}

	t.Run("Tier 7: Default fallback after all candidates removed returns ErrConfigNotFound", func(t *testing.T) {
		found, err := config.FindConfigFile(tempDir)
		if err != nil {
			t.Fatalf("unexpected error finding config: %v", err)
		}
		if found != "" {
			t.Errorf("expected empty path, got %s", found)
		}

		_, err = config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error loading absent config, got nil")
		}
		if !errors.Is(err, config.ErrConfigNotFound) {
			t.Errorf("expected ErrConfigNotFound, got %v", err)
		}
	})
}

func TestLoad_HostileFilesystemEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("root yaml is a directory should be skipped in favor of root yml", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()

		dirCandidate := filepath.Join(tempDir, ".agenticlens-config.yaml")
		if err := os.MkdirAll(dirCandidate, 0755); err != nil {
			t.Fatal(err)
		}
		fileCandidate := filepath.Join(tempDir, ".agenticlens-config.yml")
		if err := os.WriteFile(fileCandidate, []byte("project_name: SkipDirYML"+testValidSuffixYAML), 0644); err != nil {
			t.Fatal(err)
		}

		found, err := config.FindConfigFile(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if found != fileCandidate {
			t.Errorf("expected %s, got %s", fileCandidate, found)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected load error: %v", err)
		}
		if cfg.ProjectName != "SkipDirYML" {
			t.Errorf("expected SkipDirYML, got %s", cfg.ProjectName)
		}
	})

	t.Run("legacy .agenticlens path is a file rather than directory", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()

		legacyFile := filepath.Join(tempDir, ".agenticlens")
		if err := os.WriteFile(legacyFile, []byte("regular file, not directory"), 0644); err != nil {
			t.Fatal(err)
		}

		jsonCandidate := filepath.Join(tempDir, ".agenticlens-config.json")
		if err := os.WriteFile(jsonCandidate, []byte(`{"project_name": "LegacyFileSafeJSON"`+testValidSuffixJSON), 0644); err != nil {
			t.Fatal(err)
		}

		found, err := config.FindConfigFile(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if found != jsonCandidate {
			t.Errorf("expected %s, got %s", jsonCandidate, found)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "LegacyFileSafeJSON" {
			t.Errorf("expected LegacyFileSafeJSON, got %s", cfg.ProjectName)
		}
	})

	t.Run("empty legacy directory in child does not shadow root config in parent", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()

		parentConfig := filepath.Join(tempDir, ".agenticlens-config.yaml")
		if err := os.WriteFile(parentConfig, []byte("project_name: ParentRootYAML"+testValidSuffixYAML), 0644); err != nil {
			t.Fatal(err)
		}

		childDir := filepath.Join(tempDir, "child_repo")
		emptyLegacyDir := filepath.Join(childDir, ".agenticlens")
		if err := os.MkdirAll(emptyLegacyDir, 0755); err != nil {
			t.Fatal(err)
		}

		found, err := config.FindConfigFile(childDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if found != parentConfig {
			t.Errorf("expected parent config %s, got %s", parentConfig, found)
		}

		cfg, err := config.Load(childDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "ParentRootYAML" {
			t.Errorf("expected ParentRootYAML, got %s", cfg.ProjectName)
		}
	})
}

func TestLoad_BoundaryAndCorruptedConfigs(t *testing.T) {
	t.Parallel()

	t.Run("empty root yaml config file returns validation error", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgPath := filepath.Join(tempDir, ".agenticlens-config.yaml")
		if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error on empty file, got nil")
		}
		if !errors.Is(err, config.ErrInvalidConfig) {
			t.Errorf("expected ErrInvalidConfig, got: %v", err)
		}
	})

	t.Run("whitespace-only root yaml config file returns validation error", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgPath := filepath.Join(tempDir, ".agenticlens-config.yaml")
		if err := os.WriteFile(cfgPath, []byte("   \n\n   \n"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error on whitespace file, got nil")
		}
		if !errors.Is(err, config.ErrInvalidConfig) {
			t.Errorf("expected ErrInvalidConfig, got: %v", err)
		}
	})

	t.Run("tab-character in root yaml config file returns descriptive error", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgPath := filepath.Join(tempDir, ".agenticlens-config.yaml")
		if err := os.WriteFile(cfgPath, []byte("\tinvalid_tab: true\n"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error on tab-character yaml file, got nil")
		}
		if !strings.Contains(err.Error(), "decoding yaml configuration") {
			t.Errorf("expected error containing 'decoding yaml configuration', got: %v", err)
		}
	})

	t.Run("corrupted root yaml syntax returns descriptive error without panicking", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgPath := filepath.Join(tempDir, ".agenticlens-config.yaml")
		if err := os.WriteFile(cfgPath, []byte("project_name: [unclosed_mapping:\n"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error loading corrupted yaml, got nil")
		}
		if !strings.Contains(err.Error(), "decoding yaml configuration") {
			t.Errorf("expected error containing 'decoding yaml configuration', got: %v", err)
		}
	})

	t.Run("corrupted root json syntax returns descriptive error without panicking", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgPath := filepath.Join(tempDir, ".agenticlens-config.json")
		if err := os.WriteFile(cfgPath, []byte(`{"project_name": "bad_json", `), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error loading corrupted json, got nil")
		}
		if !strings.Contains(err.Error(), "decoding json configuration") {
			t.Errorf("expected error containing 'decoding json configuration', got: %v", err)
		}
	})

	t.Run("malformed ticket prefix scalar tag in root config returns descriptive error", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgPath := filepath.Join(tempDir, ".agenticlens-config.yaml")
		content := `
ticket:
  prefixes:
    - !!int 12345
`
		if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error decoding non-string scalar prefix, got nil")
		}
		if !strings.Contains(err.Error(), "invalid ticket prefix scalar type") {
			t.Errorf("expected error containing 'invalid ticket prefix scalar type', got: %v", err)
		}
	})

	t.Run("legacy subfolder .agenticlens/config.yml parses correctly", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		legacyDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(legacyDir, 0755); err != nil {
			t.Fatal(err)
		}
		cfgPath := filepath.Join(legacyDir, "config.yml")
		if err := os.WriteFile(cfgPath, []byte("project_name: LegacyYMLOnly"+testValidSuffixYAML), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "LegacyYMLOnly" {
			t.Errorf("expected LegacyYMLOnly, got %s", cfg.ProjectName)
		}
	})

	t.Run("legacy subfolder .agenticlens/config.json with legacy root fields is rejected", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		legacyDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(legacyDir, 0755); err != nil {
			t.Fatal(err)
		}
		cfgPath := filepath.Join(legacyDir, "config.json")
		content := `{
			"project_name": "LegacyProject",
			"ticket_prefix": "LEGACYPREFIX",
			"workspaces_dir": "legacy_workspaces_dir",
			"src_folders": ["src_legacy_1", "src_legacy_2"]
		}`
		if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := config.Load(tempDir)
		if err == nil {
			t.Fatalf("expected error loading legacy json without canonical sections, got nil")
		}
		if !errors.Is(err, config.ErrInvalidConfig) {
			t.Errorf("expected ErrInvalidConfig, got %v", err)
		}
	})
}

func TestLoad_CustomExtensionAndEmptyPathEdgeCases(t *testing.T) {
	t.Run("FindConfigFile with empty string uses working directory", func(t *testing.T) {
		tempDir := t.TempDir()
		expectedPath := filepath.Join(tempDir, ".agenticlens-config.yaml")
		if err := os.WriteFile(expectedPath, []byte("project_name: CwdTest"+testValidSuffixYAML), 0644); err != nil {
			t.Fatal(err)
		}
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		found, err := config.FindConfigFile("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if found != expectedPath {
			t.Errorf("expected %s, got %s", expectedPath, found)
		}

		cfg, err := config.Load("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "CwdTest" {
			t.Errorf("expected CwdTest, got %s", cfg.ProjectName)
		}
	})

	t.Run("LoadFromFile with non-standard extension falls back to YAML then JSON", func(t *testing.T) {
		tempDir := t.TempDir()
		yamlConf := filepath.Join(tempDir, "config.conf")
		if err := os.WriteFile(yamlConf, []byte("project_name: NonStandardYAML"+testValidSuffixYAML), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.LoadFromFile(yamlConf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ProjectName != "NonStandardYAML" {
			t.Errorf("expected NonStandardYAML, got %s", cfg.ProjectName)
		}

		jsonConf := filepath.Join(tempDir, "config.ext")
		if err := os.WriteFile(jsonConf, []byte(`{"project_name": "NonStandardJSON"`+testValidSuffixJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cfg2, err := config.LoadFromFile(jsonConf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg2.ProjectName != "NonStandardJSON" {
			t.Errorf("expected NonStandardJSON, got %s", cfg2.ProjectName)
		}

		badConf := filepath.Join(tempDir, "config.bad")
		if err := os.WriteFile(badConf, []byte(`[invalid: {`), 0644); err != nil {
			t.Fatal(err)
		}

		_, err = config.LoadFromFile(badConf)
		if err == nil {
			t.Fatalf("expected error on invalid file, got nil")
		}
	})
}

func TestLoad_CodeMetricsBaseBranch(t *testing.T) {
	t.Run("unmarshals base_branch from yaml", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgFile := filepath.Join(tempDir, ".agenticlens-config.yaml")
		content := `project_name: "TestProject"` + testValidSuffixYAML + `
code_metrics:
  base_branch: "release/v1.0"
`
		if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.LoadFromFile(cfgFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.CodeMetrics.BaseBranch != "release/v1.0" {
			t.Errorf("expected base_branch 'release/v1.0', got %q", cfg.CodeMetrics.BaseBranch)
		}
	})

	t.Run("unmarshals base_branch from json", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgFile := filepath.Join(tempDir, ".agenticlens-config.json")
		content := `{"project_name": "TestProject", "code_metrics": {"base_branch": "master"}` + testValidSuffixJSON
		if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.LoadFromFile(cfgFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.CodeMetrics.BaseBranch != "master" {
			t.Errorf("expected base_branch 'master', got %q", cfg.CodeMetrics.BaseBranch)
		}
	})

	t.Run("defaults base_branch to empty string when omitted", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgFile := filepath.Join(tempDir, ".agenticlens-config.yaml")
		content := `project_name: "TestProject"` + testValidSuffixYAML
		if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.LoadFromFile(cfgFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.CodeMetrics.BaseBranch != "" {
			t.Errorf("expected default base_branch '', got %q", cfg.CodeMetrics.BaseBranch)
		}
	})
}

func TestLoadFromFile(t *testing.T) {
	t.Run("correctly loads base_branch from repo .agenticlens-config.yaml", func(t *testing.T) {
		repoConfigPath := filepath.Join("..", "..", "..", "..", ".agenticlens-config.yaml")
		cfg, err := config.LoadFromFile(repoConfigPath)
		if err != nil {
			t.Fatalf("failed to load root .agenticlens-config.yaml: %v", err)
		}
		if cfg.CodeMetrics.BaseBranch != "origin/main" {
			t.Errorf("expected base_branch %q, got %q", "origin/main", cfg.CodeMetrics.BaseBranch)
		}
	})

	t.Run("correctly loads custom base_branch from yaml file", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgFile := filepath.Join(tempDir, ".agenticlens-config.yaml")
		content := `project_name: "TestProject"` + testValidSuffixYAML + `
code_metrics:
  base_branch: "origin/release-2.0"
`
		if err := os.WriteFile(cfgFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		cfg, err := config.LoadFromFile(cfgFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.CodeMetrics.BaseBranch != "origin/release-2.0" {
			t.Errorf("expected base_branch %q, got %q", "origin/release-2.0", cfg.CodeMetrics.BaseBranch)
		}
	})
}
