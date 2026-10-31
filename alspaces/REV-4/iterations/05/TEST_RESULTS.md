# REV-4 iteration 05 independent test results

## Result and tested identity

PASS for the bounded test reduction. All 47 frozen candidates are accounted for: seven DELETE, 35 MERGE and five RETAIN. No missing transfer, weakened surviving oracle, unrelated receiver change, new test declaration or new helper suite was found. This report does not replace the independent reviewer verdict or the parent's required integration gates.

Baseline: `0c4d1fea90b8ec259e052c484361299435a7b84c`. Inspected the shared worktree after both coders declared their patches settled. Initial observed status contained the 15 authorized test paths, parent-owned SUMMARY.md and iteration 05 artifacts; these shared changes were preserved. The tester changed only this report. No application/test edits, new tests, commits, publication, telemetry or child agents were performed.

Read AGENTS.md, tester persona, configuration, TECHNICAL_GUIDELINES.md, root contract, iteration 05 contract/plan/candidate ledger and prior iteration 04 review. Configured workspace root is `alspaces`.

Test-only patch SHA-256 (Git diff over changed backend/web paths against the baseline): `31938d3409d37c6ef17fcb1f8e6cd4308556bd74971f24efbcffa6a693c60fb0`. Exactly 15 test files change: **231 added, 765 deleted, 534 net lines removed**. No production/runtime/schema/CI/Makefile/dependency/fixture-data path changes. `git diff --check` passed.

## Independent AST and source proof

Used Go go/parser + go/ast and the installed TypeScript compiler parser, independently extracting baseline files directly with `git show <baseline>:<path>` and current files from disk. Compared declaration identity multisets and body multisets, preserving duplicates with the same file/title; helper functions were counted separately. The direct-Git baseline exactly matches the supplied frozen body inventory.

The supplied fingerprint `5d66f496f12d93c6244b0484a131f61af2ebb6853e4ff2d06ce512658077b08d` was independently reproduced: SHA-256 of sorted `(file,line)` records formatted as `file:line:sha256(body UTF-8)\n`. It is a canonical source fingerprint, distinct from the raw JSON serialization SHA-256 (`c286e9af5891bc76097659567a20452de70c463a0b0013d7789485bf5eeb15a0`).

| Inventory | Baseline | Current | Removed | Added |
|---|---:|---:|---:|---:|
| Go top-level tests, excluding TestMain | 499 | 494 | 5 | 0 |
| TS declarations, before parameter expansion | 698 | 661 | 37 | 0 |
| Total declarations | 1,197 | 1,155 | 42 | 0 |
| Frontend expanded runtime cases | 749 (prior evidence) | 712 (observed) | 37 | 0 |

The 42 removed identities exactly equal the authorized non-RETAIN candidate multiset. No renamed/additional top-level tests, test.each replacements or helper wrappers were introduced. All **25 parameterized TS declarations** remain byte-identical. All **101 extracted non-test Go functions (including TestMain) and named TS function/arrow helpers outside executable test callbacks** have unchanged name/body multisets. Surviving bodies change only in the **23 authorized receivers** listed below: two Go and 21 TS; all other surviving test body multisets are unchanged.

For changed TS receivers, independently extracted original expect statements remain present after whitespace normalization. Two deliberate representation rewrites were inspected semantically: L14 replaces four true expectations with a four-member narrowing loop whose false branch throws; L46 binds the same role/model tooltip elements to variables before the unchanged placement assertions. No original receiver assertion is lost in either rewrite. The Go receiver diffs preserve their old checks and add only parser dispatch/default-order checks.

## Every MERGE obligation

Each row is a separate checked obligation; all 35 pass. Existing recipient titles are preserved and their current file/line identities appear below.

| Candidate | Observed surviving oracle |
|---|---|
| L05 | WARNING recipient asserts sm then md on rerender; CRITICAL asserts lg. Both original severity labels, classes and status-role queries remain. |
| L06 | InfoTooltip accessible render receives header-tip-custom and verifies both container classes. Original trigger/bubble/content/aria-describedby checks remain. |
| L09 | Existing mixed-case enum row keeps custom priority and adds the exact omitted-priority sorted/deduplicated ten-role result with length and element checks. No new row or t.Run. |
| L12 | Zero-cost recipient adds the title while retaining advisor/po labels, exact two zero percentages and no-NaN assertion. Other existing render cases cover canonical labels. |
| L13 | Same accessible InfoTooltip render verifies placement-top and rerenders for bottom/left/right with exact placement class mappings. |
| L14 | All four original rule IDs execute through the production guard; a compiler-checked ToolFailureAnomalyRule assignment and equality assertion run on success, and the false branch throws. Existing non-tool membership negatives remain. |
| L16 | claude-3-7-sonnet is added to the existing sample loop, which checks hex shape, repeatability and palette membership for each model. |
| L17 | Existing stronger zero-ticket/empty-list FleetOverview test now sets total_cost_usd to zero; heading and Refresh assertions remain. |
| L18 | Entire overflow-visible/not-hidden source fragment is byte-identical inside the existing responsive-wrapper recipient. Static source oracle is preserved. |
| L19 | Entire width/overflow/touch/min-width/nowrap source fragment is byte-identical in that same recipient. CSS is read exactly once for both fragments. |
| L20 | TestEnumsParsing now independently asserts ticket_id dispatch to SortFieldTicketID, retaining existing cost/default checks. |
| L21 | Same parser recipient independently asserts completed_at dispatch to SortFieldCompletedAt. |
| L23 | Existing clickable badge checks visible Clickable Warning text independently of its accessible button name, then retains callback counts. |
| L24 | Clickable badge checks exact default Click to open Anomaly Inspector title before its first callback click. |
| L25 | Same clickable badge rerenders with Custom Inspection Tooltip, queries current button, verifies title override and second callback execution. |
| L26 | WARNING status recipient rerenders with the original Static Informational Title and asserts it while preserving status/severity text. |
| L27 | Initial non-clickable WARNING render explicitly asserts absence of title before the custom-title rerender. |
| L28 | Single-item DistributionCard render retains model and 100.0% checks and gains exact tooltip content plus its title-specific accessible trigger. |
| L29 | Existing focus/Escape flow dispatches Enter, Tab, ArrowDown and space while focused, asserting visible after each; Escape and blur checks remain. |
| L30 | Dedicated Navbar workspace recipient already has custom/workspaces text, pill class/title/aria-label and both Live Server/Mock Mode absence checks; its body is unchanged. No additional transfer was necessary. |
| L31 | TrendChart recipient asserts supplied tooltip content/trigger with populated data and rerenders empty data with the same content/trigger; original hover/focus/Escape checks remain. |
| L32 | Same TrendChart recipient tests omitted and empty-string tooltip absence for each of populated and empty data, preserving all four negative branches. |
| L33 | Healthy SaturationBar recipient preserves original numeric/status/default-tooltip assertions, then reuses its fixture to verify exact custom tooltip override. |
| L35 | Existing populated ThreadViewer fixture gains the section trigger/content and all ten original header labels, aria links, bottom placement and exact explanations before both propagation checks. |
| L36 | Existing dropdown open/options test verifies arrow presence, aria-hidden=true, v closed, ^ open and v after closing. All original option/expanded assertions remain. |
| L37 | Complex-version Navbar recipient retains exact semver text/commit/date title and gains version-badge class; ordinary version fixture adds no distinct render branch. |
| L38 | Same version recipient rerenders with changed version, commit and build date, asserts updated title/text and absence of old version. |
| L39 | Existing project-click recipient asserts explicit name, badge class/title and permanent branding before its unchanged landing callback assertion. |
| L40 | Same project recipient rerenders without projectName using configured ConfiguredRepo, asserting configured badge/class/title and both branding labels. |
| L41 | Same recipient then supplies explicit empty projectName over populated configuration, asserting no project badge and persistent AgenticLens/Telemetry labels. |
| L42 | Existing TicketTable propagation fixture gains all eight original accessible header names, nonempty aria-describedby and bottom placement before its complete original sort/order sequence. |
| L43 | Existing TraceWaterfall badge-selection fixture checks the exact default inspection title before preserving selection count, payload and propagation isolation. |
| L44 | Existing populated MetricProgression render preserves all loading/KPI/chart assertions and gains all five exact accessible chart trigger/explanation pairs. |
| L45 | Same MetricProgression render retains Token Volume aria linkage, initially hidden bubble, focus-visible and blur-hidden sequence. Donor title mentions hover, but its body only exercised focus/blur. |
| L47 | Existing L46 role/model help-wiring test gains the original role hover enter/leave, focus/Escape dismissal and model focus/blur sequence, with initial-hidden checks and same left placement. |

No candidate requires additional retention. L30 is a pre-existing equivalent oracle rather than a receiver edit. L46 is the intentional RETAIN-body exception: its declaration survives and receives L47, as the frozen ledger requires. These facts do not change scope or scores.

## DELETE rationale checked against sources

| Candidate | Independent check |
|---|---|
| L01 | Both DistributionCard cases pass empty items/zero total and check the same empty notice; stronger original empty recipient remains byte-identical. |
| L02 | Existing granularity case clicks all three options with active classes and cumulative absence; initial dashboard recipient retains Total Invocations presence. |
| L03 | Existing empty-ledger recipient preserves zero-progression heading plus explanatory empty text. |
| L07 | InfoIcon remains a direct InfoTooltip alias; production reference search finds no importer. Only newly unused test import was removed. Dropping reference identity is an acknowledged implementation constraint reduction. |
| L08 | Both public ordering functions delegate to unchanged orderRolesHelper; existing comprehensive table independently invokes each API and covers priority, deduplication, arbitrary names and alphabetic remainder/defaults. Different advisor/researcher literals add no branch. |
| L10 | BuildInfo is a value struct containing only strings. GetBuildInfo returns copied values; later global replacement cannot mutate a previous value in Go. Custom-variable construction and fallback tests remain unchanged. This intentionally drops a language-semantics constraint. |
| L11 | Existing same mocked 404 branch still asserts Workspace AL-1 not found for the interpolated ticket ID. Removed NONEXISTENT-999 literal adds no distinct response/error branch. |

Consolidations remain within existing component/fixture contracts: badge props, individual tooltip contracts, one table header affordance, view chart help, and responsive DOM-plus-CSS guards. L46 retains its separate, substantial view help contract rather than absorbing unrelated KPI tests. No cross-component catch-all test or universal tooltip helper was introduced. Reduced failure isolation is an inherent tradeoff of these named consolidations; no mutation-equivalence or coverage-percentage claim is made.

## Five retained tests and compiler/CSS guards

All five declarations remain at the ledger's original scores (2, 3, 4, 4, 4). Four bodies are byte-identical; the fifth is the explicitly authorized receiver described above.

| Candidate | Current declaration | Score | Body SHA-256 evidence |
|---|---|---:|---|
| L04 | `backend/internal/domain/enums_test.go:484` / TestAgentRole_FallbackConstants | 2 | unchanged: `c21c7563b193c5abb749b6b5f8d9bbe240f32c45a164c733f2b865d12d6f3971` |
| L15 | `web/src/utils/theme.test.ts:15` / defines valid hex colors across all theme tokens | 3 | unchanged: `f3caa5ab23c483f7643c86d1fb3f35426f8da2248ccc0f7278f3cd99c70c23db` |
| L22 | `backend/internal/domain/enums_test.go:9` / TestCanonicalTicketTypes | 4 | unchanged: `4b73fe3c329a7282529d60036b3a718fcf38e5598914b9d25e4006588d8c10dd` |
| L34 | `web/src/components/Components.test.tsx:537` / renders empty notice and section header tooltip when threads array is empty | 4 | unchanged: `e7b48e7a65fc941aa751d0f49fd9bca40b092354ce4660e827afc7f35ed3483f` |
| L46 | `web/src/views/TicketDeepDive.test.tsx:794` / renders information tooltips on velocity, table headers, and refined KPI tooltips | 4 | baseline `f855d74b7a955dd48902ff51d6f7aa87d79e733e8de9a42aa7abe1ad6802b526`; extended `49a37111d1138b27562a11e1209330c964b27a0b20f475ba6493cf5b881fa155` |

L46 preserves all 33 original expect statements semantically (only two repeated DOM lookups became local roleBubble/modelBubble variables), and has 41 expect statements after the authorized interaction addition. Existing velocity/KPI content, tooltip header cardinality, explanations and placements are preserved.

The two baseline tests containing negative `@ts-expect-error` directives have identical bodies. The positive narrowed assignment remains inside the four-member loop. `web/tsconfig.json` includes all src files in strict/noImplicitAny/exactOptionalPropertyTypes mode, so the successful tsc run checks both positive narrowing and active negative directives rather than relying on Vitest transpilation.

The CSS regex-source fragments from L18 and L19 survive **byte-for-byte**, including comments, regexes, fallback empty strings and every assertion. Their fragment sizes/hashes are:

- L18: 671 bytes, SHA-256 `540bed8e87dcbf1a0b7a211277d2e80562e3251eb0f860c6f216c97472f8a050`.
- L19: 1,168 bytes, SHA-256 `fd78f3e79238ab63deebe36badb7c7814d395cbe8e069a83d99f34edf819ec9b`.

Both execute inside the existing responsive-wrapper test after its original two-table wrapper/header/placement checks, with one stylesheet read. These source guards establish CSS declaration constraints; jsdom does not prove rendered overflow geometry.

## Exact removed declarations

The following 42 identities are the complete baseline-minus-current multiset, with no additions. Candidate IDs map directly to frozen CANDIDATES.md.

| Candidate | Removed declaration |
|---|---|
| L01 (DELETE) | `web/src/components/Components.test.tsx` / renders empty notice when items dictionary is empty |
| L02 (DELETE) | `web/src/views/MetricProgression.test.tsx` / toggles granularity between hourly, daily, and weekly |
| L03 (DELETE) | `web/src/views/MetricProgression.test.tsx` / renders zero telemetry state when no ledgers exist |
| L05 (MERGE) | `web/src/components/AnomalyBadge.test.tsx` / renders small, medium, and large sizes according to size prop |
| L06 (MERGE) | `web/src/components/Components.test.tsx` / applies custom className to container element |
| L07 (DELETE) | `web/src/components/Components.test.tsx` / exports InfoIcon alias identically |
| L08 (DELETE) | `backend/internal/domain/aggregations_test.go` / TestOrderRoles |
| L09 (MERGE) | `backend/internal/domain/aggregations_test.go` / TestOrderSubagentRoles |
| L10 (DELETE) | `backend/internal/domain/version_test.go` / TestGetBuildInfo_ImmutabilityValueSemantics |
| L11 (DELETE) | `web/src/api/client.test.ts` / throws structured error when ticketId does not exist on 404 |
| L12 (MERGE) | `web/src/components/Components.test.tsx` / handles zero total cost without divide-by-zero distortion |
| L13 (MERGE) | `web/src/components/Components.test.tsx` / applies custom placement classes correctly |
| L14 (MERGE) | `web/src/domain/anomalies.test.ts` / narrows candidate to ToolFailureAnomalyRule via type guard across all valid members |
| L16 (MERGE) | `web/src/utils/theme.test.ts` / returns deterministic hex colors for models via getModelColor |
| L17 (MERGE) | `web/src/views/FleetOverview.test.tsx` / displays empty state when total_tickets is zero |
| L18 (MERGE) | `web/src/views/TicketDeepDive.test.tsx` / ensures .table-card and .thread-accordion-item in web/src/index.css have overflow: visible to prevent tooltip clipping |
| L19 (MERGE) | `web/src/views/TicketDeepDive.test.tsx` / ensures .table-responsive, .data-table, and .thread-models-table rules in web/src/index.css satisfy responsive constraints |
| L20 (MERGE) | `backend/internal/domain/aggregations_test.go` / TestParseLedgerSortField_TicketID |
| L21 (MERGE) | `backend/internal/domain/aggregations_test.go` / TestParseLedgerSortField_CompletedAt |
| L23 (MERGE) | `web/src/components/AnomalyBadge.test.tsx` / renders custom label string when label prop is provided |
| L24 (MERGE) | `web/src/components/AnomalyBadge.test.tsx` / renders default title attribute on clickable badge when title prop is omitted |
| L25 (MERGE) | `web/src/components/AnomalyBadge.test.tsx` / renders custom title attribute on clickable badge when title prop is provided |
| L26 (MERGE) | `web/src/components/AnomalyBadge.test.tsx` / renders custom title attribute on non-clickable badge when title prop is provided |
| L27 (MERGE) | `web/src/components/AnomalyBadge.test.tsx` / does not set title attribute on non-clickable badge when title prop is omitted |
| L28 (MERGE) | `web/src/components/Components.test.tsx` / renders InfoTooltip when tooltip prop is provided |
| L29 (MERGE) | `web/src/components/Components.test.tsx` / retains visibility when non-Escape keys are pressed while focused |
| L30 (MERGE) | `web/src/components/Components.test.tsx` / renders workspace pill when workspaces directory is configured |
| L31 (MERGE) | `web/src/components/Components.test.tsx` / renders InfoTooltip when tooltip prop is provided in both populated and empty states |
| L32 (MERGE) | `web/src/components/Components.test.tsx` / does not render InfoTooltip when tooltip prop is omitted or empty in populated and empty states |
| L33 (MERGE) | `web/src/components/Components.test.tsx` / renders custom tooltip prop on SaturationBar when provided |
| L35 (MERGE) | `web/src/components/Components.test.tsx` / renders InfoTooltip on section header and all 10 column headers of thread models table |
| L36 (MERGE) | `web/src/components/MultiSelectDropdown.test.tsx` / renders arrow indicator correctly: v when closed, ^ when open |
| L37 (MERGE) | `web/src/components/Navbar.test.tsx` / renders version badge with provided versionInfo prop and displays tooltip |
| L38 (MERGE) | `web/src/components/Navbar.test.tsx` / dynamically updates version badge when versionInfo prop changes on rerender |
| L39 (MERGE) | `web/src/components/Navbar.test.tsx` / renders project badge when projectName prop is provided |
| L40 (MERGE) | `web/src/components/Navbar.test.tsx` / renders project badge when projectConfig prop is provided |
| L41 (MERGE) | `web/src/components/Navbar.test.tsx` / preserves permanent AgenticLens and Telemetry branding when project is empty |
| L42 (MERGE) | `web/src/components/TicketTable.test.tsx` / renders InfoTooltip on all 8 table headers with placement="bottom" |
| L43 (MERGE) | `web/src/components/TraceWaterfall.test.tsx` / renders AnomalyBadge with title tooltip affordance on anomalous turns |
| L44 (MERGE) | `web/src/views/MetricProgression.test.tsx` / renders InfoTooltip triggers and explanations across all five trend charts |
| L45 (MERGE) | `web/src/views/MetricProgression.test.tsx` / handles focus and hover transitions on trend chart tooltips in MetricProgression |
| L47 (MERGE) | `web/src/views/TicketDeepDive.test.tsx` / handles hover and focus on Role and Model Resource Consumption header tooltips with placement-left |

## Exact changed surviving recipients

This is the complete body-change list. All are named recipients in the frozen ledger. Navbar's L30 workspace recipient is already sufficient and unchanged, so it is intentionally absent.

| Current source | Existing declaration |
|---|---|
| `backend/internal/domain/aggregations_test.go:566` | TestEnumsParsing |
| `backend/internal/domain/enums_test.go:138` | TestOrderSubagentRoles_OrderingAndPreservation |
| `web/src/components/AnomalyBadge.test.tsx:39` | invokes onClick callback when clicked and changes role to button |
| `web/src/components/AnomalyBadge.test.tsx:22` | renders CRITICAL severity with red styling classes and accessible role status |
| `web/src/components/AnomalyBadge.test.tsx:6` | renders WARNING severity with amber styling classes and accessible role status |
| `web/src/components/Components.test.tsx:595` | clicking column header InfoTooltip triggers stops propagation and does not toggle thread details |
| `web/src/components/Components.test.tsx:300` | handles hover and keyboard focus states on TrendChart InfoTooltip trigger |
| `web/src/components/Components.test.tsx:159` | handles single item with 100% cost allocation cleanly |
| `web/src/components/Components.test.tsx:142` | handles zero total cost without NaN or division by zero |
| `web/src/components/Components.test.tsx:394` | renders healthy saturation status and 200k ceiling |
| `web/src/components/Components.test.tsx:179` | renders trigger button with accessible aria attributes and role="tooltip" |
| `web/src/components/Components.test.tsx:223` | toggles visibility class on focus, blur, and dismisses on Escape key |
| `web/src/components/MultiSelectDropdown.test.tsx:30` | opens dropdown on click and displays options |
| `web/src/components/Navbar.test.tsx:296` | invokes onSelectTab with landing when project badge is clicked |
| `web/src/components/Navbar.test.tsx:241` | renders complex semver tags with prerelease and build metadata in badge |
| `web/src/components/TicketTable.test.tsx:265` | clicking InfoTooltip trigger on sortable headers stops click propagation and preserves sort order |
| `web/src/components/TraceWaterfall.test.tsx:421` | invokes onSelectEvent with anomaly when explicit AnomalyBadge is clicked and stops propagation |
| `web/src/domain/anomalies.test.ts:35` | returns true for all four tool failure anomaly rule IDs |
| `web/src/utils/theme.test.ts:76` | handles special characters and complex model tags deterministically |
| `web/src/views/FleetOverview.test.tsx:104` | renders empty state when workspace list is empty but total_tickets is 0 |
| `web/src/views/MetricProgression.test.tsx:20` | renders loading state initially and then displays progression KPIs and trend charts |
| `web/src/views/TicketDeepDive.test.tsx:794` | renders information tooltips on velocity, table headers, and refined KPI tooltips |
| `web/src/views/TicketDeepDive.test.tsx:977` | wraps Role and Model Resource Consumption tables in .table-responsive containers and keeps headers outside |

## Verification and limits

| Owner | Command/evidence | Result |
|---|---|---|
| Tester | Independent baseline/current Go and TS AST parses; multiset comparison; helper/parameterized/retained-body comparisons | PASS: exact 42 authorized removals; 23 authorized receiver body changes; no additions or unrelated changes. |
| Tester | `(cd web && npm run typecheck && npm run lint)` | PASS, exit 0 for strict tsc and eslint with max-warnings 0. |
| Tester | `(cd web && npm test -- --reporter=json --outputFile=/tmp/rev4-test05-vitest.json)` | PASS, exit 0. 34 actual file reports, all passed; 712 expanded test-case records, all passed; zero failed/pending/todo. Aggregate nested-suite counters: 148 passed, zero failed/pending. |
| Tester | `git diff --check` | PASS. |
| Coder B (reported focused evidence) | `npm --prefix web test --` followed by the nine owned frontend paths in PLAN.md | PASS: 9 files, 254 cases, zero failed/skipped; 15.80 seconds. Existing React act warnings reported in unchanged MetricProgression preset test. Full JSON frontend run independently passed subsequently. |

Full frontend JSON output and audit scripts/inventories remain outside Git under `/tmp/rev4-test05-*`. The JSON run log contains no warnings; this does not erase the coder's recorded focused-run warnings. Baseline expanded count 749 is prior iteration evidence; this tester observed the current 712 directly.

The parent owns the required full integrated `make`, full backend race suite, frontend asset build/embedded binary verification, final catalog/doc integration, telemetry, commits and PR delivery. Those gates are not claimed completed by this report. No standalone browser geometry, mutation testing, benchmark speedup, real-network integration or release/publication exercise was performed.
