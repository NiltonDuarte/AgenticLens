# AgenticLens Fallback Architecture Audit: Comprehensive System-Wide Evaluation

## 1. Executive Summary

This audit catalogs and critiques every fallback cascade ("If A is not provided, try B, else C") across AgenticLens. The audit spans the Go 1.23+ backend (`cmd/agentic-lens`, `internal/domain`, `internal/adapters`), the TypeScript 5.x web dashboard (`web/src`), and developer shell automation scripts.

AgenticLens maintains rigorous Functional Core, Imperative Shell (FCIS) separation and strict type-safety with zero `any` or `interface{}` across Go and TypeScript. However, fallback cascades—designed to preserve system operation across varied environments (local CLI, Docker containers, multi-agent CI/CD, heterogeneous Git repositories)—have evolved independently across different packages. 

This audit identified six concrete risks and architectural blind spots. Based on steering feedback, two core architectural directives have been established:
1. **Zero-Fallback Completion Date Mandate**: `completed_at` must always be present and non-zero when telemetry is generated. Telemetry generation must validate this field and fail if missing. Domain aggregations and frontend analytics must **never fall back** to `updated_at` or `started_at` for ticket completion.
2. **Immediate Tier 4 Session Resolver Deletion**: Tier 4 ("most recent message in SQLite DB") must be deleted entirely. Session resolution must fail fast if explicit identifiers (nonce, session ID, ticket title/directory) cannot be resolved.
3. **Frontend Silent Fallback Elimination**: `web/src/views/TicketDeepDive.tsx` must remove `ledger.ticket_type ?? 'feature'` and explicitly render `'unclassified'` with warning styling.
4. **Hardcoded Build Version Synchronization**: Synchronize fallback version strings between Go backend (`domain.DefaultVersion = "v0.0.0-dev"`) and TypeScript frontend (`Navbar.tsx: buildInfo?.version ?? 'v0.0.0-dev'`).
5. **Leaky Container Path Sniffing De-prioritization**: `adapters.ResolveWorkspacesDir` must evaluate configured project paths (`ProjectConfig.Workspaces.Dir`) before container paths (`/app/workspaces`).
6. **Dual-Path Query Parameter Normalization**: Deprecate alias query parameters (`startDate`/`endDate`) in favor of canonical `since`/`until`.

---

## 2. Comprehensive Fallback Matrix

The table below catalogs all 24 distinct fallback chains discovered across the platform, ranked by resilience grade (A: Deterministic & safe; B: Acceptable with minor drift; C: Fragile / leaky; D: Active risk; E: Slated for removal).

| # | Subsystem / Location | Source Code Anchor | Fallback Cascade ("If A -> B -> C") | Resilience Grade | Architectural Status / Directive |
|---|---|---|---|:---:|---|
| 1 | Configuration File Discovery | `backend/internal/adapters/config/loader.go:FindConfigFile` | 1. Explicit path parameter<br>2. Root `.agenticlens-config.yaml`<br>3. Root `.agenticlens-config.yml`<br>4. Legacy `.agenticlens.yaml`<br>5. Legacy `.agenticlens.yml`<br>6. Return `""` (use struct defaults) | **A** | **Retain**: Smooth migration from legacy filenames to root YAML; deterministic. |
| 2 | Configuration Struct Defaults | `backend/internal/adapters/config/loader.go:DefaultProjectConfig` | 1. User YAML values<br>2. Hardcoded defaults (`ProjectName: "AgenticLens"`, `WorkspacesDir: "workspaces"`, `BaseBranch: "main"`, `Prefixes: ["AL", "GEN", "PO"]`, `CostRates: DefaultModelRates`) | **A** | **Retain**: Immutable base configuration guarantees non-nil runtime state. |
| 3 | OpenCode Root Session Resolution | `backend/internal/adapters/opencode/resolver.go:ResolveActiveRootSession` | 1. Telemetry nonce in message part<br>2. Explicit root session ID<br>3. Ticket title / workspace folder in message parts<br>4. ~~Most recent message in database~~<br>5. Fail with error | **E** | **Drop Tier 4 / Fail Fast**: Delete database fallback. If tiers 1-3 fail, return `ErrSessionNotFound` immediately to prevent concurrent agent session contamination. |
| 4 | OpenCode Session Agent Inference | `backend/internal/adapters/opencode/extractor.go:extractAgentInvocations` | 1. Agent name on message part<br>2. Role `orchestrator` if root session<br>3. Role `unassigned` if subagent session | **B** | **Retain**: Predictable fallback conforming to FCIS multi-agent model. |
| 5 | OpenCode Model ID Extraction | `backend/internal/adapters/opencode/extractor.go:extractAgentInvocations` | 1. Model identifier on message part<br>2. Default model on session record<br>3. Fallback string `"default"` | **B** | **Retain**: Prevents empty strings; pricing calculations cleanly fallback to `DefaultCostRates`. |
| 6 | OpenCode Message Timestamp | `backend/internal/adapters/opencode/extractor.go:parseTimeField` | 1. Millisecond integer timestamp<br>2. RFC3339 string timestamp<br>3. Fallback to Unix Epoch (zero `time.Time`) | **B** | **Retain**: Defensive unmarshaling against heterogeneous SQLite timestamp formats. |
| 7 | Git Author / Committer Resolution | `backend/internal/adapters/git/inference.go:GetGitUser` | 1. `git config user.name`<br>2. `git config user.email`<br>3. Combined `"Name <email>"`<br>4. Fallback string `"unknown"` | **A** | **Retain**: Never panics in detached or headless CI containers. |
| 8 | Git Base Branch Resolution | `backend/internal/adapters/git/analyzer.go:ResolveBaseBranch` | 1. Configured `base_branch`<br>2. Target of `refs/remotes/origin/HEAD`<br>3. Probe `origin/main` -> `main` -> `origin/master` -> `master`<br>4. Fail with explicit descriptive error | **B** | **Retain**: Heuristic probing covers common repository naming conventions. |
| 9 | Git Numstat Diffing Strategy | `backend/internal/adapters/git/analyzer.go:GetTicketNumstat` | 1. Merge-base diff (`base...HEAD`)<br>2. Two-dot diff (`base..HEAD`)<br>3. Worktree diff (`git diff base`)<br>4. Staged diff (`git diff --staged base`)<br>5. Return zero stats `(0, 0, 0)` | **A** | **Retain**: Comprehensive fallback handles detached heads, squashed rebases, and clean working trees. |
| 10 | Ticket Type Inference (4-Tier) | `backend/internal/adapters/git/inference.go:InferTicketType` | 1. Explicit CLI `--ticket-type`<br>2. Conventional branch prefix (`feat/`, `fix/`, etc.)<br>3. Ticket prefix mapping (`PO-*` -> `docs`, `GEN-*` -> `chore`)<br>4. Return `nil` (unclassified / prompt required) | **A** | **Retain**: Strict, eliminates arbitrary git log grepping; leaves unclassified as explicit `nil`. |
| 11 | Workspaces Directory Resolution | `backend/internal/adapters/repository.go:ResolveWorkspacesDir` | 1. Non-empty override argument<br>2. Container path `/app/workspaces`<br>3. Configured `cfg.Workspaces.Dir`<br>4. Local `./workspaces`<br>5. Walk up parent directories for `workspaces`<br>6. Fallback string `"workspaces"` | **C** | **Reorder**: Prioritize `cfg.Workspaces.Dir` over `/app/workspaces` so host configurations are never hijacked by container paths. |
| 12 | Ticket ID Inference in Scanner | `backend/internal/adapters/repository.go:ScanWorkspaces` | 1. `ledger.TicketID` from JSON file<br>2. Parent directory basename | **A** | **Retain**: Enables scanning legacy or unversioned directories lacking explicit ID keys. |
| 13 | Telemetry Markdown Header Detection | `backend/cmd/agentic-lens/main.go:updateSummaryFile` | 1. Existing `"### Workflow Efficiency"`<br>2. Existing `"### Agentic Execution Metrics"`<br>3. Append new section to end of file | **B** | **Retain**: Preserves backward compatibility with legacy markdown templates. |
| 14 | Telemetry Ledger Schema Detection | `backend/internal/domain/migration/engine.go:DetectVersion` | 1. `schema_version` integer in JSON<br>2. If null, 0, or <= 1 -> fallback to Schema Version 1 | **A** | **Retain**: Cleanly detects legacy unversioned ledgers and triggers progressive upgrade pipeline. |
| 15 | Ledger Deserialization Fallback | `backend/internal/adapters/repository.go:LoadExistingFactLedger` | 1. Atomic Fact Schema (`facts.events`, `facts.agent_invocations`)<br>2. Legacy Thread Schema (`thread.turns`, `workflow_efficiency`) | **A** | **Retain**: Dual-schema unmarshaling prevents data loss when scanning historical workspaces. |
| 16 | Ticket Completion Date Resolution | `backend/internal/domain/aggregations.go` & `formatters.ts` | ~~1. `completed_at`<br>2. `updated_at`<br>3. `started_at`~~ | **E** | **Abolish Fallback Chain**: `completed_at` must be strictly present on telemetry generation. Zero fallback to `updated_at` or `started_at`. Ledgers lacking `completed_at` are treated as uncompleted. |
| 17 | Anomaly Detection Tool Telemetry | `backend/internal/domain/anomalies.go:detectToolFailureLoops` | 1. Schema v2 `ledger.Facts.Events`<br>2. Fallback to v1 `ledger.Facts.Tools` | **A** | **Retain**: Dual-schema rule engine handles both legacy and migrated ledgers seamlessly. |
| 18 | Context Saturation Anomaly Detection | `backend/internal/domain/anomalies.go:detectContextWindowSaturation` | 1. Breach events in `ledger.Facts.Events`<br>2. Fallback to `inv.Tokens.PeakContext >= 160,000` | **A** | **Retain**: Accurately identifies bloated context in ledgers lacking fine-grained token events. |
| 19 | Live Anomaly Ceiling Token Fallback | `backend/internal/domain/live_anomaly.go:AssessLiveWindow` | 1. Configured model ceiling<br>2. Fallback `HealthyCeilingTokens = 200,000` | **B** | **Retain**: Provides baseline safety threshold when model registry metadata is absent. |
| 20 | Live Anomaly Error Snippet Fallback | `backend/internal/domain/live_anomaly.go:AssessLiveFailures` | 1. Non-empty `part.ErrorMessage`<br>2. Non-empty `part.OutputSnippet`<br>3. Fallback string `"part <ID>"` | **B** | **Retain**: Ensures descriptive anomaly alert message even if output snippet was truncated. |
| 21 | Live Workspace Events Dual-Plane API | `backend/internal/adapters/http/handlers.go:WorkspaceEventsHandler` | 1. Live SQLite database hydration (`IsHydrated: true`, `DataSource: "live_sqlite"`)<br>2. Committed ledger fallback (`IsHydrated: false`, `DataSource: "committed_ledger"`) | **A** | **Retain**: Zero-downtime dual-plane resilience: works seamlessly without SQLite access. |
| 22 | HTTP Query Parameter Aliases | `backend/internal/adapters/http/handlers.go:FleetHandler` | 1. `since` parameter -> fallback `startDate`<br>2. `until` parameter -> fallback `endDate`<br>3. `window` parameter -> fallback `"all"`<br>4. `sort` parameter -> fallback `"updated_at"` | **B** | **Normalize**: Retain for backward compatibility, but document `since`/`until` as canonical. |
| 23 | Web UI Live Stream Connection | `web/src/api/client.ts:connectWorkspaceEventsLive` | 1. Server-Sent Events (SSE) `/api/live/workspaces/:id`<br>2. Fallback to 3000ms periodic HTTP polling upon SSE disconnect/error | **A** | **Retain**: Rock-solid frontend fallback for environments behind aggressive proxies or buffering firewalls. |
| 24 | Web UI Ticket Type Badge | `web/src/views/TicketDeepDive.tsx:221` | 1. `ledger.ticket_type`<br>2. ~~Fallback `'feature'`~~ | **D** | **Eliminate Fallback**: Remove default `'feature'`. Render `'unclassified'` with explicit neutral badge styling. |

---

## 3. Deep Dive: Architectural Directives & Concrete Mandates

### 3.1 Directive 1: Zero-Fallback Completion Date Mandate

#### Current Problem
Across the backend and frontend, multiple date resolution functions implement conflicting fallback chains:
- `backend/internal/domain/aggregations.go:210` (`GetLedgerCompletionDate`): `completed_at` -> `updated_at` -> `started_at`
- `backend/internal/domain/aggregations.go:1022` (`GetTicketEffectiveDate`): `completed_at` -> `started_at` -> `updated_at`
- `web/src/domain/formatters.ts:16` (`getEffectiveCompletionDateMs`): `completed_at` -> `updated_at` -> `started_at`
- `web/src/domain/progression.ts:60` (`getEffectiveDate`): `completed_at` -> `started_at` -> `updated_at`

This creates temporal corruption: an in-progress ticket that started on Monday and was modified on Friday lands on Friday in Fleet Overview, but on Monday in Progression trends. Furthermore, treating `updated_at` or `started_at` as completion falsifies ticket lifecycle metrics.

#### Directive & Architectural Invariant
1. **Mandatory Telemetry `completed_at`**: When telemetry is generated (`cmd/agentic-lens generate`), the engine must enforce that `completed_at` is populated and non-zero. If the OpenCode session lacks a completion timestamp, the generator must record the current finalization time or halt with a validation error.
2. **Zero Fallback**: In domain aggregations and frontend charts, **never fall back** to `updated_at` or `started_at`.
   - `GetLedgerCompletionDate`:
     ```go
     func GetLedgerCompletionDate(ledger AgenticFactLedger) (time.Time, bool) {
         if !ledger.CompletedAt.IsZero() {
             return ledger.CompletedAt, true
         }
         return time.Time{}, false
     }
     ```
   - Filtering tickets by completion date strictly includes ledgers where `CompletedAt` is non-zero. Tickets without `completed_at` are classified as in-progress or unfinalized.
   - Frontend `getEffectiveCompletionDateMs` and `progression.ts` must align directly with this rule, removing fallback logic completely.

---

### 3.2 Directive 2: Drop Session Resolver Tier 4 (Fail Fast)

#### Current Problem
In `backend/internal/adapters/opencode/resolver.go`:
```go
// Tier 4: Most recent message in SQLite DB
row := db.QueryRowContext(ctx, `
    SELECT session_id FROM messages 
    WHERE session_id IS NOT NULL 
    ORDER BY created_at DESC LIMIT 1
`)
```
If a ticket title does not match workspace folder names and no `--nonce` or `--session-id` was provided, the generator binds to whatever session wrote to SQLite last. In parallel or multi-agent execution environments, this captures an arbitrary foreign session and attaches its telemetry, tokens, tool failures, and git diff to the wrong ticket.

#### Directive & Architectural Invariant
1. **Delete Tier 4 Entirely**: Remove the query `SELECT session_id FROM messages ... ORDER BY created_at DESC LIMIT 1`.
2. **Fail Fast**: When Tiers 1-3 fail (no nonce match, no explicit `--session-id`, no workspace directory/title match), immediately return:
   ```go
   return "", fmt.Errorf("%w: unable to resolve OpenCode session for ticket %q; specify --session-id or use --nonce", ErrSessionNotFound, ticketID)
   ```
3. This guarantees deterministic, isolated telemetry generation with zero cross-ticket contamination.

---

### 3.3 Directive 3: Eliminate Frontend Ticket Type Masking

#### Current Problem
In `web/src/views/TicketDeepDive.tsx:221`:
```tsx
<span className="ticket-type-pill">{ledger.ticket_type ?? 'feature'}</span>
```
Ticket AL-29 removed silent fallbacks to `feature` in the Go backend, surfacing `unclassified` tickets. However, the UI continues to mask `null`/`undefined` ticket types as `'feature'`, presenting unclassified tickets as features.

#### Directive & Architectural Invariant
Replace the fallback with explicit unclassified handling:
```tsx
<span className={`ticket-type-pill ${ledger.ticket_type ? '' : 'unclassified'}`}>
  {ledger.ticket_type ?? 'unclassified'}
</span>
```
Add CSS styling for `.ticket-type-pill.unclassified` (neutral slate gray with a warning border) to make unclassified status visible to the user.

---

### 3.4 Directive 4: De-prioritize Leaky Container Path Sniffing

#### Current Problem
In `backend/internal/adapters/repository.go:ResolveWorkspacesDir`:
```go
if strings.TrimSpace(overrideDir) != "" {
    return overrideDir
}
// Docker container mount check
if _, err := os.Stat("/app/workspaces"); err == nil {
    return "/app/workspaces"
}
```
If a host machine or CI runner happens to have an `/app/workspaces` path, it takes precedence over the user's `.agenticlens-config.yaml` (`cfg.Workspaces.Dir`).

#### Directive & Architectural Invariant
1. Read configuration first: `cfg.Workspaces.Dir` must be checked before probing `/app/workspaces`.
2. Restrict container path sniffing to verified container environments (e.g., verifying `/.dockerenv` or container cgroups).

---

### 3.5 Directive 5: Synchronize Platform Build Version Defaults

#### Current Problem
- Backend constant: `domain.DefaultVersion = "v0.0.0-dev"` (`backend/internal/domain/types.go:17`)
- Frontend default: `buildInfo?.version ?? 'v0.1.0-dev'` (`web/src/components/Navbar.tsx:75`)

#### Directive & Architectural Invariant
Standardize both backend and frontend fallback versions on `v0.0.0-dev`.

---

## 4. Remediation Roadmap & Implementation Tickets

The following vertical implementation tickets are prepared to execute these directives:

### Ticket 1: REV-FIX-1 — Enforce Mandatory `completed_at` in Telemetry & Dismantle Date Fallback Chains
- **Scope**: Backend CLI, Domain, Adapters, and TypeScript Frontend
- **Actions**:
  1. Update `backend/cmd/agentic-lens/main.go` (`generateCmd`): validate that `ledger.CompletedAt` is non-zero. If zero, assign current generation timestamp before serialization.
  2. Refactor `backend/internal/domain/aggregations.go`:
     - Update `GetLedgerCompletionDate` to only return `CompletedAt` (zero fallback to `updated_at` or `started_at`).
     - Remove `GetTicketEffectiveDate` or deprecate its fallback ordering in favor of strict `CompletedAt`.
  3. Refactor `web/src/domain/formatters.ts` and `web/src/domain/progression.ts`: eliminate date fallback chains; require `completed_at`.
  4. Update domain unit tests in Go and Vitest asserting zero fallback behavior.

### Ticket 2: REV-FIX-2 — Drop Session Resolver Tier 4 & Fail Fast
- **Scope**: Backend Adapters (`backend/internal/adapters/opencode/resolver.go`)
- **Actions**:
  1. Delete Tier 4 query from `ResolveActiveRootSession`.
  2. Return explicit `ErrSessionNotFound` when Tiers 1-3 fail.
  3. Update `resolver_test.go` to assert fast failure instead of binding to the latest database message.

### Ticket 3: REV-FIX-3 — Eliminate UI Ticket Type Fallback to `'feature'`
- **Scope**: TypeScript Frontend (`web/src/views/TicketDeepDive.tsx`, CSS)
- **Actions**:
  1. Replace `{ledger.ticket_type ?? 'feature'}` with `{ledger.ticket_type ?? 'unclassified'}`.
  2. Add `.ticket-type-pill.unclassified` styling.
  3. Add Vitest coverage verifying unclassified pill rendering when `ticket_type` is undefined or null.

### Ticket 4: REV-FIX-4 — Fix Workspaces Container Directory Sniffing Precedence
- **Scope**: Backend Adapters (`backend/internal/adapters/repository.go`)
- **Actions**:
  1. Reorder resolution in `ResolveWorkspacesDir`: CLI override -> `cfg.Workspaces.Dir` -> `./workspaces` -> container probe `/app/workspaces`.
  2. Gate container probe with check for container environment.

### Ticket 5: REV-FIX-5 — Harmonize Default Version Strings
- **Scope**: TypeScript Frontend (`web/src/components/Navbar.tsx`)
- **Actions**:
  1. Change default version fallback in `Navbar.tsx` from `'v0.1.0-dev'` to `'v0.0.0-dev'`.
