# PO-19: Reconcile the active roadmap with development

## Request and boundaries

User request: organize TICKETS.md, remove To Do and move unblocked work to Ready, and review every existing ticket against current development. Baseline: clean checkout at 03f5b05d; fetched origin/main is the same revision. PO-19 was unused in the workspace/index search and is assigned to this documentation task. No application, tests, runtime scripts or developer contracts are changed. Preserve existing ticket IDs, requirements, historical completed/discarded files and the portable ledger contract.

## Evidence and method

Read all 23 active specifications, the shared registry, current source contracts and CLI dispatch/flags. Compare merged history (#109 AL-62, #110 GEN-21, #112 AL-67, #113 REV-6, #114 AL-77, #115 AL-66), completed AL-50/AL-57/AL-59/AL-76 specifications, AL-67 follow-up evidence and open GitHub PRs. Source anchors: backend/cmd/agentic-lens/main.go; backend/internal/domain/cost.go, live.go and events.go; backend/internal/adapters/live/sqlite.go, backfill/recovery.go, http/server.go; web/src/utils/debugPrompt.ts, components/AnomalyDrawer.tsx and views. Negative symbol/route searches are evidence of missing scoped surfaces, not proof that every acceptance case is broken. Original audit findings retain their frozen evidence and require developer revalidation.

Open delivery: AL-64 draft PR #121 at c68fbc9079d2f24c624d5de0fe87d313ed095b19 reported successful checks and DIRTY merge state at review time. Preserve its owner's scope; do not close the ticket or edit that PR. Release PR #55 is human controlled and outside this work.

## Ticket-by-ticket disposition

- **AL-63 (Ready)**: AL-62, AL-58 and AL-59 are merged foundations. Retain the unperformed cross-source capability matrix; distinguish historical hydration from live support. Coordinate FIX-18 and REV-7 without making them prerequisites.

- **AL-64 (In Progress)**: In Progress: draft PR #121 is open at c68fbc9079d2f24c624d5de0fe87d313ed095b19. Checks reported success, but GitHub reports merge conflicts. The fix is not merged; preserve every source-family acceptance case and let the existing owner finish delivery.

- **AL-65 (Blocked)**: Still blocked by AL-63. Current backfill help already describes --recompute-db without --force; research changing the default, not adding this existing option. Preserve local-source precedence and unavailable-owner history.

- **AL-69 (Ready)**: Retain the scoped corrections. Current domain/live.go still inspects crash phrases, and AL-67 F05 independently supports the child-lookup diagnostic gap. AL-62 lean ownership and AL-66 optional output are shipped baselines to preserve, not work to repeat.

- **AL-70 (Ready)**: Retain configuration/path/identity work. AL-67 F01/F02/F11 supplies recent composition evidence; http/server.go still binds :port. AL-59 source selection exists, so correct propagation rather than recreating selection.

- **AL-71 (Ready)**: Retain request ownership, date boundaries and live validation/recovery. Current TicketComparisons, MetricProgression and live hooks remain the consumers. AL-60 exclusion search is shipped and must survive these corrections.

- **AL-72 (Ready)**: Retain keyboard, clipboard, chart and accessibility corrections. Existing prompt copying and chart views are present; this work corrects their behavior rather than introducing them.

- **AL-73 (Ready)**: Retain installer/checksum failure acceptance. FIX-16 and GEN-19 repaired release-candidate workflow behavior; that delivery does not establish installer artifact validation or complete checksum-manifest rejection.

- **AL-74 (Ready)**: Retain remaining migration, oracle, isolation and unused-code corrections. REV-6 (#113) already split test lanes and gates integration on unit success; do not repeat that orchestration. Current domain/events.go retains anomalyEvidenceSelection.events; AL-67 F10 reconfirms the required-column oracle issue.

- **AL-75 (Ready)**: Retain evidence-led dependency and reproducibility investigation. REV-6 changed build/test ordering, not proof of lockfile integrity or dependency reachability. Historical audit counts are not current vulnerability evidence.

- **AL2-1 (Ready)**: Partially implemented: AL-50 and current web/src/utils/debugPrompt.ts already provide browser prompt generation and copying. Remaining scope is CLI/API availability, persisted-fact fallback and a consistent preview/entry flow; preserve verbatim raw payloads and reuse existing controls.

- **AL2-2 (Ready)**: Not delivered: current CLI has no archive subcommand and scanner has no is_archived contract. AL-76 compact JSON/format is shipped but is not workspace archiving. Preserve configured workspace roots and portable history.

- **AL2-3 (Ready)**: Not delivered: project-name configuration exists, but no ProjectDefinition/ScanMultiProjectWorkspaces contract was found. Single-project behavior remains the baseline; this ticket unlocks AL2-4 and AL2-5.

- **AL2-4 (Blocked)**: Still blocked by AL2-3. No /api/projects registry/sync or GitCacheManager implementation was found. Existing Git code-metric analysis is not remote project ingestion.

- **AL2-5 (Blocked)**: Still blocked by AL2-3 and AL2-4. FleetOverview exists for one project; no project registry client or ProjectSelector was found. Retain project identity across API selection and views.

- **AL2-6 (Ready)**: Not delivered: AL-57 and domain/cost.go already distinguish observed zero, partial and unavailable cost. No model_pricing rate-card engine was found. Add estimates alongside this evidence; observed zero must remain observed and must not trigger estimation.

- **AL2-7 (Blocked)**: Still blocked by AL2-6. Current cost views expose observed cost/coverage, not CostDisplayMode or rate-card simulation. Simulation must not rewrite measured evidence or shared ledgers.

- **AL2-8 (Ready)**: Not delivered: reasoning token metrics exist, but no configured reasoning_effort/variant fact or display contract was found. Historical OpenCode sample counts are original observations, not a current schema guarantee; validate v1/v2 independently and preserve native unknown values.

- **AL2-9 (Ready)**: Ready, discovery-only. No signing implementation was found or is authorized. Portable ledger reads remain supported without local receipts, private keys or mandatory authenticity checks; any proposed policy change needs an explicit product decision.

- **FIX-18 (Ready)**: Still open: AL-67 F03 distinguishes v2 historical support from v1 live queries. Current live/sqlite.go still queries session/message/part; no merged corrective delivery was found.

- **FIX-19 (Ready)**: Still open: AL-67 F04 documents duplicate active/archive owner snapshots through generation/backfill versus first-readable hydration. No merged corrective delivery was found; retain exact-owner ambiguity semantics.

- **FIX-20 (Ready)**: Still open: live/sqlite.go uses non-context SQL reads and backfill/recovery.go calls CaptureSessionSnapshot with context.Background(). Retain cancellation and stored-fact preservation acceptance.

- **REV-7 (Ready)**: Ready, review-only. AL-67 delivered the broad architecture assessment; this narrower follow-up owns remaining mixed-family, snapshot, ordering and resource guarantees. It does not block concrete fixes.

## Acceptance and verification limits

The registry has no To Do or Discarded section, each of 23 active specifications appears exactly once, and status/dependency metadata agrees between registry and specification. AL-64 is In Progress; 18 tickets are Ready and four remain Blocked. Dependency edges are reciprocal, acyclic and reference active tickets. No ticket is closed based on partial implementation. Existing gaps in prior numbering are preserved. Historical done/discarded artifacts are not rewritten.

AL2-1 retains its CLI/API delivery while acknowledging shipped browser prompts; AL2-6 respects observed zero/partial/unavailable costs from AL-57; AL2-8 preserves native effort values and validates source-family assumptions rather than treating old sample counts as current evidence. These are compatibility corrections within existing ticket goals. All other acceptance criteria remain intact; review notes distinguish delivered foundations from remaining scope. No new feature policy, signing requirement or development authorization is introduced.

This is author reconciliation, not an independent correctness audit or runtime confirmation of historical findings. No subagents launched. Full make is required for delivery; document checks additionally cover file links, unique IDs, exact status metadata and dependency graph consistency. Source evidence remains separate from synthetic test results.
