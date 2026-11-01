# AL-68 data integrity re-review: data-sol-03

Scoped code verdict: **CHANGES_REQUESTED**. CAP-01 and CAP-03 are fixed for the assigned remaining reproductions. CAP-12 fixes progress within each queue, but a reproduced interaction still lets failed stop work permanently starve start-only child discovery. The separate required real Codex interaction gate remains **INCOMPLETE**.

## Revision, identity and boundary

- Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Frozen revision: `corrected-03`, `CANDIDATE-03.json`, fingerprint `26898644f86547fb9013c161120b291c1c6ed7f130307f214351b25c1741534d`. Independently recomputed the listed path/NUL/content/NUL digest before and after investigation; both match.
- Run: `data-sol-03`, final bounded affected re-review. Requested identity: `codex/gpt-6.1-sol/high`. Observed interface: Codex tools on the shared Linux checkout. Actual model/effort metadata and usage/cost are not independently exposed; requested settings are not asserted as measured runtime identity.
- Read the assigned data persona, common review procedure, contract/approved interaction clarification, corrections/findings, prior own Sol-02 report, source, affected regression bodies and implementation/test evidence. No sibling persona or sibling review conclusions were loaded. Product and tests remained read-only; only this report was written in the repository. Probes/binary/synthetic fixtures remain under `/tmp`.

## Remaining blocking finding

### CAP-12 remainder — P2: persistent stop failures permanently starve start-only child discovery

Locations: `backend/internal/adapters/capture/children.go:212`, `:227`, `:249`; `backend/internal/adapters/capture/reconcile.go:80`.

Within one proven Codex parent, queued stop paths always consume the recovery allowance before pending-start metadata discovery. Failed stop entries correctly remain durable. Their new rotation helps later stop entries, but never rotates between stop work and the start-discovery walk. With `--max-sessions 2`, reconciliation reserves one attempt for child recovery. One permanently malformed queued stop therefore consumes that one attempt on every pass; the `budget.sessions < 1` guard skips the pending-start walk forever. The same scheduling condition applies whenever unresolved stop work fills the recovery allocation, including larger/default allowances.

Native candidate-03 reproduction:

1. Register/bind a synthetic root to TEST-100 and reconcile its accepted source facts.
2. Deliver a SubagentStop whose absolute child path contains malformed metadata. Retain that unresolved stop entry.
3. Write a different valid child transcript with authoritative native parent `root`, and deliver only SubagentStart with an opaque, different interface agent ID. Do not deliver its Stop.
4. Run 20 successive `capture reconcile --ticket-id TEST-100 --max-sessions 2` passes.

Every pass exits zero and uses two total source attempts, but local registrations still contain only `root`; `pending_children` remains 2. The valid child never reaches the metadata walk. No guessed identity, invalid parent or unavailable source explains this outcome. Control: remove only the malformed synthetic stop-queue entry, preserve the same source/start evidence, and repeat the same bounded command. The exact child `child-only-start` is discovered on the third pass. The opaque interface ID is never substituted for its source ID.

Evidence: `/tmp/al68-data-sol-03-interactions.py`; failing disposable fixture paths are printed by the script. The initial verified failure/control fixture was `/tmp/al68-data03-interaction-bm50avg2`; its control removed the stop row after the 20 failing passes. A final replay retained another failing fixture without removing its queue. All records are synthetic. This is a remaining CAP-12/CAP-04 interaction in the correction path, not a new broad audit finding. Required correction: ensure eventual bounded progress between queued stops and start discovery while preserving unresolved evidence and the shared source limits. Existing `TestQueuedCodexChildRecoveryContinuesPastMalformedAndMismatchedMetadata` covers later queued stops, not this cross-queue schedule.

## Fixed cases and carried evidence

| ID | Disposition and evidence |
| --- | --- |
| CAP-01 | Fixed for the exact old-source replay and equal/unknown source-order boundaries. Native replay: registry source 378 bytes / timestamp 02 accepts 10; absent-registry explicit generation reads 549 bytes / timestamp 03 and publishes 30; restoring exact 378-byte source then successfully refreshing through the original registry keeps canonical 30 and reports pending. Old/new SHA256 remain `77747170045c7ab513d75b627e3f109cafecf444432b558a8fe5ffe6731f9c7d` / `94032149b7add1410ef9b288db690cb46e23649ae1ae508b6d57490e39e72123`. `filterCapturePrecedence` and `rejectOlderExplicitOwners` compare complete compact owner facts and source invocation completion, with no token maxima. Independent native equal/unknown canonical-checkpoint probes on both registered and explicit paths preserve the entire accepted owner; registered reports are pending and explicit output diagnoses unresolved checkpoint precedence. |
| CAP-03 | Fixed for actual-store aliases. Native Codex selected symlink conflict now exits nonzero without a conflicting ledger. Independent OpenCode SQLite probes verify selected and pre-normalization stored aliases in both symlink and hardlink cases: all four reject TEST-200 against known TEST-100 with no ledger. A separately copied real SQLite store reusing `root` is allowed in all four controls. `SameStore` resolves symlinks and uses `os.SameFile`; selection checks legacy rows, while `Store.Change` compares/migrates actual-store identities inside its binding transaction. |
| CAP-12 within-queue cases | Fixed for previously reproduced OpenCode known-prefix starvation and later queued Codex metadata. Native original OpenCode replay discovers the previously hidden `child-b` on bounded pass 8 with `--max-sessions 1`, preserving root/child-a. Source-ID paging advances before LIMIT and parent rotation persists. The Codex queued cursor advances before failed/mismatched reads; the tester's existing production-adapter regression reaches the valid later queued child while preserving three unresolved entries. That passing result is reused, with its oracle inspected. Cross-queue starvation above remains blocking. |
| CAP-02 | Prior fix carries forward: publication ownership is still acquired before canonical baseline load, reconciliation and explicit reads, and remains held through staged ledger/report/summary publication. New precedence filtering operates inside that same ownership boundary. No changed lock/write dependency invalidates the prior overlap/crash reasoning; reused candidate-03 overlapping-publisher regression and full race evidence. |
| CAP-05 | Prior failed-read accounting fix carries forward: exact extraction reserves/charges before inspecting failure; metadata attempts decrement and reserve before parsing. New cursors change progression, not charging. The demonstrated CAP-12 starvation respects the per-pass allowance and therefore is not a byte/session-accounting recurrence. Reused affected failed-read regression/race evidence. |
| CAP-10 | Prior cutoff fix carries forward: completion is assigned after reconciliation work, successful reads retain per-owner refreshed_at, and declared semantics exclude a common historical source snapshot. New publication filtering marks unresolved precedence pending rather than inventing a source completion. Reused checkpoint regression/race evidence; no new clock or publication-time authority was introduced. |

Native scripts: `/tmp/al68-data-sol-03-probe.py`, `/tmp/al68-data-sol-03-children.py`, `/tmp/al68-data-sol-03-interactions.py`, `/tmp/al68-data-sol-03-aliases.py`. Built `/tmp/al68-data-sol-03-cli` using `go build -o /tmp/al68-data-sol-03-cli ./cmd/agentic-lens` from `backend`; exit zero. Scripts executed successfully with the outcomes above. An initial unknown-checkpoint probe omitted the required DTO key and failed input validation; corrected it to explicit null before assessing product behavior. That fixture error is not a product finding.

## Verification and remaining limits

Reused the tester's candidate-03 six-package race evidence and stable fingerprints in TEST_REPORT.md. Reused parent full `make` PASS and unchanged candidate digest in VERIFICATION.md; no suite was duplicated. Parent reports origin/main fetched with no commits beyond the original baseline. No product changes, staging, commits, telemetry or publication were performed by this reviewer.

The recorded actual OpenCode 1.19.5 context/tool-binding/source round trip remains applicable because these corrections do not change its installed feedback protocol. It does not establish the new bounded schedules or untested versions/platforms. Required actual Codex trust/callback/agent-response/source acceptance remains unavailable; no bypass or trust question was attempted. Fixture success cannot satisfy that explicit acceptance boundary.

**Scoped code verdict: CHANGES_REQUESTED for the demonstrated CAP-12 remainder. Overall acceptance evidence: INCOMPLETE independently because real Codex interaction remains unverified.** Parent owns consolidation, correction assignment and delivery.
