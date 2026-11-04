# AL-64 candidate 03 data and failure-handling correction review

Outcome: **APPROVED for the scoped data corrections**, conditional on the still-pending product policy for persisted open source intervals. This is not delivery approval and does not resolve process-liveness semantics.

Persona: reviewer-data. Requested and parent-acknowledged identity: gpt-6-luna / medium. Runtime metadata independently confirming model/effort was unavailable; no usage or cost claim is made. Phase: candidate-03 bounded correction re-review, not blind discovery.

## Revision and scope

Baseline: `a84adbe5be46c2ef52ede0987a417562715d38e1`; branch `fix/AL-64-active-turn-duration`. Frozen application/test manifest: `/tmp/al64-review-03/fingerprint.json`, SHA-256 `f9b8bb31e5d7a3776adf9fef2b2f3f456711e76bfbceb7f276880e7e4634eb49`. All 46 path hashes matched before review and again immediately before this verdict. Inspected corrected Codex lifecycle parsing, OpenCode v1/v2 activity evidence, operation cutoff propagation through generation/backfill/capture and exact-owner/fallback recovery, related regression cases, contract, and supplied reports. Product and test paths were read-only; only this report was written.

## Supplied correction dispositions

No new distinct data-integrity finding was demonstrated in this bounded pass. Discovery credit for supplied issues remains with their original discovering reviewers.

| Finding | Severity | Disposition |
| --- | ---: | --- |
| DUR01 invalid Codex completion | 5/10 | Fixed. A matching completion that is malformed or precedes its start sets an explicit duration limitation before the unusable start is removed. `finalizeFacts` keeps the known completed subtotal, does not extend the invalid interval to cutoff, and projects partial evidence. Supplied red/green overlay cases discriminate the prior false-complete behavior. |
| DUR02 v1 activity start vs generic update | 6/10 | Fixed. The v1 query bounds open running-tool or timed text/reasoning evidence by native activity start, falling back to record creation only when native start is absent. A generic progress update after cutoff no longer suppresses valid earlier activity. Supplied boundary controls cover missing, future, zero and malformed starts. |
| DUR03 legitimate supersession | 6/10 | Fixed. A later message suppresses the earlier v1 open evidence only when that message was created at or before the operation cutoff. Future messages cannot invalidate the snapshot; before-cutoff and exact-cutoff controls preserve legitimate supersession. |
| DUR04 fallback cutoff propagation | 4/10 | Fixed. `recoverOpenCodeAt` passes the fixed operation cutoff through `readOpenCodeTree` into strict extraction. The enabled fallback retains the missing exact owner and adds the distinct discovered owner once, using the requested cutoff and exact provisional duration. |
| DUR05 no-tool streaming paths | 5/10 | Regression guard is effective. The v1 native timed text/reasoning cases and v2 idle-null streaming case exercise those production paths; supplied path-removal mutants fail while controls pass. This re-check confirms the oracle evidence and takes no discovery credit. |
| DUR06 malformed/missing Codex start | 6/10 | Fixed. Invalid or absent start explicitly marks partial/unsupported duration while retaining the known completed subtotal. It does not synthesize an interval or completion from cutoff. Supplied pre-fix branch overlay fails the new table cases; corrected Codex race suite passes. Discovery credit remains with the candidate-02 architecture reviewer. |

## Data and replacement interactions assessed

- Codex valid completed intervals remain in `activeMillis`; invalid completion/start boundaries add a limitation without adding elapsed time. Valid open starts alone can contribute provisional duration through the fixed cutoff. Future lifecycle records are cutoff-gated. This preserves known subtotal semantics while preventing invalid intervals from certifying full coverage.
- OpenCode v1 uses the source-native start and only uses record creation as the documented fallback. The SQL supersession predicate is scoped to the same session and messages at/before cutoff. v2 uses native idle-null or running-tool evidence, while a later at-cutoff record ends the open proof. These paths do not turn generic `time_updated` into a completion measurement.
- The fallback change carries one cutoff through tree extraction. The exact-owner overlap guard remains in place; heuristic enrichment does not replace a missing exact owner. Existing exact-owner replacement remains a replacement of the successfully extracted snapshot, not a union or cumulative addition. Unavailable owners and other-harness history remain preserved outside the refresh set.
- Invalid/missing evidence is represented as partial/unsupported rather than measured zero or invented completion. The supplied lifecycle and recovery tests cover those important transitions. The reader-first v6 compatibility boundary remains as documented; no writer rollout or release is authorized here.

## Verification

Executed from `backend/` (exit 0):

```text
go test -race ./internal/adapters/codex ./internal/adapters/opencode ./internal/adapters/backfill -run 'TestRolloutActiveDurationCutoffAndTerminalBoundaries|TestSchemaV[12]ActiveDurationSnapshotAndCompletion|TestBackfillEngine_ExplicitFallbackResolvesWhenAllowed|TestBackfillActiveDurationReplacesSnapshotsAcrossSourceFamilies|TestAL64' -count=1
```

Reused the supplied candidate-03 original-code red/current-code green overlays and streaming path-removal mutant results, affected Codex/OpenCode/backfill race suite results, and candidate-01 full Go race, frontend 834 tests, typecheck and lint for unchanged scopes. I did not repeat the full suites. Integrated `make`, final embedded-asset/binary verification, and live native Codex generation/refresh remain parent-owned and pending. The runtime evidence reports native Codex fresh generation blocked at the correct project-ownership check; this is an explicit unverified runtime path, not a failed duration assertion.

## Policy and remaining limits

The user's clarification on provisional persisted source intervals versus additional process-liveness proof remains pending. This review assesses the current source-state provisional implementation and does not decide that product policy, establish process liveness, or approve delivery across that acceptance gap. Native v1 cancellation evidence shows open persisted markers can outlive the process; retain the disclosed uncertainty. Native OpenCode live generation/refresh/completion/repeat evidence is reused, but does not establish the pending policy. Parent owns integrated verification, policy resolution, consolidated verdict and delivery.
