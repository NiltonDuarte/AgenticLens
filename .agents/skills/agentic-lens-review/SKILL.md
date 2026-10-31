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
- **Pure, Deterministic Domain**: Preserve zero I/O and zero mocks in domain tests. Measure performance against an actual requirement; rounded runner timings do not establish a microsecond SLA.
- **Test Value Before Volume**: Apply [test purpose and proportionality](../../../TECHNICAL_GUIDELINES.md#test-purpose-and-proportionality) and the [test review checklist](../../../docs/REGRESSION_PREVENTION.md#adding-reviewing-or-reducing-tests). Scores, coverage, counts, and line ratios prioritize inspection; they never authorize automatic deletion or require additional tests.

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
   - `.agenticlens-config.yaml` (Project parametrization)

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

#### Tier 4: Test Value, Performance & Flakiness
- **Behavior and Coverage**: Identify the production contract and plausible regression each candidate detects. Inspect helpers, compiler checks, and existing peers. Confirm deletion rationales or assertion-to-recipient mappings; preserve distinct ownership and layer coverage. If scores are requested, state the rubric and review confidence without treating a score band as an automatic removal rule.
- **Boundary Evidence**: Distinguish text/source assertions, rejection guards, compiler contracts, DOM behavior, and real build/release execution. Report exactly which boundary was exercised and what remains unverified.
- **Measured Bottlenecks**: Profile when performance is in scope or a concrete slowdown warrants investigation. Run from `backend/` after `make web-assets`, using comparable cache, race-detector, and concurrency settings. For example:
  ```bash
  go test -race -count=1 -json ./... > /tmp/agenticlens-test-profile.jsonl
  jq -r 'select(.Action=="pass" or .Action=="fail") | select(.Test != null) | "\(.Elapsed)s \(.Package) \(.Test)"' /tmp/agenticlens-test-profile.jsonl | sort -nr | head -20
  ```
  Check the test command exit status separately from report generation. Recorded durations guide diagnosis; they are not a universal speed budget, and summed parent/subtest timings can double-count work.
- **Safe Parallelism**: Consider `t.Parallel()` only for independent setup and state. Environment, process-global, and linker-metadata mutations require serial execution or effective isolation. Do not parallelize solely to meet a runtime target.
- **Timing and Synchronization**: Inspect arbitrary sleeps and flaky waits. Prefer deterministic synchronization or controlled timers where practical; an arbitrary short timeout is not a deterministic substitute.
- **Adapter Fidelity**: Real filesystem, SQLite, and Git fixtures may be essential integration proof. Diagnose redundant setup or subprocess repetition before proposing optimization; in-memory substitutes must preserve the claimed boundary. Keep domain tests pure.
- **Frontend Environment**: Use a lightweight environment for pure utilities where appropriate, while preserving real component interactions and existing failure oracles. Do not replace meaningful integration coverage solely to reduce duration.

---

### 3. Deliver Findings

Adapt the report to the actual scope. Include:

1. **Assessment and evidence**: audited baseline/diff, inspected paths, observed findings versus inference, and review confidence. Link precise source locations.
2. **Test decisions, when in scope**: each removal rationale or existing recipient for meaningful assertions, retained independent coverage, and any requested score with its rubric. Use syntax-aware inventory and distinguish declarations from expanded runtime cases.
3. **Performance, when measured**: commands, environment/cache/race settings, actual results, diagnosed bottlenecks, and proposed changes that preserve fidelity. Omit invented durations or target budgets when profiling was not done.
4. **Architecture findings by severity**: affected boundary, concrete smell, subtraction opportunity, proposed shape, benefit, blast radius, and evidence required for safety. Preserve the Tier 1-3 checks above.
5. **Prioritized recommendations and limits**: actual passed/failed/skipped checks, unverified claims, and proposed work within the authorized scope. A passing suite alone does not establish mutation equivalence or justify a deletion threshold.
