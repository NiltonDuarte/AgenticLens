# Implementation Plan: AL-32 Telemetry Event Pruning, Database Message Stripping, Agent Data Investigation, and Ledger Backfill

## 1. Goal and Iteration Scope

Deliver complete telemetry event pruning, database message stripping, agent structured metadata investigation, and ledger backfill across the AgenticLens Go backend. This iteration eliminates database message retention in the Git plane, preserves bare error indicators, guarantees exact in-memory metric derivation before persistence pruning, introduces schema migration step StepV2ToV3Sanitized (upgrading CurrentSchemaVersion to 3), wires migration into the CLI and backfill engine, and sanitizes historical workspace ledgers to under 5 KB.

---

## 2. Grounded Reference Architecture and Concrete Paths

The implementation maps to the following concrete files and packages:

### Domain Core Layer (backend/internal/domain/)
- backend/internal/domain/constants.go: Promote CurrentSchemaVersion from 2 to 3. Add migration step constant MigrationStepSanitizedEvents.
- backend/internal/domain/events.go: Implement pure function PruneEventsForPersistence(ledger AgenticFactLedger) []EventFact to filter events to bare errors, compactions, and anomaly-associated occurrences, setting ErrorMessage to nil.
- backend/internal/domain/anomalies.go: Tune anomaly evaluation helper detectToolHighFailureRates and detectToolRetryStreaks to ensure deterministic evaluation parity between raw and pruned event representations.
- backend/internal/domain/events_test.go: Add table-driven tests asserting pure event pruning behavior, bare error retention, and nil ErrorMessage enforcement.
- backend/internal/domain/aggregations_test.go: Add table-driven tests asserting that aggregate metrics and derived ledgers maintain bit-identical equality before and after event pruning.

### Schema Migration Engine (backend/internal/domain/migration/)
- backend/internal/domain/migration/step_v2_v3.go: Create StepV2ToV3Sanitized implementing MigrationStep (SourceVersion 2, TargetVersion 3).
- backend/internal/domain/migration/registry.go: Register StepV2ToV3Sanitized in NewEngine default registry.
- backend/internal/domain/migration/validator.go: Update InvariantValidator to validate version 3 specific invariants (asserting SchemaVersion 3, identity preservation, invocation/tool equality, bit-identical derived totals, and zero error message strings).
- backend/internal/domain/migration/engine_test.go: Add unit tests for StepV2ToV3Sanitized and full v1 -> v2 -> v3 chain execution.

### Boundary Adapters Layer (backend/internal/adapters/)
- backend/internal/adapters/opencode/extractor.go: Remove regex text sanitization function sanitizeTelemetryText. Set ErrorMessage to nil on extracted error tool calls intended for persistence.
- backend/internal/adapters/opencode/extractor_test.go: Update extractor unit tests to assert nil ErrorMessage on extracted tool error events.
- backend/internal/adapters/backfill/migration_runner.go: Verify runner executes v2 -> v3 migrations with CurrentSchemaVersion defaulting to 3, supporting dry-run and backup creation.
- backend/internal/adapters/backfill/engine.go: Integrate schema v3 and event pruning into BackfillEngine execution and RecomputeDB workflow.
- backend/internal/adapters/backfill/migration_runner_test.go: Add tests verifying batch migration of historical workspace fixtures reduces file size to under 5 KB.
- backend/internal/adapters/backfill/engine_test.go: Add unit tests verifying BackfillEngine sanitizes events and updates schema version.

### CLI Application Layer (backend/cmd/agentic-lens/)
- backend/cmd/agentic-lens/main.go:
  - In runGenerate, compute metrics and detect anomalies in memory against the raw fact ledger, then prune events using domain.PruneEventsForPersistence before writing JSON to disk via adapters.SaveFactLedger.
  - In runMigrate, default target-version flag to domain.CurrentSchemaVersion (3).
- backend/cmd/agentic-lens/main_test.go: Add CLI integration tests asserting generated JSON ledgers contain zero error_message keys, retain bare error events, and produce valid markdown summary tables.

---

## 3. Phased Implementation Tasks

### Phase 1: Domain Core Schema Version Promotion and Pure Event Pruning
- Task 1.1: Update Schema Constants in backend/internal/domain/constants.go:
  - Update CurrentSchemaVersion constant from 2 to 3.
  - Define MigrationStepSanitizedEvents = "schema_v3_sanitized".
- Task 1.2: Implement PruneEventsForPersistence in backend/internal/domain/events.go:
  - Define function PruneEventsForPersistence(ledger AgenticFactLedger) []EventFact.
  - Evaluate DetectAnomalies(ledger) to identify detected anomalies and gather set of associated sessions, turns, and tools.
  - Iterate ledger.Events:
    - Retain tool call events where Status == "error". Create bare EventFact with ToolName, Status ("error"), Turn, Timestamp, DurationMs, ContextTokens, and ErrorMessage set to nil.
    - Retain compaction events (EventType == EventTypeCompaction) with ErrorMessage set to nil.
    - Retain events matching detected anomalies (matching session, turn, or tool) with ErrorMessage set to nil.
    - Omit routine completed turn messages and routine successful tool calls.
  - Return pruned slice sorted stably via SortEvents.
- Task 1.3: Align Anomaly Evaluation in backend/internal/domain/anomalies.go:
  - In detectToolHighFailureRates, check if ledger.Tools is populated. When ledger.Tools is present, use ledger.Tools for tool call volumes and error counts to guarantee failure rates are calculated accurately even when successful tool calls are pruned from events.
  - In detectToolRetryStreaks, ensure streaks are evaluated correctly across consecutive turn sequences.
- Task 1.4: Add Pure Domain Tests in backend/internal/domain/events_test.go and aggregations_test.go:
  - Add TestPruneEventsForPersistence asserting routine completed turns and successful tool calls are pruned.
  - Assert bare error indicators retain ToolName, Status, Turn, Timestamp, and DurationMs, while ErrorMessage is nil.
  - Assert anomaly-associated events are preserved without ErrorMessage text.
  - Assert ComputeTotals, ComputeWorkflowEfficiency, ComputeRoleMetrics, ComputeModelMetrics, and ComputeThreadSummaries yield bit-identical results when evaluated on unpruned versus pruned ledgers.

### Phase 2: Schema Migration Step V2 to V3 and Invariant Validation
- Task 2.1: Implement StepV2ToV3Sanitized in backend/internal/domain/migration/step_v2_v3.go:
  - Create struct StepV2ToV3Sanitized.
  - SourceVersion() returns 2.
  - TargetVersion() returns 3.
  - Description() returns "Prunes routine completed events and strips database message text from error events while maintaining bit-identical aggregate metrics".
  - Apply(ledger AgenticFactLedger) executes:
    - Target version integer 3 assigned to cloned.SchemaVersion.
    - cloned.Events updated via PruneEventsForPersistence(ledger).
    - Update cloned.Migration: SchemaVersion set to 3, "schema_v3_sanitized" appended to StepsCompleted, LastMigratedAt set to cloned.UpdatedAt, LastMigratedBy set to "universal-migration-engine".
    - Return cloned ledger and nil error.
- Task 2.2: Register Step in backend/internal/domain/migration/registry.go and engine.go:
  - In NewEngine(registry *Registry), register NewStepV2ToV3Sanitized() alongside NewStepV1ToV2() when registry is nil.
- Task 2.3: Update InvariantValidator in backend/internal/domain/migration/validator.go:
  - In Validate(before, after, expectedTargetVersion):
    - When expectedTargetVersion <= 2: preserve legacy v1 -> v2 check requiring len(after.Events) >= len(before.Events).
    - When expectedTargetVersion == 3:
      - Assert after.SchemaVersion equals 3.
      - Assert ticket identification, timestamps, code facts, invocations, and tools match before exactly.
      - Assert len(after.Events) <= len(before.Events).
      - For each event in after.Events, assert event.ErrorMessage == nil.
      - Assert derived metrics (ComputeTotals, ComputeWorkflowEfficiency, ComputeRoleMetrics, ComputeModelMetrics, ComputeThreadSummaries) match between before and after bit-for-bit.
- Task 2.4: Add Migration Engine Unit Tests in backend/internal/domain/migration/engine_test.go:
  - Test migration from v2 to v3: assert schema_version is 3, routine events pruned, ErrorMessage nil, and invariants pass.
  - Test chained migration from v1 to v3: assert both StepV1ToV2 and StepV2ToV3Sanitized execute sequentially and successfully.
  - Test idempotency on v3: assert migrating already-v3 ledger returns identical ledger without modification.

### Phase 3: Boundary Adapter Extraction & Regex Sanitization Elimination
- Task 3.1: Refactor Extractor in backend/internal/adapters/opencode/extractor.go:
  - Remove regex pattern emailPattern and function sanitizeTelemetryText.
  - In ExtractFacts, when encountering failed tool calls, record eventStatus = "error" and set errExcerpt = nil (or omit textual assignment).
  - Ensure EventFact for tool calls preserves ToolName, Status, Turn, Timestamp, and DurationMs with ErrorMessage set to nil.
- Task 3.2: Update Extractor Tests in backend/internal/adapters/opencode/extractor_test.go:
  - Update assertions on extracted tool error events to verify ErrorMessage is nil.
  - Assert bare metadata (tool name, turn, duration) remains populated.

### Phase 4: CLI Application Generation & In-Memory Derivation Sequence
- Task 4.1: Update runGenerate in backend/cmd/agentic-lens/main.go:
  - Assemble factLedger in memory with the complete raw events slice extracted from the database.
  - Derive metricsLedger via domain.DeriveMetricsLedger(factLedger, projCfg.RoleMappings).
  - Format summary markdown table via adapters.FormatMarkdownMetricsTable(metricsLedger).
  - Prior to writing factLedger to disk via adapters.SaveFactLedger, prune events:
    - Create persistedFactLedger copying factLedger.
    - Set persistedFactLedger.Events = domain.PruneEventsForPersistence(factLedger).
    - Assign CurrentSchemaVersion (3) to persistedFactLedger.SchemaVersion.
  - Pass persistedFactLedger to adapters.SaveFactLedger(*jsonOutputFlag, persistedFactLedger).
- Task 4.2: Update runMigrate in backend/cmd/agentic-lens/main.go:
  - Verify default flag target-version is domain.CurrentSchemaVersion (3).
- Task 4.3: CLI Integration Tests in backend/cmd/agentic-lens/main_test.go:
  - Add test asserting generate command produces JSON output without any error_message keys.
  - Assert that tool failure events in output contain tool_name, status: "error", and turn number.
  - Assert that markdown metrics table matches values derived from in-memory metrics.
  - Assert file size of generated JSON is under 5 KB.

### Phase 5: Backfill Engine Integration & Historical Ledger Sanitization
- Task 5.1: Update BackfillEngine in backend/internal/adapters/backfill/engine.go:
  - In Execute, ensure default migration target version is CurrentSchemaVersion (3).
  - In processSingleWorkspace, when RecomputeDB extracts events, apply domain.PruneEventsForPersistence before assigning to factLedger.Events.
  - Set factLedger.SchemaVersion to CurrentSchemaVersion (3).
- Task 5.2: Update MigrationRunner in backend/internal/adapters/backfill/migration_runner.go:
  - Verify default target version is CurrentSchemaVersion (3).
  - In dry-run mode, verify no files are written and report correctly indicates planned upgrade to v3.
  - In execution mode, verify atomic write creates sanitized v3 ledger with trailing newline and valid formatting.
- Task 5.3: Add Tests in backend/internal/adapters/backfill/migration_runner_test.go and engine_test.go:
  - Test batch migration across synthetic workspace directory with v1 and v2 ledgers.
  - Verify all output ledgers are upgraded to v3, contain zero ErrorMessage values, and have file sizes under 5 KB.
  - Verify dry-run mode does not modify workspace files on disk.

### Phase 6: Historical Workspaces Ledger Migration
- Task 6.1: Run Migration on Historical Workspaces:
  - Execute agentic-lens migrate across workspaces directory to upgrade existing historical ledgers to schema version 3.
  - Verify all workspaces/<TICKET_ID>/agentic_metrics.json files drop from thousands of lines to concise ledgers under 5 KB.
  - Verify zero database message strings or error excerpts remain in any committed workspace ledger.

### Phase 7: Verification, Benchmarks, and Compliance Audit
- Task 7.1: Execute backend unit test suite: go test -v -race ./... from backend directory.
- Task 7.2: Execute static analysis: go vet ./... from backend directory.
- Task 7.3: Execute repository quality gate: make check from repository root.
- Task 7.4: Verify zero any or empty interface across Go backend.
- Task 7.5: Verify zero emojis across all modified code, tests, docs, and artifacts.

---

## 4. Execution Dependencies and Delivery Sequence

1. Phase 1 (Domain Core Schema Version Promotion and Pure Event Pruning) establishes the pure mathematical core and pruning algorithms, ensuring all aggregate rollups remain invariant.
2. Phase 2 (Schema Migration Step V2 to V3 and Invariant Validation) implements the pure version transition and invariant checks, enabling safe automated upgrades.
3. Phase 3 (Boundary Adapter Extraction) aligns the SQLite extractor with bare error indicators and removes regex sanitization.
4. Phase 4 (CLI Application Generation) coordinates in-memory derivation before persistence pruning during generate CLI execution.
5. Phase 5 (Backfill Engine Integration) updates batch migration runners and tests.
6. Phase 6 (Historical Workspaces Ledger Migration) executes the migration to sanitize repository history and verify file sizes drop below 5 KB.
7. Phase 7 (Verification and Compliance) runs automated quality gates and confirms zero any and zero emojis.

---

## 5. Verification Plan

### Automated Test Commands
- Domain unit tests: cd backend && go test -v -race -run TestPruneEvents ./internal/domain/...
- Domain aggregations invariance: cd backend && go test -v -race -run TestDeriveMetricsLedger ./internal/domain/...
- Migration engine tests: cd backend && go test -v -race -run TestStepV2ToV3 ./internal/domain/migration/...
- Adapter extractor tests: cd backend && go test -v -race -run TestExtractFacts ./internal/adapters/opencode/...
- Backfill runner tests: cd backend && go test -v -race -run TestMigrationRunner ./internal/adapters/backfill/...
- CLI integration tests: cd backend && go test -v -race -run TestCLIGenerate ./cmd/agentic-lens/...
- Complete backend test suite: cd backend && go test -race ./...
- Static analysis: cd backend && go vet ./...
- Repository quality gate: make check

### Manual and Integration Verification Scenarios
- File Size Verification: Verify migrated historical workspaces (e.g. workspaces/AL-6/agentic_metrics.json, workspaces/AL-30/agentic_metrics.json) drop from >100 KB and thousands of lines to <5 KB.
- Error Message Absence: Search serialized JSON ledgers for "error_message" keys; assert zero instances exist.
- Dry-Run Verification: Run ./bin/agentic-lens migrate --workspaces workspaces/ --dry-run; assert status displays migrated (dry-run) and files on disk remain untouched.
- Bit-Identical Aggregates: Run agentic-lens metrics on a migrated ticket; assert derived total cost, token totals, active duration, and tool counts match pre-migration values exactly.
