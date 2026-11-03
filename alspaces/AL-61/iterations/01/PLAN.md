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
