---
name: agentic-lens-operator
description: Operate an installed AgenticLens CLI and dashboard from an adopting repository. Use for telemetry generation, scanning, backfill, schema migration, configuration, dashboards, and troubleshooting; does not develop AgenticLens itself.
---

# AgenticLens Operator

Help users operate an installed AgenticLens release from the repository where they use it. Do not assume the AgenticLens source checkout is available. Prefer evidence in this order:

1. Installed CLI help and version output.
2. Documentation, configuration, ledgers, and dashboard available in the adopting repository.
3. The public AgenticLens source at the exact matching release tag, only when the first two sources do not answer the question.

## Operating Workflow

1. Identify the requested operation and inspect the relevant existing state before proposing a command: current worktree, stable user common wrapper and setup scripts, shared exact pin/selected native path/version, shared and local config, workspace path, ledger migration state, or dashboard URL. A missing ignored/untracked file is not evidence that the shared configuration needs initialization; follow the worktree repair instructions in the CLI reference. Git tracking and ignore choices remain with the repository owner.
2. Read [references/cli.md](references/cli.md) for command syntax and argument behavior. Read [references/dashboard.md](references/dashboard.md) for UI routes and interpretation. Read [references/configuration.md](references/configuration.md) before reading or editing config.
3. Explain what the command will read and change. A dry run is useful when supported, but it does not replace explaining side effects.
4. Make a change only when the user directly requested or already allowed that specific change. Otherwise ask permission before running it. Honor access controls enforced by the adopting repository and host environment; do not add a separate access-control model.
5. Verify outcomes using command output and persisted ledger state. Report the exact scope, completed work, remaining pending steps, and any errors. Do not infer team-wide completion from one developer's local run.

## Migration and Backfill Coordination

Ledgers are shared project records across developers, machines and harnesses. Preserving supported recorded facts does not require the producing machine's private database or logs; local sources are needed for new extraction or missing enrichment. Follow the [portable ledger contract](../../../TECHNICAL_GUIDELINES.md#portable-ledger-contract); report an installed release's conflicting behavior as a compatibility gap rather than inventing receipts or declaring existing history untrusted.

- `agentic-lens migrate` changes schema shape in workspace ledgers. A shape-only migration may leave `migration.status: partial` with `source_backfill` in `migration.steps_pending`. Report pending source enrichment and the installed release's actual runtime-ingestion limits; preserve supported recorded fields without treating missing enrichment as evidence that history is untrusted.
- Source-backed completion uses `agentic-lens backfill` on a machine that has the original agent sessions. Select the correct harness and source path. For OpenCode, `--recompute-db` reads that machine's local SQLite database. For Codex, choose Codex as the source and point `--codex-home` to the session files.
- Backfill updates only sessions present in the selected local source; preserve all unavailable sessions. Local source values take precedence for successfully refreshed records; the existing ledger supplies unavailable history. The runtime recomputes full-ledger aggregates. Migration is best-effort: keep a hard blocker and its remaining work in explicit resumable state for another developer/machine; do not require private local receipt state to continue or manually discard records.
- Use recorded session user names to identify developers whose source sessions may be needed. Ask the team directly to coordinate who has run their part. The skill cannot distinguish remotely between another developer's work being pending, blocked, or failed.
- If the source session cannot be found or its original data is unavailable, preserve the metrics JSON in partial migration state. Explain the pending step; do not mark it complete or remove migration metadata manually.
- Reinspect the ledger after each authorized backfill and report whether pending steps remain. A developer's local success does not establish team-wide completion.

## Source Inspection as a Last Resort

When installed help and adopting-repository artifacts are insufficient, read the installed binary's version metadata (`agentic-lens version --json`) and identify its exact public release tag. Inspect only the matching tag in `https://github.com/NiltonDuarte/AgenticLens` and cite the paths or symbols that answer the question. If the installed build has no matching tag, use its reported commit only if the public repository contains that exact commit; otherwise report that the source cannot be matched reliably. Never use the moving `main` branch to explain an installed release when a release tag is available.

## References

- [CLI commands and arguments](references/cli.md)
- [Dashboard pages and interpretation](references/dashboard.md)
- [Configuration schema and editing](references/configuration.md)
