package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

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

		yamlContent := `project_name: FromYAML`
		jsonContent := `{"project_name": "FromJSON"}`

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

		yamlContent := `project_name: FromYAML`
		ymlContent := `project_name: FromYML`

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

		ymlContent := `project_name: FromYML`
		jsonContent := `{"project_name": "FromJSON"}`

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

		jsonContent := `{"project_name": "FromJSON"}`
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

func TestLoad_LegacyJSONCompatibility(t *testing.T) {
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

	cfg, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("unexpected error loading legacy json: %v", err)
	}

	if cfg.ProjectName != "LegacyProject" {
		t.Errorf("expected LegacyProject, got %s", cfg.ProjectName)
	}
	if len(cfg.Ticket.Prefixes) != 1 || cfg.Ticket.Prefixes[0].Prefix != "LEG" {
		t.Errorf("expected prefixes [LEG], got %v", cfg.Ticket.Prefixes)
	}
	if cfg.Ticket.Prefixes[0].Description != "" {
		t.Errorf("expected empty description for legacy prefix, got %q", cfg.Ticket.Prefixes[0].Description)
	}
	if cfg.Workspaces.Dir != "custom_workspaces_dir" {
		t.Errorf("expected custom_workspaces_dir, got %s", cfg.Workspaces.Dir)
	}
	if len(cfg.Workspaces.SrcFolders) != 2 || cfg.Workspaces.SrcFolders[0] != "api" || cfg.Workspaces.SrcFolders[1] != "ui" {
		t.Errorf("expected src_folders [api, ui], got %v", cfg.Workspaces.SrcFolders)
	}

	// Sane defaults preserved for other fields
	if cfg.Database.QueryTimeoutMs != 5000 {
		t.Errorf("expected default database timeout 5000, got %d", cfg.Database.QueryTimeoutMs)
	}
	if cfg.Live.ActiveThresholdSeconds != 300 {
		t.Errorf("expected default live active threshold 300, got %d", cfg.Live.ActiveThresholdSeconds)
	}
}

func TestLoad_MissingConfigurationFileFallback(t *testing.T) {
	tempDir := t.TempDir()

	cfg, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("expected nil error on missing config, got %v", err)
	}

	defaultCfg := domain.DefaultProjectConfig()
	if cfg.ProjectName != defaultCfg.ProjectName {
		t.Errorf("expected default project name %s, got %s", defaultCfg.ProjectName, cfg.ProjectName)
	}
	if cfg.Workspaces.Dir != defaultCfg.Workspaces.Dir {
		t.Errorf("expected default workspaces dir %s, got %s", defaultCfg.Workspaces.Dir, cfg.Workspaces.Dir)
	}
	if len(cfg.Ticket.Prefixes) != len(defaultCfg.Ticket.Prefixes) {
		t.Errorf("expected default prefixes %v, got %v", defaultCfg.Ticket.Prefixes, cfg.Ticket.Prefixes)
	}
	for i, exp := range defaultCfg.Ticket.Prefixes {
		if cfg.Ticket.Prefixes[i].TicketType != exp.TicketType {
			t.Errorf("expected default prefix[%d] %q TicketType %q, got %q", i, exp.Prefix, exp.TicketType, cfg.Ticket.Prefixes[i].TicketType)
		}
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

	// Unspecified sections retain defaults
	if cfg.Workspaces.Dir != "workspaces" {
		t.Errorf("expected default workspaces dir, got %s", cfg.Workspaces.Dir)
	}
	if len(cfg.Ticket.Prefixes) != 3 || cfg.Ticket.Prefixes[0].Prefix != "AL" || cfg.Ticket.Prefixes[1].Prefix != "GEN" || cfg.Ticket.Prefixes[2].Prefix != "PO" {
		t.Errorf("expected default ticket prefixes, got %v", cfg.Ticket.Prefixes)
	}
	if cfg.Database.QueryTimeoutMs != 5000 {
		t.Errorf("expected default query timeout 5000, got %d", cfg.Database.QueryTimeoutMs)
	}
	if cfg.Live.ActiveThresholdSeconds != 300 {
		t.Errorf("expected default active threshold 300, got %d", cfg.Live.ActiveThresholdSeconds)
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

		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("project_name: FromYAML"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yml"), []byte("project_name: FromYML"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"project_name": "FromJSON"}`), 0644); err != nil {
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

		if err := os.WriteFile(filepath.Join(cfgDir, "config.yml"), []byte("project_name: FromYML"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"project_name": "FromJSON"}`), 0644); err != nil {
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
	defaults := domain.DefaultProjectConfig()

	t.Run("zero byte empty file preserves full defaults", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assertConfigMatchesDefaults(t, cfg, defaults)
	})

	t.Run("empty object preserves full defaults", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("{}\n"), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assertConfigMatchesDefaults(t, cfg, defaults)
	})

	t.Run("nested single field overrides retain other defaults", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		partialContent := `
ticket:
  case_insensitive: false
workspaces:
  dir: custom_workspaces_only
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

		// Unspecified fields in same sections must retain defaults
		if len(cfg.Ticket.Prefixes) != 3 || cfg.Ticket.Prefixes[0].Prefix != "AL" || cfg.Ticket.Prefixes[1].Prefix != "GEN" || cfg.Ticket.Prefixes[2].Prefix != "PO" {
			t.Errorf("unexpected ticket prefixes: %v", cfg.Ticket.Prefixes)
		}
		if len(cfg.Workspaces.SrcFolders) != 2 || cfg.Workspaces.SrcFolders[0] != "web" || cfg.Workspaces.SrcFolders[1] != "backend" {
			t.Errorf("unexpected src_folders: %v", cfg.Workspaces.SrcFolders)
		}
		if cfg.Database.Path != defaults.Database.Path || cfg.Database.QueryTimeoutMs != 5000 || cfg.Database.BusyTimeoutMs != 5000 {
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
	t.Run("legacy keys translate and empty values do not clobber defaults", func(t *testing.T) {
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

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Ticket.Prefixes) != 3 || cfg.Ticket.Prefixes[0].Prefix != "AL" {
			t.Errorf("expected default prefixes AL, GEN, PO when legacy prefix is empty whitespace, got %v", cfg.Ticket.Prefixes)
		}
		if cfg.Workspaces.Dir != "workspaces" {
			t.Errorf("expected default workspaces dir when legacy workspaces_dir is empty, got %s", cfg.Workspaces.Dir)
		}
		if len(cfg.Workspaces.SrcFolders) != 2 || cfg.Workspaces.SrcFolders[0] != "web" {
			t.Errorf("expected default src_folders when legacy src_folders is empty, got %v", cfg.Workspaces.SrcFolders)
		}
	})

	t.Run("modern nested keys supersede legacy root keys", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		mixedJSON := `{
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
	t.Run("nonexistent directory returns defaults and nil error", func(t *testing.T) {
		nonexistentDir := filepath.Join(os.TempDir(), "agenticlens_nonexistent_test_dir_12345")
		_ = os.RemoveAll(nonexistentDir)

		cfg, err := config.Load(nonexistentDir)
		if err != nil {
			t.Fatalf("expected nil error on nonexistent directory, got %v", err)
		}

		defaults := domain.DefaultProjectConfig()
		assertConfigMatchesDefaults(t, cfg, defaults)
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
ticket:
  prefixes:
    - prefix: AL
      description: Core features
    - prefix: GEN
      description: General tasks
    - prefix: PO
      description: Product owner
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
ticket:
  prefixes:
    - AL
    - GEN
    - PO
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
ticket:
  prefixes:
    - prefix: AL
      description: Core features
    - GEN
    - prefix: PO
      description: Product owner
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
			"ticket": {
				"prefixes": [
					{"prefix": "AL", "description": "Core features"},
					{"prefix": "PO", "description": "Product Owner"}
				]
			}
		}`
		scalarJSON := `{
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

	t.Run("legacy flat JSON with PO prefix", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		legacyJSON := `{"ticket_prefix": "PO"}`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(legacyJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Ticket.Prefixes) != 1 || cfg.Ticket.Prefixes[0].Prefix != "PO" {
			t.Errorf("expected [PO] prefix, got %+v", cfg.Ticket.Prefixes)
		}
		if cfg.Ticket.Prefixes[0].Description != "" {
			t.Errorf("expected empty description for legacy prefix, got %q", cfg.Ticket.Prefixes[0].Description)
		}
	})

	t.Run("priority after config.json deletion", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		yamlContent := `project_name: ModernOnly`
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

		yamlContent := `project_name: OmittedRoleOrderYAML`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error loading config: %v", err)
		}

		if len(cfg.RoleOrder) != len(defaultOrder) {
			t.Fatalf("expected default RoleOrder len %d, got %d: %v", len(defaultOrder), len(cfg.RoleOrder), cfg.RoleOrder)
		}
		for i, r := range defaultOrder {
			if cfg.RoleOrder[i] != r {
				t.Errorf("expected default RoleOrder[%d] %q, got %q", i, r, cfg.RoleOrder[i])
			}
		}
	})

	t.Run("fallback to DefaultProjectConfig when role_order omitted from JSON", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}

		jsonContent := `{"project_name": "OmittedRoleOrderJSON"}`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(jsonContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error loading config: %v", err)
		}

		if len(cfg.RoleOrder) != len(defaultOrder) {
			t.Fatalf("expected default RoleOrder len %d, got %d: %v", len(defaultOrder), len(cfg.RoleOrder), cfg.RoleOrder)
		}
		for i, r := range defaultOrder {
			if cfg.RoleOrder[i] != r {
				t.Errorf("expected default RoleOrder[%d] %q, got %q", i, r, cfg.RoleOrder[i])
			}
		}
	})

	t.Run("fallback to DefaultProjectConfig when config file completely absent", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()

		cfg, err := config.Load(tempDir)
		if err != nil {
			t.Fatalf("unexpected error loading config from empty dir: %v", err)
		}

		if len(cfg.RoleOrder) != len(defaultOrder) {
			t.Fatalf("expected default RoleOrder len %d, got %d: %v", len(defaultOrder), len(cfg.RoleOrder), cfg.RoleOrder)
		}
		for i, r := range defaultOrder {
			if cfg.RoleOrder[i] != r {
				t.Errorf("expected default RoleOrder[%d] %q, got %q", i, r, cfg.RoleOrder[i])
			}
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
}
