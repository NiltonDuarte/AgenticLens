# AL-64 iteration 01 data review

- Reviewer role: independent data integrity and failure-handling review.
- Requested runtime identity: gpt-6-luna / medium. Runtime settings acknowledged as requested; no independent runtime metadata was available to assert.
- Candidate: `AL64-candidate-01`; baseline `a84adbe5be46c2ef52ede0987a417562715d38e1`; branch `fix/AL-64-active-turn-duration`.
- Candidate manifest: `/tmp/al64-review-01/fingerprint.json`, SHA-256 `6aab550b873d4b5de1242b889c26bb3b7c6838d4900ecae6d418b426c2a35695`; rechecked before verdict and unchanged. It lists the 46 application/test paths. Checkout contains 59 staged paths including work artifacts; this report is the only path this reviewer changed.
- Scope: duration evidence validation/persistence; source failure and unavailable history behavior; exact-owner replacement; relevant Codex/OpenCode producers and aggregation/read consumers. Reviewed only assigned data persona, shared review context/procedure, contract, iteration plan/investigation/runtime evidence/test report, technical guidelines, and relevant code paths. No sibling review reports were read.

## Findings

No confirmed data-integrity finding in the inspected scope.

## Consequential hypotheses and evidence

1. **Duration evidence could corrupt legacy or portable facts.** Rejected for the inspected paths. Runtime v6 decoding validates optional evidence and components against `active_duration_seconds`; legacy absence remains accepted and unknown. Migration decode/copy carries `DurationEvidence`, and invariant validation compares it. Existing numeric values are not rewritten to zero when evidence is absent. Aggregation derives provisional/unknown summaries from persisted invocation facts; HTTP DTOs carry the summary through. This is source inspection, supported by the reported schema/migration tests; I did not repeat suites.

2. **Invalid or incomplete source reads could delete prior owners or masquerade as authoritative empty snapshots.** No defect found in the traced paths. Backfill `recoverOpenCodeAt` retains stored sessions on nil DB, read/extraction errors, and owner identity/count mismatch; exact successful owner snapshots are returned for replacement. Codex parse warnings are retained as diagnostics and prevent that owner from being accepted as complete. Capture reconciliation’s exact extraction errors remain failed registrations and do not publish extracted facts. This follows the existing complete-owner refresh boundary in the portable ledger contract. The provided test report describes cross-source save/reload tests for unavailable owners and siblings; I reused that evidence, without rerunning it.

3. **Exact-owner replacement could append cumulative duration or alter unrelated children/harnesses.** No defect found by source trace. Existing refresh reconciles sessions by ID and recomputes aggregates over refreshed plus preserved ledger facts; duration is a per-invocation total with its source evidence, not an additive delta. Runtime evidence reports live v1/v2 generate/refresh/complete/repeat equality for the session arrays; the test report reports isolated persisted cross-family owner, child, and other-harness checks. These are bounded checks and do not establish every multi-machine scenario.

4. **Cutoffs or source timing could fabricate completed facts.** The new domain interval helper distinguishes an end after the snapshot cutoff as provisional and an absent end without current source-state evidence as partial/unsupported. Extraction paths set `CompletedAt` only from source completion at/before cutoff; Codex clears completion for provisional evidence. Existing generic `time_updated` is no longer used as assistant completion. Native runtime report observed live v1/v2 unfinished activity and later completion, and confirmed SIGINT can leave persisted v1 open markers stale. This validates source-state provisional behavior while also demonstrating that source fields cannot prove process liveness.

## Pending source-state policy boundary

The implementation counts explicitly open Codex/OpenCode source intervals provisionally through the recorded cutoff, with persisted cutoff and UI/CLI disclosure. The contract currently says authoritative unfinished activity may be measured provisionally, while the user clarification says the product decision on provisional source-state evidence versus required live-process proof remains pending. Native evidence also shows stale v1 open state after SIGINT and no persisted-field discriminator for that case. I do not resolve or reinterpret this policy. If the pending clarification requires live-process proof, the current source-state-only behavior would conflict and needs a new acceptance decision/change before approval. If provisional source-state intervals with disclosed liveness uncertainty are accepted, this behavior is consistent with the implemented model. This unresolved boundary prevents a final delivery approval.

## Verification and limits

- Manifest SHA-256 rechecked and matched the supplied value immediately before writing this report; baseline `HEAD` remained the supplied revision at review start.
- Reused supplied evidence: full Go race, frontend 834 tests, typecheck and lint pass at frozen candidate fingerprint; native OpenCode v1/v2 generation, `--recompute-db`, completion and unchanged-repeat probes pass as described in `RUNTIME_EVIDENCE.md` and `TEST_REPORT.md`.
- No full suites rerun. No disposable probe was necessary to substantiate a candidate defect after tracing the relevant replacement/validation paths.
- Native probes do not prove universal process liveness, Codex native end-to-end behavior, all multi-machine combinations, or absence of defects. The user’s source-state policy clarification remains pending. No production file or test was changed.

## Verdict

**INCOMPLETE** for delivery approval solely because the expressly pending source-state/liveness policy boundary is material to whether provisional duration is acceptable. Within the data-integrity paths examined, no substantiated blocking corruption finding was confirmed. Reassess that policy at the parent’s direction; this report does not claim whole-repository approval.
