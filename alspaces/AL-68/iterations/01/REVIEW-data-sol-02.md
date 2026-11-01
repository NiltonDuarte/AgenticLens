# AL-68 data integrity re-review: data-sol-02

Verdict: **CHANGES_REQUESTED**. CAP-01 and CAP-03 have narrower successful corrections but retain reproduced blocking cases. The new bounded OpenCode child-discovery path also has a continuation defect. Required real Codex evidence remains incomplete independently of these product findings.

## Revision, scope and identity

- Phase: revision-aware affected re-review. Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Frozen corrected candidate: `17fdae376119de798b382af56952351564389a69a174e5b0049259d387151cbc`, `CANDIDATE-02.json`. Independently recomputed its listed path/NUL/content/NUL fingerprint and obtained that exact value.
- Assignment: CAP-01 snapshot precedence, CAP-02 publication ownership/atomicity, CAP-03 exact binding conflicts, CAP-05 failed-read accounting, CAP-10 completion/checkpoints and CAP-11 ticket selection/inherited children. Read the assigned data persona, common procedure, approved contract/plan, correction instructions, findings registry, previous own data report and implementation/test/smoke evidence. No sibling persona or sibling review conclusions were needed.
- Requested profile: `codex/gpt-6.1-sol/high`. Observed interface: Codex tools on the shared Linux checkout. Actual model/effort runtime metadata is not exposed by the inspected evidence; requested settings are not claimed as observations. No usage, cost or session identifier is invented.
- Product/tests remained read-only. Only this assigned report is written in the repository; native probes, their binary and synthetic fixtures are under `/tmp`.

## Blocking findings

### CAP-01 remains partially unfixed — P1: an unchanged older source can still replace newer canonical facts

Locations: `backend/internal/adapters/capture/reconcile.go:124`, `:146`; `backend/cmd/agentic-lens/main.go:339`, `:360`.

The correction protects a newer canonical owner when registered extraction fails, is pending or is skipped. It does not establish precedence when an old source remains readable. Snapshot acceptance compares only with the local registry checkpoint. An identical old checkpoint is accepted, placed in `RefreshedSessions`, and then unconditionally replaces the canonical owner. Publication locking serializes this loss but does not prevent it.

Reproduced through the corrected native CLI:

1. Register and bind exact Codex `root` to TEST-100; registry-backed generation accepts input usage 10.
2. Append a later valid source record at `00:00:03Z`, adding 20 input tokens. Explicit generation using an absent registry publishes cumulative 30 to the same canonical ledger. The original registry still has the earlier 10 checkpoint.
3. Restore the exact original source bytes, simulating an older/truncated source view. Registry-backed generation successfully rereads those bytes; they equal its local checkpoint and pass acceptance.
4. Generation exits zero, reports the owner `current`, and replaces the canonical 30 with 10.

Observed output: `registered_initial 10`, `explicit_newer 30`, `old_source_refresh_over_canonical 10`, `old_source_state current`.

Exact checkpoint evidence: the original source is 378 bytes, source `UpdatedAt = 2026-01-01T00:00:02Z`, SHA256 `77747170045c7ab513d75b627e3f109cafecf444432b558a8fe5ffe6731f9c7d`. The appended newer source is 549 bytes, last source timestamp `2026-01-01T00:00:03Z`, SHA256 `94032149b7add1410ef9b288db690cb46e23649ae1ae508b6d57490e39e72123`; that is the source actually used by the successful explicit publication of 30. Restoring the original bytes returns exactly to the registry's accepted `(378, 00:00:02Z)` checkpoint. SQLite confirms the registry remains at revision 378 / `2026-01-01T00:00:02Z` after the regressing generation. The canonical facts had already advanced through the later 549-byte/`00:00:03Z` source; rereading an equal local checkpoint does not prove precedence over them.

The canonical owner was demonstrably newer, and the restored bytes are demonstrably the older source checkpoint. This is not a proposed fieldwise maximum over token totals. Preserve accepted canonical facts when a registry reread cannot establish a newer authoritative source revision; the persisted precedence information must survive the retained explicit generation path. Existing new regression cases remove/archive the registered path and therefore exercise extraction failure, not this successful stale reread.

Evidence: `/tmp/al68-data-sol-02-probe.py`, synthetic fixture `/tmp/al68-data-probe-cp1l6gi8`.

### CAP-03 remains partially unfixed — P1: a source-store symlink bypasses a known conflicting binding

Location: `backend/cmd/agentic-lens/capture.go:338`, `:351`; capture identity creation also uses only `filepath.Abs` at `:145`.

The new guard rejects a direct matching store string. It compares absolute lexical paths rather than the actual source store. A directory symlink to that same Codex home produces a different store string while discovery reads the same exact transcript and session.

In the same disposable fixture, `root` remains explicitly bound to TEST-100. Create `source-alias -> source`, then generate TEST-200 using the same registry, `--session-id root`, and `--codex-home source-alias`. The command exits zero and publishes `root` into a ledger with `ticket_id: TEST-200`; registry status still shows TEST-100. Observed output: `alias_other_ticket_exit 0`, `alias_other_ticket_published TEST-200`, `registry_binding TEST-100`.

This is the same real store and exact session, rather than two independent stores that happen to reuse an ID. Canonicalize/compare actual store identities consistently across registration and selection, including existing stored aliases, so an ordinary source path alias cannot evade the approved binding-conflict gate.

Evidence: `/tmp/al68-data-sol-02-probe.py`, `/tmp/al68-data-probe-cp1l6gi8/source-alias`.

### DATA-02-NEW-01 — P2: bounded OpenCode child discovery never advances past known children

Locations: `backend/internal/adapters/capture/children.go:139`, `:146`; `backend/internal/adapters/opencode/capture_snapshot.go:126`.

The CAP-11 correction adds parent-scoped discovery. Its query selects the first `budget.sessions` child IDs in sorted order, then skips already registered children in Go. There is no OpenCode child cursor or exclusion of known children before LIMIT. Consequently, a known prefix permanently hides later children when that prefix fills the selected allowance.

Native CLI reproduction with a real disposable SQLite source:

1. Seed bound parent `root` and native child `child-a`; reconcile with `--max-sessions 10`. Both become registered.
2. Insert source-native `child-b` with `parent_id = root`; it has no lifecycle registration yet.
3. Run four successive passes with the supported `--max-sessions 1` setting.
4. Every pass reports only `root`/`child-a`, and `pending_children` is zero. Local status still has no `child-b`. The query repeatedly selects known `child-a`, which is skipped without moving the discovery boundary.

The same source-level failure occurs at the default allowance once the ordered known prefix fills that allowance. Increasing the budget may work around the trigger, but continuing bounded passes never makes progress. Add resumable parent-scoped discovery that eventually visits later child IDs, and report incomplete discovery while more children remain. This is a correction-introduced regression in the required exact, bounded inherited-child route; the existing unassigned-root isolation test does not reach it.

Evidence: `/tmp/al68-data-sol-02-children.py`, SQLite fixture `/tmp/al68-data-children-otdv20no/source.db`.

## Correction dispositions and refuting evidence

| ID | Scoped disposition | Evidence |
| --- | --- | --- |
| CAP-01 | Partially fixed; still blocking | `fillMissingCaptureSessions` protects canonical owners from retained fallback; later explicit extraction wins duplicate refreshed owners. The durable failed-registry/explicit tests reach those branches. Successful stale reread still loses canonical facts as reproduced above. |
| CAP-02 | Fixed for the assigned generation ordering/process-crash boundary | `runGenerate` acquires publication ownership before baseline load and all source reads, retaining it through ledger/report/summary output. Explicit writers and absent/different registries share the output lock. The key resolves parent/existing-leaf symlinks; persistent lock inodes avoid waiter/unlink races. Unix flock and Windows LockFileEx release kernel ownership on process exit. Ledger writes serialize a validated staged file, sync and rename; report writes stage/sync/rename. Reused the tester's overlapping registered/explicit native-publisher regressions. A report/summary failure is explicitly nonzero; a surviving stale report is digest-detectable and regeneration is the recovery path. |
| CAP-03 | Partially fixed; still blocking | Direct same-path known-other-ticket selection is rejected before publication and has a native regression. The reproduced same-store alias bypass remains. |
| CAP-05 | Fixed for failed snapshot parsing/accounting | `budget.charge(revision)` occurs before inspecting extraction success. Codex error/warning paths return the reserved source size; OpenCode snapshot preflight returns its reservation on extraction failure. Reused the two-malformed-owner test: only one source fits, the other remains pending, counters respect the shared bound. Metadata reads charge reservations before parsing. The separate child continuation finding above concerns work progression, not a recurrence of this specific parse accounting defect. |
| CAP-10 | Fixed for declared reconciliation semantics | The completion cutoff is assigned by defer after work; explicit semantics deny a common historical source snapshot. A successful owner keeps `refreshed_at`; failed/pending owners retain their previous successful checkpoint. Existing native smoke and fixture checkpoint ordering support the source trace. This does not claim ticket completion or a common timestamp for later explicit generation reads. |
| CAP-11 | Unassigned isolation fixed; inherited-child route still blocking | Extraction indexes and final report projection select the requested ticket. Unrelated unassigned roots stay locally visible without consuming that ticket's extraction allowance or publishing their IDs. Source-proven child inheritance checks source/store/native-parent evidence. OpenCode pagination nevertheless prevents eventual discovery in the reproduced bounded schedule. |

Traced actual producers through registration/read-modify-write, child recovery, exact source extraction, accepted/fallback result projection, explicit selection, canonical session merge and final ledger/report persistence. Registry transactions remain `BEGIN IMMEDIATE` with rollback on incomplete writes; successful source reads are compacted before persistence. No new canonical ledger schema or invented source totals were introduced by these corrections. These observations do not refute the three reproduced cases above.

## Verification and limits

- Independently built `/tmp/al68-data-sol-02-cli` with `go build -o /tmp/al68-data-sol-02-cli ./cmd/agentic-lens` from `backend`; exit zero.
- Recomputed the candidate fingerprint again after report/probes: unchanged `17fdae376119de798b382af56952351564389a69a174e5b0049259d387151cbc`.
- Executed both isolated native production-path probe scripts above; exit zero with the stated failing product outcomes. Fixtures use synthetic records and real filesystem/SQLite adapters; no shared source/test mutation or live transcript access occurred.
- Reused `TEST_REPORT.md` corrected six-package race evidence and its coherent source/test fingerprint, rather than rerunning the expensive suites. Inspected the new CAP-01/02/03/05/10/11 regression bodies and source branches; the three added probe triggers are outside their existing oracles.
- Reused parent's corrected full `make` PASS and embedded-asset byte verification in `VERIFICATION.md`, tied to the unchanged corrected candidate fingerprint. This is pre-integration evidence; required integration checks remain parent-owned.
- `SMOKE.md` records a fresh actual OpenCode 1.19.5 binding/tool/source round trip after correction, plus earlier independent-root/native-child source evidence. Those observations support their documented interface, not all versions/platforms/families or the new bounded-discovery schedule.
- Required real Codex trust/callback/agent-response evidence remains pending. No trust bypass, fabricated source success, or user trust request was attempted. Native non-Linux execution and broader runtime claims remain limited as documented. Even without these product defects, the outstanding required Codex evidence would prevent an overall delivery approval.
- **Scoped code verdict: CHANGES_REQUESTED. Overall handoff verdict: CHANGES_REQUESTED, with the separate real-runtime evidence gate INCOMPLETE.** Parent owns consolidation, correction assignment, subsequent candidate verification, telemetry and publication.
