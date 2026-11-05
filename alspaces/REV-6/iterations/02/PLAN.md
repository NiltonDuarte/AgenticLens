# REV-6 iteration 02: completed test repair and sequencing

User explicitly approved CONTRACT_DELTA.md and this plan on 2026-11-05. Baseline 23b38163 includes main 40b16def. This concise retained plan replaces completed working detail at the user's artifact-cleanup request; original detail remains in Git history at 2b91574d. The original contract and iteration01 are historical.

## Approved implementation

1. Repair the two schema tables with valid ticket_id and intended-error assertions; preserve valid decode, nil/empty semantics and canonical rewrite. Do not change runtime validation.
2. Partition Go into internal/domain/... units and the complete complement; preserve later domain race verification and complement race coverage. Reject failed/empty discovery; use self-contained fixtures.
3. Partition frontend pure domain/utilities into Node units and the complementary jsdom/annotated suite. Centralize selectors and strictly typecheck configs. Standalone npm integration runs all frontend units first, without Go/Make prerequisites.
4. Make/CI integration, checks and setup require both languages' units, including parallel/mixed goals. Share dependencies/assets. Preserve build-only overrides, paths with spaces, native/storage/ownership/WAL/ACK/fencing/platform assertions and selected historical-release interfaces. No dependencies, skipped tests, timeout inflation or runtime feature changes.
5. Prove complete disjoint discovery/case accounting, meaningful schema-oracle mutations, real failing-unit blocking and positive reachability. Cover Make aliases/default/parallel paths, frontend-only npm, CI edges and historical release command compatibility.

## Ownership and completion

Coder owned Make/npm/Vitest/CI/compatibility orchestration and relevant contributor guidance. Tester owned schema tests and orchestration/release regression checks. Both ran gpt-6.1-sol/high, shared the checkout and preserved others' work. Parent owned artifacts, integration, telemetry, commits and publication.

Independent behavior, data/failure, architecture and test-effectiveness reviewers each ran gpt-6.1-sol/high and gpt-6-luna/medium. Four paired behavior/architecture correction reviews verified affected changes; supplied-finding verification receives no fresh discovery credit. No extra performance specialist was needed because implementation claims no new performance budget/saving.

Final tested source 8280c2bd passed make and canonical/standalone embedded-asset checks. Earlier failed candidates were corrected without weakening release guards or runtime validation. Parent consolidated APPROVED after required checks; unchanged data scope carries forward with explicit rationale. All six fixed findings/reporters remain in REVIEW_SUMMARY.json; detailed outcomes, actor attribution, source revisions, partition proof and limits are consolidated in ../../SUMMARY.md. No application change is part of later artifact cleanup.
