# Issue: Conflation of Root Session Hierarchy with Persona Role

## 1. Context and Architectural Root Cause
In AgenticLens, session telemetry is organized as a hierarchy: a root session coordinates the user workflow, and child invocations (`parent_id != nil`) represent delegated subagent tasks (`planner`, `coder`, `tester`, `reviewer`).

However, the domain metrics engine and extraction adapters conflated two orthogonal concepts:
- **Session Hierarchy (Structural)**: Whether an invocation is the root session (`ParentID == nil`) or a delegated child session (`ParentID != nil`).
- **Agent Persona (Semantic)**: The specific agent persona executing the session (`general`, `orchestrator`, `build`, `coder`, `tester`, etc.).

---

## 2. The Flawed Assumption
The existing implementation assumed that:
1. The root session is always named `"orchestrator"`.
2. Any role whose name is not strictly `"orchestrator"` is a subagent (for example, `domain.IsSubagentRole(role)` checked `role != "orchestrator"`).

This assumption is invalid:
- A user interacting directly with OpenCode runs in a root session whose role is `"general"`.
- A task initiated under any specialized persona operates in a root session carrying that persona name.
- Root sessions perform direct coordination and general conversational tasks without being subagent delegations.
- Any role must be accepted as a valid root role without being misclassified as a subagent.

---

## 3. How the Regression Manifested
When commit `f57d680` executed the batch backfill, `ExtractFacts` preserved the raw persona stored in OpenCode's SQLite database (`agent = "general"` for root sessions):

1. **Root Session Misclassified as Subagent**: Because `IsSubagentRole()` checked `role != "orchestrator"`, the root `"general"` session was categorized as a delegated subagent.
2. **Subagent Peak Context Inversion**: The root session's peak context tokens (e.g., 123.3k tokens in AL-1) were attributed to `SubagentsPeakTokens` with `SubagentsPeakRole = "general"`, rather than measuring actual delegated subagents.
3. **Root Coordination Metrics Zeroed**: Metrics tracking the root session (`OrchestratorPeakTokens`, saturation percentages) dropped to `0` because they specifically looked up the key `"orchestrator"` instead of inspecting the root invocation.
4. **Subagent Effort Distortion**: The root session's token consumption was included in subagent effort distribution percentages, diluting the proportions of the actual subagents (`coder`, `planner`, `tester`, `reviewer`).
5. **Test Failure in `parity_test.go`**: Historical assertions in `backend/internal/adapters/parity_test.go` hardcoded the old assumption where root sessions were forced to `"orchestrator"`. When compared against ledgers where the root session is `"general"` (and misclassified as a subagent), 8 test cases failed.

---

## 4. Requirements for Remediation

### 4.1 Decouple Structure from Persona
- Determine whether an invocation is a root session strictly by checking `ParentID == nil` (or matching `RootThreadIDs`), never by checking `Role == "orchestrator"`.
- Ensure that root sessions are never counted or interpreted as subagents, regardless of their role name (`general`, `orchestrator`, `build`, or custom personas).

### 4.2 Support Any Role as Root Role
- The root session can be `"general"`, `"orchestrator"`, or any user-defined persona. Configuration should allow mapping or preserving root roles cleanly.
- Root coordination metrics (peak context, saturation, active duration) must measure the root invocation regardless of its persona name.
- Subagent metrics (subagent peak role, subagent turn velocity, subagent effort distribution) must strictly filter to invocations where `ParentID != nil`.

### 4.3 Update `parity_test.go` Baselines
- Update baseline fixtures and assertions in `backend/internal/adapters/parity_test.go` to match the corrected domain calculations and the authentic root role (`"general"`).
