# REV-5 iteration 03 frontend audit

Status: CHANGES_REQUESTED. The assigned audit is complete; the findings below remain open implementation work. No application, test, dependency, or runtime files were changed.

## Scope and evidence

- Baseline: `d5227812b09a17ff15d7e6ffe058e96ecf1616eb`. All 101 assigned paths match their frozen `SCOPE.json` SHA-256 hashes after review.
- Read every source, test body, fixture, stylesheet declaration, and configuration file under the assigned frontend scope. The lockfile received structural traversal of every package entry, not a semantic review of third-party implementations. `FRONTEND_COVERAGE.json` records one disposition per assigned path.
- Applied the reviewer persona, review skill, technical guidelines, regression checklist, REV-5 contract, and iteration 03 plan. Applicable boundaries: configuration portability; persisted facts and optional hydration; classification and exported prompts; UI recovery and accessibility; polling and asynchronous updates; meaningful test oracles.
- Reused the parent's existing `make` evidence at `49cf350` (781 frontend tests), with unchanged frontend source verified by frozen hashes. This reviewer did not rerun the full suite, lint, build, embedded-binary checks, or browser checks. No browser or preview tools were used. These are reused or unperformed checks, not new passes.
- Read-only in-memory Node/esbuild experiments executed actual production prefix/filter/sort, anomaly matching, live decoding, clipboard, polling, and server-side chart rendering. No temporary test or production files were added. Synthetic inputs below prove deterministic behavior, not observations from real session telemetry.
- Confidence is high for the source-grounded defects below. Keyboard/event and CSS cascade findings are static execution-path analysis, not browser validation. This is not a claim that unreported files have no defects.

## Architecture assessment

The frontend has a useful existing split: domain modules compute immutable projections, API modules own transport, hooks/views own effects, and components render those values. Cost evidence uses explicit states and preserves unavailable values; the historical API decoder establishes a substantially stronger boundary than the live API. Event rebinding, ambiguous invocation rejection, and optional hydration have real domain/integration coverage. Those protections should survive any refactor.

The principal drift is inconsistent ownership of the same responsibilities: duplicated filter request state, server/client date calculations, live/retrospective attribution rules, clipboard outcomes, and state-versus-CSS visibility. FE-01 through FE-16 identify concrete consequences; FE-D1 through FE-D3 isolate subtraction and restructuring opportunities. File size alone did not establish a defect. No frontend domain network/filesystem side effects or explicit `any` implementation were identified in the reviewed sources; unchecked casts at live ingress still undermine runtime typing. The test fixtures and test setup are legitimate test dependencies, regardless of the repository's separate code-classification issue identified by the parent.

## Findings

### FE-01 — P2: configured prefixes containing digits are misclassified

`web/src/domain/formatters.ts:444` uses alphabetic-only prefix expressions; `:194` uses the same assumption for natural sorting. The repository config includes `AL2`. Actual production-function reproduction returned `extractTicketPrefix('AL2-1') === 'AL'`, filtering that ledger by `['AL2']` returned `[]`, and sorting `AL2-2, AL2-10` returned `AL2-10, AL2-2`. Comparisons (`TicketComparisons.tsx:126`) and progression (`MetricProgression.tsx:118`) locally discard valid server results for the configured AL2 option.

Use a canonical configured-prefix interpretation, preserving supported legacy ID formats without hardcoded repository prefixes. Extend existing formatter and view tests with AL and AL2 together, digit-bearing and delimiter-bearing custom prefixes, and natural numeric ordering. Current tests exercise alphabetical prefixes and normalization, but miss this configured production case.

### FE-02 — P2: obsolete filter requests can overwrite the current selection

`FleetOverview.tsx:89-105`, `TicketComparisons.tsx:79-89`, and `MetricProgression.tsx:74-85` commit every asynchronous result/error without an owning request generation or cleanup. Trigger: request A starts, the user selects B, B completes, then A settles. Fleet shows A's KPIs under B's controls; the other views can filter A's stale subset into an incorrect empty result. An obsolete rejection can also replace a newer success with an error.

Fence success, diagnostics, error, and loading transitions by the active request identity; cancel work where supported. Reuse the ownership approach already present in `useLiveTrace` and TicketDeepDive rather than adding a generic caching framework. Add deferred A/B success and rejection permutations to the existing view tests. Their current refetch tests use immediate settled responses and do not prove this ordering boundary.

### FE-03 — P2: applying a progression filter changes the meaning of the time window

`MetricProgression.tsx:74-77` fetches all history without type/prefix filters, but sends `preset` to the server when either is selected. Its local date calculation at `:134-167` anchors presets to the latest ledger date. The adapter reviewer confirmed `handlers.go:91-133` and `:173-188` use the actual current UTC time for server preset cutoffs. A historical dataset visible under 7d becomes empty when selecting a type that includes every displayed ticket; the server has already discarded it. The same inconsistency affects 1d, 30d, and 90d. `custom` is accepted by the server and is not an unsupported-window defect.

Choose one date-boundary owner. A minimal correction is to fetch all history for the selected types/prefixes and apply the existing latest-ledger window consistently. If current-time semantics are intended, explicitly change both paths together. Test with the clock far beyond the newest ledger and toggle only the type/prefix control. Existing mocked-client tests do not reproduce server filtering semantics.

### FE-04 — P2: live REST and SSE payloads cross the typed boundary unchecked

`liveApi.ts:23-80` casts nested arrays to live models; SSE `anomaly_alert` and `trace_update` at `:233-248` directly cast parsed JSON. Production reproduction accepted `{root_sessions:[{}]}`; the exact consumer traversal from `LiveObservability.tsx:46-47` then threw `Cannot read properties of undefined (reading 'flatMap')`. A malformed message can similarly reach `LiveTraceInspector.tsx:176` and call `toLowerCase` on an absent role. Primitive/null responses are silently converted to current-looking empty snapshots rather than reported as malformed.

Validate required nested node, message, part, and anomaly fields at REST/SSE ingress, with explicit normalization only for genuinely optional collections. Preserve the last valid snapshot on rejection. Share those decoders across transport paths. Extend `liveApi.test.ts` with malformed nested collections, null roots, wrong scalar types, and valid optional omissions; current invalid-JSON/default tests do not establish this boundary. The adapter reviewer confirmed ordinary Go leaf nodes initialize empty child slices; this finding concerns malformed transport data, not an asserted ordinary leaf-null production failure.

### FE-05 — P2: fallback polling remains in error after successful recovery

`liveApi.ts:155-179` emits `fallback-polling` only on entry, then emits `error` on failures. Successful polls reset the counter and deliver data but never restore the status. An in-memory production subscription with EventSource unavailable, two failed polls, then a successful poll produced statuses `[fallback-polling,error,error]`, one delivered tree, and final status `error`. LiveObservability consequently retains the connection error UI after data recovery.

Restore the polling status after a successful poll and test failure-threshold-to-success recovery. Also bound overlapping polls: `:185-189` schedules a fresh asynchronous fetch every two seconds regardless of an outstanding fetch. A slow response can overtake a newer one or accumulate work. Use one in-flight poll and schedule/coalesce the next; verify deferred completion ordering and teardown. Existing tests cover a transient failure and persistent failure separately, not recovery from the error state or concurrent pending polls.

### FE-06 — P2: parent keyboard handlers intercept nested controls

`TraceWaterfall.tsx:539-543` handles every bubbled Enter/Space event and prevents its default. Its nested native copy/payload buttons stop click propagation only. Keyboard activation therefore selects the row and can suppress the child's native click. `LiveSessionTree.tsx:54-58` has the same defect for the collapse button at `:118-124`. The child-session badge is explicitly protected and is not affected by this finding.

Limit row keyboard handling to events whose target is the row itself, or use separate native selection controls. Add `userEvent.keyboard` coverage for nested copy, payload, and collapse buttons, asserting both the child outcome and absence of parent selection. Existing row keydown and child mouse-click tests do not exercise this interaction; the child-session badge test proves only its dedicated propagation guard.

### FE-07 — P2: a selected event's payload cannot be hidden

`TraceWaterfall.tsx:523` defines expansion as `expandedKeys[itemKey] || isSelected`. Clicking a row selects it and opens the payload; clicking its displayed Hide Payload button (`:649-655`) toggles only `expandedKeys`, so the selected row stays expanded indefinitely. This affects both local and controlled selection. The existing payload hide test opens a payload without selecting its row, so it passes on a different state transition.

Separate initial expansion on selection from the user's explicit disclosure state. Preserve the established selection/enrichment reconciliation guarantees. Extend the existing payload test to select the row first, hide it, and rebind enriched facts while preserving the intended disclosure state.

### FE-08 — P2: clipboard failures are reported as successful copies

`utils/clipboard.ts:25-27` ignores the boolean returned by `execCommand`; production reproduction returned `true` when the command returned `false`. In addition, TraceWaterfall (`:118-119`, `:157-158`), AnomalyDrawer (`:49-50`), FailedToolAlertCard (`:55`), and LiveTraceInspector (`:71`, `:168`, `:347`) ignore the helper's false result and display Copied after a rejected clipboard write. ArtifactViewer duplicates the fallback at `:146` and has the same false-return problem, although its rejection catch correctly leaves Copy visible. TraceInvocationBadge correctly uses the helper result.

Honor the fallback boolean, remove temporary DOM nodes in `finally`, and reuse one result-aware copy interaction instead of duplicate implementations. Test denied writes, fallback false, fallback throws, retry success, and accessible failure feedback. Retain existing verbatim-payload and defensive-fence tests.

### FE-09 — P2: fallback anomaly matching accepts contradictory tool identity

`domain/anomalies.ts:113-130` falls through from a failed tool-name comparison to a matching turn/session. Production reproduction matched a bash retry-streak anomaly to a failed read call on the same turn/session. Parallel tools in one assistant turn can therefore receive the wrong badge, filter membership, inspector context, and exported diagnosis. Exact persisted evidence references are handled separately and should remain authoritative.

Require all supplied identity constraints to agree before applying weaker fallbacks. Add a same-turn, same-session, different-tool negative case alongside current successful/completed-status and session-isolation tests. The live matcher (`:207-221`) additionally uses substring identity/tool matching; constrain it with structured identifiers or explicitly parsed legacy trigger formats, and test `part-1` versus `part-10` and `read` versus `read_file` before changing compatibility behavior.

### FE-10 — P2: multiple positive bar series occupy identical chart geometry

`TrendChart.tsx:291-315` computes each series' bar x coordinate and width without a series offset or stacked baseline. MetricProgression's Token Volume Distribution (`:457-484`) supplies three positive bar series, so later series obscure earlier ones. Server-side rendering of two equal positive values produced two rectangles with identical `x=328`, `y=42.428571428571445`, `width=20`, and `height=153.57142857142856`. Transparency blends colors; it does not provide independently readable distribution bars.

Use explicit grouped bars or a defined stack with an appropriate scale. Add coordinate assertions for multiple positive series while preserving the existing diverging positive/negative baseline test and unknown-value gaps. Current tests cover one positive bar series and two bars on opposite sides of zero, not this shipped chart configuration.

### FE-11 — P2: fleet tool success is an unweighted mean of ticket percentages

`TicketComparisons.tsx:176-182` averages per-ticket percentages, but the KPI at `:258-261` claims overall tool-call success across the fleet. One failed call in one ticket and 99 successful calls in another produces 50%, although 99 of 100 calls succeeded. This is deterministic arithmetic evidence; no live telemetry is inferred.

Calculate the percentage from the selected ledgers' total calls and errors in a pure aggregation, preserving search/type/prefix/skill scope and defined zero-call behavior. Alternatively, explicitly label a product-approved ticket-average metric. Extend existing comparison tests with unequal denominators; checking the KPI label alone does not protect the calculation.

### FE-12 — P2: generic anomaly counts are exported as consecutive failures

`TraceWaterfall.tsx:133-141` passes any matching anomaly's `context.count` as `consecutiveErrors`. `domain/anomalies.ts:178-190` similarly extracts any count, and LiveTraceInspector (`:252-254`, `:440`) interprets it as a streak. A TOOL_HIGH_FAILURE_RATE count of three errors separated by successful calls is therefore exported as three consecutive failures. This is independent of FE-09: even correctly attributed anomalies can have different count semantics.

Restrict streak extraction to retry/consecutive-error rules; retain high-failure-rate totals as totals. Add a non-streak anomaly with count greater than one to existing prompt/inspector tests, alongside the existing positive retry-streak case. Keep original evidence text verbatim.

### FE-13 — P2: table actions lack keyboard equivalents

`TicketTable.tsx:99`, `:132`, `:145`, `:178`, and `:202` attach sorting/navigation only to non-focusable th/tr elements. The comparisons and landing headers use the same mouse-only sort pattern. A keyboard user cannot activate these sort operations; TicketTable also has no inner ticket navigation button. Comparisons and landing do provide inner navigation buttons, so that navigation-specific defect is confined to TicketTable.

Place native buttons in sortable headers and a native link/button in the ticket cell, with `aria-sort` on the header. Preserve InfoTooltip propagation behavior. Add keyboard tests to the existing table suites; their current fireEvent.click coverage does not establish keyboard access.

### FE-14 — P2: the modal anomaly drawer does not own keyboard focus

`AnomalyDrawer.tsx:27-42` registers Escape and `:123-124` declares an aria-modal dialog, but no opening focus, focus containment, background inertness, or return-focus behavior exists. Opening the drawer leaves focus on the underlying row; Tab can continue through obscured page controls instead of entering the inspector. The overlay protects pointer interaction only.

Use an existing/native dialog mechanism or implement focused dialog lifecycle behavior without creating a general modal framework. Test opening from keyboard, initial focus, Tab/Shift-Tab containment, Escape, and return to the invoking control. Existing tests verify role, content, and closing callbacks, not focus ownership.

### FE-15 — P3: tooltip CSS can override Escape dismissal

`InfoTooltip.tsx:25-28` removes the visible class and blurs focus on Escape, while `index.css:3658-3663` independently makes a hovered tooltip visible. With the pointer remaining over the trigger/container, Escape cannot dismiss the bubble; blurring also discards the keyboard position. The tests assert class removal without loading the stylesheet.

Give tooltip dismissal one state owner, preserve focus, and test visibility with the stylesheet or a focused browser check in a separately authorized implementation task.

### FE-16 — P3: Hide Reasoning does not hide short reasoning

`LiveTraceInspector.tsx:144-146` leaves short reasoning unchanged when showReasoning is false, and `:241` always renders it. Clicking Hide Reasoning changes only the label for short content; its test asserts that label change without checking content visibility. Make reasoning disclosure actually hide content, or use truthful expand/collapse preview labels. Extend the existing short-reasoning test with the intended content visibility assertion.

## Architecture and subtraction opportunities

### FE-D1 — P3: remove unused contracts and dependency surface before adding abstractions

`api/types.ts` has no consumers: AsyncState, ApiConfig, ApiError, and HealthBadgeVariant are speculative parallel contracts. `package.json:14` declares lucide-react, but no frontend source imports it. FleetAnomalyCard's `onViewTicket` prop (`:6`) is supplied by FleetOverview but ignored; FailedToolAlertCard's `messageId` prop (`:15`) is supplied by its caller but ignored. Repository source searches support these specific unused claims; they are not a general reachability proof.

Delete unused definitions/dependency/props if their removal remains consistent with the intended contract. If message identity is required in failure exports, implement the missing provenance deliberately rather than retaining a no-op prop. Verify typecheck, canonical frontend build, affected UI tests, and lockfile consistency. Do not invent consumers to justify existing abstractions.

### FE-D2 — P3: consolidate state ownership and move calculations out of views

FleetOverview, TicketComparisons, and MetricProgression duplicate filter request parameter construction, prefix discovery, independent loading/error/data state, and error handling. Their divergence is observable in FE-02 and FE-03. TicketComparisons contains business aggregation inside a view (FE-11), while formatters.ts combines display formatting, ID parsing, filtering, and row derivation. Prefer a small shared request lifecycle only where it removes these concrete duplicate invariants, and cohesive pure filter/date/benchmark functions. Avoid a broad framework rewrite or one-line forwarding wrappers. Keep useLiveTrace's existing coalescing and TicketDeepDive's enrichment identity logic intact.

### FE-D3 — P3: strengthen test oracles before reducing volume

The reviewed tests provide meaningful baseline-preservation, request-ordering, cost-unknown, defensive prompt fencing, and ambiguous-identity checks. Preserve those contracts. Narrow improvement candidates are the label-only reasoning assertion (FE-16) and class-only tooltip assertions (FE-15); successful-only clipboard assertions (FE-08); nested mouse clicks paired with parent-only keydown tests (FE-06); and repeated multiselect composition sequences in TraceWaterfall.test.tsx that overlap generic MultiSelectDropdown tests. Consolidation must retain composition-specific filtering, denominator invariance, empty-to-recovered transitions, and callback isolation. No automatic deletion count or measured speedup is claimed. No performance profiling was run because this audit found no measured slowdown requiring a new benchmark.

## Structural lockfile review and remaining limits

Parsed all 677 lockfile package entries (including root), checked manifest/root name and version, equality of direct dependency maps, installed-entry presence for each direct dependency, and resolution URL schemes. No manifest mismatch or non-HTTPS resolution URL was found. 287 non-root entries omit both resolved and integrity metadata; this limits an integrity-attestation claim and merits regeneration/inspection under the dependency-maintenance workflow. This audit did not fetch registry metadata, audit known vulnerabilities, attest package contents, or run a clean install. The lockfile is not marked semantically reviewed.

CSS was reviewed statically for cascade, focus, overflow, scroll containment, and responsive constraints; visual geometry and narrow-device layout were not browser-validated. The embedded Go asset path is preserved by the package build script, but independent embedded-binary validation remains the parent/integration owner's responsibility. The source files remain unchanged, so none of these findings is represented as fixed.

Final verdict: CHANGES_REQUESTED.
