# REV-5 independent upstream integration review

Status: APPROVED

Reviewed incoming main 47af923 (AL-57 cost availability/schema v5) against premerge aa86f4b71abe6cbb91a43e855b91bbc11fb31515, plus the resolved staged source/test diff and retained REV-5 changes against origin/main. Read the contract, integration plan, cost semantics, and implementation/test reports. Reviewed both explicit conflicts and automatic merges. No blocking defect found in the inspected incoming code or integration.

## Contract and correctness

- The split fact DTOs preserve presence-aware decoding and explicit nullable amounts. Invocation/task decoders distinguish omitted/null from numeric zero; v5 validates required members and evidence, while v4 compatibility retains historical observed zero. Domain accumulators keep internal numeric compatibility separate from precise cost summaries, and HTTP/export DTOs derive complete amounts from evidence. Persisted ingestion and computed-output contracts remain separate; deleted reverse-conversion tests do not justify restoring an unused API.
- OpenCode v1/v2 extraction tracks reported versus missing source costs and retains partial evidence when records coalesce. New Codex extraction marks unavailable cost. Task-link cost remains auxiliary and is excluded from invocation totals. Existing owner-snapshot refresh and unavailable-source baseline retention remain intact.
- Domain summary composition propagates complete, partial, unavailable and empty states through totals, role/model/thread projections, fleet and cumulative time buckets. Reviewed precise-subtotal versus rounded-scalar handling and model ordering. Incomplete cost does not become an observed zero or a complete cumulative amount. Cost-derived role percentages require a complete nonzero denominator.
- Migration uses version-specific typed inputs and presence probes before domain assembly. v4-to-v5 preserves numeric-zero/missing distinctions and ownership without creating a source-backfill requirement; the pre-v4 recovery rule remains tied to the v4 boundary rather than CurrentSchemaVersion. Implicit memberPresence/JSON/YAML codec methods remain reachable through decoders and must not be deleted based on identifier-only scanning.
- Frontend shape and financial validation compose in the existing decoders with one response-body read. Nullable costs/maps, legacy omitted invocation cost, partial evidence, required v5 members and summary coverage contradictions are handled at named schema paths. Dynamic role/model keys and source/artifact payloads remain opaque. Existing malformed non-cost rejection, omitted artifact descriptions and malformed-header warnings preserving valid sibling data remain.
- UI cost formatting, cards, distribution shares, chart gaps, progression/scatter and complete-only comparison averages were inspected. Observed zero is eligible for Average Spend; partial/unavailable tickets remain in other metrics, and known subtotals are labeled. Missing values are not plotted as zero. No browser geometry claim follows from this source review.

## Subtraction, architecture and coverage

The merge preserves removed ExtractFacts, MigrateLedger, CalculateRoleEffortRollup and obsolete frontend fetch APIs. The reintroduced ValidateToolFacts/containsString and new test-only SummarizeCost wrapper were removed with assertions retargeted to active extraction/aggregation APIs. Reviewed representative test transfers, including the intentionally different active role-versus-fleet rounded scalar behavior. Tests of deleted output-only reverse conversions and unused request readers are explicitly identified as obsolete; real persisted-codec, SQLite, HTTP and UI ownership remain independently covered.

The DTO split is justified by distinct persisted, compact-export and HTTP contracts; no generic speculative reverse-conversion framework was added. Pure domain calculation/version decoding remains free of I/O and clocks. Original REV-5 configuration cleanup/rollback and actual live revision cache survive the merge. Codecs, routes and legacy configuration acceptance are preserved. No newly introduced any/interface{} or source-workspace fixture dependency was identified in the inspected delta.

Reused coder declaration/reference screening: all 127 Go files report only implicit codec candidates; the frontend named-function scan reports no zero-reference candidates. These scans are conservative and are not a binding-aware call graph or proof that every unused declaration is absent. The incoming production scope was screened broadly with deeper examination of cost, decoding, migration, projection and UI calculations; this is not an exhaustive numerical or whole-program proof.

## Verification and remaining gate

Reused documented backend/domain race runs, frontend typecheck/lint/build and affected 232-case suite, final client rerun, and independent tester's uncached cost-codec/migration race checks plus 113 client/warning cases. Actual merged-Go HTTP smoke used isolated v4 observed-zero, v4 missing-cost and v5 partial fixtures and accepted fleet/list/detail payloads through the merged decoders. That is coder evidence reused by review, not a second independent server execution. No tests were silently represented as run by this reviewer.

Reviewer staged and working-tree whitespace checks passed. Root still owns merge commit, integrated make, final standalone asset verification, complete Codex telemetry and PR delivery. Approval does not claim those pending gates have passed. Provider source hydration, full release execution, browser geometry and performance profiling remain unverified here.

Reviewed staged backend/web diff SHA-256: 43c5e89503c35ad0f89f9f7cc458f6e2e855c36dd6ea25fce7ec5df060440945.
