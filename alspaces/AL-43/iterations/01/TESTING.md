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
