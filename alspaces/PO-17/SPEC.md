# PO-17: Cost semantics and schema priorities

Status: AL-55 bounded scope approved and reconciled in iteration 02; cost design remains a review proposal only. User selected AL-only development, AL2 queued, and deferred all cost changes pending PO review. No cost design is approved. Only the existing AL-55 shared specification and PO-17 docs are released for this PO delivery.
Baseline: `bdef2d0ef98e37fe259921bc30ba2daefbfda717`.

## Problem statement

Operators see numeric zero for cost that source traces never measured. Numeric totals, rankings, charts, and summaries then imply free usage. AL-57 correctly identifies that defect but prescribes nullable values before resolving partial totals and provenance. AL2-6 would estimate on zero, including potentially genuine zero. AL-55's original acceptance criteria recreate already-shipped schema work.

## Personas and user stories

- As an engineer, I want unavailable cost distinguished from observed zero so I can trust a session report.
- As a team lead, I want the known cost subtotal and its coverage so mixed-source work remains useful without claiming a complete financial total.
- As an operator, I want historical facts preserved with explicit uncertainty so upgrades do not rewrite evidence.
- As an analyst, I want API-equivalent estimates separated from source-reported costs so subscription usage is not represented as actual invoice spend.

## Recommended product semantics, pending acceptance

Cost has two independent dimensions: availability/evidence and valuation basis. Observed source-reported USD, unavailable source-reported USD, unverified historical USD, API-equivalent estimate, and simulation are distinct concepts. A source-reported amount is not proof of settled invoice billing. True observed zero is known. Missing data is never free usage. A harness label alone is insufficient to classify an arbitrary imported historical amount.

Use explicit states/provenance at fact boundaries and coherent aggregate projections. Null may represent an unavailable complete amount at a wire boundary; pervasive nullable arithmetic is not itself the product requirement. Developer planning selects representation under FCIS and strict typing. Do not introduce rate cards, subscription allocation, or estimates in AL-57.

A cost aggregate reports a known source-reported subtotal, the number of cost-bearing invocation facts included, unknown and unverified counts, and whether its complete total is available. Coverage denominator is cost-bearing invocation facts in the selected dataset, not sessions, tool calls, or dollars. Split-model invocation facts count separately; retain session-owned identity and multiplicity. Invocation coverage is not financial coverage. Empty datasets have an explicit empty state rather than a claim of complete measured free execution. Filtered scopes recompute coverage using the same included facts as their subtotals; excluded invalid ledgers remain separately diagnosed.

For mixed facts, show the known subtotal with a partial label and coverage, while the complete observed total stays unavailable. Unverified historical numeric facts remain visible as legacy values but cannot establish a complete observed subtotal. Do not count repeated role/model/tool projections as independent costs. Derived averages, ratios, shares, rankings, and efficiencies inherit availability and valuation basis; a partial subtotal cannot serve as a complete cost denominator.

Illustrative acceptance examples (synthetic, not session evidence):

- Observed 2 USD + unavailable: known subtotal 2 USD; 1 of 2 facts known; complete total unavailable.
- Observed 0 USD + unavailable: known subtotal 0 USD, marked partial; 1 of 2 known; complete total unavailable.
- Observed 0 USD + observed 2 USD: complete source-reported total 2 USD, 2 of 2 known.
- All unavailable: cost unavailable, zero known facts; do not show an unqualified 0 USD headline.
- Legacy numeric 0 without reliable provenance: retain original 0 and mark unverified; do not silently classify as measured zero or replace with estimate.
- Observed 2 USD + separate 3 USD estimate for a disjoint unknown fact: observed subtotal 2 USD; API-equivalent estimate 3 USD; modeled combined value 5 USD only under an explicitly named modeled view. It is not billed spend. If all-fact API-equivalent mode is later selected, price every eligible fact on that basis rather than adding its estimate to its observed amount.

## Presentation and compatibility

Go CLI text/Markdown, JSON/API, and TypeScript ticket/fleet/model/role/comparison/progression views communicate the same availability and valuation basis. Known zero renders as zero; unknown renders as unavailable; historical unverified values identify their status. Programmatic consumers need a documented wire evolution and actionable upgrade behavior rather than silent changes to numeric meaning. Human summary formatting must never coerce missing values to zero.

Recommended sort semantics: complete known values sort numerically, incomplete/unverified values remain after them in either direction with stable identity tie-breaking. Do not rank partial subtotals as full costs. Recommended charts: unavailable points are gaps/omissions with excluded-count explanations, never zero points or fabricated connecting lines. Cumulative known subtotal may continue only as explicitly partial after a missing bucket; a complete cumulative total cannot recover by treating earlier unknowns as zero. Keep non-cost axes and unrelated metrics usable. Log charts preserve measured zero as known even where a log axis cannot plot it.

## Migration and source safety

Preserve original historic numeric facts, known nonzero values, and observed zero. Reclassify a proven placeholder only through explicit migration/backfill with reliable record-level source/provenance evidence, a diagnostic and recoverable original. Source-less and ambiguous history remains unverified. Never infer a price from model name, numeric zero, subscription membership, or absent source records. Preserve other-device/harness sessions, Git facts, invocation multiplicity, and ownership.

Retain AL-56's ordinary runtime ingestion boundary: explicit migration and source-backed backfill are separate operations; shape-only migration is not recovered evidence. Backup, dry-run, atomic writes, idempotency, ticket scoping, partial diagnostics, and unavailable-source behavior remain acceptance requirements. Do not add client-side legacy ingestion. Do not prescribe v4 again; coordinate the next schema transition between owners using the current migration framework and expand-contract boundaries. AL-57 must provide its own safe compatibility path without requiring AL-55 solely for numbering.

## Priority and existing ticket boundaries

1. Product acceptance: decide cost reporting direction and authorized planning breadth. AL-57 implementation remains on hold.
2. AL-57: deliver source evidence/availability, aggregates and coverage, migration compatibility, CLI/API and existing UI surfaces as one coherent correctness slice, with existing numerical values preserved. Keep the existing ID. If execution needs batches, expand fact/contract compatibility, then migrate all consumers, then contract legacy behavior; do not ship a partially misleading production path.
3. AL2-6, if selected: optional deterministic rate-card calculation and basic CLI/detail presentation with estimate provenance. Depends on accepted and delivered AL-57 semantics. Never use zero alone as the estimate trigger; observed zero remains intact. Missing/unmatched models and unsupported token categories stay explicitly unpriced. Catalog identity/version and applied rates must make estimates reproducible. No current rate research required for this PO task.
4. AL2-7, if selected: modes and fleet/comparison/progression projections followed by simulation under the same ID. Depends on AL2-6. Distinguish source-reported, API-equivalent, and modeled combined views with coverage. Simulations change projections only, never stored facts or configuration. Accept the exact hybrid definition before development; an estimate for every fact and an estimate filling only missing facts are different measures.
5. AL-55: independently ground the remaining skills/task source-recovery value in its dedicated developer thread. Remove duplicate v4 and client migration requirements from the active proposal after approval. Keep optional metadata unless a concrete consumer need justifies a new required-array contract. No dependency on AL-57 by default; serialize schema-sensitive integration if both plans change it.

Proposed adjacency: AL-57 blocks AL2-6; AL2-6 blocks AL2-7. AL-55 has no cost dependency unless developer grounding establishes a necessary shared capability. These edges are proposals; the shared index is unchanged.

## Alternatives and decisions for manager coordination

Alternative A (recommended): source-reported known subtotal plus explicit unknown/unverified coverage; complete total unavailable when incomplete. Useful mixed-source reports, with more aggregate metadata.

Alternative B: strict unavailable total whenever any fact is unavailable/unverified; omit partial subtotal. Simpler headline semantics, less visibility into already-known spend; still requires provenance, zero preservation, sorting/chart and migration rules.

Alternative C: defer cost changes; hold AL-57/AL2-6/AL2-7 implementation and retain an explicit known reporting limitation. This leaves the current zero ambiguity unresolved and is not a correctness fix.

Manager owns the pending question and follow-up, avoiding conflicting prompts. Required decisions:

1. AL-only delivery/planning versus including AL2 cost tickets; broader analysis does not authorize AL2 dispatch.
2. Choose A, B, or C; no recommendation is accepted merely because it appears first.
3. Confirm ambiguous historical values remain preserved and unverified, with correction only on source evidence. Any contrary historical rewriting needs an explicit evidence policy.
4. After AL-55 grounding, choose the residual user outcome or retire superseded requirements; required arrays alone do not justify a migration project.
5. If AL2 proceeds, accept the semantics of each mode, especially the disjoint modeled combination. This can remain deferred while AL-57 proceeds.

## Non-functional and verification boundaries

Go domain remains pure; adapters own source/file effects; TypeScript and CLI own presentation. Preserve strict typing, zero any, no emojis, configured workspaces, privacy, and embedded single-binary distribution. Tests belong to separate developer sessions, use self-contained fixtures, and must cover the synthetic cases above through actual adapter, migration, API, aggregation, CLI, and UI boundaries. Existing relevant test owners are listed in iterations/01/REVIEW.md. Full make and embedded asset verification remain developer delivery gates.

PO-17 is documentation only. No new developer ticket, developer workspace, CONTRACT.md, implementation, or price catalog is authored. Shared planning files remain untouched until product acceptance and explicit manager release of those paths. PO final delivery later requires its own source-grounded Codex telemetry, docs commit classification, verification, and a documentation PR; no release action or PR merge is authorized.

## Authoritative steering and AL-55 reconciliation

User explicitly selected AL-only development and deferred all cost changes pending PO review. This selects a temporary execution hold, not acceptance of the proposed cost semantics. AL2 analysis identifies conflicts only; it authorizes no dispatch. The manager alone routes remaining questions. AL-55_ACCEPTANCE_PROPOSAL.md supersedes the preliminary residual-scope discussion above with five reviewable choices: optional v4 arrays, conservative snapshot retention policy, all-required-owner completion, explicit recompute/Force behavior, and existing backup scope. Choice 2 partial-refresh policy is accepted; choices 1, 3, 4 and 5 remain unfrozen pending the requested revision. The cost dependency graph is conditional future sequencing, never a blocker for conservative AL-55 recovery.

Partial acceptance update (explicit user question-tool result): choice 2 permits independently complete OpenCode session-tree refresh while retaining prior whole skill/task collections, with explicit CLI diagnostics and the mixed-age limitation. Other AL-55 choices are unfrozen: user selected "Revise before implementation" without identifying a change. Manager has requested freeform clarification; no revision or implementation approval is inferred. Costs remain deferred, AL2 queued, and shared-plan/developer-contract edits unauthorized. This decision record refines the current review iteration; it does not start a new implementation plan.

## Current AL-55 approval supersedes earlier holds

Explicit user approval to continue AL-55 implementation and direct confirmation of remaining choices supersede the earlier partial-acceptance/revision hold for AL-55. Wire-v4 skills/tasks remain optional; internal reads require arrays and normalize omitted fields to empty arrays without claiming recovery or forcing a file rewrite. All-needed-owner completion, baseline eligibility on failed optional refresh, recompute without Force/routine skips, and existing backup/atomic-write scope are accepted alongside choice 2. Iteration 02 and the released AL-55 spec are the current delivery record. Historical unresolved-choice text above records discovery only. No AL-57/AL2 cost design is approved.
