# AL-61 Iteration 01

## Scope and approval

Implement the frozen CONTRACT.md inventory. User explicitly pre-approved this ticket scope before planning; no renewed approval is required for these boundaries. Baseline 122e983f8c365a2aa248b1b6d492252555e922ae, initially clean. Parent owns integration, artifact lifecycle, telemetry and PR delivery. Shared checkout participants preserve others’ changes.

## Implementation owner: coder

1. Inspect localized formatters in web/src/domain/formatters.ts and existing localized timestamp tests in formatters.test.ts. Add a historical local date/time formatter with second precision and explicit missing fallback; preserve formatLocalTimeWithMs and formatLocalDateTimeTooltip semantics for existing consumers.
2. Apply the formatter to Started, Completed and Updated in web/src/views/TicketDeepDive.tsx, chronological event labels in web/src/components/TraceWaterfall.tsx, skill Loaded At in web/src/components/SkillsSubagentsPanel.tsx, and selected historical event Timestamp in web/src/components/AnomalyDrawer.tsx. Preserve valid machine-readable source instants and full tooltip context; retain existing invalid text compatibility and missing completion semantics.
3. Inspect timestamp-group, timeline-timestamp, timeline event header, skills-subagents-table and drawer value styles in web/src/index.css. Make only necessary scoped wrap/layout changes, preserving table containment, tooltip visibility and keyboard access. Do not change global time-only formatting or live inspector.
4. Confirm existing API timestamp validation/decoding in web/src/api/validation.ts and client.ts passes original values unchanged. No backend or wire changes are expected.

## Verification owner: tester

Inspect and extend existing localized timestamp cases in formatters.test.ts, malformed workspace timestamp and skills assertions in TicketDeepDive.test.tsx, localized trace case in TraceWaterfall.test.tsx, and event detail assertions in AnomalyDrawer.test.tsx. Use independent expected calendar/time values under controlled non-UTC timezone, including cross-day conversion and both DST offsets. Preserve timezone environment isolation; do not merely compare the formatter to itself. Assert no fractional seconds in visible values and preserve tooltip/source datetime. Exercise missing start/completion/update and malformed strings, without adding unrelated generic edge matrices. Keep existing live formatter tests as a guard against precision changes.

Run focused frontend tests during development, npm run typecheck and lint, then the parent runs integrated make after clean commit and upstream merge. Canonical npm run build copies assets into backend/internal/adapters/web/dist; inspect both trees and validate embedded standalone delivery. Record actual commands and limits, including browser geometry if unexercised.

## Selected regression checklist

UI layout: long visible dates wrap/read at narrow widths, tables remain scrollable and controls remain keyboard accessible. Source evidence: raw diagnostics and stored/API timestamps unchanged. Frontend embedding: build before Go tooling and inspect ignored generated asset trees without adding ignored output. Tests: self-contained fixtures, meaningful assertions in coherent existing recipients, no live workspace dependencies.

## Independent review roster

This is a bounded display-only feature with no persistence, asynchronous or domain behavior change. A narrow generic reviewer may cover behavior, date/data failure semantics, architecture and test effectiveness together: all risks converge on the formatter and four direct callers, with unchanged API/persistence inspected as dependencies. Separate full architecture/data specialists would not add a distinct changed boundary. Both configured profiles remain mandatory: gpt-6.1-sol/high and gpt-6-luna/medium. Parent must record this narrow rationale, freeze one revision/fingerprint, dispatch fresh independent contexts and assign unique report paths iterations/01/REVIEW_SOL.md and REVIEW_LUNA.md. Each reviewer owns only its report, inspects production/test files read-only, and returns hypotheses, evidence, limitations and verdict. Parent owns consolidated approval, findings attribution and compact retention; incomplete model coverage cannot be called approved.

## Completion

After corrections and focused review, commit task-owned implementation, merge origin/main, run required make, resolve any integration effects and consolidate review. Parent moves the existing specification to plan/done, removes AL-61 index entry, checks downstream dependencies, records SUMMARY.md/PULL_REQUEST.md, generates source-grounded Codex telemetry at the last step, commits final artifacts, pushes and opens/links the PR. Do not modify AL-55 telemetry.

## Review correction and delivery steering

Sol/high independent review found SOL-1 (severity 5): JavaScript normalized impossible calendar timestamps into fabricated dates. This is a correction within approved invalid-input acceptance, not a new scope. Coder added historical RFC3339 calendar validation consistently for visible text, tooltip and datetime; live formatting remains unchanged. Tester extended coherent existing cases, including leap-date/offset compatibility. Both configured profiles re-review the frozen correction.

User steering on 2026-11-03: MGT-2 supervises eligible non-release delivery and owns merging. AL-60 is planned first; publish AL-61 when ready, then integrate AL-60 once its actual merge is notified before the management merge gate. Preserve both features and shared embedded assets. This supersedes any assumption of independent merge ownership; scope and acceptance are unchanged. Integration requires affected verification and review impact assessment.

User steering after baseline failures: preserve make evidence and defer another full-suite retry until AL-60 handoff. No weakening capture tests or unrelated product changes. Full gate ended with three unchanged CLI capture failures; focused correction verification may continue. Notify MGT-2 directly via t3_thread_send on concrete blockers and final handoff.

User steering: draft publication is explicitly authorized now and supersedes the prior publication hold. This does not change product acceptance or establish merge readiness. AL-60 is sole owner of the approved shared capture repair. Predecessor integration and full-suite retry remain on hold until MGT-2 notifies its actual merge. Publish the existing reviewed source with current evidence and send PR URL, exact head and pending gates directly to MGT-2. Final telemetry and lifecycle completion remain deferred until final delivery.

## Predecessor integration checkpoint

User notified AL-60 PR #101 actually merged at 6244585dc5cc41ef012cd3e1fb3b964ccd910a95 and explicitly requested the pending gates. Integrated current main exactly once into c40e002, producing 3c0e4cd59979309a14ad785882d609b62d40ad03 without conflicts. AL-60 comparison search, capture repair and completed index/specification changes are upstream content. Historical application/test files remain unchanged except additive independent CSS rules. Prior full-suite hold is superseded by actual-merge notification; run make on this merge, regenerate embedded assets, and verify both features in standalone delivery. Both configured narrow review profiles independently inspect integration and justify prior approval carry-forward. Parent owns INTEGRATION_SOL.md / INTEGRATION_LUNA.md report assignments, consolidated gate and final telemetry. PR #104 remains draft until checks and review pass; MGT-2 owns merging. This is routine integration within accepted scope, not a new product iteration.

Completion: integrated make passed at 3c0e4cd (790 frontend tests plus backend race/static/script checks and five builds); standalone assets match both canonical trees. Both integration profiles approved, no new findings. Final consolidated approval issued and findings compacted; final workspace cleanup/telemetry does not change reviewed application/test source.

## AL-60 follow-up integration

User notified verified merge of AL-60 PR #105 at 3fbecd2576270772462fe37235d90c5f33e2164c and requested integration plus required make without restarting the feature workflow. Initial status clean at 64d1e3d5ad36b1db1e764c48ddc57298f60a4d3f. Merged current main once without conflicts, producing 8c5e3568a53ce357d4d6870ddaec9289ed089985. Only four upstream Go test files and AL-60 artifacts changed: worker completion teardown barriers and deterministic delayed-start/early-return regressions. Production, CI, frontend source/tests/assets inputs and AL-61 historical behavior remain byte-identical to the approved prior head. Upstream backend/CI are byte-identical to origin/main. Preserve both configured profiles' existing approvals under this unchanged-source rationale; parent assesses the test-only merge and runs full make plus canonical asset parity. No new feature scope, duplicate capture repair or fresh feature review is required. AL-60 remains owner of its follow-up tests/debt; AL-61 owns integration evidence, refreshed telemetry and PR #104 handoff. MGT-2 owns merging after refreshed hosted checks.

Follow-up completion: required make passed at 8c5e3568 (790 frontend tests, CLI race96.194s, all Go/static/script checks, five builds). Standalone HTTP assets match both canonical trees and prior hashes. Parent integration assessment carries forward both configured unchanged-source approvals; no new candidate. Refresh existing 10-session Codex telemetry last, commit only AL-61 docs/telemetry, push PR #104 and send exact final head plus local readiness/hosted-check state directly to MGT-2.
