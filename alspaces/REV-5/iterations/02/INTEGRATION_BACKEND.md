# REV-5 backend integration

Premerge aa86f4b; incoming main 47af923. Owned backend/cmd and backend/internal/adapters, preserved concurrent domain/frontend changes and incoming staged history. No staging, commits, publishing or telemetry by this child.

## Conflict decisions

- extractor.go keeps math for source-cost rounding; obsolete sort import and ExtractFacts stay removed. The incoming schema-v1/v2 cost cases now exercise ExtractSessions and AgenticFactLedger.InvocationFacts. All partial/observed-zero/null/negative-cost and coverage assertions remain with their real SQLite fixtures.
- schemas.go remains deleted, honoring main's split into persisted fact, cost, anomaly, computed export and projection DTO modules. Required JSON hooks, fact reverse conversions, strict cost evidence handling and wire field ordering remain. The split had reintroduced two previously deleted unused functions, ValidateToolFacts and containsString; removed them again after confirming zero callers.
- schemas_test.go keeps main's output-only computed export contract. Removed the conflicted legacy AgenticMetricsLedgerDTO.ToDomain test and ordinary JSON-array rejection test, matching main's removal of that reverse conversion. These tests no longer describe an ingestion API. Actual persisted AgenticFactLedgerDTO legacy/canonical reads, anomaly codecs and all incoming cost-evidence validation tests remain.
- Reviewed automatic merges in CLI, Codex extraction/tests, HTTP DTOs/tests, formatter and repository boundaries. Existing rollback/config tests and source revision cache remain unchanged; new Codex facts still carry unavailable cost, historical persisted numeric zero remains observed, and v4 reads plus v5 persistence retain upstream semantics.

## Additional incoming-code review

Inspected typed invocation/task/evidence conversion, canonical runtime validation, new HTTP/projection DTO responsibilities and source-cost accumulation. Persisted fact ingestion and computed/HTTP exports remain separate contracts. No new blocking defect identified in those inspected paths. The typed DTO volume supports distinct wire contracts rather than restoring removed generic reverse conversions.

Re-ran Go AST declaration/reference screening across the merged source. It identified a new test-only domain SummarizeCost function; domain owner removed it. Two HTTP cost fixture preparations now use the active ComputeTotals(...).CostSummary path with unchanged assertions. Implicit JSON/YAML codecs are retained despite having no explicit identifier callers. This is conservative identifier screening, not exhaustive whole-program liveness proof.

## Verification

Full go test -race ./cmd/agentic-lens ./internal/adapters/... passed with existing valid generated embed assets. Initial compile attempt exposed four incoming source-cost test references to removed ExtractFacts; retargeted before the passing race run. The final HTTP-only race rerun passed after the SummarizeCost fixture correction. Final Go AST screening reports only implicit UnmarshalJSON/UnmarshalYAML methods. Backend whitespace check passed. No conflict markers remain in owned source. Parent owns merge bookkeeping, independent test/review, canonical frontend rebuild/full make and actual wire smoke.
