h3. Title: AL2-2: Interactive Debug Prompt Drawer, Clipboard Integration, and Waterfall Actions in Web Dashboard

*TARGET*: AgenticLens (presentation | web | components | tests)

*CONTEXT*
 
While AL2-1 delivers backend prompt synthesis and CLI generation, web dashboard users inspecting anomalous sessions currently have no visual mechanism to request, preview, or copy agent debugging prompts directly from the browser. Adding first-class prompt generation controls to the Anomaly Drawer, Waterfall event inspector, and Live Anomaly Banner enables engineers to debug stuck agents with a single click during active or post-mortem investigations.

*GOAL*
 
Build an interactive prompt viewer component in the TypeScript web frontend integrated directly into `AnomalyDrawer.tsx`, the Waterfall trace inspector, and `LiveAnomalyBanner.tsx`, providing instant prompt preview, markdown rendering, raw text toggling, and one-click clipboard copying.

*VERTICAL SLICE SCOPE*
 
* Domain Core: Define strict TypeScript interfaces for `DebugPromptResponse` and request parameters in `web/src/types/telemetry.ts` matching backend contracts with zero `any`.
* Boundary Adapter: Add `fetchDebugPrompt(ticketId: string, params?: { turn?: number; ruleId?: string; toolName?: string }): Promise<DebugPromptResponse>` in `web/src/api/client.ts` with error handling.
* Presentation: Add "Generate Debug Prompt" button inside `AnomalyDrawer.tsx`. Create `DebugPromptViewer.tsx` component with rendered markdown preview, raw copyable text area, and "Copy Prompt" button with transient feedback state ("Copied!"). Integrate quick-action debug prompt buttons on failed tool nodes in `Waterfall.tsx` and alerts in `LiveAnomalyBanner.tsx`.
* Quality: Comprehensive Vitest unit tests in `web/src/components/AnomalyDrawer.test.tsx`, `DebugPromptViewer.test.tsx`, and `LiveAnomalyBanner.test.tsx` verifying prompt fetching, loading skeletons, error fallback, clipboard copy interaction, and modal/drawer state transitions.

*ACCEPTANCE CRITERIA*
 
* Clicking "Generate Debug Prompt" in `AnomalyDrawer.tsx` queries `GET /api/workspaces/{ticketId}/debug-prompt` and displays the formatted prompt.
* Clicking "Copy Prompt" copies prompt text to the system clipboard via `navigator.clipboard.writeText` and displays visual confirmation.
* Waterfall inspector tool nodes with error status display an inline "Debug Prompt" trigger that opens the drawer focused on that tool turn.
* Live anomaly alerts in `LiveAnomalyBanner.tsx` include a "Debug Prompt" action opening the drawer with the live session alert context.
* Zero `any` in TypeScript code; strict type checking passes cleanly (`npm run typecheck`).
* Frontend tests pass with zero warnings or failures (`npm test`).

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
