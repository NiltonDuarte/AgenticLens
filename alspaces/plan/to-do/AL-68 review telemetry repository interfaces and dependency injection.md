h3. Title: AL-68: Review shared telemetry repository interfaces and source injection

*TARGET*: AgenticLens (Architecture / Domain Ports / OpenCode v1 and v2 / Codex / CLI Composition / Backfill / Hydration / HTTP and Live Services / Tests)

*Phase*: Phase 1 (post-v2.0 platform review)
*Status*: Ready
*Work type*: docs
*Blocked by*: None
*Blocks*: None
*Origin*: PO-18 iteration 07; user requests a broader architecture review: OpenCodeV1, OpenCodeV2 and Codex should implement the same repository interface as much as possible, and the correct repository should be injected into consumers.

*CONTEXT*

Source capabilities and duration fixes alone do not establish a coherent repository boundary. Review how callers access session facts, traces, children and recovery across OpenCode schema families and Codex, which contracts they share, and where concrete source selection occurs. A file named repository.go does not itself prove a telemetry-source interface exists: distinguish persisted workspace-ledger access from harness/source repositories and distinguish schema-version dispatch from harness selection.

*GOAL*

As an engineer reviewing AgenticLens architecture, I want an evidence-backed account of the implemented repository contracts and dependency wiring, plus a justified target design, so that all three source adapters share meaningful operations and consumers receive the correct implementation without embedding storage/schema details. Deliver a reviewable architecture assessment; this ticket does not silently authorize a broad refactor.

*VERTICAL SLICE SCOPE*

* Domain Core / Ports: Inventory existing interfaces and ownership, data/result contracts, pure aggregation boundaries, and dependency direction. Assess the smallest useful common telemetry-source contract, with explicit optional capabilities where source evidence differs.
* Boundary Adapter: Trace OpenCode v1, OpenCode v2 and Codex implementations, source/schema detection and repository construction. Identify direct filesystem/SQLite access, concrete adapter coupling, duplicated dispatch and inconsistent failure semantics outside their owning boundaries.
* Composition / Injection: Inspect CLI generate, backfill, server startup, historical hydration and live/child lookup call paths. Document configuration/explicit override precedence, mixed-source ledger routing, lifetime/cleanup and cancellation/concurrency responsibilities. Follow the actual wiring rather than assuming a named factory or dependency injection container exists.
* Presentation / Consumers: Trace repository results to CLI/API/dashboard contracts; assess whether consumer logic is source-neutral where possible and whether unsupported/unavailable/partial facts remain explicit. No UI redesign is included.
* Quality: Inspect existing interface-conformance, schema-family adapter and composition tests; run focused existing checks where useful and document gaps. Produce an as-is call/dependency map, operation/implementation matrix, findings and a target design with staged recommendations. Review only: no application, schema, runtime-script or new test changes.

*ACCEPTANCE CRITERIA*

* Identify the actual current repository abstractions (or their absence), exact interface/implementation paths and constructors/call sites. Separate ledger repositories, source-session/trace access and higher-level services; do not prescribe one oversized interface that conflates all responsibilities.
* Compare common operations across OpenCode v1/v2 and Codex: source/session discovery and exact-ID lookup, fact extraction, child traversal, historical trace hydration, refresh/recovery, and live observation where actually supported. Record which belong to the common contract and which justify smaller capability-specific contracts or explicit unsupported results.
* Evaluate whether all three implementations satisfy consistent ownership, identity, snapshot, ordering, completeness and error semantics through the common boundary. Missing/unsupported source evidence must not become successful empty data, measured zero or manufactured parity. Preserve repeated occurrences and source-proven parent/child relationships.
* Trace a concrete end-to-end request for each source/schema family from effective configuration/command selection to the injected instance and consumer output, with code references and supported test/runtime evidence. Verify OpenCode schema detection and explicit Codex selection independently; a v1/v2 dispatcher behind one OpenCode implementation is acceptable if it enforces the same contract and dispatches correctly.
* Check correct source resolution for generate, backfill, server/hydration and live paths independently, including local/project settings and explicit flags as supported. For mixed-source ledgers, stored source/session identity must determine the appropriate repository per fact owner; do not route all sessions through whichever source happens to be available first. Unknown/ambiguous schema/source selection and unavailable paths must fail or report partial results explicitly rather than silently select another harness.
* Assess whether concrete construction/selection is concentrated in suitable imperative composition boundaries and dependencies are passed into consumers. Identify service-location/global-state patterns, repeated source switches, source-specific imports, hidden defaults and direct storage access that bypass the intended interface, with severity and observable impact. Do not assume every switch is a defect; justify legitimate adapter/schema dispatch.
* Review resource ownership and operational semantics: repository lifecycle/read-only access, cancellation/timeouts, concurrent callers/subscribers, stale/cache invalidation and cleanup. Keep I/O out of the pure domain and avoid forcing live-polling behavior onto repositories that cannot support it.
* Assess testing evidence for correct injection, wrong-source/wrong-schema rejection, missing/partial records, unavailable children, root/child and mixed-harness preservation, and baseline persisted-data visibility when hydration fails. State tested commit/diff, commands/results and unexecuted checks; synthetic fixture results do not establish real-source availability.
* Deliver an as-is architecture map and source/operation contract matrix, evidence-backed findings, and a proposed target interface/composition arrangement that maximizes shared behavior without erasing actual capability differences. For each recommendation record affected paths/consumers, compatibility/behavior preservation, verification needed and a bounded remediation sequence. Explicitly mark recommendations versus implemented architecture.
* Keep boundaries with existing tickets clear: AL-63 audits observable source capabilities, AL-62 fixes skill/delegation facts and ownership, AL-64 fixes duration accounting, and AL-66 owns generation toggles. Reuse their evidence when available, without adding blocking dependencies or duplicating their implementations. This ticket owns architectural review and recommendations; any application refactor requires separately approved development scope.

*DEFINITION OF DONE*

* Publish the architecture review and findings with code references, as-is and proposed dependency diagrams, explicit source/version limits and prioritized remediation recommendations. No claim that an interface/injection refactor has shipped without implementation evidence.
* Record existing tests executed/reused, coverage gaps and environment limitations; no mocks in domain or fabricated source facts. Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md, FCIS, strict Go/TypeScript, zero any/interface{} and zero emojis.
* Preserve original AL-55 evidence and unrelated tickets. Store findings/review/summary/PR and source-grounded telemetry in the active ticket workspace when review execution starts; the PO creates only this planned specification, never a premature developer workspace or CONTRACT.md.
* Publish a reviewable documentation PR; any accepted refactor proceeds separately through agentic-lens-dev with an explicit contract and approval boundary.
