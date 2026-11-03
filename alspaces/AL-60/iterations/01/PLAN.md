# AL-60 Iteration 01

## Authorization and baseline

User instruction: “its pre-approved. full send it”. Scope is frozen in `../../CONTRACT.md`; no renewed approval is needed for this slice. Baseline: `122e983f8c365a2aa248b1b6d492252555e922ae`; initial Git status clean. Parent owns integration, workspace summary, lifecycle updates, telemetry, commits and PR publication. All participants share the checkout and preserve unrelated changes.

## Grounded implementation

1. Coder owns `web/src/components/TicketScatterPlot.tsx` and, only if necessary, scoped styles in `web/src/index.css`. Add local search query state and derive visible checkbox rows by case-insensitive ID/title matching. Use an associated visible label, native editable search input, and explicit usable clearing behavior. Keep search outside the scrollable choices area so empty results retain recovery controls. Add no-match copy within the panel. Preserve the existing independent exclusion set and feed original rows into scatter derivation.
2. Tester owns `web/src/components/TicketScatterPlot.test.tsx`. Extend its existing fixture-based interaction coverage rather than introducing a parallel fixture suite. Existing tests own chart defaults, metric/scale changes, point labels, skills tooltips and click navigation. Add meaningful cases for ID/title case-insensitive query changes, unmatched results and recovery, unchanged chart membership/counts during typing and clearing, hidden excluded state, visible choice toggling and restoration. Use user-event for keyboard input and keyboard choice activation where needed. Assert point identity as well as count to catch replacement errors. Do not duplicate pure derivation coverage in `web/src/domain/ticket_scatter.test.ts`.
3. Verify title availability through existing `TicketComparisonRow`, `fetchWorkspaces` and `deriveTicketComparisonRows` paths by inspection; no API or extraction additions are planned. Check `TicketComparisons.tsx` continues passing its ordinary filteredRows to the component. No backend tests need additions for unchanged contracts.

## Review roster and ownership

Freeze product/test edits and provide every reviewer the same baseline, candidate diff/content fingerprint, contract, plan and raw verification evidence. Parent alone owns consolidated review artifacts. Read the current model policy before dispatch: every selected persona expands to gpt-6.1-sol/high and gpt-6-luna/medium; record actual execution identity or capability limits. Run in capacity-bound waves with fresh context and independent first passes.

- Behavior and integration: `reviewer-behavior`; verify accessible search/recovery, selection persistence, chart independence and keyboard/point interactions across TicketComparisons, TicketScatterPlot and unchanged derivation. Own `REVIEW_BEHAVIOR_SOL.md` and `REVIEW_BEHAVIOR_LUNA.md` respectively.
- Data integrity and failure handling: `reviewer-data`; verify query cannot mutate exclusions or chart inputs and title/ID paths retain missing-data contracts. Own `REVIEW_DATA_SOL.md` and `REVIEW_DATA_LUNA.md` respectively.
- Architecture and maintainability: `reviewer-architecture`; verify local presentation responsibility, independent state, no redundant data contracts or unnecessary extraction, and production asset synchronization. Own `REVIEW_ARCHITECTURE_SOL.md` and `REVIEW_ARCHITECTURE_LUNA.md` respectively.
- Test effectiveness: `reviewer-tests`; inspect actual interaction oracles against hidden exclusions, no-match recovery, point identities and query-only updates; avoid mirrored implementation assertions. Own `REVIEW_TESTS_SOL.md` and `REVIEW_TESTS_LUNA.md` respectively.

No extra specialist is needed: native accessibility and embedded asset provenance are explicitly assigned above; this slice adds no asynchronous operation, trust boundary, metric or polling workload. Reports live beside this plan. Parent deduplicates findings, routes corrections to assigned coder/tester and obtains affected revision-aware reviews before consolidated approval. Retain compact REVIEW_SUMMARY.json after approval and remove only resolved detailed reports.

## Verification and delivery

Applicable regression checklist boundaries are UI recovery after empty filters, baseline persisted title availability, independent data/selection state, accessible keyboard interactions, test proportionality and embedded assets. Run focused component tests during implementation. Inspect existing metric derivation and component tests before adding coverage. Record tested diff/commit, commands/results and limitations in SUMMARY.md; reviewers may reuse unchanged valid test evidence.

Commit task-owned implementation before merging origin/main into the feature branch; retain the ticket branch and never check out main. Run required full `make` after integration, including installed frontend dependencies, typecheck, lint, frontend tests, Go race tests and production build. Canonical `npm run build` invokes `web/scripts/sync-assets.mjs` to copy web/dist into backend/internal/adapters/web/dist; the latter tree is currently untracked build output, so inspect it without force-adding ignored files. Verify standalone binary contains/serves the new search assets with no Node runtime dependency. State any unexercised browser geometry explicitly.

Finalize concise SUMMARY.md and PULL_REQUEST.md, move the matching specification to plan/done, remove only AL-60 from TICKETS.md and inspect downstream dependencies. Generate source codex telemetry with a literal random nonce only at the final step before workspace commit; verify root, expected children, source labels and preserved existing facts. Commit telemetry and workspace deliverables, push, open PR using title from PULL_REQUEST.md and immediately register its URL. Do not merge the PR.

## Delivery steering and review execution

User steering during iteration 01: MGT-2 supervises delivery and owns merging eligible non-release PRs; AL-60 is first in the proposed integration order. Scope and acceptance remain unchanged; parent publishes the final reviewed head/evidence and does not merge. No new iteration required.

Reviews ran in fresh T3 delegated contexts in capacity-bound waves, with both model/effort profiles acknowledged by the dispatcher. Actual private report ownership superseded the provisional filenames above: `/tmp/al60-review/<focus>-<model>.md`, with parent-owned run/finding/consolidation records beside them. Product/test revision stayed frozen at candidate 7a50a5e; source diff fingerprint 8703b8f22f375420ebd76f8e7a62f22e2a8b6902563133ea837fac91c4294914. The branch integration found origin/main already contained, so no code re-review was needed.

Additional user evidence/steering: MGT-2's unchanged baseline make also failed capture native-version/capability and worker timeout tests. Preserve that reported baseline evidence separately from AL-60's observed runs; do not weaken tests or broaden scope. AL-60 has priority for bounded full verification. On final handoff or concrete blocker, send PR URL, reviewed head and verification state directly to MGT-2 thread cce1ea90-2513-466c-a078-58b2765dfbb6 using t3_thread_send. Parent remains responsible for publication; manager owns merging.

User subsequently requested read-only diagnosis of shared capture failures and explicitly withheld approval for capture code/test repair. Preserve reviewed search code; do not repeatedly run the full suite. Parent assigned a read-only explorer to trace native probes, installation identity and worker deadlines; no capture changes authorized. AL-61's reported failures on unchanged 122e983 and MGT-2 baseline evidence are corroboration, distinct from this checkout's observed results. Merge gates remain incomplete pending the separate scope decision.
