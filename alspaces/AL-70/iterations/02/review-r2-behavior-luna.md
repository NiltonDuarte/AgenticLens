# AL-70 R2 behavior re-review — r2-behavior-luna

**Requested reviewer:** Codex / gpt-6-luna / medium (runtime identity not independently observable). **Candidate:** AL70-R2, base `f63e3db6b2d4baa8e22452e4ec94eff566867721`, source fingerprint `712d1b3db91b641ed2927f75081dd299a61aec84b3b907d69874c5d7614545c2`. **Outcome: APPROVED for behavior/integration scope.** No blocking finding or distinct candidate found; this is not a whole-repository approval.

## Findings

None. Re-reviewed the R1 behavior candidates CAP-02 and DATA-01, and the specified CFG-01, CAP-01, PATH-01, LIVE-01, UI-01, UI-02, FE-01 and CTX-01 boundaries. No new source-model, producer-count or A12 requirement was applied.

## Evidence and disposition

- CAP-02: traced worker admission through reconciliation to child `generate`. The child receives the same validated `MaxSessions` and `MaxSourceBytes` as explicit CLI flags, preserving precedence over local source budgets without config writes. The fresh-source publication probe in `/tmp/al70-r2-budget-final.log` passed (3.584s), including absent/competing local caps and invalid overrides.
- LIVE-01: `GetSessionTree` returns `ErrSourceUnavailable` when the selected hook is unavailable and returns an empty tree for a reachable source with no active rows. HTTP/SSE check availability, subscribe before the initial snapshot, and the client falls back to REST polling after SSE errors; UI state remains null/unavailable until a valid tree arrives. This preserves recovery when a source later becomes available. The handlers/service preserve startup policy validation. No live machine integration was exercised.
- UI-01/UI-02: the live service resolves the configured mapping before serializing the node role; LiveSessionTree and LiveTraceInspector format the projected value without applying the mapping again. Source-role consumers still use `sourceLabel`. ArtifactViewer consumes only `items`; `null`, `""`, whitespace, and nonempty content have distinct unavailable/empty/whitespace/render states, and empty content remains copyable.
- FE-01: repeated prefix query values bypass the `all` sentinel filter while other multi-value parameters retain it. The HTTP parser recognizes complete configured/observed literal prefixes before CSV splitting, including case-folding under configured policy.
- DATA-01/CTX-01: recursive discovery reports descendant walk errors while preserving reachable siblings; batch output renders warnings/errors even when no target ledger yielded a result. Summary-only publication resolves required context policy before publication ownership/writes; absent-policy summary controls remain source-independent.
- CFG-01/CAP-01/PATH-01: traced structural YAML-node validation, shared YAML/YML local selection into manager callback flow, and NUL-delimited Git metadata parsing against the R1 counterexamples. I found no remaining counterexample on these entrypoints. These are source-trace conclusions; I did not rerun their bespoke probes.

Reused supplied focused evidence: `/tmp/al70-r2-final-focused.log` (CLI/config/manager/live/HTTP/backfill packages passed), `/tmp/al70-data01-gscr5d4b/final-test.log` (backfill package passed), `/tmp/al70-frontend-r2-integration.log` (497 tests passed), and supplied frontend typecheck/lint/build logs. Parent-owned full `make` remains outside this scoped pass. Earlier R1 boundaries not changed by R2 were carried forward only where the corrected call path and guard were directly inspected above; no blanket R1 verdict is implied.

## Limits

No live developer stores, machine installation/activation, broad duplicate suite, or full `make` was run. Verification is bounded to the corrected candidate fingerprint, supplied disposable-fixture evidence, and the production call traces described here. Manifest fingerprint was confirmed on entry and exit; excluded review artifacts do not alter the source candidate.
