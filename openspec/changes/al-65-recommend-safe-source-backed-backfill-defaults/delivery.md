# docs(backfill): recommend exact-session refresh defaults (AL-65)

## Pull request

Recommend exact-session refresh for ordinary backfill **after unchanged runs can finish without writing**. Keep the current default until that condition and preservation checks pass. Today an operator can request this lookup with `--source <harness> --recompute-db` on a completed ledger. Clearer help can explain that invocation, but the ordinary command still skips it.

Availability means readable original records sufficient to reconstruct the selected owner and its required tree. Saved totals alone provide no evidence of source availability. Use the explicitly selected local store and stored harness, session ID and parent. Retain the prior tree and skill/task snapshots whenever tree or collection coverage is insufficient, preserve unrelated harnesses, and report mixed-age facts. Heuristics must remain explicitly enabled, with ambiguity reported. Keep Git measurement separately selected by ticket and ref.

### Compatibility and rollout

1. Retain Force for callers that deliberately want to reprocess stored facts or rebuild a missing summary. Keep explicit source recomputation working. Today either recompute flag already bypasses the completed-ledger skip.
2. In a separately approved implementation, introduce an opt-in preview of the proposed default and an explicit maintenance-only mode for existing batch scripts. Compare semantic facts before stamping or publishing, compare rendered summary bytes separately, and still repair a missing or stale summary when requested. A dry run should report expected owner coverage and writes.
3. Change the default only after validating bounded multi-ticket cost, unchanged bytes and mtimes, partial trees and unavailable selected stores. Announce the change before any switch. Keep explicit Git-only calls limited to their selected work and preserve explicit recompute flags and failure signals. Document a maintenance-only invocation before deprecating any Force spelling. The README already combines Git, DB, schema and summary flags; external script usage is unknown.

Proposed operator output, with migration status reported separately:

- “Codex: refreshed 5 of 5 recorded sessions.”
- “Codex: 5 sessions verified unchanged. No files written.”
- “Codex: 4 recorded sessions unavailable. Prior facts retained. Refresh incomplete.”
- “OpenCode: selected database unavailable. No files written. Refresh incomplete.”

## Observed behavior

At b6ab0471, a locally built engine ran in a temporary Git project with the shared configuration, private publication state, a controlled base/ref change and a fresh AL-55 ledger copy per case. Every invocation supplied `backfill --source codex --ticket-id AL-55 --workspace <copy> --update-summaries`. Git cases added `--code-ref HEAD`. Each ran twice. The partial input added a pending source-backfill migration step to the same ledger. The available store held five copied original rollouts from the [AL-63 audit](../al-63-verify-source-capabilities-across-harnesses/CAPABILITIES.md).

| Additional flags | Completed input, first run | Partial input, first run |
|---|---|---|
| None | Skipped ledger and summary | Processed stored facts, source step stayed pending |
| Force alone | Only ledger change was updated_at; rendered missing metrics into summary | Same outcome as ordinary partial processing |
| DB, with or without Force | Refreshed sessions | Refreshed sessions and cleared pending migration |
| Git, with or without Force | Replaced code facts from selected ref | Replaced code facts; source step stayed pending |
| DB and Git, with or without Force | Refreshed both | Refreshed both and cleared pending migration |

On completed input, ordinary runs took 28–42 ms, Force alone 80–116 ms, DB alone 311–339 ms and Git alone 90–101 ms. Every processed repeat advanced the ledger mtime and rewrote the requested summary even when its bytes were identical. On completed inputs the only changed ledger member on these successful repeats was updated_at. Partial migrations still pending also changed migration metadata. These are individual wall-clock observations from one machine with a small isolated store; they do not estimate a whole-repository scan.

### Recovery boundaries

- With only the root rollout accessible, the CLI reported incomplete recovery. The four unavailable child records survived field-for-field; the readable root's collection representation changed and four delegations remained. The completed input reported both one completed migration and one incomplete refresh. A pending migration remained partial.
- With an existing empty Codex home, all session values survived, incomplete recovery was reported and the ledger was rewritten. A nonexistent selected store failed before either file changed. Both situations require retaining the stored facts.
- On constructed OpenCode v1 and v2 selector ledgers augmented with five Codex owners, two DB refreshes against each single-family source backup reported complete and preserved all five unrelated records. Session arrays were stable on repeat. The mixed-family database reported incomplete and retained collections; an absent database failed before changing facts. With both harnesses selected, the engine opens the selected stores before processing ledgers, so an unavailable required store can prevent any refresh.
- AL-63 fixtures separately cover unreadable identified children, malformed records and prior snapshots. Positive native v2 skill/task facts, enabled heuristic lookup, large batch timing and external callers remain unverified. The isolated Git change demonstrates flag selection only. AL-55's tracked ledger matched HEAD after the probes.

## Verification

The ticket owner inspected CLI help, dispatch, current-ledger preparation and publication, ran 44 Codex flag/availability invocations and eight mixed-owner OpenCode invocations, and checked the original ledger against Git. The prior audit supplies independent fixture and native-source evidence. This delivery changes documentation only.
