# feat(config): migrate configuration to yaml, support PO tickets, and bridge definitions (AL-21)

## Overview
Migrates AgenticLens configuration from flat legacy JSON (`.agenticlens/config.json`) to structured YAML (`.agenticlens/config.yaml`), establishing shared agent-to-backend bridge definitions. Provides pure Go domain configuration models, enriched ticket prefix definitions with descriptive purposes, default support for `PO` alongside `AL` and `GEN`, case-insensitive ticket ID normalization, and a centralized configuration loader supporting format prioritization (`.yaml` -> `.yml` -> `.json` -> defaults). Deletes superseded `config.json` and migrates all agent guidelines and skills to YAML.

- Contract: [workspaces/AL-21/CONTRACT.md](workspaces/AL-21/CONTRACT.md)
- Telemetry: [workspaces/AL-21/SUMMARY.md](workspaces/AL-21/SUMMARY.md)

## Summary of Changes

### Domain Core Layer (`backend/internal/domain/`)
- Configuration models: `ProjectConfig`, `TicketConfig`, `TicketPrefix`, `WorkspacesConfig`, `DatabaseConfig`, `AnomalyThresholdsConfig`, `LiveConfig`, `CodeMetricsConfig`, and `RoleMappings`.
- Enriched `TicketPrefix` with `Prefix` and `Description` strings, plus `TicketConfig.PrefixNames()` helper.
- Baseline defaults in `DefaultProjectConfig()` including `AL`, `GEN`, and `PO` prefixes with descriptions.
- `NormalizeTicketID`: pure case-insensitive prefix matching (`al-21` -> `AL-21`, `po-4` -> `PO-4`), regex escaping, token extraction, and length-prioritized prefix matching.
- Pure role mapping with `RoleMappings.ResolveRole` and unit test coverage in `config_test.go`.

### Boundary Adapters Layer (`backend/internal/adapters/`)
- Added `gopkg.in/yaml.v3` dependency.
- Centralized configuration loader in `config/loader.go`:
  - Upward directory traversal (`.yaml` -> `.yml` -> `.json` -> defaults).
  - Flexible deserialization supporting both scalar strings and `{prefix, description}` objects.
  - Legacy root-key compatibility mapping (`ticket_prefix`, `workspaces_dir`, `src_folders`).
- Refactored `adapters.ResolveWorkspacesDir` and `git.LoadConfig` to consume the centralized loader.
- Unit and boundary tests in `config/loader_test.go` and workspace parity verification in `parity_test.go`.

### Repository Configuration & Guidelines (`.agenticlens/`, `.opencode/`, `AGENTS.md`)
- Authored documented `.agenticlens/config.yaml`.
- Deleted superseded `.agenticlens/config.json`.
- Migrated references in `AGENTS.md`, `.opencode/agent/*.md`, and `.opencode/skills/*.md`.
