# AL-57 iteration 05: Independent coverage and assertion-transfer audit

Status: PASS for independent body/contract audit. No additional tests were justified after inspecting the coder additions and existing real-boundary recipients. This report reuses their valid unchanged-source verification instead of duplicating suites; the parent owns the final integrated make gate.

Baseline: 412bc08e3199af04c03009ed60ca1bd3b592dd85, initially clean. All six iteration 05 scopes were explicitly approved. Tester remained read-only in adapter/CLI and frontend tests until their respective coder handoffs; no production or test edits were made by the tester. Only this report is tester-owned. Frozen cost semantics and the iteration 04 filtered-average exception remain unchanged.

## Domain implementation and coverage audit

Read the complete aggregation/cost diff and the affected test bodies, not just test names or counts. accumulateInvocations now merges evidence inside its existing token loop and projects its rounded scalar from that summary. ComputeTotals counts rows directly. Thread model projection reuses the same accumulator. Bucket/cumulative scalar values derive from the ordered summary merges. These changes retain pure domain boundaries, row multiplicity, empty/unknown/partial/zero behavior, tiny complete-cost ordering and configured context health.

The coder's existing-owner additions cover the actual changed boundaries:

- TestComputeTotals retains its evidence table and now asserts rounded scalars and invocation counts, including complete/partial tiny contributions merged before rounding.
- TestComputeRoleAndModelMetrics retains complete-first tiny-cost model/thread ordering and independently checks scalar precision, partial/unavailable state, token/cache/peak/health fields. TestConfigurableHealthyCeilingAggregations adds the thread per-model custom-ceiling recipient.
- TestCalculateFleetSummary uses actual DeriveMetricsLedger on two source ledgers reporting 0.00004 each. Scalar ledger/fleet/role/model sums remain 0 while complete raw summaries retain 0.00008. Its nonempty CalculateRoleEffortRollup assertions also retain row/token/cache/peak/health behavior. Its legacy numeric-zero versus empty case protects the independent in-memory compatibility policy.
- TestAggregateMetricsOverTime_Hourly supplies unordered complete tiny, partial tiny, unavailable and observed-zero rows plus a code-only bucket. It protects chronological cumulative precision, continued incompleteness after missing evidence, counts, code-only empty evidence and non-cost code sums.

All original meaningful domain assertions remain. These are source-value oracles through real derivation/aggregation paths, not copies of the implementation formula. There is no reason to add a duplicate matrix merely because the tester is a separate role. Fleet/role rounded compatibility caches intentionally remain distinct from source-precision summaries; no new rounding policy is introduced.

## Backend removal/transfer audit

Compared baseline and current bodies and inspected actual emitted-output recipients. The only deleted test declarations belong to the explicitly approved dead DTOs; the event repository owner is renamed without dropping its persistence/read body.

| Removed or replaced assertion | Actual disposition and recipient |
| --- | --- |
| Computed metrics reverse anomaly roundtrip | TestAgenticMetricsLedgerDTO_AnomaliesSerialization now asserts emitted/parsed rule, severity, tool, turn and count. Original omitted-presentation and omitempty guards remain. |
| Unsupported legacy computed-input/default and malformed computed-input ingestion | Retired with the unused reverse graph; output DTO unmarshalling was not a production ingestion contract. Malformed anomaly collection rejection now reaches the actual fact decoder in TestSchemaMigrationDTOBoundaries. Existing fact/API negatives remain. |
| Computed role/model reverse reconstruction | TestAgenticMetricsLedgerDTO_DynamicRolesAndModelsSerialization checks emitted parsed counts, custom names and order directly. |
| Computed reverse ticket type | TestTicketTypeBoundarySchemaAndSerialization checks actual emitted JSON type strings; all live fact bidirectional/null/omitted cases remain. |
| Dead BrittleToolSummaryDTO fields and stdlib roundtrip | Actual TestResponseHelpersAndSerialization/WriteFleetResponse checks two emitted tools field by field: name, anomaly count, calls, errors and failure rate. Domain arithmetic owners remain separate. |
| Dead FleetAnomalySummaryDTO fields/empty shape | Actual WriteFleetResponse checks total/warning/critical counts, full severity/rule maps, total/affected ticket counts, anomaly rate and nested tool fields. Its empty companion checks zero scalars, explicit zero map entries and non-nil empty tool array. Existing /api/fleet aggregation coverage remains. |
| Dead EventFactDTO custom role | TestCanonicalV4NestedEventsReadCompatibilityAndMetadata now reads security-auditor through a real canonical session event decoder and retains full metadata/owner expectations. Empty-role/status defaults belonged only to the unused root DTO and are intentionally retired. |
| Root Tool/Event DTOs, unreachable raw root collections and dead validateAnomalyEvidence | Existing root events/tools/invocations rejection loop, active persisted anomaly validation negatives, and the actual persistence/read plus unscoped-tool negative body remain. The repository owner is TestSessionEventPersistenceAndRepositoryHandling. Migration's legacy domain fact types and historical behavior are untouched. |
| SerializedLedgerDTO alias | Three variable declarations now use AgenticMetricsLedgerDTO. Their actual CLI/persistence/output assertions remain. No replacement compatibility alias or test-only reverse helper was introduced. |

Production AgenticFactLedgerDTO.ToDomain remains called by repository scan/load/existing-ledger and backfill paths. Live anomaly/context/fact helper conversions remain. Persisted compact facts, richer HTTP facts and compact computed export retain their distinct privacy/shape responsibilities; shared metric DTOs remain reusable. Existing TestSaveFactLedger/TestSaveMetricsLedger and HTTP endpoint/serializer tests retain field order, null/partial/count values, export facts/threads exclusions, source hydration baseline and nil/empty behavior.

Read the moved named invocation/task/root decoder mechanics. The small helper creates a strict decoder, checks EOF and probes only named cost_usd/cost_evidence presence. Owning concrete codecs retain context, strict unknown-field behavior and nullable-versus-omitted handling; v4 optional versus v5 required policy remains explicit. Migration stays separately tolerant of historic unknown fields and uses only shared pure domain validation. No generic target decoder or object-kind guessing was introduced.

## Frontend endpoint validation audit

Read all changed client test bodies, the five named endpoint readers and their client.ts call sites. The coder's actual-entrypoint additions protect both demonstrated failures: valid dynamic role/model keys named cost_usd/cost_summary are accepted as data, while v5 task cost_usd null without own cost_evidence is rejected. Matching summary-map entries are still validated, including a deliberate contradictory dynamic-name entry.

Existing and extended client cases cover v5 invocation/task own-member presence, explicit null and contradictions; observed zero/partial/unavailable task success; v4 numeric zero/null/omitted invocation and task compatibility; opaque source-event/artifact objects and role-effort map keys; all three ledger-array client functions and nested thread/model amounts; comparison row coverage/counts and array shape; bucket versus cumulative coverage and array shape. Each named reader is selected by a real endpoint client rather than property-name inference. Dynamic map keys and unrelated source/artifact data are not recursively classified.

All 100 API client cases passed in the coder's final handoff. The unchanged compact cards, filtered average and view tests remain their existing recipients; this iteration does not change those product behaviors. No extra source-text, copied-validator or generic edge matrix was added.

## Verification and scope limits

Reused final coder results recorded in AGGREGATION.md, BACKEND.md and the frontend handoff, after confirming the corresponding production/test source had not been changed by the tester:

- Domain coder: focused aggregation owner checks, go test -race ./internal/domain/... and go vet ./internal/domain/... passed after final additions.
- Backend coder: go test ./internal/adapters/... ./cmd/agentic-lens, matching race suites, go vet ./... and git diff --check passed after moves/transfers.
- Frontend coder: all 100 API client tests, typecheck, lint and canonical build/embed checks passed; no tests silently skipped.
- Tester: git diff --check passed. Targeted exact-symbol search confirms no retired DTO/alias/dead reverse helper names remain in Go code; supported SessionToolFactDTO/SessionEventFactDTO are intentionally retained. This is an audit aid, not a source-text unit test.

Current audited backend/web source-tree SHA256: a14d1909f8693e0521b58a72b8d26ae39e600959a6488439960af929f57f2aca across 230 existing files. Method: sorted unique git ls-files -co --exclude-standard paths under backend/web, hash each existing path, NUL, bytes, NUL; ignored assets/dependencies and deleted files excluded. This identifies the audited source, not a claim that a previous iteration's full make covered it.

No production defect or meaningful assertion loss was found. No additional suite was rerun solely for role duplication. Parent must run the final integrated make and embedded/standalone verification after review/integration and records final telemetry/commits/PR delivery. No browser geometry, new performance benchmark, raw transcript ingestion, live workspace fixture or delivery action is claimed by this report.

## Final frozen-source note

After the earlier a14d1909f8693e0521b58a72b8d26ae39e600959a6488439960af929f57f2aca snapshot, the backend removed two zero-call unexported anomaly evidence-reference helpers identified by final review. Retain that earlier snapshot as distinct evidence; it is not the final source hash.

The final frozen backend/web source independently recomputes to f49d9f8a69b019dbeba07fa2e5891c1da6509a40683197bc6bfffed4c02998ad across 230 existing files, using the same sorted existing Git-path/path-NUL-bytes-NUL method above. Backend focused adapter, HTTP and backfill checks passed after the deletion, as reported in the handoff. Removing unreachable helpers requires no new test or duplicate suite run. The parent runs the final integrated make on this latest source after integration; earlier verification is not represented as testing an identical snapshot.
