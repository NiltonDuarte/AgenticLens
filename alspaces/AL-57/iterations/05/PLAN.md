# AL-57 iteration 05: Explicit validation and focused DTO cleanup

Status: APPROVED for implementation. After reviewing the concrete plan, the user explicitly selected "Approve the full plan". This authorizes the six bounded changes below without another confirmation gate. Baseline: 412bc08e3199af04c03009ed60ca1bd3b592dd85, initially clean. Existing PR #89 remains the delivery PR. Frozen CONTRACT.md stays unchanged.

## Steering and preserved acceptance

Source: the latest user request identifies frontend validation, DTO responsibility, fact-codec organization, duplicate cost calculation, unused computed reverse conversion and repeated decoding mechanics. Changed boundary: investigate and propose concrete corrections, then stop for approval. Findings below are proposed fixes or justified retention/defer decisions; none is claimed fixed.

Preserve iteration 04's average over complete-cost tickets within the current filtered population, all accepted cost/partial/coverage semantics, valid historical numeric zero as observed, read-only canonical-v4 compatibility, strict v5 cost presence/evidence, v1-v3 explicit migration, source/session ownership, repeated row multiplicity, generated field order and optional-hydration baseline visibility. No general serializer/validator framework, reflection, arbitrary-object rewriting or new financial representation is approved by this planning request.

## 1. Replace recursive frontend cost inference

Disposition: PROPOSE FIX. web/src/api/costValidation.ts validateCostPayload recurses into every object and guesses financial object kinds from property names. It incorrectly interprets dynamic role/model names as structural cost fields and fails to require cost_evidence on a v5 task.

Observed through actual client entry points using an in-memory esbuild bundle of client.ts and existing mockData fixtures, with no repository file writes:

- fetchFleetSummary with valid numeric model_costs keys named cost_usd and cost_summary rejects with “Invalid cost payload at response.model_costs.cost_summary”. Model names are data, not schema selectors.
- fetchWorkspaceById with facts.schema_version 5 and a task_links row containing cost_usd null but no cost_evidence is accepted. That violates the accepted v5 fact contract.

Replace the generic entry point with endpoint-selected readers/validators for fleet summary, workspace ledger arrays, workspace detail, comparison arrays and time-bucket arrays. Proposed costValidation.ts exports are readFleetSummaryResponse, readMetricsLedgerArrayResponse, readWorkspaceDetailResponse, readComparisonRowsResponse and readTimeBucketsResponse; client.ts selects the reader from the endpoint, not JSON field coincidences. Keep small shared amount/evidence/summary guards, and explicit known-schema helpers for metric rows, workflow, session/thread projections and fact ledgers. If the file becomes unwieldy, split the named fact and summary guards by responsibility; do not create an extensible validation engine.

At the fact-ledger node, schema_version selects the invocation/task policy. Validate each declared sessions[].invocations[] and task_links[] location directly: v5 requires own cost_usd and cost_evidence members, valid evidence objects and consistent nullable amounts; v4 retains accepted optional legacy presence. Dynamic model_costs/role_costs and companion summary maps iterate opaque keys and validate their values against matching entries only. Never pass those entries back through a generic object classifier. Unrelated source event payloads and free-form artifact content remain opaque. Cost-like model/role names and user metadata must not change interpretation.

Keep existing supported older numeric response compatibility where already accepted; do not silently require newly introduced aggregate summaries from every legacy mock/consumer. Strictness comes from the known endpoint/fact schema policy. Invalid known financial data remains an explicit error, never a zero fallback.

## 2. Separate DTOs by actual responsibility

Disposition: PROPOSE FIX the catchall ownership, RETAIN JUSTIFIED wire shapes that serve different contracts. Current cost_dto.go mixes shared rollups, HTTP-only payloads and reverse helpers; schemas.go mixes persisted facts, codecs, anomaly mapping and computed exports.

| Proposed owner | Definitions and conversions |
| --- | --- |
| adapters/metric_projection_dto.go | Shared GroupedMetricDTO, RoleMetricDTO, ModelMetricDTO, WorkflowEfficiencyDTO and forward converters/slice helpers; roleEffortAmounts |
| adapters/metrics_export_dto.go | AgenticMetricsLedgerDTO and SerializeLedger; remove the unused compatibility alias SerializedLedgerDTO and use the canonical DTO in its three test declarations |
| adapters/http/workspace_dto.go | HTTP SessionFactResponseDTO, FactLedgerResponseDTO, MetricsLedgerResponseDTO, ThreadRoleMetricDTO, ThreadExecutionRecordDTO and their forward converters |
| adapters/http/fleet_dto.go | FleetSummaryDTO, forward conversion and HTTP-owned cost-map helpers |
| adapters/http/comparison_dto.go | TicketComparisonRowDTO and forward collection conversion |
| adapters/http/progression_dto.go | TimeBucketMetricDTO and forward collection conversion |
| adapters/anomaly_dto.go | AnomalyDTO, AnomalyContextDTO, AnomalyEvidenceRefDTO, required bidirectional mapping and retired-field read compatibility; keep other live anomaly boundary definitions/converters with their actual callers |

The HTTP package imports shared adapters/domain types. Adapters must not import HTTP or retain forwarding definitions that recreate the catchall. Update http/dto.go's ResponsePayload/writers and handlers to the local HTTP projections. No change to wire keys, order, nulls, nil/empty behavior, endpoint response families or export fields is intended.

Separate persisted and HTTP fact/session envelopes are justified: persisted output strips source content/presentation evidence, while HTTP serves admitted richer events/anomalies. HTTP metrics include facts and threads; computed CLI/export metrics intentionally omit them. Shared metric projections should be reused, but merging these distinct root/session envelopes would change privacy or public output. Domain scalar models versus nullable wire DTOs are also deliberate layer separation, not duplicate financial authority.

## 3. Colocate fact definitions, codecs and conversions

Disposition: PROPOSE FIX file-level fragmentation without removing live decoding.

- adapters/invocation_fact_dto.go owns InvocationTokensDTO and InvocationFactDTO, their named UnmarshalJSON, both conversion directions and invocation slice conversion. adapters/task_link_fact_dto.go owns TaskLinkFactDTO and its codec/conversions/slice helpers. Shared known-cost evidence/accumulator helpers stay in a small fact-cost helper file, not another mixed DTO catchall. This reunites definitions currently split between schemas.go and cost_fact_dto.go.
- adapters/session_fact_dto.go owns SessionFactDTO, SessionToolFactDTO and SessionEventFactDTO with their retired-field read codecs and session conversion responsibilities.
- adapters/fact_ledger_dto.go owns AgenticFactLedgerDTO, CodeFactsDTO, MigrationInfoDTO, root version dispatcher/validation, both root conversions and ticket-type helpers. fact_ledger_decode.go retains the named v4/v5 boundary entry points.

Retain AgenticFactLedgerDTO.ToDomain and its required fact/anomaly helper chain. Production callers include repository.go's scan/load/existing-ledger paths and backfill/engine.go; these are accepted persistence/read boundaries. This cleanup must not remove them merely because computed reverse helpers are unused. Record each move once, remove emptied catchall files, and preserve public fact symbol names unless an actual boundary change is necessary.

## 4. Reduce redundant cost work without changing compatibility precision

Disposition: PROPOSE FIX redundant local work; RETAIN JUSTIFIED rounded compatibility projections; DEFER replacing the entire scalar-plus-summary model because that broader representation redesign is not needed for these concerns and would alter existing callers/rounding.

In domain/aggregations.go accumulateInvocations, merge SummaryForCost into the cost summary during the existing token loop instead of independently summing a scalar and then traversing invocations again through SummarizeCost. Derive the existing rounded scalar from the same summary subtotal once. ComputeTotals, ComputeRoleMetricsWithCeiling and ComputeModelMetrics keep their accepted outputs. Reuse this accumulator in ComputeThreadSummariesWithCeiling's duplicated per-model loop, preserving ceiling health and precision-aware cost ordering. Remove bucketAccumulator.costUSD's redundant summary-subtotal cache and derive bucket/cumulative scalars from the same ordered summary merges with unchanged four-decimal projection.

Retain CalculateFleetSummary's totalCostUSD/roleCosts/modelCosts and CalculateRoleEffortRollup's scalar accumulation: these intentionally sum already-rounded per-ledger scalar values. Two ledger source amounts of 0.00004 each produce scalar ledger values 0 and fleet scalar 0, while their raw merged summary subtotal is 0.00008; rounding the raw merged summary instead would produce 0.0001. Wire complete amounts continue to derive from summaries. Do not silently change the legacy domain projection to make duplication disappear.

Compute EffectiveCost/evidence once per conversion where currently repeated in InvocationFactDTOFromDomain/ToDomain, and compute effective role/model/cumulative summaries once in FleetSummaryDTOFromDomain and TimeBucketMetricDTOFromDomain. Retain EffectiveCost/EffectiveCostSummary at independent in-memory legacy construction, migration, domain rollup/sort and output boundaries; deleting these compatibility responsibilities requires a separately approved representation change. Presence-aware wire decoding remains independent of in-memory fallback and must run before scalar defaults.

## 5. Remove computed reverse conversions with no production consumer

Disposition: PROPOSE FIX. Read-only caller audit found AgenticMetricsLedgerDTO.ToDomain invoked only by four schemas_test.go cases, not production. Remove that method and its exclusively reachable subtree: GroupedMetricDTO.ToDomain, RoleMetricDTO.ToDomain/RoleMetricDTOsToDomain, ModelMetricDTO.ToDomain/ModelMetricDTOsToDomain, WorkflowEfficiencyDTO.ToDomain, costAccumulator and roleEffortAccumulators. Do not remove shared helpers with live fact callers, including AnomalyDTO/AnomalyContextDTO.ToDomain and toDomainTicketType.

Meaningful assertion transfer is explicit:

- schemas_test.go's anomaly serialization case near line 450 keeps emitted rule/severity/context, omitted presentation strings and omitempty assertions against actual JSON/parsed DTO. Remove the computed reverse roundtrip, unsupported legacy computed-ingestion promise and malformed computed-input checks: ordinary Go unmarshalling into an output DTO does not create a supported ingestion contract. Preserve negative malformed-type guards at actual fact/API read boundaries. Expected values must be asserted against emitted JSON, not a replacement reverse converter.
- Its custom role/model case near line 980 asserts the parsed emitted roles/models, names and order directly, rather than reconstructing domain metrics solely to inspect them.
- TestTicketTypeBoundarySchemaAndSerialization retains emitted computed ticket-type strings and every live fact bidirectional assertion. Remove only the computed reverse ticket-type assertion near line 1118.
- Existing TestSaveMetricsLedger and HTTP TestResponseHelpersAndSerialization remain the production boundary owners for nullable totals, payload fields, order, facts/export exclusions and partial subtotals. No replacement test-only reverse helper is introduced.

The follow-up full caller audit also proves SerializedLedgerDTO has only three test variable declarations, in main_test.go near line 464 and repository_test.go near lines 1811/1956. SerializeLedger already returns AgenticMetricsLedgerDTO; historical AL-1 review text documents a test usage, not a consumer contract. Remove the alias, change those declarations to AgenticMetricsLedgerDTO, and preserve their actual CLI/persistence behavior assertions.

Adjacent dead boundary code is a separate, explicitly bounded proposed deletion, not part of the computed reverse graph by assumption. Remove FleetAnomalySummaryDTO/BrittleToolSummaryDTO and both conversion directions: the production fleet payload already embeds domain.FleetAnomalySummary. Replace their dead bidirectional/stdlib-roundtrip tests with missing real-output assertions in HTTP TestResponseHelpersAndSerialization's WriteFleetResponse case, using the existing domain TestCalculateFleetSummary_IncludesAnomalySummary and TestComputeFleetAnomalies_MultipleAnomaliesAndBrittleTools as arithmetic owners. Supply a populated domain summary and assert emitted JSON summary counts, severity/rule maps, ticket/anomaly rate and every nested brittle tool name/count/calls/errors/failure-rate field; retain empty-map/array behavior through an empty variant. TestWorkspaceEventsAndAnomaliesHTTP's /api/fleet branch remains the end-to-end aggregation owner. Do not claim a dead DTO assignment test proves the endpoint.

Remove obsolete adapters ToolFactDTO/EventFactDTO, their unused conversion methods and rawExistingLedger root Invocations/Tools/Events fields: the follow-up caller audit confirms the runtime gate rejects those collections before decode. Adjust any condition referencing a removed field only to remove its proven unreachable operand; broader legacy-loader fallback policy is unchanged. Preserve TestCanonicalV4NestedEventsReadCompatibilityAndMetadata's existing root events/tools/invocations rejection loop and repository event handling's legacy-unscoped-tools negative case; do not add a redundant rejection table. SessionToolFactDTO/SessionEventFactDTO and domain migration legacy ToolFact/EventFact types remain supported. Transfer the custom event-role assertion from TestEventFactDTO_DynamicRoles to actual session read in TestCanonicalV4NestedEventsReadCompatibilityAndMetadata or the existing HTTP event owner TestWorkspaceEventsAndAnomaliesHTTP; retire the obsolete DTO's empty-role/status defaults because no supported runtime decoder calls it. Keep TestEventFactDTOAndRepositoryEventHandling's actual persistence/read assertions, renaming it for the current responsibility rather than deleting it. validateAnomalyEvidence has exactly one definition and zero callers: delete that dead helper with its obsolete EventFactDTO dependency, while preserving active PersistAnomalySummaries/ValidatePersistedEvidence validation paths and their rejection tests. No historic domain migration support is removed.

## 6. Consolidate narrow decode mechanics

Disposition: PROPOSE FIX repeated strict-decoder/trailing/presence plumbing; RETAIN JUSTIFIED explicit version and record policies.

fact_ledger_decode.go currently repeats the same strict root decoder and EOF check in decodeSchemaV4FactLedger and decodeSchemaV5FactLedger; cost_fact_dto.go repeats strict typed invocation/task decoding and the same cost-member presence probe. Add a small adapters/fact_decode_helpers.go for a configured json.Decoder factory, single-value EOF helper and fixed known cost_usd/cost_evidence presence probe. Each owning schema still decodes into its concrete declared type. Share the typed canonical-root body between named v4/v5 wrappers only where their declared root layout is identical, with required-cost policy supplied explicitly. Invocation/task methods retain their concrete alias targets and reuse only these mechanics. Keep contextual errors at the owning entry point. No any/interface{}, generic target framework, reflection or arbitrary object-kind dispatch is permitted.

Do not conflate v4 optional cost members with v5 required members, or required member presence with nullable value. Preserve DisallowUnknownFields and trailing-value rejection at existing strict adapter boundaries. Migration/cost_decode.go's known-schema decodeKnownCost remains separately owned in the pure migration layer and uses shared pure domain.ValidateCostEvidence; do not import adapters, impose current strict unknown-field policy on tolerated historic input, or duplicate the adapter mechanic through an outer dependency just to share lines.

## Verification and approval gate

Extend actual web/src/api/client.test.ts endpoint tests with the two reproduced probes, opaque dynamic role/model names, v5 invocation/task omitted versus null/contradictory fields and invalid counts. Exercise every named reader through client functions, including array/detail/cumulative payloads, rather than testing only guards. Preserve current average filtered-subset and compact card behavior through unchanged view tests.

Backend recipients: schema compatibility/negative-codec cases for omitted/null/unknown/trailing fields and historical zero; TestSaveFactLedger/TestSaveMetricsLedger for actual persistence/export/order; HTTP TestResponseHelpersAndSerialization and endpoint tests for all moved response families, partial nulls, maps and cumulative values; migration/source-ownership and existing refresh-preservation tests for v1-v5 history, repeated rows and unavailable hydration. No test may depend on live ticket files. Existing aggregated acceptance cases remain the oracle, not copied converter implementations.

Add a coherent rounding oracle to TestCalculateFleetSummary and its related role-rollup assertion or a separately scoped case if setup differs; there is currently no dedicated nonempty TestCalculateRoleEffortRollup coverage. Reuse TestComputeTotals, TestComputeRoleAndModelMetrics including tiny-cost thread sorting, and TestAggregateMetricsOverTime_Hourly/AdversarialHourly for consolidated local passes. Map every additional removed meaningful assertion to a real recipient in the review, or retain it.

After approval, run affected Go adapter/HTTP/migration/domain checks and frontend API/typecheck/lint tests during rework, then the required integrated make and canonical frontend/embedded asset/standalone binary verification. Record exact tested source, results and skips. No browser geometry claim is implied by DOM or source checks. Parent retains integration, telemetry, commits and PR delivery.

Approval checkpoint: read-only audits and two reproduced frontend failures were presented first; the user then approved the full plan. All six changes may proceed. Planned moves and historical passing suites remain distinct from completed implementation and its new verification.
