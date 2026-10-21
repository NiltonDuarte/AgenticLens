# Contract: AL-32 Telemetry Event Pruning, Database Message Stripping, Agent Data Investigation, and Ledger Backfill

## 1. Problem Statement

In the current telemetry lifecycle of AgenticLens, the generate subcommand dumps chronological event streams extracted from the OpenCode SQLite database directly into committed workspace ledgers under workspaces/<TICKET_ID>/agentic_metrics.json. Persisting raw messages and error output strings from the database creates critical security vulnerabilities and severe repository bloat:

1. Security Hazards: Unbounded terminal output, compiler diagnostics, tool payloads, and prompt excerpts frequently contain confidential environment variables, API credentials, private authentication tokens, absolute filesystem paths, internal network addresses, and proprietary source snippets. Storing these strings in committed workspace files permanently leaks secrets into Git version history.
2. Inadequacy of Regex Redaction: Attempting regex-based pattern matching over arbitrary database strings is fragile, incomplete, and leak-prone. Redaction creates a false sense of security while failing against multi-line traces, structured payloads, or novel credential formats. Complete omission of all database message text from the Git plane is the only robust architectural defense.
3. Repository Bloat: Routine session activity records thousands of lines of completed turn messages (assistant token events) and routine successful tool calls. Historical workspace ledgers frequently exceed hundreds of kilobytes or even thousands of JSON lines for single-ticket iterations, despite storing mostly redundant routine occurrences.

To resolve these defects permanently, Ticket AL-32 establishes strict boundaries:
- The committed Git plane must strictly enforce zero database message persistence: no prompts, no assistant text, and no raw database error strings are ever written to workspaces/<TICKET_ID>/agentic_metrics.json.
- Failed tool calls in committed ledgers are retained strictly as bare structural error indicators (tool name, error status, turn number, timestamp, duration) with error message pointers omitted or nil. Full diagnostic context is reserved strictly for local hydration from the developer local SQLite database.
- All mathematical rollups, token aggregates, financial costs, active durations, tool execution health metrics, and anomaly detection evaluations must be computed over the complete raw event stream in memory before pruning and stripping events for disk persistence.
- A discrete schema migration step (StepV2ToV3Sanitized) is introduced to upgrade schema version 2 ledgers to schema version 3, stripping legacy message strings and pruning routine events while asserting bit-identical aggregate metrics.
- An investigation evaluates pathways for capturing clean structured metadata directly from the agent runtime versus relying solely on local dynamic hydration.
- Historical workspace ledgers across the repository are migrated and backfilled, dropping ledger file sizes to under 5 KB for typical sessions.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All changes across the Go backend must strictly adhere to TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. CLI Application Layer (backend/cmd/agentic-lens/main.go):
   - In runGenerate, coordinates extraction of raw facts into memory.
   - Derives the full metrics ledger and evaluates operational anomalies against the complete, unpruned fact ledger in memory.
   - Invokes domain event pruning and stripping on the fact ledger immediately prior to delegating persistence to repository adapters.
   - In runMigrate, defaults target schema version to CurrentSchemaVersion (schema version 3).
   - Contains zero SQL queries, zero JSON serialization logic, and zero mathematical rollup algorithms.
2. Boundary Adapters Layer (backend/internal/adapters/):
   - Database Extraction (backend/internal/adapters/opencode/extractor.go):
     - Eliminates regex text masking on database outputs.
     - Populates event facts with structural metadata; ensures persisted error messages are omitted or nil.
   - Backfill and Migration (backend/internal/adapters/backfill/):
     - Migration runner executes StepV2ToV3Sanitized across discovered workspace ledgers.
     - Supports dry-run inspection without disk mutation.
     - Backfill engine incorporates schema migration and sanitization prior to recomputation.
   - Boundary Schemas and Repository (backend/internal/adapters/schemas.go, repository.go):
     - Maintains EventFactDTO with omitempty serialization on ErrorMessage. When ErrorMessage is nil, the key is omitted entirely from JSON payloads.
3. Functional Core Domain Layer (backend/internal/domain/):
   - Constants (backend/internal/domain/constants.go):
     - Promotes CurrentSchemaVersion from 2 to 3.
     - Defines migration step identifier MigrationStepSanitizedEvents.
   - Pure Event Pruning Function (backend/internal/domain/events.go):
     - Pure function PruneEventsForPersistence takes an immutable fact ledger and returns a pruned, sanitized slice of EventFact.
     - Zero disk I/O, zero network, zero database handles, zero clocks.
   - Schema Migration Step (backend/internal/domain/migration/step_v2_v3.go):
     - Pure implementation of MigrationStep transitioning version 2 to version 3.
     - Uses PruneEventsForPersistence and updates migration metadata immutably.
   - Migration Invariant Validator (backend/internal/domain/migration/validator.go):
     - Evaluates version-specific invariants: validates version 3 ledgers maintain identity, invocations, tools, and bit-identical aggregate metrics while confirming zero residual error messages and pruned event volume.
4. Presentation Layer (web/src/):
   - Out of scope for new UI components in this ticket.
   - TypeScript telemetry types already treat error_message as optional string or undefined; AnomalyDrawer already guards error payload rendering behind presence checks.
   - Verifies TypeScript type contracts and existing test suites pass cleanly without regression.

### Pure Domain Core Isolation
- Domain models and aggregation functions operate exclusively on immutable value structs.
- Event pruning and message stripping logic is a pure transformation function in the domain core.
- No database handles, CLI flags, or filesystem paths enter backend/internal/domain/.

### Imperative Shell Boundary Rules
- Zero Database Message Persistence: Serialized JSON ledgers under workspaces/<TICKET_ID>/agentic_metrics.json must never contain raw database error text, assistant reasoning, prompt text, or command output strings. The error_message property must be absent from JSON output.
- Structural Bare Indicators: Retained error events preserve tool_name, status as "error", turn, timestamp, duration_ms, and context_tokens without textual context.
- Zero Any Mandate: Zero any or empty interface across Go and TypeScript code.
- Zero Emoji Mandate: Zero emojis across code, docstrings, error messages, terminal logs, tests, and documentation artifacts.

---

## 3. Scope Boundaries: Go Backend, CLI, Migration, and Presentation

### In Scope
1. Domain Core Constants and Pruning Logic (backend/internal/domain/):
   - Advance CurrentSchemaVersion to 3 in backend/internal/domain/constants.go.
   - Define constant MigrationStepSanitizedEvents = "schema_v3_sanitized".
   - Implement pure domain function PruneEventsForPersistence in backend/internal/domain/events.go: filters events to bare failed tool calls, compaction events, and anomaly-associated events, setting ErrorMessage to nil and omitting routine completed turns and successful tool calls.
   - Verify and tune DetectAnomalies in backend/internal/domain/anomalies.go so that anomaly evaluations against sanitized ledgers maintain bit-identical consistency with raw evaluations.
2. Migration Engine and Step V2 to V3 (backend/internal/domain/migration/):
   - Implement StepV2ToV3Sanitized in backend/internal/domain/migration/step_v2_v3.go implementing the MigrationStep interface.
   - Register StepV2ToV3Sanitized in Registry and default Engine initialization.
   - Update InvariantValidator in backend/internal/domain/migration/validator.go to branch on targetVersion: for version 3, enforce schema version 3, identity preservation, invocation and tool equality, bit-identical derived totals and efficiency rollups, zero error messages on retained events, and omission of routine completed events.
3. OpenCode Extractor Adapter (backend/internal/adapters/opencode/):
   - In extractor.go, remove regex text masking (sanitizeTelemetryText).
   - Ensure tool call events set ErrorMessage to nil when preparing facts for Git persistence, retaining bare error metadata.
4. CLI Generation and Migrate Subcommands (backend/cmd/agentic-lens/):
   - In main.go (runGenerate), derive full metrics and detect anomalies in memory against the raw fact ledger before invoking PruneEventsForPersistence to create the fact ledger saved to disk.
   - Update default target schema version in runMigrate to CurrentSchemaVersion (v3).
5. Backfill Engine and Migration Runner (backend/internal/adapters/backfill/):
   - Wire StepV2ToV3Sanitized into MigrationRunner and BackfillEngine.
   - Ensure dry-run preview operates without modifying files.
   - RecomputeDB in BackfillEngine applies event pruning and message stripping to newly extracted facts before persistence.
6. Structured Metadata Investigation:
   - Formulate and document an architectural evaluation comparing local SQLite hydration versus runtime-reported structured agent metadata (exit codes, error categories, error phases).
7. Test Suites and Fixtures:
   - Table-driven unit tests in domain verifying pure pruning logic, bare error retention, aggregate metric invariance, and migration validation.
   - Adapter tests in backfill and opencode verifying dry-run, backup creation, and file size reduction to under 5 KB.
   - CLI integration tests asserting absence of error_message keys in generated JSON ledgers.

### Out of Scope
1. Dynamic SQLite hydration HTTP endpoints and WebSocket streaming (reserved for AL-33).
2. Dual-plane visual waterfall UI changes in the TypeScript web application (reserved for AL-33).
3. Modifying OpenCode SQLite table schema or table column definitions.
4. Changes to Git commit parsing or code metrics classification logic.

---

## 4. Architectural Specifications

### Event Pruning and Retention Rules

In the committed Git plane (workspaces/<TICKET_ID>/agentic_metrics.json), the events array must contain strictly:
1. Failed Tool Calls:
   - Condition: EventType == "tool_call" and Status == "error".
   - Retention form: Bare error indicator with ToolName, Status ("error"), Turn, Timestamp, DurationMs, ContextTokens. ErrorMessage is set to nil and omitted from JSON.
2. Compaction Events:
   - Condition: EventType == "compaction".
   - Retention form: Timestamp, SessionID, Turn, Role, EventType, Status ("completed"), ContextTokens.
3. Anomaly-Associated Events:
   - Condition: Events directly associated with detected operational anomalies (e.g. tool calls participating in a detected ToolRetryStreak, turn messages marking a ContextSaturationSpike, or role transition turns participating in a SubagentReworkLoop).
   - Retention form: Retained for diagnostic traceability, with ErrorMessage set to nil.
4. Omitted Routine Events:
   - Routine turn messages (EventType == "turn_message" with Status == "completed") not associated with detected anomalies are omitted completely.
   - Successful tool calls (EventType == "tool_call" with Status == "completed") not associated with detected anomalies are omitted completely.

### In-Memory Derivation Precedence in Telemetry Generation

To guarantee that derived metrics, workflow efficiency, tool execution stats, and anomaly flags remain 100% accurate despite event pruning in Git:
1. Step 1 (Extraction): Extractor queries OpenCode SQLite database and builds the complete raw InvocationFact, ToolFact, and EventFact slices in memory.
2. Step 2 (Full Raw Fact Assembly): Fact ledger is assembled in memory containing the complete raw event stream.
3. Step 3 (In-Memory Metrics Derivation): DeriveMetricsLedger evaluates all mathematical rollups, context health, subagent velocity, role effort, and operational anomalies (DetectAnomalies) against the complete raw event stream.
4. Step 4 (Markdown Summary Generation): FormatMarkdownMetricsTable renders the summary table using the full derived metrics ledger.
5. Step 5 (Persistence Pruning): Pure function PruneEventsForPersistence transforms the fact ledger events slice into the sanitized, pruned event slice (stripping ErrorMessage and omitting routine events).
6. Step 6 (Disk Serialization): Fact ledger with pruned events is written to agentic_metrics.json via SaveFactLedger. Resulting JSON file is typically under 5 KB.

### Schema Migration Step V2 to V3 Specification

Step identifier: StepV2ToV3Sanitized
Source version: 2
Target version: 3
Description: "Prunes routine completed events and strips database message text from error events while maintaining bit-identical aggregate metrics"
Execution logic:
1. Read input AgenticFactLedger (schema version 2).
2. Evaluate DetectAnomalies on the input ledger to identify active anomalies.
3. Apply PruneEventsForPersistence to prune routine events and strip ErrorMessage.
4. Set SchemaVersion to 3.
5. Record "schema_v3_sanitized" in StepsCompleted within MigrationInfo.
6. Return upgraded AgenticFactLedger.
Validation invariants (InvariantValidator):
1. SchemaVersion must equal 3.
2. TicketID, TicketTitle, TicketType must match before state exactly.
3. StartedAt, CompletedAt, and Code must match before state exactly.
4. Invocations slice and Tools slice must match before state exactly.
5. Every event in after.Events must have ErrorMessage == nil.
6. Number of events in after.Events must be less than or equal to before.Events.
7. ComputeTotals, ComputeWorkflowEfficiency, ComputeRoleMetrics, ComputeModelMetrics, and ComputeThreadSummaries derived from before and after must be bit-identical.

### Structured Agent Metadata Investigation Evaluation

Investigation Focus: Evaluating pathways for obtaining structured error and execution metadata directly from the agent runtime versus relying solely on local dynamic hydration.

Findings and Architecture Options:
1. Option A: Local Dynamic Hydration from Local SQLite Database:
   - Mechanism: Telemetry ledger in Git stores only bare structural indicators (tool name, turn, session ID). When viewed locally in the web dashboard or CLI report, the server queries the local OpenCode SQLite database using session_id and turn/part_id to hydrate raw output strings on-the-fly.
   - Strengths: Zero risk of Git credential leakage; eliminates all repository bloat; zero modifications needed to agent execution runtimes or subagent prompts.
   - Limitations: Machine-local only. Team members or CI environments without the original SQLite database cannot view the hydrated text (they see only the bare indicators).
2. Option B: Structured Metadata Emission from Agent Runtime:
   - Mechanism: Agent runtime or subagent harnesses emit structured JSON diagnostics directly upon tool failure:
     - Process exit code (e.g. exit_code: 1, 127).
     - Standardized error classification enum (e.g. category: "assertion_failure", "compilation_error", "timeout", "not_found", "syntax_error").
     - Target command stem or file subject (e.g. command: "go test", target: "backend/internal/domain/events.go").
     - Failure phase enum (e.g. phase: "build", "test", "lint").
   - Strengths: Fully portable in Git; safe from credential leakage (enumerated categories contain zero environment variables or paths); enables rich cross-ticket analytics (fleet-wide breakdown of compilation vs test failures).
   - Limitations: Requires upstream protocol support in OpenCode or subagent tooling wrappers to classify errors at execution time.
3. Architectural Recommendation:
   - Implement a Dual-Plane Hybrid Strategy:
     - Short Term (AL-32 & AL-33): Enforce Git sanitization with bare indicators and implement dynamic local SQLite hydration for full developer diagnostics.
     - Medium Term: Extend EventFact schema to accept optional structured fields (exit_code int, error_category string) populated by adapters whenever OpenCode or tool wrappers report structured diagnostics, preserving complete message stripping on raw text.

---

## 5. Acceptance Criteria

1. Running agentic-lens generate produces workspaces/<TICKET_ID>/agentic_metrics.json containing strictly bare error events, compaction events, and anomaly-associated events. All ErrorMessage pointers are nil, and the error_message key is omitted from JSON.
2. Zero prompts, zero assistant message text, and zero raw database error outputs are present in generated JSON ledgers. No regex sanitization is used; message text is completely omitted.
3. Bare error events preserve ToolName, Status ("error"), Turn, Timestamp, DurationMs, and ContextTokens.
4. In-memory derivations (ComputeTotals, ComputeWorkflowEfficiency, ComputeRoleMetrics, ComputeModelMetrics, ComputeThreadSummaries, and DetectAnomalies) execute over the complete raw event stream, producing exact totals.
5. StepV2ToV3Sanitized successfully migrates schema version 2 ledgers to schema version 3, stripping legacy ErrorMessage text and pruning routine events while maintaining bit-identical aggregate metrics.
6. InvariantValidator asserts non-destructive migration invariants: verifies schema version 3, identity preservation, invocation/tool fidelity, bit-identical derived totals, and complete absence of error message strings.
7. Running agentic-lens migrate --workspaces workspaces/ --dry-run previews v2 -> v3 migrations without modifying disk files.
8. Running agentic-lens migrate --workspaces workspaces/ migrates historical ledgers to schema version 3, reducing file size of typical ledgers to under 5 KB.
9. Table-driven tests in backend/internal/domain/ and backend/internal/adapters/ pass with 100% success under race detector.
10. Strict compliance with TECHNICAL_GUIDELINES.md: zero any across Go and TypeScript, and zero emojis across all code, tests, logs, and artifacts.
