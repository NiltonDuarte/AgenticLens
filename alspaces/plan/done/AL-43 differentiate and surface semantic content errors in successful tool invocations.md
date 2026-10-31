h3. Title: AL-43: Differentiate and Surface Semantic Content Errors in Successful Tool Invocations

*TARGET*: AgenticLens (domain | adapters | presentation | web | tests)

*CONTEXT*
 
While AL-42 isolates hard process execution failures (non-zero exit codes and crashes), autonomous agents frequently run tools that succeed at the OS level (`status: "completed"`, exit 0) but produce semantic errors in their output (compiler syntax errors, unhandled promise rejections, test assertion failures, stack traces). Currently, if these are treated as execution failures, they corrupt tool reliability metrics; if they are ignored, critical failure forensics are lost. Treating semantic content errors as a distinct, first-class metric across domain models, aggregations, APIs, and the UI allows developers to track agent coding mistakes independently from infrastructure/tool breakage.

*GOAL*
 
Deliver end-to-end support for semantic content errors in completed tool calls: implement pure domain content classification, extend domain models and serialization schemas (`ContentErrors` on `ToolFact`, `HasContentError` on `EventFact` and `LivePartFact`), update offline and live adapters, and surface dual-indicator visualizations (Red for execution failure vs Amber for content errors) across the TypeScript web dashboard (Trace Waterfall, Live Inspector, and Deep Dive).

*VERTICAL SLICE SCOPE*
 
* Domain Core (`internal/domain/live.go`, `internal/domain/models.go`):
  - Implement pure classifier `ClassifyToolContent(toolName string, output string) ToolContentResult` returning `HasContentError bool`, `Severity FailureSeverity`, `Category string`, and `ErrorSnippet *string`.
  - Target semantic patterns in completed output: stack traces, panic messages, compiler errors (`syntax error`, `cannot find package`, `build failed`), and unhandled exceptions.
  - Extend `ToolFact` with `ContentErrors int64`.
  - Extend `EventFact` and `LivePartFact` with `HasContentError bool`, `ContentSeverity string`, and `ContentErrorSnippet *string`.
  - Ensure pure backward compatibility with legacy ledgers where `ContentErrors` defaults to 0.
* Boundary Adapters & HTTP (`internal/adapters/schemas.go`, `internal/adapters/http/dto.go`, `internal/adapters/live/service.go`, `internal/adapters/opencode/extractor.go`):
  - Extend `ToolFactDTO` with `ContentErrors int64` (`omitempty`).
  - Wire both offline extractor and live service to evaluate `ClassifyToolContent` on tool completions.
  - Expose `content_errors` and `has_content_error` across REST endpoints (`/api/workspaces/{ticketId}/events`, `/api/live/sessions/{id}/trace`).
* Web Presentation (`web/src/`):
  - Update TypeScript types in `web/src/types/telemetry.ts` and `web/src/domain/live.ts`.
  - Update `TraceWaterfall.tsx`: Render amber node outlines and `CONTENT ERROR` badges for completed tool calls containing semantic content errors, distinct from red `EXEC ERROR` badges.
  - Update `LiveTraceInspector.tsx`: Display side-by-side metric counters for `Failed Tools` (execution failures) and `Content Errors` (semantic errors).
  - Update `TicketDeepDive.tsx`: Add `Content Errors` column in tool execution breakdown tables.
  - Update `AnomalyDrawer.tsx`: Differentiate diagnostics between process execution crashes and semantic content warnings.
* Quality:
  - Table-driven Go unit tests for `ClassifyToolContent` with regex boundary cases, compiler outputs, and false-positive avoidance.
  - Serialization and deserialization tests verifying backward compatibility with legacy v6 ledgers lacking `content_errors`.
  - Vitest unit tests in `web/src/components/TraceWaterfall.test.tsx` and `LiveTraceInspector.test.tsx` asserting accurate rendering of dual badges and stat counters.

*ACCEPTANCE CRITERIA*
 
* Completed tool calls (`status: "completed"`) containing compiler errors or tracebacks have `has_content_error: true` without having their status changed to `"error"`.
* `ToolFact` tracks `errors` (execution crashes) and `content_errors` (output errors) separately.
* Deserializing existing legacy ledgers without `content_errors` defaults gracefully to 0 without errors.
* REST API endpoints include `content_errors` and `has_content_error` in JSON responses.
* Web dashboard renders red indicators for execution errors and amber indicators for content errors.
* Go test suite (`go test -race ./...`) and frontend test suite (`npm test`, `npm run typecheck`) pass with zero errors.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
