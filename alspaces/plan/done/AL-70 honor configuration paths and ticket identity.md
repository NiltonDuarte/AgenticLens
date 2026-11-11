# AL-70: Honor configuration paths and ticket identity

Completed through the approved [iteration 03](../../AL-70/iterations/03/PLAN.md), including the CI fixture-lifecycle correction; verification is recorded in the delivery summary.

Iteration 02 completed under the confirmed [AL-70 contract](../../AL-70/CONTRACT.md) and [inventory verdicts](../../AL-70/LEGACY_AND_FALLBACKS.md), with verification in the [delivery summary](../../AL-70/SUMMARY.md). The original specification below is historical; confirmed scope supersedes its former defaults, source roots, selector heuristics and compatibility proposals.

## Goal

Operators get the same configured paths, source and ticket identity in every command and view, with a local dashboard bound to loopback by default.

## Context

Configuration must round-trip faithfully, retain path origin and explicit overrides, and give the same ticket/path identity to every consumer. The local dashboard must use a loopback default.

## Acceptance criteria

### Preserve configuration values through YAML generation

Configuration generation changes Project #2 to Project and omits a configured CodeMetrics.BaseBranch; hand-interpolated YAML loses valid scalar values.

* Round-trip every supported setting without scalar reinterpretation, preserving useful generated comments and nondefault BaseBranch.
* Test comment markers, colons, quotes, boolean/null-like strings, workspace paths and source roots through generation and the production loader, including the init path.

### Preserve configuration path origin and CLI precedence

Nested-directory scans resolve workspace paths against the wrong directory; init writes a local default that overrides its explicit database path; both OpenCode schema fallback queries hardcode workspaces.

* Retain config origin and resolve configured relative paths against it; document and preserve explicit CLI override semantics and absolute paths across scan, migration and server callers.
* After init --db-path and custom --output, the next normal invocation resolves the requested database and the matching local configuration location. Assert effective behavior, not just generated shared YAML.
* Pass configured workspace policy into v1/v2 fallback resolution; verify custom roots, separators, SQL LIKE wildcard escaping, title fallback and exact-ID/nonce precedence without repository-specific constants.

### Use the same effective source settings across commands (AL-67 F01, F02, F11)

Generate honors a configured local Codex home, but production server hydration currently derives it only from environment/home. Generate also anchors configured relative source paths to the project; server/backfill currently use the caller directory. The source is available, yet a different command can search the wrong place.

Example: /repo/.agenticlens-config.local.yaml selects ./sources/codex. Generate from /repo finds it. Starting the server or backfill from /repo/backend must still use /repo/sources/codex, rather than /repo/backend/sources/codex or ~/.codex. Absolute CLI overrides must win, and explicit relative CLI paths retain caller-relative meaning.

* Resolve and inject effective OpenCode/Codex settings consistently at generate, backfill and server/hydration composition boundaries. Preserve explicit/local/environment/builtin precedence where supported, shared legacy database.path and physical symlink/.. semantics.
* Exercise real loaded local configuration with conflicting configured/environment stores, nested cwd, both harnesses and an explicitly targeted workspace belonging to another project. Do not silently mix cwd-project source settings with target-project workspace settings.
* Fix the no-fallback test's evidence gap: construct a usable competing harness containing a matching owner and prove the selected source is not replaced by it. Retain selected-error/path and no-output assertions. No fallback bug is claimed from the old fixture alone.
* Verify configured Codex hydration through production server wiring; calling NewLiveServiceWithCodexHome directly does not prove configuration reaches that constructor. Retain mixed-owner routing, unavailable history and ledger-backed visibility.
* No new source-selection framework or live Codex/v2 capability is required. FIX-18 owns live capability reporting; REV-7 owns wider read-contract investigation.

### Align ticket identity parsing and stable sorting across layers

AL2 IDs are parsed as AL and sorted lexically in Go and TypeScript. Independently, descending equal-value comparisons violate strict ordering and repeatedly reorder tied rows.

* Agree a configured-prefix grammar across normalization, prefix extraction, filtering and natural ordering. Preserve documented legacy delimiterless formats without hardcoding AL or AL2.
* Test AL versus AL2, AL2-2 before AL2-10, custom digit/delimiter prefixes, case and leading zeroes through both language boundaries and consumer views.
* Ensure all affected descending comparisons treat equality correctly and follow a documented stable tie policy; repeated sorting must be idempotent for updated time, duration, invocation and identity ties. Preserve the valid cost/completion comparators.

### Preserve Git path identity and classify test support correctly

Numstat parsing collapses filename whitespace and mishandles quoting/rename forms before classification. Project patterns classify web/src/test fixture/setup files as production.

* Parse Git numstat with a lossless protocol such as NUL-delimited output; define rename classification ownership and retain real path bytes.
* Use a temporary Git repository to verify repeated spaces, tabs, non-ASCII and quoted paths, source-to-test renames, binary rows and ordinary changes against actual classification.
* Correct this project's explicit test-support patterns, proving mockData.ts/setup.ts and existing test suffixes classify correctly while preserving adopting repositories' configured policies.

### Bind the local dashboard to loopback by default

The dashboard advertises localhost but binds all interfaces by default. The user explicitly defines AgenticLens as local-only and deliberate remote exposure as operator-owned.

* Bind loopback by default and keep the advertised URL consistent with the listener.
* Prove standalone embedded UI and API access still work locally and assert the actual listener address.
* Do not add authentication, a new remote-deployment mode or mandatory remote-host configuration. Preserve ADP-05 at P2 default hardening under the user's accepted scope.

## Implementation notes

Related work: AL-68 owns hook installation. Coordinate configuration changes while preserving its explicit session binding and local-only boundaries.

Finding references: AL-67 F01, F02, F11; ADP-02, ADP-03, ADP-04, ADP-05, ADP-11, CC-05, CC-06, CC-07, D7, FE-01

* [REV-5 ADAPTERS_REPORT.md](../../REV-5/iterations/03/ADAPTERS_REPORT.md)
* [REV-5 CORE_CLI_REPORT.md](../../REV-5/iterations/03/CORE_CLI_REPORT.md)
* [REV-5 DELIVERY_REPORT.md](../../REV-5/iterations/03/DELIVERY_REPORT.md)
* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)

* [AL-67 source configuration evidence](../../AL-67/ARCHITECTURE_ASSESSMENT.md), F01/F02/F11.

*DEVELOPMENT STATE REVIEW (2026-11-05)*

Retain configuration/path/identity work. AL-67 F01/F02/F11 supplies recent composition evidence; http/server.go still binds :port. AL-59 source selection exists, so correct propagation rather than recreating selection.

Evidence and review limits: [PO-19 reconciliation](../../PO-19/SPEC.md), merged baseline `03f5b05d`. This planning review does not certify runtime behavior or close unverified acceptance criteria.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [relevant regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. Historical findings require revalidation; this rewrite is not runtime verification or implementation approval.
