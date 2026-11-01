# REV-5 iteration 03: adapter audit

Verdict: CHANGES_REQUESTED. Audit only; none of the recommendations below has been implemented.

All 82 `owner=backend` entries in SCOPE.json received a full-file semantic read, including the complete tests and the JSONL fixture. The accompanying coverage manifest records file-specific assessments. The review followed AGENTS.md, TECHNICAL_GUIDELINES.md, docs/REGRESSION_PREVENTION.md, the reviewer persona, the review skill, and the iteration plan/contract. Baseline: `d5227812b09a17ff15d7e6ffe058e96ecf1616eb`. Source and test files were not edited. No live workspace fixtures were used for probes.

Severity uses P1 for substantial loss/exposure of stored or private evidence, P2 for reproducible correctness or material validation/test gaps, and P3 for smaller maintainability issues. Findings distinguish executed evidence from static inference. Absence of a finding in a file is not a claim that the file has no possible defects.

## ADP-01 — P1: warning-bearing Codex recovery replaces known facts with incomplete facts

Locations: `backend/internal/adapters/backfill/recovery.go:138-162`, especially 148-152; replacement at 105-127. Parser behavior: `backend/internal/adapters/codex/rollout.go:226-229`.

Trigger: a previously stored Codex session remains discoverable, but its source rollout contains malformed usage data. ParseRolloutPath returns a transcript and warnings instead of a fatal error. readCodexTree ignores those warnings, and recovery checks identity/tree membership before replacing the stored tree. Those checks do not establish completeness of token or invocation facts.

Executed evidence: a disposable-copy test seeded stored Input=1000 and a rollout containing valid root/model records followed by malformed JSON. The parser reported `line 3: malformed JSON record`; recovery returned `complete=true`, `updated=true`, no diagnostics, and Input=0. The test deliberately failed its preservation assertion. Thus this is not merely a speculative difference in totals: known stored evidence was replaced with an incomplete parse presented as complete.

Safe change: propagate parse completeness and diagnostics through the recovery boundary; retain the previous tree when material source records cannot be recovered. Decide explicitly how a trailing partial live write differs from a malformed committed record. Keep healthy source trees independently recoverable. Required proof: malformed middle and trailing records, preserved prior counts and provenance, explicit partial diagnostics, healthy replacement, and unaffected sibling-source recovery. Existing missing-child/source-failure preservation tests do not cover warning-bearing successful parses.

## ADP-02 — P2: configuration generation does not preserve valid configuration values

Locations: `backend/internal/adapters/config/writer.go:69-75` and 268-285; further manually interpolated strings at 99-105, 132-153, 175 and 302.

Trigger: generate configuration for a project name containing YAML syntax, or serialize a configuration with CodeMetrics.BaseBranch set. Plain interpolation turns `Project #2` into the YAML value `Project`. BaseBranch is not emitted at all.

Executed evidence: GenerateConfigYAML followed by yaml.Unmarshal changed ProjectName from `Project #2` to `Project` and BaseBranch from `release-main` to empty. The default-config path calls this same generator, making project-name corruption reachable during ordinary configuration generation; BaseBranch omission also violates the exported full-config serializer's contract.

Safe change: use typed YAML nodes/encoding for scalar values while retaining useful comments, and serialize every supported configuration field. Required proof: round trips for comment markers, colons, quotes, boolean/null-like strings, workspace paths and source roots, plus nondefault BaseBranch. Retain existing required-field checks. Current writer tests exercise ordinary strings but miss these values.

## ADP-03 — P2: discovered project configuration loses the directory that gives relative paths meaning

Locations: `backend/internal/adapters/repository.go:22-32`, then 37-51. Related test: `backend/internal/adapters/repository_test.go:235-263`.

Trigger: run a workspace command from a nested project directory. config.Load ascends to the project configuration, but ResolveWorkspacesDir returns its relative Workspaces.Dir unchanged. ScanWorkspaces then stats that path relative to the nested process directory.

Executed evidence: a temporary project with root configuration `workspaces.dir: spaces`, an existing root/spaces directory, and cwd=root/nested produced `Workspaces directory not found: spaces`. The existing nested-config resolver test accepts the literal returned directory without exercising a scan. CLI callers use this resolver without changing to the configuration directory.

Safe change: retain configuration origin and resolve configured relative paths against it; preserve deliberate cwd semantics for explicit CLI overrides. Required proof: nested directory scanning, load/migration/server callers, absolute configured paths, and explicit relative overrides. Do not hardcode this repository's root name.

## ADP-04 — P2: OpenCode workspace fallback ignores the configurable workspace root

Locations: `backend/internal/adapters/opencode/schema_v1_session_resolution.go:75-76` and `backend/internal/adapters/opencode/schema_v2_session_resolution.go:50-51`.

Trigger: source title does not identify the ticket and fallback must find a workspace path, but the adopting project uses a configured root such as `alspaces` rather than `workspaces`. Both schema-family resolvers only search `%workspaces/<ticket>/...`. The resolver boundary has no workspace-root input.

Static evidence: both SQL patterns contain the literal directory; resolution tests use the same conventional root. This does not affect an exact nonce/session-ID match that has already succeeded. It affects the fallback specifically and can prevent source recovery despite a matching configured workspace reference being present.

Safe change: supply the resolved workspace policy from the calling shell and construct escaped path-search patterns there or in a narrowly owned resolver. Required proof: both schema families with a custom root, title fallback, exact-ID/nonce precedence, path separators, and SQL LIKE wildcard characters treated according to a documented policy.

## ADP-05 — P2: default bind address is broader than the local-only product scope

Locations: `backend/internal/adapters/http/server.go:90-93`; global middleware at 65-66; `backend/internal/adapters/http/middleware.go:7-13`; live routes at server.go:57-59. The CLI advertises localhost at `backend/cmd/agentic-lens/main.go:562`.

Product boundary clarified by the user: AgenticLens is local-only; deliberate remote exposure and its access controls are the operator's responsibility. Absence of application authentication is not a defect under that supported scope. The original P1 classification over-weighted an externally reachable deployment assumption and is superseded by P2, normal-priority default hardening.

The remaining issue is default behavior: the listener address is `:<port>`, which binds available network interfaces rather than loopback. Starting the local dashboard can therefore accept LAN connections without an explicit remote-exposure option. Actual access depends on network reachability and firewall policy; internet exposure is not established. Live trace routes return local telemetry. No browser exploit was demonstrated, and no assumption is made about browser private-network protections. Wildcard CORS is contextual evidence, not an independent requirement to add authentication.

Minimal recommendation: bind loopback by default to match the supported local-only use and advertised URL. Do not add authentication, remote-host configuration or a new remote deployment mode solely to resolve this finding. Any deliberate tunneling/proxy exposure remains operator-owned. Required proof for a future implementation: loopback listener address and continued standalone embedded-UI/API access. No runtime change was made during this severity correction.

## ADP-06 — P2: optional child-lookup failure is indistinguishable from an absent child

Location: `backend/internal/adapters/live/service.go:441-447`. Existing regression: `backend/internal/adapters/live/sqlite_test.go:771-797`.

Trigger: required message/part reads succeed while ResolveChildSessionIDs fails, for example when parent_id is unavailable. The error is discarded and knownChildren remains empty, suppressing child links with no partial-data diagnostic.

Static evidence backed by the existing fixture: the test drops parent_id, requires the trace to survive, and explicitly expects ChildSessionID to be empty. Preserving the trace is correct. The missing behavior is reporting that linkage could not be resolved, rather than presenting an ordinary successful trace with no child. This finding does not ask that optional enrichment failure hide supported baseline trace data.

Safe change: preserve messages and parts and attach a typed optional-enrichment diagnostic/completeness indicator. Required proof: distinguish successful empty lookup, failed lookup with intact trace, and successful linked lookup; propagate the distinction through the HTTP DTO without turning the whole response into an error.

## ADP-07 — P2: required-column omission tests succeed without validating required columns

Locations: `backend/internal/adapters/opencode/schema_family_detection_test.go:40-84`, especially 43-48 and 77-79. Branch under test: schema_family_detection.go's mixed-family rejection.

Trigger: run any omission case. Each fixture creates both schema-family inventories and expects the mixed-schema error. The mixed-schema condition holds regardless of the omitted column, so the asserted rejection does not prove the advertised required-column behavior.

Executed mutation evidence: in a disposable backend copy only, changed `if !actual[col]` to `if false && !actual[col]`, disabling missing-column detection. `go test ./internal/adapters/opencode -run '^TestSchemaFamilyRejectsEveryRequiredTableColumnOmission$' -count=1` still passed (0.014s). Repository production/tests were untouched.

Safe change: start each case from one valid family, remove exactly one required column, and assert the relevant missing-structure diagnostic. Keep mixed-family rejection as a separate fixture. Required proof: the targeted column-validation mutation must fail the omission test, while complete v1/v2 fixtures pass. Preserve any meaningful mixed-schema assertions when restructuring.

## ADP-08 — P2: unit tests depend on mutable checkout configuration and branch inventory

Locations: `backend/internal/adapters/config/loader_test.go:2261-2282`; `backend/internal/adapters/git/inference_test.go:50-57` and 90-105.

Trigger: a developer legitimately changes project configuration or has a branch whose ticket matches one of the presumed-unbranched IDs. The loader test reads the live repository configuration and locks its base branch, source roots and anomaly settings. Inference tests invoke the actual Git-backed inference function while presuming selected tickets have no matching branch.

Static evidence: the assertions encode environment state rather than exclusively parser/inference semantics. This is not a claim that these tests fail in the present checkout. The nearby temporary YAML test demonstrates that the loader behavior can be tested independently of active project policy.

Safe change: use self-contained configuration fixtures and isolated temporary Git repositories with explicit branch inventories. Required proof: the tests pass with arbitrary valid repository configuration/current branch state and still fail when parser or branch-precedence behavior changes. Do not replace the Git behavior with a domain mock.

## ADP-09 — P2: several zero-write assertions observe an unrelated or absent directory

Locations: `backend/internal/adapters/live/sqlite_test.go:237-255` and 292-305; hydration fixture at 800-808 and its subsequent service construction/snapshot assertions.

Trigger: run the live tests in the package directory, where the literal relative `workspaces` directory normally does not exist. Snapshot assertions are guarded by Stat success and can both perform no checks. The hydration test separately allocates a temporary workspaces directory but does not configure that directory on the service it exercises.

Static evidence: these assertions cannot establish that writes did not occur at the service's configured destination. Other tests do provide stronger protection: later hydration failure-preservation coverage at sqlite_test.go:1393 onward configures a real temporary root and verifies contents, and HTTP tests snapshot concrete fixtures. This is a misleading local test claim, not evidence that live service currently writes.

Safe change: consolidate these assertions into the existing meaningful preservation tests, or configure and populate the exact temporary directory being monitored; treat traversal errors as failures and compare file contents as well as presence. Required proof: a temporary mutation writing to the configured workspace must fail the no-write test. Retain functional trace/hydration assertions when removing ineffective scaffolding.

## ADP-10 — P2: updating generated metrics deletes later peer sections in a summary

Location: `backend/internal/adapters/repository.go:310-338`, especially 323 and 332-336.

Trigger: SUMMARY.md has `### Workflow Efficiency` followed by another third-level section, such as `### Verification`. Replacement stops only at a second-level heading, so it consumes the peer section and its evidence.

Executed evidence: temporary summary containing a metrics section followed by `### Verification\nkeep-proof` became only the summary title and replacement metrics; `keep-proof` disappeared. The probe deliberately failed its preservation assertion. Existing replacement tests exercise a following second-level heading and miss this case.

Safe change: identify the exact generated block with stable markers, with a conservative legacy migration path, or correctly parse heading boundaries without treating code fences as headings. Required proof: preservation of peer and parent sections, headings in fenced code, legacy header compatibility, absent block append, and repeat-run idempotence.

## ADP-11 — P2: numstat parsing changes file identity before classifying code

Locations: `backend/internal/adapters/git/analyzer.go:23-41`; Git invocation around 211 does not request NUL-delimited paths.

Trigger: changed filenames contain repeated spaces, quoting-sensitive bytes, or Git rename display syntax. strings.Fields followed by strings.Join collapses whitespace inside the path; quoted/escaped Git paths and rename display text are not decoded into a real destination path. Classification then receives a different string than the file's path.

Static evidence: for the valid textual row `1\t0\tsrc/a  b.go`, the function necessarily produces `src/a b.go`. Exact include/exclude rules can therefore classify it incorrectly. Rename and quoted-path cases are additional protocol risks; they were not executed against a real Git repository in this review.

Safe change: request `git diff --numstat -z` and parse the documented path/rename protocol while retaining actual path bytes. Define which path owns a renamed file's classification. Required proof: repeated spaces, tabs, non-ASCII/quoted paths, source-to-test renames, binary rows and normal files, using a temporary Git repository as the boundary check.

## ADP-12 — P2: duplicated source decoding has drifted in transient raw trace fidelity

Locations: `backend/internal/adapters/live/service.go:580-645`, 715-723 and 1217-1258; peer decoder `backend/internal/adapters/opencode/schema_v1_session_extraction.go:496-551`.

Trigger: a string tool result has meaningful leading/trailing whitespace, such as indented source output. Live decoding calls rawMessageToString, which trims the decoded string at line 721; the offline v1 decoder preserves the decoded output string. Multiple decoder copies independently decide field extraction and status fallbacks, increasing the likelihood of further semantic drift.

Static evidence: the transformations differ directly. Existing parity tests cover ordinary records, not whitespace-sensitive raw output. This is a transient diagnostic-fidelity issue and a concrete source-ownership smell. It is NOT a finding that persistent ledgers should retain raw input/output: privacy-approved fact persistence deliberately omits transient payloads, and normalizeToolInput's compact projection must be assessed against that contract rather than described as prohibited evidence loss.

Safe change: establish one version-specific decoded source representation for fields shared by offline extraction and live views, then let consumers own their explicit projections. Preserve raw strings in the transient raw-trace view; keep persistence privacy rules and imperative cache/SQLite ownership intact. Required proof: v1 live/offline parity for whitespace-bearing strings, structured values, malformed/unknown states and ordinary records, plus existing retention tests proving no raw-payload expansion in serialized ledgers. Do not create a speculative universal abstraction across unrelated source formats.

## ADP-13 — P3: production SQLite API exposes a test-only database handle

Location: `backend/internal/adapters/live/sqlite.go:131-134`.

Static evidence: the exported DB accessor says it exists for test introspection and its repository use is the read-only-pragma test. This exposes the underlying handle as public production API solely to inspect an implementation detail. The connection is opened read-only, so this finding does not claim a demonstrated write bypass.

Safe change: move the inspection into a package-internal test or use a narrowly scoped boundary test without exporting the handle. Required proof: keep read-only enforcement coverage and compile all consumers after removal. This is small API debt; it should not become a large abstraction rewrite.

## Other assessed boundaries and limits

- Persisted facts versus optional hydration: reviewed successful and unavailable-source paths and the later preservation tests. Do not remove persisted facts merely to normalize optional enrichment. ADP-01 and ADP-06 identify specific remaining completeness failures.
- FCIS ownership: SQLite, filesystem, clocks, HTTP and source discovery remain shell responsibilities. DTO separation is useful; the concrete ownership concern is duplicated source interpretation in ADP-12, not a recommendation to move I/O into domain functions.
- SSE: the global server WriteTimeout alone is not an infinite-stream defect because LiveStreamHandler explicitly clears the write deadline. No such finding is asserted.
- Live tree leaves are initialized with nonnil empty SubSessions slices in service/domain tree construction. Normal leaf serialization is not evidence of a frontend null-child crash.
- HTTP windows `1d`, `7d`, `30d`, `90d` use server-current time; `custom` relies on explicit since/until parameters. The frontend owner is reporting mismatched client filtering semantics; this report does not duplicate that finding.
- Core/CLI owners are consolidating AL2 parser/config-threshold issues and successful-output/error-text heuristic concerns; those shared findings are not duplicated here.
- No full-suite rerun was performed for unchanged source. The parent supplied the already completed `make` result at 49cf350 (including 781 frontend tests); that is prior validation, not evidence that the defects above are absent. No new timing profile was collected, so this report makes no claim about the slowest tests.

## Targeted verification record

Disposable backend copy: `/tmp/rev5-adapter-audit-vnbstddy`; copied module files/internal tree and added two temporary probe files. This location is not a repository deliverable and is not required by committed tests.

Command: `go test ./internal/adapters ./internal/adapters/backfill -run '^TestREV5' -count=1 -v` in that copy. Result: expected nonzero exit; all four preservation/round-trip probes reproduced ADP-01, ADP-02, ADP-03 and ADP-10 as described above. These were defect probes, not a passing regression suite.

Separate mutation command: `go test ./internal/adapters/opencode -run '^TestSchemaFamilyRejectsEveryRequiredTableColumnOmission$' -count=1` after disabling missing-column detection only in the copy. Result: pass, confirming ADP-07's test blind spot. No mutant was copied into the checkout.

Coverage manifest: ADAPTERS_COVERAGE.json has one entry for each of the 82 assigned paths. All reviewed source hashes were checked against SCOPE.json before handoff.

## Subsequent remediation status

Iteration 04 implements ADP-01 and D3 in commit 302b3d0. Independent review approved the scoped changes; see ../04/REVIEW.md and TESTING.md for evidence and limits. This historical audit describes its original frozen baseline. All other findings remain open, including ADP-05 at P2 under the clarified local-only scope.
