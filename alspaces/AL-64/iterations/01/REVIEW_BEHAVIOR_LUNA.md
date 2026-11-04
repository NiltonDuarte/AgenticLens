# AL-64 iteration 01 behavior review

## Identity and scope

- Persona: reviewer-behavior
- Requested runtime: gpt-6-luna / medium (acknowledged; no claim about independently observed runtime metadata)
- Candidate: AL64-candidate-01; branch `fix/AL-64-active-turn-duration`; baseline/HEAD `a84adbe5be46c2ef52ede0987a417562715d38e1`
- Frozen manifest: 46 paths; `/tmp/al64-review-01/fingerprint.json` SHA256 `6aab550b873d4b5de1242b889c26bb3b7c6838d4900ecae6d418b426c2a35695`; independently recomputed all 46 file hashes with zero mismatches. Candidate diff supplied at `/tmp/al64-review-01/candidate.diff` SHA256 `753dcd6bbadc5777bc27aaa5b2ec14b7b27e10f9ed298dd98e5981de94b34634`.
- Scope: generation, capture reconciliation, source-backed backfill, aggregation and visible duration totals/cutoffs across Codex and OpenCode schema v1/v2, including relevant unchanged callers/consumers.
- Read only the assigned behavior persona, shared review context, canonical review skill, technical guidelines, AL-64 contract and iteration plan/investigation/runtime/test records. Did not inspect sibling review reports or their conclusions.

## Finding

No blocking behavior finding confirmed in the scoped candidate. No severity-rated findings.

## Consequential hypotheses investigated

1. **A single operation cutoff might be lost between entry points and source adapters.** Refuted by source trace. `runGenerate` captures a cutoff at entry and passes it to Codex rollout parsing, OpenCode strict extraction and capture reconciliation (`backend/cmd/agentic-lens/main.go`; `backend/cmd/agentic-lens/capture.go`). `BackfillEngine.Execute` resolves one cutoff and passes it to Codex recovery and OpenCode snapshot extraction (`backend/internal/adapters/backfill/engine.go`, `recovery.go`). The command-line wrappers without an injected cutoff take their own boundary clock, as expected.

2. **Codex active turn time could still be omitted or completion could be double counted.** Refuted for the inspected path. `readEvent` pairs turn starts with completion/failure, closes interrupted/aborted lifecycles at their terminal timestamp, and `finalizeFacts` adds only still-open starts through the snapshot cutoff. Model-specific accumulation is rounded at the invocation boundary and stored as separate completed/provisional components. Completion removes the open start, so a later extraction recomputes rather than adding to an old provisional snapshot. This preserves turn accounting, including tool time within that turn.

3. **OpenCode v1/v2 could disagree on active-vs-completed state or use an update timestamp as completion.** Refuted in the current extractor paths. V1 derives activity from an unfinished native tool or open timed text/reasoning part on the latest message. V2 uses `time_idle IS NULL` or a running tool, bounded by a later message at/before the cutoff. Both route assistant intervals through the shared `DurationInterval`; v2 no longer substitutes `time_updated` for a missing completion. The native runtime report records successful active generation, `--recompute-db` refresh and completion/repeat refresh in each family, while also documenting the v1 empty auxiliary table correction. Those runs support the path but do not prove every source state.

4. **Visible aggregate duration could omit evidence or imply a fabricated common cutoff.** Refuted for the inspected outputs. Go workflow, project/fleet and progression aggregations merge per-invocation evidence; cutoff values are retained as a sorted distinct set. The CLI/Markdown workflow summary uses `DurationDescription`; ticket detail and ticket rows label incomplete/provisional values and expose explanations; fleet and comparison summaries expose aggregate coverage/cutoffs as supporting text. The numeric totals and sorting remain cumulative additive seconds, consistent with the contract. Comparison KPI supporting text is cohort-level coverage disclosure for an average KPI; it reports the filtered cohort's provisional seconds and cutoffs rather than claiming those values are the average itself.

5. **Refresh could add snapshots, mutate unavailable-owner cutoffs, or drop unrelated facts.** The inspected recovery path replaces successfully extracted exact owners and retains stored owners when discovery/extraction fails; the invocation and workspace aggregations then recompute from the merged sessions. Existing portable-ledger preservation behavior remains at the owning recovery boundary. The AL-64 test report records the targeted matrix for fixed/advancing cutoff, replacement, completion, repeat, children and other-harness facts. I did not independently rerun the full suites.

6. **Persisted optional evidence could make older facts unreadable.** The contract's reader-first rollout is supported by the inspected DTO/domain/API path: absence of duration evidence remains accepted and summarized as unknown; newly written evidence has validation and frontend boundary validation. Pre-change strict-reader deployment compatibility remains an operational rollout constraint in the contract, not a behavior defect in this candidate. No release is authorized by this review.

## Verification and limits

- Manifest exactness checked immediately before review; no stale candidate evidence observed.
- Read source/diff and traced supported paths through generation, refresh, aggregation and UI/CLI consumers. New domain/frontend duration files were read as part of that trace.
- Reused the supplied successful full Go race, frontend, typecheck, lint and native OpenCode runtime evidence. No full suite rerun; review procedure favors targeted evidence and the supplied unchanged fingerprint is exact.
- The user clarification about persisted open source-state intervals being provisional with disclosed process-liveness uncertainty versus requiring live proof remains pending. This review does not resolve it. The inspected semantics count open source-state intervals provisionally and the UI/summary disclose their stored cutoffs and provisional status; liveness uncertainty remains a known limitation under the pending policy.
- Native Codex active generation/capture/refresh was not independently exercised in this review. AL-64 supplied evidence and the deterministic regression report do not substitute for an unobserved live Codex runtime claim.
- No claim that all possible native source states are covered, that the entire repository is defect-free, or that this authorizes release/merge.

## Outcome

**APPROVED** for this scoped behavior review at the exact candidate revision above. No blocking behavior defect was established; the scoped integration paths are coherent with the approved contract, subject to the explicitly pending source-liveness policy and the verification limits above.
