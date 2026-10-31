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

## Superseding acceptance: iteration 02, authorized feature removal

Source: the user reviewed PR #78's content-classification approach, questioned its suitability as a platform feature, and explicitly instructed: "revert this feature". This supersedes the Behavior and feature-specific Acceptance requirements above; the preceding text remains historical evidence of iteration 01. The user's instruction explicitly approves the removal scope without another confirmation round. The baseline is clean commit 9a3796ebadad7f5eb002b764b43c28e0526efb86; the delivery branch is chore/AL-43-revert-content-classification. PR #78 is already merged, so delivery requires a new PR.

Remove automatic tool-output content classification and its category/severity/snippet propagation, content-warning counters, persisted metric production, API projections, UI badges, special filtering, and explanatory copy. Remove feature-only tests because their expected behavior is intentionally retired. Preserve execution classification, failed-call counts, success rates, failure streaks, execution anomalies, ordinary source evidence, and unrelated navigation, invocation badges, filters, and live refresh behavior.

Previously written supported ledgers must remain readable without source hydration or a migration. Retired nested content fields may be accepted only at the adapter's decode boundary and discarded; they must not become domain facts, new API fields, or new persisted metrics. Keep strict rejection of unrelated unknown fields and existing schema-version/ownership validation. No rewriting historical ticket ledgers, adding a replacement classifier, configuring language packs, or changing execution semantics is authorized. Preserve historical ticket evidence and the completed specification/index lifecycle.

Verification must establish removed behavior, supported ledger compatibility, preserved source/execution semantics, and existing interaction guarantees. See iterations/02/PLAN.md for the grounded sequence and selected regression cases. Existing architecture, source fidelity, typed boundaries, zero emojis, embedded distribution, final telemetry, and delivery ownership requirements remain in force.

## Superseding acceptance: iteration 03, execution-classifier correction F1

Source: manager review MGT-1 iterations/05/PR88_REVIEW.md finding F1 and explicit user/manager instruction: "retain main’s execution classifier while removing content classification, and add these regression cases." This approves correction without a further permission gate. Starting checkout is clean e91cf0d05348537657c210601c6acc93c01f967d on chore/AL-43-revert-content-classification. The reviewed reference main is 3857fa22f49f825c82ced39c3f81f963b7447545; the parent must fetch actual main before final integration.

Restore ClassifyToolExecution exactly from the reviewed main reference, including its existing failure/severity/message logic. This supersedes iteration 02's decision to restore pre-PR #78 panic, timeout, and permission substring heuristics. No execution policy expansion or content classifier is authorized. Successful read output containing a timeout identifier, grep output containing a saved panic example, and explicit-zero bash output containing permission-denied fixture text must retain successful execution and no failure severity. The corresponding live/offline calls must not inflate failed-call totals or failure streaks; explicit authoritative failures remain failures.

All content-feature removal, strict decode-only retired-field compatibility, source/raw Codex evidence, AL-55 recovery and optional-array normalization, AL-47 live behavior, and AL-53 navigation guarantees remain accepted. Historical plans and reviews remain evidence of their original heads. Iteration 03 records the changed execution acceptance and affected verification; parent updates existing PR #88 after focused verification, review, required integrated make, and final original-root telemetry. No PR merge, auto-merge, or watch is authorized.
