# REV-5 backend implementation and test mapping

Baseline: e8d9003ea6388d1a750273478ec572863e965a72. Ownership: backend/cmd/** and backend/internal/adapters/** only; domain changes belong to DOMAIN.md. Shared checkout changes preserved. No commits, telemetry generation or publishing by this child.

## F1: Configuration transaction

writeConfigPair validates both targets before staging or moving originals. Cleanup is registered before the first staged file. One deferred finalizer rolls back replacements and restores originals on every pre-commit exit, in reverse order; errors from cleanup/restoration are joined with the initiating error with path context. Successful installation cleans backups and reports cleanup failures. Staging closes the file even if writing fails and preserves both errors.

TestInitCommand's existing local-write-failure case now covers both absent and existing shared configurations, exact existing bytes, original path usability, no partial newly created shared file, no false success, local target context, and no staged/backup files. New TestWriteConfigPairCollisionAndForce transfers the removed single-file writer's meaningful typed collision/path, unchanged-original, force replacement, and successful initial write checks to the actual two-file implementation; both files and cleanup are asserted. Existing CLI force, missing flags, output-directory and permission cases remain. No global failure-injection interface introduced. Late OS restore failure is surfaced but not deterministically injected; no claim of all filesystem failure modes tested.

## F2: Unread trace buffer

Removed live/cache.go (TraceBuffer, NewTraceBuffer, SetMessages, AppendMessages, GetMessages, Clear, Len), LiveService.buffer, constructor allocation, Buffer accessor, and source-read SetMessages. Removed unreferenced Hook accessor. traceCache, revision handling, live messages and subscribers remain. MaxTraceBufferMessages stays accepted compatibility configuration; domain comment and generated YAML comment describe it as legacy rather than implying it truncates full responses.

Removed TestTransientMemoryBufferBoundary, TestTraceBuffer_NegativeCapacityDefaultsSafely and TestTraceBuffer_AdversarialConcurrentReadWrite because their sole oracle was the removed unused cache. Removed only the Buffer().Len assertion from TestZeroDiskPersistenceVerification; source trace and no-write assertions remain. TestConcurrentReadsAndSubscriptions, TestLiveTraceSnapshotReuseAndInPlaceInvalidation and TestLiveTraceLegacyRevisionRecoveryAndBoundedOnDemandCache still exercise actual live state.

## F4: Obsolete merge implementation

Removed mergeTools, mergeEvents, toolFactKey and eventKey. Removed TestMergeToolsAndEvents_UnitAdversarial, whose flatten/dedup algorithms have no runtime caller and do not match owner-snapshot replacement. Preserved TestMergeSessionsReplacesRefreshedOwnerSnapshot (replacement, idempotence, unavailable owners) and TestMergeSessionsPreservesRowMultiplicityAndOwnership (legitimate repeated rows and ownership). No dead merge algorithm copied into tests.

## F5 and F6: Adapter candidate disposition

- ExtractFacts removed. All schema-v1/v2, schema-family error and extractor test callers now invoke ExtractSessions directly. Tests project session-owned facts using existing AgenticFactLedger.InvocationFacts/ToolFacts/EventFacts; no copied sorting or extraction logic. Original SQLite fixtures and token, invocation, user, tool, event, timestamp, malformed/schema, injection and nil/empty assertions remain. Live parity test also calls ExtractSessions. Test names now identify active extraction API.
- Transcript.ToMetricsLedger removed. Rollout metadata/usage/hierarchy and Codex retention tests now exercise SessionFacts, DeriveMetricsLedger and PersistAnomalySummaries directly with explicit fixture metadata. Existing privacy markers, context-versus-cumulative accounting, provenance, anomaly-summary retention and omitted-persisted-events assertions remain. No copy of old optional-metadata normalization.
- NewServer removed; production CLI already uses NewServerWithConfig. HTTP test setup now calls the active constructor with loaded isolated fixture config or explicit default fixture configuration; it does not fall back to the developer checkout's config. TestConfigHandler_DefaultConfig supplies explicit fixture name/prefixes/roles. Live handler tests use explicit defaults. TestNewServer_MissingConfigFailsFast removed as a contract of the deleted discovery constructor; TestServer_FailsWithoutWorkspacesConfig, TestCLI_RootConfigurationDiscovery and TestCLIServerConfigurationResolution remain on the actual CLI configuration path. Workspace synchronization assertions still run against NewServerWithConfig.
- GenerateDefaultConfigYAML removed; TestGenerateDefaultConfigYAML removed because it compared the passthrough's result to its delegate. TestGenerateDefaultYAML_Overrides, edge cases and validation plus config YAML roundtrip tests remain.
- WriteConfig and newly unreachable WriteConfigFile removed. TestWriteConfigFile_CollisionAndForce assertions transferred to TestWriteConfigPairCollisionAndForce as above. TestWriteConfigFile_PermissionHandlingAndInvalidPath removed (dead single-file writer and conditional root-dependent oracle); actual init directory/permission cases remain in TestInitCommand.
- FleetAnomalySummaryDTOFromDomain, FleetAnomalySummaryDTO, BrittleToolSummaryDTO and its conversion subtree removed. Their only callers were TestFleetAnomalySummaryDTO_BidirectionalMapping and TestBrittleToolSummaryDTO_BidirectionalMapping, also removed. Actual HTTP routes serialize domain fleet/brittle data, not these DTOs. TestWorkspaceEventsAndAnomaliesHTTP retains actual fleet anomaly response assertions; domain fleet aggregate coverage remains. Active AgenticMetricsLedgerDTO/SerializeLedger and canonical/event compatibility DTOs remain.
- EventFactDTOFromDomain, containsString, validateAnomalyEvidence, ValidateToolFacts, evidenceRefsToDomain and evidenceRefsFromDomain removed: no production or test references. Dormant evidence validation is deleted, not activated. Existing accepted schema and codec behavior unchanged.
- FormatISODatetime, FormatCurrency, FormatPercentage, FormatDuration, FormatCodeDelivery, FormatTicketLabel and newly unreachable FormatCompactNumber removed: no callers/tests. Live markdown formatters remain, including FormatNumber.
- LoggingMiddleware, WriteLiveAnomalyAlertResponse and WriteLiveAnomalyAlertsResponse removed: no callers/tests. Actual CORS, SSE and live response paths remain.
- UnmarshalJSON and UnmarshalYAML retained: encoding/json and YAML decoder entry points are implicit runtime callers; identifier-only inventory misses them. These are the only remaining zero-explicit-reference function names in the rerun Go inventory.

## EventsFromLiveTrace test coverage transferred to actual source path

Domain owner removed obsolete converter and its isolated tests. Existing TestHydrateWorkspaceEvents_SuccessAndZeroDiskWrites exercises actual SQLite hydration of turn_message/tool/compaction, prompt/reasoning, context, tool input/output and no writes. Added session/turn provenance on every resulting event, source message identity on the hydrated turn's EventID, and owning MessageID on tool events. Actual source-backed turn events use EventID and may have nil MessageID; the deleted converter's different turn MessageID convention is not imposed on production. Persisted baseline events with absent optional identifiers remain accepted.

TestHydrateWorkspaceEvents_ConsecutiveErrorsSpansTurns and TestHydrateWorkspaceEvents_MultiSessionIsolationAndStreakReset retain error classification and ownership boundaries. TestFetchSessionTrace_MalformedJSONPartsFallback and TestLiveService_CorruptedJSONAndMalformedPayloads preserve malformed/unknown optional source handling. TestLiveService_TokenExtractionAndAnomalyDetection retains context and anomaly behavior. Converter-specific empty LiveTraceView handling has no remaining API contract and was removed rather than recreated in tests.

## Verification

- Go AST scan rerun after cleanup: only implicit UnmarshalJSON and UnmarshalYAML methods remain with zero production identifier references. This is a conservative identifier screen, not a complete call graph or dead-type proof.
- go test -race ./cmd/agentic-lens ./internal/adapters/... passed across all adapter packages and CLI. Existing generated embed assets available; no manual source asset changes.
- Final focused CLI/HTTP/OpenCode race rerun covers the expanded absent/existing configuration regression and test naming updates; result communicated to parent.
- Intermediate compile/test failures caught unused imports, one missed parity caller, checkout-dependent HTTP fixture defaults, and an overstrict MessageID assertion; all corrected before handoff. No silent skips added, no domain I/O, no any/interface{}, no package dependency added.
- Parent owns tester/reviewer audit and final integrated make gate, standalone reproduction and telemetry.
