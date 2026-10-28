# AL-56 Iteration 11 Plan: Make Migration Metadata Temporary

## Objective

Correct iteration 10's runtime-readiness rule. Completed canonical schema-v4 ledgers must load without persisted migration metadata; migration metadata must be retained only while an explicit multi-step migration/backfill is incomplete, then removed by the successful final step. Keep ordinary runtime incompatible with schemas v1-v3 and with any persisted in-progress/incomplete migration state.

## User Correction and Contract Delta

- Treat the existing completed schema-v4 ledgers as already migrated; they must not need an added `source_backfill` marker to load.
- Supersede iteration 10's universal durable `source_backfill` readiness requirement in `CONTRACT.md`.
- Migration metadata records resumable workflow progress only. It may be carried between systems or steps while migration/backfill is incomplete.
- The final successful migration/backfill step validates and persists canonical v4, removes all migration metadata, and only then reports the ledger runtime-ready.
- Ordinary runtime must reject any ledger still carrying migration-in-progress or incomplete state. A metadata-free canonical v4 ledger is complete by structural validation.

## User Scope Addition: Partial Workspace Availability

- The UI must continue loading and displaying valid ticket workspaces when one or more discovered ledgers are malformed, unsupported, or unmigrated.
- A ledger rejected by the iteration 11 runtime rules remains excluded from every aggregate and detail result; it must not prevent unrelated valid ledgers from loading or make the platform appear unavailable.
- Surface a clear, persistent warning identifying each failed ticket workspace and why it was skipped. For schema v1-v3 and in-progress/incomplete migration state, include the applicable migration/backfill remediation. Do not expose raw ledger contents.
- This is an explicit scope addition to the already approved iteration 11 plan and is included in the approved implementation scope below.

## Design and Implementation Guidance

### Runtime validation

- Keep schema-version gating ahead of fact collection decoding: reject v1, v2, and v3 with the detected version and explicit migration plus source-backed backfill remediation.
- For v4, validate the canonical session-owned shape and required fields. Accept it when no migration metadata exists.
- If the v4 document contains persisted migration metadata indicating in-progress or incomplete work, reject it before returning or merging any facts. Explain that the explicit workflow must be resumed.
- Do not require, synthesize, or check a standalone `source_backfill` marker as universal eligibility state.

### Workspace scan, API, and UI behavior

- Preserve `ScanWorkspaces`' per-ledger isolation: read, validation, decode, or normalization failure for one `agentic_metrics.json` is recorded as a diagnostic while scanning continues and valid ledgers remain available.
- Trace the workspace scan result through the aggregate HTTP endpoints used by the UI (`/api/workspaces`, `/api/fleet`, `/api/comparison`, and `/api/progression`). A per-ledger diagnostic must not turn an otherwise usable partial result into an HTTP failure. Keep genuinely fatal scan/setup failures distinguishable from individual rejected workspaces.
- Return diagnostics alongside the successfully derived data (or an equivalent explicit response contract), retaining enough structured identity to name the affected ticket workspace and a user-readable reason. A rejected schema v1-v3 ledger should identify its detected version and migration/backfill path; an incomplete v4 migration should explain that the explicit workflow must be resumed.
- Update the TypeScript API types/client and UI data flow so successful partial responses populate valid workspace rows and totals while displaying a clear warning that some workspaces were skipped. Warnings identify each ticket workspace that failed to load and remain visible on the relevant workspace aggregate views; they must not be swallowed by the current `Promise.all` loading path.
- Fleet and other aggregates are calculated only from accepted ledgers. Do not fabricate metrics for skipped workspaces or report the partial result as complete without warning.

### Explicit migration and resumption

- Persist enough migration progress metadata to resume an incomplete, multi-step migration/backfill on another system or in a later invocation. Preserve existing source selection, ticket scoping, backup, dry-run, atomic-write, and Git-fact preservation safeguards.
- Intermediate migration/backfill steps must keep the incomplete state visible and must not make the ledger pass ordinary runtime validation.
- In the final successful step, complete source-backed work, validate canonical v4, and persist the completed ledger with migration metadata removed. Do not report success/readiness if removal or validation fails; preserve failure safety and recoverability.
- Ensure already-completed canonical v4 ledgers without migration metadata remain untouched and usable; do not run workspace-wide rewrites or backfills.

## Acceptance Criteria

- Ordinary runtime accepts structurally canonical schema-v4 ledgers with no migration metadata, including existing completed ledgers.
- Ordinary runtime rejects schema v1, v2, and v3 before parsing legacy facts and gives actionable migration/backfill guidance.
- Ordinary runtime rejects any persisted migration-in-progress or incomplete state without returning/merging partial facts.
- An intermediate migration state can be resumed across systems/steps and stays runtime-ineligible until completion.
- Successful completion of the final migration/backfill step removes all migration metadata before the ledger becomes runtime-ready.
- A schema-v4 ledger with completed migration metadata is finalized and persisted without metadata instead of being skipped; metadata-free v4 ledgers with no work remain skippable.
- Explicit `--recompute-git` and `--recompute-db` requests bypass completed-ledger skip predicates so requested recomputation is processed.
- No persisted `source_backfill` marker remains as a universal requirement for completed v4 eligibility.
- Tests cover: completed canonical v4 without metadata loads; intermediate state is rejected by runtime and can resume; final successful migration removes metadata; and v1-v3 each reject before legacy fact parsing.
- With a mix of valid and invalid/unmigrated ticket ledgers, workspace scanning retains valid ledgers and records a diagnostic for every skipped ticket workspace.
- Aggregate endpoints used by the UI return usable partial workspace data and per-workspace diagnostics when individual ledgers fail; a single bad ledger does not make those endpoints or the platform unavailable.
- The UI displays valid workspace data and aggregates while clearly warning which ticket workspaces failed to load and why, including actionable migration guidance for legacy or incomplete ledgers.
- Fatal workspace-directory/setup failures remain distinguishable from recoverable per-ledger failures and continue to produce an appropriate unavailable/error state.
- Coverage verifies mixed valid/invalid workspaces through scan, API response contracts, and UI partial-load/warning behavior.

## Scope Boundary

The approved implementation scope for iteration 11 includes temporary migration metadata for resumable incomplete migrations, compatibility for completed canonical v4 ledgers without migration metadata, partial backend/API scan responses that preserve valid workspace data and report per-workspace diagnostics, and frontend warning behavior that keeps those diagnostics visible while displaying valid data.

Implementation scope includes the backfill skip/finalization behavior and regression coverage for the CLI and engine lifecycle. Verification, telemetry, and commit are part of the implementation workflow; no ledger migration or edits to other ticket workspaces are included.
