# AL-64 duration investigation

Revision: working diff based on a84adbe5be46c2ef52ede0987a417562715d38e1. Implementation author: coder, gpt-6.1-sol/high (inherited runtime). Scope approval is recorded in PLAN.md. This is working evidence; consolidated review and integration checks are still pending.

## Per-family results

### Codex

Confirmed defect: readEvent previously accrued only paired start/completion intervals. Explicit task_started/turn_started remained in turnStarts and contributed no duration until completion. This corroborates, without reclassifying as a newly reproduced observation, the supplied AL-55 root 287 to 1417 seconds / total 2058 to 3188 seconds delta and unchanged four children recorded in CONTRACT.md.

The parent inspected native CLI version 0.166.0 source vscode, current root 01a1a7d8-b9d2-7383-a6f7-3c7002ba88e6. Its primary rollout has task_started 2026-11-04T18:36:02.663Z, task_complete 18:40:12.141Z, and an open task_started at 18:51:31.551Z. A child replay demonstrates inherited metadata/history and turn_aborted; parser ownership/replay isolation remains intact. These observations establish native lifecycle shapes, not completed end-to-end generation/runtime refresh verification.

ParseRolloutAt now receives a deterministic cutoff, retains completed intervals and adds open explicit lifecycle intervals provisionally. Completion after cutoff is ignored for duration at that snapshot. Explicit interruption/abort bounds measured activity; shutdown/thread end also stop pending activity. Superseded starts and orphan completions report partial coverage; duplicate IDs do not count twice. Invalid boundaries do not create negative duration. Per-model totals round once at the existing invocation boundary. Non-duration source facts are preserved; cutoff does not discard later user/token records.

Deterministic replay cases cover completed plus open (14 seconds at cutoff 20, including 10 provisional), advancing cutoff (24 total at 30), future completion, final completion (34 total), interruption (9 completed) and unchanged repeats. No live AL-55 ledger or workspace is a fixture.

### OpenCode schema v1

Confirmed accounting gap: assistant payload time.created/time.completed pairs were the only duration producer; unfinished assistant activity accrued zero. The generic message/part updated timestamp was not a reliable activity completion and is not substituted.

Native 1.19.5 isolated observations from the runtime tester establish streaming assistant creation plus native text time.start without time.end, step-start, and explicit running-tool state/time.start. Completion supplies assistant time.completed and text/tool end. Both open streaming-part timing and a running native tool on the latest assistant are now supported unfinished lifecycle evidence. A later message supersedes old open-state evidence. Missing assistant completion without those additional lifecycle markers is partial/unsupported, with known measured duration preserved.

Important limit: the runtime tester observed SIGINT leaving the v1 assistant incomplete and running tool state unchanged. Stored open lifecycle state cannot prove that its process remains alive. The parent has requested a user decision on whether such explicitly open source intervals may remain provisional with that limitation. Current candidate implements source-state provisional semantics pending that decision; it must not be represented as process-liveness verification. See RUNTIME_EVIDENCE.md for the native probes and source versions/identities.

Existing completed root 14 seconds and child 15 seconds remain covered. New deterministic extraction cases cover active 10 seconds, advancing to 20 seconds, fixed-cutoff repeat and final completion at 30 seconds, without additive snapshot accumulation.

### OpenCode schema v2

Confirmed defect: assistant message time_updated was substituted for absent payload time.completed and presented as completed duration. Supported pending/running tools also caused extraction to fail closed, preventing a truthful active snapshot. The replacement uses explicit assistant completion, native session time_idle lifecycle state where the schema provides it, or explicit running tools, and labels supported open intervals provisional. Missing completion alone remains unsupported. Later idle/user/assistant rows supersede open activity proof. Native time_suspended is not treated as terminal: runtime observations show it can predate an actively streaming assistant.

Native 2.2.0 runtime observations establish assistant time.created without completed, no idle row, and session time_idle NULL while streaming. Completion writes assistant time.completed, an idle row and session time_idle. Cancellation writes assistant completion, finish error and idle outcome interrupted. Source state is still not a general guarantee against an unrecorded process crash.

The runtime tester also found native v2 tool timing at content item.time.created/ran/completed, whereas the existing adapter read state.time. The reader now prefers the native top-level item time and retains the established imported state.time fallback. In-progress tool trace states remain pending/running rather than completed; unknown states still fail explicitly.

Deterministic family-specific extraction cases cover 10 seconds provisional, advance/fixed cutoff at 20 seconds, completion at 30 and repeat. Native runtime-generated/persisted numerical checks belong in RUNTIME_EVIDENCE.md; do not confuse fixture results with runtime execution.

## Durable and display semantics

Invocation active_duration_seconds remains the numeric sum. Optional duration_evidence carries typed complete/provisional/partial coverage, completed/provisional components, source-family basis, cutoff for provisional observations and explicit unsupported limitation. Integer components reconcile to the existing per-invocation rounded total. Legacy absence means unknown coverage, never verified completion or measured zero.

DTO validation checks nonnegative coherent components and cutoff/state combinations. Migration decode/copy and invariant validation preserve evidence. No source event journal or transcript is persisted. Workflow/fleet/comparison/progression summaries aggregate provisional seconds, counts of provisional/unknown invocations and distinct recorded cutoffs. Reads retain their stored cutoff rather than advancing by the current clock. Exact source-owner replacement and unavailable-history retention continue through the existing portable refresh mechanism.

CLI/Markdown and dashboard ticket/fleet/comparison/scatter duration consumers identify provisional or unknown coverage, including accessible descriptions and recorded cutoffs. Primary numeric KPI values remain available. Frontend progression calculations propagate the same evidence. Registered capture, explicit generation and backfill receive a common duration cutoff per operation; capture registry reconciliation-completion timestamps retain their independent operational meaning.

## Reader-first schema v6 expansion

Compatibility is bounded and explicit. Existing valid v6 inputs without duration_evidence remain accepted and retain their numeric facts. Expanded readers validate and preserve new optional evidence. Pre-change strict v6 readers reject the new optional member; this is not transparent old-binary forward compatibility. Upgrade all consuming readers before enabling new writer emission. Do not strip evidence to appease old consumers. This PR does not authorize release/distribution rollout. Parent explicitly accepted this reader-first optional expansion within the approved contract. No destructive migration or authenticity gate is introduced.

## Verification so far and remaining work

Passed: focused Go domain/migration, Codex, OpenCode v1/v2, adapter serialization, HTTP projection and backfill suites; canonical frontend build; TypeScript typecheck and lint. A broad frontend run passed 822/824 with two KPI text expectations; keeping numeric primaries and coverage in secondary text resolved the two affected suites (31 tests). The independent tester will run affected final checks and add missing persistence, cross-machine and refresh assertions.

The initial CLI suite found unrelated source freshness/capture-feedback/user-iteration regressions caused by discarding entire source records beyond the duration cutoff. Narrowing cutoff handling to duration fixed those failures: TestCaptureLaterAgentChannelReportsBoundPublicationFailure, TestCaptureNativeCheckExemptsOnlyAcceptedExactCallerTail and TestCLIGenerate_ThreeDeviceHandoffAccumulation now pass. Preserve their assertions.

Known remaining test update: main_refresh_test.go TestCLIGenerate_SessionRefreshMatchesFreshExtraction assumes active duration zero, nonnil completion for an unfinished invocation, and byte/struct equality across independently timed active generations. The new acceptance requires nonzero source-supported provisional duration, no fabricated invocation completion, and comparisons using the same explicit cutoff (or verified advancing-cutoff deltas) while retaining exact identity/token/turn/completed-session assertions. OpenCode no-running-evidence active fixture should assert partial coverage and nil completion. Do not erase meaningful idempotence checks for fixed cutoffs/completed facts.

Required integrated make, independent review, source-backed runtime generation/backfill evidence, final embedded binary verification and telemetry completeness are not claimed complete here. Raw transcripts and disposable source copies remain outside Git.

## Native v1 ingestion prerequisite follow-up

The runtime tester's live candidate generation exposed an existing schema-detection defect before duration extraction: a fresh native OpenCode 1.19.5 database initializes complete session/message/part plus an empty, correctly shaped session_message auxiliary table, without session_v2. The previous detector classified this native layout as an incomplete mixed family. The narrowly scoped correction recognizes complete v1 with that empty, valid auxiliary table only; populated auxiliary rows, malformed auxiliary columns, any present incomplete v2 owner table and genuinely competing families remain explicit errors. This does not select a family by version string or discard populated unknown history. Runtime disposable source: /tmp/al64-runtime-o27v2fde/v1/opencode.db; no source capture enters Git. Dedicated regression coverage is owned by the independent tester.

## Candidate 01 review corrections

The independent review identified four production defects, corrected without changing the pending source-state/liveness policy. DUR-01 now marks a matching reversed or unparsable Codex completion as partial coverage while retaining known completed seconds and excluding the unmeasurable interval; the invalid terminal no longer silently certifies complete coverage. DUR-02 uses native v1 tool activity start (falling back to source record creation for older running-tool records without native start) and native streaming start, rather than generic last-update time, to bound open activity at cutoff. DUR-03 bounds v1 superseding messages to creation at or before cutoff, so future messages cannot erase prior supported activity. DUR-04 passes the immutable operation cutoff through explicitly enabled fallback OpenCode tree discovery/extraction. Independent regression ownership remains with the tester, including DUR-05 no-tool streaming cases. Discovery attribution and final verification belong in the parent consolidated review.

## Candidate 02 invalid-start correction

DUR-06, discovered by the architecture reviewer (gpt-6.1-sol/high) during correction verification, is distinct from the invalid-completion defect: a missing or unparsable start timestamp marked the turn ID seen without recording an interval or limitation; later same-ID completion could then certify the known subtotal as complete. The parser now records an invalid lifecycle start limitation on that branch, retaining supported seconds as partial coverage and creating no invented ongoing interval. Valid duplicate starts and future-lifecycle cutoff behavior remain unchanged. Independent tester owns malformed/missing-start regression cases; parent retains verification and final disposition.
