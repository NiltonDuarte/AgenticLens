# Contract: AL-21 Migrate Configuration to YAML and Establish Agent-Backend Bridge Definitions

## 1. Problem Statement

Configuration in AgenticLens is currently constrained to `.agenticlens/config.json` with minimal keys (`project_name`, `ticket_prefix`, `workspaces_dir`, `src_folders`). This JSON-centric approach exhibits several critical deficiencies:
1. JSON lacks comments, preventing developers and agent operators from documenting operational thresholds, reasoning parameters, and directory conventions in-place.
2. The Go backend and agent orchestration runtimes duplicate hardcoded assumptions across disparate files:
   - Ticket prefixes and casing regexes are duplicated across CLI flags, git commit scrapers, and database inference heuristics.
   - Agent role deduplication and role assignment (`planner`, `coder`, `tester`, `reviewer`, `orchestrator`, `build`) are hardcoded directly in `backend/internal/adapters/opencode/extractor.go`.
   - SQLite pragmas, busy timeouts (`5000ms`), and query timeouts are hardcoded without configuration hooks.
   - Live session monitoring windows (`300s`) and anomaly detection thresholds are scattered as literal constants.
3. Multiple adapters (`backend/internal/adapters/repository.go` and `backend/internal/adapters/git/classifier.go`) each perform bespoke upward directory traversals to parse `.agenticlens/config.json` into unshared local private structs (`configDTO`, `rootConfig`).

Migrating configuration to `.agenticlens/config.yaml` and establishing structured configuration definitions bridges the gap between agent workflows and the Go backend runtime. The system must support human-readable YAML with inline documentation while guaranteeing seamless backward compatibility via JSON fallback and resilient default value generation when files or fields are omitted.

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All implementation must strictly comply with `TECHNICAL_GUIDELINES.md` and `AGENTS.md`.

### Inward Dependency Flow
Dependencies flow strictly inward across layers:
1. Application and CLI entrypoints (`backend/cmd/agentic-lens/main.go`) invoke adapter loaders to acquire domain configuration models and pass them inward to domain algorithms or adapter constructors.
2. Boundary Adapter Layer (`backend/internal/adapters/config/`, `backend/internal/adapters/git/`, `backend/internal/adapters/repository.go`):
   - `backend/internal/adapters/config/loader.go` executes filesystem lookups, directory traversals, format detection, YAML/JSON deserialization, and merges partial overrides on top of pure domain defaults.
   - Adapters (`repository.go`, `classifier.go`, `opencode/client.go`) consume pure domain configuration models and never re-implement custom parsing or file reading loops.
3. Functional Core Domain Layer (`backend/internal/domain/config.go`):
   - Houses pure configuration models (`ProjectConfig`, `TicketConfig`, `WorkspacesConfig`, `DatabaseConfig`, `LiveConfig`, `AnomalyThresholdsConfig`, `CodeMetricsConfig`, `RoleMappings`).
   - Contains pure ticket normalization and regex compilation logic: `NormalizeTicketID(raw string, cfg TicketConfig) (string, bool)`.
   - Contains pure role resolution methods: `ResolveRole(rawAgent string) AgentRole`.
   - Absolute isolation: zero imports of `os`, `io`, `net/http`, `database/sql`, filesystem traversal APIs, or external YAML/JSON libraries.

### Strong Typing and Zero Any Invariant
1. Zero usage of `any` or `interface{}` across all Go packages. Concrete structs, string aliases, and typed errors only.
2. Strict value semantics: domain configuration models are immutable structs.

### Zero Emojis Mandate
No emojis permitted in code, comments, docstrings, configuration YAML/JSON files, log messages, test outputs, or workspace artifacts.

## 3. Scope Boundaries

### In Scope for AL-21
1. Pure Domain Configuration Core (`backend/internal/domain/config.go`):
   - Definition of `ProjectConfig` and child structs: `TicketConfig`, `WorkspacesConfig`, `DatabaseConfig`, `LiveConfig`, `AnomalyThresholdsConfig`, `CodeMetricsConfig`, and `RoleMappings`.
   - Implementation of pure function `NormalizeTicketID(raw string, cfg TicketConfig) (string, bool)`.
   - Implementation of pure method `RoleMappings.ResolveRole(rawAgent string) AgentRole`.
   - Definition of domain configuration constructor `DefaultProjectConfig() ProjectConfig` returning baseline domain defaults.
2. Unified Configuration Loader Adapter (`backend/internal/adapters/config/loader.go`):
   - External dependency integration of `gopkg.in/yaml.v3` in `backend/go.mod`.
   - Upward directory traversal searching in order: `.agenticlens/config.yaml`, `.agenticlens/config.yml`, `.agenticlens/config.json`.
   - Deserialization supporting both YAML (via `gopkg.in/yaml.v3`) and JSON (via standard `encoding/json`).
   - Sane default application: missing fields or absent configuration files automatically populate with `DefaultProjectConfig()`.
   - Safe path resolution helpers for configuration file location.
3. Adapter Integration Refactoring:
   - Refactor `backend/internal/adapters/repository.go` (`ResolveWorkspacesDir`) to utilize the centralized loader.
   - Refactor `backend/internal/adapters/git/classifier.go` (`LoadConfig`) to delegate to the centralized loader while maintaining public signature compatibility.
4. Repository Configuration Migration:
   - Author `.agenticlens/config.yaml` in repository root with comprehensive explanatory comments and active configuration sections.
   - Retain `.agenticlens/config.json` as a backward-compatibility verification target or cleanly supersede it while validating fallback.
5. Automated Test Suite:
   - Pure domain unit tests in `backend/internal/domain/config_test.go` covering `NormalizeTicketID` and role mapping resolution across edge cases.
   - Table-driven adapter tests in `backend/internal/adapters/config/loader_test.go` testing YAML parsing, JSON fallback, missing file defaults, malformed file error reporting, and partial configuration merging.
   - Regression testing for `repository.go` and `classifier.go`.

### Out of Scope for AL-21
1. Implementation of the live SQLite hook and session tree engine (governed by AL-17).
2. Implementation of SSE streaming endpoints or live HTTP routes (governed by AL-17 and AL-18).
3. Live trace inspector and visual waterfall presentation in web (governed by AL-18 and AL-20).
4. Direct modification of OpenCode database extraction queries in `backend/internal/adapters/opencode/extractor.go` beyond role mapping consumption interfaces.

## 4. Architectural Contracts and Data Structures

### Domain Configuration Models (`backend/internal/domain/config.go`)

- `ProjectConfig`: consolidated configuration model containing `ProjectName string`, `Ticket TicketConfig`, `Workspaces WorkspacesConfig`, `Database DatabaseConfig`, `Live LiveConfig`, `CodeMetrics CodeMetricsConfig`, and `RoleMappings RoleMappings`.
- `TicketConfig`: struct containing `Prefixes []string`, `CaseInsensitive bool`, and `CustomPatterns []string`.
- `WorkspacesConfig`: struct containing `Dir string` and `SrcFolders []string`.
- `DatabaseConfig`: struct containing `Path string`, `QueryTimeoutMs int`, `BusyTimeoutMs int`, and `ReadOnly bool`.
- `AnomalyThresholdsConfig`: struct containing `StallThresholdSeconds int`, `ErrorRateThreshold float64`, and `ContextExhaustionPct float64`.
- `LiveConfig`: struct containing `ActiveThresholdSeconds int`, `PollIntervalMs int`, `MaxTraceBufferMessages int`, and `AnomalyThresholds AnomalyThresholdsConfig`.
- `CodeMetricsConfig`: struct containing `Include []string`, `Exclude []string`, `TestPatterns []string`, and `ProdPatterns []string`.
- `RoleMappings`: type map of `string` to `string`. Method `ResolveRole(rawAgent string) AgentRole` converts agent persona strings into canonical domain roles (`orchestrator`, `planner`, `coder`, `tester`, `reviewer`).
- `NormalizeTicketID(raw string, cfg TicketConfig) (string, bool)`: pure deterministic ticket normalization helper operating on immutable inputs without disk or environment access.

### Domain Defaults Specification (`DefaultProjectConfig`)

When a configuration file is missing or contains partial keys, the following defaults are guaranteed:
- `ProjectName`: `"AgenticLens"`
- `Ticket.Prefixes`: `["AL", "GEN"]`
- `Ticket.CaseInsensitive`: `true`
- `Ticket.CustomPatterns`: `[]string{}`
- `Workspaces.Dir`: `"workspaces"`
- `Workspaces.SrcFolders`: `[]string{"web", "backend"}`
- `Database.Path`: `"~/.local/share/opencode/opencode.db"`
- `Database.QueryTimeoutMs`: `5000`
- `Database.BusyTimeoutMs`: `5000`
- `Database.ReadOnly`: `true`
- `Live.ActiveThresholdSeconds`: `300`
- `Live.PollIntervalMs`: `2000`
- `Live.MaxTraceBufferMessages`: `500`
- `Live.AnomalyThresholds.StallThresholdSeconds`: `120`
- `Live.AnomalyThresholds.ErrorRateThreshold`: `0.25`
- `Live.AnomalyThresholds.ContextExhaustionPct`: `0.85`
- `CodeMetrics`: all slice fields initialized as empty slices
- `RoleMappings`: baseline mapping covering `"orchestrator"` -> `"orchestrator"`, `"build"` -> `"orchestrator"`, `"planner"` -> `"planner"`, `"coder"` -> `"coder"`, `"tester"` -> `"tester"`, `"reviewer"` -> `"reviewer"`.

### Centralized Loader Adapter Contract (`backend/internal/adapters/config/loader.go`)

- `FindConfigFile(startDir string) (string, error)`: traverses directory parents to locate `.agenticlens/config.yaml`, `.agenticlens/config.yml`, or `.agenticlens/config.json`.
- `Load(repoRoot string) (domain.ProjectConfig, error)`: discovers and parses project configuration, merging values on top of domain defaults.
- `LoadFromFile(filePath string) (domain.ProjectConfig, error)`: decodes specific file based on file extension (`.yaml`/`.yml` via `yaml.v3`, `.json` via `json.Unmarshal`).

## 5. Backward Compatibility Guarantees

1. JSON Fallback:
   If `.agenticlens/config.yaml` and `.agenticlens/config.yml` do not exist, the loader automatically discovers and deserializes `.agenticlens/config.json`.
2. Missing File Resilience:
   If no configuration file exists anywhere in the directory tree, `Load(repoRoot)` returns `DefaultProjectConfig()` and a `nil` error, ensuring the CLI and HTTP server continue running without failure.
3. Legacy Key Mapping:
   In legacy `.agenticlens/config.json`, keys exist at the root: `"ticket_prefix": "AL"`, `"workspaces_dir": "workspaces"`, `"src_folders": ["web", "backend"]`. The loader must parse these legacy root keys into their corresponding structured domain locations (`cfg.Ticket.Prefixes = []string{legacyPrefix}`, `cfg.Workspaces.Dir = legacyWorkspacesDir`, `cfg.Workspaces.SrcFolders = legacySrcFolders`).
4. Adapter Call-Site Compatibility:
   Existing functions `git.LoadConfig(repoRoot string) (*git.ClassificationConfig, error)` and `adapters.ResolveWorkspacesDir(overridePath string) string` must remain functioning public symbols, delegating internally to the new centralized config loader.

## 6. Acceptance Criteria

1. `.agenticlens/config.yaml` is authored in the repository root with comprehensive comments and valid YAML structure covering project, ticket, workspaces, database, live, code metrics, and role mapping sections.
2. Go dependency `gopkg.in/yaml.v3` is registered in `backend/go.mod` and compiles cleanly under `CGO_ENABLED=0`.
3. `backend/internal/domain/config.go` houses pure domain models, `DefaultProjectConfig()`, `NormalizeTicketID()`, and `RoleMappings.ResolveRole()` with zero I/O imports.
4. `backend/internal/adapters/config/loader.go` provides centralized loading with priority: YAML -> YML -> JSON -> Defaults.
5. Legacy root keys in JSON are mapped to nested domain config structures.
6. Case-insensitive ticket normalization is verified via table-driven tests (e.g. `al-12` -> `AL-12`, `GEN-42` -> `GEN-42`, `feat/al-9-trace` -> `AL-9`, invalid strings -> `("", false)`).
7. `backend/internal/adapters/repository.go` and `backend/internal/adapters/git/classifier.go` successfully consume configuration through the unified loader.
8. Zero `any` or `interface{}` types exist across all added/modified Go code.
9. Zero emojis exist in code, comments, logs, or configuration files.
10. All checks pass cleanly via `make check` and `make test`.
