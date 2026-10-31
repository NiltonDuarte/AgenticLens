h3. Title: AL-68: Configure optional skills, task_links and anomalies output for generate

*TARGET*: AgenticLens (Configuration / Domain / Source Adapters / CLI / Persistence / Summary / Dashboard Compatibility / Tests)

*Phase*: Phase 1 (post-v2.0 platform review)
*Status*: Ready
*Work type*: feat
*Blocked by*: None
*Blocks*: None
*Origin*: PO-18 iteration 03; user requests config toggles for generate features including task_links, skills and anomalies. Disabled metrics must not be written.

*CONTEXT*

Operators need control over optional telemetry emitted by generate. The current command gathers skills/task links, derives metrics and persists anomaly summaries; --anomalies is a diagnostic display flag, not a configuration toggle for persisted anomaly facts. Output policy must consistently apply across source extraction, retained facts, persistence and generated summaries, including the granular ownership planned in AL-67.

*GOAL*

As an operator, I want independent configuration toggles for skills, task_links and anomalies so that generate writes only enabled metric categories. Keep existing output behavior when toggles are omitted, and distinguish disabled collection/output from measured zero or unsupported source evidence.

*VERTICAL SLICE SCOPE*

* Domain Core: Model an explicit typed generation policy and apply it consistently to optional fact/derived-output categories. Separate disabled, unavailable and recorded-empty semantics; keep unrelated metrics intact.
* Boundary Adapter: Load independent booleans from the existing project configuration through normal validation/defaulting. Pass the effective policy through OpenCode v1/v2 and Codex generation paths and output preparation; do not require a source to support a metric merely because its toggle is enabled.
* Persistence: Apply the policy to the complete emitted ledger, including nested session/invocation collections, derived duplicates and facts retained from an existing output ledger. Coordinate with AL-67's source ownership without relying solely on root-field omission.
* Presentation: Apply toggles to generate JSON and generated Markdown metrics sections; document effective settings and --anomalies interaction in help/config examples. Existing API/dashboard readers must accept intentionally omitted categories without implying successful zero measurements. No new dashboard settings UI is required.
* Quality: Focused config/domain tests and actual generate persistence/summary integration tests cover enabled/default/disabled combinations, existing ledgers, multiple owners/sources, and repeated runs. Fixtures remain self-contained.

*ACCEPTANCE CRITERIA*

* Project configuration provides independent boolean enable/disable settings for skills, task_links and anomalies. Missing settings preserve current behavior (enabled); explicit false survives defaults/merging. Invalid values fail with an actionable configuration error before output files are modified. Document the exact keys and effective configuration precedence in the developer contract and user-facing config reference.
* Disabling a category makes generate omit its structured facts and category-specific derived metrics from written output, rather than writing empty arrays, null placeholders or zero counts as if measured. Apply this at every canonical owner and any compatibility projection, including session/invocation scope introduced by AL-67. Minimal explicit disabled-policy metadata/diagnostics may identify intentional omission without containing the suppressed metrics.
* Each toggle works independently and all three can be off together. Enabled categories and core session/invocation identity, token/cost/duration/code metrics, source attribution and parent-child structure remain intact. Disabling task_links does not remove child sessions; disabling skills does not remove the ordinary invocation metrics for work that used a skill.
* Anomalies disabled suppresses persisted anomaly facts at root/session scope and anomaly-derived generated counts, breakdowns and health claims. Do not replace absent anomaly analysis with healthy/zero-error assertions. Ordinary tool execution/error facts remain valid independent telemetry; toggles do not redact raw transcript text or unrelated tool evidence merely because it mentions a category name.
* Regenerating an existing target output with a category disabled omits previously persisted facts of that category from the new output as an intentional policy choice; retained-session merging or optional hydration must not silently restore them during generate. Preserve existing safe-write/backup guarantees and all enabled/unrelated history. This deliberate configured omission is distinct from source-unavailable recovery, which must still retain enabled stored facts safely.
* Generated Markdown metrics sections reflect the effective policy, including when --append replaces an earlier generated section; disabled metric values must not remain stale in that managed section. Preserve user-authored summary prose. Any category-specific console diagnostic requested while disabled reports that it is disabled; existing --anomalies does not silently re-enable persistence or analysis. No new CLI override flags are required by this ticket.
* Re-enabling permits source-backed recovery of available category facts on a later generation; do not promise recovery of previously omitted facts when original source data is unavailable. Enabled but unsupported/unavailable facts retain their existing explicit availability semantics, not fabricated successful zeros.
* Configuration applies to generate; no automatic rewriting of other ledgers, global dashboard filtering, live feature suppression, or change to backfill defaults is included. Later explicit backfill/hydration policies are separate scope; document that these generation toggles are not a global source redaction guarantee. Consumers must continue to display supported persisted facts when enrichment is unavailable.
* Demonstrate config omission/defaults, each false separately, all false, mixed values, invalid types, and off/on regeneration using meaningful nonempty fixtures. Cover both first creation and existing target ledgers, nested owners, an unrelated source's retained session, safe failure and unchanged repeat stability. Assert emitted JSON field absence and managed Markdown omission rather than only in-memory filtering.
* Keep schema/read compatibility explicit if distinguishing disabled from missing evidence requires additive metadata; coordinate schema evolution with AL-67 and existing owners. Do not introduce a schema version solely by assumption. Pass required make gates and existing frontend compatibility checks; any frontend change includes canonical embedded-asset verification.

*DEFINITION OF DONE*

* Execute in a separate agentic-lens-dev session with a reviewed typed policy/config/output contract and implementation scope approval under the usual gate. This PO task authors the ticket only.
* Deliver configuration-to-output behavior and regression tests end to end, with documented keys, enabled-by-default semantics, intentional omission on regeneration, restoration limits and command scope.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md: FCIS, strict Go/TypeScript, zero any/interface{}, zero emojis, no domain mocks and self-contained fixtures. Preserve the original AL-55 evidence.
* Publish reviewed changes, verification evidence and source-grounded telemetry in the ticket's implementation PR. No developer workspace or CONTRACT.md is created by this PO session.
