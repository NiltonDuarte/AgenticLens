# AL-64 iteration 04 data and failure review

## Outcome

**APPROVED — scoped data/failure review.** No blocking data-integrity finding was demonstrated in the frozen candidate. This is not approval of unverified native caller or continuation capabilities, nor of the parent-owned integrated delivery gates.

## Reviewed candidate and evidence identity

- Requested focus: data integrity, evidence freshness/identity, retained owners, publication/failure handling, portable facts and configuration boundary.
- Baseline: `d8a7c14cfd9f0ee626fedc33c70f92bd928d9192`.
- Candidate: shared worktree source/test/config/guidance paths in `/tmp/al64-04-review/fingerprint.json`; manifest SHA-256 `c766dd83ec3e9f601e6e5b302422defa2bbe6c86f5906648cae1e272a338a32d`. Independently checked all 411 listed paths: zero missing, zero mismatched. Supplied `/tmp/al64-04-review/candidate.diff` SHA-256: `9afb94b071ce4f74e32fad4ea409f1887899d2c3b9a7f3057fa3f4695d41da82`.
- The brief did not supply a review run ID. Persona/model request was gpt-6-luna / medium; runtime metadata for this review was not independently exposed, so no actual usage attribution is asserted.
- Read the approved iteration 04 plan, implementation, tester report, runtime capability report, data reviewer persona/shared context, portable ledger contract and regression cases. Did not read sibling review conclusions or parent consolidation.
- No product or test files were changed. This report is the only file written by this reviewer.

## Findings

None. No severity score applies because no concrete defect is being reported.

## Data hypotheses investigated

1. **Retained facts are mistaken for fresh native proof.** Refuted in the inspected path. `MergeSourceEvents` retains prior checkpoint evidence only for owners not marked complete by the fresh read; retained ledger owners with no matching fresh evidence reach the selector as unresolved and can only produce qualified F04 from their own stored provisional subtotal. The capability report separately distinguishes genuine Codex source-open evidence from caller identity and process liveness. OpenCode intervals are labeled assistant-message intervals, and current-source openness is qualified independently from timing.
2. **Evidence can be matched to the wrong model or invocation after grouping/replacement.** No counterexample found. Evidence is captured per source/session/model and one-based invocation row before model grouping is discarded; the selector requires source/session/row, model, and exact cutoff to match. Complete source owners replace their prior facts and evidence together. Duplicate evidence row keys become ambiguous rather than authoritative. Existing selector/adapter tests exercise mismatches, duplicate rows and multiple owners/models.
3. **Partial, future, terminal or stale source records create exact-current advice.** No false-positive path was demonstrated. The selector requires valid native kind/ID/start, start no later than the explicit cutoff, and `SourceOpen` for exact F03. Exact caller cases additionally require a matching independently verified interval. Codex native IDs are not reconstructed from accounting IDs; later historical activity disables source-open wording. OpenCode v1/v2 mark superseded/future/partial cases conservatively and retain qualified subtotal fallback where allowed. Complete known-short timing is checked before source-open qualification, so a 179999ms interval cannot become F04 merely because it is no longer open.
4. **Optional feedback failure can corrupt or falsely acknowledge publication.** Refuted by call ordering in `main.go`: feedback runs only after requested publication and its diagnostic acknowledgement succeed; its writer error is logged as optional and does not change success. The tester report records a real generate subprocess with read-only stderr, successful publication, publication-failure suppression and stdout/ledger/summary/API contamination checks. I reused that evidence rather than rerunning broad tests.
5. **The change introduces a private-state gate on portable history or fabricates empty/zero source evidence.** No such gate found. Checkpoint evidence is transient (`json:"-"`) and not added to ledger/API/summary persistence. Missing evidence results in qualified output only when that owner’s recorded subtotal crosses the configured threshold; it does not erase stored facts. Configuration remains optional/defaulted and strictly decoded; threshold validation is separate from history acceptance. Existing portable-ledger and unavailable-owner tests are cited by the tester report, but this review did not rerun migration or full portable suites.

## Verification and limits

- Independent verification performed: manifest byte/hash verification for all 411 paths; source trace across Codex/OpenCode extraction, `SourceEvents` merge, capture reconciliation, final ledger selection and publication ordering; `git diff --check` passed.
- Reused from the final tester report: public `make test-unit` (including 190 frontend units), frozen affected source/config/capture race suite, focused generate CLI race suite, and `git diff --check`. No duplicate broad suite was run.
- The runtime report's installed versions and source observations support only source-open evidence. Generic generate caller/current-native-interval handoff and a real natural-completion-to-next-turn-refresh cycle remain explicitly unverified. Tests of typed F01/F02/F05 selectors do not establish native support; production F01/F02/F05 remain gated as described in the plan.
- Parent integrated `make`, standalone binary verification and review consolidation remain outside this scoped approval.
