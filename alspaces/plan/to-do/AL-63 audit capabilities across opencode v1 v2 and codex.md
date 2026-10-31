h3. Title: AL-63: Audit source capability parity across OpenCode v1, OpenCode v2, and Codex

*TARGET*: AgenticLens (Multi-source ingestion / hydration / recovery)

*Phase*: Phase 1 (post-v2.0 platform review)
*Status*: Blocked
*Work type*: docs
*Blocked by*: AL-62
*Blocks*: AL-65
*Origin*: PO-18; user review dated 2026-10-31, build v2.0.0-rc.106+9707d41-1-g45cb679.

*CONTEXT*

The user reports Codex lacks trace hydration and requests a capability comparison. Identical capabilities across harnesses are not assumed. Incorporate the merged Codex-fact fix and granular ownership delivery instead of rerunning unchanged evidence. The standalone active OpenCode walkthrough has been removed from this plan; this capability audit retains its original matrix scope and reports unexecuted live checks explicitly.

*GOAL*

As an operator, I want a versioned capability matrix with evidence and limits so that I can choose sources and diagnose gaps accurately.

*VERTICAL SLICE SCOPE*

* Domain Core: Audit fact availability, persisted baseline projections, execution-error semantics, tree completeness, and mixed-age state.
* Boundary Adapter: Inspect OpenCode schema v1/v2 and Codex paths independently for ingestion, hydration, live activity, and exact-identity recovery.
* Presentation: Verify CLI/API/dashboard exposure, skill and delegation facts, child navigation and error reporting for each source.
* Quality: Bounded verification using existing suites and controlled disposable ledgers. Produce one matrix and reproducible evidence inventory; no parity implementation, schema change, or new test code.

*ACCEPTANCE CRITERIA*

* Provide a three-source matrix covering ingestion, historical trace hydration, live observability, skill facts, delegation facts, child-session navigation, execution-error reporting, and source-backed recovery.
* Every cell distinguishes supported and verified, supported but failing, unsupported, unavailable source evidence, or not yet executed. Cite declared support separately from runtime evidence, including build, source version/schema, command, IDs and result.
* Check persisted-data behavior with optional hydration available and unavailable; supported stored facts remain visible when enrichment fails.
* For both OpenCode schema families, check missing and malformed source records, unavailable children, retention of previous trees and skill/task snapshots, and explicit incomplete/mixed-age diagnostics. Do not treat unavailable lookup as successful empty data.
* Check exact stored-session identity, parent/source preservation, other-harness facts, growing-source refresh, unchanged repeat runs, and explicit opt-in for heuristic recovery using disposable evidence copies.
* Use AL-62 diagnosis/implemented-fix verification; record residual gaps without asserting a successful backfill or preservation check based on the original notes. Findings propose follow-up tickets rather than silently expanding this audit.

*DELIVERY CONSTRAINTS*

* Phase: Phase 1 (post-v2.0 platform review); use AL identifiers, not AL2.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md. FCIS; Go 1.23+ and TypeScript 5.x strict; zero any/interface{} and zero emojis.
* Tests and experiments use self-contained fixtures/disposable copies, never live ticket workspaces. Preserve alspaces/AL-55/agentic_metrics.json.
* No new capability is inferred from missing data. Separate source evidence, recorded observations, and hypotheses.

*DEFINITION OF DONE*

* Deliver the evidence report, limitations and follow-up recommendation in the active ticket workspace; no application fixes are part of this ticket.
* Record exact checks executed, failures, unavailable evidence and unexecuted checks. Existing test results may be reused for unchanged code.
* Maintain ticket summary/review/PR artifacts and source-grounded telemetry; publish a reviewable PR. Completion requires evidence, not just a plan.
