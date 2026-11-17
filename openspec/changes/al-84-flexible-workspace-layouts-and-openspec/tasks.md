# AL-84 iteration 01

Implementation and verification are complete. Explicit user approval: “$agentic-lens-dev plan and iteration 1 is approved.” [Design](design.md) owns the approved contract and all material steering. Baseline `163181428ee9a5a599a2da1e91d3af5e618a9682`; feature branch `AL-84/flexible-workspace-layouts-openspec`, initially clean. All user refinements and corrections remain in iteration01.

- [x] Record scope approval before coder dispatch and preserve isolated writable-path ownership.
- [x] Implement explicit roots, bounded ledger-only discovery, canonical recorded identity and common selection across readers, generation, capture, formatting and maintenance.
- [x] Implement private publication state, portable pending detection, durable recovery and request-scoped physical read binding without workspace registration or identity sidecars.
- [x] Adopt pinned OpenSpec execution changes while preserving separate roadmap authority, historical homes, custom setup and PO/development role boundaries.
- [x] Verify portable history, aliases, duplicates, arbitrary/external moves, interrupted publication, native Codex/OpenCode source publication, archive/late refresh and standalone embedded reads.
- [x] Complete behavior/data/architecture/tests reviews on both configured profiles, correct all findings and retain model-specific discovery credit. Test-only integration corrections received narrow review on both profiles; unchanged runtime approvals carry forward.
- [x] Commit the candidate, integrate main and pass full make at `273d90aa16966fe34984b78543fc42cdb6e0c7c0`, plus final OpenSpec and standalone checks.
- [x] Extract final telemetry, verify all expected source owners and prepare the final artifact commit and PR handoff.

[Delivery](delivery.md) owns the tested revision, outcome, verification and publication context. [Test evidence](evidence/iterations/01/TEST_RESULTS.md) retains actual failures, corrections, current passes and limits; [compact findings](evidence/iterations/01/REVIEW_SUMMARY.json) retains every canonical finding and reporter tuple. The parent publishes the prepared commit and PR and links the actual URL in the handoff. No merge is authorized.

Hosted Pages caller follow-up: both behavior profiles approved explicit Pages configuration plus CLI precedence; old/new callers and11 site/9 preview tests passed. Core sources remain unchanged. The parent refreshes telemetry and watches PR144 checks after pushing the correction.

## Requested Astra review

The user requested a fresh Astra review and fixes for any issues found. This is iteration 01 follow-up work on PR144, frozen at `2b4118736b5bde24d17cc382da7445c8046ce884` against `163181428ee9a5a599a2da1e91d3af5e618a9682`. For this request, Astra/high replaces the default paired model profiles without changing repository policy. Product and test files stay frozen during independent discovery.

- [x] Behavior first pass: `CHANGES_REQUESTED`; pending publication can disappear from implicit uniqueness/first-publication admission. Parent reproduced both failures. Report: `evidence/iterations/01/reviews/astra-behavior.md`. Correction and recheck remain open.
- [x] Data first pass: `CHANGES_REQUESTED`; export/summary publication loses the selected physical directory between load and write admission. Parent reproduced same-ticket replacement writes with the stable control passing. Report: `evidence/iterations/01/reviews/astra-data.md`. Correction and recheck remain open.
- [x] Architecture first pass: `CHANGES_REQUESTED`; independently found pending identity exclusion and reproduced inconsistent discovery/selection depth for in-root symlinks. Parent reproduced the alias case; discovery advertises a home that explicit selection rejects. Report: `evidence/iterations/01/reviews/astra-architecture.md`. Correction and recheck remain open.
- [x] Test first pass: `CHANGES_REQUESTED`; independently reproduced pending-duplicate selection and identified the missing fully available-root regression boundary. Report: `evidence/iterations/01/reviews/astra-tests.md`.
- [ ] Parent reconciles findings, routes necessary fixes and tests, verifies the corrected revision and publishes current evidence. Prior passing checks remain evidence only for unchanged code.

Three distinct findings are confirmed. Pending identity uncertainty must block implicit uniqueness and first-publication absence while preserving invalid-ledger exclusion, partial listing and explicit healthy selection. Artifact publication must retain physical selection through write admission without persisting identity or blocking writers during ordinary reads. Discovery must apply the same physical depth semantics as explicit selection/publication, including in-root aliases and overlapping configured roots. These are corrections to approved acceptance, not a new iteration or scope boundary. Coder owns affected production Go paths; tester owns regression tests after the API correction; parent owns artifacts and integration.

Corrections and regression tests are implemented. Affected ordinary/race checks pass, including five guard mutations failing at their intended assertions. One broad callback deadline/SQLite rollback failure did not reproduce in six targeted runs or the broad rerun; retain it as an unexplained reliability observation. The corrected source is frozen in the private candidate content manifest for Astra/high rechecks: behavior, data, architecture and tests each own `reviews/astra-<focus>-recheck.md`. Parent integrated make and publication remain pending.
