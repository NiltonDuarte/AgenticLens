# Iteration 01: AgenticLens operator skill discovery

## Discussion Input

The setup skill helps users install AgenticLens in another repository, and metric generation covers generating metrics. Users still need guided help to operate the installed platform, including migration, historical backfill, and issue diagnosis. The agent working in that repository may not be able to inspect AgenticLens source code, so the operator skill should rely on the installed CLI, local configuration, telemetry ledgers, and command output.

The user clarified that migration means running `agenticlens migrate` to migrate stored telemetry. They also clarified the lifecycle: migration can leave a ledger in an intermediate state until source-backed backfill is completed, and the backfill must run on a machine with the original agent session data. The ledger includes user names, making it straightforward to identify which developers own relevant sessions.

The skill should help a team recognize an ongoing migration, identify the developers who still need to contribute their local backfill, guide each developer through that action on the machine containing their source sessions, and verify when pending work has cleared.

The user clarified that migration completion across developers is team-coordinated. The operator cannot infer whether another developer's backfill is merely pending or blocked; it must ask the team developers. If backfill cannot find a session or the original source data is unavailable, leave the metrics JSON in partial migration state rather than treating the migration as complete.

The skill does not need to implement or second-guess the installed repository's access controls. For changes, it must act only when the user directly requested or already allowed the action; otherwise it must ask permission first.

## Approved Skill Scope

The user requested implementation of `.agents/skills/agentic-lens-operator/` with focused references that let an agent operate AgenticLens from an adopting repository. The skill must cover each CLI command and its arguments in detail, the embedded dashboard UI, reading and modifying shared and machine-local configuration, migration/backfill coordination, and a last-resort procedure for consulting the public source repository at the tag matching the installed release. The source repository is the last resort after installed command help and local documentation are insufficient.

## Repository Findings

- `agenticlens migrate` performs stepwise schema migrations across ticket workspace ledgers. Its controls include dry run, backups, ticket filtering, a target schema version, and summary refresh.
- A schema-only migration can mark the ledger `partial` with `source_backfill` pending; the migration notes say source-backed backfill is required before runtime ingestion.
- `agenticlens backfill --recompute-db` re-queries the configured local OpenCode SQLite database. The relevant session data must be available on the machine running it.
- Backfill clears the pending source-backfill step when it finds the source session. If the session is not found, migration remains pending.
- `agenticlens update` upgrades installed platform binaries; it is separate from telemetry schema migration.
- The current migrate command scans workspace directories for `agentic_metrics.json` ledgers. It does not coordinate or report migration progress across developers' local environments.
- The local OpenCode trace database is a source for generation and database recomputation, distinct from the workspace telemetry ledgers that migration changes.

## Working Product Direction

Explore a dedicated AgenticLens operator skill for use from a repository where the platform is installed. It should help agents perform supported maintenance and recovery tasks even when they cannot inspect the AgenticLens source repository. For migration, it should distinguish schema migration from source-backed backfill, identify session owners from partially migrated ledgers, explain that each relevant developer must run backfill on the machine holding their original session data, and verify pending-step status afterward. It should make clear that one successful local action does not establish team-wide completion when other owners still have work.

For unavailable source data, preserve the ledger's partial migration state and explain the remaining pending step. For team-wide progress, use team coordination and ask developers directly; do not claim to determine remotely whether their work is pending, blocked, or failed. Respect the host repository's existing access controls, and require direct user authorization or permission for every change.

## Open Questions

No open questions remain from this discussion. The user decided to preserve partial state when source data is unavailable, coordinate status with the team, rely on the repository's own access controls, and require explicit user authorization for changes.

## Status

Skill implementation requested by the user and completed. Scope above was approved; no AgenticLens application code was changed.
