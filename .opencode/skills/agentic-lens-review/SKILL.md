---
name: agentic-lens-review
description: Use when auditing the AgenticLens Go and TypeScript codebase, diagnosing slow test suites, profiling test execution bottlenecks, or reviewing PR diffs for architectural compliance and performance.
---

# AgenticLens Code & Test Suite Review Skill

Audits the **AgenticLens** codebase (Go backend/CLI and TypeScript frontend). Performs deep profiling of the test suite to diagnose and fix slow test executions, audits architectural integrity against `TECHNICAL_GUIDELINES.md`, and verifies zero `any`, FCIS layer isolation, and single-binary packaging.

---

## 1. Diagnostics & Review Modes

This skill supports two specialized audit modes:

### Mode 1: Test Suite Performance & Runtime Optimization (Primary)
Use when tests take too long to run or CI execution is sluggish.
Profiles both Go and TypeScript test suites, pinpoints the slowest tests, diagnoses the root cause (I/O churn, missing parallelism, artificial sleeps, heavy fixtures), and generates concrete performance optimization patches.

### Mode 2: Full Codebase & Diff Review
Use when reviewing a PR, branch diff, or active iteration (`workspaces/<TICKET_ID>/`).
Audits changes against `TECHNICAL_GUIDELINES.md` (FCIS inward flow, Go 1.23+ idioms, TypeScript strict mode, zero `any`, `embed.FS` bundling, zero emojis) and originating ticket specifications.

---

## 2. Test Suite Profiling & Diagnostics Protocol

### Step 1: Measure & Profile Go Test Suite
Run the Go test suite with verbose JSON output to extract precise execution durations:
```bash
go test -json ./... | jq -r 'select(.Action=="pass" or .Action=="fail") | select(.Test != null) | "\(.Elapsed)s \(.Package) \(.Test)"' | sort -nr | head -20
```
Also measure total package-level durations:
```bash
go test -json ./... | jq -r 'select(.Action=="pass" or .Action=="fail") | select(.Test == null) | "\(.Elapsed)s \(.Package)"' | sort -nr
```

### Step 2: Measure TypeScript / Frontend Test Suite (if present)
```bash
test -d web && (cd web && npm test -- --reporter=verbose) || true
```

### Step 3: Investigate Root-Cause Bottlenecks

Inspect flagged slow tests against the following anti-pattern checklist:

1. **Missing Test Parallelism (`t.Parallel()`)**:
   - Are independent unit and table-driven subtests executing serially?
   - Add `t.Parallel()` to top-level test functions and inside `t.Run()` loops (capturing `tc := tc` or relying on Go 1.22+ loop scoping).
2. **Artificial Delays & Sleeps**:
   - Search for `time.Sleep(...)` across `*_test.go` and `web/**/*.test.ts`.
   - **Rule**: Banish `time.Sleep` from tests. Replace with synchronized channels (`select { case <-done: ... case <-time.After(50*time.Millisecond): t.Fatal(...) }`), mutex conditions, or deterministic polled assertions.
3. **Redundant Disk I/O & Database Fixture Churn**:
   - Are tests creating fresh SQLite databases on disk (`tempfile`) instead of using in-memory SQLite (`:memory:` or `file::memory:?cache=shared`)?
   - Are file traversal tests generating thousands of files on physical disk rather than using memory-backed filesystems or minimal representative trees?
4. **Subprocess Spawning in Loops**:
   - Are tests executing `exec.Command("git", ...)` or `exec.Command("go", ...)` inside tight test loops?
   - Refactor tests to call pure domain functions or memory adapters directly instead of shelling out to subprocesses.
5. **Slow Pure Domain Tests**:
   - Pure functions in `internal/domain` should execute in **microseconds (< 0.001s)**. If any domain test exceeds 5ms, check for accidental I/O, heavy memory allocations, or quadratic algorithms.
6. **Frontend Test Environment Overhead**:
   - Are pure utility formatters running under heavy `jsdom` rather than lightweight node/happy-dom test runners?
   - Are tests repeatedly mounting full React trees when testing pure functional calculations?

---

## 3. Architecture & Standards Audit Checklist

Verify compliance with `TECHNICAL_GUIDELINES.md`:

### Go Standards
- **FCIS Isolation**: `internal/domain` contains zero imports of `os`, `net/http`, `database/sql`, or outer adapter packages.
- **Value Semantics**: Domain models use value-semantic structs; aggregations are pure deterministic functions.
- **Zero `any` / `interface{}`**: All structs, interfaces, and function parameters use concrete types or strictly bounded generics.
- **Error Discipline**: Errors are wrapped with domain context using `%w`. Zero discarded errors (`_ = ...`).
- **Single Binary**: Static assets in `web/dist` are bundled via `embed.FS`.

### TypeScript Standards
- **Strict Mode**: `"strict": true` and `"noImplicitAny": true`. Zero `any` types.
- **Pure Presentation**: View components cleanly separated from API clients. Graceful empty states.
- **Zero Emojis**: Audit code, comments, UI strings, charts, metric cards, and terminal logs for zero emojis.

---

## 4. Review Report Structure

The review must output an actionable report in markdown format:

```markdown
# AgenticLens Code & Test Review Report

## 1. Test Suite Performance Diagnostics
- **Total Suite Execution Time**: <X.XXs>
- **Target Execution Time**: <Y.YYs>

### Top Slowest Tests
| Duration | Package | Test Name | Diagnosed Bottleneck | Recommended Optimization |
|---|---|---|---|---|
| 1.45s | `internal/adapters` | `TestScanWorkspaces` | Disk fixture generation | Switch to in-memory fixtures / t.Parallel() |
| 0.82s | `internal/adapters` | `TestExtractMetrics` | Subprocess git invocation | Mock git output / stub helper |

### Test Performance Findings & Actionable Fixes
1. **<Bottleneck Category 1>**:
   - Problem: ...
   - Prescription / Code Patch:
   ```go
   // Before vs After
   ```

## 2. Architecture & Standards Audit
| Area | Status | Finding / Violation |
|---|---|---|
| FCIS Layer Flow | PASS / WARN | ... |
| Go Type Strictness | PASS / FAIL | ... |
| Error Handling (%w) | PASS / FAIL | ... |
| TypeScript Strictness | PASS / FAIL | ... |
| Emoji Audit | PASS / FAIL | ... |

## 3. Prioritized Remediation Plan
- [ ] Task 1: ...
- [ ] Task 2: ...
```
