# Implementation Plan: FIX-2 Iteration 03 - Core Domain Defaults Relocation, Complete Default Role Elimination, and Strict Empty-Role Discipline

## 1. Executive Summary and User Decisions

During the decoupled test workspace refactoring under ticket FIX-2, configuration loading logic was updated to prevent unconfigured fallback states. As part of this refactoring, an unexported helper function named getDefaultConfig was temporarily extracted inside backend/internal/adapters/config/loader.go with empty role mappings and an empty role order.

This caused an immediate compilation break:
- backend/internal/adapters/config/loader_test.go resides in package config_test (an external black-box test package).
- At line 1268, loader_test.go attempts to call getDefaultConfig().RoleOrder, which fails compilation because getDefaultConfig was unexported and cannot be called across package boundaries.

Beyond the compilation break, defining domain baseline defaults inside an outer adapter violates the Functional Core, Imperative Shell (FCIS) architecture:
- Core domain models and baseline defaults belong exclusively to the pure functional core in backend/internal/domain/.
- Outer adapters in backend/internal/adapters/config must consume domain models rather than defining domain defaults themselves.
- Default agent roles (orchestrator, build, planner, coder, tester, reviewer) and default role order were previously injected as synthetic fallbacks in multiple locations across domain derivation, configuration loading, YAML generation, and CLI output.

Iteration 03 executes four explicit architectural decisions:
1. Relocate the baseline default configuration constructor to backend/internal/domain/config.go as the pure exported function DefaultProjectConfig() domain.ProjectConfig.
2. Completely eliminate hardcoded default role mappings and hardcoded default role ordering across the entire system (domain aggregations, loader merge logic, writer YAML emission, and scaffolded templates). Empty role mappings (domain.RoleMappings{}) and empty role order ([]string{}) are treated as valid, first-class states. When role order is empty, role ordering across tables and cards resolves dynamically to alphabetical order via domain OrderRoles.
3. Fix all broken tests and compiler failures across Go backend packages (loader_test.go, writer_test.go, aggregations_test.go, server_test.go, main_test.go) and ensure full TypeScript alignment.
4. User directive: DO NOT COMMIT. Maintain an uncommitted working tree upon successful verification.

---

## 2. FCIS Architectural Invariants

1. Functional Core Purity:
   - backend/internal/domain/config.go must remain 100% pure. Zero imports of backend/internal/adapters/, backend/cmd/, net/http, or os.
   - DefaultProjectConfig() must construct and return domain.ProjectConfig by value with zero external dependencies, zero I/O, zero network, and zero clocks.
2. Value Semantics:
   - All configuration types use value semantics. DefaultProjectConfig returns a concrete struct, not a pointer.
3. Strict Typing and Zero Any:
   - Strictly zero uses of any or interface{} across all Go packages and TypeScript modules.
4. Boundary Discipline:
   - File traversal, YAML/JSON parsing, and serialization belong strictly in backend/internal/adapters/config/.
   - HTTP routing and response serialization belong strictly in backend/internal/adapters/http/.
5. Zero Emojis Mandate:
   - No emojis permitted in code, comments, docstrings, commit messages, plan files, review files, or test outputs.

---

## 3. Concrete Phased Implementation Tasks

### Phase 1: Pure Domain Defaults and Aggregation Alignment (backend/internal/domain/)

- Task 1.1: Add DefaultProjectConfig to backend/internal/domain/config.go:
  - Define exported function DefaultProjectConfig() ProjectConfig.
  - Construct and return a ProjectConfig value with pure baseline defaults:
    - ProjectName: empty string.
    - Ticket: TicketConfig with Prefixes initialized to an empty slice of TicketPrefix, CaseInsensitive set to true, CustomPatterns initialized to an empty slice of string.
    - Workspaces: WorkspacesConfig with Dir set to empty string, SrcFolders initialized to an empty slice of string, Artifacts initialized to an empty slice of WorkspaceArtifactConfig.
    - Database: DatabaseConfig with Path set to local OpenCode database path (~/.local/share/opencode/opencode.db), QueryTimeoutMs set to 5000, BusyTimeoutMs set to 5000, ReadOnly set to true.
    - Live: LiveConfig with ActiveThresholdSeconds set to 300, PollIntervalMs set to 2000, MaxTraceBufferMessages set to 500, AnomalyThresholds configured with StallThresholdSeconds set to 120, ErrorRateThreshold set to 0.25, ContextExhaustionPct set to 0.85.
    - CodeMetrics: CodeMetricsConfig with Include, Exclude, TestPatterns, ProdPatterns all initialized as non-nil empty slices of string.
    - RoleMappings: RoleMappings{} (non-nil, empty map).
    - RoleOrder: an empty slice of string (make([]string, 0), non-nil).
  - Verify that DefaultProjectConfig contains zero I/O, zero network calls, zero clock calls, and zero external dependencies.

- Task 1.2: Refactor DeriveMetricsLedger in backend/internal/domain/aggregations.go:
  - Remove hardcoded canonical role order slice containing orchestrator, planner, coder, tester, reviewer.
  - Initialize projection configuration via cfg := DefaultProjectConfig().
  - If roleMappings variadic argument is non-empty and roleMappings[0] is not nil, assign cfg.RoleMappings = roleMappings[0].
  - Delegate execution to DeriveMetricsLedgerWithConfig(factLedger, cfg).
  - Note: because cfg.RoleOrder is empty, OrderRoles and OrderSubagentRoles automatically sort present roles in ascending alphabetical order.

- Task 1.3: Author unit test TestDefaultProjectConfig in backend/internal/domain/config_test.go:
  - Call domain.DefaultProjectConfig().
  - Assert ProjectName is empty string.
  - Assert Workspaces.Dir is empty string and Workspaces.SrcFolders is empty slice.
  - Assert Ticket.Prefixes is empty slice and Ticket.CaseInsensitive is true.
  - Assert Database connection parameters match expected timeouts (5000ms) and ReadOnly is true.
  - Assert Live parameters match active threshold (300s), poll interval (2000ms), max trace (500), and anomaly thresholds (120s, 0.25, 0.85).
  - Assert CodeMetrics pattern slices are non-nil and empty.
  - Assert len(cfg.RoleMappings) is 0 and cfg.RoleMappings is not nil.
  - Assert len(cfg.RoleOrder) is 0 and cfg.RoleOrder is not nil.

- Task 1.4: Update TestDeriveMetricsLedger_RoleMappings in backend/internal/domain/aggregations_test.go:
  - Under empty RoleOrder, OrderRoles sorts present roles alphabetically.
  - Update expectedRoles from the old canonical order [orchestrator, coder, tester] to alphabetical order [coder, orchestrator, tester].

### Phase 2: Adapter Config Loader Clean-Up (backend/internal/adapters/config/)

- Task 2.1: Remove private getDefaultConfig from backend/internal/adapters/config/loader.go:
  - Delete func getDefaultConfig() domain.ProjectConfig completely from loader.go.

- Task 2.2: Update mergeConfig in backend/internal/adapters/config/loader.go:
  - Replace cfg := getDefaultConfig() with cfg := domain.DefaultProjectConfig().
  - Ensure that when dto.RoleMappings is nil or omitted, cfg.RoleMappings remains domain.RoleMappings{}.
  - Ensure that when dto.RoleOrder is nil or omitted, cfg.RoleOrder remains []string{}.
  - Ensure that when dto.RoleOrder is present:
    - Sanitize items by trimming whitespace and converting to lowercase.
    - Filter out empty strings.
    - If cleaned items exist, assign to cfg.RoleOrder. If cleaned slice is empty, assign make([]string, 0).

- Task 2.3: Fix loader_test.go compilation and test assertions in backend/internal/adapters/config/loader_test.go:
  - In TestLoad_RoleOrderParsingAndFallback:
    - Replace defaultOrder := getDefaultConfig().RoleOrder on line 1268 with defaultOrder := domain.DefaultProjectConfig().RoleOrder.
    - In sub-test fallback to DefaultProjectConfig when role_order omitted from YAML:
      - Assert that len(cfg.RoleOrder) is 0 and cfg.RoleOrder is not nil.
      - Assert that len(cfg.RoleOrder) equals len(defaultOrder).
    - In sub-test fallback to DefaultProjectConfig when role_order omitted from JSON:
      - Assert that len(cfg.RoleOrder) is 0 and cfg.RoleOrder is not nil.
      - Assert that len(cfg.RoleOrder) equals len(defaultOrder).
  - In TestLoad_PartialConfigurationMerging:
    - Add assertions verifying that when role_mappings and role_order are omitted from partial YAML, cfg.RoleMappings is empty map and cfg.RoleOrder is empty slice.

### Phase 3: Adapter Config Writer Clean-Up (backend/internal/adapters/config/)

- Task 3.1: Update GenerateConfigYAML in backend/internal/adapters/config/writer.go:
  - In the Role Mappings section:
    - Remove the fallback block that populates the 6 canonical roles (orchestrator, build, planner, coder, tester, reviewer) when len(cfg.RoleMappings) == 0.
    - If len(cfg.RoleMappings) == 0:
      - Write role_mappings: {}\n\n.
    - If len(cfg.RoleMappings) > 0:
      - Write role_mappings:\n followed by sorted keys formatted as indentation plus key colon space value newline, ending with a trailing newline.
  - In the Role Order section:
    - Remove the fallback block that populates the 5 canonical roles (orchestrator, planner, coder, tester, reviewer) when len(cfg.RoleOrder) == 0.
    - If len(cfg.RoleOrder) == 0:
      - Write role_order: []\n.
    - If len(cfg.RoleOrder) > 0:
      - Write role_order:\n followed by each role formatted as indentation hyphen space role newline.

- Task 3.2: Update GenerateDefaultYAML in backend/internal/adapters/config/writer.go:
  - Inside GenerateDefaultYAML, construct ProjectConfig with:
    - RoleMappings: domain.RoleMappings{}.
    - RoleOrder: []string{}.
  - Do NOT populate any default roles or default role orders.

- Task 3.3: Export GenerateDefaultConfigYAML in backend/internal/adapters/config/writer.go:
  - Define func GenerateDefaultConfigYAML with identical parameters to GenerateDefaultYAML.
  - Delegate directly: return GenerateDefaultYAML(projectName, workspacesDir, dbPath, srcFolders, prefixes, customArtifacts...).

- Task 3.4: Synchronize tests in backend/internal/adapters/config/writer_test.go:
  - In TestGenerateConfigYAML_Default:
    - When passed a configuration with empty RoleMappings and empty RoleOrder, assert that output contains role_mappings: {} and role_order: [].
    - Assert that output does NOT contain orchestrator: orchestrator, planner: planner, or - orchestrator.
  - Add TestGenerateConfigYAML_CustomRoles:
    - Pass a configuration with custom RoleMappings and custom RoleOrder.
    - Assert that custom mappings and custom order are rendered accurately.
  - In TestGenerateDefaultYAML_Overrides:
    - Assert that output contains role_mappings: {} and role_order: [].
    - When unmarshaled via LoadFromFile, assert that len(loaded.RoleMappings) == 0 and len(loaded.RoleOrder) == 0.
  - Add TestGenerateDefaultConfigYAML:
    - Verify GenerateDefaultConfigYAML outputs identical YAML bytes to GenerateDefaultYAML.

### Phase 4: Server, Repository, CLI, and Root Template Synchronization

- Task 4.1: Update NewServer in backend/internal/adapters/http/server.go:
  - Initialize cfg := domain.DefaultProjectConfig().
  - Attempt config.Load(workspacesDir); if successful, cfg = loaded.
  - Else attempt config.Load(""); if successful, cfg = loaded.
  - Pass resolved cfg to NewServerWithConfig.

- Task 4.2: Update .agenticlens-config.yaml in project root:
  - Note: Keep .agenticlens-config.yaml at root aligned with its configured values or remove role_mappings/role_order default assumptions if testing against root.

- Task 4.3: Synchronize backend/internal/adapters/http/server_test.go:
  - In TestConfigHandler_DefaultConfig:
    - Ensure test verifies that when role_order is empty, it serializes to empty slice and not null.

- Task 4.4: Synchronize backend/cmd/agentic-lens/main_test.go:
  - In TestInitCommand:
    - Verify that newly scaffolded configurations preserve empty RoleMappings and RoleOrder without injecting synthetic roles.

### Phase 5: Verification and Non-Commit Boundary Enforcement

- Task 5.1: Execute automated verification suite:
  - Run make check (formatting, vet, lint across backend and web).
  - Run make test (backend test suite with -race flag, frontend Vitest suite).
  - Run make build (web production build and Go single binary compilation).

- Task 5.2: Enforce Non-Commit Invariant:
  - Respect user command: DO NOT COMMIT.
  - Do NOT execute git commit, git push, or gh pr create.
  - Leave modified files in the working tree for user review.

---

## 4. Verification Gates for Tester

### Gate 1: Go Backend Compilation & Package Separation
- Command: cd backend && go test -c ./internal/adapters/config
- Expected: Compilation succeeds cleanly with zero errors. External test package config_test references domain.DefaultProjectConfig() without access violations.

### Gate 2: Pure Domain Baseline Defaults Test
- Command: cd backend && go test -v -run TestDefaultProjectConfig ./internal/domain
- Expected: All assertions in TestDefaultProjectConfig pass. RoleMappings is empty map, RoleOrder is empty slice.

### Gate 3: Config Loader Role Order Fallback Tests
- Command: cd backend && go test -v -run TestLoad_RoleOrderParsingAndFallback ./internal/adapters/config
- Expected: YAML and JSON without role_order load with len(RoleOrder) == 0. Custom role orders load as specified.

### Gate 4: Config Writer Empty Role Serialization Tests
- Command: cd backend && go test -v -run "TestGenerateConfigYAML|TestGenerateDefault" ./internal/adapters/config
- Expected: Emits role_mappings: {} and role_order: []. Zero default role injection. Round-trip unmarshaling verifies empty collections.

### Gate 5: HTTP Server /api/config Serialization Tests
- Command: cd backend && go test -v -run TestConfigHandler ./internal/adapters/http
- Expected: GET /api/config returns 200 OK with "role_order":[] when empty, never null.

### Gate 6: CLI Init Command Configuration Scaffold Test
- Command: cd backend && go test -v -run TestInitCommand ./cmd/agentic-lens
- Expected: Scaffolded configuration files contain role_mappings: {} and role_order: [].

### Gate 7: Domain Role Ordering Invariant (Alphabetical Fallback)
- Command: cd backend && go test -v -run TestDeriveMetricsLedger ./internal/domain
- Expected: DeriveMetricsLedger sorts roles alphabetically when RoleOrder is empty.

### Gate 8: Full Repository Verification Suite
- Commands: make check && make test && make build
- Expected: All checks, race-detector tests, and production build pass with exit code 0. Confirm git status shows modified files without new commits.

---

## 5. Reviewer Audit Checklist

### Axis 1: Spec Alignment
- DefaultProjectConfig() is defined in backend/internal/domain/config.go and exported.
- DefaultProjectConfig() returns empty RoleMappings (domain.RoleMappings{}) and empty RoleOrder ([]string{}).
- Private helper getDefaultConfig is completely removed from backend/internal/adapters/config/loader.go.
- mergeConfig in loader.go calls domain.DefaultProjectConfig().
- loader_test.go compiles cleanly without undefined symbol errors.
- GenerateConfigYAML in writer.go serializes role_mappings: {} when RoleMappings is empty, without injecting default roles.
- GenerateConfigYAML in writer.go serializes role_order: [] when RoleOrder is empty, without injecting default roles.
- GenerateDefaultYAML in writer.go constructs ProjectConfig with empty RoleMappings and empty RoleOrder.
- GenerateDefaultConfigYAML is exported in writer.go with identical signature and behavior to GenerateDefaultYAML.
- DeriveMetricsLedger in aggregations.go uses DefaultProjectConfig() rather than hardcoding canonical role order.
- User command DO NOT COMMIT was strictly respected (zero commits created).

### Axis 2: Engineering Standards & Invariants
- Functional Core, Imperative Shell (FCIS): backend/internal/domain/config.go maintains zero dependencies on adapters, filesystem, or I/O.
- Value Semantics: DefaultProjectConfig() returns domain.ProjectConfig by value.
- Type Safety: strictly zero uses of any or interface{} across all modified code.
- Zero Emojis Mandate: strictly zero emojis in code, comments, docstrings, plan files, review files, or test outputs.
- Zero Code Blocks in Plan: plan file is 100% descriptive prose, architectural specifications, and structured task lists with zero code blocks.
- Quality Gates: make check, make test, and make build pass cleanly.
