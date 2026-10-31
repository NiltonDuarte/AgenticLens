# AL-52 Technical Contract

## Approval and baseline

The user explicitly pre-approved the plan and contract on 2026-10-31: "plan and contract is pre-approved. full send it". This contract grounds the approved AL-52 specification without expanding its acceptance boundaries. Initial checkout: AL-52/rich-timeline-badges-subagent-session-linking; baseline commit: 85a2af94b5036b0918aaf9a7a89aed3a990f2650. Initial Git status was clean. Workspace root is the configured alspaces directory.

## Problem and delivered behavior

Skill and task tool invocations currently appear as generic tools in the retrospective TraceWaterfall and live LiveTraceInspector. Render prominent skill-name and task-role badges using existing AL-51 facts. A recorded child session identifier permits navigation to that child's existing view when available; unavailable children retain an informative full-identifier tooltip and copy action. Long badge labels truncate without hiding the complete value from inspection.

## Acceptance boundaries

1. Both trace surfaces display distinct skill and task badges when supported metadata is present, including the labels [skill: name] and [task -> subagent_type].
2. Retrospective badges consume ledger.skills and ledger.task_links supplied by TicketDeepDive, with safe event-input fallback where available. Live badges consume LivePartFact.skill_name, subagent_type and child_session_id.
3. Metadata must belong to its invocation. Ledger fact collections lack a parent session identifier; match only uniquely attributable compatible tool occurrences using timestamp and turn (and skill role where useful). Never attach another session's child through a first-match turn heuristic. Ambiguous facts preserve ordinary trace behavior.
4. A known child navigates within the current LiveObservability session tree or the current TicketDeepDive session/thread view. Missing or pending children expose their complete recorded ID and a copy button without manufacturing a session or guessing a link.
5. Existing supported v4 ledgers with absent optional metadata remain usable. Hydration failure cannot hide recorded badges or existing events. Generic tools, anomaly interactions, filters, collapse defaults and diagnostic source copy retain their behavior.
6. Badges conform to existing theme conventions, use nowrap and ellipsis, retain accessible full labels/tooltips, and support keyboard activation. Nested navigation/copy controls must not activate the parent timeline event.

## Architecture and exclusions

Scope is TypeScript presentation, pure presentation metadata helpers, current view wiring, CSS and relevant tests. No Go domain, extraction, schema, API or migration changes are required. Existing ledger and live contracts remain unchanged. Zero any and zero emojis apply. Network work stays in existing view boundaries; badge components do not fetch. No new polling or child discovery is introduced. Generated frontend assets remain ignored and are rebuilt into the Go embed tree through the canonical build.

## Verification and delivery

Extend coherent component/view coverage for metadata rendering, known child selection, unavailable child copy, optional fields and ambiguous ownership. Use isolated fixtures. Run affected Vitest tests, frontend typecheck/lint/build and the integrated make gate after upstream integration. Inspect web/dist and backend/internal/adapters/web/dist, exercise the existing embedded asset guard and standalone binary. Record actual evidence and limitations in iteration review and SUMMARY.md. Parent orchestrator owns lifecycle, telemetry, commits and PR delivery.
