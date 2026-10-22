# Implementation Plan: AL-5 Single Binary Integration Parity and Deprecation

## 1. Goal and Iteration Scope
Deliver single-binary integration parity, Go codebase relocation, and complete Python decommissioning:
1. Audit and document all visualizations present in the legacy Streamlit application missing in TypeScript web (completed in workspaces/AL-5/STREAMLIT_GRAPH_PARITY_REPORT.md).
2. Execute cross-engine parity verification comparing Go domain calculations against Python reference rollups across real workspaces and edge-case fixtures.
3. Relocate the Go application from the repository root to a dedicated backend directory (backend/cmd, backend/internal, backend/go.mod, backend/go.sum).
4. Coordinate frontend and backend compilation in Makefile and build workflows so that make build reliably yields a self-contained executable at bin/agentic-lens.
5. Completely purge all Python artifacts: agentic_lens/, scripts/extract_agentic_metrics.py, tests/, pyproject.toml, uv.lock.
6. Modernize Dockerfile into a multi-stage build producing a minimal standalone container running the Go binary with zero Node.js and zero Python runtime dependencies.
7. Verify standalone execution of ./bin/agentic-lens server without Node.js or Python runtime dependencies.

Zero code blocks are permitted in this plan. Implementation belongs exclusively to the coder subagent.

## 2. Grounded Reference Architecture and Symbol Map
- Backend Domain Engine: backend/internal/domain/aggregations.go (ComputeTotals, ComputeWorkflowEfficiency, ComputeRoleMetrics, ComputeModelMetrics, DeriveMetricsLedger, CalculateFleetSummary, GetBaseHealthStatus)
- Backend Domain Models: backend/internal/domain/models.go (AgenticFactLedger, AgenticMetricsLedger, WorkflowEfficiency, GroupedMetric, ContextWindowHealth)
- Backend Embedded Asset Adapter: backend/internal/adapters/web/embed.go (GetFS, distFS), backend/internal/adapters/web/embed_test.go
- Backend HTTP Server: backend/internal/adapters/http/server.go (NewServer, Server.ListenAndServe), backend/internal/adapters/http/spa.go (SPAHandler)
- Backend CLI and Server Entrypoint: backend/cmd/agentic-lens/main.go (runServer, runGenerate, runBackfill, main)
- Frontend Build Config: web/package.json, web/vite.config.ts, web/tsconfig.json
- Root Configuration and Manifests: Makefile, Dockerfile, .agenticlens/config.json, TECHNICAL_GUIDELINES.md
- Visual Parity Report: workspaces/AL-5/STREAMLIT_GRAPH_PARITY_REPORT.md

## 3. Phased Implementation Tasks for Coder Subagent

### Phase 1: Cross-Engine Parity Verification Suite
- Task 1.1: Author synthetic and real-world parity test fixtures:
  - Create fixture ledgers covering real workspaces (AL-1, AL-2, AL-11, AL-12, PO-1) and edge cases (zero invocations, orchestrator-only, high token saturation exceeding 200k ceiling, pure test code 1:0 ratio, pure prod code 0:1 ratio).
- Task 1.2: Execute cross-validation test comparing Go domain calculations against Python reference rollups:
  - Assert 100% parity across:
    - Invocations count, input tokens, output tokens, cache read tokens, cache hit rate percentage.
    - Cost USD within floating point tolerance of 0.0001.
    - Context health status string, orchestrator saturation percentage, subagent peak context, and peak role.
    - Active duration in seconds and formatted duration string.
    - Subagent average turns and peak turns.
    - Test to production code ratio string, test lines added, and prod lines added.
    - Tool calls count, error corrections count, and tool success rate percentage.
- Task 1.3: Encode permanent Go parity regression tests:
  - Ensure backend/internal/domain/aggregations_test.go permanently tests all mathematical parity scenarios with zero external dependencies.

### Phase 2: Relocate Go Application to Dedicated backend/ Directory
- Task 2.1: Move Go files:
  - Move cmd/ to backend/cmd/.
  - Move internal/ to backend/internal/.
  - Move go.mod and go.sum to backend/.
- Task 2.2: Update root Makefile:
  - Configure backend commands to run in backend/ directory (e.g. go test -race ./..., go vet ./... inside backend).
  - Configure build target: run npm run build in web/, ensure compiled assets are placed in backend/internal/adapters/web/dist, and run go build to output the binary to root bin/agentic-lens.
  - Retain graceful fallback when Node.js is absent but pre-existing dist assets are present in backend/internal/adapters/web/dist.
- Task 2.3: Update project configuration:
  - Update .agenticlens/config.json src_folders to ["web", "backend"].
  - Update any skill references or script paths if needed.
- Task 2.4: Update Dockerfile:
  - Multi-stage container build:
    - Stage 1: node alpine to build web/ distribution.
    - Stage 2: golang alpine to copy web/dist into backend/internal/adapters/web/dist and compile backend/cmd/agentic-lens to bin/agentic-lens.
    - Stage 3: minimal alpine or scratch image with only the compiled binary and ca-certificates.
    - Entrypoint runs /bin/agentic-lens server --port 8085 --workspaces /workspaces.

### Phase 3: Complete Python Purge and Cleanup
- Task 3.1: Verify parity passed cleanly before deletion.
- Task 3.2: Delete Python source packages and scripts:
  - Delete agentic_lens/ (entire presentation, domain, and adapter packages).
  - Delete scripts/extract_agentic_metrics.py.
  - Delete tests/ (Python test directory).
- Task 3.3: Delete Python dependency manifests and virtual environment artifacts:
  - Delete pyproject.toml.
  - Delete uv.lock.
  - Delete .python-version, .pytest_cache, and any __pycache__ folders if present.
- Task 3.4: Verify repository hygiene:
  - Confirm zero Python files (.py) remain in the codebase.
  - Confirm git status is clean of untracked Python residue.

### Phase 4: Standalone Execution Verification and Server Validation
- Task 4.1: Verify standalone execution of bin/agentic-lens server:
  - Build executable via make build.
  - Launch server in temporary background process.
  - Verify HTTP responses:
    - Root path GET / returns HTTP 200 with HTML content containing the React root div.
    - Embedded assets GET /assets/index-*.js and GET /assets/index-*.css return HTTP 200 with correct JavaScript and CSS mime types.
    - SPA client-side routing fallback: requesting a non-file path such as GET /deep-dive returns HTTP 200 with index.html content.
    - API endpoint GET /api/workspaces returns HTTP 200 with JSON payload containing scanned workspace ledgers.
    - Non-existent API route GET /api/nonexistent returns HTTP 404 with JSON error format, not index.html.
  - Test graceful shutdown behavior upon receiving interrupt signal.

### Phase 5: Quality Gate Audits and Workspace Deliverables
- Task 5.1: Run full quality gates:
  - Run make check (go vet in backend, npm run typecheck, npm run lint, git diff check).
  - Run make test (go test -race in backend, npm test in web).
  - Run make build to produce standalone binary.
- Task 5.2: Generate ticket telemetry and documentation:
  - Run AgenticLens CLI generate command with unique nonce to record telemetry into workspaces/AL-5/agentic_metrics.json and append metrics table to workspaces/AL-5/SUMMARY.md.
  - Author workspaces/AL-5/SUMMARY.md detailing changes across layers.
  - Draft workspaces/AL-5/PULL_REQUEST.md.
  - Verify zero emojis across code, commit messages, comments, and documentation files.

## 4. Testing and Verification Strategy

### Automated Backend Tests
- go vet ./... inside backend/.
- go test -race ./... inside backend/ verifying domain aggregations, repository scanning, HTTP handlers, SPA routing, and embedded filesystem access.
- Dedicated mathematical parity tests verifying all rollup calculations with zero mocks.

### Automated Frontend Tests
- npm run typecheck inside web to enforce strict TypeScript typing with zero any.
- npm run lint inside web to enforce ESLint standards.
- npm test inside web to run Vitest unit and component tests.
- npm run build inside web to produce production bundle.

### Standalone Binary Verification
- Execute ./bin/agentic-lens server --port 8085 --workspaces workspaces in a temporary subshell.
- Issue HTTP requests against /, /assets/..., /api/workspaces, and /api/nonexistent, verifying status codes, headers, and payloads.
- Terminate process gracefully via interrupt signal.

## 5. Invariant Compliance Checklist
- Streamlit vs TypeScript graph parity report authored and saved: Confirmed (workspaces/AL-5/STREAMLIT_GRAPH_PARITY_REPORT.md).
- Go application moved to dedicated backend/ folder: Confirmed.
- Parity verified before removing Python: Confirmed.
- Complete removal of scripts/extract_agentic_metrics.py and all Python artifacts: Confirmed.
- Zero code blocks present in this implementation plan: Confirmed.
- Zero emojis across all plan artifacts, code, comments, and commit messages: Confirmed.
- Inward dependency flow preserved: backend/internal/domain imports no outer packages: Confirmed.
- Single binary self-contained execution with zero Node.js or Python runtime dependency in production: Confirmed.
- Division by zero defended across all calculations: Confirmed.
- Strict typing with zero any or interface{} in Go and TypeScript: Confirmed.
