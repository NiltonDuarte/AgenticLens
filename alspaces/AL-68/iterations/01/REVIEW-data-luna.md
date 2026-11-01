# AL-68 Data Integrity Review — data-luna-01

## Assessment

**CHANGES_REQUESTED** for the reviewed candidate. The capture registry preserves accepted per-session facts on failed extraction, but its advertised cutoff does not bound the source snapshot, so published coverage can claim an earlier cutoff than the facts it contains. Full Codex runtime acceptance also remains incomplete; this is an acceptance evidence gap, not a code defect inferred from the missing smoke.

- Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Candidate fingerprint: `c7dade3313a1d65a4269d595c782750655be8a044c020dc1f3ac502464cf325a`. Recomputed from the sorted tracked and nonignored untracked regular repository files outside `alspaces`, using path/NUL/content/NUL, and matched `CANDIDATE.json`.
- Requested profile: `codex / gpt-6-luna / medium`, run `data-luna-01`. The T3 runtime declaration identifies this session as gpt-6-luna with medium effort; no additional provider usage metadata was available to me.
- Scope: candidate diff plus relevant unchanged source extraction, persisted ledger merge, store ownership and generation paths. Product/tests remained read-only. Required TEST_REPORT, SMOKE, SETUP_IMPLEMENTATION and VERIFICATION were read as raw evidence. The recorded full `make` passed and is reused; I did not rerun it.
- Requirements: `CONTRACT.md`, `iterations/01/PLAN.md`, and ticket spec `alspaces/plan/to-do/AL-68 install harness lifecycle hooks for automatic session capture.md`. The plan explicitly forbids guessing. OpenCode 1.19.5 smoke records actual root binding and a native child binding whose source `parent_id` matches the root; independent roots and repeated reconciliation are also source-checked. It records unknown effort rather than inventing one. Codex 0.166.0 trust and callback interaction are still pending human review, so the required Codex real parent/child and independent-root acceptance is not established.

## Finding

### High — Report cutoff can precede facts extracted into the same report

`backend/internal/adapters/capture/reconcile.go:90` assigns `Cutoff` before listing and extracting registrations. Later, at lines 134–150, each source is read and its own `updatedAt` is used only for comparison with the previously accepted snapshot; it is never compared with the report cutoff. OpenCode obtains `time_updated` during its later SQLite read in `backend/internal/adapters/opencode/capture_snapshot.go:37–45`. Codex reads the source file later in `backend/internal/adapters/capture/reconcile.go:274–303` and derives its timestamp from parsed records.

Trigger: a registered session is still changing, or source flushes between reconcile start and that owner's extraction. The newly observed cumulative snapshot can then include source facts later than `Cutoff`, while both `capture_report.json` and the appended summary label that earlier instant as the capture snapshot cutoff. This makes the report's coverage boundary inaccurate and prevents consumers from knowing which source facts the published ledger includes at the claimed cutoff. The contract requires a documented cutoff and says source flush/revision checkpoints govern replacement.

Evidence is source-level and direct; no disposable timing probe was run. Fix by defining and enforcing cutoff semantics (for example, capture a source revision/checkpoint no later than the cutoff, or set the reported cutoff only after all accepted snapshots and describe it as a completion boundary). Add a deterministic test where a source timestamp advances past the initial reconcile time and assert the ledger/report boundary remains truthful.

## Consequential hypotheses assessed

- **Failed or partial extraction erases accepted session usage:** refuted for the inspected paths. Extraction errors, empty/unflushed sources, and older/truncated snapshots update state/detail without replacing `Snapshot`; `TestReconcilePreservesAcceptedSnapshotsThroughResumeTruncationAndMissingSource` and partial-parse coverage exercise these cases.
- **Cumulative resume snapshots double count:** no defect found in the inspected capture path. It replaces the exact registration snapshot, and the smoke records an unchanged repeat with the same compact snapshot digest. Existing ledger union uses `mergeSessions`; the raw tests cover growth and unchanged repeats.
- **A child callback's parent ID is mistaken for child identity:** refuted in the inspected Codex path. `SubagentStart` does not bind/register a child, and `SubagentStop` queues the transcript path; reconciliation parses and verifies the child transcript's `ParentThread` before registration.
- **Concurrent writes/crash leave a half-written registration:** no defect established. `Store.Change` uses SQLite `BEGIN IMMEDIATE` and rollback; tests cover simultaneous writers and process exit during an uncommitted transaction. Reconciliation is serialized per registration mutation rather than one transaction around the entire pass; I did not establish an observable data corruption from that design in this pass.
- **Report/ledger write failure leaves a stale pair:** `main.go` writes the ledger before atomically replacing the report. A report failure returns nonzero with an explicit diagnostic, but does leave the newly written ledger and any prior report on disk; the report digest lets a consumer detect mismatch. This is a recoverable publication limitation, not a separate finding here because the CLI does not claim the pair succeeded. The parent should keep this state visible in final handoff.

## Verification and limits

- Reused: recorded `make` exit 0 on the frozen fingerprint; focused race suites and publication digest tests as documented in raw `TEST_REPORT.md`/`VERIFICATION.md`.
- Inspected source ownership by exact `(harness, store, session ID)`, registration binding, source snapshot replacement, OpenCode exact-family selection, Codex child-parent verification, final merge and report generation. Report redaction removes request tokens and source-store paths before publication; local registration state retains them for reconciliation.
- No independent runtime probe was run. OpenCode source-backed smoke is meaningful actual harness/agent interaction on Linux amd64 1.19.5. Codex callbacks and user trust remain unverified; no bypass or user prompt was attempted. Other platforms, OpenCode v2 native smoke, resumed-session interaction and real interrupted-hook recovery remain outside observed runtime evidence.
- Outcome is scoped to data integrity/failure handling and this frozen candidate only. The finding requires correction; the Codex smoke gap must remain explicit in any completion claim.
