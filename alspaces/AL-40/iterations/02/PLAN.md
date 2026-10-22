# Implementation Plan: AL-40 Iteration 02 - Removal of Hardcoded Defaults from ConfigHandler

## 1. Goal and Iteration Scope

Deliver the follow-up refactoring for ticket AL-40 in response to architectural steering:
- Remove hardcoded domain defaults and synthetic fallback values from ConfigHandler in backend/internal/adapters/http/config_handler.go.
- Ensure ConfigHandler strictly reflects the configuration loaded into domain.ProjectConfig (sourced from .agenticlens-config.yaml and merged via backend/internal/adapters/config/loader.go) without fabricating synthetic project names, default ticket prefixes, workspace paths, or canonical role orders.
- Preserve boundary sanitization, defensive copying, and non-nil slice initialization so that empty slices serialize as empty JSON arrays rather than JSON nulls.
- Update test cases in backend/internal/adapters/http/server_test.go to verify that empty configurations remain empty in the serialized payload, while whitespace-only inputs are trimmed cleanly without injecting artificial fallbacks.
- Confirm frontend compatibility across web/src/ to verify that empty configuration fields are handled gracefully by existing presentation components and formatters without errors or layout shifts.
- Execute full verification via make check, make test, and make build under strict zero-any and zero-emoji mandates.

---

## 2. Grounded Reference Architecture and File Paths

The delta plan grounds against existing packages, files, and symbols:

### Go Imperative Boundary Shell (backend/internal/adapters/http/)
- backend/internal/adapters/http/config_handler.go:
  - Function ConfigHandler(cfg domain.ProjectConfig, workspacesDir string) http.HandlerFunc.
  - Current implementation contains fallback defaults for ProjectName (defaulting to AgenticLens), Prefixes (defaulting to AL, GEN, PO slice), WorkspacesDir (defaulting to workspaces), and RoleOrder (defaulting to 5 canonical roles).
  - Target implementation removes all four fallback injections, relying exclusively on values provided in cfg and workspacesDir while maintaining defensive copy and non-nil empty slice guarantees.
- backend/internal/adapters/http/dto.go:
  - Struct ProjectConfigDTO (fields: ProjectName, Prefixes, WorkspacesDir, RoleOrder, Version).
  - Function WriteProjectConfigResponse(w http.ResponseWriter, dto ProjectConfigDTO).
- backend/internal/adapters/http/server.go:
  - Struct Server with projectConfig domain.ProjectConfig field.
  - Constructor NewServerWithConfig(workspacesDir string, staticFS fs.FS, cfg domain.ProjectConfig, liveService ...LiveServiceInterface) *Server.
  - Constructor NewServer(workspacesDir string, staticFS fs.FS, liveService ...LiveServiceInterface) *Server resolving configuration via config.Load.
- backend/internal/adapters/http/server_test.go:
  - TestConfigHandler_DefaultConfig: Verifies GET /api/config when server is instantiated via NewServer, asserting response reflects values loaded from .agenticlens-config.yaml.
  - TestConfigHandler_CustomConfig: Verifies GET /api/config when server is instantiated via NewServerWithConfig with custom configuration parameters.
  - TestConfigHandler_EmptyFallbacks: Must be refactored to TestConfigHandler_EmptyConfigPreserved, asserting that empty ProjectConfig yields empty ProjectName, empty WorkspacesDir, empty non-nil Prefixes array, and empty non-nil RoleOrder array.
  - TestConfigHandler_AdversarialEdgeCases: Update subtests to assert trimmed empty strings instead of fallback string substitutions.

### Go Configuration Loading (backend/internal/adapters/config/)
- backend/internal/adapters/config/loader.go:
  - Functions Load(repoRoot string) and LoadFromFile(filePath string).
  - Function mergeConfig(dto rawProjectConfigDTO) domain.ProjectConfig, which establishes baseline defaults when loading configuration files.
- .agenticlens-config.yaml:
  - Project configuration containing project_name: AgenticLens, ticket prefixes (AL, AL2, GEN, PO, FIX), workspaces.dir: workspaces, role_mappings, and role_order.

### TypeScript Frontend Boundary & Presentation (web/src/)
- web/src/domain/types.ts:
  - Interfaces ProjectConfigResponse and TicketPrefixConfig.
- web/src/domain/formatters.ts:
  - Function combineDistinctPrefixes(discovered: readonly string[], configured?: readonly string[]): readonly string[]. Already provides fallback to AL, GEN, PO when prefix collections are completely empty.
- web/src/components/Navbar.tsx:
  - Conditionally renders project badge only when projectName is non-empty, preventing blank badge elements.
- web/src/views/LandingPage.tsx:
  - Conditionally renders project context only when projectName is present and non-empty.

---

## 3. Architectural Specifications and FCIS Invariants

### Functional Core, Imperative Shell (FCIS) Separation
1. Configuration Ownership Invariant:
   - Business defaults and configuration normalization belong exclusively to the configuration loading adapter in backend/internal/adapters/config/loader.go.
   - HTTP transport handlers must not synthesize domain models or invent substitute configuration data. Doing so creates discrepancies between the configuration active in domain aggregations and the metadata reported via the REST API.
2. Boundary Sanitization and Non-Nil Guarantees:
   - Strings: cfg.ProjectName and workspacesDir/cfg.Workspaces.Dir must be trimmed of leading and trailing whitespace via strings.TrimSpace. If empty, they remain empty string ("").
   - Slices: cfg.Ticket.Prefixes and cfg.RoleOrder must never serialize as JSON null. If nil or empty, allocate an empty slice (make with length 0) to guarantee JSON serialization produces []. If non-empty, copy defensively to avoid race conditions or external mutations.
   - Version: Derived from domain.GetBuildInfo().Version.

### REST Contract Delta
1. Response Payload for Populated Configuration (e.g. via .agenticlens-config.yaml):
   - project_name: "AgenticLens"
   - prefixes: [AL, AL2, GEN, PO, FIX] (matching .agenticlens-config.yaml)
   - workspaces_dir: "workspaces" (or server directory)
   - role_order: ["orchestrator", "planner", "coder", "tester", "reviewer"]
   - version: current semantic release or dev version string
2. Response Payload for Empty Configuration (e.g. zero-value domain.ProjectConfig{}):
   - project_name: ""
   - prefixes: []
   - workspaces_dir: ""
   - role_order: []
   - version: current semantic release or dev version string

---

## 4. Phased Work Breakdown

### Phase 1: Go HTTP Handler Refactoring
- File: backend/internal/adapters/http/config_handler.go
- Task 1.1: Remove ProjectName default fallback:
  - Read cfg.ProjectName and trim whitespace using strings.TrimSpace.
  - Do not check for empty string or assign "AgenticLens". Retain trimmed string as-is.
- Task 1.2: Remove Ticket.Prefixes default fallback:
  - Check len(cfg.Ticket.Prefixes).
  - If len is 0, initialize sanitizedPrefixes as make([]domain.TicketPrefix, 0).
  - If len > 0, initialize sanitizedPrefixes as make([]domain.TicketPrefix, len(cfg.Ticket.Prefixes)) and copy elements from cfg.Ticket.Prefixes.
  - Remove hardcoded AL, GEN, PO slice construction.
- Task 1.3: Remove WorkspacesDir default fallback:
  - Read workspacesDir argument; if trimmed workspacesDir is non-empty, use it.
  - Otherwise, read cfg.Workspaces.Dir and trim whitespace.
  - Do not check for empty string or assign "workspaces". Retain trimmed string as-is.
- Task 1.4: Remove RoleOrder default fallback:
  - Check len(cfg.RoleOrder).
  - If len is 0, initialize sanitizedRoleOrder as make([]string, 0).
  - If len > 0, initialize sanitizedRoleOrder as make([]string, len(cfg.RoleOrder)) and copy elements from cfg.RoleOrder.
  - Remove hardcoded orchestrator, planner, coder, tester, reviewer slice construction.
- Task 1.5: Populate ProjectConfigDTO and invoke WriteProjectConfigResponse:
  - Assign ProjectName, Prefixes, WorkspacesDir, RoleOrder, and Version (from domain.GetBuildInfo().Version).
  - Verify zero code lint or compilation issues.

### Phase 2: Go HTTP Server Test Suite Updates
- File: backend/internal/adapters/http/server_test.go
- Task 2.1: Update TestConfigHandler_DefaultConfig:
  - Assert that GET /api/config against NewServer returns 200 OK.
  - Assert that dto.ProjectName matches the configured project name from .agenticlens-config.yaml ("AgenticLens").
  - Assert that dto.WorkspacesDir matches wsDir.
  - Assert that dto.Prefixes contains the prefixes configured in .agenticlens-config.yaml (verifying AL, AL2, GEN, PO, FIX are present).
  - Assert that dto.RoleOrder matches the configured role order from .agenticlens-config.yaml.
- Task 2.2: Refactor TestConfigHandler_EmptyFallbacks to TestConfigHandler_EmptyConfigPreserved:
  - Create a server using NewServerWithConfig with empty workspacesDir ("") and empty domain.ProjectConfig{}.
  - Perform GET /api/config.
  - Assert HTTP status is 200 OK.
  - Assert dto.ProjectName is "" (empty string).
  - Assert dto.WorkspacesDir is "" (empty string).
  - Assert len(dto.Prefixes) is 0 and dto.Prefixes is non-nil.
  - Assert len(dto.RoleOrder) is 0 and dto.RoleOrder is non-nil.
  - Verify raw JSON response string contains `"prefixes":[]` and `"role_order":[]`, asserting no `"prefixes":null` or `"role_order":null`.
- Task 2.3: Update TestConfigHandler_AdversarialEdgeCases:
  - In subtest whitespace-only fields, pass hostileCfg with whitespace-only ProjectName and Workspaces.Dir, and empty workspacesDir argument.
  - Assert dto.ProjectName trims cleanly to "" without panic.
  - Assert dto.WorkspacesDir trims cleanly to "" without panic.
  - Confirm raw JSON output contains `"prefixes":[]` and `"role_order":[]`.
  - Maintain non-GET HTTP methods subtest ensuring 405 Method Not Allowed or 404 Not Found.

### Phase 3: Frontend Resilience & Integration Verification
- Files: web/src/components/Navbar.tsx, web/src/views/LandingPage.tsx, web/src/domain/formatters.ts
- Task 3.1: Verify Navbar presentation with empty project name:
  - Confirm Navbar renders platform brand "AgenticLens Telemetry" and does not render empty badge markup when projectName is empty or undefined.
- Task 3.2: Verify LandingPage hero presentation with empty project name:
  - Confirm LandingPage hero section renders generic telemetry header and omits project pill when projectName is empty or undefined.
- Task 3.3: Verify prefix combining when configured prefixes slice is empty:
  - Confirm web/src/domain/formatters.test.ts verifies combineDistinctPrefixes([], []) returns default fallback ['AL', 'GEN', 'PO'].
- Task 3.4: Execute frontend test suites:
  - Run npm test to verify all client, formatter, component, and view tests pass.

### Phase 4: Full Quality Gate Execution
- Task 4.1: Static analysis:
  - Run make check (golangci-lint, go vet, npm run lint, npm run typecheck).
- Task 4.2: Automated test execution:
  - Run make test (go test -race ./... and vitest run).
- Task 4.3: Binary compilation:
  - Run make build (npm run build and go build ./cmd/agentic-lens).
- Task 4.4: Standards audit:
  - Confirm zero emojis across all modified files, test files, and workspace artifacts.
  - Confirm zero any or untyped interface{} usage in Go and TypeScript code.

---

## 5. Verification & Quality Assurance Gates

1. Unit Test Verification:
   - Go: go test -v -race ./backend/internal/adapters/http -run TestConfigHandler
   - All 4 test cases (DefaultConfig, CustomConfig, EmptyConfigPreserved, AdversarialEdgeCases) must pass.
2. End-to-End Build Verification:
   - make clean && make
   - Verify single binary artifact bin/agentic-lens compiles with embedded web assets.
3. Contract Compliance Verification:
   - ConfigHandler contains zero hardcoded default strings ("AgenticLens", "workspaces", canonical roles, or canonical prefix arrays).
   - Zero code blocks in PLAN.md.
   - Zero emojis in all artifacts.

---

## 6. Risk Analysis & Mitigation

1. Risk: Nil slice serialization emitting `"prefixes": null` or `"role_order": null`.
   - Mitigation: Explicitly initialize slices using make with capacity 0 when input slice length is 0, ensuring Go json.Marshal produces `[]`.
2. Risk: TestConfigHandler_DefaultConfig failing in environments where .agenticlens-config.yaml is not found.
   - Mitigation: NewServer relies on config.Load which traverses upward to find the repository root configuration file. The test runs within the repository tree where .agenticlens-config.yaml is present.
3. Risk: Frontend crash if API returns empty prefixes.
   - Mitigation: combineDistinctPrefixes in web/src/domain/formatters.ts is already guarded and provides default prefixes when both discovered and configured collections are empty.
