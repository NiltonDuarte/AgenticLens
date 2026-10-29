# Feature Specification: Per-Developer Runtime Configuration and Multi-Harness Sources

## Recommendation

Add a separate, optional local developer configuration layer. Keep `.agenticlens-config.yaml` as the shared, reviewable project configuration and move machine-specific database locations and installed-runtime selections into an ignored local file. A local file can name Codex and OpenCode sources independently and select a default source, allowing one checkout to work when either or both harnesses are installed.

Use explicit source selection when running ingestion and keep automatic discovery limited to a documented fallback. If multiple sources are configured, AgenticLens should not silently choose one based on whichever database happens to exist. Local values should override shared defaults only for machine-specific settings; CLI flags should take precedence over both. Commit an example local config, but never commit a developer's real paths or credentials.

## Problem Statement

The root `.agenticlens-config.yaml` currently mixes shared project settings with an OpenCode-specific local SQLite path. The path may not exist on another developer's machine, and there is no equivalent first-class configuration for Codex. Developers who have both Codex and OpenCode installed need to switch telemetry sources without editing shared project configuration or relying on implicit environment detection.

## User Stories

- As a developer, I want machine-specific database paths in an ignored local config, so that my setup does not make shared project config machine-dependent.
- As a developer with Codex and OpenCode installed, I want to configure both sources and choose a default or explicitly select one, so that I can ingest the intended telemetry predictably.
- As a project maintainer, I want a committed example and clear precedence rules, so that new contributors can configure their local environment without exposing local paths.

## Functional Scope

### Configuration layers

- Preserve `.agenticlens-config.yaml` for shared project settings and backward-compatible defaults.
- Add an optional, ignored local developer configuration file and a tracked example template.
- Keep local configuration focused on machine-specific values, including per-harness database paths and the selected default source.
- Define deterministic precedence: CLI overrides, then local developer settings, then shared project configuration and built-in defaults.
- Provide actionable errors for an explicitly selected source that is missing or invalid; do not silently switch to another configured source.

### Harness sources

- Model Codex and OpenCode as independently configurable telemetry sources, each with its own path and availability diagnostics.
- Support users configuring either source or both in one checkout.
- Let commands select a source explicitly and use the configured default when the command omits a source.
- Keep existing OpenCode-only configuration functional during migration.
- Document that source selection does not imply merging records across harnesses; a combined scan is only supported if explicitly requested and defined by the command.

### Safety and documentation

- Ignore the actual local developer config in version control and commit only a sanitized example.
- Never place secrets in the local config schema.
- Document setup, source selection, precedence, and behavior when a configured runtime database is absent.

## Acceptance Criteria

1. A developer can configure Codex and OpenCode database paths independently in a local-only config file.
2. The local config can select a default harness, and CLI selection overrides that default.
3. An explicitly selected unavailable source returns an actionable error and never silently falls back.
4. Existing shared configuration and OpenCode database settings continue to work during migration.
5. The actual local config is ignored; a sanitized template and documentation are committed.
6. Tests cover configuration precedence, one-source and two-source setups, absent databases, and backward compatibility.
7. Harness-specific filesystem and database access remains in adapters; domain calculations stay pure.
8. Go and TypeScript remain strictly typed with no `any` or `interface{}`; no emojis are introduced.

## Out of Scope

- Installing, upgrading, or managing Codex or OpenCode runtimes.
- Storing credentials or secrets in configuration.
- Automatically merging telemetry from multiple harnesses into one session history.
- Replacing shared project settings or supporting arbitrary per-user project behavior beyond local runtime settings.

## Open Decisions for Development

- Confirm the repository's existing config discovery and CLI source-selection surfaces before freezing the contract.
- Select the local filename and precedence details consistently with existing conventions; ensure the filename is ignored without ignoring the tracked example.
- Define whether a future explicit multi-source mode is needed separately from selecting one source.
