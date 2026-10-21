h3. Title: AL2-1: End-to-End Agent Anomaly Debug Prompt Generator, CLI, and Interactive Web Drawer

*TARGET*: AgenticLens (domain | adapters | http | cli | presentation | web | tests)

*CONTEXT*
 
When an autonomous coding agent encounters execution failures, consecutive tool errors, context saturation spikes, or subagent rework loops, engineers must manually inspect log files and event waterfalls to synthesize guidance prompts. This manual triage slows down development and often produces incomplete instructions that cause the agent to repeat the same error. Delivering an end-to-end vertical slice from pure domain prompt synthesis to REST API, CLI, and interactive web dashboard enables engineers to generate and copy context-rich, ready-to-use LLM debugging prompts with a single click or command.

*GOAL*
 
Implement an end-to-end agent debug prompt generator cutting through Go domain prompt synthesis, REST API endpoint, CLI command, and TypeScript web dashboard components (integrated into the Anomaly Drawer, Waterfall inspector, and Live Anomaly Banner with one-click clipboard copying).

*VERTICAL SLICE SCOPE*
 
* Domain Core: Define `DebugPrompt` models and implement `BuildDebugPrompt(anomaly *Anomaly, event *EventFact, causalTrace []EventFact, ticketID string, cfg ProjectConfig) DebugPrompt` in `internal/domain`. Format structured markdown containing `# Incident Overview`, `## Anomaly Diagnostics`, `## Tool Invocation Forensics`, `## Preceding Execution Trace`, and `## Prescriptive Remediation Directives` tailored to rule types (`TOOL_RETRY_STREAK`, `TOOL_HIGH_FAILURE_RATE`, `CONTEXT_SATURATION_SPIKE`, `SUBAGENT_REWORK_LOOP`). Pure functions, value semantics, zero I/O imports.
* Boundary Adapter: Add `GET /api/workspaces/{ticketId}/debug-prompt` in `internal/adapters/http/handlers.go` accepting optional `turn`, `rule_id`, and `tool_name` parameters, returning JSON with prompt markdown and diagnostic metadata. Support hydrated SQLite event traces with graceful fallback to committed ledger facts.
* CLI Presentation: Implement `agentic-lens prompt` subcommand in `backend/cmd/agentic-lens/main.go` supporting `--ticket <ticketId>`, `--turn <n>`, `--rule <rule_id>`, and `--output <file>`, outputting markdown to stdout or file.
* Web Presentation: Define `DebugPromptResponse` in `web/src/types/telemetry.ts` and `fetchDebugPrompt` in `web/src/api/client.ts`. Build `DebugPromptViewer.tsx` inside `AnomalyDrawer.tsx` with rendered markdown preview, raw text area, and one-click "Copy Prompt" button with transient feedback state ("Copied!"). Integrate prompt triggers on error nodes in `Waterfall.tsx` and alerts in `LiveAnomalyBanner.tsx`.
* Quality: Comprehensive table-driven Go tests in `internal/domain` with zero mocks, HTTP endpoint tests, and CLI execution tests. Vitest tests in `web/src/components/AnomalyDrawer.test.tsx` and `DebugPromptViewer.test.tsx` verifying fetching, loading states, clipboard copying, and error fallbacks.

*ACCEPTANCE CRITERIA*
 
* `BuildDebugPrompt` synthesizes deterministic, richly structured markdown including exact tool inputs, error messages, and rule-tailored directives.
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
