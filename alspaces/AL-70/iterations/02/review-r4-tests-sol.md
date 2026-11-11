# AL-70 R4 test-effectiveness re-review

**APPROVED within the assigned UI-04/LIVE-03 scope.** No new blocking finding or unresolved scoped evidence gap. Run `r4-tests-sol`, re-review; requested Codex `gpt-6.1-sol` / `high`. Runtime model/effort and usage were not independently observed.

Reviewed AL70-R4 over base `f63e3db6b2d4baa8e22452e4ec94eff566867721`, frozen 270-input fingerprint `b656706936a5c83845a3390af67add958ba7a7a378385a01d5dcfbf1679c6d45`. All manifest entries, including deletions, matched at entry and exit. R3 findings were supplied correction context, not independent discovery credit. No sibling R4 conclusions were read.

The native regression at `backend/internal/adapters/live/partial_wire_contract_test.go:23` constructs isolated portable v6 history and a native Codex rollout, then calls real repository/LiveService HTTP detail and events routes. Before fixture equality, independent assertions establish partial owner coverage/diagnostics, one retained anomaly, projected detail roles, raw canonical invocation/anomaly roles, enclosing-owner restoration, no invented events and unchanged ledger bytes. Both actual HTTP payloads must match the shared frontend JSON pair without field normalization; no aggregate-count assumption substitutes for source completeness.

`TicketDeepDive.test.tsx:1812` consumes that exact pair through actual fetch helpers and validators, delays events, opens the raw baseline occurrence, and asserts one row, mapped roles and the same open dialog after partial hydration. Domain tests independently protect duplicate raw occurrences, distinct personas, per-occurrence evidence/severity, incomplete children, root/unknown owners, undefined versus empty categories, and input immutability. The selector uses raw facts and has no mapping argument that could reverse-guess aliases.

Compared R3/R4 test bodies and helpers: prior meaningful lifecycle, selection/rebinding, accessibility, role and absence assertions remain. `withAnomalyFacts` supplies explicit intended canonical facts to projected-only fixtures. The changed raw event `builder` now correctly displays `planner`; the distinct genuine raw `planner` displaying `other` remains independently guarded by `AnomalyDrawer.test.tsx:38`. Projected recorded-summary inspection remains separate. These changes follow the revised source contract; no deletion was justified by counts or scores.

LIVE-03 tests control deferred fetch responses and fake timers while invoking production `subscribeToLiveStream`. They cover newer-success/older-success, older failure after newer success, newer-failure/older-success, recovery, useful progress with newer work pending, teardown, cadence and callback-error handling. They neither sleep nor copy the subscription implementation.

Executed successfully:

- `go test -count=1 ./internal/adapters/live -run 'TestNativePartialAnomalyWiresMatchFrontendFixtures|TestNativeLiveHTTPWireMatchesFrontendFixture'`.
- Vitest unit config, `src/domain/events.test.ts`: 9 passed. The unit configuration excludes `liveApi.test.ts`; its checks were explicitly run with integration config.
- Vitest integration config, targeted `liveApi.test.ts`: 6 passed; targeted `TicketDeepDive.test.tsx`: 9 passed; `AnomalyDrawer.test.tsx` role-origin table: 3 passed. Other cases were intentionally deselected.

Disposable-copy mutations under `/tmp/al70-r4-tests-sol-0gdvqlt5/web` demonstrate useful failure oracles: removing accepted-completion ownership fails all three stale-poll cases; admitting only the latest-started request fails useful progress; appending retained source anomalies instead of reconciling fails the actual-pair UI guard with two rows instead of one. Shared source/tests were untouched.

Reused R4 recorded 217-unit/114-integration, typecheck, lint and canonical-build results. Source/Git/config/capture/publication code is unchanged from R3; this pass does not recertify those separate investigations. Parent owns candidate commit, main integration and full `make`; this scoped approval does not establish that delivery gate or authorize publication.
