h3. Title: AL-70: Preserve configuration, paths and ticket identity

*TARGET*: AgenticLens (Configuration / CLI / Git and OpenCode Adapters / HTTP Listener / Frontend Identity)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit; consolidated draft batch at the user's request in iteration 09.
*Findings*: ADP-02, ADP-03, ADP-04, ADP-05, ADP-11, CC-05, CC-06, CC-07, D7, FE-01
*Related*: AL-68 owns hook installation. Coordinate configuration changes while preserving its explicit session binding and local-only boundaries.

*CONTEXT*

Configuration must round-trip faithfully, retain path origin and explicit overrides, and give the same ticket/path identity to every consumer. The local dashboard must use a loopback default.

The source reports retain affected paths, triggers, evidence and limits at their frozen baseline. Revalidate before implementation; this specification does not claim remediation. Overlapping findings and test/architecture analyses are counted once in the current finding map.

*GOAL*

Correct the scoped behavior through its production consumers and verification boundaries, satisfying every acceptance section below.

*VERTICAL SLICE SCOPE*

* Own the changes through Configuration / CLI / Git and OpenCode Adapters / HTTP Listener / Frontend Identity, including real consumer behavior and regression proof. Keep pure calculations in the core and I/O in adapters; preserve supported stored facts when enrichment is unavailable.
* Use bounded implementation and review iterations within this ticket. Grouping reduces tracking overhead; it does not authorize dropping acceptance cases, skipping focused reviews or calling a partial fix complete. No broad framework rewrite or new speculative abstraction is required.

*ACCEPTANCE CRITERIA*

h4. Preserve configuration values through YAML generation

Configuration generation changes Project #2 to Project and omits a configured CodeMetrics.BaseBranch; hand-interpolated YAML loses valid scalar values.

* Round-trip every supported setting without scalar reinterpretation, preserving useful generated comments and nondefault BaseBranch.
* Test comment markers, colons, quotes, boolean/null-like strings, workspace paths and source roots through generation and the production loader, including the init path.

h4. Preserve configuration path origin and CLI precedence

Nested-directory scans resolve workspace paths against the wrong directory; init writes a local default that overrides its explicit database path; both OpenCode schema fallback queries hardcode workspaces.

* Retain config origin and resolve configured relative paths against it; document and preserve explicit CLI override semantics and absolute paths across scan, migration and server callers.
* After init --db-path and custom --output, the next normal invocation resolves the requested database and the matching local configuration location. Assert effective behavior, not just generated shared YAML.
* Pass configured workspace policy into v1/v2 fallback resolution; verify custom roots, separators, SQL LIKE wildcard escaping, title fallback and exact-ID/nonce precedence without repository-specific constants.

h4. Align ticket identity parsing and stable sorting across layers

AL2 IDs are parsed as AL and sorted lexically in Go and TypeScript. Independently, descending equal-value comparisons violate strict ordering and repeatedly reorder tied rows.

* Agree a configured-prefix grammar across normalization, prefix extraction, filtering and natural ordering. Preserve documented legacy delimiterless formats without hardcoding AL or AL2.
* Test AL versus AL2, AL2-2 before AL2-10, custom digit/delimiter prefixes, case and leading zeroes through both language boundaries and consumer views.
* Ensure all affected descending comparisons treat equality correctly and follow a documented stable tie policy; repeated sorting must be idempotent for updated time, duration, invocation and identity ties. Preserve the valid cost/completion comparators.

h4. Preserve Git path identity and classify test support correctly

Numstat parsing collapses filename whitespace and mishandles quoting/rename forms before classification. Project patterns classify web/src/test fixture/setup files as production.

* Parse Git numstat with a lossless protocol such as NUL-delimited output; define rename classification ownership and retain real path bytes.
* Use a temporary Git repository to verify repeated spaces, tabs, non-ASCII and quoted paths, source-to-test renames, binary rows and ordinary changes against actual classification.
* Correct this project's explicit test-support patterns, proving mockData.ts/setup.ts and existing test suffixes classify correctly while preserving adopting repositories' configured policies.

h4. Bind the local dashboard to loopback by default

The dashboard advertises localhost but binds all interfaces by default. The user explicitly defines AgenticLens as local-only and deliberate remote exposure as operator-owned.

* Bind loopback by default and keep the advertised URL consistent with the listener.
* Prove standalone embedded UI and API access still work locally and assert the actual listener address.
* Do not add authentication, a new remote-deployment mode or mandatory remote-host configuration. Preserve ADP-05 at P2 default hardening under the user's accepted scope.

* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with an approved concrete contract. Complete every scoped correction and evidence case; this planning session changes no application, test or runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md: strict typing, zero any/interface{}, zero emojis, no domain mocks and self-contained fixtures. Pass affected checks, independent focused review and the required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Preserve source identity, unavailable values, meaningful existing assertions and applicable local-only boundaries. Record tested revision, reproduction/probe evidence, reused checks and environment limits; update finding dispositions only after verified fixes. Publish the implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 ADAPTERS_REPORT.md](../../REV-5/iterations/03/ADAPTERS_REPORT.md)
* [REV-5 CORE_CLI_REPORT.md](../../REV-5/iterations/03/CORE_CLI_REPORT.md)
* [REV-5 DELIVERY_REPORT.md](../../REV-5/iterations/03/DELIVERY_REPORT.md)
* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)
