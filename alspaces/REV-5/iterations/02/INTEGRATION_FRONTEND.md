# REV-5 frontend integration

Parents: branch aa86f4b and main47af923. Ownership web/** only; no staging, commit or publication. Preserved other agents' changes. This report records integration decisions, not an exhaustive upstream correctness proof.

## Resolution and review

Resolved client.ts conflicts in favor of active REV-5 fetch functions. Removed upstream response-reader wrappers from costValidation.ts and exposed two pure cost validators consumed by the existing shape decoders. Response bodies are read exactly once. Cost checks run after known Go nil collection normalization and before returning typed domain data; schema-v5 required evidence, null/zero contradictions, finite nonnegative amounts, summary count/state agreement, nested models and opaque map keys retain upstream validation. Non-cost shape guards remain intact.

Shape checks now accept nullable metric/fact/map costs and nullable role effort shares. Optional role_effort_unavailable_reason is validated. InvocationFact.cost_usd is optional in the TypeScript model because upstream deliberately accepts omitted legacy invocation cost; v5 cost validation still requires both amount and evidence. Missing legacy cost stays absent/unavailable, never becomes zero. One progression test fixture now maps an absent invocation cost to null when constructing its required aggregate amount; numeric inputs retain their previous assertions.

fetchComparisonRows, fetchProgressionLedgers, fetchWorkspaceAnomalies and fetchProgressionMetrics stay removed: no new application caller exists. Their five response-reading wrappers likewise do not survive as test-only production code. Source named-function inventory parsed 93 TS/TSX files and found no zero-production-reference candidate; this conservative name scan is not a binding-aware whole-program proof.

Reviewed automatic changes in cost.ts, formatters, progression/scatter derivations, CostKPICard, distribution/chart rendering, comparison average, and cost display call sites in deep dive/thread/task panels. Kept complete-only comparison Average Spend (observed zero included), missing-value chart gaps, known-subtotal coverage and partial/unavailable rows in non-cost results. No unrelated upstream redesign was performed. Existing malformed-diagnostic warnings, valid sibling data and omitted artifact-description recovery remain.

## Test decisions

- Retained new upstream fleet cost tests for opaque role/model keys, partial metadata and missing summary rejection.
- Retained v5 fact/task evidence cases, legacy missing/null/zero cases, and nested thread-model rejection at active workspace fetch boundaries. Removed only the deleted progression-client entry from the shared financial-location table; workspace ledgers and projected rows still exercise that same oracle.
- Upstream comparison-client summary/count and opaque source assertions now have an active recipient: the workspace projection table covers empty, complete zero, partial and unavailable summaries; checks cost-summary contradiction, non-cost malformed tokens and exactly one JSON read. Source-event/artifact opacity remains covered by the active detail test.
- Corrected upstream opacity fixture from an invalid event containing only source metadata to a valid event plus that metadata; retained the exact opaque source assertion. The general event guard was not weakened to admit malformed events.
- Did not restore obsolete progression/comparison network/query tests. Active fleet/workspace query/error tests remain; original REV-5 removal rationale still applies. Upstream time-bucket HTTP-reader tests are obsolete with their unused reader. Active progression.test.ts retains cumulative unknown cost behavior, and backend progression HTTP tests own emitted bucket contracts. No claim that an unused client validator still exists.

## Verification

- Typecheck and ESLint passed after conflict resolution.
- Focused Vitest run: 232 cases across client, formatters, progression, TicketComparisons and TrendChart passed.
- Canonical frontend build passed; web/dist matched backend/internal/adapters/web/dist. Vite emitted its advisory chunk-size warning; no build stage skipped.
- Final typecheck/lint/client/build rerun after simplifying the cost-validator export passed, including all 110 client cases. No runtime semantics changed in that simplification.
- git diff --check for web passed; no conflict markers remain in frontend source. Index conflict flags remain for parent staging.

Actual merged-Go wire smoke: built current backend with go build -o /tmp/rev5-merge-cli ./cmd/agentic-lens. Compiled validation.ts/costValidation.ts to a temporary CommonJS output with TypeScript. Created a temporary adopting project from root configuration and three independent copies of static testdata/workspaces/AL-1 facts: AL-901 schema-v4 with every reported cost zero; AL-902 schema-v4 with invocation cost omitted; AL-903 schema-v5 with partial evidence and null complete amounts (migration metadata version updated accordingly). Started the freshly built CLI server with this workspace and an unavailable temporary SQLite path. Queried /api/fleet, /api/workspaces and all three /api/workspaces/:id endpoints, and passed each actual JSON body through the merged decoders. All accepted. Returned totals were respectively numeric 0/complete, null/unavailable and null/partial. Server stopped and temporary project removed. No live ticket history or developer source database was used. This smoke proves these real serialized wire shapes, not provider hydration, browser layout or all possible persisted records.

Independent tester/reviewer and full make remain parent gates. Test/source editing ownership released after handoff.
