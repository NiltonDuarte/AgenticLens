# FIX-20: Stop source reads when callers cancel

## Goal

Cancelling a request or backfill operation stops its in-flight native work promptly and releases resources without discarding accepted facts.

## Context

Some consumers accept a context but lose it before expensive work. Live SQL uses non-context queries, Codex discovery/parsing lacks cancellation, and backfill can call exact snapshots with a background context despite a workspace deadline. Cancellation is checked between owners, which does not interrupt the owner currently being read.

Example: closing a historical-trace request during a large rollout scan leaves the scan running. A slow live read can also keep the shared trace-cache lock occupied after its caller leaves. Latency and production incident frequency were not measured.

## Acceptance criteria

* Already-cancelled and mid-read cancellation stop native work at the relevant query/read boundaries. A context parameter that is ignored below the interface does not satisfy acceptance.
* Cancellation reaches public HTTP/backfill consumers with an explicit error or partial result; it cannot certify an empty complete snapshot or replace an unavailable owner's stored facts.
* Read connections, file/gzip readers and snapshot transactions clean up on cancellation. Preserve exact capture's dedicated-handle/snapshot assumptions.
* A cancelled trace read does not leave the cache lock held indefinitely or prevent subsequent independent reads. Preserve accepted cached snapshots, revision invalidation and useful partial data.
* Normal uncancelled v1/v2/Codex extraction and mixed-owner history remain unchanged. Record the controlled test conditions; do not claim a universal latency SLA or unmeasured speedup.

## Implementation notes

Finding references: F06

No global timeout-policy redesign, new source repository framework or speculative cache optimization. REV-7 investigates wider read/lifetime guarantees; this concrete cancellation correction can proceed independently.

* [AL-67 assessment](../../AL-67/ARCHITECTURE_ASSESSMENT.md), F06; application baseline a84adbe5be46c2ef52ede0987a417562715d38e1. Findings are unimplemented.
* [Follow-up ownership](../../AL-67/iterations/02/FOLLOW_UP_TICKETS.md).

*DEVELOPMENT STATE REVIEW (2026-11-05)*

Still open: live/sqlite.go uses non-context SQL reads and backfill/recovery.go calls CaptureSessionSnapshot with context.Background(). Retain cancellation and stored-fact preservation acceptance.

Evidence and review limits: [PO-19 reconciliation](../../PO-19/SPEC.md), merged baseline `03f5b05d`. This planning review does not certify runtime behavior or close unverified acceptance criteria.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [relevant regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. Historical findings require revalidation; this rewrite is not runtime verification or implementation approval.
