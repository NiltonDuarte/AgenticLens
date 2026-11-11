# AL-70 R4 behavior re-review — reviewer-behavior / `r4-behavior-luna`

**Outcome: APPROVED (scoped).** Candidate AL70-R4, base `f63e3db6b2d4baa8e22452e4ec94eff566867721`; 270-input fingerprint `b656706936a5c83845a3390af67add958ba7a7a378385a01d5dcfbf1679c6d45` verified at entry and exit. Requested profile: Codex `gpt-6-luna` / medium; runtime identity was not independently observed. This is the assigned behavior/integration re-review of UI-04 and LIVE-03 only, not whole-repository approval.

## Behavior traced and hypotheses

- **UI-04 partial detail/events pair:** Followed native Codex partial hydration through `partial_wire_contract_test.go`, the actual `/api/workspaces/{id}` and `/events` JSON pair, frontend `decodeWorkspaceEvents`, `TicketDeepDive`, anomaly entry selection, `TraceWaterfall`, and drawer selection. The pair preserves the raw builder/source-tester anomaly under its enclosing missing owner while the detail endpoint's projected planner/tester summary remains in its distinct recorded-summary section. One-to-one occurrence reconciliation preserves repeated identical findings; severity is associated with the same raw occurrence, evidence enrichment does not overwrite raw context, and role aliases are not used as identity. Unknown/root and missing owner facts remain retained. Initial/pending, partial, unavailable, failed enrichment, and open selection/rebinding preserve stored facts and selection; failed/absent hydration does not fabricate trace rows. Existing raw ledger bytes remain unchanged in the producer test.
- **LIVE-03 fallback polling:** Traced monotonic poll IDs and `lastAcceptedPoll` through success/error callbacks and timer teardown. A stale older result/error cannot replace a newer accepted result/error; an older useful completion still applies if a later request is merely pending. Subsequent current recovery is accepted, and post-unsubscribe completions are suppressed. The fetched configured cadence and existing missing/invalid/unsupported policy errors remain on the same path. This carries forward the prior cadence/policy flow because this correction changes completion ordering only.

No blocking defect found in these scoped flows. R3 counterexamples were treated as correction context, not new discovery credit. No unrelated source defects were encountered.

## Evidence and limits

Read-only evidence: `/tmp/al70-frontend-r4-focused.log` (217 unit and 114 focused integration tests passed); frontend typecheck, lint and production build logs passed; `/tmp/al70-ui04-wire-final.log` (`TestNativePartialAnomalyWiresMatchFrontendFixtures` and live HTTP wire test passed); `/tmp/al70-ui04-http-peer.log` passed. The standalone smoke result at `/tmp/al70-r4-standalone-zo06gqpb/RESULT.json` records five cases passing, unchanged ledger bytes and matching embedded JS bytes; this is corroborating runtime evidence, not a repeated test run. The build emitted only the existing large-chunk advisory.

Parent-owned integrated `make`, main integration, and final delivery remain outside this scoped verdict. No additional tests were run. Source-session/model identity was not observed. Requested model/effort above is dispatch context only.
