# AL-57: Evidence-aware telemetry cost

Status: APPROVED baseline, amended at first approval by the user's instruction: “remove legacy_unverified type. treat legacy costs as observed. keep it simple”. Implementation is authorized under this contract and iterations/02/PLAN.md. The unapproved iteration 01 proposal remains historical context only.

Baseline: 45cb679f21475cd14b7dd4cbc25996a39d734403. Branch: fix/AL-57-cost-semantics. Workspace: alspaces/AL-57, from the configured workspaces.dir.

## Authority and scope

The user selected AL-57, requested a contract and plan, then approved the proposed scope with the explicit simplification quoted above. This acceptance supersedes the prior cost implementation hold and the proposal's separate historical-cost classification. AL-55 is delivered and has no cost dependency. AL2 development remains queued.

The existing AL-57 specification requires unavailable cost to differ from numeric zero, safe historical compatibility, and consistent CLI/API/UI behavior. This accepted contract refines its mixed-total rule and versioned representation. Explicit current user steering governs historical numeric values, including historical Codex zero: treat them as observed, without a separate historical state, counter, label, or amount.

Included: source extraction, persisted cost evidence, pure aggregation, explicit migration/backfill compatibility, all existing cost displays and consumers, and embedded frontend delivery. Excluded: price catalogs, API-equivalent estimates, subscription allocation, simulated costs, verified invoice billing, unrelated telemetry changes, and rewriting repository history.

## Fact semantics

Cost means reported USD, including accepted historical numeric values. A number is not proof of a settled invoice. Use a typed evidence state at each invocation: observed, partial, or unavailable. TypeScript uses a discriminated union; Go uses a typed value and validated constructors/accessors with value semantics. Bare floating-point zero is not the default for newly unknown cost.

| State | Meaning | Complete observed amount | Known source subtotal |
| --- | --- | --- | --- |
| observed | Every contributing current source record supplied a valid amount, or a valid legacy numeric amount is present | Numeric, including zero | Same amount |
| partial | Some contributing current source records supplied amounts and others did not | Unavailable | Sum of supplied amounts, including zero |
| unavailable | Current source cost or legacy numeric cost is absent/null | Unavailable | Unavailable |

Both OpenCode schema families currently coalesce source messages by model before producing invocation rows. A row with one reported 2 USD contribution and one missing contribution is partial, with known subtotal 2 USD, not observed 2 USD and not a fabricated zero. User messages, tool events, and idle records are not additional cost-bearing records. Cost-bearing records are assistant execution/usage records entering the owning extractor's model accumulator; absent cost in such a record is unknown even when token counts are zero. Existing empty-session fallback rows without measured cost are unavailable. Preserve the existing invocation grouping and row multiplicity.

Presence determines availability: explicit source numeric zero is observed zero; absence/null is unavailable. A fresh Codex invocation is unavailable because its rollout records do not establish USD cost. An OpenCode source label alone establishes neither presence nor completeness. Nonfinite/invalid costs are rejected with explicit diagnostics rather than enter arithmetic as zero. Valid legacy numeric values, including zero and historical Codex zero, are observed by the user's compatibility policy regardless of harness. Missing/null legacy values remain unavailable; decode their presence before any scalar default can invent zero.

Persistence: schema v5 retains invocation cost_usd as number or null and adds required cost_evidence metadata containing the typed state and known subtotal for partial evidence. Observed rows store their complete amount; partial/unavailable rows store null. Migration preserves each valid legacy numeric amount in cost_usd and marks it observed. No separate historical amount is introduced. API fact consumers inspect evidence; aggregate cost_usd is the nullable complete observed amount. Cost fields and evidence combinations are validated together; contradictory values are rejected at the boundary. Concrete type layout may vary without changing these meanings.

Evidence metadata must stay minimal: state and the partial known subtotal establish the required cost semantics. Existing session source identifies the harness; no historical provenance reconstruction is required. Do not store raw prompts/transcripts or invent message identifiers after aggregation. Source ownership remains the containing session's exact ID/source; rows retain their occurrence within that session. No invocation ID synthesis, fuzzy deduplication, or new billing record dataset is required.

## Aggregate contract

Every cost-bearing aggregate exposes nullable complete cost, nullable known source subtotal, state, and invocation coverage counts. Counts are fully_observed, partial, unavailable, and total. The three state counts sum to total. A complete total exists only for a nonempty selection of entirely observed invocation facts. Empty is an explicit aggregate state, with unavailable complete amount and subtotal, not a measured free run.

Known subtotal sums observed amounts, including accepted legacy numeric values, and evidence-backed contributions from partial facts exactly once. It is unavailable when no selected fact contains observed contributions; an observed contribution of zero yields a known subtotal of zero. Historical numbers use the ordinary observed display and arithmetic without special labels or counters.

Coverage counts invocation facts, including separate model rows and legitimate repeated rows, never sessions or tool calls. A partial invocation is not a fully observed invocation. Coverage is not a percentage of financial spend or source-message completeness. Present explicit counts when partial facts exist, such as one complete, one partial, one unavailable out of three invocations. Filters, time buckets, roles, models, and thread trees use exactly the included facts; repeated role/model/thread projections are not additional money. Fleet rollups merge the same cost summaries/counts and do not sum both projections and underlying facts.

Round only at existing display/aggregate precision boundaries, preserving recorded legacy values and current extraction precision. Merge subtotals before rounding so composition does not lose contributions.

| Synthetic input | Known subtotal | Complete total | Invocation coverage |
| --- | --- | --- | --- |
| Observed 2 + unavailable | 2, labeled partial | Unavailable | 1 observed, 1 unavailable |
| Observed 0 + unavailable | 0, labeled partial | Unavailable | 1 observed, 1 unavailable |
| Observed 0 + observed 2 | 2 | 2 | 2 observed |
| One coalesced row: reported 2 + missing | 2, labeled partial | Unavailable | 1 partial |
| Observed 1 + preceding partial row | 3, labeled partial | Unavailable | 1 observed, 1 partial |
| Legacy numeric 7 + observed 2 | 9 | 9 | 2 observed |
| Historical Codex numeric 0 | 0 | 0 | 1 observed |
| Legacy missing/null + observed 2 | 2, labeled partial | Unavailable | 1 unavailable, 1 observed |
| Only unavailable facts | Unavailable | Unavailable | Explicit unavailable counts |
| No invocation facts | Unavailable, empty | Unavailable, empty | Total 0; no coverage percentage |

Derived cost averages, percentages, efficiencies, and distribution shares require complete compatible amounts and valid denominators. Do not divide by a partial subtotal or turn an unavailable denominator into zero. A complete observed zero remains a known cost but a zero denominator has an explicit undefined/unavailable ratio. Existing non-cost metrics remain usable. The current cost-based role-effort distribution becomes unavailable with an explanatory reason when its subagent selection is incomplete or its complete denominator is zero; token/turn measures must not silently replace it.

## Compatibility, migration, and source refresh

CurrentSchemaVersion is 4. Implement a v4-to-v5 cost-evidence migration in the existing registry and engine, with generation/persistence targeting v5. Do not recreate v4 or alter AL-55's accepted skills/task normalization.

The approved runtime compatibility expansion keeps valid canonical v4 readable through a version-specific, read-only projection: valid numeric costs become observed, missing/null cost becomes unavailable, and files are not rewritten or made source-dependent. This explicitly expands the present current-version-only runtime gate. Presence-aware decoding must precede float defaults on both read and explicit migration paths. v1-v3 and malformed/interim ambiguous shapes still require the existing explicit migration workflow; do not add browser-side legacy ingestion. Existing partial/pending recovery metadata remains pending and subject to its established eligibility rules.

Explicit migration preserves valid original numerical values, session IDs, hierarchy, source labels, invocation order/multiplicity, timestamps, code facts, optional skills/task data, and unrelated harnesses. Shape conversion does not prove source recovery and must not create a permanent source-recovery requirement solely because cost is unknown. Unknown cost is supported uncertainty, not structural invalidity. Invalid costs are rejected explicitly rather than silently converted. Migration uses existing dry-run, ticket filtering, backup option, and atomic writes. Old binaries will reject v5: document that compatibility limit and retain recoverable original files through the migration backup workflow.

Evidence recovery occurs only through explicit source-backed refresh using existing exact-owner complete-tree recovery. Refreshes replace successful owners' cumulative snapshots; unavailable, malformed, or unresolved owners retain their prior facts and diagnostics. Partial cost within a successfully recovered tree is valid partial evidence, distinct from a failed/incomplete tree lookup. Repeated unchanged refresh is idempotent. A blanket conversion of historical Codex zeros to null is forbidden; fresh exact-tree source extraction may replace a proven placeholder while preserving the recoverable original and describing the change. No ordinal-only join of old rows to new rows when source grouping changes.

## API, CLI, and presentation behavior

All existing cost fields in derived totals, role/model/thread metrics, fleet cost maps, comparison rows, and bucket/cumulative progression become nullable complete amounts paired with the same cost summary contract. Update Go HTTP serialization and TypeScript API validation/types together. This is an intentional JSON/API cost contract change; document it for external consumers and ship the bundled frontend in the same deliverable. An old numeric-only external client must upgrade rather than receive invented zero compatibility values.

Human headlines show complete reported cost or explicitly labeled known subtotal with coverage. All unavailable/empty states say Unavailable or No invocation data. Valid historical numbers use ordinary observed formatting, including zero. CLI text/Markdown and dashboard use the same semantics; neither claims verified invoice spend.

Only complete observed costs participate in cost rankings and numeric cost coordinates. In either sort direction, incomplete/empty rows follow complete rows, with stable identity ties. Cost charts use gaps/omissions with excluded-count explanations, not zero or lines spanning missing values. Known-subtotal series may continue only when explicitly labeled partial. Once cumulative complete cost becomes unavailable, a later known bucket cannot make it complete again. Non-cost chart axes and filtering remain usable. Existing zero-safe log behavior may plot observed zero; do not discard its known state.

Cover task-link cost displays too: current TaskLinkFact.CostUSD has no source producer and omits zero in JSON. Decode the optional field by presence: missing/null task cost is unavailable; a valid historical numeric task amount, including explicitly present zero, is observed. An omitted old zero is unavailable because its original presence cannot be recovered. No special historical type or display is added. Task links are auxiliary references, excluded from invocation totals; do not derive task spend by adding a child tree already counted elsewhere.

## Architecture and guarantees

Pure cost states, summary algebra, availability rules, ordering, and ratios belong to backend/internal/domain. Sources, JSON validation, migration I/O, filesystem, HTTP and CLI orchestration remain in adapters/cmd. Go/TypeScript use concrete strict types, zero any, and zero emojis. Persisted facts remain the read baseline even when optional enrichment fails; GET never writes a ledger. Invalid individual ledgers produce per-ticket diagnostics while valid siblings remain visible.

## Approved boundary and steering

The user's approval response accepts subtotal plus coverage, partial coalesced-row evidence, v5 persistence, read-only canonical-v4 compatibility, and coherent CLI/API/UI rollout, amended to treat valid legacy numeric costs as observed and remove historical special treatment. Implementation may proceed without repeating the approval question. Live historical workspaces are not test fixtures and are not bulk-migrated merely to verify implementation.

This first approved baseline supersedes iteration 01's proposed historical classification. Current invocation states are only observed, partial, and unavailable; valid historical amounts participate normally in totals and cost-dependent calculations. Historical Codex zero remains observed unless an explicit successful source-backed refresh replaces its owning snapshot with current evidence. No general conversion by harness is permitted.

Changes to legacy classification, coverage denominator, explicit migration safety, or the v4 runtime compatibility policy require renewed approval. Routine implementation layout and fixes within accepted boundaries do not. Iterations/02/PLAN.md records the acceptance delta and supersedes conflicting historical planning text.
