# AL-56 Iteration 03 Plan: Cross-Device Invocations Merging and SDLC Iterations Accumulation

## Overview

When developers or agents work across multiple devices on the same ticket branch, running telemetry generation on a secondary device currently results in code.sdlc_iterations being overwritten. Because the secondary device only possesses local session history in its SQLite database or rollout store, blind overwriting of the invocations slice and extraction of local-only steering turns discards historical sessions and resets the SDLC iteration count.

The user explicitly rejected inspecting workspace iteration directories on disk, reaffirming that the platform must remain strictly grounded in database facts and configuration.

This iteration implements session accumulation and cross-device iteration tracking by:
1. Elevating CurrentSchemaVersion from 3 to 4 across backend domain constants, migration runners, and TypeScript definitions.
2. Adding user_iterations to the atomic InvocationFact domain model and its corresponding serialization DTOs in Go and TypeScript.
3. Recording root user steering turns into user_iterations on root session invocations across both OpenCode and Codex extractors.
4. Replacing blind invocations overwriting with an ID-aware merge in the CLI generator, preserving historical invocations from prior devices.
5. Computing code.sdlc_iterations as the exact sum of user_iterations across merged invocations, strictly rejecting fallback heuristics and requiring schema migration for pre-v4 ledgers.

## Architectural Boundaries & Principles

- Functional Core, Imperative Shell (FCIS):
  The domain layer in backend/internal/domain/ remains a pure functional core. InvocationFact receives a primitive integer field with value semantics. All file loading, database querying, and slice merging remain strictly within the imperative boundary adapters and CLI command runners.
- Platform Data Grounding:
  Zero inspection of workspace iteration directories. The platform reads only the database, rollout logs, configuration, and existing atomic fact ledgers.
- Deterministic Merging:
  Merging invocations is an idempotent operation based on unique invocation identifiers. Running generation multiple times on the same device must produce identical, stable ledgers without duplicating invocations or inflating iteration counts.
- Strict Zero-Code Mandate:
  This specification is written entirely in descriptive prose and structured task lists.

## Phased Implementation Tasks

### Phase 1: Schema Version 4 Elevation and Domain DTO Model Updates
- Targets:
  - backend/internal/domain/constants.go
  - backend/internal/domain/models.go
  - backend/internal/domain/migration.go
  - backend/internal/adapters/schemas.go
  - backend/internal/adapters/repository.go
  - backend/internal/adapters/backfill/engine.go
  - backend/internal/adapters/backfill/migration_runner.go
  - web/src/domain/types.ts
- Tasks:
  1. In backend/internal/domain/constants.go:
     Update CurrentSchemaVersion to 4.
  2. In backend/internal/domain/models.go:
     Add UserIterations of type int64 with JSON tag user_iterations to domain.InvocationFact. Ensure domain models remain pure value types with zero external dependencies.
  3. In backend/internal/adapters/schemas.go:
     Add UserIterations of type int64 with JSON tag user_iterations to InvocationFactDTO.
     Update ToDomain method on InvocationFactDTO to map UserIterations into domain.InvocationFact.
     Update InvocationFactDTOFromDomain to map UserIterations from domain.InvocationFact into InvocationFactDTO.
  4. In backend/internal/adapters/repository.go:
     Confirm LoadExistingFactLedger parses user_iterations correctly through InvocationFactDTO across schema v4 and legacy schemas (v1, v2, v3). For historical ledgers lacking this field, verify that Go zero-initializes the field to 0 without parsing errors.
  5. In backend/internal/adapters/backfill/engine.go and migration_runner.go:
     Ensure migration runners and backfill workflows recognize schema-v4 as the target version, migrating v1, v2, and v3 ledgers into schema-v4.
  6. In web/src/domain/types.ts:
     Add readonly user_iterations to the TypeScript InvocationFact interface, ensuring full structural parity between Go backend JSON output and TypeScript frontend consumption.

### Phase 2: Extraction Adapters for OpenCode and Codex
- Targets:
  - backend/internal/adapters/opencode/extractor.go
  - backend/internal/adapters/codex/rollout.go
- Tasks:
  1. In backend/internal/adapters/opencode/extractor.go:
     During ExtractFacts, track user messages per session while processing message rows.
     For root sessions (parent session ID is nil), compute the total number of user-authored steering messages.
     When constructing InvocationFact records, assign this count to UserIterations for the primary invocation (index 0).
     For multi-model sessions emitting secondary invocation records (index greater than 0) or for child subagent sessions (parent session ID is non-nil), explicitly set UserIterations to 0 to prevent multi-model duplication.
  2. In backend/internal/adapters/codex/rollout.go:
     In transcript fact assembly and CombineTranscripts, assign the root transcript steering turns to UserIterations on the primary invocation (index 0) of the root thread.
     Ensure child rollouts, subagent transcripts, and non-primary model slices have UserIterations set to 0.

### Phase 3: Invocations Merging and Cross-Device Session Preservation
- Targets:
  - backend/cmd/agentic-lens/main.go
- Tasks:
  1. In backend/cmd/agentic-lens/main.go (runGenerate):
     Replace destructive slice assignment of invocations with an ID-aware merge function.
     Initialize the merged slice with cached invocations from cachedFacts.
     For each newly extracted invocation from the local database or rollout:
     If an invocation with the same ID and model already exists in the cached set, replace it with the fresh authoritative local record.
     If an invocation does not exist in the cached set, append it.
     Retain all cached invocations that originated from earlier sessions or different devices whose IDs were not present in the local database.
  2. Apply equivalent non-destructive aggregation to tool usage facts by merging tool calls and error counts across distinct sessions, and append novel event records from local sessions.

### Phase 4: Strict SDLC Iteration Summation (Zero Fallbacks)
- Targets:
  - backend/cmd/agentic-lens/main.go
- Tasks:
  1. In backend/cmd/agentic-lens/main.go:
     Compute totalSDLC by iterating through the merged invocations slice and summing UserIterations across all root invocations.
  2. Implement strict zero-fallback policy:
     Assign codeFacts.SDLCIterations = totalSDLC.
     If totalSDLC is 0 (for example, when an existing ledger lacks populated user_iterations because it was generated under an older schema version), do not fall back to loose steering turn heuristics, cached values, or default guesses.
     Check whether cachedFacts is present with schema_version less than 4; if so, emit an explicit diagnostic notifying that the existing ledger uses schema version less than 4 and must be migrated using migrate-schema or backfill.
  3. Ensure that when cross-device sessions accumulate, the sum strictly reflects total user steering turns across all recorded root sessions without loss or artificial inflation.

### Phase 5: Adversarial and Parity Testing
- Targets:
  - backend/internal/adapters/opencode/extractor_test.go
  - backend/internal/adapters/codex/rollout_test.go
  - backend/cmd/agentic-lens/main_test.go
  - backend/internal/adapters/repository_test.go
  - web/src/
- Tasks:
  1. In extractor_test.go:
     Assert that UserIterations is accurately populated for root sessions and remains 0 for subagents and secondary model splits.
  2. In rollout_test.go:
     Assert that Codex transcript parsing populates UserIterations on the root invocation and isolates child invocations.
  3. In main_test.go:
     Construct a multi-device handoff simulation test:
     Create an existing ledger representing Device A with invocation A having user_iterations = 3.
     Run generate against an isolated test database representing Device B containing only invocation B with user_iterations = 2.
     Verify that the resulting ledger contains both invocation A and invocation B, and that code.sdlc_iterations equals 5.
     Test idempotency by executing generate repeatedly against the same database, verifying no duplicate invocations or iteration count inflation occurs.
     Test strict rejection of fallbacks: verify that older schema ledgers lacking user_iterations are not silently guessed and require migration.
  4. In repository_test.go:
     Verify that serializing and deserializing ledgers with UserIterations preserves exact values.
  5. In web/ frontend tests:
     Execute npm test to confirm TypeScript type checks and component tests pass without errors.

### Phase 6: Workspace and Telemetry Synchronization
- Tasks:
  1. Verify that all documentation artifacts in alspaces/AL-56/ strictly obey the Zero Emojis Mandate.
  2. Maintain alspaces/AL-56/SUMMARY.md and PULL_REQUEST.md.
  3. Keep telemetry generation synchronized as the final step of iteration completion.
