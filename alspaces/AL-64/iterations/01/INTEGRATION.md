# AL-64 routine upstream integration

Status: resolving conflicts; not final delivery approval.

On 2026-11-05 the user requested merging main into this branch. Pending AL-64 planning documents were checkpointed at e9ca2676 before fetching. Merge target: origin/main 40b16def, containing AL-66 optional metric output, AL-67 architecture assessment and AL-77 cross-ticket session overlap warnings. Merge base: a84adbe5.

This is routine integration within existing AL-64 implementation scope. Iteration02 remains proposed and unapproved. No hybrid caller/liveness implementation, session poke, final telemetry or PR publication is authorized by this merge request. No direct user message quotations are retained.

## Ownership and preserved behavior

Coder owns all15 conflicted product/test files and necessary adjacent compatibility fixes; parent owns staging and merge commit. Shared checkout edits preserve both reviewed AL-64 behavior and upstream changes. Independent tester then owns meaningful missing integration coverage; reviewers remain product/test read-only.

Preserve completed/provisional/unsupported duration semantics, all six fixed regressions, deterministic cutoffs, exact-owner replacement/retention, compact evidence and duration displays. Preserve AL-66 runtime optional metrics and absent/empty/nonempty coverage without resurrecting removed persisted policy markers. Preserve AL-77 full-loaded-dataset overlap detection, async reload/error states and ticket navigation.

## Initial conflicts

- backend/cmd/agentic-lens/capture.go
- backend/cmd/agentic-lens/main.go
- backend/internal/adapters/backfill/engine_test.go
- backend/internal/adapters/capture/reconcile.go
- backend/internal/adapters/http/comparison_dto.go
- backend/internal/adapters/http/fleet_dto.go
- backend/internal/adapters/opencode/capture_snapshot.go
- backend/internal/adapters/opencode/extractor.go
- backend/internal/adapters/opencode/schema_v1_session_extraction.go
- backend/internal/adapters/opencode/schema_v1_session_extraction_test.go
- backend/internal/adapters/opencode/schema_v2_session_extraction.go
- backend/internal/adapters/schemas_test.go
- backend/internal/domain/models.go
- web/src/api/validation.ts
- web/src/views/FleetOverview.tsx

## Verification and review plan

After conflict resolution, verify affected Go/CLI/source/DTO/backfill tests and frontend types/build. Run required integrated make and verify the standalone binary serves canonical embedded assets. Original pre-merge tests/reviews remain historical evidence, not approval of changed interactions.

All four integration focuses run both configured profiles: reviewer-behavior (policy/cutoff/source/UI coexistence), reviewer-data (optional evidence versus omission/retention), reviewer-architecture (composition signatures and responsibility), reviewer-tests (preserved assertions and interaction coverage). Profiles: gpt-6.1-sol/high and gpt-6-luna/medium. Unique reports: INTEGRATION_BEHAVIOR_SOL.md, INTEGRATION_BEHAVIOR_LUNA.md, INTEGRATION_DATA_SOL.md, INTEGRATION_DATA_LUNA.md, INTEGRATION_ARCHITECTURE_SOL.md, INTEGRATION_ARCHITECTURE_LUNA.md, INTEGRATION_TESTS_SOL.md and INTEGRATION_TESTS_LUNA.md under this iteration. Freeze one merged candidate and use capacity-limited waves; parent owns consolidated integration verdict and exact revision attribution.

Final ticket approval remains gated on iteration02 scope resolution. Integrated checks and local merge completion do not publish a PR or authorize release.
