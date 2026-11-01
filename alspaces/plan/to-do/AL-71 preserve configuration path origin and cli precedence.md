h3. Title: AL-71: Preserve configuration path origin and CLI precedence

*TARGET*: AgenticLens (Configuration / CLI / Workspace Repository / OpenCode / Tests)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: CC-05, ADP-03, ADP-04
*Related*: None

*CONTEXT*

Nested-directory scans resolve workspace paths against the wrong directory; init writes a local default that overrides its explicit database path; both OpenCode schema fallback queries hardcode workspaces.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to preserve configuration path origin and CLI precedence, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected Configuration / CLI / Workspace Repository / OpenCode / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Retain config origin and resolve configured relative paths against it; document and preserve explicit CLI override semantics and absolute paths across scan, migration and server callers.
* After init --db-path and custom --output, the next normal invocation resolves the requested database and the matching local configuration location. Assert effective behavior, not just generated shared YAML.
* Pass configured workspace policy into v1/v2 fallback resolution; verify custom roots, separators, SQL LIKE wildcard escaping, title fallback and exact-ID/nonce precedence without repository-specific constants.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 ADAPTERS_REPORT.md](../../REV-5/iterations/03/ADAPTERS_REPORT.md)
* [REV-5 CORE_CLI_REPORT.md](../../REV-5/iterations/03/CORE_CLI_REPORT.md)
