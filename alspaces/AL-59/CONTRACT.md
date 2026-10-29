# AL-59 Technical Contract

## Objective

Move machine-specific telemetry source selection into an optional ignored local developer configuration while retaining shared project configuration compatibility. Support independent OpenCode and Codex locations, an optional default harness, and explicit CLI selection. Selecting a source must never silently route to another harness when the selected source is unavailable.

## Grounded repository boundaries

- `backend/internal/domain/config.go` owns value-shaped project configuration data and defaults. Any added source configuration values remain data only; no path probing, environment reads, filesystem access, or database access belongs in the domain.
- `backend/internal/adapters/config/loader.go` owns configuration file discovery, parsing, merging, validation, and local/shared precedence because these are filesystem and serialization concerns.
- `backend/internal/adapters/opencode/` owns SQLite opening and OpenCode queries. `backend/internal/adapters/codex/` owns Codex home/session rollout discovery and parsing. Neither harness adapter may select or fall back to the other harness.
- `backend/cmd/agentic-lens/main.go` owns CLI flag resolution and source dispatch. Existing `generate --source` currently defaults to OpenCode, OpenCode uses `--db-path`, Codex uses `--codex-home` / `CODEX_HOME`, and `backfill` supports explicit source selection including `both`.
- `backend/internal/adapters/backfill/` owns backfill orchestration and adapter calls. Its multi-source `both` behavior remains distinct from selecting one default harness.
- `docs/SETUP_MANUAL.md` and `README.md` document current source behavior; `.gitignore` excludes database files but has no local AgenticLens configuration entry yet.

## Configuration contract

Use `.agenticlens-config.local.yaml` as the optional machine-local file, discovered from the repository root alongside `.agenticlens-config.yaml`. Add it to `.gitignore`; track `.agenticlens-config.local.example.yaml` containing only sanitized placeholders and no credentials. The local file is optional; a missing file is not an error. A present but malformed or invalid local file is an actionable configuration error rather than silently ignored.

The local schema has `telemetry.default_source`, `telemetry.sources.opencode.database_path`, and `telemetry.sources.codex.home`. Codex currently reads rollout files under a Codex home directory; it does not consume a SQLite database. Accordingly, `home` represents the Codex source location and the ticket's “database path” wording must not introduce a fictitious Codex database adapter. The OpenCode local database path is `database_path`.

For a single-harness command, resolve source choice in this order: explicit `--source`, local `telemetry.default_source`, shared source setting if one exists in a compatible future or current config, then built-in default `opencode`. For a selected source location, resolve CLI-specific path override (`--db-path` or `--codex-home`), local harness path, legacy shared `database.path` for OpenCode, then that harness adapter's established built-in path (`~/.local/share/opencode/opencode.db` for OpenCode; `$CODEX_HOME` then `~/.codex` for Codex). No OpenCode value is interpreted as a Codex location. Expand supported home notation at the imperative boundary.

The current shared `.agenticlens-config.yaml` has `database.path: ~/.local/share/opencode/opencode.db`; preserve its behavior as the OpenCode fallback during migration. Do not require local config for existing installations. Do not move unrelated shared settings or alter config initialization defaults as part of this ticket.

## Selection and failure semantics

- An omitted `--source` uses the configured default according to precedence; explicit `--source` always wins.
- `--source` accepts only supported single harnesses for `generate`. Preserve `backfill`'s explicit `both`/`all` semantics; a single-source default must not silently expand to both.
- Once selected, failure to find, read, or open that harness's configured source returns a clear error that names the selected harness and actionable path/config remediation. It must not retry another configured source.
- OpenCode paths remain read-only SQLite inputs. Codex paths remain rollout directories. Existing session identity requirements such as Codex `--nonce` or `--session-id` remain unchanged.
- Existing shared OpenCode-only checkouts continue to work without local config. Codex-only and dual-source local configurations work independently.

## Scope and exclusions

In scope: typed source config data; local config loading and merge/validation; ignored local file and sanitized example; source/path resolution in telemetry CLI and relevant backfill entry points; adapter-boundary failure clarity; setup documentation; focused Go tests for precedence, source combinations, missing/invalid selections, and migration compatibility.

Out of scope: domain metric or ledger schema changes; database writes; credential storage; automatic source probing/fallback; dashboard source selector; changing telemetry calculations; unrelated CLI commands; treating Codex rollouts as SQLite.

## Acceptance boundary

The implementation is acceptable when each source can be configured independently in local-only settings; explicit CLI source and path overrides take precedence; omitted source honors local default then compatibility/default rules; unavailable selected sources fail clearly without fallback; legacy shared OpenCode config still works; the local file is ignored and a sanitized example is tracked; and tests cover precedence, Codex-only/OpenCode-only/both settings, failure cases, and legacy config. Domain calculations remain pure and Go remains statically typed without `any` or `interface{}`.

## Verification scope

Add deterministic tests at the config loader/resolver, CLI selection boundary, and adapter error boundary. Use temporary directories and SQLite fixtures where needed; do not use domain mocks. Run focused Go package tests during implementation, then the repository's standard `make` quality gates for delivery as required by the ticket's definition of done.

## Init generation and setup guidance

### Init command behavior

- `agentic-lens init` must generate both the shared project configuration (`.agenticlens-config.yaml`, subject to the existing `--output` behavior) and the machine-local source configuration (`.agenticlens-config.local.yaml`).
- The local output must be kept distinct from the shared output. If the caller explicitly directs `--output` to the local-config filename, `init` must fail clearly rather than overwrite one artifact with the other or silently change the shared output destination.
- Preserve existing shared-config fields and flag behavior for compatibility. `--dry-run` must display both generated artifacts with clear filename labels and separately identifiable valid YAML, without writing either file; `--force` must consistently govern overwrites of both generated files. Without `--force`, protect either pre-existing destination.
- The generated local file must contain only telemetry source settings: default source and independently configurable OpenCode database path and Codex home. It must not duplicate shared project settings or contain credentials.

### Local file defaults, privacy, and ignore behavior

- Generated defaults must be sanitized and portable: use home-relative paths (`~/.local/share/opencode/opencode.db` and `~/.codex`) rather than embedding the invoking user's absolute home path. Default the selected harness to `opencode` for compatibility; do not assert that either source exists during generation.
- The generated `.agenticlens-config.local.yaml` must be ignored by Git. Preserve the tracked sanitized example and ensure it has no user-specific absolute paths, credentials, or other machine secrets.
- `init` must not silently overwrite a pre-existing local file unless `--force` is supplied. A shared `--output` destination equal to the local filename must fail clearly. Failure to write either output must identify the artifact, must not report success, and must avoid leaving a misleading partial result (preflight and staged temporary outputs before replacement).

### Setup-skill and documentation requirements

- The AgenticLens setup skill must tell users that `init` creates both shared and local config files, explain which settings belong in the local file, guide them to review/edit per-harness paths and the default source, and instruct them to ensure `.agenticlens-config.local.yaml` is in `.gitignore` before committing.
- The skill must explain that the Codex setting is a Codex home/rollout directory, while the OpenCode setting is a SQLite database path, and retain the no-credentials guidance.
- The setup manual and README must describe the generated local file and its ignore/privacy role consistently with the skill and CLI behavior.

### Verification additions

- Add CLI tests for generating both files, stable sanitized defaults, no user-home leakage, dry-run behavior, existing-file protection/`--force`, output-path collision, and actionable partial-write failures.
- Add or update documentation/skill checks as available, and verify all setup paths describe the generated local config and `.gitignore` requirement. Existing source precedence, independent source, legacy compatibility, and no-fallback tests remain required by the base contract.
