# REV-5: full repository code audit

**Verdict: CHANGES_REQUESTED.** Every tracked program-code file is covered. The repository is not free of correctness defects, code smells or architectural weaknesses. This iteration records findings only; none of these newly identified issues is represented as fixed.

Baseline: `d5227812b09a17ff15d7e6ffe058e96ecf1616eb`. Initial working tree was clean. This expands the earlier change-focused review and supersedes any inference that its approval established whole-repository cleanliness.

## Coverage and method

[COVERAGE.json](COVERAGE.json) contains exactly one disposition per scoped file, with content hash, method, assessment and finding IDs. [SCOPE.json](SCOPE.json) freezes the inventory. Reconciliation found no missing or duplicate paths and no changed audited file hashes.

- 48 core/CLI files: all 46 Go production/test files read completely, two dependency manifests structurally reviewed.
- 82 adapter files: every production/test file and the source fixture read, with relevant caller/consumer tracing.
- 101 frontend files: all source, tests, CSS, fixture and configuration content reviewed; the lockfile inspected structurally across all package entries, not third-party implementation code.
- 77 delivery/support files: five shell scripts, Python compatibility provider, build/container files, workflows, active configuration and supporting fixture data. Runnable code under the setup skill and historical AL-58 evidence directory is included. Historical fixture Markdown is treated as inert fixture data, not endorsed instructions.

Total: 308 scoped code/configuration/support files. A separate global tracked-source-extension and executable-file check found no program-code path omitted from the manifest. This total is not a claim that 308 files are production code. Ordinary historical ticket documents/telemetry, prose documentation, ignored dependencies/build output and Git internals are excluded. Sources referenced as documentation context are not falsely counted as code inspection.

Full semantic reading does not execute every possible path or prove absence of defects. Source-level conclusions, reproduced behavior, test mutation evidence, and structural manifest checks are identified separately. Prior passing tests are reused as baseline evidence; they do not invalidate newly reproduced defects.

## Highest-priority risks

| Priority | Finding | Evidence and owner report |
| --- | --- | --- |
| P1 | Warning-bearing Codex recovery can replace known stored facts with incomplete zero-valued facts while claiming complete recovery. | Disposable probe: stored input1000 became0, parser warned, recovery returned complete/updated without diagnostics. [ADP-01](ADAPTERS_REPORT.md). |
| P2 | Default listener binds network interfaces despite the local-only product scope. | User clarified intentional remote exposure is operator-owned; authentication is not required for supported local use. Residual concern is implicit LAN binding; recommend a loopback default. [ADP-05](ADAPTERS_REPORT.md). |
| P1 | Release force_tag can attach binaries built from the dispatch revision to an older release tag. | Workflow control-flow evidence; no remote publishing performed. [D3](DELIVERY_REPORT.md). |
| P2 | Partial generation preserves remote sessions but deletes their historical auxiliary skill/task facts. | Isolated CLI/SQLite reproduction. [CC-02](CORE_CLI_REPORT.md). |
| P2 | Configuration values lose meaning across writer, loader, command and aggregation boundaries. | YAML roundtrip/name/BaseBranch loss; nested-root scan failure; init DB override; configured anomaly thresholds omitted or unused. [ADP-02/03/04](ADAPTERS_REPORT.md), [CC-03/04/05](CORE_CLI_REPORT.md). |
| P2 | Successful inspection output is misclassified as tool failure. | Actual classifier returns errors for successful reads containing exit/crash phrases; live/offline callers feed those strings. [CC-01](CORE_CLI_REPORT.md). |
| P2 | Configured AL2 tickets are misparsed, incorrectly filtered and lexically sorted. | Actual Go and TypeScript functions reproduced the same defect. CC-06 and FE-01 are one cross-layer problem, not two independent counts. [Core](CORE_CLI_REPORT.md), [frontend](FRONTEND_REPORT.md). |
| P2 | Frontend request ownership and validation gaps permit stale selection data and malformed live payloads. | Unfenced view requests; malformed nested live payload accepted then consumer throws; fallback status remains error after successful recovery. [FE-02/03/04/05](FRONTEND_REPORT.md). |
| P2 | Summary updates can erase subsequent peer-level verification evidence. | Disposable file probe deleted a following third-level Verification section. [ADP-10](ADAPTERS_REPORT.md). |
| P2 | Several tests pass without checking their named contract. | Disabling all required-column checks still passed the omission test; other no-write/readiness/sorting/UI assertions observe the wrong state or boundary. [ADP-07/08/09](ADAPTERS_REPORT.md), [CC-11/12](CORE_CLI_REPORT.md), [FE-15/16/D3](FRONTEND_REPORT.md). |
| P2 | Delivery scripts can report success after failed updates or partial checksum generation. | Local failure injection produced update exit0 after all downloads failed; package and verify both exited0 with a partial checksum manifest. [D1/2/5](DELIVERY_REPORT.md). |

These are priorities, not the complete issue list. The four reports below retain every finding, including lower-severity subtraction opportunities and the exact test/verification implications.

## Complete finding catalog

- [CORE_CLI_REPORT.md](CORE_CLI_REPORT.md): CC-01 through CC-12. Covers status/content confusion, auxiliary-fact preservation, threshold propagation, ignored error-rate configuration, DB-path precedence, ticket parsing, descending equality ordering, dropped scan diagnostics, migration input aliasing, dead computation/duplicate aggregation, discarded test setup and non-isolated server smoke tests.
- [ADAPTERS_REPORT.md](ADAPTERS_REPORT.md): ADP-01 through ADP-13. Covers incomplete recovery, YAML serialization, configured path origin/fallbacks, HTTP exposure, silent child-linkage failure, test blind spots/environment coupling, destructive summary replacement, lossy numstat path parsing, duplicated source-decoding fidelity, and the test-only exported SQLite DB accessor.
- [FRONTEND_REPORT.md](FRONTEND_REPORT.md): FE-01 through FE-16, plus FE-D1 through FE-D3. Covers parsing and async/validation issues above; keyboard interception, disclosure controls, false clipboard success, contradictory anomaly identity, overlapping chart bars, incorrectly weighted tool-success KPI, failure-count semantics, keyboard/focus access, tooltip and reasoning visibility. Also identifies unused type contracts/dependency/props, duplicated state ownership, and specific test-oracle weaknesses. FE-D2/D3 are cross-cutting architecture/test analyses of related defects, not extra independent runtime failures.
- [DELIVERY_REPORT.md](DELIVERY_REPORT.md): D1 through D7. Covers update/integrity/release workflow issues, a compatibility harness tied to historical workspace artifacts, checksum failure masking, duplicated installers, and test fixture/setup files classified as production by configured telemetry patterns. Additional toolchain/lockfile reproducibility observations are explicitly bounded.

## Architecture assessment

The architecture has useful foundations: Go domain functions generally avoid filesystem/network/clocks, adapters own I/O, session-owned facts preserve identity, explicit cost evidence distinguishes unknown from zero, and the frontend remains embedded in a standalone binary. A large file or one implementation of a consumer interface was not treated as a defect by itself.

The weak point is consistency across boundaries, not merely naming or formatting:

1. **Facts and completeness have multiple owners.** Generate and backfill preserve different auxiliary facts; a tolerant parser's warnings are discarded before snapshot replacement; optional linkage failure looks like successful absence. Give each refresh a shared completeness/ownership contract and preserve known facts on incomplete recovery.
2. **Effective configuration is not carried through the full operation.** Names/paths/base branches/thresholds survive some boundaries but are overridden, dropped, reinterpreted relative to a different directory, or replaced by defaults at later projections. Resolve configuration and its origin once, then pass the needed typed policy explicitly.
3. **Source interpretation is duplicated.** Live/offline decoders disagree about raw text; output content is interpreted as execution status; Go/TypeScript ID parsers use a grammar inconsistent with configured prefixes. Share the specific semantic contract and parity cases, not a speculative universal adapter.
4. **UI request and disclosure state have competing owners.** Controls, selected rows, pending network responses, transport status and CSS can disagree about what is current or visible. Use explicit ownership for requests and disclosures; preserve existing good ownership logic in useLiveTrace and TicketDeepDive.
5. **Subtraction opportunities remain beyond function-name reachability.** The core allocates values that are never consumed, thread model aggregation duplicates existing logic, a SQLite handle accessor exists only for a test, frontend contracts/props/dependency have no consumers, and installer copies require parallel fixes. Earlier zero-name-reference screening could not detect writes-only data flow, unused types/props, or same-named methods with different receiver types.
6. **Some tests validate implementation scaffolding instead of the behavior named by the test.** Strong migration/cost/hydration tests should remain. Repair ineffective oracles before using green suites as evidence; consolidate only with explicit assertion recipients, not count targets.

Prefer targeted corrections and deletion before a broad rewrite. Fix fact preservation and release identity first, then configuration/identity propagation and async boundaries, then repair oracles and remove unused surfaces. The reports specify plausible regressions and recipients; implementation requires a separate approved remediation iteration.

## Reproductions and validation

Core review built the unchanged CLI and used isolated domain/CLI/SQLite fixtures. Adapter review used disposable copies for four expected-failing defect probes and one targeted mutation. Frontend review executed actual production functions/transports/server-side rendering in memory on synthetic input. Delivery probes used unchanged scripts/Makefile with local harmless binaries and HTTP/hash-tool stubs; no external release or network download was performed. See [DELIVERY_REPRODUCTIONS.json](DELIVERY_REPRODUCTIONS.json), [PACKAGING_REPRODUCTION.json](PACKAGING_REPRODUCTION.json) and [CLASSIFICATION_REPRODUCTION.json](CLASSIFICATION_REPRODUCTION.json).

The expected-failing probes are findings, not failed repository modifications. All audited code hashes remain unchanged. Shell syntax and Python AST checks passed; JSON/YAML/TOML structural parsing passed. No browser geometry, real release publication, OpenCode compatibility session, dependency vulnerability scan, or runtime performance profile was performed. Full runtime/provider compatibility is not inferred from source inspection.

Root delivery checks and final telemetry are recorded in SUMMARY.md. New findings remain open regardless of the unchanged suite result.

## Audit final verification

At audit commit 8721e36, origin/main (47af923) was already merged. Required `make` exited 0: Go vet/race suites, TypeScript typecheck/ESLint, 781 frontend tests across 35 files, canonical embedded asset build and four Linux/Darwin binaries passed. All 308 audited file hashes remained unchanged afterward. Existing passing tests do not close the documented defects or ineffective-oracle findings. No application code changed in iteration 03. Earlier standalone asset-serving evidence remains applicable to the unchanged frontend sources.

Build warnings remain: npm install reported 21 dependency advisories (18 moderate, one high, two critical), and Vite reported a bundle over 500 kB. These are tool-reported signals, not independently triaged vulnerability or performance findings; production reachability was not assessed. No automated dependency upgrades were applied.

## ADP-05 severity correction

User clarified that the product is local-only and deliberate remote exposure is the operator's responsibility. ADP-05 is reduced from P1 to P2 and narrowed to the default bind address, which is broader than the advertised localhost use. Lack of authentication is not a defect within the supported scope. Loopback binding is the minimal recommendation; authentication and remote-deployment features are not required. This supersedes the original priority, preserves the source evidence and all other findings, and makes no runtime change. Existing source hashes and verification remain valid.
