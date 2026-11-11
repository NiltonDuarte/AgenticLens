# A12 architecture review

**Outcome: APPROVED (scoped architecture review).** Reviewed AL-70 iteration 03 against baseline `54d58be40c74864d645c5168ba23820a21389c01`, candidate fingerprint `974d1449fc440115b563067461d43b9b6381baf60e178ff45e6854f884b10e6a`. All 29 current paths in the supplied frozen manifest match their declared SHA-256 values. The diff artifact SHA-256 is `37e968da167087ccaaffde0b817bb6fe3768d3b40c58875c62793a6fd7f43b8e`. Scope: OpenCode/Codex model identity extraction, live resolution/formatting, and frontend display consumers. No product or test files were changed. Coordinator run ID and requested model/effort were not included in the handoff, so I do not infer them from the persona name.

## Findings

No blocking architecture or maintainability findings in scope (severity 0). The Go producer changes store the canonical empty key in the existing model fields/maps; aggregation still groups by raw `Model` string. Codex `assignTurnModel` distinguishes map absence from a present empty value and only moves provisional empty-model turns when a nonempty observed model arrives. `turnNumber` likewise checks map presence, so empty identity is not confused with absent bookkeeping. This retains existing structures and avoids a schema or compatibility layer.

The TypeScript `formatModelLabel` is used at presentation sites and does not alter keys used for colors, grouping, sorting, or React identity. The Markdown formatter has the analogous display-only conversion. The prompt generator deliberately maps an explicitly supplied empty model to `Unknown`; its session overview model is required, while optional issue/turn model fields still omit only `undefined`. I found no path that rewrites stored ledger data or merges empty, literal `default`, and literal `unknown` groups.

## Hypotheses assessed

- Empty string might collapse into “missing map entry” and drop attribution. Refuted by the explicit `present` check in `assignTurnModel`, the present-key check in `turnNumber`, and existing map-key grouping.
- Human labels might replace raw keys before grouping/sorting/color selection. Refuted by tracing the frontend call sites: formatting occurs in rendered text; raw `m.model` remains the table key/color input. Markdown sorting also precedes label substitution.
- A new helper/schema abstraction might duplicate identity policy or add compatibility burden. Not substantiated: the helper is a small presentation function; the backend uses existing string identity fields, with no new schema or migration.

## Verification and limits

Reused the supplied tester evidence: backend command `cd backend && go test -count=1 ./internal/domain/... ./internal/adapters/codex ./internal/adapters/opencode ./internal/adapters/live ./internal/adapters ./internal/adapters/http` passed; frontend unit command `cd web && npm run test:unit -- src/domain/formatters.test.ts src/utils/debugPrompt.test.ts` passed (112 tests); focused integration command `cd web && ./node_modules/.bin/vitest run --config vitest.integration.config.ts src/components/Components.test.tsx src/components/LiveSessionTree.test.tsx src/components/LiveTraceInspector.test.tsx src/views/TicketDeepDive.test.tsx` passed (149 tests); `git diff --check` passed. Independently verified the manifest path hashes and baseline revision. No additional tests were run. Review is limited to the frozen A12 diff and assigned architecture focus; it does not certify other review focuses, final `make`, integration against `main`, or standalone-binary delivery.
