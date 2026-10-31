# AgenticLens Architecture, Code & Test Review Report: Recent commits (last 3 hours)

## 1. Executive Summary

The recent commits add useful event and comparison features, with substantial subtraction in the debug prompt refactor. The primary risk is live polling work: each 2-second tick now reconstructs and queries a full trace for every session in the live tree, even when no subscriber needs trace events. The trace query also performs a child-session lookup whose errors are silently discarded. This creates avoidable SQLite load and makes link completeness dependent on transient DB success.

The configured Go test profiling command was run. The top individual tests complete in 0.09–0.32s. A full `go test ./...` attempt exited before testing two embed-dependent packages because the frontend bundle was absent in the checkout; the recent CI change does build that bundle before Go checks, which addresses the CI path.

## 2. Test Suite Performance Diagnostics

- **Total Go command elapsed:** 2.83s for `go test ./...`, ending with setup failures in two embed-dependent packages.
- **Profiled package tests:** all other listed packages passed; top individual tests are below 0.33s.
- **Target execution time:** no project target is specified.

### Slowest Tests & Bottlenecks

| Duration | Package | Test Name | Diagnosed Bottleneck | Recommendation |
|---|---|---|---|---|
| 0.32s | `internal/adapters/git` | `TestResolveBaseBranch_ProbeCandidates` | Git candidate probing in test fixture | Acceptable at current scale; avoid subprocess repetition if this grows |
| 0.30s | `internal/adapters/backfill` | `TestBackfillEngine_MultiMachineLifecycle` | Backfill integration fixture and filesystem work | Keep as integration coverage; separate pure transformations if timing rises |
| 0.22s | `internal/adapters/git` | `TestComputeSDLCIterations_ZeroCallSitesOrExports` | Static repository scan | Consider narrowing scanned files if this becomes a frequent bottleneck |
| 0.19s | `internal/adapters/git` | `TestInferTicketType_AdversarialEdgeCases` | Subtests share package setup | No immediate optimization needed |

### Test Performance Findings & Concrete Patches

1. **Embed prerequisite in local Go test command**:
   - Problem: `go test ./...` fails package setup for `cmd/agentic-lens` and `internal/adapters/web` because `//go:embed all:dist` has no matching directory in a clean checkout.
   - Evidence: `backend/internal/adapters/web/embed.go:9`; command failure: `pattern all:dist: no matching files found`.
   - Prescription: run the frontend build before standalone Go checks, as `.github/workflows/ci.yml` now does. A Make target for backend-only checks should share that prerequisite to make local and CI invocation consistent.

## 3. Architecture & Code Smells by Severity

### [High]: Per-poll full trace reads scale with every active session

- **Tier:** Tier 2 (Module Depth / Runtime Cost), Tier 1 (Boundary discipline)
- **Affected files:** `backend/internal/adapters/live/service.go:144-165`, `:201-240`, `:385-395`
- **Observed smell:** `pollAndBroadcast` unconditionally calls `broadcastTraceUpdates` every poll. That walks every root and descendant and invokes `GetSessionTrace` for each session. `GetSessionTrace` checks existence, fetches all messages and parts, then separately queries child IDs. With the configured 2-second poll, this makes full-history reads proportional to sessions × poll frequency, whether or not any trace changed. It can overwhelm SQLite as session count or history length grows and can delay the polling goroutine from catching subsequent ticks.
- **Subtraction target:** Remove the unconditional full trace reconstruction from the session-tree poll path.
- **Proposed target shape:** Track trace freshness/change at the SQLite event boundary (or use a bounded per-session incremental cursor), and query only sessions with new events. If full snapshots remain necessary, gate by active subscribers and apply a bounded concurrency/budget per tick.
- **Benefits:** Reduces repeated SQLite reads and JSON allocations, keeps tree/anomaly polling responsive, and makes live update work proportional to changed data.
- **Blast radius:** Live SSE consumers and trace retrieval. Preserve initial trace hydration and event ordering; verify reconnect and child-session behavior.

### [Medium]: Child-session lookup failures silently omit task links

- **Tier:** Tier 1 (Boundary error handling)
- **Affected files:** `backend/internal/adapters/live/service.go:389-395`
- **Observed smell:** A failed `ResolveChildSessionIDs` is converted into an empty set with no surfaced warning/error. `GetSessionTrace` then returns a successful trace with missing `ChildSessionID` values, indistinguishable from no child sessions being present. This hides DB/schema faults and makes task-link quality nondeterministic.
- **Subtraction target:** Remove the silent `childErr == nil` suppression.
- **Proposed target shape:** Either return a wrapped error from `GetSessionTrace` when child IDs are required for its contract, or include an explicit partial-data/degradation signal and report the underlying error through the adapter's established diagnostics path.
- **Benefits:** Makes telemetry completeness observable and prevents silently incomplete trace views.
- **Blast radius:** Live trace endpoint and SSE snapshots. Decide whether child links are required or optional, then test the selected error contract.

### [Low]: New persisted domain fields are weakly typed and partly unpopulated

- **Tier:** Tier 3 (Domain Modeling)
- **Affected files:** `backend/internal/domain/models.go:61-80`
- **Observed smell:** `SkillFact.CallerRole` is a raw string despite `domain.AgentRole` existing. `TaskLinkFact.CostUSD` is defined but extraction does not populate it, and there is no visible source of that value in the OpenCode event input.
- **Subtraction target:** Delete `CostUSD` until an actual source/requirement exists; use `AgentRole` for `CallerRole`.
- **Proposed target shape:** Keep the persisted model limited to values derivable from extracted events, with existing domain types reused for role classification.
- **Benefits:** Avoids implying telemetry is measured when it is always zero and strengthens compile-time consistency.
- **Blast radius:** JSON ledger consumers and type references. If `CostUSD` has an external planned producer, document/version that contract before removal.

## 4. Top Remediation Recommendations

1. Replace full-tree per-tick trace reconstruction with incremental or change-triggered fetching.
2. Give child-session lookup failures an explicit error or partial-data contract.
3. Remove the currently unsupported `TaskLinkFact.CostUSD` field and type the caller role as `AgentRole`.
4. Make local Go test entry points generate the embedded frontend assets, matching the CI workflow.
