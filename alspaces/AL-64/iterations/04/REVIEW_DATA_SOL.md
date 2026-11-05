# AL-64 iteration 04 data and failure-handling review

Verdict: CHANGES_REQUESTED for the assigned data/failure scope. One independently reproduced introduced feedback defect; no numeric accounting or persistence correction is requested.

## Revision and identity

Run: checkpoint_data_sol, independent first-pass discovery, CP-01. Baseline: `d8a7c14cfd9f0ee626fedc33c70f92bd928d9192`; candidate is the frozen staged/unstaged diff supplied in `/tmp/al64-04-review/candidate.diff`. Manifest `/tmp/al64-04-review/fingerprint.json` SHA-256: `c766dd83ec3e9f601e6e5b302422defa2bbe6c86f5906648cae1e272a338a32d`. All 411 paths matched before investigation and again immediately before this verdict. Requested/parent-supplied reviewer attribution: data integrity and failure handling, gpt-6.1-sol / high. This reviewer did not independently inspect native model metadata.

Product and tests remained read-only. Only this assigned report was written in the checkout; isolated overlay artifacts are under `/tmp/al64-04-data-sol/`. Initial checkout already contained the parent/coder/tester AL-64 candidate changes. No sibling personas or review conclusions were read, and no children, staging, commits, telemetry or publication were performed.

## Finding DATA-SOL-01: native v2 idle state can still produce source-open feedback

Severity: 4/10. Introduced consequence at `backend/internal/adapters/opencode/schema_v2_session_extraction.go:283`; relevant existing inputs at lines 100-101 and 262-266.

Trigger: an accepted schema-v2 owner has a final assistant message with no message completion, a tool still marked running, and native `session_v2.time_idle` after that assistant's start. There need not be a separate later idle message. The idle timestamp may be before or after the fixed duration cutoff.

The adapter only retains IDs whose `time_idle IS NULL` in `activeSessions`. It loses the actual non-null idle timestamp. The existing tool loop then changes `running` back to true, and the new `SourceOpen` expression treats that numeric ongoing input as sufficient current source-open proof. The exact interval reaches the production selector through `ExtractSessionsStrictWithPolicyAt`/`ExtractSessionsStrictWithSourceAt`, `SourceEvents.CheckpointEvidence` and explicit `runGenerate`; registered extraction reaches the same implementation through `CaptureSessionSnapshotWithPolicyAt`.

Observed result: at cutoff start+180000ms, a final assistant with a running tool and `time_idle=start+120000` or `start+200000` emits F03 claiming the source reports an open assistant message at 180s. Both fixtures retain active180/provisional180/input17. A control with the same fixture and `time_idle=NULL` correctly emits F03.

This violates the approved distinction between historical numeric timing and current source-open evidence: native idle/terminal evidence after the interval start cannot certify that interval as currently source-open, including when that idle was observed after the numeric cutoff. This is a source lifecycle contradiction, not a proposed process-liveness or caller-authenticity requirement. Existing numeric behavior is unchanged and is not reported as a newly introduced accounting defect.

Required correction: retain enough native idle state to disqualify current source-open wording for this interval, without altering the accepted numeric duration behavior. The eligible180 subtotal should receive qualified F04; a fully known short interval should retain no-feedback suppression. Add a focused native-idle control covering before/after-cutoff observations and an active/null-idle control.

Reproduction: disposable overlay `/tmp/al64-04-data-sol/overlay.json` maps only the existing checkpoint test file to an external copy containing `TestReviewDataNativeIdleDisqualifiesSourceOpen`. From backend:

`go test -overlay /tmp/al64-04-data-sol/overlay.json ./internal/adapters/opencode -run '^TestReviewDataNativeIdleDisqualifiesSourceOpen$' -count=1 -v`

Executed twice during investigation; final log `/tmp/al64-04-data-sol/native-idle.log` records exit1, null-idle control PASS, both non-null-idle cases FAIL with actual F03/expected F04. These are synthetic real-adapter SQLite fixtures, not authentic installed-runtime evidence. No shared product/test mutation was used.

## Other consequential hypotheses assessed

- Owner/model/row attribution through grouping: refuted for the inspected supported paths. Codex emits evidence before sorted model grouping; OpenCode emits it alongside each invocation row. `SessionsFromFacts`, `ProjectGenerationFacts` and optional-category transforms preserve invocation order. Selector matches source/session/one-based row/model/cutoff, treats duplicate row evidence as ambiguous, and preserves legitimate row multiplicity. Cross-harness ID collisions are rejected by existing merge validation before publication.
- Stale capture or retained history promoted to fresh proof: refuted in inspected callers. Fresh exact reads carry request-local evidence; capture snapshots persist only SessionFact data. Failed/unread owners preserve recorded facts without fresh interval evidence, so eligible retained subtotals receive F04. Main's explicit source facts and evidence both supersede the same captured owner. `MergeSourceEvents` removes replaced owner evidence, including complete empty replacement.
- Fabricated Codex interval identity/currentness: the supplied pre-review correction is present. Synthetic accounting IDs become empty native IDs and qualify fallback. After-cutoff lifecycle observations disqualify source-open evidence; native start/model attribution is kept before grouping. Existing warning failures remain existing failures. This verification does not claim discovery credit for the parent/coder/tester corrections.
- Optional feedback corrupting successful publication: refuted by code trace and reused subprocess checks. Selection/emission occurs after requested ledger/summary commit and diagnostic acknowledgement; write failure is reported without exit/fact rollback. Automatic capture generation does not emit. SourceEvents/Transcript exclude checkpoint evidence from JSON; DTO/schema/summary/frontend surfaces are untouched.
- Invalid or disabled thresholds and optional metrics changes: strict YAML-node decoding/defaults reject prohibited types/values, preserve int64 overrides and existing metrics, and zero disables selection. Shared/local precedence is unchanged. Reused loader/writer and CLI cases exercise these boundaries.
- Caller/root confusion, pooling and misleading exact ages: no additional finding. Production supplies no verified caller, so F01/F02/F05 remain gated pure cases. Selector uses individual native age, not pooled rounded subtotals; complete short evidence avoids fallback. Partial timing and unsupported identity remain qualified. No new authenticity/liveness policy is inferred.

## Verification reused and limits

Read approved PLAN, implementation/runtime capability reports, final tester report, own persona/common review procedure, technical standards and applicable portable-ledger/regression boundaries. Inspected changed domain/config/source/capture/CLI paths plus unchanged grouping, merge, publication and retention callers.

Reused final frozen evidence, confirming log contents: `/tmp/al64-04-tester-frozen-affected-race.log` passes domain, migration, config, Codex, OpenCode and capture (capture57.417s); `/tmp/al64-04-tester-cli-final-pass.log` passes focused CLI race14.083s. Tester reports final `make test-unit` including190 frontend units and existing optional metric/portable retained-owner matrices; these are reused evidence, not rerun by this reviewer. The new isolated probe exposes a missing native-idle branch despite those passing checks.

Scoped conclusion does not certify all repository behavior. Installed native F01/F02 caller/current-interval and completion/next-turn transports remain unverified as documented; approved F03/F04 fallback is sufficient within its truthful evidence boundaries. Parent owns reconciliation, fixes, integrated make, standalone verification and publication. Existing unrelated retention/null-numeric debt was not expanded or reclassified here.
