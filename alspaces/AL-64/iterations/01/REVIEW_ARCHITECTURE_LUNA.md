# AL-64 iteration 01 architecture review

Reviewer: gpt-6-luna / medium (runtime settings acknowledged). Role: independent architecture discovery.

## Revision and scope

- Branch: `fix/AL-64-active-turn-duration`
- Reviewed base/candidate revision: `a84adbe5be46c2ef52ede0987a417562715d38e1` (candidate changes are present in the shared checkout as a working-tree diff).
- Frozen manifest: `/tmp/al64-review-01/fingerprint.json`, SHA-256 `6aab550b873d4b5de1242b889c26bb3b7c6838d4900ecae6d418b426c2a35695`; all 46 listed paths matched their recorded SHA-256 values before review. The branch HEAD equals the supplied baseline.
- Scope: cutoff and source/projection responsibility boundaries; bounded schema-v6 additive evidence; relevant unchanged consumers and source readers. Required OpenCode runtime evidence and full-suite outcomes were read as context, not independently repeated.

## Finding

### Severity 5 — v1 activity proof can be suppressed by a post-cutoff message

**Location:** `backend/internal/adapters/opencode/schema_v1_session_extraction.go:195-201`.

**Trigger:** During an extraction whose cutoff is `C`, the source contains an assistant message with an unfinished native activity part observed by `C`, plus a newer message whose `time_created` is after `C`. This can occur when source data advances after cutoff capture or from future-dated source timestamps.

**Expectation:** The contract says duration is computed as of the supplied cutoff and future activity is excluded. A message created after `C` cannot supersede activity at `C`.

**Consequence:** The `NOT EXISTS` guard considers every newer row, without `newer.time_created <= C`. It therefore omits the older message from `runningMessages`. At `schema_v1_session_extraction.go:383`, `DurationInterval` receives `ongoing=false`; an unfinished interval becomes unsupported and contributes no provisional seconds, even when its open part has a start and update at or before `C`. This makes v1's result depend on source rows outside the requested snapshot boundary. The v2 equivalent explicitly checks `later.created <= cutoff` at `schema_v2_session_extraction.go:255-260`.

**Evidence:** Direct query and consumer trace at the cited lines; comparison with the v2 cutoff guard; contract clauses requiring cutoff-consistent duration and exclusion of future activity. The current v1 active snapshot test exercises repeated cutoffs and completion but does not add a post-cutoff newer message (`schema_v1_session_extraction_test.go:1713-1752`). This is an introduced defect in the candidate. The narrow correction is to apply the same cutoff bound to the v1 newer-message predicate and exercise that boundary.

## Architecture assessment

- **Pure and cohesive cutoff semantics:** Domain interval arithmetic is deterministic and clock-free in `domain.DurationInterval`; component rounding and validation are centralized in `MeasuredDuration`/`ValidateDuration`. CLI, backfill, capture and parser boundaries supply explicit cutoffs. Source-specific proof stays in Codex and OpenCode adapters, so the core does not infer liveness from raw source schemas. The v1 finding above is a boundary inconsistency, not a domain-layer purity or ownership violation.
- **Source/projection ownership:** Codex lifecycle markers are interpreted in `codex`; OpenCode v1/v2 native message/session markers are interpreted in their schema-family extractors; persistence DTOs validate evidence at the ledger boundary; domain rollups merge compact evidence; TypeScript validates and formats projections. I found no evidence that raw source records or process state leaked into the domain or durable evidence.
- **Schema-v6 expansion:** The invocation evidence is optional and compact (state, completed/provisional components, cutoff, basis, optional limitation). Current fact-ledger DTO and migration decoder now recognize it, preserve absence for legacy facts, validate components, and copy it through conversion. Frontend evidence is optional at read boundaries. The change does not widen the durable schema to transcripts or lifecycle history. The stated reader-first rollout constraint remains important: a pre-change strict v6 reader rejects the new member, so readers must be deployed before a writer rollout; this candidate does not authorize a release.
- **Aggregation:** Workflow, fleet, comparison and time-bucket projections retain the numeric total and carry provisional seconds, invocation counts, unknown coverage and distinct cutoffs. Legacy evidence absence becomes unknown coverage, not a completed zero. Cutoffs are deduplicated and sorted. This is bounded to compact evidence fields, though the distinct cutoff list naturally grows with distinct persisted snapshots.

## Consequential hypotheses assessed

- **Does the domain infer ongoing work from a missing end alone?** No: `DurationInterval` returns a limitation unless source adapters pass positive ongoing evidence. OpenCode v1 uses running tool/open part state; v2 uses its idle state or a running tool; Codex uses explicit open lifecycle starts. The user’s pending policy clarification about provisional persisted source-state intervals versus live-process proof remains unresolved. This review assessed the implemented source-state provisional semantics without deciding that policy.
- **Can schema-v6 evidence be lost on ordinary read, projection, or explicit migration?** The candidate adds it to strict invocation DTO decode/encode and the migration input/domain conversion and equality validator. Existing absence remains accepted and interpreted as unknown. No loss path was identified in the inspected paths.
- **Does a derived projection invent a globally common observation time?** No: `MergeDuration` unions distinct cutoff strings rather than selecting one. API, workflow, comparison, fleet and progression carry the summary.
- **Can a future source row affect cutoff-bounded v1 activity?** Yes; this is the reported finding. The v2 source-message scan bounds its supersession check by the cutoff.

## Verification and limits

- Verified the frozen manifest hashes and baseline revision before review; no manifest drift was observed.
- Static source tracing only for the reported v1 edge case; no disposable SQLite probe was run.
- Reused reported verification: full Go race suite, frontend tests (834), typecheck and lint passed at the frozen candidate revision; native OpenCode v1/v2 live generation and refresh evidence is documented in `RUNTIME_EVIDENCE.md`. These do not cover the post-cutoff v1 message case and are not a substitute for the targeted boundary check.
- No full suites repeated. No product or test files edited. Only this assigned review artifact was written. No commits, telemetry, PR actions or children.

## Verdict

**CHANGES_REQUESTED.** Correct and verify the v1 post-cutoff supersession guard. Apart from that concrete cutoff inconsistency, this scoped architecture review found the duration core pure/cohesive, ownership correctly placed, and the v6 evidence addition compact and backward-compatible for upgraded readers. This verdict does not resolve the pending source-state/liveness policy, certify bug absence, or authorize release.
