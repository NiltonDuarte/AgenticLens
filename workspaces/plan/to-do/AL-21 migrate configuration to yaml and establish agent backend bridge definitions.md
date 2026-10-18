h3. Title: AL-21: Migrate Configuration to YAML and Establish Agent-Backend Bridge Definitions

*TARGET*: AgenticLens (adapters | domain | cli | tests)

*CONTEXT*

Currently, configuration is minimally defined in `.agenticlens/config.json` and only used for basic workspace path resolution and git metric patterns. JSON lacks readability and comments, and the backend duplicates hardcoded assumptions (such as ticket prefix regexes, role classifications, database timeouts, and active session windows). Migrating configuration to YAML and enriching it with concrete, structured definitions bridges the gap between agent workflows and the Go backend, allowing both to share consistent operational parameters.

*GOAL*

Migrate `.agenticlens/config.json` to human-readable `.agenticlens/config.yaml` with a backward-compatible Go configuration parser (YAML with JSON fallback), establishing rich configuration definitions for ticket pattern matching (case-insensitive prefixes), database access pragmas, live observability thresholds, agent-to-role mappings, and workspace layouts.

*VERTICAL SLICE SCOPE*

* Domain Core:
  - Create configuration models in `internal/domain/config.go`:
    - `ProjectConfig` (`ProjectName`, `TicketConfig`, `WorkspacesConfig`, `DatabaseConfig`, `LiveConfig`, `CodeMetricsConfig`, `RoleMappings`).
    - `TicketConfig` (`Prefixes []string`, `CaseInsensitive bool`, `CustomPatterns []string`).
    - `LiveConfig` (`ActiveThresholdSeconds int`, `PollIntervalMs int`, `MaxTraceBufferMessages int`, `AnomalyThresholds AnomalyThresholdsConfig`).
    - `DatabaseConfig` (`Path string`, `QueryTimeoutMs int`, `BusyTimeoutMs int`, `ReadOnly bool`).
  - Pure helper to compile regex and normalize ticket IDs: `NormalizeTicketID(raw string, cfg TicketConfig) (string, bool)`.
* Boundary Adapter:
  - Create `internal/adapters/config/loader.go` using a robust Go YAML parser (`gopkg.in/yaml.v3`):
    - Searches for `.agenticlens/config.yaml` or `.agenticlens/config.yml`.
    - Falls back to `.agenticlens/config.json` if YAML is not present.
    - Applies sensible domain defaults when fields are omitted.
  - Update `internal/adapters/repository.go` and `internal/adapters/git/classifier.go` to use the unified configuration loader.
  - Convert `.agenticlens/config.json` to `.agenticlens/config.yaml` in repository root.
* Quality:
  - Go unit tests verifying YAML parsing, JSON fallback, missing file handling, and default value population.
  - Ticket normalization tests asserting case-insensitive matching (e.g. `al-12` -> `AL-12`).
  - Validation that zero `any` types and zero emojis exist across added code.

*ACCEPTANCE CRITERIA*

* `.agenticlens/config.yaml` is created and validated as the primary configuration format.
* The Go backend loads `.agenticlens/config.yaml` and seamlessly falls back to `.agenticlens/config.json` if YAML is absent.
* Ticket pattern matching in configuration supports case-insensitive prefixes (`al-1` matches prefix `AL`).
* Database connection, live observability thresholds, and role mappings are loaded dynamically from configuration.
* All Go tests and static checks pass cleanly (`go vet ./...`, `go test -race ./...`).

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`go vet ./...`, `go test -race ./...`).
* Verified backward-compatible configuration loading with table-driven tests.
* Telemetry tested with `agentic-lens` CLI.
