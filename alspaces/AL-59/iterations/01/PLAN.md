# AL-59 Unified Implementation Plan

## Goal

Deliver local per-harness source configuration, deterministic source selection, and dual-artifact `init` generation while preserving legacy shared OpenCode behavior. Follow the unified frozen boundaries in `../../CONTRACT.md`.

## Implementation sequence

1. **Configuration model and loader**
   - Add typed source settings to the config value model and raw YAML DTOs without adding I/O or validation side effects to the domain.
   - Extend adapter config loading to read optional `.agenticlens-config.local.yaml` beside the discovered shared config, merge only local telemetry source settings, validate source identifiers/default values and non-empty configured paths, and report parse/validation errors with the local filename.
   - Retain existing `database.path` unchanged as the legacy OpenCode source path. Resolve actual defaults and home expansion at the adapter/CLI boundary.
2. **Source/path resolution and CLI wiring**
   - Introduce a small typed adapter-side resolver for source choice and source-specific location precedence; keep it deterministic by passing parsed config and CLI inputs explicitly.
   - Change single-source `generate` parsing so omitted `--source` is distinguishable from explicit `opencode`; apply CLI source, local default, and built-in default in contract order.
   - Resolve `--db-path` and `--codex-home` before invoking their respective adapters, preserving explicit path flags and existing environment/default behavior.
   - Review `backfill` source behavior: preserve required explicit `both` support; where a single source is accepted, apply the configured default only when the CLI source is omitted if compatible with current command contract. A selected source's missing path must produce an error, never a cross-source retry.
   - Keep server/live OpenCode behavior compatible; use the resolved OpenCode path when consuming source config, without adding a dashboard source selector.
3. **Local artifacts and documentation**
   - Ignore `.agenticlens-config.local.yaml` and track `.agenticlens-config.local.example.yaml` with placeholder paths for both harnesses and an example default.
   - Extend `agentic-lens init` to generate the shared `.agenticlens-config.yaml` and distinct local `.agenticlens-config.local.yaml`. Keep existing shared-config fields, `--output` compatibility, and flag behavior. Local output contains only `telemetry.default_source` and independently configurable OpenCode `database_path` and Codex `home`.
   - Use sanitized portable local defaults: `opencode`, `~/.local/share/opencode/opencode.db`, and `~/.codex`. Do not probe source availability or serialize an expanded user home path.
   - Apply overwrite protection consistently to both outputs: `--force` governs both; without it, protect either existing file. Reject `--output` when it collides with the local-config filename. Label both YAML artifacts clearly during `--dry-run` and write neither. Report artifact-specific write errors, avoid success output until both writes succeed, and prevent or clean up a misleading partial result (preflight and stage temporary outputs before replacement).
   - Update `.agents/skills/agentic-lens-setup/SKILL.md` to explain dual-file generation, guide review of default source and paths, verify the local file is ignored before commit, distinguish the Codex rollout home from OpenCode SQLite, and retain credential-safety guidance.
   - Update README and setup manual with local setup, path meanings (OpenCode SQLite path vs Codex home), precedence, independent configurations, source flag behavior, migration compatibility, safe missing-source failures, and the prohibition on credentials.
4. **Tests**
   - Add table-driven config tests for absent local file, valid local settings, malformed local YAML, invalid harness/default values, and merging local settings without changing shared project settings.
   - Cover Codex-only, OpenCode-only, and both-source configurations; assert source/path precedence including CLI override, local source path, legacy shared OpenCode path, and adapter built-in fallback.
   - Cover explicit unavailable Codex/OpenCode locations and invalid database paths; assert actionable error and no alternate source selection.
   - Add CLI-level tests proving omitted source uses the configured local default, explicit `--source` overrides it, legacy OpenCode shared config remains usable, and `backfill both` remains explicit multi-source behavior.
   - Keep tests deterministic with temporary filesystem fixtures and real temporary SQLite where appropriate; do not add domain mocks or change metric/domain test scope.
   - Add CLI `init` tests for both files, sanitized defaults and absence of absolute-home leakage, clearly labeled write-free dry-run, existing-file protection and `--force`, shared-output collision, and actionable write failures without misleading success or partial state.
   - Retain all config, resolver, source-selection, and failure tests above; cover source independence and migration compatibility alongside init behavior.

## Expected change areas

- `backend/internal/domain/config.go`
- `backend/internal/adapters/config/loader.go` and config loader tests
- Focused source/path resolver under `backend/internal/adapters/config/` or adjacent adapter package, with unit tests
- `backend/cmd/agentic-lens/main.go` and command tests
- `backend/internal/adapters/backfill/` and focused tests only if required for source resolution
- `.gitignore`, `.agenticlens-config.local.example.yaml`
- `.agents/skills/agentic-lens-setup/SKILL.md`, `README.md`, `docs/SETUP_MANUAL.md`

## Iteration acceptance checks

- No code path silently changes the selected harness after a path/open/discovery failure.
- Existing shared OpenCode configuration works with no local file.
- Explicit source and path flags win over local values; local values win over shared OpenCode path; built-in defaults apply only when no higher-precedence value exists.
- Local configuration is optional, ignored, and has a sanitized tracked example.
- `init` generates both distinct artifacts using sanitized home-relative local defaults without checking whether either source exists; dry-run, overwrite, collision, and write failures are clear and safe.
- Setup skill, README, and setup manual consistently explain dual-file setup, path meanings, local-file review, ignore verification, and credential safety.
- Source selection and I/O remain at adapter/CLI boundaries; domain remains pure.
- Focused tests and full `make` checks pass before delivery.

## Risks and decisions to verify during implementation

- The specification describes a Codex “database path,” but the existing Codex adapter consumes a CODEX_HOME rollout tree. Preserve the real adapter contract and document this distinction; do not invent a Codex SQLite source.
- Existing `generate` currently declares OpenCode as the flag's default. Implementation must distinguish omitted from explicitly supplied `--source opencode` to honor local default precedence.
- Existing config loading requires shared project fields. Local-only source configuration supplements rather than replaces the required shared project file.
- Backfill's `both` mode is an explicit multi-source operation and must not be conflated with a single default source.
