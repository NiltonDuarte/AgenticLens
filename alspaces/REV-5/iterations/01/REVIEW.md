# GEN-14 repository audit

Baseline: `36bea1f18f2256392428329136b3241e98d09854`. Findings remain open; this PR changes only audit artifacts.

## Method and limits

Read configured roots and architecture/testing rules; examined recent 50-commit history. Parsed all 111 Go files with Go AST and all 89 TypeScript/TSX files under web/src with TypeScript AST. Inventoried named functions against production/test identifier occurrences, then inspected candidate bodies, runtime alternatives, and selected test peers. Included CLI, adapters, domain, frontend API/domain/views/components; screened Makefile and build configuration. This is repository-wide screening with deeper review of findings, not a line-by-line proof of every function or exhaustive static call graph. Scripts/workflows received build-path screening, not full security audits.

After excluding implicit codec methods, the Go screen produced 42 zero-production-reference function names, 24 with test identifier references. The frontend function screen produced nine test-referenced exports. These are screening candidates, not a certified dead-code total; the additional unread-buffer chain was found by manual tracing.

Identifier inventory is conservative: shared names can hide dead methods; occurrences are not call counts. Codec methods invoked by encoding/json and YAML are false positives and explicitly excluded. Tests, web/src/test fixtures/setup, embedded/generated assets, and dependencies are not application production candidates. No claim that all unreachable declarations were found. No test deletion, mutation testing, browser geometry validation, release publishing, or performance profiling performed. No domain I/O/clock or actual any/interface{} violation found in the targeted screen. Stateful live subscription/cache ownership is intentional; the redundant buffer below is a concrete issue rather than a blanket objection to state.

## Prioritized findings

### F1 - P1: Failed forced initialization removes the existing config from its path

`backend/cmd/agentic-lens/main.go:1122` through `1154`: writeConfigPair renames the first existing target to a backup, then returns directly for a directory or stat failure on the second target without invoking rollback. With an existing shared config and a directory named .agenticlens-config.local.yaml, init --force fails and the shared config disappears from its expected path. Its bytes survive only in a randomly named backup. This is a functional rollback defect, not just style.

Verified with a baseline-built CLI in a temporary directory; see rollback-reproduction.json. Exit 1, original_path_exists=false, backup retains original bytes. The existing test at main_test.go:5146 covers the local-directory failure only when the shared config does not yet exist, so it misses preservation of an existing file.

Proposed fix: validate both destinations before moving either; install rollback protection before the first destructive step; report rollback failures. Also arrange staged-file cleanup before staging the second file (current deferred cleanup is registered after both writes). Blast radius: init only. Required proof: extend existing init failure cases with an existing first file and assert exact byte preservation after second-target failure; retain force/collision tests.

### F2 - P2: Live trace buffer has runtime writes but no runtime reads

`backend/internal/adapters/live/service.go:87,99,573` and `live/cache.go:12`: every source trace rebuild copies messages into TraceBuffer, but GetMessages and Len are read only by sqlite_test.go. Buffer itself is only exposed to tests. The response uses the local messages slice and the actual revision cache lives separately in traceCache (service.go:398). Clear has no callers, so the session map has no eviction path. Per-session truncation does not bound the number of retained sessions.

Delete the unused buffer, its construction/write/accessor and buffer-only tests, while retaining traceCache and the actual source-backed trace tests. Do not delete or reinterpret configured max trace settings without tracing remaining consumers and defining their intended contract. Benefit: remove copies, locks, retained message slices, and a misleading second caching API. No measured memory or latency estimate. Validate trace ordering/content, revision reuse and concurrent reads through existing service tests; remove only assertions about the redundant internal buffer.

### F3 - P2: Frontend test-only progression client asserts the wrong response shape

`web/src/api/client.ts:203`: fetchProgressionLedgers casts /api/progression to AgenticMetricsLedger[], but `backend/internal/adapters/http/handlers.go:401` returns AggregateMetricsOverTime buckets. Its test mock supplies ledgers, masking the incompatible wire contract. It has no application caller; MetricProgression uses fetchWorkspaceLedgers (MetricProgression.tsx:74). fetchProgressionMetrics at client.ts:397 has the correct bucket shape but is also test-only.

Delete the obsolete ledger-returning client and its invalid mock-contract tests; keep HTTP progression behavior tests. Review the other three test-only client exports (fetchComparisonRows, fetchWorkspaceAnomalies, fetchProgressionMetrics) for deletion rather than maintaining unused parallel fetch APIs. Do not delete public server routes merely because this UI does not use them. Scope: frontend client surface; retain live consumers and their failure/diagnostic tests.

### F4 - P2: Obsolete CLI merge helpers are tested instead of executed

`backend/cmd/agentic-lens/main.go:1403,1432`: mergeTools and mergeEvents have only test callers in merge_internal_test.go. Production generate refresh uses mergeSessions at main.go:299. The old event helper deduplicates by timestamp/session/turn/type, omitting event identity; its continued tests do not protect the production owner-refresh path.

Delete helpers and their exclusive key types after verifying no remaining references. Remove their obsolete algorithm-specific tests; preserve existing mergeSessions assertions for ownership, multiplicity, idempotence and unavailable source preservation (merge_internal_test.go:45,84). Do not move tests to a test-only copy of the dead algorithm or claim passing tests prove equivalence. Benefit: remove a misleading alternate refresh implementation.

### F5 - P2: Legacy test-only domain and adapter APIs retain parallel behavior

Representative locations: domain/aggregations.go:532 (ComputeThreadSummariesFromSessions forwards directly), domain/live.go:143 (ClassifyToolFailure compatibility tuple), adapters/opencode/extractor.go:141 (ExtractFacts flattens session-owned data), domain/aggregations.go:1447 (unused temporal filter), domain/live.go:260 (EventsFromLiveTrace alternate conversion).

See inventory for the wider candidate set. Pure wrappers should disappear first; meaningful extractor/classification/filter assertions should be retargeted to the actual session extraction, typed classification, and plural-filter APIs. ExtractFacts has substantial test coverage, so its tests must be inspected case by case before deletion: real SQLite/schema cases remain valuable. EventsFromLiveTrace is not the current hydration conversion; production uses convertPartToEventFact in live/service.go:1236. Do not assume its entire test suite is redundant without mapping provenance and event-kind assertions to the real path.

Frontend counterparts: computeSkillAdoption (domain/aggregations.ts:17), formatIsoDateOnly, filterLedgersByCompletionDate, extractDistinctPrefixes and filterComparisonRows (domain/formatters.ts:163,268,506,601) have test references only. Remove duplicate calculations only after deciding whether the server or active frontend helper owns each behavior; retain tests of the live owner. Blast radius: moderate because tests currently encode obsolete API shapes. These are removal candidates, not authorization for bulk test deletion.

### F6 - P3: Unreferenced formatter, schema, and compatibility surfaces obscure current contracts

AST/reference screening found unused formatters FormatISODatetime, FormatCurrency, FormatPercentage, FormatDuration, FormatCodeDelivery and FormatTicketLabel in adapters/formatters.go; unused LoggingMiddleware (http/middleware.go:26); and unreachable schema helpers validateAnomalyEvidence, containsString, evidenceRefsToDomain and evidenceRefsFromDomain (schemas.go:645,429,856,867). These have neither production nor test identifier references. InfoIcon at web/src/components/InfoTooltip.tsx:82 is similarly unused.

Prefer deletion over new tests for dead implementations. In particular, do not wire the old evidence validator back in as cleanup: recent AL-43 changes removed behavior, and enabling a dormant validator changes accepted persistence contracts. Check the canonical schema contract before removing related DTO fields, which may still be wire compatibility data. Preserve codec entry points UnmarshalJSON and UnmarshalYAML despite their lack of explicit calls. Inventory lists further candidates with confidence limited to repository source reachability.

### F7 - P2: Most frontend HTTP payloads bypass runtime validation

`web/src/api/client.ts:82,117,159` and similar fetch functions cast res.json() to domain models. diagnosticsFromResponse checks only Array.isArray, not diagnostic element shape. TypeScript assertions do not validate server/proxy/version-skew responses. Unlike a proven current incident, this is a boundary contract weakness: malformed successful responses reach views as supposedly valid values.

After deleting unused fetchers, add focused boundary decoders for supported payloads and reject malformed structures with explicit request errors. Avoid speculative framework-wide interfaces. Preserve current optional-field compatibility and workspace diagnostics. Required proof: malformed HTTP-200 bodies and valid legacy/optional shapes through real client functions and existing view recovery tests. Scope: active client boundaries and consumers.

## Test decisions and follow-up order

No tests removed. Prioritize F1 regression correction, then F2 and F4 subtraction, then obsolete frontend API removal, then domain/adapter surface consolidation. F1's existing test has a real oracle but lacks an existing-file branch; extend it rather than create a duplicate suite. F3's mocked ledger response is an invalid server contract. F4's helper-specific tests guard unreachable behavior; live mergeSessions tests must remain. F5 requires per-assertion migration before any test reduction. Do not delete independent adapter/HTTP/DOM coverage based on counts.

Small interfaces at consumed HTTP boundaries are justified by their role; no recommendation to remove them merely because there is one live implementation. Single binary embed wiring remains in place. Unknown cost/source facts and session ownership must remain unchanged by future cleanup.

## Verification

See SUMMARY.md for final integrated gate and telemetry results. AST tools and raw candidate inventories are included for reproducibility. Inventory counts describe parsed source files and identifier occurrences, not runtime test case totals or executable coverage.
