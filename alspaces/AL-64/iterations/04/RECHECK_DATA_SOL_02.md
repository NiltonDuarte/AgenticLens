# AL-64 iteration 04 data and failure correction recheck

Verdict: APPROVED within the bounded CP-02 correction scope. No new distinct finding. CP-REV-01 (severity 4/10) is resolved for source-open qualification; numeric extraction remains governed by iteration 03. This does not approve integration, publication or the whole feature.

Reviewer responsibility: data integrity and failure handling. Run identity: checkpoint_data_recheck02; phase: bounded correction recheck. Parent-assigned native identity: gpt-6.1-sol / high, inherited; runtime model metadata was not independently inspected by this reviewer. Product and tests remained read-only; only this assigned report was written in the checkout. No sibling persona or review conclusion was loaded.

## Revision and scope

- Baseline HEAD: `d8a7c14cfd9f0ee626fedc33c70f92bd928d9192`.
- Frozen CP-02 manifest: `/tmp/al64-04-review/fingerprint-02.json`; SHA-256 `512721e4192e55dcad98f2f1d7a080360178b73ce016c0225b50840dc69ff629`. All 411 entries matched at entry and at final verification.
- Full candidate `/tmp/al64-04-review/candidate-02.diff`: SHA-256 `b426a55c735e39e4331ed5e1b03bd0095d7283c472bf64d484e2301b506486f9`. The live `git diff HEAD -- backend docs/SETUP_MANUAL.md` has the same hash.
- Three-path correction `/tmp/al64-04-review/correction.diff`: SHA-256 `c8cd7bf243b285d68b50d095538da34b9ddbaa68e4c230700fa73342f8848522`.
- Independently reproduced tester-compatible source/test fingerprint: 403 paths, `29b4e1215134acdc5baa39caed9bc88d68f329b9eff5b17ab49ef2444553b6ba`.
- Correction paths: `backend/internal/adapters/opencode/schema_v2_session_extraction.go`, `backend/internal/adapters/opencode/checkpoint_test.go`, `backend/cmd/agentic-lens/checkpoint_generate_test.go`.

Read approved iteration 04 PLAN, correction brief, implementation/test/runtime reports, assigned persona/shared procedure, technical portable-ledger contract and applicable refresh/failure regression boundaries. Initial working-tree changes belonged to the shared AL-64 work and were preserved.

## Consequential hypotheses and disposition

1. **Native idle still certifies a stale running tool as source-open: refuted.** At schema_v2_session_extraction.go:121 the optional lookup is separate from the existing active-session query. At :287-293 only `CheckpointInterval.SourceOpen` consumes its result. Absent optional column preserves older behavior; native NULL and a positive idle strictly before the message permit existing proof. Equal/later idle, including after the numeric cutoff, disqualifies source-open wording. Text, fractional, zero and negative metadata cannot be interpreted as active. A long interval then yields F04 through checkpoint.go; its ID/start remain available for exact known-short suppression. No numeric cutoff/DurationInterval, token/cost, owner/model/row or completion assignment changes in this correction.

2. **Unavailable optional metadata rejects extraction or erases measured facts: refuted for the new lookup.** schemaV2CheckpointIdle at :482-511 returns nil on query, scan, iteration or close error; it does not return an extraction error. Missing entries never establish active state. The existing mandatory numeric extraction failure paths remain unchanged. Three isolated Go overlays exercised query failure, no matching optional rows and Scan failure through the public strict extraction/selector path. All passed the adapted native-idle table: successful extraction retains active180/provisional180/input17 and the owner/model/row; long optional-column cases qualify to F04, absent-column control stays F03, known179999 cases stay silent. Iteration/close error handling was inspected statically rather than dynamically injected.

3. **Qualification becomes an accounting or ownership gate: refuted.** DurationInterval is evaluated and accumulated before the feedback interval is appended. The new map does not control activeSessions, running, projectSourceFacts, authoritative replacement or invocation values. ExtractSessionsStrictWithSourceAt and capture snapshot callers append transient evidence only after successful extraction. Complete-owner MergeSourceEvents replacement removes older evidence for refreshed owners. Retained ledger DTOs are preserved by retainUnrefreshedGenerationOwners; retained owners receive no invented native evidence and can use F04. No new receipt, local-source, provenance or authenticity prerequisite was introduced for portable history.

4. **Later separate-message supersession can regain exact feedback: guarded.** The existing v2 `mi == len(msgs)-1` qualification remains; both-family added cases reach a real later native user record after cutoff while retaining numeric180/input17. The tester's mutation log shows both long assertions fail with F03 when existing supersession guards are removed; short controls remain silent. CP-TEST-01 (severity 3/10, test-effectiveness reviewer discovery) is verification coverage, with no product fix required.

5. **Feedback errors change publication success: refuted within the inspected path and reused subprocess evidence.** main.go:709 calls the selector/writer after ledger/summary journal commit and acknowledgement, only for explicit generation, using an empty caller proof. A writer error is reported without a new failure exit. checkpoint_generate_test.go:54 opens stderr with os.Open, and :59 now proves the inherited descriptor rejects writes before the child runs. Root privilege does not make an O_RDONLY descriptor writable. Existing publication failures still exit before feedback. No ledger/summary/API/frontend payload or persisted feedback state changes in the correction.

## Verification evidence

Reused the tester's final CP-02 evidence after reproducing its source/test fingerprint; no duplicate broad suite was run:

- `/tmp/al64-04-tester-corrections/pre-fix.log`: original source overlay fails eight native-idle qualification cases with F03 versus expected F04; qualifying controls pass.
- `/tmp/al64-04-tester-corrections/supersession-mutant.log`: expected failures for v1/v2 future superseding long cases.
- `/tmp/al64-04-tester-corrections/green-focused-race.log`: affected OpenCode and CLI race tests PASS (1.713s / 7.972s), including the stderr descriptor preflight and publication boundary.
- `/tmp/al64-04-tester-corrections/opencode-race.log`: full OpenCode race package PASS (6.156s).
- TEST_REPORT.md records correction `make test-unit` PASS, including 190 frontend units; unchanged CP-01 capture/config/portable-history evidence is carried forward with its original scope, not represented as re-executed here.

Reviewer-executed isolated probes, from backend, for each `query_failure`, `missing_rows`, `scan_failure`:

`go test -overlay /tmp/al64-04-data-sol-recheck02/<variant>.json ./internal/adapters/opencode -run '^TestOpenCodeV2CheckpointIdleBoundaryPreservesNumericFacts$' -count=1`

All three returned exit 0 (0.022s, 0.022s, 0.024s). Overlay copies are outside Git. The production copy perturbs only the optional metadata query; the copied table changes only NULL/old-positive expected feedback to F04 under forced unavailability. Its numeric/identity assertions and absent-column/known-short controls are retained. These are synthetic failure probes, not native runtime observations or new product defects.

## Limits and handoff

This approval addresses the three-path correction and its data/failure dependencies. Native generic caller/current-interval handoff and natural-completion/next-turn refresh remain unverified as documented in RUNTIME_CAPABILITIES.md. F03 is source observation, not process liveness; F01/F02/F05 typed tests do not enable native support. Existing numeric running-tool treatment is intentionally unchanged and is not a newly approved accounting rule. Parent retains integrated make, standalone asset/binary checks, consolidation, final approval and delivery ownership. No new unresolved data/failure defect or material evidence gap was found within this bounded scope.
