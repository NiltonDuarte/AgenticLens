# Configuration Reference and Editing

AgenticLens uses shared project configuration and optional machine-local source configuration:

- `.agenticlens-config.yaml`: project identity, ticket parsing, workspace layout, metrics rules, dashboard behavior, and role presentation. Usually tracked.
- `.agenticlens-config.local.yaml`: machine-specific telemetry source locations and default source. Usually ignored by Git; verify the adopting repository's policy before changing or committing it.
- `.agenticlens-config.local.example.yaml`: portable sanitized example when present.

Never put credentials or secrets in either file. Local source paths identify the current machine's own OpenCode database or Codex home; they do not migrate or copy that data.

## Read Before Editing

1. Locate both files and read the relevant sections, plus any example file.
2. Determine which setting belongs in shared config versus local config. Do not move machine-specific paths into a tracked shared file.
3. Preserve comments, unrelated keys, types, and existing values. Change only the requested key.
4. Explain the exact file/key/value and expected effect. Edit only if the user directly requested or already allowed that change; otherwise ask permission.
5. After an authorized edit, inspect the diff and check YAML syntax with an available YAML parser. Use a config-consuming read-only command such as `agentic-lens --list` only when a valid ledger/workspace exists. There is no dedicated config validation command.

Do not use `agentic-lens init --force` to change one setting: it overwrites shared and local generated configuration. `init --dry-run` previews generated defaults but is not a round-trip editor and may not preserve existing custom values.

## Source Selection and Path Precedence

For telemetry extraction, explicit command flags take precedence over machine-local source paths. The default harness is selected by `generate --source`, then `telemetry.default_source`, then OpenCode. If the selected source is unavailable, the CLI fails rather than silently switching to another harness.

OpenCode uses a SQLite database path; Codex uses a home directory containing rollout files. Relevant keys:

```yaml
telemetry:
  default_source: opencode
  sources:
    opencode:
      database_path: ~/.local/share/opencode/opencode.db
    codex:
      home: ~/.codex
```

OpenCode path resolution generally prefers CLI override, machine-local `telemetry.sources.opencode.database_path`, legacy shared `database.path`, then adapter default. For Codex, use `--codex-home`, local `telemetry.sources.codex.home`, `$CODEX_HOME`, then `~/.codex` as supported by the installed version. Paths are local to the machine executing the CLI.

## Shared Configuration Keys

| Key | Purpose and guidance |
|---|---|
| `project_name` | Project label shown in CLI/dashboard. |
| `ticket.prefixes` | Prefix strings or objects with `prefix`, `description`, optional default `ticket_type`. |
| `ticket.case_insensitive` | Whether ticket prefix matching ignores case; default true. |
| `ticket.custom_patterns` | Additional regexes for nonstandard ticket IDs. |
| `workspaces.dir` | Root directory containing ticket workspaces and ledgers. |
| `workspaces.artifacts` | Relative artifact paths and labels shown in ticket detail. |
| `database.path` | Legacy/shared OpenCode DB path; prefer local source path for per-machine setup. |
| `database.query_timeout_ms` | Query timeout in milliseconds. |
| `database.busy_timeout_ms` | SQLite lock wait timeout in milliseconds. |
| `database.read_only` | OpenCode DB access mode; preserve read-only behavior unless the user asks for an explicitly supported change. |
| `live.active_threshold_seconds` | Window used to classify sessions as active. |
| `live.poll_interval_ms` | Live server polling interval. |
| `live.max_trace_buffer_messages` | Maximum in-memory live trace events. |
| `live.anomaly_thresholds.*` | Stall seconds, tool error ratio, context exhaustion ratio, and other supported anomaly limits. |
| `code_metrics.base_branch` | Git comparison base; when omitted, the CLI probes common upstream/main/master refs. |
| `code_metrics.source_roots` | Directories included as source roots. |
| `code_metrics.include` / `exclude` | Glob patterns to include or omit from code metrics. |
| `code_metrics.test_patterns` / `prod_patterns` | File classification patterns. |
| `role_mappings` | Map raw agent personas to canonical reporting categories. |
| `role_order` | Preferred display order; unlisted roles follow alphabetical order. |

Use the installed version's config example and help/source as authoritative if a key or default differs. Unknown or invalid values may fail config loading; avoid inventing keys.

## Safe Editing Pattern

For a directly authorized, narrow edit, preserve the rest of the file and show the diff:

```yaml
telemetry:
  default_source: codex
  sources:
    opencode:
      database_path: ~/.local/share/opencode/opencode.db
    codex:
      home: ~/.codex
```

Do not replace the entire configuration with this excerpt. Keep machine-local files ignored by Git. Before changing a path, check that it exists and belongs to the intended local harness. If the source is missing, report it; do not silently substitute another developer's or harness's data.
