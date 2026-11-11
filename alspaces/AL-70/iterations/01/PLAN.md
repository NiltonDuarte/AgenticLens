# AL-70 iteration 01: Proposed implementation plan

Awaiting scope approval; no application changes or runtime verification performed.

Initial checkout was clean on `AL-70/honor-configuration-paths-and-ticket-identity` at `eb7d169c7f18fb935d4c2ceac998d4eca2ac36f4`. Integration with `origin/main` preceded planning and fast-forwarded to `f63e3db6b2d4baa8e22452e4ec94eff566867721`, clean afterward. This is the planning baseline. [Contract](../../CONTRACT.md) and [specification](../../../plan/to-do/AL-70%20honor%20configuration%20paths%20and%20ticket%20identity.md) own scope.

## Implementation and ownership

One coder owns application changes after approval, across these boundaries; the tester separately owns additional regression tests after coder handoff. They share the checkout and preserve others' edits. The planner owns only this plan/contract; the parent owns integration, delivery artifacts, telemetry and publication.

1. Correct `config/writer.go` scalar emission and missing settings; retain comments and shared/local separation. Trace CLI init output/local-pair placement and preserve AL-79 setup behavior. Extend writer/loader and CLI init coverage.
2. Retain configuration origin through `config/loader.go`, `selection.go`, `source.go`, `adapters/repository.go` and CLI scan/migrate/backfill/server composition. Resolve foreign workspace ownership before source injection. Reuse current resolvers rather than create a selection framework. Pass workspace policy to `opencode/schema_v1_session_resolution.go` and `schema_v2_session_resolution.go`.
3. Inject configured Codex home through production server wiring into `live.NewLiveServiceWithCodexHome`; retain mixed-owner hydration and source failure behavior. Current main already anchors local sources in `Load` and passes init's database flag to `InitialLocalYAML`; revalidate those paths rather than duplicate historical fixes. Server still calls `NewLiveService`, which derives Codex home from environment/home.
4. Unify prefix interpretation in domain `config.go`/`aggregations.go`, frontend `domain/formatters.ts`, configuration API and filtering/sorting consumers (`FleetOverview`, `TicketComparisons`, `MetricProgression`). Use the contract's longest-prefix and legacy-reader rules; preserve strict normalization/custom-pattern boundaries and existing raw-spelling numeric ties. Replace descending `!less` in `SortLedgers` with strict comparisons while preserving secondary keys.
5. Switch Git analyzer output/parser to NUL-delimited numstat; classify destination paths on rename. Correct this repository's `.agenticlens-config.yaml` test-support patterns without broadening generic defaults.
6. Bind `http.Server.ListenAndServe` to loopback and align CLI advertisement. Verify the actual listener and standalone embedded engine.

## Verification and regression boundaries

Tester inspects existing peers before extending coverage. Use real temporary config/filesystem, SQLite v1/v2, rollout and Git fixtures; never live ticket artifacts or developer stores. Cover YAML metacharacters/base branch, init then load with custom output, nested cwd, explicit relative/absolute overrides, base-local linked worktrees, physical symlink/parent traversal, foreign project roots and missing target configuration. Source fixtures must contain conflicting usable stores: strengthen `TestGenerateSelectedSourceDoesNotFallback` with a matching competing owner and retain selected-error/path/no-output assertions. Reach configured Codex hydration through server composition, preserving mixed owners and unavailable persisted history. Include copied valid ledgers with absent original sources where refresh/hydration changes.

Exercise fallback separators, SQL wildcards, title and exact-ID/nonce precedence; compare Go/TypeScript prefix cases, digit/delimiter prefixes, case and leading zeros through API/UI consumers. Repeat ascending/descending sorting with complete ties and verify cost/completion behavior. Real Git fixtures cover whitespace, tabs, non-ASCII, quotes, source-to-test renames and binary changes; prove `mockData.ts`, `setup.ts`, existing suffixes and adopting-policy preservation. Assert actual loopback listener plus standalone UI/API requests.

Relevant [regression boundaries](../../../../docs/REGRESSION_PREVENTION.md): configuration/project portability, extraction/refresh, persisted facts/optional hydration, portable ledgers, classification, frontend embedding and source availability. Run focused checks after changes, canonical frontend typecheck/lint/build, `make web-assets` before embed-dependent probes, and integrated `make` after main integration. Verify current ignored embedded assets; never force-add them. Record actual revision, commands, results and limits.

## Independent review

Freeze the candidate and run four focuses, each with `gpt-6.1-sol`/high and `gpt-6-luna`/medium, in capacity-limited waves per the [protocol](../../../../.agents/skills/agentic-lens-review/references/parallel-review.md). Unique report paths under this iteration: `review-<focus>-<profile>.md`.

| Focus / evidence owner | Risk question and scope |
| --- | --- |
| reviewer-behavior | Do CLI, API, UI and standalone server use the same effective paths/identity and loopback behavior? |
| reviewer-data | Do config precedence, fallback matching, Git paths and unavailable portable history preserve facts? |
| reviewer-architecture | Do origin handling and shared identity rules fit existing adapters/domain without duplicating selection or crossing manager boundaries? |
| reviewer-tests | Do production-wiring, conflicting-source, real-Git and repeated-sort tests detect the claimed regressions? |

The parent consolidates findings and retains attribution. No specialist is presently needed beyond these concrete scopes. Historical findings require current reproduction; this plan supplies no runtime approval. Explicit scope approval, implemented behavior, executed checks and complete review evidence remain outstanding.
