# Implementation Plan: AL-6 Chronological Event Extraction and Schema Evolution in Go

## 1. Goal and Iteration Scope
Deliver chronological event extraction, schema evolution, and existing workspace backfill natively in Go across domain core, boundary adapters, SQLite extractor, CLI commands, and TypeScript frontend type contracts. This enables fine-grained event telemetry extraction from OpenCode SQLite databases into agentic_metrics.json, maintains complete backward compatibility with existing ledgers that omit events, and upgrades historical workspace ledgers on disk with populated event streams.

---

## 2. Grounded Reference Architecture and Concrete Paths

The implementation maps to the following existing files and packages:

### Domain Core Layer (backend/internal/domain/)
- backend/internal/domain/enums.go: Add EventType string enum type and typed constants (EventTypeToolCall, EventTypeTurnMessage, EventTypeCompaction).
- backend/internal/domain/models.go: Add EventFact struct. Add optional Events []EventFact to AgenticFactLedger.
- backend/internal/domain/aggregations.go: Verify pure aggregation functions preserve Events without regression.

### Boundary Adapters Layer (backend/internal/adapters/)
- backend/internal/adapters/schemas.go: Add EventFactDTO struct with bidirectional mapping to domain.EventFact. Update AgenticFactLedgerDTO with Events []EventFactDTO (json tag events,omitempty).
- backend/internal/adapters/repository.go: Update ExistingLedgerFacts and rawExistingLedger to parse and hold Events. Ensure safe decoding when events is absent.
- backend/internal/adapters/opencode/extractor.go: Extend extraction logic to query messages and parts, construct EventFact slices, correlate turn numbers and context tokens, detect tool errors, compute execution duration, and sort events chronologically.
- backend/internal/adapters/backfill/engine.go: Wire extracted events into processSingleWorkspace when --recompute-db is executed, ensure ledgers lacking events are candidate targets for recompute, and preserve non-destructive facts when local sessions are absent.

### CLI Application Layer (backend/cmd/agentic-lens/)
- backend/cmd/agentic-lens/main.go: Update runGenerate to receive extracted events from opencode.ExtractFacts, cache existing events if SQLite is not queried, and persist events in AgenticFactLedger. Ensure backfill subcommand runs smoothly for batch workspace updates.

### TypeScript Presentation Layer (web/src/domain/)
- web/src/domain/types.ts: Define EventFact interface with readonly fields matching domain models. Add optional readonly events?: readonly EventFact[]; to AgenticFactLedger.

### Test Suites
- backend/internal/adapters/opencode/extractor_test.go: Unit and adversarial tests for chronological event extraction, tool error detection, compaction correlation, and sorting.
- backend/internal/adapters/repository_test.go: Backward compatibility tests loading ledgers without events, and round-trip tests with events.
- backend/internal/adapters/backfill/engine_test.go: Backfill tests asserting event population on database match and preservation on database miss.
- backend/cmd/agentic-lens/main_test.go: CLI integration tests for generate and backfill with event emission.

---

## 3. Phased Implementation Tasks

### Phase 1: Domain Core Models and Event Types (backend/internal/domain/)
- Task 1.1: Define EventType string enum type in backend/internal/domain/enums.go:
  - Constant EventTypeToolCall = "tool_call"
  - Constant EventTypeTurnMessage = "turn_message"
  - Constant EventTypeCompaction = "compaction"
- Task 1.2: Define EventFact struct in backend/internal/domain/models.go:
  - Field Timestamp of type string with JSON tag timestamp
  - Field SessionID of type string with JSON tag session_id
  - Field Turn of type int64 with JSON tag turn
  - Field Role of type string with JSON tag role
  - Field EventType of type EventType with JSON tag event_type
  - Field ToolName of type *string with JSON tag tool_name,omitempty
  - Field Status of type string with JSON tag status
  - Field ErrorMessage of type *string with JSON tag error_message,omitempty
  - Field DurationMs of type *int64 with JSON tag duration_ms,omitempty
  - Field ContextTokens of type int64 with JSON tag context_tokens
- Task 1.3: Update AgenticFactLedger in backend/internal/domain/models.go:
  - Add field Events of type []EventFact with JSON tag events,omitempty
- Task 1.4: Verify DeriveMetricsLedger in backend/internal/domain/aggregations.go assigns factLedger directly to Facts, automatically propagating Events without modifying existing mathematical totals.

### Phase 2: Boundary DTOs and Backward-Compatible Serialization (backend/internal/adapters/)
- Task 2.1: Define EventFactDTO in backend/internal/adapters/schemas.go:
  - Fields matching EventFact using standard JSON tags and pointer semantics for optional values.
  - Implement value receiver method ToDomain() domain.EventFact performing validation and enum casting.
  - Implement standalone function EventFactDTOFromDomain(domain.EventFact) EventFactDTO.
- Task 2.2: Update AgenticFactLedgerDTO in backend/internal/adapters/schemas.go:
  - Add field Events of type []EventFactDTO with JSON tag events,omitempty.
  - Update ToDomain() method on AgenticFactLedgerDTO: if dto.Events is nil or empty, populate Events as empty slice; otherwise convert each EventFactDTO to domain.EventFact.
  - Update AgenticFactLedgerDTOFromDomain function to map ledger.Events to []EventFactDTO.
- Task 2.3: Update backend/internal/adapters/repository.go:
  - Add Events []domain.EventFact to ExistingLedgerFacts struct.
  - Add Events []EventFactDTO (json tag events,omitempty) to internal rawExistingLedger struct.
  - In LoadExistingFactLedger, map raw.Events into result.Events.
  - Ensure unmarshaling files without events leaves Events as nil or empty without returning an error.

### Phase 3: OpenCode SQLite Chronological Event Extractor (backend/internal/adapters/opencode/)
- Task 3.1: Update ExtractFacts signature in backend/internal/adapters/opencode/extractor.go:
  - Return signature: ([]domain.InvocationFact, []domain.ToolFact, []domain.EventFact, *string, *string, error).
- Task 3.2: Implement message event extraction and turn mapping:
  - During message iteration for each session, establish sequential turn indexing per session.
  - Build lookup map storing message timestamp, turn index, and cumulative context tokens (tokens.input + tokens.cache.read) keyed by message ID.
  - Emit domain.EventFact with EventTypeTurnMessage for messages, recording active duration when completion timestamps exist.
- Task 3.3: Implement tool call event extraction:
  - Query parts where data LIKE '%"type":"tool"%'.
  - Extract tool name, status, and output.
  - Apply error classification: flag error when status is error, output contains Traceback, or output prefix starts with Error:.
  - When error detected, set status to error and record truncated error excerpt in ErrorMessage.
  - Calculate DurationMs from time_updated - time_created when valid.
  - Resolve turn index and context tokens from parent message lookup.
  - Emit domain.EventFact with EventTypeToolCall.
- Task 3.4: Implement compaction event extraction:
  - Query parts where data LIKE '%"type":"compaction"%'.
  - Correlate turn index from parent message lookup.
  - Emit domain.EventFact with EventTypeCompaction.
- Task 3.5: Implement chronological sorting:
  - Combine message events, tool call events, and compaction events into a single slice.
  - Sort slice ascending by Timestamp, using Turn and session creation order as deterministic tie-breakers.
- Task 3.6: Defensive error handling:
  - Gracefully ignore corrupted JSON strings in message or part rows without aborting the extraction pass.

### Phase 4: CLI Generation and Backfill Integration (backend/cmd/agentic-lens/, backend/internal/adapters/backfill/)
- Task 4.1: Update backend/cmd/agentic-lens/main.go in runGenerate:
  - Initialize events slice from cachedFacts.Events if existing ledger is loaded.
  - Update opencode.ExtractFacts invocation to capture extracted events.
  - If database extraction succeeds, assign extracted events to factLedger.Events.
  - Assign Events: events when building domain.AgenticFactLedger.
- Task 4.2: Update backend/internal/adapters/backfill/engine.go in processSingleWorkspace:
  - When --recompute-db is executed and session is resolved in SQLite, capture extracted events and assign to factLedger.Events.
  - Enhance completion evaluation: if a ledger already has schema version 2 but has zero events, treat it as requiring event backfill when --recompute-db is enabled or when --force is supplied.
  - When --recompute-db is not executed or session is absent locally, preserve pre-existing factLedger.Events loaded from disk.
  - Marshal updated AgenticFactLedgerDTO containing events atomically to disk.

### Phase 5: TypeScript Frontend Domain Parity (web/src/domain/)
- Task 5.1: Update web/src/domain/types.ts:
  - Add EventFact interface with readonly properties: timestamp: string, session_id: string, turn: number, role: string, event_type: string, tool_name?: string, status: string, error_message?: string, duration_ms?: number, context_tokens: number.
  - Add optional readonly events?: readonly EventFact[]; to AgenticFactLedger.
- Task 5.2: Verify TypeScript strict compilation across all web views, components, and tests.

### Phase 6: Automated Verification, Adversarial Testing, and Workspace Backfill Delivery
- Task 6.1: Update backend/internal/adapters/opencode/extractor_test.go:
  - Add tests validating that ExtractFacts returns chronologically sorted events.
  - Add tests asserting tool call status, error extraction, duration calculation, and turn association.
  - Add tests asserting compaction events are captured.
  - Add adversarial tests with missing parts, corrupted part JSON, and empty sessions.
- Task 6.2: Add repository tests in backend/internal/adapters/repository_test.go:
  - Test decoding legacy ledger fixtures omitting events. Assert zero errors and empty Events.
  - Test round-trip encoding and decoding of ledgers containing populated Events.
- Task 6.3: Update backfill tests in backend/internal/adapters/backfill/engine_test.go:
  - Assert that backfill populates events when database session is found.
  - Assert that backfill preserves existing events when database session is missing.
  - Assert that ledgers lacking events are upgraded during DB recomputation.
- Task 6.4: Execute full quality suite:
  - Run go vet ./... in backend/.
  - Run go test -race ./... in backend/.
  - Run npm run typecheck and npm test in web/.
  - Build standalone CLI binary: go build ./cmd/agentic-lens.
- Task 6.5: Execute workspace backfill delivery:
  - Run ./bin/agentic-lens backfill --workspaces workspaces/ --recompute-db --force --update-summaries.
  - Verify that existing workspaces on disk with sessions in OpenCode SQLite have their agentic_metrics.json upgraded to include chronologically sorted events.
  - Inspect git diff across workspaces/ to ensure backward-compatible preservation of non-local sessions and accurate population of event streams.
- Task 6.6: Audit diff to ensure complete absence of emojis and zero usage of any or empty interface{}.
