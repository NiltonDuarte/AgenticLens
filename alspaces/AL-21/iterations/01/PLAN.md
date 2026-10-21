# Implementation Plan: AL-21 Migrate Configuration to YAML and Establish Agent-Backend Bridge Definitions

## 1. Goal and Iteration Scope

The goal of iteration 01 is to deliver ticket AL-21: migrating AgenticLens configuration from the legacy flat JSON format to a structured, human-readable YAML format with backward-compatible JSON fallback, and establishing shared agent-to-backend bridge definitions.

This iteration encompasses adding gopkg.in/yaml.v3 to the Go backend dependencies, defining pure domain configuration models and ticket normalization logic in backend/internal/domain/config.go, implementing the centralized file discovery and parsing adapter in backend/internal/adapters/config/loader.go, wiring existing workspace resolution and git classification adapters to use the centralized loader, authoring the fully annotated .agenticlens/config.yaml in the project root, and implementing comprehensive unit and boundary tests.

## 2. Grounded Reference Architecture & File Paths

The implementation maps directly to the following repository locations:

- Dependency Manifest:
  - backend/go.mod and backend/go.sum: Add gopkg.in/yaml.v3 for YAML parsing under pure Go execution.
- Domain Functional Core:
  - backend/internal/domain/config.go: Pure configuration structs, default values, role mappings, and pure ticket ID normalization functions.
  - backend/internal/domain/config_test.go: Pure unit tests for ticket normalization and role resolution without mocks or I/O.
- Imperative Configuration Adapter:
  - backend/internal/adapters/config/loader.go: Centralized configuration discovery, YAML/JSON parsing, legacy JSON key migration, and domain defaults merging.
  - backend/internal/adapters/config/loader_test.go: Table-driven boundary tests for file hierarchy, YAML parsing, JSON fallback, legacy format translation, and corrupted file handling.
- Integrated Adapters:
  - backend/internal/adapters/repository.go: Refactor ResolveWorkspacesDir to query the centralized loader instead of running duplicate JSON reading loops.
  - backend/internal/adapters/git/classifier.go: Refactor LoadConfig to delegate to the centralized configuration loader.
- Repository Configuration Artifacts:
  - .agenticlens/config.yaml: New primary configuration manifest containing extensive explanatory comments, ticket pattern rules, database pragmas, live observability thresholds, and role mappings.
- Automation & Quality Gates:
  - Makefile: Verify that vet, test-backend, and check targets validate new packages.

## 3. Phased Step-by-Step Implementation Tasks for Coder

### Phase 1: Dependency Registration and Go Module Configuration
- Task 1.1: In backend directory, register gopkg.in/yaml.v3 using go get gopkg.in/yaml.v3. Verify that backend/go.mod and backend/go.sum update cleanly and that go vet ./... compiles without errors under pure Go settings.

### Phase 2: Domain Functional Core Scaffolding (backend/internal/domain/config.go)
- Task 2.1: Define pure configuration domain models in backend/internal/domain/config.go:
  - ProjectConfig struct containing ProjectName, Ticket, Workspaces, Database, Live, CodeMetrics, and RoleMappings with both json and yaml tags.
  - TicketConfig struct containing Prefixes as string slice, CaseInsensitive as boolean, and CustomPatterns as string slice.
  - WorkspacesConfig struct containing Dir as string and SrcFolders as string slice.
  - DatabaseConfig struct containing Path as string, QueryTimeoutMs as integer, BusyTimeoutMs as integer, and ReadOnly as boolean.
  - AnomalyThresholdsConfig struct containing StallThresholdSeconds as integer, ErrorRateThreshold as float64, and ContextExhaustionPct as float64.
  - LiveConfig struct containing ActiveThresholdSeconds as integer, PollIntervalMs as integer, MaxTraceBufferMessages as integer, and AnomalyThresholds as AnomalyThresholdsConfig.
  - CodeMetricsConfig struct containing Include, Exclude, TestPatterns, and ProdPatterns as string slices.
  - RoleMappings type defined as map of string to string.
- Task 2.2: Implement pure helper method RoleMappings.ResolveRole:
  - Clean and lowercase input agent string.
  - Check explicit lookup in map; if present, cast to domain.AgentRole.
  - If absent, match known agent defaults: orchestrator, build, and root map to RoleOrchestrator; planner maps to RolePlanner; coder maps to RoleCoder; tester maps to RoleTester; reviewer maps to RoleReviewer.
  - Fall back to casting input string to AgentRole.
- Task 2.3: Implement pure constructor DefaultProjectConfig:
  - Returns ProjectConfig populated with canonical operational defaults specified in CONTRACT.md: project name AgenticLens, prefixes AL and GEN, case insensitive true, workspaces directory workspaces, source folders web and backend, database path ~/.local/share/opencode/opencode.db, timeouts 5000ms, read only true, active threshold 300s, poll interval 2000ms, buffer 500 messages, anomaly thresholds (120s stall, 0.25 error rate, 0.85 saturation), empty code metrics slices, and standard role map.
- Task 2.4: Implement pure function NormalizeTicketID:
  - Accepts raw input string and TicketConfig.
  - Trims leading and trailing whitespace. If empty, returns empty string and false.
  - If TicketConfig Prefixes slice is empty, defaults to AL and GEN.
  - For each configured prefix, builds case-insensitive or case-sensitive regex pattern depending on CaseInsensitive flag.
  - Handles standalone ticket IDs (e.g. AL-21, al-21, gen-4) and tickets embedded in branch or commit identifiers (e.g. feat/al-21-yaml-config, bugfix/GEN-42-fix).
  - When a prefix match is identified, extracts the numeric portion and formats the canonical string as uppercase prefix followed by hyphen and numeric digits. Returns canonical string and true.
  - If prefixes do not match, tests any configured CustomPatterns regexes. On match, returns canonical identifier and true.
  - Returns empty string and false if no match is found.

### Phase 3: Centralized Imperative Configuration Loader (backend/internal/adapters/config/loader.go)
- Task 3.1: Implement FindConfigFile helper:
  - Accepts starting directory path. If empty, resolves from current working directory.
  - Traverses directory path upwards toward root filesystem boundary.
  - At each directory level, checks candidate paths in strict priority order: .agenticlens/config.yaml, .agenticlens/config.yml, .agenticlens/config.json.
  - If an existing file is found, returns its absolute path and nil error.
  - If root directory is reached without finding any file, returns empty string and nil error.
- Task 3.2: Implement legacy JSON migration and partial merging:
  - Define an internal unmarshaling DTO capable of reading both modern nested configuration keys and legacy flat JSON keys (ticket_prefix at root, workspaces_dir at root, src_folders at root).
  - Define merge helper that initializes with DefaultProjectConfig() and overlays non-zero values parsed from the configuration file.
  - If legacy root ticket_prefix is supplied and prefixes slice is empty, populates prefixes with a slice containing the legacy prefix.
  - If legacy root workspaces_dir is supplied and workspaces.dir is empty, populates workspaces.dir.
  - If legacy root src_folders is supplied and workspaces.src_folders is empty, populates workspaces.src_folders.
- Task 3.3: Implement LoadFromFile and Load functions:
  - LoadFromFile takes an explicit file path, reads file contents, inspects extension (.yaml/.yml vs .json), invokes appropriate unmarshaler (yaml.Unmarshal or json.Unmarshal), merges values on top of defaults, and returns domain.ProjectConfig.
  - Load takes a repoRoot directory string, calls FindConfigFile, and if a path is located, invokes LoadFromFile. If no file is found, returns DefaultProjectConfig() with nil error.

### Phase 4: Adapter Integration Refactoring
- Task 4.1: Refactor backend/internal/adapters/repository.go:
  - Remove duplicate private configDTO struct and custom file search loops in ResolveWorkspacesDir.
  - Update ResolveWorkspacesDir to invoke config.Load(""), inspect cfg.Workspaces.Dir, and resolve candidate path.
  - Retain container volume mount /app/workspaces check and current working directory fallbacks for unconfigured environments.
- Task 4.2: Refactor backend/internal/adapters/git/classifier.go:
  - Remove duplicate rootConfig struct and private JSON parsing loop.
  - Update LoadConfig(repoRoot string) (*ClassificationConfig, error) to invoke config.Load(repoRoot).
  - Map cfg.CodeMetrics fields into ClassificationConfig struct and return pointer.
  - Retain existing ClassificationConfig type and ClassifyFile signatures to ensure complete zero-breakage backward compatibility.

### Phase 5: Repository Configuration Artifact Migration
- Task 5.1: Create .agenticlens/config.yaml in repository root:
  - Populate with structured sections: project_name, ticket, workspaces, database, live, code_metrics, and role_mappings.
  - Add comprehensive YAML comments explaining each parameter, accepted values, and agent bridge implications.
  - Mirror existing settings from .agenticlens/config.json (AL prefix, workspaces directory, web and backend folders) while establishing explicit defaults for database and live observability.
- Task 5.2: Ensure backward compatibility validation:
  - Verify that the system operates correctly whether .agenticlens/config.yaml is present alone, or alongside legacy .agenticlens/config.json.

## 4. Adversarial Testing Strategy for Tester

### Pure Domain Unit Tests (backend/internal/domain/config_test.go)
- Test Case 1: Ticket Normalization - Exact Matches
  - Test inputs matching canonical format: AL-21 returns AL-21 and true. GEN-100 returns GEN-100 and true.
- Test Case 2: Ticket Normalization - Case Insensitivity
  - Test lowercase and mixed case inputs: al-21, Al-21, aL-21 all normalize to AL-21 with true result.
  - Test lowercase gen-5 returns GEN-5 and true.
- Test Case 3: Ticket Normalization - Git Branch and Token Patterns
  - Test branch names: feat/al-21-yaml-config extracts AL-21 and true.
  - Test bugfix/GEN-33-db-lock extracts GEN-33 and true.
- Test Case 4: Ticket Normalization - Case Sensitive Mode
  - When CaseInsensitive is set to false, al-21 fails to match prefix AL and returns empty string and false.
- Test Case 5: Ticket Normalization - Negative and Invalid Cases
  - Test non-matching strings: PROJ-123, random-text, ticket-, empty string, whitespace only. All return empty string and false.
- Test Case 6: Role Mapping Resolution
  - Test default mappings: planner returns RolePlanner, coder returns RoleCoder, build and orchestrator return RoleOrchestrator.
  - Test custom configured mappings: custom-agent mapped to tester correctly resolves to RoleTester.
  - Test unknown agent names: returns AgentRole with input string without panic.
  - Test empty agent string: safely defaults to RoleOrchestrator.

### Imperative Loader Boundary Tests (backend/internal/adapters/config/loader_test.go)
- Test Case 7: YAML Configuration Parsing
  - Create temporary directory with .agenticlens/config.yaml containing full configuration. Call Load and assert that all fields across all sections match expected values.
- Test Case 8: File Priority Resolution
  - In a temporary directory containing both config.yaml and config.json, assert that config.yaml takes precedence and config.json is ignored.
  - In a temporary directory containing only config.json, assert that config.json is loaded and properly deserialized.
- Test Case 9: Legacy JSON Compatibility
  - Create legacy config.json with root keys ticket_prefix, workspaces_dir, and src_folders. Call Load and assert that cfg.Ticket.Prefixes contains the prefix and cfg.Workspaces fields are populated correctly.
- Test Case 10: Missing Configuration File Fallback
  - Call Load on a directory tree with zero config files. Assert that it returns DefaultProjectConfig() without returning an error.
- Test Case 11: Partial Configuration Merging
  - Provide a YAML file specifying only project_name and database.path. Assert that unspecified sections (ticket, live, workspaces) are populated with pure domain defaults.
- Test Case 12: Malformed File Error Handling
  - Provide a syntactically invalid YAML file. Assert that Load returns an explicit descriptive error wrapping the YAML parsing failure.

### Adapter Integration Tests
- Test Case 13: Repository Workspace Resolution Regression
  - Execute existing and updated tests in backend/internal/adapters/repository_test.go to verify that ResolveWorkspacesDir respects configured directories in both YAML and JSON setups.
- Test Case 14: Git Classifier Configuration Regression
  - Execute backend/internal/adapters/git/analyzer_test.go to confirm that file classification correctly applies include, exclude, test_patterns, and prod_patterns loaded from the new configuration adapter.

### Static Analysis and Invariant Validation
- Verify zero any and zero interface{} in backend/internal/domain/config.go and backend/internal/adapters/config/loader.go using grep or static inspection.
- Verify zero emojis across all code, comments, test names, and config files.
- Run make check and make test-backend to guarantee clean compilation and zero race conditions.

## 5. Review Criteria for Reviewer

### Axis 1: Specification and Contract Alignment
- Does backend/internal/domain/config.go contain all requested structs (ProjectConfig, TicketConfig, WorkspacesConfig, DatabaseConfig, LiveConfig, AnomalyThresholdsConfig, CodeMetricsConfig, RoleMappings)?
- Does NormalizeTicketID correctly support case-insensitive matching for configured prefixes, converting lowercase inputs to canonical uppercase format?
- Does the configuration loader discover .agenticlens/config.yaml, fall back to .agenticlens/config.yml, fall back to .agenticlens/config.json, and default gracefully when missing?
- Are existing call-sites in repository.go and classifier.go cleanly migrated to the unified configuration loader?
- Is .agenticlens/config.yaml authored in the repository root with comprehensive, informative comments?

### Axis 2: Engineering Standards and Technical Invariants
- FCIS Discipline: Is backend/internal/domain/config.go 100% pure functional core with zero filesystem, disk I/O, or network imports? Does file traversal and YAML decoding remain strictly contained in backend/internal/adapters/config/?
- Strong Typing: Is there zero usage of any or empty interface{} in the new code? Are concrete types and typed enums employed consistently?
- Value Semantics: Are configuration models immutable value structs with safe defaults?
- Error Handling: Are errors wrapped with descriptive context using %w rather than discarded?
- Emoji Prohibition: Are all authored files, docstrings, comments, log lines, and YAML keys 100% free of emojis?
- Backward Compatibility: Does the test suite prove that legacy config.json setups continue to function without breakage?
- Quality Gates: Do make check, go vet ./..., and go test -race ./... pass with zero warnings or failures?
