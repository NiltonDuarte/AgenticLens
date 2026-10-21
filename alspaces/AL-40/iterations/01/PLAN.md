# Implementation Plan: AL-40 Dynamic Configuration REST API Endpoint and Observability Dashboard Branding

## 1. Goal and Iteration Scope

Deliver the complete end-to-end implementation of ticket AL-40 across the Go backend and TypeScript frontend:
- Add ProjectConfigDTO and response serialization in backend/internal/adapters/http/dto.go.
- Implement ConfigHandler in backend/internal/adapters/http/config_handler.go and register GET /api/config on the server.
- Update Server and NewServer in backend/internal/adapters/http/server.go to accept and hold project configuration, providing NewServerWithConfig.
- Wire loaded configuration in backend/cmd/agentic-lens/main.go.
- Write Go HTTP server tests in backend/internal/adapters/http/server_test.go validating the new endpoint.
- Define TicketPrefixConfig and ProjectConfigResponse in web/src/domain/types.ts.
- Enhance prefix combining and deduplication in web/src/domain/formatters.ts with test coverage in web/src/domain/formatters.test.ts.
- Define mockProjectConfig in web/src/api/mockData.ts and implement fetchProjectConfig in web/src/api/client.ts with test coverage in web/src/api/client.test.ts.
- Update web/src/components/Navbar.tsx to render the observed project name badge alongside permanent AgenticLens branding, with test coverage in web/src/components/Navbar.test.tsx.
- Update web/src/views/LandingPage.tsx to reference the observed project context, with test coverage in web/src/views/LandingPage.test.tsx.
- Update web/src/App.tsx to hydrate project configuration on startup and pass configured prefixes into views.
- Update web/src/views/FleetOverview.tsx, web/src/views/TicketComparisons.tsx, and web/src/views/MetricProgression.tsx to display configured prefixes alongside discovered prefixes in filter dropdowns.
- Verify all quality gates pass: make check, make test, and make build, ensuring zero any and zero emojis.

---

## 2. Grounded Reference Architecture & File Paths

The plan grounds against existing concrete packages, files, and symbols:

### Go Functional Core (backend/internal/domain/)
- backend/internal/domain/config.go:
  - ProjectConfig struct (fields: ProjectName, Ticket, Workspaces, Database, Live, CodeMetrics, RoleMappings, RoleOrder).
  - TicketPrefix struct (fields: Prefix, Description, TicketType).
  - TicketConfig struct (fields: Prefixes, CaseInsensitive, CustomPatterns).
  - DefaultProjectConfig() returning standard baseline defaults.
- backend/internal/domain/version.go:
  - GetBuildInfo() returning BuildInfo struct with Version, GitCommit, BuildDate.

### Go Boundary Adapters (backend/internal/adapters/)
- backend/internal/adapters/http/dto.go:
  - Add ProjectConfigDTO struct with JSON tags: project_name, prefixes, workspaces_dir, role_order, version.
  - Extend ResponsePayload interface to include ProjectConfigDTO.
  - Add WriteProjectConfigResponse(w http.ResponseWriter, dto ProjectConfigDTO) function.
- backend/internal/adapters/http/config_handler.go (new file):
  - Implement ConfigHandler(cfg domain.ProjectConfig, workspacesDir string) http.HandlerFunc.
  - Build and return ProjectConfigDTO with version from domain.GetBuildInfo().Version.
- backend/internal/adapters/http/server.go:
  - Extend Server struct with projectConfig domain.ProjectConfig.
  - Implement NewServerWithConfig(workspacesDir string, staticFS fs.FS, cfg domain.ProjectConfig, liveService ...LiveServiceInterface) *Server.
  - Update NewServer to load configuration or fallback to DefaultProjectConfig() and delegate to NewServerWithConfig.
  - Register route mux.HandleFunc("GET /api/config", ConfigHandler(s.projectConfig, s.workspacesDir)).
- backend/internal/adapters/http/server_test.go:
  - Add unit tests verifying GET /api/config returns 200 OK, Content-Type application/json, and matches ProjectConfigDTO schema.
  - Add test asserting default configuration values when no config file is passed.
  - Add test asserting custom project name and prefix values when custom configuration is passed.
- backend/cmd/agentic-lens/main.go:
  - In runServer: replace NewServer call with NewServerWithConfig(workspacesDir, staticFS, projCfg, liveService).

### TypeScript Domain & Pure Formatters (web/src/domain/)
- web/src/domain/types.ts:
  - Define TicketPrefixConfig interface: prefix (string), description (string), ticket_type (optional string).
  - Define ProjectConfigResponse interface: project_name (string), prefixes (readonly TicketPrefixConfig[]), workspaces_dir (string), role_order (readonly string[]), version (string).
- web/src/domain/formatters.ts:
  - Update extractDistinctPrefixes or add combineDistinctPrefixes(discovered: readonly string[], configured?: readonly string[]): readonly string[].
  - Merge discovered and configured prefix strings, trim, filter empty, deduplicate via Set, sort ascending alphabetically, and fallback to default ['AL', 'GEN', 'PO'] if resulting set is empty.
- web/src/domain/formatters.test.ts:
  - Add unit tests verifying merging of configured prefixes with empty discovered prefixes.
  - Add unit tests verifying deduplication when configured and discovered prefixes overlap.
  - Add unit tests verifying alphabetical sorting and whitespace trimming.

### TypeScript API Client & Mock Fixtures (web/src/api/)
- web/src/api/mockData.ts:
  - Define and export mockProjectConfig: ProjectConfigResponse with project_name 'AgenticLens', standard prefixes, 'workspaces' directory, canonical role order, and 'v0.1.0-dev' version.
- web/src/api/client.ts:
  - Implement fetchProjectConfig(): Promise<ApiResponse<ProjectConfigResponse>>.
  - Perform fetch('/api/config') and parse JSON on 200 OK, with fallback to mockProjectConfig on HTTP errors or network exceptions.
- web/src/api/client.test.ts:
  - Add test suite for fetchProjectConfig verifying successful response parsing, 500 error fallback, 404 error fallback, and network rejection fallback.

### TypeScript Presentation Shell (web/src/components/, web/src/views/, web/src/)
- web/src/components/Navbar.tsx:
  - Extend NavbarProps with optional projectConfig?: ProjectConfigResponse and projectName?: string.
  - In Navbar, preserve permanent "AgenticLens" brand title and "Telemetry" subtitle.
  - When projectName is non-empty, render a styled project badge adjacent to the brand title/subtitle.
  - If projectConfig is not passed as a prop, trigger fallback fetchProjectConfig in useEffect similar to fetchVersion.
- web/src/components/Navbar.test.tsx:
  - Add tests verifying project badge rendering when projectName / projectConfig is supplied.
  - Add tests verifying fallback behavior when projectName / projectConfig is omitted.
  - Verify zero emojis in rendered markup.
- web/src/views/LandingPage.tsx:
  - Extend LandingPageProps with optional projectConfig?: ProjectConfigResponse | null or projectName?: string.
  - In the hero section, render an observed project pill or contextual subtitle when project name is provided.
- web/src/views/LandingPage.test.tsx:
  - Add tests asserting that observed project context renders when prop is supplied.
- web/src/App.tsx:
  - In App component, declare state for projectConfig: ProjectConfigResponse | null.
  - In useEffect on mount, call fetchProjectConfig and update projectConfig state.
  - Pass projectConfig and projectName to Navbar.
  - Pass projectConfig and projectName to LandingPage.
  - Pass configuredPrefixes to FleetOverview, TicketComparisons, and MetricProgression.
- web/src/views/FleetOverview.tsx:
  - Accept configuredPrefixes?: readonly string[] in FleetOverviewProps.
  - Update prefixOptions memoization to use combineDistinctPrefixes uniting knownPrefixes with configuredPrefixes.
- web/src/views/TicketComparisons.tsx:
  - Accept configuredPrefixes?: readonly string[] in TicketComparisonsProps.
  - Update prefixOptions memoization to use combineDistinctPrefixes uniting knownPrefixes with configuredPrefixes.
- web/src/views/MetricProgression.tsx:
  - Accept configuredPrefixes?: readonly string[] in MetricProgressionProps.
  - Update prefixOptions memoization to use combineDistinctPrefixes uniting knownPrefixes with configuredPrefixes.
- web/src/views/FleetOverview.test.tsx, TicketComparisons.test.tsx, MetricProgression.test.tsx:
  - Update tests to assert that configured prefixes (e.g. new prefix without workspaces) render in dropdown filter options.

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Go HTTP Boundary Adapter & Server Wiring
- Task 1.1: In backend/internal/adapters/http/dto.go:
  - Define ProjectConfigDTO struct with fields:
    - ProjectName string (json:"project_name")
    - Prefixes []domain.TicketPrefix (json:"prefixes")
    - WorkspacesDir string (json:"workspaces_dir")
    - RoleOrder []string (json:"role_order")
    - Version string (json:"version")
  - Add ProjectConfigDTO to the ResponsePayload interface type list.
  - Implement WriteProjectConfigResponse(w http.ResponseWriter, dto ProjectConfigDTO). Ensure nil slices are normalized to empty slices before serialization.
- Task 1.2: Create backend/internal/adapters/http/config_handler.go:
  - Implement ConfigHandler(cfg domain.ProjectConfig, workspacesDir string) http.HandlerFunc.
  - Build ProjectConfigDTO:
    - project_name: cfg.ProjectName if non-empty, otherwise "AgenticLens".
    - prefixes: cfg.Ticket.Prefixes if non-empty, otherwise domain.DefaultProjectConfig().Ticket.Prefixes.
    - workspaces_dir: workspacesDir if non-empty, otherwise cfg.Workspaces.Dir (or "workspaces").
    - role_order: cfg.RoleOrder if non-empty, otherwise domain.DefaultProjectConfig().RoleOrder.
    - version: domain.GetBuildInfo().Version.
  - Dispatch response using WriteProjectConfigResponse(w, dto).
- Task 1.3: In backend/internal/adapters/http/server.go:
  - Add projectConfig domain.ProjectConfig to Server struct.
  - Implement NewServerWithConfig(workspacesDir string, staticFS fs.FS, cfg domain.ProjectConfig, liveService ...LiveServiceInterface) *Server.
  - In NewServerWithConfig, initialize Server and register mux.HandleFunc("GET /api/config", ConfigHandler(cfg, workspacesDir)).
  - Update NewServer(workspacesDir string, staticFS fs.FS, liveService ...LiveServiceInterface) *Server to resolve configuration via domain.DefaultProjectConfig() and delegate directly to NewServerWithConfig.
- Task 1.4: In backend/cmd/agentic-lens/main.go:
  - In runServer function, update server instantiation to pass projCfg via adaptershttp.NewServerWithConfig(workspacesDir, staticFS, projCfg, liveService).

### Phase 2: Go Backend Testing & Verification
- Task 2.1: In backend/internal/adapters/http/server_test.go:
  - Create TestConfigHandler_DefaultConfig asserting GET /api/config returns 200 OK, Content-Type application/json, and default fields matching domain.DefaultProjectConfig().
  - Create TestConfigHandler_CustomConfig creating a Server with custom ProjectConfig (e.g. ProjectName: "CustomApp", custom prefixes, custom workspaces dir) and asserting response payload values match custom inputs.
  - Verify that version matches domain.GetBuildInfo().Version.
- Task 2.2: Execute Go test suite:
  - Run go test -race ./backend/internal/adapters/http/...
  - Run go vet ./backend/...

### Phase 3: TypeScript Domain Contracts & Pure Formatters
- Task 3.1: In web/src/domain/types.ts:
  - Define and export TicketPrefixConfig interface:
    - readonly prefix: string
    - readonly description: string
    - readonly ticket_type?: string
  - Define and export ProjectConfigResponse interface:
    - readonly project_name: string
    - readonly prefixes: readonly TicketPrefixConfig[]
    - readonly workspaces_dir: string
    - readonly role_order: readonly string[]
    - readonly version: string
- Task 3.2: In web/src/domain/formatters.ts:
  - Update extractDistinctPrefixes or add combineDistinctPrefixes(discovered: readonly string[], configured?: readonly string[]): readonly string[].
  - Clean and trim tokens, insert into Set, sort ascending alphabetically, and if empty return default ['AL', 'GEN', 'PO'].
- Task 3.3: In web/src/domain/formatters.test.ts:
  - Add tests for combineDistinctPrefixes:
    - Combines configured prefixes with empty discovered prefixes.
    - Merges overlapping configured and discovered prefixes without duplicates.
    - Handles case differences, whitespace trimming, and alphabetical sort order.
    - Returns ['AL', 'GEN', 'PO'] when both inputs are empty.

### Phase 4: TypeScript API Client & Mock Data
- Task 4.1: In web/src/api/mockData.ts:
  - Define and export mockProjectConfig: ProjectConfigResponse populated with:
    - project_name: 'AgenticLens'
    - prefixes: standard AL, GEN, PO prefixes with descriptions and ticket types
    - workspaces_dir: 'workspaces'
    - role_order: CANONICAL_ROLE_ORDER
    - version: 'v0.1.0-dev'
- Task 4.2: In web/src/api/client.ts:
  - Export fetchProjectConfig(): Promise<ApiResponse<ProjectConfigResponse>>.
  - Make GET request to /api/config, return data with isMock: false on 200 OK.
  - Catch network exceptions or non-OK status codes and return mockProjectConfig with isMock: true.
- Task 4.3: In web/src/api/client.test.ts:
  - Add unit tests for fetchProjectConfig:
    - Live 200 OK response parsing.
    - Fallback on network failure.
    - Fallback on 500 server error.
    - Fallback on 404 response.
    - Fallback on invalid JSON payload.

### Phase 5: Navigation Bar & Branding Presentation Components
- Task 5.1: In web/src/components/Navbar.tsx:
  - Update NavbarProps to include optional projectConfig?: ProjectConfigResponse and optional projectName?: string.
  - Add internal state for projectConfig/projectName, with fallback useEffect fetching fetchProjectConfig when prop is omitted.
  - Retain permanent core platform brand: "AgenticLens" title and "Telemetry" subtitle.
  - When observed project name is present, render an adjacent project badge (e.g. span with className "project-badge" or "brand-project-tag" and title "Observed Project: <name>").
- Task 5.2: In web/src/components/Navbar.test.tsx:
  - Add test: renders project badge when projectConfig or projectName prop is provided.
  - Add test: gracefully handles omission and renders default brand without broken badges.
  - Add test: verifies permanent presence of AgenticLens platform brand regardless of project configuration.
  - Add test: asserts zero emojis across rendered HTML.
- Task 5.3: In web/src/views/LandingPage.tsx:
  - Update LandingPageProps to include optional projectConfig?: ProjectConfigResponse | null or projectName?: string.
  - In the hero section, render an observed project pill or contextual subtitle when project name is provided.
- Task 5.4: In web/src/views/LandingPage.test.tsx:
  - Add test verifying observed project context renders in the hero section when prop is provided.

### Phase 6: Multi-Select Prefix Filtering Integration Across Views
- Task 6.1: In web/src/App.tsx:
  - In App component, declare state for projectConfig: ProjectConfigResponse | null.
  - In useEffect on mount, call fetchProjectConfig and update projectConfig state.
  - Pass projectConfig and projectName to Navbar.
  - Pass projectConfig and projectName to LandingPage.
  - Pass configuredPrefixes to FleetOverview, TicketComparisons, and MetricProgression.
- Task 6.2: In web/src/views/FleetOverview.tsx:
  - Accept configuredPrefixes?: readonly string[] in FleetOverviewProps.
  - Update prefixOptions memoization to use combineDistinctPrefixes uniting knownPrefixes with configuredPrefixes.
- Task 6.3: In web/src/views/TicketComparisons.tsx:
  - Accept configuredPrefixes?: readonly string[] in TicketComparisonsProps.
  - Update prefixOptions memoization to use combineDistinctPrefixes uniting knownPrefixes with configuredPrefixes.
- Task 6.4: In web/src/views/MetricProgression.tsx:
  - Accept configuredPrefixes?: readonly string[] in MetricProgressionProps.
  - Update prefixOptions memoization to use combineDistinctPrefixes uniting knownPrefixes with configuredPrefixes.
- Task 6.5: In view test files (FleetOverview.test.tsx, TicketComparisons.test.tsx, MetricProgression.test.tsx):
  - Add test cases verifying that passed configuredPrefixes (e.g. ['FEATURE', 'INFRA']) appear in prefix dropdown options even when no ledgers for those prefixes exist.

### Phase 7: Comprehensive Verification & Quality Gates
- Task 7.1: Go backend verification:
  - Run go test -race ./...
  - Run go vet ./...
- Task 7.2: TypeScript frontend verification:
  - Run npm test
  - Run npm run typecheck
  - Run npm run lint
- Task 7.3: Single binary compilation:
  - Run npm run build
  - Run go build -o bin/agentic-lens ./cmd/agentic-lens
  - Verify single binary starts and responds to GET /api/config.
- Task 7.4: Verification of engineering invariants:
  - Verify zero any across all modified Go and TypeScript files.
  - Verify zero emojis across all code, tests, and UI strings.
