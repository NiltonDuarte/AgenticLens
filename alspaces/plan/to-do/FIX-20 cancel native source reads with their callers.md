h3. Title: FIX-20: Cancel native source reads with their callers

*TARGET*: AgenticLens (Source Adapters / Live and Historical HTTP / Backfill)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: AL-67 architecture assessment; user-requested follow-up ticketing.
*Findings*: F06

*CONTEXT*

Some consumers accept a context but lose it before expensive work. Live SQL uses non-context queries, Codex discovery/parsing lacks cancellation, and backfill can call exact snapshots with a background context despite a workspace deadline. Cancellation is checked between owners, which does not interrupt the owner currently being read.

Example: closing a historical-trace request during a large rollout scan leaves the scan running. A slow live read can also keep the shared trace-cache lock occupied after its caller leaves. Latency and production incident frequency were not measured.

*GOAL*

Cancelling a request or backfill operation stops its in-flight native work promptly and releases resources without discarding accepted facts.

*VERTICAL SLICE SCOPE*

* Native reads: propagate caller cancellation through reached OpenCode v1/v2 SQL and Codex discovery/file-read loops.
* Consumers: carry the correct request/workflow context through live trace/tree, historical hydration and backfill; preserve explicit partial/error semantics.
* Quality: use controlled disposable blocked/slow reads and cancellation/cleanup evidence rather than timing-only sleeps.

*ACCEPTANCE CRITERIA*

* Already-cancelled and mid-read cancellation stop native work at the relevant query/read boundaries. A context parameter that is ignored below the interface does not satisfy acceptance.
* Cancellation reaches public HTTP/backfill consumers with an explicit error or partial result; it cannot certify an empty complete snapshot or replace an unavailable owner's stored facts.
* Read connections, file/gzip readers and snapshot transactions clean up on cancellation. Preserve exact capture's dedicated-handle/snapshot assumptions.
* A cancelled trace read does not leave the cache lock held indefinitely or prevent subsequent independent reads. Preserve accepted cached snapshots, revision invalidation and useful partial data.
* Normal uncancelled v1/v2/Codex extraction and mixed-owner history remain unchanged. Record the controlled test conditions; do not claim a universal latency SLA or unmeasured speedup.

*SCOPE LIMITS*

No global timeout-policy redesign, new source repository framework or speculative cache optimization. REV-7 investigates wider read/lifetime guarantees; this concrete cancellation correction can proceed independently.

*DEFINITION OF DONE*

* Revalidate the AL-67 source evidence at the implementation/review revision. Record actual checks, remaining limits and disposition; do not claim a production incident from a static counterexample.
* Work in a separate dedicated ticket session under the appropriate development/review workflow. Preserve portable ledger history, exact source/session/parent identity, legitimate repeated occurrences and explicit unavailable values. No application implementation is authorized by this planning document alone.
* Use self-contained fixtures; keep I/O outside the pure domain and follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md. Implementation requires affected checks, independent review and the required make gate; frontend changes also require canonical embedded-asset verification. Review-only work records evidence and recommendations without application/test edits.
* Publish the scoped deliverables and source-grounded telemetry. Do not mark completion from passing tests or a proposed design alone.

*SOURCE EVIDENCE*

* [AL-67 assessment](../../AL-67/ARCHITECTURE_ASSESSMENT.md), F06; application baseline a84adbe5be46c2ef52ede0987a417562715d38e1. Findings are unimplemented.
* [Follow-up ownership](../../AL-67/iterations/02/FOLLOW_UP_TICKETS.md).
