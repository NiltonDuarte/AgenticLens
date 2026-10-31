# AL-55 developer contract proposal

Status: Grounded proposal only; not frozen or approved for implementation. Pending PO-17 scope reconciliation and explicit user approval conveyed directly or by the manager. A Ready label is not approval.

Baseline: `bdef2d0ef98e37fe259921bc30ba2daefbfda717`, branch `feat/AL-55-migration-backfill-scope`. Initial working tree was clean. Configured workspace root is `alspaces`. Only AL-55 developer planning artifacts are writable in this phase. Shared specifications and ticket index remain owned by PO-17 and the manager.

## Problem remaining after delivered dependencies

Historical skill/task recovery already exists but its safe refresh and completeness semantics have gaps. The remaining useful work is a narrowly bounded correction to source-backed backfill, not a second schema-v4 introduction. This finding is based on source inspection, not a reproduced failure or completed test run.

| Existing behavior | Evidence and implication |
| --- | --- |
| Go schema v4 and the complete registered migration chain already exist | `backend/internal/domain/migration/engine.go`, `step_v3_v4.go`, `engine_test.go`: `StepV3ToV4UserIterations` is registered. Preserve it. |
| Shape migration explicitly leaves source recovery pending | `migration.Engine.applyChain` adds `source_backfill` and partial status. `adapters.ValidateRuntimeFactLedger` rejects incomplete migration. Shape validity does not prove recovered history. |
| Skill/task collections deliberately remain optional | `backend/internal/domain/models.go` and `backend/internal/adapters/schemas.go` use `omitempty`; `web/src/domain/types.ts` declares optional arrays and read helpers. Empty read projections are convenience defaults, not extraction evidence. |
| OpenCode extraction already produces normalized skill/task facts | `backend/internal/adapters/opencode/skills_tasks.go`: source events, skill status and input names, task metadata and direct-child verification. Both schema-family extraction test files exercise this producer. |
| Backfill already populates these facts | `backend/internal/adapters/backfill/engine.go`, `processSingleWorkspace`, calls `ExtractSkillTaskFacts` in the replacement branch. It is false that AL-55 must introduce extraction from nothing. |
| Migration runner already supports dry-run, backups and atomic ledger replacement | `backend/internal/adapters/backfill/migration_runner.go`, `migration_runner_test.go`. Backfill itself uses temporary-file rename and dry-run, but exposes no backup option. Do not conflate the two commands. |
| Client migration module does not exist | There is no `web/src/domain/migration.ts`; normal ingestion is server-controlled. Do not introduce client legacy ingestion. |

## Observed remaining gaps

1. Canonical v4 with no migration metadata can hit the completed fast path even when `RecomputeDB` is explicitly requested. The skip guard checks `RecomputeGit` but not `RecomputeDB`. The similarly named existing recompute test supplies completed migration metadata and therefore misses the ordinary canonical case.
2. In partial OpenCode resolution, backfill assigns the entire ledger-level skills/task arrays from only the refreshed sessions, then preserves unresolved sessions separately. Old facts belonging to unavailable roots can be lost. These fact types have no session/source identifier, so individual old rows cannot safely be attributed using timestamp, role, turn or task child ID.
3. A single `dbSessionFound` boolean represents aggregate recovery. Any successful Codex tree sets it true, and either harness can therefore clear the global pending `source_backfill` step despite other selected stored roots or children being unavailable. Root-count comparison in OpenCode also does not independently establish stored-child coverage.
4. Existing tests cover migration/source separation, unavailable sources, multi-root extraction, source filtering and Codex tree replacement. No skill/task assertions occur in `backfill/engine_test.go`; recovery multiplicity and preservation of these arrays through incomplete refresh remain unguarded there.

## Minimal proposed acceptance boundaries

Subject to decisions below, retain schema v4 and optional arrays. AL-55 owns backfill command selection, safe source refresh, preservation of historical skills/task facts, and truthful recovery outcome. It does not own cost availability, rate cards, UI simulation, a new schema family, or AL-43 diagnostics/content fields.

An explicit source recompute must reach the requested source for an otherwise complete canonical ledger without requiring `--force`. Routine invocation without recompute retains existing skip semantics. Existing force and heuristic discovery are separate controls; exact stored session identity remains the default, and heuristic substitution remains opt-in.

Resolve and parse each selected source tree before replacement. Successful refresh replaces cumulative snapshots for that owner rather than adding totals. Preserve unavailable sessions, all non-selected harness sessions, legitimate repeated invocation rows, parent links and existing persisted facts when discovery, parsing or grouping fails. Confirm stored-child coverage as well as roots; source absence is incomplete, not a successful empty extraction.

Recommended conservative array policy: the current producer is explicitly OpenCode-owned, but historical ledger-level rows have no independently verifiable owner. Replace the whole skill/task arrays only after complete exact refresh of all stored OpenCode roots and children required to reconstruct that producer's history. On incomplete OpenCode recovery preserve both old arrays unchanged and report incomplete recovery; do not append an unattributable fresh subset or guess ownership. Codex-only refresh never manufactures or replaces skill/task facts. A successful complete OpenCode recovery with no observed invocations may produce an empty optional collection; absence alone never establishes that condition. This policy avoids a speculative persisted provenance extension; partial fresh skill/task enrichment is intentionally deferred until full history is available.

Under that recommendation, available session snapshots may refresh while skill/task arrays remain the prior snapshot. Diagnostics must explicitly say those arrays were retained because recovery was incomplete; they must not claim the arrays reflect the refreshed sessions. If that temporary age difference is unacceptable, the alternative is to defer the whole OpenCode refresh until complete coverage; this must be chosen in decision 2. Stage extraction results before modifying the persisted aggregate: a late parse/grouping failure cannot publish replacement arrays based on the earlier successful subset. Completeness means successful exact discovery and parsing of every required stored root and descendant, plus grouping of the complete discovered tree; a nonempty returned slice or root count is insufficient. A source-confirmed successful zero-observation result is distinct from failed discovery or unavailable source.

Track coverage per selected harness/tree during execution. Clear a pre-existing global source-backfill requirement only when all stored owners required by that requirement are demonstrably covered; a success from one harness must not discharge missing owners in another. Report missing source/tree and parse/grouping failures with useful context. Preserve pending steps unrelated to the successful owner. Do not impose permanent migration metadata on already-complete canonical data merely because optional skill/task arrays are absent.

Persist only after final domain preparation and boundary validation succeed. Keep dry-run free of file changes. Repeated unchanged refresh preserves the semantic fact set and row multiplicity; administrative timestamps may advance under current behavior. No read hydration or GET persistence changes are proposed. Preserve AL-43 additions during eventual integration.

## Decisions required before freezing

1. Approve preserving the AL-51 optional-array contract, superseding AL-55's stale required-array acceptance paragraphs. If required arrays are still desired, PO must specify a new compatible contract and version strategy first; no silent tightening under existing v4.
2. Approve conservative whole-set skill/task replacement after complete OpenCode coverage, with unchanged arrays on partial recovery and explicit retained-snapshot diagnostics. Choose whether successful session snapshots may refresh while those arrays remain old (recommended), or the entire OpenCode refresh waits for complete coverage. Alternative per-session replacement requires an approved provenance model and legacy-row policy; it is a larger, currently unjustified schema change.
3. Approve global migration completion semantics across mixed sources: recommended behavior keeps global source-backfill pending while any stored required owner remains unresolved, including owners outside a narrowly selected source. Selection can update one source without asserting entire-ledger recovery. Define whether failed refresh of an already-complete ledger records only a command diagnostic or makes runtime ingestion ineligible; recommendation is preserve baseline eligibility unless an existing migration requirement remains unmet.
4. Confirm explicit recompute refreshes current snapshots without `--force`, including skill/task arrays after complete coverage. Current code otherwise retains invocation snapshots in some completed-ledger paths; acceptance must not promise refresh while only adding events.
5. Confirm backup scope: retain existing standalone `migrate --backup` and atomic backfill writes for this minimal slice, or explicitly require an additional backfill backup interface. The latter is not already delivered and is not assumed authorized.

## Architecture and exclusions

Pure coverage evaluation and source-independent replacement policy belong in Go domain only if useful beyond local adapter bookkeeping; discovery, source errors, CLI wiring, validation and writes stay in adapters/CLI. Do not move OpenCode payload parsing into domain. Zero `any`, zero emojis, typed states, no domain I/O or clocks. No frontend implementation is justified by the proposed slice. Preserve nullable-cost and rate-card allocation for PO-17/manager and held AL-57. No shared plan lifecycle changes, ticket completion claim, coder dispatch, release action, PR merge or auto-merge is authorized by this proposal.

Verification plan and existing coverage owners are in `iterations/01/PLAN.md`. No application behavior or full gate has been verified by this planning artifact.
