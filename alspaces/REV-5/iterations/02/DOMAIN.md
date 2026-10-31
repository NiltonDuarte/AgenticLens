# REV-5 domain remediation

Baseline: e8d9003ea6388d1a750273478ec572863e965a72. Ownership: backend/internal/domain/** only. All 15 domain candidates in iteration 01 are removed after repository reference inspection. Three dependencies that became unreachable are also removed. No production algorithm, wire field, enum, or codec behavior was changed. MaxTraceBufferMessages remains accepted and is documented as legacy compatibility; it is not repurposed to truncate live responses.

## Candidate and test dispositions

| Removed production symbol | Preserved assertion owner or deletion rationale |
| --- | --- |
| ComputeRoleMetrics | Existing TestComputeRoleAndModelMetrics, TestComputeRoleMetrics_DynamicOrdering and TestComputeRoleMetrics_AdversarialEdgeCases now call ComputeRoleMetricsWithCeiling with the previous default ceiling. Numerical results, role ordering, empty input and custom order assertions retained. |
| ComputeThreadSummariesFromSessions; newly unreachable ComputeThreadSummaries | Existing thread-summary tests, empty-thread subtest and session_facts_test hierarchy/multiplicity assertions now use ComputeThreadSummariesWithCeiling with the same ceiling. |
| ComputeWorkflowEfficiencyWithConfig | Existing configured critical-ceiling assertion passes cfg.EffectiveHealthyCeiling and cfg.RoleOrder directly to the production ComputeWorkflowEfficiencyWithCeiling API. |
| ClassifyToolFailure | TestClassifyToolExecution_MessageTruncation preserves failure, severity, message presence and exact 2048-character truncation assertions against the typed production result. |
| IsValidTicketType | TestNormalizeTicketType_AcceptedAndRejectedSpellings retains every valid/invalid alias, whitespace and case input using NormalizeTicketType's returned validity. |
| FilterLedgersByTicketType; FilterLedgersByPrefix | TestFilterLedgersByTicketTypes_AllAndAliases and TestFilterLedgersByPrefixes_AllAndMultiple retain selection and empty/all assertions against plural APIs. Comma-string parsing was only the obsolete wrapper contract; live HTTP parameter parsing remains independently tested by TestMultiFilterEndpoints and its adversarial peers. |
| FilterLedgersByDateRange | Renamed TestFilterLedgersByCompletionDate_MissingCompletion exercises the actual filter with original selection and missing/empty/whitespace/corrupt completion cases. |
| FilterLedgersByTimeWindow | Original range fixtures/assertions preserved in TestFilterLedgersByCompletionDate_Ranges, TestFilterLedgersByCompletionDate_ExactBoundaries and TestAdversarialDomainAggregations using explicit bounds. Includes inclusive start/end, just-outside boundaries, empty inputs and no fallback to started/updated timestamps. No preset-dispatch algorithm was copied into a test helper. |
| GetTicketEffectiveDate; parseISODate (newly unreachable) | TestGetLedgerCompletionDate_NoFallback retains nil, empty, whitespace, corrupt calendar and no-fallback guards on the active timestamp parser. The removed helper's midnight truncation is not the production contract: expected valid result now preserves the original full timestamp. |
| ResolveDatePreset | TestResolveDatePreset deleted: tested an unused alternate preset vocabulary, including midnight and unknown-input behavior not used by HTTP. Active ParseTimeWindow tests remain, as do TestProgressionHandler and TestMultiFilterEndpoints exercising live temporal requests. No claim that dead preset semantics are preserved. |
| CalculateRoleEffortRollup | Only test was an empty-map assertion in TestFleetSummaryAndComparisonEdgeCases. Deleted that obsolete-calculation assertion; existing fleet zero-value and real aggregation assertions remain. No runtime aggregate uses this calculation. |
| ExtractDistinctPrefixes | TestExtractDistinctPrefixes and its nil/empty/pathological-ID assertions in TestMultiFilter_AdversarialEdgeCases deleted. Uniqueness/sorting of a returned distinct-prefix list is obsolete behavior with no runtime producer. TestExtractTicketPrefix and plural-filter pathological-ID/selection assertions remain for active extraction/filter behavior. |
| GetStep | TestRegistry now verifies registration and exact step identity through PlanChain(1, 2). Its concurrent registration/chain-resolution test remains; duplicate unused direct lookup removed. |
| MigrateLedger | Existing serialized-ledger subtests now invoke Engine.Migrate using a JSON-marshaling fixture helper; upgrade, invalid target, same-version and downgrade assertions retained at the actual production boundary. Helper performs serialization only, no migration logic. |
| EventsFromLiveTrace | Removed TestEventsFromLiveTrace with the unused conversion. Active source-backed event coverage is owned by adapter tests; see coordination below. Nil LiveTraceView conversion, whitespace normalization and caller-provided partial trace behavior are obsolete helper-specific contracts and are not reproduced in tests. |

## Event-conversion coverage coordination

Backend owner identified TestHydrateWorkspaceEvents_SuccessAndZeroDiskWrites, TestHydrateWorkspaceEvents_ConsecutiveErrorsSpansTurns, TestHydrateWorkspaceEvents_MultiSessionIsolationAndStreakReset, TestFetchSessionTrace_MalformedJSONPartsFallback and TestLiveService_TokenExtractionAndAnomalyDetection as active recipients for event kinds, turns, session ownership, compaction, status, source fallback and context. Backend owner is inspecting/strengthening message provenance and reasoning assertions in that layer; BACKEND.md records final evidence. No adapter files were changed by this agent.

## Verification

Focused go test -race ./internal/domain/... passed after removals and assertion retargeting. Earlier compile feedback caught one remaining session_facts_test reference and a missing test JSON import; corrected before final run. Post-cleanup AST inventory has no remaining zero-production-reference named functions in domain packages. This is conservative identifier screening, not a whole-program liveness proof. git diff --check passed for domain paths. Parent owns final integrated gate and independent tester/reviewer assessment.
