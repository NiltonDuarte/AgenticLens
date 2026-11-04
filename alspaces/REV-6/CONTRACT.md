# REV-6 technical contract

## Problem and authorized scope

Review test performance and actual unit/integration boundaries, beginning with capture adapters and the native CLI. The reported duration near 43 seconds is an unverified observation, not an individual-test baseline. Deliver measured findings and bounded follow-up recommendations for inexpensive unit checks that gate slower integration checks.

This is a review-only ticket. Application code, tests, fixtures, schemas, manifests, CI workflows, runtime scripts, timeouts and coverage remain unchanged. No coder or implementation phase is authorized. Later remediation requires its own approved ticket; reuse AL-74's existing ownership rather than duplicating its reliability and cleanup work. Unrelated concrete defects are routed to separate FIX tickets by the parent under the management rule.

## Baseline and source authority

Initial checkout: clean, branch REV-6/review-test-performance-unit-integration-boundaries, commit a84adbe5be46c2ef52ede0987a417562715d38e1. The parent fetched origin/main and confirmed that same revision. AL-62's delivered work is merged PR #109, commit 86bea597; closed unmerged PR #103 is not delivery evidence. Reconfirm revision and worktree status before profiling; freeze one revision for comparable measurements and independent review.

Workspace root is alspaces from .agenticlens-config.yaml. REV-6 is the supplied existing identifier even though the configured prefix list does not enumerate REV. Do not allocate a second ticket or alter prefix configuration for this review.

## Acceptance boundaries

- Separate per-test/subtest elapsed time, package elapsed time and whole-command wall time. Attribute compilation, package TestMain fixture builds, subprocess startup, intended waits, race instrumentation and resource consumption where evidence supports attribution.
- State hardware/OS, toolchain, source revision, cache assumptions, race settings, package/test concurrency and fixture controls for every timing cohort. Distinguish observations, source-derived hypotheses and unmeasured estimates.
- Classify tests by exercised boundaries. Pure deterministic transformations belong to unit coverage; real filesystems, SQLite, Git, processes, native entrypoints, lifecycle hooks and external harness protocols require explicit boundary classification. An integration test may be fast; a slow test is not thereby integration coverage.
- Preserve distinct oracles and native ownership, current-installation identity, pending allocation, generation ACK, failure/cancellation, fencing and realistic fixture guarantees in every recommendation. Any consolidation recommendation names an assertion recipient and the boundary still requiring independent proof.
- Recommend practical measured budgets, local commands and hosted CI dependencies that prevent slower integration execution before fast unit success. Cover shared fixture-build ownership, frontend asset prerequisites, race coverage and Windows-specific tests without claiming unexecuted platform proof.
- Deliver prioritized, bounded follow-ups, evidence limits and AL-74 overlap. Never recommend blanket timeout inflation, silent skips, arbitrary coverage removal or test-count targets.

## Architecture and data guarantees

FCIS boundaries remain authoritative: backend/internal/domain stays pure, backend/internal/adapters owns external I/O, backend/cmd/agentic-lens composes the native shell, and web owns presentation. Single-binary asset embedding, strict typing and zero emojis remain unchanged. Profiling must use disposable projects/state and existing self-contained fixtures; do not mutate the installed capture state, private harness histories or unrelated ticket artifacts. Raw transcripts remain outside Git. Supported portable ledger facts, native source identities and uncertainty remain intact.

## Deliverables and completion

The parent owns SUMMARY.md, PULL_REQUEST.md, review consolidation/retention, final telemetry, ticket lifecycle, commits and PR publication. Iteration 01 owns its PLAN.md and measured review evidence. Completion requires comparable measurements, an explicit classification inventory, sequencing recommendations, scoped follow-ups and honest limits; a plan or a successful aggregate suite alone does not satisfy the review. Source-grounded telemetry is generated only at final delivery, with explicit Codex source and the review's chosen conventional type, and verified for root/child coverage before commit.

## Approval state

The user explicitly approved this contract and iteration 01 plan on 2026-11-04. Approval authorizes the review and disposable measurement, without authorizing implementation.
