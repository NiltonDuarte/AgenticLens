# AL-70 R3 test-effectiveness review

**Run:** `r3-tests-luna`; requested/dispatch-acknowledged identity `codex / gpt-6-luna / medium` (runtime identity was not independently observed). **Candidate:** AL70-R3, base `f63e3db6b2d4baa8e22452e4ec94eff566867721`, fingerprint `ee805c71aae0b2d57ba393f31fe49a1506a71afdc70f6e38ed1441d161f51550`. I verified all 265 manifest entries and the canonical fingerprint at entry; no mismatches. Scope was test effectiveness for UI-03, LIVE-02, POLL-01 and preserved R2 test guarantees.

## Verdict: APPROVED

No blocking test-effectiveness defect found in this scope. Tests exercise the relevant callers and distinguish the source of values:

- UI-03 has domain-level selection/merge/rebind tests for projected recorded anomalies versus raw hydrated anomalies (including retained unavailable owners and the genuine raw `planner` alias case), plus TicketDeepDive tests covering pending, failed, unavailable, partial and complete hydration, mixed source-event/recorded-anomaly selection, and summary rebinding. The drawer assertions inspect both raw event-role fallback and anomaly context roles. These catch role-origin loss or inferring origin from role text.
- LIVE-02 is tested through actual REST responses: a 503 followed by a valid 200 must render the returned empty tree, restore “Fallback Polling (REST)” and remove the stale error banner. The API test uses fake timers to establish the configured interval boundary and confirms unsubscribe prevents later requests and callbacks.
- POLL-01 tests the actual App fetching `/api/config`, entering Live Observability, and scheduling exactly 333 ms; invalid/missing/unsupported direct timer values start no timer or fetch. Go HTTP tests exercise configured 333, absent policy, and source-free absent `live` with saved-ledger byte preservation. This pairs with source-level validation that static persisted views do not require the optional live field. Standalone evidence independently exercised real binary/API behavior and confirmed unchanged ledger bytes.

R2 coverage remains relevant because its contract paths are unchanged: App policy handoff, producer/HTTP wire preservation, and byte-preserving read behavior. The R3 additions strengthen UI recovery and cadence-through-App without deleting those prior oracles. No count-based coverage inference or artificial source authority was used. The corrected backend recomputation/double-projection hypothesis remains refuted by the existing producer path; it is not reported as an unresolved test issue.

## Verification and limits

Reused supplied logs: `/tmp/al70-frontend-r3-focused.log` (216 unit and 159 affected integration cases pass), `/tmp/al70-frontend-r3-mixed-selection.log` (final TicketDeepDive recipient: 60 cases pass), `/tmp/al70-frontend-r3-typecheck.log`, `-lint.log`, and `-build.log` (all pass; build reports its existing >500 kB chunk advisory); `/tmp/al70-r3-http-poll.log` (`go test` full HTTP package pass, 1.196 s); and `/tmp/al70-r3-standalone-_89smbrx/RESULT.json` (12 disposable ledgers, config/absence/static-view/port cases, unchanged recorded bytes). R2's own prior report and focused test evidence were carried forward only for unchanged dependencies. I did not rerun broad suites; parent owns integrated `make`. This verdict covers the assigned test focus and does not certify the full candidate.
