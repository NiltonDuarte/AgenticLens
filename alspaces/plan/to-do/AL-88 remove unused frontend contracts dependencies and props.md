h3. Title: AL-88: Remove unused frontend contracts, dependencies and props

*TARGET*: AgenticLens (Frontend Contracts / Components / Dependencies / Tests)

*Status*: Ready
*Work type*: refactor
*Priority*: P3
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: FE-D1
*Related*: None

*CONTEXT*

Unused api/types contracts, lucide-react and ignored component props enlarge the production surface without consumers.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to remove unused frontend contracts, dependencies and props, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected Frontend Contracts / Components / Dependencies / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Revalidate reachability and remove unused AsyncState/ApiConfig/ApiError/HealthBadgeVariant, dependency and no-op props where the intended contract permits.
* If message identity is required for an export, implement explicit provenance in its owning feature rather than keeping a no-op prop; do not invent consumers to justify speculative abstractions.
* Preserve callers and meaningful assertions; validate lockfile consistency, typecheck and canonical embedded asset build. No test-count target or unmeasured performance claim.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)
