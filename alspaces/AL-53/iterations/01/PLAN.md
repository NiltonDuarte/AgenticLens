# AL-53 Iteration 01 Plan

## Approval and scope

User source: the initial AL-53 request explicitly pre-approved the plan and contract and instructed full delivery. The contract records that approved presentation-only scope. No implementation was present initially and no prior ticket-local planning artifacts existed. Baseline is `85a2af94b5036b0918aaf9a7a89aed3a990f2650`, with clean initial status and an already-up-to-date upstream merge. This iteration introduces no material steering delta.

## Grounded data path

- `backend/internal/domain/models.go` defines persisted `SkillFact` and `TaskLinkFact`; `backend/internal/adapters/opencode/skills_tasks.go` extracts them from OpenCode tool calls. Facts include caller role, turn, timestamp, boolean skill success, optional skill error, optional verified child ID, and optional duration/cost. The current extractor does not produce delegation cost; unavailable is expected and must remain visible as unavailable.
- `backend/internal/adapters/schemas.go` and `repository.go` already preserve these collections through ledger DTO/read paths. They require no modification for this slice.
- `web/src/domain/types.ts` mirrors those fields as `AgenticFactLedger.skills` and `.task_links`, and already supplies `ledgerSkills` and `ledgerTaskLinks`. `web/src/domain/events.test.ts` covers defensive accessor behavior; avoid duplicating literal accessor tests.
- `TicketDeepDive.loadData` in `web/src/views/TicketDeepDive.tsx` shows persisted workspace detail before optional trace enrichment. Render the panel from `detail.ledger.facts`, not the hydrated event array, so pending/failed/empty enrichment cannot erase it.
- `web/src/components/ThreadViewer.tsx` renders `ThreadExecutionRecord` accordions, initially opening the first thread, but currently has no external child-selection or focus API. Add only the minimal integration needed for exact-ID jumps. Retain ordinary defaults and user toggles.
- `web/src/domain/formatters.ts` already supplies currency, durations, and timestamp labels/tooltips. Reuse them with presence guards; add pure boolean status and raw-role presentation helpers. `web/src/index.css` supplies existing thread-card and responsive-table conventions. `TraceWaterfall` provides a neighboring collapsible-card pattern.

## Implementation sequence and ownership

1. Coder owns the necessary production frontend paths: a dedicated panel component if useful, TicketDeepDive integration, pure formatters, minimal ThreadViewer navigation integration, and relevant styles. Respect the frozen contract, maintain independent skill/delegation rendering and row multiplicity, and avoid schema/backend edits. Use explicit unavailable text for absent optional metrics and preserve source text.
2. Tester owns relevant test additions/adjustments after implementation. Inspect the existing view, thread, and formatter cases before adding coverage. Extend coherent suites, using local v4 fixture variants and real DOM interactions rather than live `alspaces` artifacts. Execute focused frontend checks and communicate exact results to the parent.
3. Reviewer owns `iterations/01/REVIEW.md` and reads the production/test diff against this plan and contract. Audit spec alignment, unknown-value handling, exact child identity, accessibility, source preservation, and architectural rules. Record reviewed diff identity, reused valid checks, any unverified geometry boundary, and actionable findings. Coder/tester resolve review corrections within iteration 01.
4. Parent owns ticket index/status, summary and PR artifacts, implementation commit, merge from origin/main, final integrated `make`, lifecycle updates, final Codex telemetry extraction/validation, final commit, push, and PR registration. Stage only task-owned files and preserve unrelated changes in the shared checkout.

## Focused regression cases

- Optional facts: Extend `TicketDeepDive.test.tsx` to show the exact empty message with omitted and empty collections on supported v4 detail; rendering must not crash or log React errors. The plausible regression is unconditional collection access.
- Source fidelity: Exercise skill success and failure, caller roles and invocation coordinates, error details, duplicate invocations, and delegations-only input. The plausible regressions are skill-gated delegation rendering, accidental deduplication, status inferred from text, and lost diagnostics.
- Unknown metrics and navigation: Verify absent cost/duration use unavailable labels, explicit measured zero remains zero, and exact matched child links expand/focus the correct thread. Verify missing IDs and unavailable thread targets preserve truthful state. Existing ThreadViewer tests own initial expansion and toggling; extend them only for new selection behavior.
- Accessibility and containment: Use keyboard interactions for collapsing/reopening the panel and following its child action, checking semantic headings/table labels, expanded state, focus, and full long-value access. Preserve surrounding default collapse behavior. DOM assertions prove behavior and semantics; do not claim browser geometry from CSS inspection.
- Optional hydration: Extend an existing baseline/enrichment lifecycle test or add one coherent panel scenario proving persisted skill/delegation facts remain visible while enrichment is unavailable. Do not duplicate the entire pre-existing enrichment matrix.
- Shipping boundary: Parent runs `make` after integration, inspects matching frontend/embedded asset trees, and verifies the host standalone binary serves the freshly built panel assets. Generated assets are ignored and must not be force-added.

## Verification record

Planning only: grounded source paths, existing tests, configuration, Makefile, and current frontend scripts. No application changes or tests were executed by the planner. Full test results belong in REVIEW.md and SUMMARY.md after implementation. No live ticket workspaces or source databases are test fixtures.

## Routine integration correction: MGT-1 F1 and F5 (2026-10-31)

Source: user-authorized management instruction to integrate actual main after PRs 76, 79 and 75 merged. This is a review correction within iteration 01, not new feature scope. Base is origin/main `6bc1064943ae22e1b6586a6d145543c6a9cd23f2`; starting AL-53 head `57058521c498b9e376bbab5ec2355adec63d7729`, clean status. Preserve the frozen product guarantees and all existing meaningful assertions. Consolidate competing focusSession/jumpRequest contracts into one explicit state/prop used by both waterfall and panel navigation; retain exact child focus/expansion/scroll, repeat activation, keyboard access, full IDs, ticket reset and no stale parent trace. Preserve upstream live refresh/reading anchors and fleet skills features. Supersedes the earlier standalone navigation API only; no speculative AL-43 integration or AL-55 implementation. Parent reconciles planning artifacts, retaining completed-ticket removals and marking AL-55 Ready with no blockers. Coder owns production conflict resolution; tester owns missing combined regressions; reviewer independently audits combined behavior. Required full make and embedded/standalone checks run on the actual merged branch. Own Codex telemetry must preserve existing roots/children. Manager alone gates and merges; this owner only prepares, pushes and reports exact readiness.

## Verified startup-fix integration (2026-10-31)

Source: MGT-1 explicit unblock after PR 82 verified merged at `c3ce116611638ea1bee47dc3140d8af992814a44`. Clean starting head `dff7f7dbab8bad3991d2ce4cbde08d24e736835f`; actual fetched main matched the instructed base. Clean merge `8a853069d1f037ab6935988f7f5bfbd64fbd3910` adds upstream AL-47 startup-ordering correction/regression and AL-47 artifacts only. No AL-53 frontend/test or planning dependency change; existing independent navigation review remains valid. Go source and tests match main exactly. Original failed gate and concrete proof remain historical evidence. Run fresh full make plus embedded/standalone verification on this merge, refresh only own telemetry preserving session identities, and report to manager. No merge/auto-merge, AL-43 speculation or release action. Acceptance scope unchanged; same iteration.
