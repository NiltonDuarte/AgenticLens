# AL-68 behavior review: behavior-sol-01

Outcome: CHANGES_REQUESTED. Two reachable behavior defects were reproduced. Required real Codex evidence also remains incomplete; this report grants no runtime support approval.

## Revision and identity

- Phase: independent discovery, behavior/integration focus. Requested profile: codex/gpt-6.1-sol/high.
- Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Candidate: `c7dade3313a1d65a4269d595c782750655be8a044c020dc1f3ac502464cf325a`, using CANDIDATE.json's sorted repository path/NUL/content/NUL definition outside alspaces. Independently recomputed the identical fingerprint after the probes.
- Directly observed environment strings: CODEX_THREAD_ID `01a19857-1005-7980-8c08-3d69e249e33c`; CODEX_SESSION_ID `01a19830-6b3f-7f63-964f-a7521381d0fb`. These are environment evidence, not verified source attribution. No actual model/effort, usage or cost was independently established for this reviewer; the requested profile is not asserted as measured metadata.
- Product and tests remained read-only. No sibling reviewer reports were loaded. Requirements inspected: ticket specification, CONTRACT.md and iteration PLAN.md, including actual agent interaction and explicit ownership requirements.

## Findings

### B-SOL-1: Failed registry refresh overwrites newer accepted ledger facts (P1)

Locations: `backend/internal/adapters/capture/reconcile.go:135`, `:247`; `backend/cmd/agentic-lens/main.go:201`, `:312`, `:333`; unchanged `mergeSessions` at `main.go:1464`.

Reconcile correctly retains its own previous snapshot when source extraction fails, but returns that snapshot in Sessions without distinguishing it from a successfully refreshed owner. Generate treats every returned row as authoritative and overwrites the existing canonical row by ID. Registry precedence checks compare only with the registry's previous revision, never the newer canonical facts or a successful explicit-selector extraction.

Reproduction against a freshly built native candidate, entirely in a temporary repository:

1. Deliver SessionStart for exact root, bind its request to TEST-100, and generate with the registry: input usage 10.
2. Advance that exact source to input usage 20. Generate using the supported explicit `--session-id root --codex-home <store>` path with a different, nonexistent `--capture-dir`. The existing ledger now contains 20.
3. Remove the root source file. Generate the same ledger with the original registry and no explicit selector.

Observed: command succeeds, report marks root `failed`, and canonical input usage regresses from 20 to 10. Expected: unavailable source enrichment preserves the already accepted 20. Existing missing-source tests hold identical snapshots in registry and ledger, so they do not expose this ordering failure. The same unconditional merge also lets retained registry rows replace freshly selected rows when both paths are used.

Required correction: preserve authoritative existing/freshly selected owner facts when the registry did not accept a newer source snapshot; add an end-to-end regression where canonical and registry snapshots differ before the failure.

### B-SOL-2: Native Codex children depend on SubagentStop despite a received start callback (P2)

Locations: `backend/internal/adapters/capture/hooks.go:75`, `:80`, `:83`; `backend/internal/adapters/capture/reconcile.go:263`; `backend/cmd/agentic-lens/main.go:201`.

SubagentStart registers the parent and returns a warning to the child, but retains neither pending child work nor unresolved start correlation. Only SubagentStop queues the child's source path. Exact registered extraction reads the parent alone, while registered generation bypasses the existing native-descendant discovery path. Therefore a running child, or one interrupted before its stop callback, is omitted even after its authoritative source has flushed.

Disposable reproduction: bind root, place a valid child rollout in the same source store with source-proven parent_thread_id=root, deliver SubagentStart with session_id=root and a distinct agent_id, then generate through the registry without delivering SubagentStop.

Observed: ledger IDs `[root]`, pending_children=0, and root `current`. Run the existing explicit `--session-id root` selection against the same source files: ledger IDs `[child, root]`. No agent_id was assumed to equal a source ID. This is a source-backed child the baseline discovery can find, not an independent session requiring guessed ownership.

Expected: recover the child through authoritative parent evidence within the reconciliation bound, or retain and expose unresolved child-start work until identity can be proven. Discarding a received start and depending exclusively on the stop event conflicts with interrupted-session recovery and the accepted prohibition on relying solely on shutdown callbacks. This source-level/probe finding is distinct from the unavailable real Codex smoke; it does not claim actual callback execution was verified.

## Hypotheses investigated

- Ambient ticket guessing or cross-session challenge reuse: inspected request/Register/Bind, identity validation and OpenCode runtime tool context. Explicit request tokens are persisted per source/store/session; conflicts reject, and plugin binding uses context.sessionID. No finding in those inspected paths.
- Parent session_id incorrectly becoming a Codex child ID: the handler guard explicitly prevents that. The resulting missing-start recovery is B-SOL-2, rather than fabricated parentage.
- Lost independent roots or double counting: inspected registered exact-owner extraction, fixed-point inherited binding and ID replacement, and reused integration/idempotence tests. Independent explicitly bound roots are included. Repeated snapshots replace rows; failure preservation across unequal registry/canonical snapshots is B-SOL-1.
- OpenCode family/title fallback: capture selects the uniquely owning exact ID across complete v1/v2 families and passes exact extraction; duplicated IDs reject. Unchanged general extraction retains its earlier default descendant behavior. Fixture tests cover missing IDs alongside unrelated matching titles. No additional behavior defect established.
- Installer compatibility and unrelated settings: inspected additive hook groups, plugin hash ownership, native protocol preflight, both downloader delegation paths and wrapper capture dispatch. Raw distribution tests exercise preservation/conflicts and argument-safe delegation. These do not establish real Windows/macOS execution or trust.
- Source failure and publication: inspected unchanged canonical loading/merge, source parsers, compact persistence and digest/report generation. Registry retention alone does not establish final-ledger preservation, as reproduced. No claim of full publication failure/concurrency certification is made in this focus report.

## Verification and evidence limits

Executed `go build -o <temporary>/agentic-lens ./cmd/agentic-lens` from backend, followed by a Python-driven native CLI probe using temporary Git/config/source/registry files. Build and all probe commands exited zero. Probe directory: `/tmp/al68-behavior-sol-wa4lm0ia`; binary SHA-256: `a6c283bf32aa70f64e5df11d11ab24e758627daf08b60aa1d68603d77da2438e`.

Probe output:

```text
initial_registry_input 10
newer_accepted_ledger_input 20
after_failed_registry_refresh_input 10
failed_registration_state failed
registered_generation_session_ids ['root']
registered_generation_pending_children 0
registered_generation_states [('root', 'current')]
explicit_selector_session_ids ['child', 'root']
```

Reused TEST_REPORT.md's focused race checks and raw SMOKE.md/SETUP_IMPLEMENTATION.md evidence. Parent VERIFICATION.md records full `make` exit 0 on the identical frozen product candidate, including all Go race tests, frontend 781 tests, typecheck/lint/build, release tests and distribution tests. No duplicate full suite was run here. Passing checks do not refute the reproduced cases.

OpenCode Linux CLI 1.19.5 smoke reports actual agenticlens_bind exchanges, two independent roots, a native child and source comparisons; this is reused evidence rather than independently rerun smoke. Codex trust and actual callback/agent/parent-child source verification remain pending. Do not bypass trust. Windows/macOS runtime paths, OpenCode v2 native runs, resumed-runtime interaction and T3 Codex children remain unverified as recorded in SMOKE.md. Fixes require affected regression checks and revision-aware review; runtime evidence gaps remain separate delivery gates.
