---
name: agentic-lens-review
description: Use when auditing, reviewing, or finding refactoring, performance, and architectural improvement opportunities in the AgenticLens Go and TypeScript codebase, with deep test suite profiling for slow tests.
---

# AgenticLens Architecture, Code & Test Suite Review

Audits the **AgenticLens** codebase (Go backend/CLI/engine and TypeScript frontend) for architectural drift, code smells, shallow abstractions, and test suite execution bottlenecks. Synthesizes deep module design with strict Functional Core, Imperative Shell (FCIS) invariants to deliver prioritized, actionable refactoring and performance optimization proposals.

---

## Core Review Principles

- **Subtraction Before Addition**: Always look for what can be deleted first (dead code, speculative wrappers, redundant adapters, unused interfaces) before proposing new abstractions.
- **Deep Modules Over Passthroughs**: A module should provide significant capability behind a compact interface. Flag shallow wrappers that just delegate calls without compressing complexity.
- **Minimize Reader Load**: Measure maintainability by mental tracing cost. Can a developer answer "where does X come from?" and "what computes X?" in under 30 seconds? If not, flag for simplification.
- **Domain Modeling Over Scattered Conditionals**: Replace ad-hoc boolean flags, status strings, and nested conditionals with typed value structs, string enums, and exhaustive type switches in Go, and Discriminated Unions in TypeScript.
- **Microsecond Pure Domain SLA**: Pure domain functions in `internal/domain` must execute in microseconds without disk, network, or clock side effects. Tests must run deterministically with zero artificial delays.

---

## Audit Workflow

### 1. Scope & Hotspot Discovery
Determine the audit boundary before reading files:
1. If the user specifies a subsystem, package, or ticket (e.g. `internal/domain`, `internal/adapters`, `web/`, or a workspace diff), focus on that scope.
2. Otherwise, identify churn hotspots:
   ```bash
   git log --oneline -n 50 --stat
   ```
3. Read baseline documents:
   - `TECHNICAL_GUIDELINES.md` (Go & TypeScript architectural invariants, FCIS, typing, `embed.FS`)
   - `.agenticlens/config.json` (Project parametrization)

---

### 2. Multi-Tier Inspection

Execute four inspection tiers against the target scope:

#### Tier 1: Layering & Boundary Discipline (Go + TypeScript FCIS)
- **FCIS Isolation Breach**: Does `internal/domain` import any outer layer (`internal/adapters`, `cmd/`) or I/O library (`os`, `net/http`, `database/sql`, `path/filepath`)? Reject immediately.
- **Impure Domain Functions**: Do functions in `internal/domain` call `time.Now()`, generate random numbers, or access global mutable variables instead of receiving values as immutable arguments?
- **Boundary Model Leakage**: Are raw JSON dictionaries (`map[string]any`) leaking into domain functions? External JSON must be unmarshaled and validated in `internal/adapters` into immutable value structs before passing to domain logic.
- **Stateless Handlers & Scanner**: Do adapter structs or HTTP handlers accumulate mutable runtime state between calls instead of remaining strictly stateless?

#### Tier 2: Module Depth & Structural Friction
- **Shallow Passthroughs & One-Caller Wrappers**: Are there Go functions or TypeScript helpers that merely forward arguments to another standard library or third-party function without adding validation, domain conversion, or caching? Flag for inlining or deletion.
- **Unnecessary Interfaces**: Are Go interfaces defined in producer packages or declared with only one implementation before any second implementation exists? In Go, define small interfaces where they are consumed. Flag speculative interfaces.
- **The Deletion Test**: If a helper or adapter is deleted, does it simplify the system or scatter complexity? If deleting simplifies the codebase, recommend deletion first.
- **Speculative Generality**: Are there unneeded Go generics (`[T any]`), extra hook callbacks, or hypothetical configuration options not backed by actual telemetry requirements?

#### Tier 3: Domain Modeling & Fowler Code Smells
- **Primitive Obsession**: Are raw strings or integers used for core domain concepts (e.g. ticket IDs, model names, token counts, health categories) instead of dedicated typed string/int types (e.g. `type TicketID string`, `type HealthStatus string`)?
- **Feature Envy**: Does a method in package A constantly reach into and manipulate fields of struct B? Move the calculation onto struct B or into `internal/domain/aggregations`.
- **Data Clumps**: Do the same group of token counters (input, output, cache read/write, reasoning) travel together as separate function parameters? Bundle them into a dedicated value struct (`TokenMetrics`).
- **Missing Discriminated Unions / Sum Types**: Are multi-state UI results or Go workflow results represented with multiple optional pointers and boolean flags instead of typed state structs or TypeScript Discriminated Unions?
- **Shotgun Surgery**: Does adding a new metric dimension require coordinated edits across multiple disconnected files? Gather metric computation into cohesive domain aggregations.

#### Tier 4: Test Suite Performance, Execution Speed & Flakiness
- **Millisecond Profiling**: Run test suite profiling to identify all tests exceeding 100ms:
  ```bash
  go test -json ./... | jq -r 'select(.Action=="pass" or .Action=="fail") | select(.Test != null) | "\(.Elapsed)s \(.Package) \(.Test)"' | sort -nr | head -20
  ```
- **Missing Test Parallelism (`t.Parallel()`)**:
  - Are independent unit tests and table-driven subtests executing serially?
  - Verify `t.Parallel()` is applied to independent test functions and inside `t.Run()` loops.
- **Artificial Delays & Sleeps**:
  - Search for `time.Sleep(...)` across `*_test.go` and `web/**/*.test.ts`.
  - Banish `time.Sleep`. Replace with deterministic channel synchronization (`select { case <-done: ... case <-time.After(50*time.Millisecond): ... }`), mutex conditions, or polled assertions.
- **Disk & SQLite I/O Churn**:
  - Flag tests generating physical on-disk files or temporary database files instead of in-memory representations (`:memory:` or memory-backed filesystems).
- **Subprocess Spawning in Loops**:
  - Flag tests shelling out to `exec.Command("git", ...)` or `exec.Command("go", ...)` inside tight test loops. Replace with fast in-memory adapters or static fixtures.
- **Pure Domain Microsecond SLA**:
  - All unit tests in `internal/domain` must execute in **microseconds (< 0.001s)** with zero mocks.
- **Frontend Test Environment Overhead**:
  - Check for heavy `jsdom` setup on pure utility tests and unmocked component tree mounts.

---

### 3. Deliver Findings

Output findings as a structured Markdown review report:

```markdown
# AgenticLens Architecture, Code & Test Review Report: <Scope>

## 1. Executive Summary
Concise assessment of the audited scope, highlighting the top 1-2 architectural risks, highest-yield code simplifications, and test execution bottlenecks.

## 2. Test Suite Performance Diagnostics
- **Total Suite Execution Time**: <X.XXs>
- **Target Execution Time**: <Y.YYs>

### Slowest Tests & Bottlenecks
| Duration | Package | Test Name | Diagnosed Bottleneck | Recommended Optimization |
|---|---|---|---|---|
| 1.45s | `internal/adapters` | `TestScanWorkspaces` | Disk fixture generation | In-memory fixtures + `t.Parallel()` |
| 0.82s | `internal/adapters` | `TestExtractMetrics` | Subprocess git invocation | In-memory adapter stub |

### Test Performance Findings & Concrete Patches
1. **<Bottleneck Category 1>**:
   - Problem: ...
   - Prescription / Code Patch:
   ```go
   // Before vs After
   ```

## 3. Architecture & Code Smells by Severity

### [High | Medium | Low]: <Finding Title>
- **Tier**: Tier 1 (Layering) | Tier 2 (Module Depth) | Tier 3 (Domain Modeling)
- **Affected Files**: `<path/to/file.go>:<line>`
- **Observed Smell**: Concrete description of the smell, boundary breach, or shallow abstraction.
- **Subtraction Target**: Exactly what dead code, wrapper, or flag can be eliminated first.
- **Proposed Target Shape**: Proposed simplified structure (describe types and seams; zero speculative code blocks).
- **Benefits**: Impact on maintainability, testability, reader load (30-second rule), and runtime performance.
- **Blast Radius**: What downstream callers or external packages touch this; safety proof needed.

## 4. Top Remediation Recommendations
Prioritized list of the highest-leverage refactorings and performance improvements to execute first.
```

---

## 4. Interactive Next Steps

Prompt the user using the `question` tool to select how to proceed:
1. **Execute Optimization/Refactor**: Hand off chosen findings directly to `agentic-lens-dev` for implementation in `workspaces/<TICKET_ID>/`.
2. **Draft Jira Ticket**: Decompose finding into a vertical tracer-bullet ticket via `agentic-lens-po`.
3. **Deepen Test Profiling**: Run targeted CPU/memory profiling (`go test -cpuprofile`) on a specific package.
