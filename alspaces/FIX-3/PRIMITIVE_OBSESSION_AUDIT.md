# Primitive Obsession Audit and Domain Typing Strategy

## Executive Summary

Following the completion of FIX-3 Iteration 04, a systematic audit of `backend/internal/domain` and `web/src/` was conducted to assess type safety across domain models. 

While recent iterations successfully established strict types for anomaly rules (`LiveAnomalyRuleName`, `LiveAnomalyRule`, `ToolFailureAnomalyRule`), significant primitive obsession persists across the rest of the codebase. Specifically, core entities, lifecycle statuses, health ratings, and roles are predominantly represented as unconstrained scalar types (`string`, `*string`, `int64`).

---

## Findings Matrix

### 1. Dangling Enums (Declared Types Bypassed for Raw Strings)

Several domain enums exist in `backend/internal/domain/enums.go` and `web/src/domain/types.ts`, but the structs situated directly alongside them continue to use raw `string`.

| Domain Concept | Declared Enum / Type | Struct Fields Using Raw `string` |
|---|---|---|
| **Context Health** | Go: `ContextHealthStatus`<br>TS: `ContextHealthStatus` | • `domain.GroupedMetric.ContextHealth` (`string`)<br>• `domain.ContextWindowHealth.Status` (`string`)<br>• `domain.ModelMetric.ContextHealth` (`string`)<br>• `domain.RoleMetric.ContextHealth` (`string`)<br>• `domain.TicketComparisonRow.ContextHealthStatus` (`string`)<br>• TS `GroupedMetric.context_health` (`string`)<br>• TS `ContextWindowHealth.status` (`string`)<br>• TS `ModelMetric.context_health` (`string`)<br>• TS `RoleMetric.context_health` (`string`)<br>• TS `TicketComparisonRow.context_health_status` (`string`) |
| **Tool / Part Failure Severity** | Go: `FailureSeverity` (`SeverityNone`, `SeverityLow`, `SeverityMedium`, `SeverityHigh`)<br>TS: `LiveAnomalySeverity` | • `domain.LivePartFact.Severity` (`string`)<br>• `domain.LiveAnomalyAlert.Severity` (`string`)<br>• TS `LivePartFact.severity` (`string`) |
| **Agent Role** | Go: `AgentRole`<br>TS: `type AgentRole = string` *(no-op alias)* | • `domain.InvocationFact.Role` (`string`)<br>• `domain.EventFact.Role` (`string`)<br>• `domain.LiveMessageFact.Role` (`string`)<br>• `domain.RoleMetric.Role` (`string`)<br>• `domain.ThreadRoleMetric.Role` (`string`)<br>*(Only `LiveSessionNode.Role` utilizes `AgentRole`)* |
| **Conventional Ticket Type** | Go: `ConventionalTicketType`<br>TS: `StandardTicketType` | • `domain.AgenticMetricsLedger.TicketType` (`*string`)<br>• `domain.TicketComparisonRow.TicketType` (`*string`)<br>• TS `AgenticMetricsLedger.ticket_type` (`string \| null`)<br>• TS `TicketComparisonRow.ticket_type` (`StandardTicketType \| string \| null`) |

---

### 2. Missing Entity Identifiers (Stringly-Typed Parameters)

Domain entities share the same underlying representation (`string`), leaving critical function signatures vulnerable to argument transpositions that the compiler cannot detect:

- **`TicketID`**: Handled as `string` / `*string` throughout all ledger aggregations, comparison tables, and artifact parsers.
- **`SessionID`**: Handled as `string` across `EventFact`, `ThreadExecutionRecord`, `LiveSessionNode`, `LiveTraceView`, and `LiveAnomalyAlert`.
- **`MessageID` & `PartID`**: Handled as `string` / `*string` across `EventFact`, `LiveMessageFact`, and `LivePartFact`.
- **`ModelName`**: Handled as raw `string` across all metric groupings and node representations.

**Concrete Example of Vulnerability:**
In `backend/internal/adapters/live` and `web/src/utils/debugPrompt.ts`, helper methods accept consecutive string parameters:
```go
// Ambiguous signature: sessionId and messageId can be accidentally swapped
func BuildTurnKey(sessionId string, messageId string, partId string) string
```
With distinct types (`SessionID`, `MessageID`, `PartID`), the Go compiler and TypeScript type checker reject inverted argument orders at build time.

---

### 3. Unconstrained Status Strings

- **Execution Status**: `EventFact.Status` and `LivePartFact.Status` use raw strings. Code checks against `"completed"` and `"error"`, but typos like `"complete"` or `"errored"` fail silently at runtime.
- **Session Lifecycle Status**: `LiveSessionNode.Status` uses raw `string` instead of a union of known lifecycle states (`"active" | "idle" | "completed"`).
- **Execution Part Type**: `LivePartFact.Type` uses `string`, whereas `EventFact.EventType` has a defined enum (`tool_call`, `turn_message`, `compaction`).

---

## Risks and Blind Spots

1. **JSON Serialization & Backward Compatibility**:
   - In Go, creating a named type based on string (`type TicketID string`) preserves JSON marshalling without wire-format changes.
   - However, using pointer fields (`*TicketID` vs `*string`) requires careful conversion when reading from external YAML/JSON or database rows (`sql.NullString`).
2. **Open-Ended Agent Fleets vs Closed Enums**:
   - While canonical agent roles are known (`orchestrator`, `planner`, `coder`, `tester`, `reviewer`), operators can run custom subagent roles.
   - In TypeScript, closing `AgentRole` as a strict union (`'orchestrator' | 'planner' | ...`) will cause type errors or UI drops when third-party or custom agents appear in telemetry.
   - *Mitigation:* Use loose autocomplete unions: `type AgentRole = CanonicalRole | (string & {})`.
3. **Branded Types Cognitive & Testing Overhead**:
   - Introducing nominal/branded types (`type TicketId = string & { readonly __brand: unique symbol }`) in TypeScript provides compile-time isolation, but forces the creation of constructor assertions (`asTicketId(...)`) across hundreds of test fixtures and UI mock data files.
   - *Mitigation:* Use semantic type aliases first (`type TicketId = string; type SessionId = string;`) or introduce branded types only at domain boundaries.
4. **Scope Creep on FIX-3**:
   - FIX-3 is an open bugfix PR (#24) with all 4 iterations tested, verified, and ready. Conflating FIX-3 with a broad domain refactor creates an oversized diff, increases regression surface, and obscures git history.

---

## Recommended Action Plan

- **Track A (Complete FIX-3)**: Finalize, merge, or link PR #24 as-is. FIX-3's contracted scope (false tool errors, waterfall collapsibility, debug identifiers, AI prompts) is 100% delivered.
- **Track B (Dangling Enums Sweep)**: Execute an incremental refactoring ticket to substitute existing declared enums (`ContextHealthStatus`, `FailureSeverity`, `AgentRole`) into domain models.
- **Track C (Strong Entity Identifiers)**: Formulate a dedicated architectural ticket (`REF-1`) for domain-driven identifier types (`TicketID`, `SessionID`, `MessageID`, `PartID`) across backend adapters, scanner, and frontend.
