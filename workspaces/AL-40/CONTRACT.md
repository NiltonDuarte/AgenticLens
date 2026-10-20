# Contract: AL-40 Dynamic Configuration REST API Endpoint and Observability Dashboard Branding

## 1. Problem Statement

During the pre-release source code audit (documented in workspaces/AL-35/AUDIT_REPORT.md, Section 3 and Section 4 Recommendation 2), a significant architectural gap was identified: the Go HTTP server exposes no route providing project configuration metadata to clients. Consequently, the TypeScript single-page application dashboard operates without dynamic awareness of ProjectConfig.ProjectName, Ticket.Prefixes, or Workspaces.Dir.

As a result, the frontend UI statically hardcodes "AgenticLens Telemetry", providing no mechanism to display the observed project name or brand context. In addition, multi-select prefix dropdowns across Fleet Overview (web/src/views/FleetOverview.tsx), Ticket Comparisons (web/src/views/TicketComparisons.tsx), and Metric Progression (web/src/views/MetricProgression.tsx) rely exclusively on runtime ledger scraping through extractDistinctPrefixes. Under this constraint, configured ticket prefixes for new or upcoming initiatives never appear in filter controls until at least one completed telemetry ledger has been committed to disk.

Ticket AL-40 resolves these issues by exposing a dedicated, sanitized GET /api/config REST endpoint in the Go HTTP server, creating TypeScript API client methods and state hydration, dynamically rendering the observed project name as a distinct badge in the navigation bar alongside permanent AgenticLens platform branding, updating the landing view to reflect project context, and populating prefix dropdowns via a pure union of configured and discovered prefixes.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All implementation deliverables must strictly adhere to TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. Go Functional Core (backend/internal/domain/):
   - Pure domain models in backend/internal/domain/config.go (ProjectConfig, TicketConfig, TicketPrefix, WorkspacesConfig) and version metadata in backend/internal/domain/version.go (GetBuildInfo, BuildInfo) remain immutable value structures.
   - Zero external imports: no imports of net/http, os, database/sql, or filesystem traversal packages.
   - Pure value semantics: domain functions receive immutable configuration copies and return deterministic models with zero side effects.
2. Go Imperative Boundary Shell (backend/internal/adapters/):
   - Boundary HTTP handler ConfigHandler in backend/internal/adapters/http/config_handler.go inspects Server configuration, sanitizes internal fields, and constructs the typed response payload.
   - DTO definitions and response writers in backend/internal/adapters/http/dto.go enforce concrete response type constraints via the ResponsePayload interface.
   - Server multiplexer wiring in backend/internal/adapters/http/server.go and entrypoint initialization in backend/cmd/agentic-lens/main.go cleanly route GET /api/config.
3. TypeScript Domain & Formatters (web/src/domain/):
   - Pure domain contracts in web/src/domain/types.ts define TicketPrefixConfig and ProjectConfigResponse matching Go DTO semantics.
   - Pure formatters in web/src/domain/formatters.ts implement deterministic prefix extraction, union, deduplication, and alphabetical sorting with zero DOM, browser, or network dependencies.
4. TypeScript Boundary & Presentation Shell (web/src/api/, web/src/components/, web/src/views/):
   - Boundary API client in web/src/api/client.ts performs asynchronous HTTP fetching with resilient fallback to deterministic fixtures in web/src/api/mockData.ts.
   - Pure presentation components (web/src/components/Navbar.tsx, web/src/views/LandingPage.tsx, web/src/views/FleetOverview.tsx, web/src/views/TicketComparisons.tsx, web/src/views/MetricProgression.tsx) consume typed props without embedding unvalidated network mutations.

### Strong Typing & Zero Any Invariant
1. Zero usage of any or empty interface{} across Go backend and TypeScript frontend codebases.
2. Go DTOs use explicit concrete types with struct tags for JSON serialization.
3. TypeScript definitions use strict read-only interfaces and discriminated unions where multi-state lifecycles are represented.

### Zero Emojis Mandate
Zero emojis permitted across all Go code, TypeScript code, comments, docstrings, commit messages, PR drafts, UI elements, badges, tooltips, or workspace artifacts.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Go Imperative Boundary Shell (backend/internal/adapters/http/):
   - Define ProjectConfigDTO in backend/internal/adapters/http/dto.go containing project_name, prefixes, workspaces_dir, role_order, and version.
   - Add ProjectConfigDTO to the ResponsePayload type union and create WriteProjectConfigResponse.
   - Implement ConfigHandler in backend/internal/adapters/http/config_handler.go returning sanitized JSON.
   - Extend Server in backend/internal/adapters/http/server.go with projectConfig domain.ProjectConfig.
   - Provide NewServerWithConfig accepting workspacesDir, staticFS, cfg domain.ProjectConfig, and optional liveService.
   - Maintain NewServer backward compatibility by resolving default configuration and delegating to NewServerWithConfig.
   - Register route GET /api/config on the server multiplexer.
   - Wire loaded configuration into NewServerWithConfig inside backend/cmd/agentic-lens/main.go.
   - Implement comprehensive unit tests in backend/internal/adapters/http/server_test.go covering 200 OK responses, JSON structure, sanitization, default fallbacks, and custom overrides.
2. TypeScript Domain Layer & Pure Formatters (web/src/domain/):
   - Define TicketPrefixConfig and ProjectConfigResponse interfaces in web/src/domain/types.ts.
   - Update extractDistinctPrefixes or implement combineDistinctPrefixes in web/src/domain/formatters.ts to merge discovered ticket prefixes with configured prefixes, deduplicate, sort ascending alphabetically, and provide canonical fallback (AL, GEN, PO) when empty.
   - Add unit tests in web/src/domain/formatters.test.ts verifying prefix merging, deduplication, sorting, and edge case fallbacks.
3. TypeScript API Client & Mock Fixtures (web/src/api/):
   - Define mockProjectConfig in web/src/api/mockData.ts adhering to ProjectConfigResponse.
   - Implement fetchProjectConfig in web/src/api/client.ts returning Promise of ApiResponse of ProjectConfigResponse with graceful mock fallback on network or parse failures.
   - Add unit tests in web/src/api/client.test.ts verifying live fetch, network error fallback, 500 error fallback, and 404 error fallback.
4. TypeScript Presentation Shell (web/src/components/, web/src/views/, web/src/):
   - Update web/src/components/Navbar.tsx to accept optional projectConfig or projectName props, preserve AgenticLens permanently as the core platform brand, and display the observed project name as a distinct badge when non-empty.
   - Update web/src/components/Navbar.test.tsx asserting brand preservation, project badge rendering, and graceful omission handling.
   - Update web/src/views/LandingPage.tsx and web/src/views/LandingPage.test.tsx to accept project context and display the observed project name in the hero section when present.
   - Update web/src/App.tsx to hydrate project configuration on startup using fetchProjectConfig, pass project context to Navbar and LandingPage, and pass configured prefixes to analytical views.
   - Update web/src/views/FleetOverview.tsx, web/src/views/TicketComparisons.tsx, and web/src/views/MetricProgression.tsx to union configured prefixes with discovered prefixes for filter controls.
   - Update corresponding view test suites to verify that configured prefixes appear in dropdowns even in the absence of matching workspaces.

### Out of Scope
1. Altering OpenCode SQLite database schema, live query hooks, or SSE stream architectures.
2. Modifying Git commit diff analysis, inference rules, or code classification patterns in backend/internal/adapters/git/.
3. Changing telemetry generation CLI commands, backfill migrations, or schema validator logic.
4. Replacing or altering the AgenticLens platform name or root brand identity.

---

## 4. Architectural Contracts and Data Specifications

### REST API Endpoint Contract: GET /api/config
1. HTTP Method: GET
2. Path: /api/config
3. Response Headers:
   - Content-Type: application/json; charset=utf-8
4. Response Status Codes:
   - 200 OK: Valid JSON payload containing sanitized configuration parameters.
5. Response Payload JSON Schema:
   - project_name: string (e.g. "AgenticLens" or observed project name)
   - prefixes: array of objects, where each object contains:
     - prefix: string (e.g. "AL")
     - description: string (e.g. "Core AgenticLens development features")
     - ticket_type: optional string (e.g. "chore", "docs")
   - workspaces_dir: string (e.g. "workspaces" or active directory path)
   - role_order: array of strings (e.g. ["orchestrator", "planner", "coder", "tester", "reviewer"])
   - version: string (e.g. "v0.1.0-dev" or semantic release string)
6. Sanitization Rules:
   - Omit database credentials, SQLite paths, and internal query timeouts.
   - Omit internal filesystem glob patterns, include/exclude lists, and git author mappings.
   - Omit live threshold numbers unless needed for presentation.
   - If project_name is empty, default to "AgenticLens".
   - If prefixes is empty, default to standard prefixes AL, GEN, PO.
   - If workspaces_dir is empty, default to "workspaces".

### Frontend Domain Types & API Client Contract
1. TicketPrefixConfig Interface:
   - readonly prefix: string
   - readonly description: string
   - readonly ticket_type?: string
2. ProjectConfigResponse Interface:
   - readonly project_name: string
   - readonly prefixes: readonly TicketPrefixConfig[]
   - readonly workspaces_dir: string
   - readonly role_order: readonly string[]
   - readonly version: string
3. Client Method:
   - fetchProjectConfig(): Promise<ApiResponse<ProjectConfigResponse>>
   - Returns ApiResponse wrapping data and isMock flag.
   - Returns mockProjectConfig on network failure, abort, or HTTP error.

### Navbar & Observability Portal Branding Contract
1. Platform Brand Invariant:
   - The brand title "AgenticLens" and subtitle "Telemetry" remain permanently rendered as the core platform identifier.
   - Clicking or keyboard-activating the brand element navigates to the landing view.
2. Observed Project Badge:
   - When project_name is non-empty, render an adjacent badge or subtitle tag (e.g. className "project-badge" or "brand-project-tag") containing the observed project name.
   - When project_name matches "AgenticLens" or is omitted, the badge cleanly renders or gracefully simplifies without breaking layout.
   - Tooltip title provides descriptive text: "Observed Project: <project_name>".
3. Landing Page Context:
   - Landing view hero section displays the observed project name when provided, maintaining clear distinction between the AgenticLens observability tool and the target repository under observation.

### Prefix Filter Population & Formatters Contract
1. Multi-Select Dropdown Integration:
   - Dropdown options are formed by uniting configured prefixes (from ProjectConfigResponse.prefixes) and discovered prefixes (from loaded workspace ledgers).
   - Each option has value equal to prefix and label equal to prefix.
   - Options are deduplicated and sorted in alphabetical ascending order.
   - If neither configured nor discovered prefixes are available, fallback to default prefixes AL, GEN, PO.
