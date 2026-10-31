# AL-55: Preserve historical facts and report truthful backfill recovery

- **Subsystem**: Domain / Adapters / Backfill / CLI / Tests
- **Phase**: Phase 8 (v2 Observability / Skills)
- **Blocked by**: None
- **Blocks**: None
- **Approval**: Explicit user approval to continue AL-55 implementation, followed by direct confirmation of the boundary/read-normalization, completeness, recompute and backup choices. See PO-17 iteration 02 for the decision record. Developer owns execution status in the shared index.

## User outcome

An operator can explicitly refresh an existing ledger without silently skipping it or losing history that is unavailable on this machine. The command reports what refreshed, what was retained, and what recovery remains incomplete. Valid current ledgers continue to load normally; no new format or mandatory migration/backfill rollout is introduced.

## Context and superseded scope

AL-56 already delivered schema v4, the registered Go v3-to-v4 step and the explicit distinction between shape migration and source-backed recovery. AL-51 delivered optional skills/task metadata and its producers; AL-52/53/54 delivered consuming views. The former proposal to introduce v4 again, recreate its migration, require persisted arrays, or add client-side legacy ingestion is superseded. Those requirements are not active acceptance criteria.

Current backfill can replace ledger-level skill/task collections from only a refreshed subset, skip canonical complete ledgers despite explicit source recompute, or treat one source-tree success as global recovery. This ticket corrects those behaviors within the existing format and command workflow. It does not depend on deferred AL-57 cost decisions or queued AL2 work.

## Approved scope

### Data boundary and internal reads

Keep `skills` and `task_links` optional at the persisted schema-v4 boundary. Internally, both collections are required arrays: reading a ledger that omits either collection creates an empty array for internal use. An empty read projection is not evidence of a source-confirmed empty history or completed recovery. Do not require rewriting existing files solely to materialize arrays, add a schema version, or impose permanent migration metadata on valid ledgers. Preserve the existing supported-v4 ingestion and explicit legacy-upgrade boundaries.

### Safe refresh and retained history

Refresh independently complete session trees selected by the operator, preserving exact source/session identity, existing unavailable roots and descendants, other-harness facts, parent linkage, and legitimate repeated invocation rows. Discovery, parsing, grouping or validation failure must not publish a partial replacement for an affected owner. Refresh cumulative snapshots rather than adding them to their old totals.

On incomplete OpenCode recovery, retain the prior whole skill/task collections. Their historical rows lack reliable ownership for safe subset replacement or guessed merging. Replace the whole collections only when all necessary exact OpenCode owners and descendants have recovered sufficiently to reconstruct that producer's history. A nonempty root result or a root count alone does not establish child coverage. Ownerless collections with no reconstructible owner history are unresolved, not proven empty. Codex-only refresh does not manufacture or replace these collections.

The user accepts that successfully refreshed session facts may be newer than retained skill/task collections. CLI per-ticket results and command summary must explicitly describe that mixed-age condition, the retained collections, and incomplete recovery. A new persisted provenance model or frontend freshness feature is not required.

### Completion and optional failure

Clear an existing pending source-recovery requirement only after every necessary stored owner and descendant for that requirement has recovered across the required sources. A successful Codex tree cannot discharge unavailable OpenCode owners or other missing trees. Source selection authorizes that source's refresh; it does not redefine the global outstanding obligation. Preserve unrelated pending steps and report unavailable owners and parse/grouping failures truthfully.

A failed optional refresh of an already-valid canonical ledger preserves baseline eligibility and facts, with an explicit command diagnostic. It must not introduce a new mandatory migration just because enrichment failed or optional collections were absent. Ledgers with an existing unmet recovery requirement remain incomplete and subject to the existing runtime boundary. Completion is bounded by known required owners and discovered source evidence, not a claim of exhaustive recovery from unavailable external history.

### Explicit recompute and persistence

Explicit source recompute works without `--force`, including a current v4 ledger without migration metadata. Successful eligible recompute refreshes the requested snapshots, rather than only bypassing the skip or refreshing events. Routine runs retain their skip behavior. Exact stored identity remains the default; heuristic fallback stays separately opt-in. Force does not permit erasing unresolved facts or guessing ownership. Align CLI help with the accepted behavior.

Retain existing standalone migration backup support, dry-run guarantees and atomic backfill writes. No new automatic backfill backup interface is required. Atomic replacement is not a rollback copy. Preserve Git facts unless explicitly requested for recomputation, with existing ticket/ref guardrails. Persist only validated final data; dry-run changes no ledger or backup files.

## Acceptance criteria

1. A schema-v4 file may omit skills/tasks and still load; its internal read result supplies both arrays as empty. Existing populated collections remain intact. Read normalization alone neither rewrites the ledger nor claims recovered history.
2. Explicit recompute of canonical v4 with no migration metadata reaches and refreshes the requested eligible source without Force; a routine run still skips when appropriate. Repeating unchanged input preserves semantic facts and row multiplicity.
3. With two OpenCode roots where one is unavailable, or an available root with a missing stored child, unavailable historical session facts and the whole prior skill/task collections survive. Independently complete trees may refresh. CLI results explain retained older collections and incomplete coverage.
4. Complete exact OpenCode recovery replaces the whole producer collections. A source-confirmed successful empty observation differs from missing sources, empty discovery or extraction failure; the latter never authorizes clearing historical collections.
5. With mixed sources or multiple Codex trees, one success plus another required-owner failure leaves existing global source recovery pending. Narrow source selection preserves unrelated requirements.
6. Failed optional refresh preserves valid baseline readability; failed required recovery preserves pending diagnostics and existing ingestion restrictions. Missing optional arrays alone do not create a recovery obligation.
7. Discovery, parse, grouping, validation and persistence errors do not publish an invalid or incomplete replacement. Non-selected harness facts, parent links, Git facts and legitimate repeated invocations remain preserved. Dry-run and existing migration-backup behavior retain their guarantees.
8. Tests use self-contained synthetic fixtures through actual read, source, backfill and CLI boundaries. Cover canonical migration-metadata absence, absent arrays, skill/task retention, missing children and mixed-source completeness using coherent existing test recipients. Full `make` passes with review evidence and environment limits recorded.

## Definition of done and exclusions

A separate AL-55 developer session implements and reviews this approved behavior, verifies the affected boundaries and full gate, and delivers its own PR/telemetry. PO-17 only reconciles this specification. Developer may take lifecycle ownership of this file only after PO-17's documentation PR merges and manager hands off.

No cost semantic changes, nullable-cost rollout, rate cards, AL2 development, new source/session provenance schema, client migration, or mandatory migration/backfill campaign is included. Existing observed financial values are preserved. No release or merge action is authorized by this specification.
