# AL-53 Technical Contract

## Authority and baseline

The user explicitly approved the plan and contract scope on 2026-10-31: "plan and contract is pre-approved. full send it". This records the provided ticket scope against the current implementation; no new approval gate is required within these boundaries. The source specification is `alspaces/plan/to-do/AL-53 dedicated skills and subagents autopsy panel in ticket deep dive.md` (relocated to `plan/done/` at delivery).

Initial Git status was clean on branch `AL-53/dedicated-skills-subagents-autopsy-panel`, baseline `85a2af94b5036b0918aaf9a7a89aed3a990f2650`. Upstream synchronization preserved that baseline. Configuration resolves the workspace root to `alspaces`.

## Delivered behavior

Ticket Deep Dive gains a dedicated collapsible `Skills & Subagents` card consistent with existing cards. Its accessible heading and controls remain discoverable when collapsed. Opening it exposes persisted skill and task delegation facts independently; a ticket with only delegations must display those delegations. When both collections are absent or empty, show `No modular skills or subagent delegations recorded for this ticket.` Supported schema-v4 ledgers may omit these optional collections. This ticket does not broaden runtime support to schemas v1-v3.

The skills table displays skill name, recorded invocation turn and time, invoking role, Success or Failed status, and recorded error details when supplied. Status follows the existing boolean fact rather than parsing error text. Repeat invocations remain separate rows. Unknown caller roles remain identifiable rather than being remapped to another persona.

The delegation table displays recorded subagent type, task description, parent turn, duration, USD cost, and child session identity/navigation when available. Preserve long descriptions, errors, and identifiers for reading and inspection. Missing duration or cost remains explicitly unavailable rather than becoming measured zero. Recorded zero values, when present, remain distinct from absence. Do not infer delegation cost from unrelated aggregate metrics or manufacture parent-child relationships.

A delegation with a child ID matching an available Session Execution Thread offers a keyboard-operable jump that reveals the exact child's thread and brings its heading into view/focus. An absent child ID has no fabricated destination. If an explicitly recorded child ID has no available thread, retain the ID and explain that its thread is unavailable instead of presenting a broken jump. Existing default thread expansion and manual toggling remain intact.

## Layer boundaries

`web/src/domain/types.ts` already defines `SkillFact`, `TaskLinkFact`, and defensive `ledgerSkills` / `ledgerTaskLinks` accessors on `AgenticFactLedger`. Consume `ledger.facts.skills` and `ledger.facts.task_links`; the prose specification's shorthand `ledger.skills` is not a new top-level API contract.

`web/src/views/TicketDeepDive.tsx` owns data retrieval and integration. A presentation component may own panel rendering and local collapse state. `web/src/domain/formatters.ts` owns pure skill-status and role-label formatting. `web/src/components/ThreadViewer.tsx` may receive minimal selection/navigation integration for child jumps. Styling belongs in `web/src/index.css` using existing visual conventions and bounded table overflow.

No Go domain, adapter, CLI, persistence schema, extraction, configuration, polling, or hydration API changes are authorized. Existing persisted detail facts are the panel's baseline; optional `fetchWorkspaceEvents` enrichment must neither hide them nor replace these collections. Preserve current stale-request protection, trace inspection, artifact viewing, and ticket selection behavior.

Follow `TECHNICAL_GUIDELINES.md`: strict TypeScript, zero any, zero emojis, pure domain formatters, declarative presentation. Use native controls, semantic tables with identified headings, and keyboard-accessible row content and actions. Collapse controls expose state and association; actionable rows or row controls work without a mouse. Preserve full source evidence without destructive truncation.

## Acceptance and verification

Demonstrate the absent/empty supported-v4 state, success and failure skills with roles/turns/time/error details, delegations without skills, present and unavailable duration/cost, repeated rows, and exact child navigation including an unavailable target. Verify collapse and child actions by keyboard and preserve normal thread behavior.

Extend existing test owners coherently, particularly `web/src/views/TicketDeepDive.test.tsx`, `web/src/components/ThreadViewer.test.tsx`, and `web/src/domain/formatters.test.ts`; a dedicated panel test is appropriate for rendering and keyboard behavior with distinct setup. Fixtures must be self-contained and independent of live ticket workspaces or session stores. Optional enrichment failures must leave panel facts visible.

The integrated delivery gate is `make`, including installed frontend checks/tests, Go race tests, and canonical asset/binary builds. Verify `backend/internal/adapters/web/dist/` matches the fresh frontend build and the standalone binary serves embedded assets; generated ignored output remains untracked. Record test evidence and limitations in iteration review and summary. The parent owns integration, final telemetry, commits, lifecycle synchronization, push, and PR delivery.
