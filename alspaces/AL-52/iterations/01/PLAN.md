# AL-52 Iteration 01 Plan

## Authorized scope

Source: the linked AL-52 specification and the user's explicit pre-approval on 2026-10-31. No new acceptance scope is introduced. Contract: ../../CONTRACT.md. Baseline 85a2af94b5036b0918aaf9a7a89aed3a990f2650; initial status clean. Preserve unrelated changes in the shared checkout.

## Grounded implementation

1. Inspect existing rendering and relevant tests before edits. web/src/components/TraceWaterfall.tsx renders EventFact rows; EventFact carries tool_input but no rich direct attributes. web/src/domain/types.ts already defines SkillFact, TaskLinkFact, ledgerSkills and ledgerTaskLinks. backend/internal/adapters/opencode/skills_tasks.go produces turn and timestamp metadata but no parent session ID on these collections. Use a small pure metadata projection with conservative unique matching across all events; do not join solely by turn or consume ambiguous duplicates. Valid event tool_input can provide a skill name or task role where optional fact arrays are absent. Child identifiers must come from recorded task links, not guessed output text.
2. Add a shared presentation badge/navigation component if it reduces duplication. TraceWaterfall accepts optional metadata and child-selection availability/callback props; TicketDeepDive supplies ledger facts and known session IDs. Navigate a known retrospective child to its actual thread/session details using a focused, expanded ThreadViewer record or an equivalently direct existing deep-dive view. Preserve the unfiltered trace and current defaults.
3. In web/src/components/LiveTraceInspector.tsx, forward child navigation props through its internal LiveMessageTurnItem and LiveMessagePartItem. Render LivePartFact.skill_name, subagent_type and child_session_id. web/src/views/LiveObservability.tsx already owns selectedSessionId, handleSelectSession and findSessionNode: reuse those boundaries and supply the current known child set. On selection changes, keep displayed trace bound to the selected session so the parent stream does not appear beneath the child heading while a request is pending.
4. For a recorded unavailable child, expose full ID in an informative tooltip and accessible copy action using web/src/utils/clipboard.ts. Do not issue speculative fetches or create fake session nodes. Keep ID access available on narrow screens. Prevent mouse and keyboard bubbling from child actions into TraceWaterfall row selection.
5. Update web/src/index.css using theme-compatible colors for skill and task badges. Apply min-width: 0, nowrap, overflow and ellipsis to the label rather than clipping the entire interactive region. Preserve adjacent tooltips, table containment and scroll behavior.

## Tests and plausible regressions

- Extend TraceWaterfall.test.tsx with actual rendered skill/task metadata, committed facts with unavailable hydration, absent optional arrays, and repeated timestamp/turn occurrences across sessions that must not gain an incorrect child link. Existing filter/collapse and anomaly coverage remains authoritative.
- Extend LiveTraceInspector.test.tsx for direct attributes, available child callback, unavailable ID tooltip and clipboard behavior, and controls usable by keyboard. Verify child navigation does not trigger parent event selection.
- Extend LiveObservability.test.tsx for parent-to-child selection and loading/current-trace identity. Extend TicketDeepDive.test.tsx and ThreadViewer coverage as needed to prove the retrospective action reaches the correct expanded child details rather than only calling a mocked callback.
- Long metadata remains accessible through complete labels/tooltips; CSS source or DOM assertions establish styling intent only, not browser geometry. Do not claim browser geometry verification without evidence.

## Selected regression checklist boundaries

Persisted facts and optional hydration: recorded facts survive unavailable enrichment; ambiguous repeated occurrences retain identity. UI state/layout: preserve filters, defaults, full identifier access, keyboard interaction and adjacent tooltip visibility. Polling/asynchronous updates: reuse existing work bounds, ignore obsolete session trace responses and never display old trace under a new selection. Frontend embedding: canonical frontend build, matching ignored asset trees, existing embedded-assets tests and standalone binary verification. Ticket fixtures are self-contained and independent of live alspaces contents.

## Execution and evidence ownership

Coder owns presentation/helper/CSS and necessary view wiring paths; tester owns relevant test files after coder handoff; reviewer audits read-only and owns REVIEW.md. Parent owns integration and delivery artifacts. Every delegate receives this plan, the contract, baseline and current diff and preserves shared changes. Record focused test commands/results, then reuse them for unchanged code; run make after merging origin/main. At delivery update SUMMARY.md and PULL_REQUEST.md, synchronize the ticket specification/index, and generate real Codex-source telemetry only at finalization. Never substitute fixture or empty telemetry for unavailable source evidence.

## Routine integration steering, 2026-10-31

Source: explicit MGT-1 management instruction from the user. Integrate origin/main after #76 and #79, preserving active refresh, scroll/reasoning anchors and AL-52 child badges/navigation. No acceptance boundary changes or new product scope; existing contract and approval remain valid. No speculative #77/#78 merge, PR merge or auto-merge. Later AL-53 owner will unify ThreadViewer navigation. Initial continuation status clean at f57dbf87167439313de57a58f4bc0c6f64548b35. Merged base b1268234cde742705728ea33ee89d30581803c47 without conflicts into bc312da81ebae5f0270ff3bb5eb3b36196587ccb. Routine lifecycle correction removes AL-52 from AL-55 blockers, preserves AL-54 removal and AL-53 blocker; completed entries remain absent. Full make, embed/standalone checks and independent integrated review required before push. Preserve source sessions in final telemetry; manager retains merge sequencing.
