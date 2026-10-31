# AL-43 Iteration 01 Testing

Baseline: 85a2af94b5036b0918aaf9a7a89aed3a990f2650. Tested the working diff after coder implementation and the TypeScript diagnostic regex correction, before delivery integration. Parent owns the integrated full make gate and embedded binary verification.

## Added and extended guards

- domain/tool_content_test.go exercises real classifier and execution classifier together for compiler, TypeScript, Python traceback, panic, promise rejection, assertion failure, stack trace, prose, inspection/search, clean output and exclusion of failed calls from content totals.
- adapters/schemas_test.go exercises canonical JSON decoding, validation, production compact encoding and restoration with missing legacy fields and populated diagnostics. Content snippets and completed status survive persistence without hydration.
- adapters/http/server_test.go extends the existing recorded-data fixture through actual detail and events handlers with no hydrator, failed hydration and partial SQLite/Codex enrichment. Wire diagnostics survive, ownership remains intact, and GET leaves stored bytes unchanged.
- adapters/live/parity_test.go compares independent content totals and diagnostic event fields through the same SQLite source using offline extraction and live service.
- adapters/opencode/schema_v2_session_extraction_test.go checks a successful v2 command with semantic output retains completed status, snippet and content count without execution errors.
- adapters/codex/rollout_test.go checks Codex completed command output with exit zero through production parsing, preserving independent counts and diagnostics.
- TraceWaterfall.test.tsx verifies default filtering includes content warnings, renders separate CONTENT ERROR and EXEC ERROR badges, excludes clean events until All Events is selected.
- LiveTraceInspector.test.tsx verifies distinct execution and content counter values.

Tests extend coherent existing boundary coverage where possible. No tests were deleted. Fixtures are self-contained; no live ticket workspace or developer database is consumed.

## Finding and resolution

The first classifier run failed for standard TypeScript output `src/app.ts(3,2): error TS2345: bad argument`. Reported to parent; coder extended the compiler diagnostic matcher. The retained regression passed on the corrected product code.

## Executed verification

- `cd backend && go test -race ./internal/domain ./internal/adapters ./internal/adapters/live ./internal/adapters/opencode ./internal/adapters/codex ./internal/adapters/http`: PASS, all six package suites.
- `cd web && npm test -- --run src/components/TraceWaterfall.test.tsx src/components/LiveTraceInspector.test.tsx`: PASS, 104 tests across two suites.
- `cd web && npm run typecheck`: PASS.
- `git diff --check`: PASS.

No silent skips in these targeted commands. This evidence does not claim browser geometry, standalone embedding, full-suite completion or final telemetry; those remain parent delivery checks. Frontend tests prove DOM and filter behavior rather than visual geometry. Source-format coverage exercises existing fixture families, not external real-runtime compatibility.

## Manager review corrections F3/F4

Baseline 8c4e2dc7a25586b13af6a7af0c0859b73edef65b; verified working correction diff in events.go, events.ts and Codex rollout.go with the following additions to existing test owners. No application changes, upstream integration or telemetry performed by tester.

- F3: Go events_test.go exercises production ReconcileEvents with an older event gaining a warning, negative/missing local metadata preserving stored evidence, positive empty/none fields falling back, and nonempty positive fields enriching metadata. Checks preserved event/session identity, execution status, payload enrichment, idempotence and input immutability. TypeScript events.test.ts covers the corresponding cases, explicitly distinguishes absent versus false flags, and exercises both reconcileEvents and selectWorkspaceEvents.
- F4: Codex rollout_test.go exercises production ParseRollout for equivalent plain diagnostic strings, MCP-style object content arrays, ordered multiple text blocks, input_text arrays, nested command-output wrappers and serialized wrappers inside text blocks. It checks equivalent flags, snippets, severity and independent counts while retaining exact original ToolOutput. Clean, mixed clean/nontext, nontext-only, unknown-envelope and unknown-block cases remain clean. A separate lifecycle guard checks actual MCP begin/end events: completed output warns; failed execution stays an execution failure without also incrementing successful content warnings. Both preserve the raw envelope and call identity.
- The structured examples are explicitly synthetic fixtures of supported shapes. They are not claimed to be sanitized runtime evidence; source-shape provenance remains the coder's separately recorded evidence.

Verification on the correction diff:

- `cd backend && go test -race ./internal/domain ./internal/adapters/codex`: PASS.
- After adding the MCP lifecycle guard, `cd backend && go test -race ./internal/adapters/codex -run 'TestCodex(ContentEnvelopes|MCPContent)' -count=1`: PASS.
- `cd web && npm run typecheck && npm test -- --run src/domain/events.test.ts`: PASS; 31 tests.
- `git diff --check`: PASS.

No new product defects found in this correction pass. The initial frontend typecheck rejected explicit undefined optional fields in the new tester fixture; corrected it to represent absence by omitting properties, then reran successfully. Full make and final integration remain parent responsibilities subject to the documented manager hold. These checks prove supported fixture paths, not universal Codex-version compatibility.

## Actual upstream integration verification

Tested the resolved working combination of AL-43 head bf575e1afd66ba17a82c5fbc1c4f3b3a2e65fcc8 with fetched main bdef2d0ef98e37fe259921bc30ba2daefbfda717, before the parent merger commit. Tester changed only the existing LiveTraceInspector invocation test and this evidence log. No existing assertions were removed.

The extended live test now exercises a single completed warning task that retains its rich invocation metadata and known-child navigation while showing the content-error class/badge. The same fixture also renders the separate execution-error class/badge and independent counters, expands failed execution details and long reasoning, copies the completed task's raw diagnostic in its debug prompt, and retains unavailable child identifier copy success/failure assertions. This specifically guards the merge boundary between AL-43 warnings and upstream invocation navigation rather than duplicating standalone feature tests.

Existing upstream tests were reused for useLiveTrace selection ownership, late-response rejection, cancellation on pause/close, SSE versus REST ordering, same-turn reading anchors and reasoning preservation; unified repeated child navigation from badges/panel; adoption/filter analytics and recovery. The F3/F4 tests remain present and passing. No upstream test fixture, timeout, SSE implementation or assertion bypass was introduced by testing.

Executed on the resolved combination:

- `cd web && npm test -- --run src/components/LiveTraceInspector.test.tsx src/components/TraceWaterfall.test.tsx src/views/LiveObservability.test.tsx src/views/TicketDeepDive.test.tsx src/views/FleetOverview.test.tsx src/views/TicketComparisons.test.tsx src/domain/events.test.ts src/domain/aggregations.test.ts`: PASS, 259 tests across eight suites.
- `cd web && npm run typecheck`: PASS.
- `cd backend && go test -race ./internal/adapters/http -run 'TestLiveStreamAndTraceRecoverInPlaceSQLiteUpdate|TestLiveStreamHandler_SSEProtocol|TestLiveStreamHandler_InitialTreeSubscriptionCleanup' -count=1`: PASS. This executes the concrete merged #82 in-place SQLite recovery and SSE protocol/subscription cleanup regressions, without changing their timeout guards or bypassing real production subscription behavior.
- `cd backend && go test -race ./internal/domain ./internal/adapters/codex`: PASS; includes F3 additive enrichment and F4 structured output/raw-evidence guards.

Stable readiness: focused integration verification passed. Parent owns full make, canonical frontend/embedded asset comparison, standalone binary verification, merger commit, telemetry and final delivery. DOM classes prove state selection; they do not establish browser visual geometry. Existing anchor tests use deterministic geometry, and no browser run is claimed.
