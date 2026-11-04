# AL-64 candidate 02 data and failure-handling correction review

Outcome: **APPROVED for the scoped corrections**, conditional on the existing source-state provisional interpretation. This is not delivery approval and does not decide the pending process-liveness requirement.

Persona: reviewer-data. Requested and parent-acknowledged inherited execution profile: gpt-6.1-sol / high. No independent provider execution/effort or cost metadata is exposed; this report makes no measured usage claim. Run: AL64-candidate-02 / recheck_data_sol; correction verification, not first-pass discovery.

Baseline: a84adbe5be46c2ef52ede0987a417562715d38e1. Frozen manifest: /tmp/al64-review-02/fingerprint.json, SHA-256 027324fa31f45b9933adb35bd34835b54918757047556354e8b2c3ce82f4c24c. Verified the manifest digest and all 46 application/test file hashes before investigation and again immediately before this verdict: 46 matched, zero differences. The initial working tree already contained staged and unstaged ticket/application changes; they were preserved. Only this assigned report was written in the repository.

## Findings and supplied correction dispositions

No new substantiated correctness finding in this bounded investigation. Numeric severities below are the supplied original findings, now verified corrected; discovery credit remains with their original reviewers in REVIEW.md.

| Finding | Original severity | Disposition and evidence |
| --- | ---: | --- |
| DUR-01 | 5 | Fixed. rollout.go:440-455 retains the measured subtotal, marks an invalid matching completion as a duration limitation, and removes the unusable open start. The reversed and unparsable terminal cases each retain 4 completed seconds, exclude the invalid interval, report partial coverage, and contribute one unknown invocation. No cutoff or invalid terminal is substituted as a completion. |
| DUR-02 | 6 | Fixed. schema_v1_session_extraction.go:198-211 uses native tool start, with record creation only when that start is absent, rather than rejecting activity because its generic update is after cutoff. The retained later-progress-update oracle passes. Independent SQLite overlay cases additionally verify absent-start fallback, future native start, zero native start, malformed native start, and absent native plus invalid record start. Only the supported fallback counts provisionally; invalid/future proof preserves partial/unsupported coverage. |
| DUR-03 | 6 | Fixed. The same v1 query bounds superseding message creation to at or before cutoff. The retained future-message oracle passes. Independent overlay controls with newer user messages before and exactly at cutoff confirm legitimate supersession still excludes stale open proof and reports partial coverage, without fabricated completion. |
| DUR-04 | 4 | Fixed. recovery.go:203-216 and :298 pass the immutable operation cutoff into fallback tree extraction. The existing enabled-fallback test persists exactly 3 provisional seconds and the supplied cutoff, retains the missing exact owner, adds the discovered owner once, and keeps recovery incomplete. |
| DUR-05 | 5 | Regression protection added. Existing family tests now exercise v1 native timed text and reasoning without tools and v2 idle-null streaming without tools, including advancing/fixed cutoff, completion replacement and completed repeat. Inspected tester overlay logs show both path-removal mutants fail these retained tests; current controls pass. This is verification of the supplied coverage finding, not new discovery. |

## Consequential persistence and failure paths

- Traced Codex limitation projection through finalizeFacts, MeasuredDuration and DurationForInvocation. The corrected error is represented as partial coverage while known completed milliseconds remain authoritative; the deleted invalid interval cannot become later provisional time. Valid completion, future completion and interruption controls still pass.
- Traced exact-owner recovery and replaceRecoveredOwners. Extraction completion remains distinct from duration coverage. Fresh facts replace the exact successful owner rather than being added to old cumulative duration; unavailable owners and unrelated harnesses stay outside replacement. A partial duration estimate is explicitly represented, not silently certified as completed coverage.
- Re-executed the three-family persisted backfill lifecycle matrix. It checks active 10 seconds, advance to 20, fixed-cutoff repeat, completion to 30 and completed repeat; unchanged 3-second child and linkage; immutable retained 7-second unavailable-owner snapshot/cutoff; 17-second other-harness history; exactly four owners; full-ledger rollups; and no source recomputation under Force alone. All passed.
- Fallback retains the existing overlap guard: heuristic enrichment cannot replace an unavailable exact stored owner or certify complete historical recovery. The cutoff correction does not broaden fallback authorization.
- Invocation DTO and migration duration decode/copy paths retain the optional typed evidence and validate it with ValidateDuration. No correction changes those serializers, reader authority, migration policy or ordinary-read clock behavior. The accepted reader-first rollout limitation remains: older strict v6 binaries reject the newly emitted member.

## Verification

Executed from backend, exit 0:

```text
go test ./internal/adapters/codex ./internal/adapters/opencode ./internal/adapters/backfill -run 'TestRolloutActiveDurationCutoffAndTerminalBoundaries|TestSchemaV[12]ActiveDurationSnapshotAndCompletion|TestBackfillEngine_ExplicitFallbackResolvesWhenAllowed|TestBackfillActiveDurationReplacesSnapshotsAcrossSourceFamilies' -count=1
go test -overlay /tmp/al64-recheck-data-sol/overlay.json ./internal/adapters/opencode -run TestAL64RecheckV1ActivityBounds -count=1 -v
```

The second command adds seven independent SQLite boundary cases through a temporary test overlay; shared product/tests were unchanged. Inspected the tester's original-code failures in /tmp/al64-tester-prefix/results.log and no-tool path-removal failures in streaming-mutant.log. They discriminate the supplied root causes and the missing streaming guards. Reused the recorded candidate-02 affected-package race successes and candidate-01 full Go race, frontend 834 tests, typecheck and lint for unchanged scopes; no full suite was repeated without a new concern.

## Remaining limits

The policy decision about persisted open lifecycle state versus additional live-process proof remains pending. This review verifies current provisional source-state accounting only; it does not establish that a process is still alive or authorize that product decision. Native OpenCode generation/active refresh/completion/repeat evidence predates these narrow corrections and is reused with the deterministic corrected-boundary evidence above. Native Codex lifecycle records are observed, but the fresh native CLI generation/refresh probe was blocked by correct project ownership validation; those runtime operations remain unverified. Integrated make, final standalone embedded asset verification, consolidated approval, telemetry and publication remain parent-owned and pending. No release or merge authorization is given.
