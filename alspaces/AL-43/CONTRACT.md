# AL-43 Technical Contract

## Approval and baseline

The user explicitly pre-approved the ticket plan and contract and authorized full implementation and delivery on 2026-10-31. This contract records the supplied specification's boundaries; it does not require another approval round. Initial checkout was clean at 85a2af94b5036b0918aaf9a7a89aed3a990f2650 on AL-43/surface-semantic-content-errors. Workspace root is configured as alspaces.

## Behavior

Successful tool invocations can contain semantic diagnostics without failing execution. Introduce pure ClassifyToolContent with HasContentError, FailureSeverity, Category, and optional ErrorSnippet. Recognize compiler diagnostics, stack traces, panic messages, unhandled exceptions, and test assertion failures conservatively. Inspection/search tools returning source or documentation must avoid matching incidental error words. Matching output is evidence of a content warning, not proof of process failure.

Classify completed successful tools separately from explicit execution failures. Preserve completed status and execution success metrics for content-only warnings. Content warnings do not advance execution failure streaks or inflate execution anomalies. Explicit failed status and authoritative nonzero execution outcomes retain execution failure behavior. Avoid double counting a failed invocation as a successful content-warning invocation.

Expose ContentErrors separately from Errors on tool facts and expose HasContentError, ContentSeverity, and optional ContentErrorSnippet on event and live-part facts. Carry these fields through session-owned canonical facts, DTO conversion, persistence, read projections, aggregation, HTTP, and frontend types. Missing fields in legacy v6 ledgers default to zero/false. Keep schema compatibility and original source/session identities; optional enrichment must not remove persisted content diagnostics.

Dashboard traces show amber CONTENT ERROR and red EXEC ERROR indicators. Live inspector shows separate Failed Tools and Content Errors counters. Deep Dive exposes content totals in tool execution presentation. AnomalyDrawer explains semantic warnings separately from process failures while preserving source evidence. Filters must allow content warnings to be discovered and inspected.

## Architecture and exclusions

Use Go pure deterministic domain classification and aggregation, adapter-owned source parsing and JSON validation, and strict TypeScript presentation. Follow TECHNICAL_GUIDELINES.md, including zero any, zero emojis, immutable fact ownership, and embedded frontend assets. No new execution anomaly rule, schema migration, unrelated reliability metric redesign, or speculative source support is required. Supported OpenCode schema families and Codex tool result paths that already produce tool facts must preserve the same distinction.

## Acceptance and verification

Prove completed compiler/traceback output becomes a content warning with unchanged execution success; prove explicit failures remain failures and incidental inspection text stays clean. Prove legacy defaults and populated canonical ledger persistence/API fields survive real production conversion paths. Prove live/offline consistency, independent counters, amber/red rendering, discoverable content-warning traces, and separate diagnostic copy. Reuse coherent existing tests, add only missing regression guards, run focused checks and the required integrated make gate, build and inspect ignored embedded assets and standalone binary. Parent owns telemetry, commits, plan lifecycle, and PR delivery.
