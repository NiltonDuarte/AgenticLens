h3. Title: AL2-1: Agent Anomaly Debug Prompt Generator Engine, HTTP Endpoint, and CLI

*TARGET*: AgenticLens (domain | adapters | http | cli | tests)

*CONTEXT*
 
When an autonomous coding agent encounters execution failures, consecutive tool errors, context window saturation spikes, or subagent rework loops, engineers must manually inspect log files and event waterfalls to construct guidance prompts for the agent. This manual triage slows down development and often produces incomplete instructions that cause the agent to repeat the same error. A deterministic, domain-driven prompt generation engine is needed to synthesize anomaly forensics and causal event context into actionable, LLM-ready debugging prompts available via HTTP API and CLI.

*GOAL*
 
Implement a pure Go domain prompt synthesis engine that transforms detected anomalies and causal tool failure events into structured markdown debugging prompts, expose the capability via `GET /api/workspaces/{ticketId}/debug-prompt`, and provide the `agentic-lens prompt` CLI subcommand for automated or terminal-driven agent steering.

*VERTICAL SLICE SCOPE*
 
* Domain Core: Implement `DebugPrompt` models and `BuildDebugPrompt(anomaly *Anomaly, event *EventFact, causalTrace []EventFact, ticketID string, cfg ProjectConfig) DebugPrompt` in `internal/domain`. Format markdown with structured sections: `# Incident Overview`, `## Anomaly Diagnostics`, `## Tool Invocation Forensics`, `## Preceding Execution Trace`, and `## Prescriptive Remediation Directives` tailored to rule types (`TOOL_RETRY_STREAK`, `TOOL_HIGH_FAILURE_RATE`, `CONTEXT_SATURATION_SPIKE`, `SUBAGENT_REWORK_LOOP`). Pure functions, value semantics, zero I/O imports.
* Boundary Adapter: Add `GET /api/workspaces/{ticketId}/debug-prompt` in `internal/adapters/http/handlers.go` accepting optional `turn`, `rule_id`, and `tool_name` query parameters, returning JSON containing formatted prompt and diagnostic metadata. Support hydrated events from local SQLite database when available.
* Presentation: Add `agentic-lens prompt` CLI subcommand in `backend/cmd/agentic-lens/main.go` supporting `--ticket <ticketId>`, `--turn <n>`, `--rule <rule_id>`, and `--output <file>`. Output formatted markdown directly to terminal stdout or write to file.
* Quality: Comprehensive table-driven unit tests in `internal/domain` with zero mocks validating all anomaly rule directive mappings and edge cases. HTTP handler integration tests and CLI subcommand tests validating query parameters, error responses, and output formatting.

*ACCEPTANCE CRITERIA*
 
* `BuildDebugPrompt` generates deterministic, richly structured markdown including exact tool inputs, error messages, and rule-tailored directives.
* `GET /api/workspaces/{ticketId}/debug-prompt` returns HTTP 200 with `DebugPromptResponse` JSON for valid tickets, HTTP 404 for unknown tickets, and HTTP 400 for malformed parameters.
* When hydrated database events are available, the prompt includes full causal tool inputs and stderr traces; falls back gracefully to pruned ledger facts when unhydrated.
* `agentic-lens prompt --ticket <ticketId>` prints formatted markdown prompt to stdout and exits with status 0.
* Zero any across all Go structs and handlers.
* `go test -race ./...` and `go vet ./...` pass with zero warnings or errors.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
