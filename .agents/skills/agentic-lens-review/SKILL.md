---
name: agentic-lens-review
description: Use when auditing, reviewing, or finding refactoring, performance, and architectural improvement opportunities in the AgenticLens Go and TypeScript codebase, with independent correctness investigation and risk-directed verification.
---

# AgenticLens Architecture, Code & Test Suite Review

Find concrete correctness defects in AgenticLens Go, TypeScript and delivery code. Plans, implementation reports and passing suites are evidence to evaluate, not conclusions to endorse. Use this procedure for development reviews as well as standalone audits; architecture and test profiling are additional lenses, not substitutes for investigating behavior.

## Review coverage

When coordinating substantive application changes or broad audits, use [specialized parallel review](references/parallel-review.md): four core focuses, risk-selected specialists, independent first passes and evidence-based consolidation. The reference owns dispatch, capacity-limited waves and revision-aware delivery gates. Narrow low-risk documentation/mechanical edits may use fewer focuses with a recorded rationale. A focused reviewer reads only its own persona, shared reviewer rules and the investigation procedure below; the coordinator loads the role-selection reference. Do not load sibling personas or reports on the first pass.

## Independent correctness review

1. Establish the requested boundary, base/head revision (or diff hash), supported deployment assumptions, and requirements. For a fix review distinguish verifying the supplied fix from discovering additional defects. Whole-repository claims require a reconciled file inventory; a diff review makes no such claim.
2. Read source and the diff, then identify plausible failure hypotheses before relying on the author's explanation or prior approval. Prefer a fresh reviewer context containing requirements and raw evidence over the implementation conversation. Do not withhold requirements or useful test evidence; reconcile reports after forming the initial assessment. Independence of context is not proof of independence of reasoning.
3. Follow the consequential execution paths through callers, producers, transformations, persistence and consumers as applicable. Inspect unchanged dependencies when needed. Implementation scope limits do not prohibit relevant read-only investigation. Challenge whether the plan itself preserves required behavior; report conflicts rather than silently changing scope.
4. Investigate the highest-impact plausible failures. For data replacement/migrations, release identity and concurrency, seek an isolated probe, existing behavioral test or precise source-level counterexample. Exercise actual boundaries where feasible; explain when evidence is only static. Do not require a generic test matrix, minimum finding count or repetitive full-suite execution. For low-risk changes a clear source trace may suffice.
5. Verify candidate findings before reporting: identify the trigger, reachable call path, incorrect observable result and supported expectation; check callers/guards for evidence that refutes the hypothesis. Distinguish introduced regressions, pre-existing defects and optional maintenance suggestions. Severity depends on actual supported use and impact, not an imagined deployment. A credible finding from one reviewer does not require majority agreement.
6. Assess the test oracle: would the relevant test fail for the plausible regression, and does its fixture reach the claimed branch? For consequential fixes, seek before/after failure evidence or a targeted disposable mutation when practical. Reuse valid results for unchanged code, but do not treat a green suite as an independent correctness argument. Do not mutate shared production files to demonstrate a bug.
7. Reconcile the investigation with requirements, implementation/test reports and engineering standards. Report findings first and bound the conclusion to the inspected revision and paths. Stop when the scoped risks have sufficient evidence; name material gaps instead of repeating reviews until someone approves.

### Evidence and outcomes

Use a compact report: reviewed revision/scope; findings with locations, triggers, consequences and evidence; consequential hypotheses investigated and their disposition; reused/executed verification; remaining limits. This is a record of investigation, not a mandatory essay for trivial edits. Separate unresolved introduced defects from existing debt and suggestions; do not silently omit an existing defect because it is outside the diff.

- `CHANGES_REQUESTED`: substantiated blocking correctness or required-contract defects. Explain what must change and why.
- `INCOMPLETE`: material access, context or verification gaps prevent a defensible conclusion. State the missing evidence and next action. This does not pass the delivery gate.
- `APPROVED`: no blocking finding in the scoped investigation, with material hypotheses assessed and no unresolved evidence gap. This is not proof of bug absence, whole-repository quality or authorization to merge.

### Reviewer effectiveness

When substantially changing review behavior, use bounded independent forward tests on defect and corrected/control cases without supplying expected findings. Keep a holdout outside the examples used to write the instructions. Record prompt versions, source revisions, model/context settings, observed misses, false positives and limits; record cost/time only when measured. Compare equivalent runs before claiming improvement. Small fixtures establish usability, not production recall. Turn recurring deterministic failures into targeted tests or checks; keep this skill as a short route to canonical rules rather than accumulating incident-specific prohibitions.

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

Select relevant lenses after the correctness investigation; use all four for a broad architecture/test audit:

#### Tier 1: Layering & Boundary Discipline (Go + TypeScript FCIS)
- **FCIS Isolation Breach**: Does `internal/domain` import any outer layer (`internal/adapters`, `cmd/`) or I/O library (`os`, `net/http`, `database/sql`, `path/filepath`)? Reject immediately.
- **Impure Domain Functions**: Do functions in `internal/domain` call `time.Now()`, generate random numbers, or access global mutable variables instead of receiving values as immutable arguments?
- **Boundary Model Leakage**: Are raw JSON dictionaries (`map[string]any`) leaking into domain functions? External JSON must be unmarshaled and validated in `internal/adapters` into immutable value structs before passing to domain logic.
- **State Ownership**: Trace mutable adapter state, caches and subscriptions to their owners, invalidation and synchronization. Necessary shell state is not itself a defect; identify a concrete leak, stale read or race.

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
