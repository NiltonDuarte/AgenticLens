# docs(plan): tool execution vs content error disentanglement and pipeline parity (PO-10)

## Summary
Defines the product specification, architectural contracts, risk analysis, and 2 vertical tracer-bullet Jira tickets (AL-42 and AL-43) to resolve tool error conflation and unify live and offline telemetry pipelines in v1:
1. Unification of tool execution error classification strictly around hard process failures (non-zero exit codes, crashes, timeouts) with complete consistency across offline and live pipelines, eliminating false positives from benign search operations like grep/glob with 0 matches (AL-42).
2. First-class differentiation and telemetry representation for semantic content errors in completed tool calls (compiler errors, traces, unhandled exceptions) across domain aggregations, schemas, APIs, and the TypeScript dashboard with dual-badge visualization (AL-43).

## Initiatives Covered

1. **Tool Execution Error Classification & Pipeline Parity (AL-42)**:
   - Pure domain execution classifier isolating process exit codes and signals from output strings.
   - Elimination of erroneous error classifications on benign exploratory commands (`grep` and `glob` with 0 matches).
   - Migration of `internal/adapters/opencode/extractor.go` to use the shared domain classifier, guaranteeing offline and live telemetry parity.
   - Refinement of `TOOL_RETRY_STREAK` and `TOOL_HIGH_FAILURE_RATE` anomaly rules to strictly monitor genuine tool execution failures.

2. **Semantic Content Errors in Successful Tool Calls (AL-43)**:
   - First-class tracking of semantic content errors on completed tool calls (`ContentErrors` in `ToolFact`, `HasContentError` in `EventFact` and `LivePartFact`).
   - Pure classifier `ClassifyToolContent` identifying stack traces, panics, and compiler errors without mutating tool execution status to `error`.
   - Backward-compatible schema serialization defaulting `ContentErrors` to 0 for older ledgers.
   - Web presentation in `TraceWaterfall.tsx`, `LiveTraceInspector.tsx`, and `TicketDeepDive.tsx` featuring distinct red (execution failure) and amber (content warning) visual badges.

## Ticket Breakdown
- Ready: `AL-42: Unify Tool Execution Error Classification and Offline-Live Pipeline Parity` (Phase 7 - Ready, blocks AL-43)
- Pending: `AL-43: Differentiate and Surface Semantic Content Errors in Successful Tool Invocations` (Phase 7 - Pending, blocked by AL-42)

## Pointers
- Specification: `workspaces/PO-10/SPEC.md`
- Shared Roadmap: `workspaces/plan/TICKETS.md`
- Development Tickets:
  - `workspaces/plan/to-do/AL-42 unify tool execution error classification and offline live pipeline parity.md`
  - `workspaces/plan/to-do/AL-43 differentiate and surface semantic content errors in successful tool invocations.md`
- Session Summary: `workspaces/PO-10/SUMMARY.md`
