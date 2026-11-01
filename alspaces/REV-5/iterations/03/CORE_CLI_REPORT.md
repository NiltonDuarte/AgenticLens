# REV-5 iteration 03: core and CLI audit

Verdict: **CHANGES_REQUESTED**. Audit only; no application or test changes made.

Baseline: `d5227812b09a17ff15d7e6ffe058e96ecf1616eb`. Scope: all 48 `core_cli` paths in `SCOPE.json`: full semantic inspection of all 46 Go production/test files, including every test body; structural inspection of `go.mod` and `go.sum`. Findings concern current code, not only the baseline diff. `CORE_CLI_COVERAGE.json` records each path exactly once.

## Findings

### CC-01 — P2: successful file contents become failed tool executions

Locations: `backend/internal/domain/live.go:103-118`; callers `backend/internal/adapters/opencode/schema_v1_session_extraction.go:507` and `backend/internal/adapters/live/service.go:645`.

**Confirmed:** `ClassifyToolExecution("read", "completed", "README: handle exit code 2", 0)` returns `IsError=true`, medium severity; the same successful read containing `docs: segmentation fault means a crash` returns high severity. The classifier searches arbitrary output for exit/crash text across every tool, despite its process-level-only contract. The callers feed file-read output into this function and turn the classification into error counts/event statuses. This is ordinary successful content inspection, not an explicit failed source status, and remains after the AL43 semantic-error rollback.

Fix: carry authoritative status/process-exit evidence separately from content. Restrict any required legacy output fallback to execution tools and a recognized execution envelope. Preserve explicit `error`/`failed` behavior. Verify successful reads/search results containing these phrases, real process failures, grep/glob no-match behavior, and both live/offline consumers. Existing `tool_execution_test.go` and `live_test.go` cover other successful error-like text but omit these residual triggers.

### CC-02 — P2: partial generate refresh removes historical skill/task facts

Location: `backend/cmd/agentic-lens/main.go:279-283`, followed by session merge at `:301-303`.

**Confirmed with the CLI:** a canonical cached ledger containing remote session `remote` and skill `historical-remote-skill`, refreshed against a local SQLite source containing only session `local`, exits 0 and retains both sessions but omits the historical skill. Cached `Skills`/`TaskLinks` are replaced wholesale whenever any OpenCode session is extracted. Session merge deliberately retains unavailable session facts, so auxiliary facts acquire different preservation semantics. The data model lacks owning-session identity for an exact partial merge.

Fix: use the existing backfill snapshot-completeness policy for generate: replace auxiliary snapshots only with complete authoritative coverage; otherwise retain and report the previous snapshot. Introduce ownership only if precise incremental replacement is required. Verify missing-root, cross-device, mixed-source and complete-refresh cases, including genuine deletion after a complete refresh. Existing multi-device generation tests protect invocations/iterations but do not assert auxiliary facts; backfill tests already demonstrate retention semantics.

### CC-03 — P2: configured anomaly thresholds are dropped across projections

Locations: `backend/cmd/agentic-lens/main.go:407-411`; `backend/internal/domain/aggregations.go:872`; persistence removes event rows at `backend/internal/domain/events.go:274`.

**Confirmed in an isolated domain harness:** two consecutive errors with configured warning=2/critical=4 produce one anomaly in `DeriveMetricsLedger`. The CLI-equivalent `PersistAnomalySummaries(facts)` yields zero after reload; passing the configured thresholds yields one. `CalculateFleetSummary` of the correctly configured fresh metrics also reports zero because it recomputes from facts with defaults. Dropping events makes the persistence mismatch irreversible without source re-extraction.

Fix: pass effective project thresholds into persistence and avoid default re-detection when fleet input already contains configured projections (or explicitly provide the same config). Verify generate/save/reload/detail/fleet parity for thresholds both below and above defaults, keeping configured live rules aligned. Existing anomaly tests mostly exercise defaults; configurable aggregation tests check health ceilings, not this persistence/fleet path.

### CC-04 — P2: advertised error-rate threshold has no effect

Locations: `backend/internal/domain/anomalies.go:218,271-275`; config field `backend/internal/domain/config.go:200`; documented meaning `docs/SETUP_MANUAL.md:301`.

**Confirmed:** setting `ErrorRateThreshold=0.90` and supplying 3 errors in 10 calls still emits a high-failure anomaly. The function accepts thresholds but compares fixed 0.20 and 0.50; repository references show the field is loaded, written and normalized, never used for detection. The documented default is 0.25 while detection uses 0.20.

Fix: define how the configured alert threshold relates to warning/critical levels, then apply it consistently and render the actual threshold. Alternatively retire the setting and documentation explicitly if fixed levels are intended. Verify custom thresholds and exact boundaries in event and aggregate fallback paths. Configuration serialization tests currently prove only that the unused number survives loading.

### CC-05 — P2: init silently overrides its own database-path argument

Location: `backend/cmd/agentic-lens/main.go:1042-1047`.

**Confirmed with isolated HOME:** `init --db-path <temp>/custom.db` writes that path into shared config but also writes the default OpenCode path into the higher-priority local config. A subsequent generate without `--db-path` attempts `<isolated-home>/.local/share/opencode/opencode.db`, not the requested database. `ResolveOpenCodePath` gives the local source path precedence.

Fix: derive the local source setting from the requested path or omit the local override when shared configuration already supplies it. Verify the effective resolver/next CLI invocation, not only `Database.Path` in the generated YAML. The custom-path init test checks the lower-priority shared field and misses the effective result. Also align local-file placement with the shared `--output` directory when custom output is supported; this placement issue was inspected statically, not separately reproduced.

### CC-06 — P2: ticket parsing rejects configured prefixes containing digits

Locations: `backend/internal/domain/aggregations.go:13,33-34,1180-1188`.

**Confirmed:** `ExtractTicketPrefix("AL2-10")` returns `AL`; `CompareTicketIDs("AL2-10", "AL2-2")` returns -1. The parser treats the first digit as the ticket sequence, so filtering by configured prefix `AL2` loses those tickets and natural ordering becomes lexical. The checked-in configuration supports `AL2`, and configured normalization uses a different grammar. Frontend review independently found the matching UI issue; consolidate across layers.

Fix: use one agreed grammar/configured-prefix contract across normalization, filtering and ordering, preserving explicitly supported legacy delimiterless IDs. Verify `AL2-2 < AL2-10`, AL versus AL2 filtering, case, delimiters, leading zeroes and legacy IDs. Existing extensive sort/prefix tables contain alphabetic prefixes only.

### CC-07 — P2: descending comparator returns true for equal rows

Location: `backend/internal/domain/aggregations.go:1375-1378`.

**Confirmed:** sorting three equal-`UpdatedAt` rows A/B/C descending repeatedly yields C/B/A, then A/B/C, then C/B/A. `return !less` violates strict ordering because equal operands compare true in both directions. Duration/invocation ties and equal ticket/time identities also use this branch. Cost and completed-at use separate valid comparators.

Fix: reverse operands or use a three-way comparison; equality must remain false. Keep the chosen stable identity/tie policy explicit. Verify repeated sort idempotence and equality in each affected field. Existing equal-value assertions focus on cost, whose independent branch avoids this defect.

### CC-08 — P2: scan diagnostics disappear from the root CLI

Location: `backend/cmd/agentic-lens/main.go:1326-1335`.

**Confirmed with the CLI:** one valid canonical ledger plus one malformed sibling produces exit 0, empty stderr and `Found 1 ticket workspaces`. The command consumes only `report.Ledgers`; scan rejection/error diagnostics are never reported. A request for the rejected ticket instead says not found, hiding recovery information.

Fix: surface per-ticket rejection diagnostics and any fatal scan error while continuing to present valid siblings according to the established partial-scan contract. Verify mixed valid/corrupt, unsupported-schema, unreadable and target-rejected cases. Existing CLI tests cover empty scans and missing tickets, not mixed accepted/rejected siblings.

### CC-09 — P3: v1 migration mutates its input through a shallow copy

Location: `backend/internal/domain/migration/step_v1_v2.go:34-43`.

**Confirmed:** applying `NewStepV1ToV2().Apply` to a ledger with one session whose `Events` is nil makes the original input session's `Events` non-nil. `cloned := ledger` shares the sessions backing array. This violates functional-core value semantics and weakens the migration engine's before/after comparison when both values alias.

Fix: clone the session slice before assigning event fields and audit appendable migration metadata ownership similarly. Verify input snapshots and alias independence for nil/non-nil event collections. Existing step tests inspect transformed output but do not assert input nonmutation for this case.

### CC-10 — P3: discarded work and duplicated model aggregation remain in the core

Locations: `backend/internal/domain/aggregations.go:681-689,607-644`; `backend/internal/domain/events.go:302-310`.

**Static evidence:** `mappedEvents` is fully built and role-mapped but never read; session events are separately mapped later. `anomalyEvidenceSelection.events` is allocated and populated but never consumed. Thread role/model aggregation duplicates the grouping, projection and cost comparator already implemented by `ComputeModelMetrics` (`aggregations.go:480-523`), requiring parallel maintenance of cost and health semantics.

Fix: delete the unused array/map work and call the existing model aggregation with the thread ceiling. No new abstraction is needed. Preserve existing cost precision, incomplete-cost ordering, role ordering and custom-ceiling tests (`TestComputeRoleAndModelMetrics`, thread-order tests and `TestConfigurableHealthyCeilingAggregations`). No performance timing was taken; only the extra allocation/traversal is established, not a measured speedup.

### CC-11 — P3: sorting test does not exercise its stated boundary; fixtures contain discarded setup

Locations: `backend/internal/domain/anomalies_test.go:1918-1947,627-647`.

**Static evidence:** the sorting tie test constructs both nil-turn and non-nil-turn anomalies in `sorted`, then never passes that slice to production code and ends with `_ = sorted`. Its actual assertion checks a single anomaly, which cannot test the intended ordering. The rate-parity test builds events in two loops and then overwrites the entire slice with a literal before detection.

Fix: remove discarded setup; either exercise the nil-turn tie through an achievable public input or remove the unsupported claim/unused case while preserving the legitimate tool-name sorting assertion. Retain the literal parity fixture and its existing rate/count/severity assertions. Consolidation opportunity: `TestGetLedgerCompletionDate_NoFallback` repeats the valid/nil/empty/bad cases already covered by `TestGetLedgerCompletionDate`; merge unique cases into the earlier table. These are concrete subtraction opportunities, not a test-count target. No timing savings claimed.

### CC-12 — P3: server smoke test can pass against an unrelated service and its readiness timeout is not bounded

Locations: `backend/cmd/agentic-lens/main_test.go:102-137,421-446`.

**Static evidence:** the server smoke test uses fixed port 19090 and asserts only HTTP 200. If another service owns that port, the spawned binary can fail to bind while the test succeeds against the other service. The shared readiness helper checks a wall-clock deadline around `http.Get`, but the default client has no request timeout, so a connection that accepts without responding can outlive the advertised deadline indefinitely.

Fix: use isolated listener allocation as peer tests already do, verify the expected fixture identity/config, observe child exit, and give readiness requests an explicit deadline. Verify occupied-port failure and an accepting-but-nonresponding fixture. No live service or occupied port was probed during this audit; this finding follows directly from process/HTTP control flow.

## Architecture and test assessment

The production domain contains no newly discovered filesystem/network/clock dependency; migration decoding is pure transformation. Session-owned invocation facts, explicit cost evidence, and live/offline boundaries provide useful separation. The principal architectural weaknesses are inconsistent refresh/preservation policies between CLI paths, configuration omitted at projection boundaries, arbitrary output mixed with execution status, and shallow mutation inside a nominally pure migration step. The large test files provide meaningful migration, precision, hierarchy and error-path checks; their size alone is not a finding. They also contain repeated default-case matrices while the reproduced cross-boundary cases are missing.

## Verification and limits

- Built the unchanged CLI with `go build -o /tmp/rev5-core-audit-cli ./cmd/agentic-lens` from `backend/`; succeeded.
- Copied the domain tree into a temporary `agenticlens` module and ran a standalone harness against the unchanged production implementation. It reproduced CC-01, CC-03, CC-04, CC-06, CC-07 and CC-09. The harness used synthetic values only.
- Ran the built CLI against `/tmp/rev5-core-cli-evidence-b_dgftp6`, a synthetic project/SQLite store with HOME redirected into that directory. This reproduced CC-02, CC-05 and CC-08. No live ticket fixture or repository source was changed.
- Reused the parent's valid unchanged-source `make` baseline (`49cf350`, including 781 frontend tests). No whole-suite rerun, race run, benchmark or dependency-vulnerability analysis was performed by this reviewer. No claim that the baseline suite detects the new findings.
- Full semantic inspection means every assigned Go file and test body was read, not that all possible runtime paths were executed or every smell is excluded. Manifest checks cover dependency declarations/checksum structure only. Reported diagnoses distinguish reproduced behavior from static control-flow evidence; suggested fixes remain unimplemented.
