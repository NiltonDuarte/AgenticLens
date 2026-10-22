# Implementation Plan: FIX-3 Iteration 04 - Strong Anomaly Rule Typing and OpenCode Database Provenance Attribution

This plan is authored for `alspaces/FIX-3/iterations/04/PLAN.md`.

---

## 1. Goal and Iteration Scope

Incorporate user steering feedback for ticket FIX-3 across two focused technical deliverables:

1. Strongly Type Anomaly Rule Identifiers and Decouple Predicate Logic:
   - Replace the loose `ruleId: string` parameter in `isToolFailureAnomalyRule` with a strongly typed union over domain and live observability rules.
   - Eliminate heuristic string content guessing (`startsWith('TOOL_')` and `includes('TOOL_FAILURE')`), replacing it with an exact membership check against the tool failure rule subset.
   - Enforce strict typing on `LiveAnomalyAlert.rule_name` by removing the loose `| string` union escape hatch.
   - Introduce a named Go domain type `LiveAnomalyRuleName` in `backend/internal/domain/live_anomaly.go` to mirror `AnomalyRuleID` in `backend/internal/domain/enums.go`.

2. OpenCode Database Telemetry Provenance Attribution in All Prompt Layers:
   - Update all four debug prompt generators in `web/src/utils/debugPrompt.ts` (Layer 1: Tool Execution, Layer 2: Execution Issue / Anomaly Card, Layer 3: Message Turn Review, Layer 4: Session Overview) to explicitly state that the telemetry was extracted from the OpenCode database.

Non-goals:
- Modifying OpenCode SQLite schema or table layouts.
- Altering existing prompt heading structures or checklist questions.
- Re-introducing internal ticket identifiers or git branch names into debug prompts.

---

## 2. Grounded Reference Architecture and File Paths

The plan is grounded directly against existing packages, modules, and symbols:

### Go Functional Core (`backend/internal/domain/`)
- `backend/internal/domain/enums.go`:
  - Contains existing `AnomalyRuleID` named string type with constants: `RuleToolRetryStreak`, `RuleToolHighFailureRate`, `RuleContextSaturationSpike`, `RuleSubagentReworkLoop`.
- `backend/internal/domain/live_anomaly.go`:
  - Contains untyped constants: `RuleConsecutiveToolErrors`, `RuleContextCeilingApproach`, `RuleTokenVelocityRunaway`, `RuleToolThrashing`.
  - Contains `LiveAnomalyAlert` struct with field `RuleName string`.
- `backend/internal/domain/live_anomaly_test.go`:
  - Exercises live anomaly evaluation rules and verifies sorting and field mappings.

### TypeScript Domain Models and Types (`web/src/`)
- `web/src/types/telemetry.ts`:
  - Contains `AnomalyRuleID` string union (`'TOOL_RETRY_STREAK' | 'TOOL_HIGH_FAILURE_RATE' | 'CONTEXT_SATURATION_SPIKE' | 'SUBAGENT_REWORK_LOOP'`).
- `web/src/domain/live.ts`:
  - Contains `LiveAnomalyRule` string union (`'ConsecutiveToolErrors' | 'ContextCeilingApproach' | 'TokenVelocityRunaway' | 'ToolThrashing'`).
  - Contains `LiveAnomalyAlert` interface with `rule_name: LiveAnomalyRule | string`.
- `web/src/domain/types.ts`:
  - Central re-export barrel for domain types.
- `web/src/domain/anomalies.ts`:
  - Exports `isToolFailureAnomalyRule(ruleId: string): boolean`.
  - Exports `matchesAnomaly`, `findMatchingAnomaly`, `extractStreakCount`, `findMatchingLiveAlert`.
- `web/src/domain/anomalies.test.ts`:
  - Tests anomaly matching predicates and rule categorization.
- `web/src/domain/live.test.ts`:
  - Tests live anomaly alert structures and evaluations.

### TypeScript Debug Prompt Generator (`web/src/utils/`)
- `web/src/utils/debugPrompt.ts`:
  - Exports `translateRuleToPlainEnglish(ruleId: string, toolName?: string): string`.
  - Exports `generateToolExecutionPrompt`, `generateExecutionIssuePrompt`, `generateMessageTurnPrompt`, `generateSessionOverviewPrompt`.
- `web/src/utils/debugPrompt.test.ts`:
  - Unit tests verifying Markdown formatting and sanitization across all four layers.

### UI Presentation Components (`web/src/components/`)
- `web/src/components/TraceWaterfall.tsx`:
  - Re-exports `isToolFailureAnomalyRule` from `../domain/anomalies`.
- Component test suites: `TraceWaterfall.test.tsx`, `LiveTraceInspector.test.tsx`, `FailedToolAlertCard.test.tsx`, `AnomalyDrawer.test.tsx`.

---

## 3. Architectural Specifications and Contracts

### Anomaly Rule Type System Contract

1. Backend Go Domain (`backend/internal/domain/live_anomaly.go`):
   - Define named type `LiveAnomalyRuleName` as an underlying `string`.
   - Update constants `RuleConsecutiveToolErrors`, `RuleContextCeilingApproach`, `RuleTokenVelocityRunaway`, and `RuleToolThrashing` to be typed explicitly as `LiveAnomalyRuleName`.
   - Update `LiveAnomalyAlert.RuleName` field type from `string` to `LiveAnomalyRuleName`.
   - Untyped string constants assigned in tests remain fully compatible in Go, while providing concrete domain typing.

2. Frontend Live Observability Types (`web/src/domain/live.ts`):
   - In `LiveAnomalyAlert`, change property `rule_name` from `LiveAnomalyRule | string` to strictly `LiveAnomalyRule`.

3. Frontend Shared Anomaly Types (`web/src/domain/anomalies.ts`):
   - Import `AnomalyRuleID` from `../types/telemetry` and `LiveAnomalyRule` from `./live`.
   - Export union type `AnomalyRule = AnomalyRuleID | LiveAnomalyRule`.
   - Export exact tool failure subset type `ToolFailureAnomalyRule = 'TOOL_RETRY_STREAK' | 'TOOL_HIGH_FAILURE_RATE' | 'ConsecutiveToolErrors' | 'ToolThrashing'`.
   - Re-export `AnomalyRule` and `ToolFailureAnomalyRule` from `web/src/domain/types.ts` and `web/src/components/TraceWaterfall.tsx`.

4. Tool Failure Anomaly Predicate Contract (`web/src/domain/anomalies.ts`):
   - Function signature: `isToolFailureAnomalyRule(ruleId: AnomalyRule | string): ruleId is ToolFailureAnomalyRule`.
   - Parameter type accepts `AnomalyRule` (or `string` for safe boundary and test compatibility).
   - Return type is a TypeScript type guard narrowing `ruleId` to `ToolFailureAnomalyRule`.
   - Implementation: Return true if and only if `ruleId` matches `'TOOL_RETRY_STREAK'`, `'TOOL_HIGH_FAILURE_RATE'`, `'ConsecutiveToolErrors'`, or `'ToolThrashing'`.
   - Remove string heuristics: delete calls to `.startsWith('TOOL_')` and `.includes('TOOL_FAILURE')`.
   - Callers in `matchesAnomaly` pass `anomaly.rule_id` (typed as `AnomalyRuleID`, which belongs to `AnomalyRule`).
   - Callers in `findMatchingLiveAlert` pass `alert.rule_name` (strictly typed as `LiveAnomalyRule`, which belongs to `AnomalyRule`). Both call sites become 100% type-checked at compile time.

5. Plain English Rule Translation Contract (`web/src/utils/debugPrompt.ts`):
   - Update `translateRuleToPlainEnglish(ruleId: AnomalyRule | string, toolName?: string): string`.
   - Explicitly type the input parameter to accept `AnomalyRule` alongside generic string fallback.

### OpenCode Database Telemetry Provenance Attribution Contract

All four prompt templates must explicitly incorporate the statement that telemetry was extracted from the OpenCode database within their narrative provenance paragraph:

1. Layer 1: Tool Execution Prompt (`generateToolExecutionPrompt`):
   - Narrative provenance sentence structure:
     `During turn [turn] of session [sessionId] ([provenance items]), the agent executed [toolName] and it [outcome]. This data was extracted from the OpenCode database.`
   - If turn is omitted, format prefix as:
     `In session [sessionId] ([provenance items]), the agent executed [toolName] and it [outcome]. This data was extracted from the OpenCode database.`

2. Layer 2: Execution Issue / Anomaly Card Prompt (`generateExecutionIssuePrompt`):
   - Narrative provenance sentence structure:
     `During turn [turn] of session [sessionId] ([agent details]), [description].[partTrigger] This data was extracted from the OpenCode database.`
   - If turn is omitted, format prefix as:
     `In session [sessionId] ([agent details]), [description].[partTrigger] This data was extracted from the OpenCode database.`

3. Layer 3: Message Turn Review Prompt (`generateMessageTurnPrompt`):
   - Narrative provenance sentence structure:
     `In turn [turn] of session [sessionId] ([message provenance]), the agent deliberated and executed tools. This data was extracted from the OpenCode database.`
   - If message provenance is absent, omit parenthetical.

4. Layer 4: Session Overview Prompt (`generateSessionOverviewPrompt`):
   - Narrative provenance sentence structure:
     `Session [sessionId] is a [role] subagent running [model] [parent session note]. The session is currently **[status]** after [turnCount] turns and [totalToolCalls] tool executions ([failedToolCalls] failed). This data was extracted from the OpenCode database.`

---

## 4. Phased Implementation Tasks for Coder

### Phase 1: Go Backend Domain Anomaly Rule Typing
- Task 1.1: In `backend/internal/domain/live_anomaly.go`:
  - Define `type LiveAnomalyRuleName string`.
  - Type constants `RuleConsecutiveToolErrors`, `RuleContextCeilingApproach`, `RuleTokenVelocityRunaway`, and `RuleToolThrashing` with `LiveAnomalyRuleName`.
  - Update `LiveAnomalyAlert` struct definition to use `RuleName LiveAnomalyRuleName`.
- Task 1.2: Run `go test -race ./...` and `go vet ./...` to verify zero compile or regression issues.

### Phase 2: TypeScript Anomaly Rule Typing and Predicate Refactoring
- Task 2.1: In `web/src/domain/live.ts`:
  - Update `LiveAnomalyAlert.rule_name` to strictly `LiveAnomalyRule`.
- Task 2.2: In `web/src/domain/anomalies.ts`:
  - Import `AnomalyRuleID` from `../types/telemetry` and `LiveAnomalyRule` from `./live`.
  - Define and export `type AnomalyRule = AnomalyRuleID | LiveAnomalyRule`.
  - Define and export `type ToolFailureAnomalyRule = 'TOOL_RETRY_STREAK' | 'TOOL_HIGH_FAILURE_RATE' | 'ConsecutiveToolErrors' | 'ToolThrashing'`.
  - Update `isToolFailureAnomalyRule(ruleId: AnomalyRule | string): ruleId is ToolFailureAnomalyRule`.
  - Replace body with strict equality against the four union members. Remove `.startsWith('TOOL_')` and `.includes('TOOL_FAILURE')`.
- Task 2.3: In `web/src/domain/types.ts`:
  - Re-export `AnomalyRule` and `ToolFailureAnomalyRule` from `./anomalies`.
- Task 2.4: In `web/src/components/TraceWaterfall.tsx`:
  - Re-export `AnomalyRule` and `ToolFailureAnomalyRule` alongside `isToolFailureAnomalyRule`.
- Task 2.5: In `web/src/utils/debugPrompt.ts`:
  - Update `translateRuleToPlainEnglish(ruleId: AnomalyRule | string, toolName?: string): string`.

### Phase 3: OpenCode Database Telemetry Attribution in Prompt Generators
- Task 3.1: In `web/src/utils/debugPrompt.ts`:
  - Update `generateToolExecutionPrompt` to append ` This data was extracted from the OpenCode database.` to `provenanceSentence`.
  - Update `generateExecutionIssuePrompt` to append ` This data was extracted from the OpenCode database.` to `provenanceSentence`.
  - Update `generateMessageTurnPrompt` to append ` This data was extracted from the OpenCode database.` to `provenanceSentence`.
  - Update `generateSessionOverviewPrompt` to append ` This data was extracted from the OpenCode database.` to `provenanceSentence`.

### Phase 4: Test Suite Updates
- Task 4.1: In `web/src/domain/anomalies.test.ts`:
  - Update `isToolFailureAnomalyRule` tests:
    - Assert `true` for all four `ToolFailureAnomalyRule` members (`TOOL_RETRY_STREAK`, `TOOL_HIGH_FAILURE_RATE`, `ConsecutiveToolErrors`, `ToolThrashing`).
    - Assert `false` for non-tool `AnomalyRule` members (`CONTEXT_SATURATION_SPIKE`, `SUBAGENT_REWORK_LOOP`, `ContextCeilingApproach`, `TokenVelocityRunaway`).
    - Remove assertions expecting arbitrary unmapped strings with `TOOL_` prefix to evaluate to true.
    - Assert `false` for unknown rule strings and empty strings.
    - Verify TypeScript type guard narrowing on `ToolFailureAnomalyRule`.
- Task 4.2: In `web/src/domain/live.test.ts`:
  - Update test `supports custom rule identifier strings` or align with strict `LiveAnomalyRule` typing.
- Task 4.3: In `web/src/utils/debugPrompt.test.ts`:
  - Add explicit assertions verifying `expect(prompt).toContain('This data was extracted from the OpenCode database.');` across all four layers.
  - Verify all existing sanitization, edge case, and boundary tests pass cleanly.
- Task 4.4: In component tests (`LiveTraceInspector.test.tsx`, `FailedToolAlertCard.test.tsx`, `AnomalyDrawer.test.tsx`, `TraceWaterfall.test.tsx`):
  - Ensure all assertions remain green.

### Phase 5: Quality Gates and Verification
- Task 5.1: Run `npm run typecheck` to confirm zero TypeScript compilation errors.
- Task 5.2: Run `npm run lint` to verify code style and formatting standards.
- Task 5.3: Run `npm test` across all 30 frontend test suites.
- Task 5.4: Run `go test -race ./...` and `go vet ./...` to verify Go backend integrity.
- Task 5.5: Confirm zero emojis across all modified files and test fixtures.
