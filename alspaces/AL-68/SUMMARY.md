# AL-68: Automatic session capture hooks

## Baseline and planning

- Active ticket: AL-68; specification: `alspaces/plan/to-do/AL-68 install harness lifecycle hooks for automatic session capture.md`.
- Initial Git status: clean. Baseline commit: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Existing ticket branch renamed to `feat/AL-68-automatic-session-capture-hooks` to match the feature classification.
- Workspace root from `.agenticlens-config.yaml`: `alspaces`.
- Planner completed `CONTRACT.md` and `iterations/01/PLAN.md`; application implementation awaits explicit approval.
- Proposed coverage publication uses a versioned `capture_report.json` bound to the canonical v5 ledger digest, avoiding a canonical ledger migration.
- Installed harness versions observed during planning: Codex 0.166.0 and OpenCode 1.19.5. Callback execution, identity mapping and real capture remain unverified.
- Parent reviewed the planning artifacts against the specification and repository workflow. Required implementation reviews cover behavior, data, architecture, tests, security and release integrity under both configured model profiles.
- No application changes or implementation checks performed. Telemetry deferred until workflow finalization.

## Scope approval

The user approved implementation on 2026-11-01 and required hooks to interact with the agent for correct information without guessing. The approved plan records that constraint. Missing ticket context is requested through supported harness channels; exact source/session identity and a per-session challenge validate the reply. Model, usage, cost and native parents come from source records, not an agent's guess.

## Implemented behavior

- Domain: exact source/store/session identity, explicit conflict-safe ticket association, checkpoint precedence and explicit pending/unknown states. Canonical schema v5 is preserved.
- Adapters: embedded Codex callback/OpenCode plugin handlers; bounded durable local registration; runtime-supplied OpenCode binding tool; source-proven parent recovery; ignored local state; fair bounded reconciliation; source warning/failure preservation; additive owned installation/rollback/disable/removal.
- CLI/publication: capture diagnostics/request/bind/reconcile commands; registered independent roots and exact legacy selectors; owner/source conflict checks; recoverable kernel publication locking and staged ledger/report writes; digest-bound redacted coverage report with truthful completion/per-owner checkpoint semantics.
- Setup/distribution: staged compatible native upgrades and rollback in both downloaders/wrappers; network-free callback dispatch; direct Unix/PowerShell and operator guidance.

## Correctness corrections

Individual reports and FINDINGS.json retain twelve confirmed finding classes and model-specific evidence. Corrections cover stale/failed/ambiguous snapshot replacement, overlapping publication, exact binding conflicts and source aliases, child recovery, failed-read accounting, safe templates, stdin deadlines, compatible replacement, setup guidance, truthful cutoffs and ticket-scoped diagnostics. Persistent rotation across four work classes prevents both session and byte starvation while retaining unresolved evidence.

Existing facts remain visible when ordering is unknown. Differing snapshots with equal/unavailable source invocation completion checkpoints cannot be assumed newer, for explicit or registered selection. One existing update-in-place fixture now supplies the actual message JSON timestamps for its already intended chronology; all numerical, steering, compaction and multiplicity assertions were retained, with separate missing-checkpoint guards.

## Verification and runtime evidence

Full make passed on candidates01/02/03 (all Go race/vet, frontend typecheck/lint/build, 35 frontend files/781 tests, release tests and capture distribution tests). Final candidate04 affected domain/capture/CLI race tests pass; unchanged source-package results carry forward. Distribution correction suite passes fourteen tests, including actual native rollback probes. Corrected Windows/Darwin cross-builds pass; native non-Linux execution remains unverified. The final integrated make remains a separate delivery gate.

Real OpenCode1.19.5 Linux smoke verified two independent roots and a native child, actual agent binding responses, exact source identities/parent, available usage and unknown effort, plus unchanged-repeat preservation. A fresh corrected interaction/reconciliation also passed. Details: iterations/01/SMOKE.md.

Codex0.166.0 capability/configuration is observed, but actual trusted callbacks, agent binding and native-child capture remain unverified. The parent requested human trust review of the concrete isolated hooks at /tmp/al68-codex-smoke-jykimf10/.codex/hooks.json; no answer has been received. No trust edits or bypasses were used. Other platforms, versions/interfaces and OpenCode v2 native runtime are explicitly unverified.

## Current delivery state

Frozen source candidate04: c34cc949278608f4c777f7428a620e2ae29542e7fe33875d5b13ae9f0d17e6ec. All twelve final focus/profile reviews approve their corrected code scopes. Overall delivery remains incomplete pending integrated checks and required live Codex trust/interaction evidence. The ticket index remains In Progress and the specification stays under plan/to-do; no PR has been published. Telemetry generation is intentionally deferred to workflow finalization.
