# FIX-18 — Iteration 01

## Selection and implementation

- [x] Trace baseline production selection from `backend/cmd/agentic-lens/main.go:runServer` into `OpenReadOnlyHook`, `LiveService`, the live handlers and dashboard. Reconcile the ticket's historical mixed-layout premise against current source and CAP-02.
- [x] Record explicit user approval in [design.md](design.md) before coder dispatch. Coordinate any overlap in server composition, schema detection and tests with active owners.
- [x] Implement capability selection, ownership of the historical handle, startup output and live reader gating. Route initial and subsequent failures through REST, SSE and the dashboard, with visible recovery.

## Verification

- [x] Exercise the production server selection and router with isolated configuration, explicit path override and real SQLite files. Assert startup diagnosis, REST tree and trace status/body, SSE preflight and historical endpoints from the same selection. Include clean v1 with active and idle cases, optional timestamps and native auxiliary table, clean v2 with recorded events, generic mixed, unknown SQLite, incomplete schema, corrupt, unset, nonexistent and inaccessible inputs. Use a deterministic inaccessible fixture even under elevated privileges.
- [x] Exercise a real v1 read failure after selection for the initial tree and after SSE admission for tree/trace polling, then restore the source. Assert visible diagnosis, recovery with unchanged data, event order, retry and teardown. Extend the existing recovery and initial subscription cleanup cases in `live_handlers_test.go` while retaining their ordering and cancellation assertions. Verify the live API and view display source detail across REST fallback, SSE and manual retry.
- [x] Reuse OpenCode family detection, ordinary v1/v2 extraction, live parity and mixed-source hydration cases. Verify clean v2 exact-owner events through the production router alongside rejected live requests. For competing mixed tables verify ordinary extraction rejection and exact-owner historical diagnostics. Compare ledger bytes and recorded owner facts across GETs in an isolated second-machine fixture with unavailable sources; check available Codex evidence independently. Keep required-column oracle repair with AL-74.
- [x] Record candidate revision, actual commands, fixture scope and limits. Run affected Go and web tests with canonical assets. Verify embedded assets and standalone server behavior. Attribute independent architecture, behavior, data and test review against the frozen diff; route findings to their owners.

## Delivery

- [x] Integrate the reviewed candidate with main, rerun affected checks and integrated `make`, and consolidate review. Update delivery and roadmap from verified results, generate and inspect final source-backed telemetry, then publish the PR under repository policy.

## Independent review coverage

The first pass paired `gpt-6.1-sol/high` with `gpt-6-luna/medium` for each focus. The user updated the Sol profile to `gpt-6.2-sol/high` during correction review. Earlier dispatches retain their actual profiles. Their reports use the focus and profile in the filename under `evidence/iterations/01/`.

| Focus | Question | Report files |
|---|---|---|
| Behavior | Do startup, tree, trace, stream and dashboard agree across source selection, read failure and recovery? | behavior-sol.md, behavior-luna.md |
| Data | Can unavailable live observation or a failed read hide historical facts, overwrite observations or modify the ledger? | data-sol.md, data-luna.md |
| Architecture | Are schema detection, capability selection, resource ownership and consumer state managed at the appropriate boundaries? | architecture-sol.md, architecture-luna.md |
| Tests | Do real configured sources and meaningful assertions cover selection, failures, ordering, recovery and independent history? | tests-sol.md, tests-luna.md |

### Review corrections

- [x] Preserve the previous usable tree observations and report trace enrichment errors to REST fallback clients, including clients without a selected session. Reproduce loss of observed anomalies and verify recovery on identical rows.
- [x] Exercise a regular file with denied read access through the production server. Protect the ordering of an outstanding trace request against a newer same-session stream diagnosis.
- [x] Clear a session diagnostic after a proven fresh complete tree on reconnect or REST fallback, including a session whose trace is unselected. Reject older pending REST results after a newer diagnostic.
- [x] Reconcile the initiation order of fallback and manual tree requests, and distinguish a superseded source-read error from an independent trace download failure.
