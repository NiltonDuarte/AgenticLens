# AL-71: Technical design

## Scope and boundaries

Implement the [dashboard query](specs/dashboard-query-reliability/spec.md) and [live reporting](specs/live-source-reporting/spec.md) deltas. Preserve configured ticket matching, tag and prefix catalogue behavior, AL-60 exclusion search, existing trace and enrichment ownership, structured anomaly evidence and historical access. Keep benchmark arithmetic, live source selection, Go date semantics for existing API callers and ledger storage outside this change.

## Decisions and ownership

- **Request ownership.** Follow Fleet's generation and teardown pattern in Comparisons and Progression. Capture ownership before each request and test it before any settled state change, including catalogue updates and loading finalization. Retire the old owner when the selection changes, on refresh or retry, and on teardown. Preserve each view's existing presentation during a refresh and keep Fleet's summary and workspaces paired. Share the request lifecycle or filter parameters only where the views need the same invariant, using existing configured prefix and tag helpers.
- **Progression range.** Request all dates from `/api/workspaces` for the applicable type and prefix filters. Resolve the effective selection with existing tag handling, then calculate its UTC bounds in the progression domain. Use the newest usable ledger reference, with completion, start and update precedence per ledger. This reference sets the window without assigning a completion time to incomplete records. Preserve existing 24-hour, UTC date-only and custom boundary conventions, including open intervals. Apply bounds to recorded invocation dates using existing completion fallback and to delivered code using ticket completion. Date changes can reuse the fetched population. Show the range used by the chart.
- **Live ingress.** Reuse the tree, anomaly and trace boundary functions for REST and SSE. Decode required nested collections and every typed field consumed by the UI before emitting an observation. Require the trace's own identity. Keep the supported null empty `recent_errors` representation, optional metadata and partial relationship diagnostics. Validate against Go-produced wire data so the decoder preserves real zero values, source strings and documented omissions. Keep malformed payload errors separate from successful empty observations and structured source diagnostics.
- **Polling.** Reuse the configured cadence and existing `LiveReadOrder` ordering across automatic, manual and SSE tree reads. Skip timer ticks while the active subscription has an unfinished fallback poll. That gives each subscription at most one poll awaiting completion, with the next regular tick eligible after settlement and no backlog proportional to response time. Preserve the existing failure threshold and apply success, failure and status only to accepted reads. Teardown clears scheduled work and retires pending results. This bounds the browser's polling contribution per subscription without increasing per-read tree work or altering selected-trace refresh.

The implementation coder owns product paths in `web/src/views/`, `web/src/api/` and `web/src/domain/`, with `web/src/hooks/` available for a shared request lifecycle. The tester owns the affected frontend tests, the server date or live wire contract tests under `backend/internal/adapters/{http,live}/` and their self-contained fixtures. The orchestrator owns OpenSpec artifacts, integration and delivery. Independent reviewers investigate without editing product paths.

## Implementation approval

The user explicitly approved implementation with “$agentic-lens-dev aprovado” on 2026-11-23, after reviewing the proposal handoff. The approval covers the scope and acceptance linked above.

## Implementation approach

Revalidate the historical findings at the candidate revision. At planning baseline `db4477fa`, Fleet already fenced metrics results, and the live API already validated trees and anomalies, restored polling status when an accepted read recovered, and suppressed obsolete completions. Keep those guarantees. A direct adapter probe reproduced synthetic empty traces from null bodies, unchecked nested messages and parts, and overlapping fallback reads.

Verify request ownership with deferred responses through the views, progression against the real workspace handler, and decoder compatibility with a real native live wire fixture. Exercise polling with controlled time, in-flight teardown and ordering against an independent read or observation. Select the [polling and asynchronous-update](../../../docs/REGRESSION_PREVENTION.md#polling-or-asynchronous-updates) and [embedded frontend](../../../TECHNICAL_GUIDELINES.md#5-native-manager-engine--static-asset-bundling) boundaries for delivery.

Independent review covers filter and refresh ordering, date attribution across the workspace HTTP boundary, producer and decoder compatibility, fallback cadence and teardown, and the ability of the tests to catch these regressions. Behavior, data integrity, architecture and test effectiveness each have a separate evidence owner.

## Steering

The exploration compared current-clock presets with historical presets. The user asked to proceed with the proposal after the recommendation to anchor to the newest selected ledger and expose the resolved range. This resolves the roadmap ticket's open anchor choice, supersedes the filter-dependent clock, and preserves the existing unfiltered historical behavior. Verification must hold the selected results constant with a filter while the server clock is remote from the fixture dates, then remove the anchor to prove the intended shift.
