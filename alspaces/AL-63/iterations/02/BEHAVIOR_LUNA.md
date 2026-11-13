# Behavior review: historical hydration and live trace

**Outcome: APPROVED (scoped behavior review).** No blocking behavior defect found in the inspected paths. This does not mean native harnesses were observed or all capability claims verified.

## Identity and evidence

Requested profile: `codex / gpt-6-luna / medium` (acknowledged). Native execution metadata was unavailable; no claim that this run was confirmed as that model/profile. Candidate: `f4152225af06e2594477574c0e9f175a81218053`. Scope: historical event hydration, live HTTP trace and execution-error reporting, and error/partial behavior for OpenCode v1/v2 and Codex. No native OpenCode/Codex source database or harness version/session IDs were inspected. No tests or probes were run; source inspection only. Parent reports unchanged application suites passed at `0a8cfcee`; reused as context, not independent verification.

## Capability matrix (declared support vs observed runtime)

| Source | Historical trace hydration | Live observation / HTTP trace | Execution errors and enrichment failure |
|---|---|---|---|
| OpenCode v1 | **Declared** via schema-family dispatch to v1 exact-owner source extraction (`opencode/extractor.go:194-215`, `schema_family_detection.go`). No native run here. | **Declared** via configured read-only SQLite hook and live trace API (`cmd/agentic-lens/main.go:812-824`, `live/service.go:378-430`). No native run here. | Source event extraction returns partial events with an error when strict parse fails (`extractor.go:206-216`); hydrator records diagnostics and classifies partial/unavailable. Live `/api/live/sessions/{id}/trace` returns errors as HTTP 400/404/500, while parsed tool failures populate `failed_tool_calls`/`recent_errors` (`http/live_handlers.go:46-79`, `live/service.go:readSessionTrace`, `parsePartRecord`). Not exercised against a native database.
| OpenCode v2 | Same declared path, selected by exact schema family; extraction dispatches v2 at `extractor.go:200-205`. | Same OpenCode SQLite live path. No separately versioned live-observation adapter is visible in the inspected call path; native v2 compatibility remains unverified. | Same declared handling. Schema-v2 fixture tests exist (`backend/internal/adapters/opencode/schema_v2_session_extraction_test.go`), but no test was run in this review and fixtures are not native observations.
| Codex | **Declared** via exact `SessionID` rollout discovery, parse, parent consistency check, and combined source events (`live/service.go:842-899`; Codex parser in `adapters/codex`). Missing/malformed owners become diagnostics and partial/unavailable; no synthetic empty success. | No Codex-backed active-session tree/trace HTTP reader was found in the inspected live service: its live trace path requires the SQLite hook. Thus support for Codex historical hydration is distinct from live observability. | Historical parse warnings mark incomplete coverage; available events are retained. Live API database errors use HTTP 500; missing IDs use 404. No native Codex rollout was read in this run.

## Persisted-fact boundary and error path

`WorkspaceEventsHandler` loads the ledger first, builds an unavailable response retaining recorded anomaly summaries, then optionally hydrates. Hydration failure/invalid response yields an unavailable response with diagnostics; the GET response does not write the ledger (`http/handlers.go:245-299`). `HydrateWorkspaceEvents` reads each recorded owner from its corresponding local source; unsupported/missing owners are diagnosed. It derives source anomaly associations from the response's transient `SourceEvents`, while the domain validator forbids committed facts from qualifying as source trace evidence (`domain/trace.go:1-55`). This supports keeping persisted facts separate from request-local enrichment. I did not run a disposable HTTP test for the failure branch.

For live OpenCode trace, service errors are propagated to HTTP status handling, and tool execution failures are represented within a successful trace rather than conflated with request failure. Child lookup errors are ignored for child linkage (`live/service.go:readSessionTrace`), so missing child IDs can remain unlinked; this is a capability limitation, not evidence that the parent trace request failed.

## Hypotheses checked and limits

- **Optional hydration erases stored facts:** not supported by the handler path; it loads and returns the stored ledger independently, while trace events are transient.
- **Missing source becomes successful empty trace:** not supported by the visible status/coverage construction and `WorkspaceEventsResponse.Validate` rules. Native sources were unavailable, so real-world parser compatibility is unverified.
- **Execution text mentioning an error becomes failed execution:** parser classification uses tool status/output through `ClassifyToolExecution`; Codex parser has explicit command status handling (e.g. source test `rollout_test.go:456`). No source execution was performed here.

No candidate finding. Recommendation/limit for the matrix: mark all three families' runtime/native observations **not executed** in this review; distinguish repository fixture coverage from actual harness-version verification. OpenCode v2 live compatibility and child-lookup diagnostic behavior remain evidence gaps.
