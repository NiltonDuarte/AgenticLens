# AL-69 iteration 01: Accurate telemetry and preserved evidence

Status: initial plan; required-configuration acceptance is superseded by [iteration 02](../02/PLAN.md). Initial Git status was clean at `eb7d169c7f18fb935d4c2ceac998d4eca2ac36f4`, branch `AL-69/preserve-accurate-telemetry-and-evidence`. The user's "start working on it" instruction authorizes planning; it is not recorded as approval of the initial rate formula, which the user has since superseded.

1. Coder traces and corrects execution classification in `domain/live.go` and its OpenCode/live consumers using structured evidence only. Remove all textual process and severity inference, including execution-wrapper and legacy fallbacks. Decode supported structured execution fields at source boundaries; preserve unknown/malformed states without invented success or failure. Share schema-v1 decoded records from the owning OpenCode adapter with live parsing; preserve consumer projections, source strings, privacy and cache ownership.
2. Coder threads effective thresholds through `domain/anomalies.go`, `events.go`, `aggregations.go`, generation, capture reconciliation and live historical hydration. Add a typed live failure-rate rule backed by the shared evaluator; update TypeScript contracts/validation and consumers. Preserve recorded summaries when source coverage cannot authorize refresh.
3. Coder surfaces existing `ScanReport` diagnostics in CLI selection/listing, preserving partial siblings and fatal boundaries. Add typed live relationship completeness in `LiveTraceView`, service snapshots and response consumers; retain content when child queries fail.
4. Coder replaces `adapters.UpdateSummaryFile` boundaries with explicit generated markers and conservative heading/fence-aware legacy handling. Correct historical/live anomaly identity matching and rule-specific streak extraction in `web/src/domain/anomalies.ts`, `TraceWaterfall` and export/inspector callers.
5. Tester extends existing coverage rather than duplicating suites: `tool_execution_test.go`, live/OpenCode parity, anomaly/event tests, repository/CLI writer and scan tests, HTTP responses, and frontend anomaly/waterfall/inspector/export tests. Repair `TestLiveTraceSurvivesChildSessionLookupFailure` to require diagnostics while preserving content assertions.
6. Verify non-default thresholds above/below defaults at exact boundaries through generate/save/reload/detail/fleet/live, including event pruning. Exercise two-machine unavailable-history preservation, mixed harnesses, recorded-empty versus omitted collections, complete/partial refresh, raw-payload exclusion and no GET writes. Test successful error-text reads, real failures, no-match, malformed/structured output, rejected siblings, failed/empty/linked child lookup, Markdown peer/parent/fenced headings, legacy append and repeat idempotence, conflicting tools and exact part identifiers, separated errors versus streaks.
7. Parent freezes the candidate and assigns the review roster below under the configured model policy. Reviewers preserve the shared checkout and own reports only; parent owns consolidation and retained attribution.
8. Parent integrates main and requires canonical `make`, current embedded frontend assets, embedded-asset checks and standalone-binary verification. Record actual revision/results and environmental limits; only then finalize review, artifacts, telemetry and PR delivery.

## Grounded implementation entry points

- `ClassifyToolExecution` currently reads exit-code/crash text for every tool. Existing `tool_execution_test.go` and `live_test.go` assertions treating completed shell output as proof of failure are superseded by the user's structured-only decision. Replace those expectations with text-invariance checks, including recognized-looking wrappers, and independent structured-failure assertions.
- `detectToolHighFailureRates` ignores its thresholds argument and hardcodes `>0.20` / `>=0.50`, with minimum five calls.
- `capture/reconcile.go` calls `PersistSourceAnomalySummariesWithPolicy` without project thresholds. `live/service.go` calls `DetectAnomaliesWithEvidence` and its refresh fallback without thresholds.
- `ScanWorkspaces` already accumulates `WorkspaceDiagnostic` values and `FatalError`; CLI selection reports a rejected target as not found.
- `LiveService.readSessionTrace` ignores `ResolveChildSessionIDs` errors. The existing regression removes `session.parent_id`, retaining useful trace reads.
- `UpdateSummaryFile` searches only for the next `## ` heading, consuming peer `### Verification` sections.
- `schemaV1PartState.Output` is a string pointer; live `rawMessageToString` trims strings. Sharing must reconcile structured output without widening persisted payloads.
- `matchesAnomaly` permits turn fallback after tool mismatch; `findMatchingLiveAlert` uses substring part/tool matches. `extractStreakCount` returns generic counts, and `TraceWaterfall` also reads `context.count` directly.

## Review ownership

Every focus runs independently under both `gpt-6.1-sol`/high and `gpt-6-luna`/medium, in capacity-limited waves. Report names expand to `REVIEW_<FOCUS>_<PROFILE>.md` in this iteration; assign concrete run IDs and a frozen candidate before dispatch.

| Focus / evidence owner | Risk question and call paths | Report prefix |
| --- | --- | --- |
| reviewer-behavior | Do classification, thresholds and identity agree from CLI/live source through API, dashboard and exports? | REVIEW_BEHAVIOR |
| reviewer-data | Do partial scans, child lookup failures, source refresh and Markdown writes preserve supported evidence and portable history? | REVIEW_DATA |
| reviewer-architecture | Are v1 decoding and pure policy shared without leaking I/O, payloads or cache ownership into the domain? | REVIEW_ARCHITECTURE |
| reviewer-tests | Do actual consumer tests detect boundary failures, distinguish missing from empty, and cover exact rates and identities? | REVIEW_TESTS |

Full implementation approval remains pending, including the new typed live rule. The [active configuration delta](../02/PLAN.md) owns required values and unresolved numeric choices. No runtime tests or review gate have run during planning.

## Scope clarification: structured execution evidence only

Source: the user's explicit instruction, "Drop any textual process, only trust the structured data. I dont want to guess anything from free form data". This accepts the structured-only classification boundary and supersedes the initial, unapproved proposal to retain a legacy execution-envelope fallback. The draft contract is corrected within this planning iteration; no approved implementation or completed verification is displaced.

Changed acceptance: output text never determines execution failure, severity, exit status, timeout, crash or no-match behavior. Preserve explicit source tool status separately from an unavailable process outcome; missing or malformed structured evidence remains unknown. This applies to shell tools as well as reads/searches, including legacy textual wrappers. Preserve all source text for inspection, live/offline parity, privacy, portable history and the other five ticket areas.

Verification impact: vary free-form stdout/stderr/error text and wrapper-looking text while keeping structured facts fixed, and require unchanged classification and severity through actual live/offline consumers. Separately vary structured status and supported exit/signal fields and assert their documented semantics. Exercise contradictory, absent, malformed, pending and unknown fields and structured grep/glob no-match cases. Remove tests that require free-form text to supply execution facts; retain their raw-evidence preservation assertions. Rate-policy approval remains pending.
