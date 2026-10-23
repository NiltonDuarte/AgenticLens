h3. Title: AL-54: Rich Timeline Badges and Subagent Session Linking in Trace Views

*TARGET*: AgenticLens (presentation / web / components / waterfall / inspector / tests)

*CONTEXT*
 
In both `TraceWaterfall.tsx` and `LiveTraceInspector.tsx`, tool calls currently render plain text labels (e.g. `skill` or `task`) without distinguishing what skill was loaded or what subagent role was spawned. Furthermore, when an operator inspects a parent session running a `task` tool call, there is no direct link to navigate to the spawned child session.

*GOAL*
 
Render distinctive contextual badges for `skill` (`[skill: <name>]`) and `task` (`[task -> <subagent_type>]`) calls across Trace Waterfall and Live Trace Inspector, providing a one-click jump button on task badges to navigate directly to the child session.

*VERTICAL SLICE SCOPE*
 
* Presentation / Waterfall:
  - In `TraceWaterfall.tsx`, check if an event represents a `skill` tool call with a non-empty `skill_name`; render a styled badge (`.badge-skill`) displaying the skill name.
  - Check if an event represents a `task` tool call with `child_session_id`; render a `.badge-task-link` button.
  - Clicking `.badge-task-link` triggers session selection, jumping the waterfall to the target child session.
* Presentation / Live Inspector:
  - In `LiveTraceInspector.tsx`, update turn tool lists to render `[skill: <name>]` and clickable `[task -> <child_session>]` jump chips.
* CSS:
  - Add accessible styling in `web/src/index.css` for skill badges (distinct violet/indigo tone) and subagent delegation links.
* Quality:
  - Component tests in `TraceWaterfall.test.tsx` and `LiveTraceInspector.test.tsx` verifying badge formatting, long name truncation with tooltip, and click-to-navigate handlers.

*ACCEPTANCE CRITERIA*
 
* Skill invocations render as `[skill: <name>]` badges across both waterfall and live inspector views.
* Task tool invocations render with subagent role badge and a clickable session link whenever `child_session_id` is present.
* Clicking the child session link smoothly selects and displays the child session without page reload.
* Graceful fallback to plain `skill` or `task` label when metadata is missing.
* `npm test`, `npm run typecheck`, and `npm run lint` pass cleanly.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
