# AL-70 R4 test-effectiveness re-review

**Scope and candidate.** Test-effectiveness re-review of UI-04 and LIVE-03 for AL-70 iteration 02. Base: `f63e3db6b2d4baa8e22452e4ec94eff566867721`; candidate: frozen 270-input fingerprint `b656706936a5c83845a3390af67add958ba7a7a378385a01d5dcfbf1679c6d45`. Requested profile: `gpt-6-luna/medium`; source-observed runtime model metadata was unavailable to this reviewer. Checked the candidate manifest: all present entries matched their listed hashes; its ten absent paths are explicitly marked deleted in the manifest. My conclusion is limited to the assigned tests and boundaries.

**Verdict: APPROVED.** No blocking test-oracle gap found for UI-04/LIVE-03.

## UI-04

`TestNativePartialAnomalyWiresMatchFrontendFixtures` uses a real schema-6 ledger, configured many-to-one role aliases and a native Codex session store with one missing owner. It starts the actual HTTP server and obtains both detail and `/events` responses. Independent assertions verify partial availability, required/complete coverage and diagnostic; canonical raw role/persona and owner restoration; the projected detail roles; the partial-source raw roles; one retained occurrence with no synthetic event; and unchanged stored ledger bytes. Full JSON equality against the frontend's two fixtures catches wire drift without normalizing away metadata. See [partial_wire_contract_test.go](../../../../backend/internal/adapters/live/partial_wire_contract_test.go:23).

The frontend integration test fetches the same fixture pair through the actual API client and validators into `TicketDeepDive`; it asserts exactly one row before/after hydration and that the open inspector retains its selection and projected display. Existing domain tests separately guard multiplicity, distinct raw personas under aliases, per-occurrence evidence/severity association, raw owner restoration, ownerless/unknown facts, absent-vs-empty categories, and partial/unavailable/complete semantics ([events.test.ts](../../../../web/src/domain/events.test.ts:38)). Ticket view cases preserve the recorded summary across failed, unavailable, partial and complete hydration. The explicit raw facts added to prior inconsistent projected-only fixtures are directionally correct: they state intended underlying persona facts and do not derive aliases in reverse. The wire test's single occurrence is complemented by domain tests with duplicates and distinct personas, so it is not being used as the sole multiplicity oracle. No score/count-based deletion was found; prior distinct UI assertions remain in the revised test file.

## LIVE-03

The actual `subscribeToLiveStream` REST fallback is exercised with controlled promises and fake-timer cadence, not sleeps or a mocked subscription implementation. Tests cover older success and failure after a newer success, older success after newer failure, sequential recovery, useful completion while newer work remains pending, and suppression of callbacks after unsubscribe ([liveApi.test.ts](../../../../web/src/api/liveApi.test.ts:175)). These assertions would detect stale tree/error/status replacement, latest-start starvation, and post-teardown callbacks. Existing cadence and invalid-policy cases remain separate.

## Verification and limits

Read the actual test code and linked consumers; reused the supplied candidate logs: frontend focused suite (217 units and 114 affected integrations), typecheck, lint, build; Go native partial-wire test PASS and HTTP peer PASS. The standalone result at `/tmp/al70-r4-standalone-zo06gqpb/RESULT.json` is complete and records unchanged ledger bytes. I did not run the repository-wide `make`; parent owns that integrated gate. No test edits or product writes were made.
