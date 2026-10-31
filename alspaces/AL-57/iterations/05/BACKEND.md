# AL-57 iteration 05 backend implementation

Baseline: 412bc08e3199af04c03009ed60ca1bd3b592dd85. User approved the full iteration plan. This coder owned adapters, necessary test updates and the three retired-alias test declarations in the CLI. Domain aggregation/cost and frontend changes belong to their separate owners and were preserved.

## Finding dispositions

| Concern | Delivered disposition |
| --- | --- |
| 2. Mixed DTO responsibilities | Removed cost_dto.go and schemas.go. Shared nullable Grouped/Role/Model/Workflow projections are in metric_projection_dto.go; compact computed export is in metrics_export_dto.go. HTTP-only workspace/session/fact/thread, fleet, comparison and progression shapes and converters moved into http/workspace_dto.go, fleet_dto.go, comparison_dto.go and progression_dto.go. No facade aliases or adapter-to-HTTP dependency were added. Existing persisted compact facts and richer admitted HTTP facts remain separate. |
| 3. Fragmented fact codecs | Invocation tokens/facts, codecs and both live conversions now colocate in invocation_fact_dto.go; task links in task_link_fact_dto.go; session/tool/event facts and conversion responsibilities in session_fact_dto.go; root/code/migration facts and validation/conversions in fact_ledger_dto.go. Anomaly/context/evidence live bidirectional mappings are in anomaly_dto.go. Small shared fact-cost helpers remain in fact_cost.go. Removed cost_fact_dto.go. |
| 4. Repeated local projection work | Invocation forward/reverse converters compute effective evidence once. Fleet converter computes effective role/model summaries once and reuses them for maps and companion summaries. Progression converter computes cumulative summary once. Domain financial precision and compatibility policy are unchanged. |
| 5. Unused reverse/dead boundary code | Removed computed metrics ToDomain and its exclusively reachable Grouped/Role/Model/Workflow reverse methods/slice helpers, costAccumulator and roleEffortAccumulators. Removed SerializedLedgerDTO alias; its three tests use AgenticMetricsLedgerDTO. Removed unused FleetAnomaly/BrittleTool DTOs and conversions, root Tool/Event DTOs and conversions, unreachable rawExistingLedger root collections, and unused validateAnomalyEvidence. Kept production FactLedger.ToDomain, session conversion, fact/anomaly helpers, migration legacy domain types, current runtime root-collection rejection and legacy loader fallback policy. |
| 6. Repeated decoder mechanics | fact_decode_helpers.go owns only configured strict Decoder creation, contextual single-value EOF check and fixed cost member presence probe. Concrete invocation/task aliases retain their owning decode policies. Named v4/v5 root entry points select required evidence explicitly and share the identical canonical root decode body. No reflection, any, generic target decoder, arbitrary recursion or object-kind guessing was added. |

Struct declarations and forward assignments were moved intact except the local caches above. Existing nullable values, field order, nil/empty collection behavior and compact export's exclusion of facts/threads remain guarded by actual serialization tests. Persisted session output still compacts events; HTTP still serves admitted event metadata. No rounding or financial representation change was introduced by this coder.

## Removed assertions and actual recipients

| Removed source assertion | Disposition / recipient |
| --- | --- |
| TestAgenticMetricsLedgerDTO_AnomaliesSerialization computed reverse roundtrip | Same test now inspects actual emitted/parsed anomaly rule, severity, tool, turn and count. Omitted presentation text and empty-anomaly omitempty checks remain. |
| Same test's unsupported legacy computed-input and malformed computed-input roundtrips | Retired because the DTO has no production ingestion consumer. Malformed anomaly collection rejection is asserted at actual fact decoding in TestSchemaMigrationDTOBoundaries/fact ledger rejects malformed anomaly collection. Existing malformed fact/API owners remain. |
| TestAgenticMetricsLedgerDTO_DynamicRolesAndModelsSerialization reverse reconstruction | Same test directly asserts emitted parsed role/model counts, custom names and their order. No reverse test helper replaces production code. |
| TestTicketTypeBoundarySchemaAndSerialization computed reverse ticket-type assertion | Same case marshals the computed export and checks its emitted ticket_type string. All live fact bidirectional/null/omission assertions remain. |
| TestBrittleToolSummaryDTO_BidirectionalMapping | Dead DTO assignment/standard-library roundtrip retired. Actual HTTP TestResponseHelpersAndSerialization/WriteFleetResponse writes valid JSON checks both emitted tool records' name, anomaly count, calls, errors and failure rate field by field. |
| TestFleetAnomalySummaryDTO_BidirectionalMapping | Actual HTTP WriteFleetResponse case now checks emitted total/warning/critical counts, full severity/rule maps, ticket counts, anomaly rate and nested brittle tool fields. Its empty companion checks zero scalar values, explicit zero map entries and non-nil empty tool array. Existing domain arithmetic and end-to-end /api/fleet owners remain unchanged. |
| TestEventFactDTO_DynamicRoles custom-role assertion | TestCanonicalV4NestedEventsReadCompatibilityAndMetadata now reads security-auditor from the actual canonical session event decoder and asserts the full resulting event metadata/owner. Removed empty-role/status defaults belonged only to the dead unsupported root DTO. |
| Retired root DTO fields / alias types | Existing canonical root events/tools/invocations rejection loop retained. Repository's live persistence/read and legacy-unscoped-tools negative assertions retained under the responsibility name TestSessionEventPersistenceAndRepositoryHandling. Three alias declaration updates retain actual CLI and persistence output checks. |

## Verification

Passed against the integrated checkout after these moves and transfers:

- From backend: go test ./internal/adapters/... ./cmd/agentic-lens.
- From backend: go test -race ./internal/adapters/... ./cmd/agentic-lens.
- From backend: `go vet ./...`.
- Root: git diff --check.
- Targeted symbol/caller search found no remaining retired alias, root Tool/Event DTO, computed reverse subtree or validateAnomalyEvidence declarations/callers.

Current embedded assets were already available; no build stage was silently skipped by these direct Go commands. Independent tester owns subsequent test changes and audit. Parent owns integrated make, review, telemetry, commits and PR delivery. This coder performed no live ledger rewrite, telemetry extraction, commit or push.

## Reviewer correction

Verified with repository-wide backend symbol search that evidenceRefsToDomain and evidenceRefsFromDomain had definitions only and no callers. Deleted those two unexported unreachable helpers from anomaly_dto.go. AnomalyEvidenceRefDTO, live AnomalyDTO/AnomalyContextDTO conversions and the fact/anomaly ingestion chain remain unchanged. No tests were added or edited for unreachable helper deletion. Focused adapter tests and whitespace checks passed after removal.
