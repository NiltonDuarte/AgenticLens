h3. Title: AL2-3: Multi-Project Configuration Schema, Domain Aggregations, and Multi-Folder Scanner

*TARGET*: AgenticLens (domain | adapters | cli | config | tests)

*CONTEXT*
 
Currently, AgenticLens only scans a single workspace directory and reads a single root configuration manifest. Engineering leads managing multiple services or teams need a single AgenticLens instance to monitor and aggregate telemetry from multiple project directories. The platform requires a multi-project configuration schema, domain models tagged by project, project-level filtering, and a scanner capable of traversing multiple distinct project folders while respecting each project's individual configuration.

*GOAL*
 
Implement the multi-project configuration schema in `.agenticlens-config.yaml` and CLI flags, tag all ingested ledgers with `ProjectID` and `ProjectName`, extend domain aggregations to support multi-project rollups and filtering, and implement a multi-folder repository scanner that loads per-project configuration manifests.

*VERTICAL SLICE SCOPE*
 
* Domain Core: Define `ProjectDefinition` (`ID`, `Name`, `SourceType`, `SourcePath`, `WorkspacesDir`, `ConfigPath`), `ProjectConfig.Projects []ProjectDefinition`. Extend `AgenticMetricsLedger` and `ScanReport` with `ProjectID string` and `ProjectName string`. Implement pure `FilterLedgersByProjects(ledgers []AgenticMetricsLedger, projectIDs []string) []AgenticMetricsLedger` and `CalculateMultiProjectSummary(ledgers []AgenticMetricsLedger) MultiProjectSummary`. Pure functions, zero I/O imports.
* Boundary Adapter: Extend `internal/adapters/config/loader.go` to parse the `projects:` array from `.agenticlens-config.yaml`. Implement `ScanMultiProjectWorkspaces(projects []ProjectDefinition) MultiProjectScanReport` in `internal/adapters/repository.go` that iterates configured project paths, loads each project's specific `.agenticlens-config.yaml` for ticket prefixes and role mappings, and extracts workspaces.
* Presentation: Add repeatable `--project <id>=<path>` flags and `--projects-config <path>` to `cmd/agentic-lens` for `scan` and `server` commands. Extend CLI table formatter to display project origin column when multiple projects are scanned.
* Quality: Table-driven domain unit tests in `internal/domain` verifying multi-project filtering and summary rollups. Adapter tests in `internal/adapters` using `t.TempDir()` with multiple project directories each containing distinct configs and ticket prefixes, asserting accurate configuration scoping and ledger tagging.

*ACCEPTANCE CRITERIA*
 
* `.agenticlens-config.yaml` supports defining an array of `projects:` with custom IDs, names, and workspace directories.
* CLI commands accept `--project core=./workspaces --project billing=/path/to/billing` alongside config definitions.
* The scanner loads each project's own `.agenticlens-config.yaml` and applies its ticket prefixes and role mappings to its ledgers.
* Domain rollups accurately compute individual project metrics and aggregate cross-project totals.
* All Go tests pass cleanly with zero mocks in domain logic (`go test -race ./...`).

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
