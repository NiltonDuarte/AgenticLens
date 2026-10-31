# AL-51: Additive Domain Contracts, OpenCode Extractor, and Live Stream Ingestion for Skills and Tasks

- **Subsystem**: Domain / Adapters / OpenCode / Live / HTTP / Tests
- **Phase**: Phase 8 (v2 Observability / Skills)
- **Status**: Ready
- **Blocked by**: None
- **Blocks**: AL-52, AL-53, AL-54

## Context
Operators currently lack first-class visibility into modular skill invocations (`tool == "skill"`) and subagent delegations (`tool == "task"`), both during live real-time observation and in historical ticket autopsies. To allow iterative UI exploration without premature migration churn, domain contracts for skills and tasks must be introduced additively. By adding optional fields (`omitempty`) to `AgenticFactLedger` and maintaining `SchemaVersion` at 3, existing ledgers remain fully valid and backward-compatible while both offline extractors and live SSE streams gain full dual-plane parity.

## Goal
Establish pure Go and TypeScript domain contracts for `SkillFact` and `TaskLinkFact` as optional fields on `AgenticFactLedger`, implement extraction in the OpenCode adapter from SQLite session logs, and extend the live SQLite watcher and SSE streaming engine to capture skill and task facts in real time.

## Scope Across Layers

### 1. Functional Core (`internal/domain/models.go`, `internal/domain/live.go`)
- Define `SkillFact`:
  ```go
  type SkillFact struct {
      Name         string `json:"name"`
      TurnIndex    int    `json:"turn_index"`
      MessageIndex int    `json:"message_index"`
      CallerRole   string `json:"caller_role"`
      Timestamp    string `json:"timestamp"`
      Success      bool   `json:"success"`
      Error        string `json:"error,omitempty"`
  }
  ```
- Define `TaskLinkFact`:
  ```go
  type TaskLinkFact struct {
      SubagentType string  `json:"subagent_type"`
      Description  string  `json:"description"`
      ParentTurn   int     `json:"parent_turn"`
      ChildSession string  `json:"child_session_id,omitempty"`
      Timestamp    string  `json:"timestamp"`
      DurationMs   int64   `json:"duration_ms,omitempty"`
      CostUSD      float64 `json:"cost_usd,omitempty"`
  }
  ```
- Add additive fields to `AgenticFactLedger`:
  ```go
  Skills    []SkillFact    `json:"skills,omitempty"`
  TaskLinks []TaskLinkFact `json:"task_links,omitempty"`
  ```
- Extend `LivePartFact` with additive optional fields:
  ```go
  SkillName      string `json:"skill_name,omitempty"`
  ChildSessionID string `json:"child_session_id,omitempty"`
  SubagentType   string `json:"subagent_type,omitempty"`
  ```
- Invariant: Maintain `CurrentSchemaVersion = 3`. No migration step is executed in this ticket.

### 2. Frontend Contracts (`web/src/types.ts`, `web/src/domain/live.ts`)
- Mirror Go structures in TypeScript:
  ```typescript
  export interface SkillFact {
    name: string;
    turn_index: number;
    message_index: number;
    caller_role: string;
    timestamp: string;
    success: boolean;
    error?: string;
  }

  export interface TaskLinkFact {
    subagent_type: string;
    description: string;
    parent_turn: number;
    child_session_id?: string;
    timestamp: string;
    duration_ms?: number;
    cost_usd?: number;
  }

  export interface LivePartFact {
    // ... existing fields ...
    skill_name?: string;
    child_session_id?: string;
    subagent_type?: string;
  }
  ```
- Add optional fields to `AgenticFactLedger` in `types.ts`:
  ```typescript
  skills?: SkillFact[];
  task_links?: TaskLinkFact[];
  ```
- Enforce defensive accessors (`ledger.skills ?? []`, `ledger.task_links ?? []`) across frontend consumers to prevent runtime exceptions when processing legacy ledgers.

### 3. Imperative Shell - Offline OpenCode Extractor (`internal/adapters/opencode/extractor.go`)
- Enhance OpenCode session extraction:
  - When a tool call with name `"skill"` is processed, extract `args.name` into `SkillFact`.
  - When a tool call with name `"task"` is processed, extract `args.subagent_type`, `args.description`, and locate the spawned child session id if recorded in the tool output or OpenCode metadata into `TaskLinkFact`.
  - Append extracted facts to the ledger.

### 4. Imperative Shell - Live Stream Watcher (`internal/adapters/live/watcher.go`)
- During incremental SQLite polling of active sessions:
  - Detect tool invocation parts where `Tool == "skill"`: parse JSON arguments to populate `SkillName`.
  - Detect tool invocation parts where `Tool == "task"`: parse JSON arguments for `SubagentType` and correlate completion parts or child session spawn records to populate `ChildSessionID`.
  - Propagate these attributes through the SSE event channel (`/api/live/stream`) and trace payload endpoints (`/api/live/traces/:id`).

### 5. Automated Tests
- Unit tests in `internal/domain/models_test.go` verifying JSON serialization with and without skills/task_links.
- Extractor tests in `internal/adapters/opencode/extractor_test.go` using synthetic SQLite fixture databases containing skill and task invocations.
- Live watcher tests in `internal/adapters/live/watcher_test.go` verifying tool extraction for skills and task delegations during live streaming.
- Frontend tests in `web/src/domain/events.test.ts` and `web/src/domain/live.test.ts` validating parsing of legacy, skill-augmented, and live SSE event payloads.

## Acceptance Criteria
1. `AgenticFactLedger` includes optional `Skills` and `TaskLinks` slices without breaking existing v3 ledgers.
2. `CurrentSchemaVersion` remains 3.
3. OpenCode extractor accurately parses `tool == "skill"` calls and populates `SkillFact` records with name, turn index, caller role, and success state.
4. OpenCode extractor accurately parses `tool == "task"` calls and populates `TaskLinkFact` records with subagent type, description, parent turn, and child session identifier if available.
5. Active session live stream emits `LivePartFact` events with `skill_name` populated when skills are invoked, and `subagent_type` / `child_session_id` populated when tasks are delegated.
6. Zero regressions in legacy test suites.

## Definition of Done
- Pure domain models and TypeScript interfaces authored with zero `any` or `interface{}`.
- Extractor and live watcher logic covered by unit tests with synthetic SQLite fixtures.
- All Go backend tests and Vitest suites pass (`make check`).
- Zero emojis across all code, tests, and documentation.
