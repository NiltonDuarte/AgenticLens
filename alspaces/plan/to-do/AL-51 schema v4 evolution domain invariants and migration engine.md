h3. Title: AL-51: Schema v4 Evolution, Domain Invariants, and V3->V4 Migration Engine

*TARGET*: AgenticLens (domain / models / migration / tests)

*CONTEXT*
 
AgenticLens telemetry ledgers are currently at Schema v3 (`SchemaVersion: 3`). While Schema v3 captures code facts, invocations, generic tool call counts, and compactions, it lacks structured representations for modular OpenCode skill loads and task subagent delegation links.
To enable first-class skill and subagent tracking without corrupting historical workspaces, the schema must evolve to Schema v4 with strict backward compatibility and deterministic domain migration steps.

*GOAL*
 
Define Schema v4 domain models (`SkillFact`, `TaskLinkFact`), update `AgenticFactLedger` and `EventFact`, and implement the `StepV3ToV4` migration engine with comprehensive domain invariant validation.

*VERTICAL SLICE SCOPE*
 
* Domain Core:
  - Add `SkillFact` (`Name string`, `Invocations int64`, `Errors int64`, `FirstTurn int64`, `LoadedByRole AgentRole`).
  - Add `TaskLinkFact` (`ParentSessionID string`, `Turn int64`, `SubagentRole AgentRole`, `ChildSessionID string`, `Description string`).
  - Update `AgenticFactLedger` in `backend/internal/domain/models.go` with `Skills []SkillFact` and `TaskLinks []TaskLinkFact`. Bump default schema constant to 4.
  - Extend `EventFact` in `models.go` with optional `SkillName *string` and `ChildSessionID *string`.
  - Update TypeScript types in `web/src/domain/types.ts` to mirror Go domain models with strict readonly fields and zero `any`.
* Domain Migration:
  - Implement `StepV3ToV4` in `backend/internal/domain/migration/step_v3_v4.go`.
  - Register `step_v3_v4` in `registry.go` and update `validator.go` to validate Schema v4 ledgers.
* Quality:
  - Unit tests in `engine_test.go` and `validator_test.go` verifying clean migrations from v1, v2, and v3 to v4.
  - Assert that missing `skills` and `task_links` arrays in legacy ledgers deserialize safely to non-nil empty slices.

*ACCEPTANCE CRITERIA*
 
* `SchemaVersion` constant bumped to 4 in domain constants.
* `SkillFact` and `TaskLinkFact` domain structs defined with complete JSON tags and immutability semantics.
* `StepV3ToV4` cleanly transitions valid v3 ledgers to v4, appending migration step identifiers.
* Deserializing historical v1/v2/v3 ledgers without skills or task links produces valid structs with initialized empty slices.
* `npm run typecheck`, `npm run lint`, and `go test ./...` pass with zero errors.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
