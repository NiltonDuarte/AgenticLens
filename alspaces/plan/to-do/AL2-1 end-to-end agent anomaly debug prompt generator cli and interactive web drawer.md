h3. Title: AL2-1: End-to-End Agent Anomaly Debug Prompt Generator, CLI, and Interactive Web Drawer

*TARGET*: AgenticLens (domain | adapters | http | cli | presentation | web | tests)

*Status*: Ready
*Blocked by*: None
*Blocks*: None

*CONTEXT*
 
When an autonomous coding agent encounters execution failures, consecutive tool errors, context saturation spikes, or subagent rework loops, the browser already offers evidence-preserving prompt generation and copy actions from AL-50. CLI/API access and a consistent preview flow remain missing. Extend the existing capability without rebuilding shipped browser controls or changing raw evidence. Delivering an end-to-end vertical slice from pure domain prompt synthesis to REST API, CLI, and interactive web dashboard enables engineers to generate and copy context-rich, ready-to-use LLM debugging prompts with a single click or command.

*GOAL*
 
Implement an end-to-end agent debug prompt generator cutting through Go domain prompt synthesis, REST API endpoint, CLI command, and TypeScript web dashboard components (integrated into the Anomaly Drawer, Waterfall inspector, and Live Anomaly Banner with one-click clipboard copying).

*VERTICAL SLICE SCOPE*
 
* Domain Core: Provide deterministic domain prompt synthesis for the CLI/API while preserving the current browser prompt guarantees. Format structured markdown containing `# Incident Overview`, `## Anomaly Diagnostics`, `## Tool Invocation Forensics`, `## Preceding Execution Trace`, and `## Prescriptive Remediation Directives` tailored to rule types (`TOOL_RETRY_STREAK`, `TOOL_HIGH_FAILURE_RATE`, `CONTEXT_SATURATION_SPIKE`, `SUBAGENT_REWORK_LOOP`). Pure functions, value semantics, zero I/O imports.
* Boundary Adapter: Add `GET /api/workspaces/{ticketId}/debug-prompt` in `internal/adapters/http/handlers.go` accepting optional `turn`, `rule_id`, and `tool_name` parameters, returning JSON with prompt markdown and diagnostic metadata. Support available native-source hydration and gracefully fall back to supported committed facts when sources are unavailable. Do not fabricate missing raw evidence.
* CLI Presentation: Implement `agentic-lens prompt` subcommand in `backend/cmd/agentic-lens/main.go` supporting `--ticket <ticketId>`, `--turn <n>`, `--rule <rule_id>`, and `--output <file>`, outputting markdown to stdout or file.
* Web Presentation: Define `DebugPromptResponse` in `web/src/types/telemetry.ts` and `fetchDebugPrompt` in `web/src/api/client.ts`. Extend the existing anomaly and trace prompt controls with a prompt viewer with rendered markdown preview, raw text area, and one-click "Copy Prompt" button with transient feedback state ("Copied!"). Integrate prompt triggers on error nodes in `Waterfall.tsx` and alerts in `LiveAnomalyBanner.tsx`.
* Quality: Comprehensive table-driven Go tests in `internal/domain` with zero mocks, HTTP endpoint tests, and CLI execution tests. Vitest tests in `web/src/components/AnomalyDrawer.test.tsx` and `DebugPromptViewer.test.tsx` verifying fetching, loading states, clipboard copying, and error fallbacks.

*ACCEPTANCE CRITERIA*
 
* The shared domain generator synthesizes deterministic, richly structured markdown including exact tool inputs, error messages, and rule-tailored directives.
* `GET /api/workspaces/{ticketId}/debug-prompt` returns HTTP 200 with `DebugPromptResponse` for valid tickets, HTTP 404 for unknown tickets, and HTTP 400 for malformed parameters.
* `agentic-lens prompt --ticket <ticketId>` outputs formatted markdown prompt to stdout and exits with status 0.
* Clicking "Generate Debug Prompt" in `AnomalyDrawer.tsx` queries the API and presents the prompt preview.
* Clicking "Copy Prompt" copies prompt text to the system clipboard via `navigator.clipboard.writeText` with visual feedback.
* Waterfall inspector tool nodes with error status and live alerts in `LiveAnomalyBanner.tsx` offer immediate triggers to open the prompt drawer.
* Zero `any` in Go and TypeScript code; all static checks and test suites pass (`make check`, `make test`).

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.

*DEVELOPMENT STATE REVIEW (2026-11-05)*

Partially implemented: AL-50 and current web/src/utils/debugPrompt.ts already provide browser prompt generation and copying. Remaining scope is CLI/API availability, persisted-fact fallback and a consistent preview/entry flow; preserve verbatim raw payloads and reuse existing controls.

Evidence and review limits: [PO-19 reconciliation](../../PO-19/SPEC.md), merged baseline `03f5b05d`. This planning review does not certify runtime behavior or close unverified acceptance criteria.
