# Contract: AL-5 Single Binary Integration Parity and Deprecation

## 1. Problem Statement
AgenticLens previously relied on a Python Streamlit application for telemetry visualization, analytics views, and local dashboard serving. That legacy architecture introduced heavy runtime dependencies (Python 3.14, Streamlit, Pandas, Altair), high process initialization latency, and complex environment configuration. 

Previous tickets delivered the pure Go functional core (AL-1), Go HTTP API services and SPA serving adapter (AL-2), modern TypeScript dashboard views (AL-3, AL-4), native telemetry extraction (AL-11), and telemetry backfill utilities (AL-12). However, three architectural cleanup tasks remain:
1. The Go codebase currently resides in the project root (`cmd/`, `internal/`, `go.mod`, `go.sum`), while the frontend is compartmentalized under `web/`. The Go application must be relocated to a dedicated directory (`backend/`).
2. Legacy Python code and artifacts (`agentic_lens/`, `scripts/extract_agentic_metrics.py`, `tests/`, `pyproject.toml`, `uv.lock`) remain in the repository.
3. Cross-engine numerical and categorical parity between the Go domain engine and the legacy Python aggregations must be verified before purging Python, and missing Streamlit visualizations must be audited and cataloged.

The purpose of AL-5 is to achieve full architectural and integration closure:
- Audit and document all visualizations present in Streamlit that are missing in the TypeScript web frontend (recorded in `workspaces/AL-5/STREAMLIT_GRAPH_PARITY_REPORT.md`).
- Relocate the entire Go application to a dedicated `backend/` directory (`backend/cmd/`, `backend/internal/`, `backend/go.mod`, `backend/go.sum`).
- Automate single-binary production bundling across npm and Go, producing a standalone executable at `bin/agentic-lens` with embedded web assets.
- Prove 100% numerical and categorical parity between Go domain aggregations and Python outputs on real and synthetic fixtures.
- Following parity confirmation, completely eliminate all Python artifacts (`scripts/extract_agentic_metrics.py`, `agentic_lens/`, `tests/`, `pyproject.toml`, `uv.lock`), resulting in a repository requiring zero Python runtime and zero Node.js runtime for production execution.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All changes must strictly adhere to TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. Go domain logic in `backend/internal/domain` has zero external dependencies, no file or network I/O, no system clock calls, and imports no outer packages.
2. Adapters in `backend/internal/adapters` handle file scanning, JSON marshaling, HTTP routing, and embedded asset retrieval via `io/fs.FS`.
3. Server and CLI entrypoints in `backend/cmd/agentic-lens` coordinate configuration and process lifecycle.
4. TypeScript presentation in `web/src` depends on `web/src/domain` and `web/src/api`, compiling to pure static assets with zero Go runtime knowledge.
5. Single-binary packaging embeds `web/dist` into `backend/internal/adapters/web/dist` via `embed.FS`, completely isolated from runtime execution environments.

### Pure Domain Core Isolation
1. Parity validation must confirm that Go `backend/internal/domain` functions produce exact numerical, categorical, and structural equivalents to the original Python domain definitions.
2. Mathematical rollups (costs, tokens, cache hit rates, saturation percentages, turn velocities, test-to-code ratios) must defend against division by zero across all edge-case ledgers.
3. All inputs to domain functions are supplied as immutable value arguments; all outputs are returned as concrete structs or typed errors.

### Dedicated Backend Folder Relocation Invariants
1. The Go module root is relocated to `backend/`:
   - `cmd/` -> `backend/cmd/`
   - `internal/` -> `backend/internal/`
   - `go.mod` -> `backend/go.mod`
   - `go.sum` -> `backend/go.sum`
2. The root `Makefile` must coordinate commands across `web/` and `backend/`. The compiled binary destination remains at the project root `bin/agentic-lens` to ensure seamless execution for CLI commands, telemetry generation, and developers.
3. Configuration in `.agenticlens/config.json` must update `src_folders` to `["web", "backend"]`.

### Imperative Boundary Shell and Asset Embedding
1. Package `backend/internal/adapters/web` wraps the embedded static filesystem via `embed.FS`. The exported file tree must be rooted at the distribution root so that `index.html` and assets reside at predictable relative paths.
2. The HTTP SPA handler must intercept all non-API paths, serving exact static assets when present on disk/embed, and falling back to `index.html` for client-side navigation. API routes under `/api` must never fall back to `index.html` on missing endpoints; they must return structured 404 JSON errors.

### Complete Python Purge Invariant
1. Cross-engine parity is executed and verified prior to artifact deletion.
2. Once parity is verified, all Python artifacts are deleted:
   - `agentic_lens/` (entire presentation, domain, adapter tree)
   - `scripts/extract_agentic_metrics.py` (superseded by native Go `agentic-lens generate` from AL-11)
   - `tests/` (Python tests)
   - `pyproject.toml` and `uv.lock`
   - Any `.pytest_cache`, `__pycache__`, or `.venv` directories.
3. The repository becomes a clean dual-tier Go backend + TypeScript frontend architecture.
4. Zero emojis are permitted in build logs, console output, Markdown tables, UI labels, or code comments.

---

## 3. Scope Boundaries: Go Backend, TypeScript Frontend, and Python Decommissioning

### In Scope for AL-5
1. Streamlit Visualization Gap Audit: Scan the Python Streamlit app for graphs missing in the TypeScript web implementation and output the comprehensive audit report to `workspaces/AL-5/STREAMLIT_GRAPH_PARITY_REPORT.md`.
2. Go Codebase Relocation:
   - Move `cmd/` and `internal/` to `backend/cmd/` and `backend/internal/`.
   - Move `go.mod` and `go.sum` to `backend/`.
   - Update root `Makefile`, `Dockerfile`, and `.agenticlens/config.json`.
   - Ensure `bin/agentic-lens` builds to root `./bin/agentic-lens`.
3. Cross-Engine Parity Verification:
   - Implement an automated parity test suite comparing Go domain calculations against Python reference outputs on existing workspace ledgers (AL-1, AL-2, AL-11, AL-12, PO-1) and synthetic edge-case fixtures.
   - Assert 100% numerical and categorical equivalence across total tokens, prompt cache hit rates, USD costs, context window health classifications, turn velocity, test-to-production ratios, and tool execution health.
4. Complete Python Purge:
   - Remove `scripts/extract_agentic_metrics.py`.
   - Remove `agentic_lens/` presentation, domain, and adapter code.
   - Remove `tests/` Python test suite.
   - Remove `pyproject.toml` and `uv.lock`.
5. Embedded Static Asset Wiring and Single Binary Packaging:
   - Wire `npm run build` from `web/` to sync into `backend/internal/adapters/web/dist`.
   - Compile standalone executable at `bin/agentic-lens`.
   - Verify `./bin/agentic-lens server` opens and serves the SPA without Node.js or Python.
6. Multi-Stage Dockerfile Modernization:
   - Multi-stage build (Node build -> Go build -> minimal Alpine/scratch final image).
   - Container has zero Python and zero Node.js runtime dependencies.

### Out of Scope for AL-5
1. Re-implementing the full Analytics Studio ad-hoc query engine in TypeScript (identified as a candidate for a future dedicated ticket in `STREAMLIT_GRAPH_PARITY_REPORT.md`).
2. Introducing new metric formulas or modifying existing health threshold constants.

---

## 4. Architectural Contracts and Parity Invariants

### Build and Packaging Contract
1. Binary Destination: The build target must produce a standalone executable at `bin/agentic-lens`.
2. Dependency Independence: The compiled binary must run without `node`, `npm`, `python`, `uv`, or external web servers installed on the host.
3. Fresh Assets Guarantee: Building via `make build` coordinates `web` build and `backend` build.
4. Fallback Safety: The `backend/internal/adapters/web` package embeds a valid `dist` filesystem so that Go developers without Node.js can build the Go binary using pre-committed or cached dist assets.

### Cross-Engine Parity Mathematical and Categorical Contract
The Go domain engine and the reference Python engine must agree on the following calculations given identical `AgenticFactLedger` inputs:
1. Token Metrics:
   - Total Input Tokens: Sum of all invocation input tokens.
   - Total Output Tokens: Sum of all invocation output tokens.
   - Total Cache Read Tokens: Sum of all invocation cache read tokens.
   - Prompt Cache Hit Rate: Cache read tokens divided by sum of input and cache read tokens, expressed as a percentage rounded to one decimal place. Zero when denominator is zero.
   - Peak Context Tokens: Maximum peak context across all invocations.
2. Financial Cost Metrics:
   - Total Cost USD: Sum of all invocation costs, rounded to four decimal places.
   - Role Cost Breakdown: Sum of costs per role, ordered canonically.
   - Model Cost Breakdown: Sum of costs per model, sorted descending by cost.
   - Subagent Role Effort Distribution: Percentage of subagent cost incurred by each subagent role, rounded to one decimal place. Zero if subagent total cost is zero.
3. Context Window Health Scoring:
   - Healthy Ceiling: Fixed at 200,000 tokens.
   - Status Mapping: Peak tokens under 100,000 maps to HEALTHY; 100,000 to 160,000 maps to ELEVATED; 160,001 to 200,000 maps to BLOATED; greater than 200,000 maps to CRITICAL.
   - Orchestrator Saturation: Orchestrator peak context tokens divided by 200,000, formatted as percentage rounded to one decimal place.
   - Subagent Peak Context: Maximum peak context among non-orchestrator invocations, paired with the role name that produced it.
4. SDLC and Efficiency Metrics:
   - Active Duration: Cumulative duration in seconds across all invocations. Formatted as Xh Ym when 60 minutes or greater; Xm Ys when under 60 minutes.
   - Subagent Turn Velocity: Average turns per subagent invocation rounded to one decimal place; peak turns among subagents; average turns per subagent role.
   - Test-to-Production Ratio: Net test lines added divided by net production lines added, formatted as X.Y:1. Special cases: 1:0 when production lines added is zero and test lines added is positive; 0:1 when test lines added is zero and production lines added is positive; 0:0 when both are zero.
   - Tool Execution Health: Total calls, error corrections, and success rate percentage rounded to one decimal place. 100.0% when total calls is zero.
   - Reasoning Token Ratio: Reasoning tokens divided by total output tokens, expressed as percentage rounded to one decimal place.

### Post-Parity Deletion Contract
1. Python Verification First: The parity test suite must execute and pass before deleting Python source code.
2. Complete Purge: All Python files, manifests, and virtual environment configs are deleted from Git tracking.
3. Verification of Zero Regressions: After deletion, `make check`, `make test`, and `make build` must all pass cleanly with no dangling references or broken paths.
