# docs(backfill): recommend exact-session refresh defaults (AL-65)

## Pull request

Recommend exact-session refresh for ordinary backfill **after unchanged runs can finish without writing**. Keep the current default until that condition and preservation checks pass. Today an operator can request this lookup with `--source <harness> --recompute-db` on a completed ledger. Clearer help can explain that invocation, but the ordinary command still skips it.

Availability means readable original records sufficient to reconstruct the selected owner and its required tree. Saved totals alone provide no evidence of source availability. Use the explicitly selected local store and stored harness, session ID and parent. Retain the prior tree and skill/task snapshots whenever tree or collection coverage is insufficient, preserve unrelated harnesses, and report mixed-age facts. OpenCode title or commit-range matching must remain opt-in via `--allow-fallbacks`. Codex has no title fallback. Report missing identities or ambiguity without guessing an owner. Keep Git measurement separately selected by ticket and ref.

In disposable probes on `b6ab0471`, an unchanged DB repeat rewrote both the ledger and requested summary. In a 12-ledger synthetic batch, the CLI reported 12 completed ledgers alongside four incomplete source refreshes. Both results matter for unattended backfill.

The roadmap includes two follow-ups. AL-94 is ready for scope review and covers unchanged writes, summary repair and retained-owner reporting. AL-95 depends on AL-94 and covers the opt-in preview, compatibility, bounded batch cost and a decision to switch, defer or retain the current default.

### Compatibility and rollout

1. Retain Force for callers that deliberately want to reprocess stored facts or regenerate a summary from stored facts. Keep explicit source recomputation working. Today either recompute flag already bypasses the completed-ledger skip.
2. In a separately approved implementation, introduce an opt-in preview of the proposed default and an explicit maintenance-only mode for existing batch scripts. Compare semantic facts before stamping or publishing, compare rendered summary bytes separately, and still repair a missing or stale summary when requested. A dry run should report expected owner coverage and writes.
3. Change the default only after validating bounded multi-ticket cost, unchanged bytes and mtimes, partial trees and unavailable selected stores. Announce the change before any switch. Keep explicit Git-only calls limited to their selected work and preserve explicit recompute flags and failure signals. Document a maintenance-only invocation before deprecating any Force spelling. The README already combines Git, DB, schema and summary flags; external script usage is unknown.

Proposed operator output, with migration status reported separately:

- “Codex: refreshed 5 of 5 recorded sessions.”
- “Codex: 5 sessions verified unchanged. No files written.”
- “Codex: 4 recorded sessions unavailable. Prior facts retained. Refresh incomplete.”
- “OpenCode: selected database unavailable. No files written. Refresh incomplete.”

Full `make` passed on integrated revision `bfb50c03` with main at `69b8c8d8`, including the CLI race suite, 231 frontend unit tests, 804 frontend integration tests and four-platform builds.

## Observed behavior on b6ab0471

At b6ab0471, a locally built engine ran in a temporary Git project with the shared configuration, private publication state, a controlled base/ref change and a fresh AL-55 ledger copy per case. Every invocation supplied `backfill --source codex --ticket-id AL-55 --workspace <copy> --update-summaries`. Git cases added `--code-ref HEAD`. Each ran twice. The partial input added a pending source-backfill migration step to the same ledger. The available store held five copied original rollouts from the [AL-63 audit](../al-63-verify-source-capabilities-across-harnesses/CAPABILITIES.md).

| Additional flags | Completed input, first run | Partial input, first run |
|---|---|---|
| None | Skipped ledger and summary | Processed stored facts, source step stayed pending |
| Force alone | Only ledger change was updated_at; rendered missing metrics into summary | Same outcome as ordinary partial processing |
| DB, with or without Force | Refreshed sessions | Refreshed sessions and cleared pending migration |
| Git, with or without Force | Replaced code facts from selected ref | Replaced code facts; source step stayed pending |
| DB and Git, with or without Force | Refreshed both | Refreshed both and cleared pending migration |

On completed input, ordinary runs took 28–42 ms, Force alone 80–116 ms, DB alone 311–339 ms and Git alone 90–101 ms. Every processed repeat advanced the ledger mtime and rewrote the requested summary even when its bytes were identical. On completed inputs the only changed ledger member on these successful repeats was updated_at. Partial migrations still pending also changed migration metadata. These are individual wall-clock observations from one machine with a small isolated store; they do not estimate a whole-repository scan.

A synthetic batch placed the same five owners in 12 distinct test ledgers under one configured root (256 sessions, 64 MiB per pass). With summaries disabled for the command, ordinary backfill skipped all 12 in 0.11 s. Two DB passes took 1.94 s and 2.11 s; each reported 12 completed ledgers and four incomplete refreshes. A dry run identified source budget exhaustion. Reusing five rollouts deliberately stresses repeated reads and does not model 12 independent real tickets. A default should name the affected tickets and permit a targeted retry against the same exact IDs.

### Recovery boundaries

- With only the root rollout accessible, the CLI reported incomplete recovery. The four unavailable child records survived field-for-field; the readable root's collection representation changed and four delegations remained. The completed input reported both one completed migration and one incomplete refresh. A pending migration remained partial.
- With an existing empty Codex home, all session values survived, incomplete recovery was reported and the ledger was rewritten. A nonexistent selected store failed before either file changed. Both situations require retaining the stored facts.
- On constructed OpenCode v1 and v2 selector ledgers augmented with five Codex owners, two DB refreshes against each single-family source backup reported complete and preserved all five unrelated records. Session arrays were stable on repeat. The mixed-family database reported incomplete and retained collections; an absent database failed before changing facts. With both harnesses selected, the engine opens the selected stores before processing ledgers, so an unavailable required store can prevent any refresh.
- AL-63 fixtures separately cover unreadable identified children, malformed records and prior snapshots. Positive native v2 skill/task facts, enabled heuristic lookup, large batch timing and external callers remain unverified. The isolated Git change demonstrates flag selection only. AL-55's tracked ledger matched HEAD after the probes.

## Integration with FIX-19

At merge revision `bfb50c03`, explicit Codex backfill rejects duplicate rollout paths within a selected family, including identical copies. The integrated regression checks two refreshes against identical, divergent compressed and warning-bearing copies. It requires an ambiguity diagnostic and incomplete coverage, preserves repeated skills and delegations, a missing owner and an unrelated OpenCode owner, and still refreshes a healthy Codex owner. AL-94 can use that distinction when reporting retained owners.

FIX-19 also charges repeated registered Codex discovery reads and shares the candidate allowance across refresh classes within one reconciliation and corrects cursor advancement for owners deferred by earlier reads. Explicit backfill uses its own source-read budget. The historical 12-ledger run measured that command on `b6ab0471`; its elapsed times and four incomplete refreshes have not been remeasured on the integrated base. AL-95 still needs representative multi-ticket preview evidence. [FIX-19 verification](../fix-19-reject-ambiguous-codex-snapshots/delivery.md)

## Verification

The ticket owner inspected CLI help, dispatch, current-ledger preparation and publication, ran 44 single-ticket Codex matrix invocations, supplementary root-only and batch controls and eight mixed-owner OpenCode invocations, and checked the original ledger against Git. Four existing backfill tests passed on the integrated revision, covering parse-failure tree preservation, skill/task snapshots, unavailable analysis and a readable parent with a missing child. Strict OpenSpec validation, three workflow tests and local Markdown link resolution passed. The prior audit supplies independent fixture and native-source evidence. This delivery changes documentation only.

The integration review focused on source selection, read budgets and roadmap closure because application, test and build inputs match main. Strict OpenSpec validation, three workflow tests and local links passed. The index preserves the AL-73, AL-80 and FIX-19 closures, AL-94 Ready and AL-95 blocked by AL-94. The full gate used serial Make targets, two Go package slots, a 30-minute Go package timeout and isolated temporary resources. The CLI race suite passed in 275 seconds.

On the earlier candidate 625837ae, integrated with origin/main b6ab0471, full `make` reported a capture-worker fixture cleanup race and exceeded the Go CLI package timeout of 10 minutes. All remaining backend packages passed. A 20-case CLI continuation reported two worker timing failures; those two passed together under `go test -race -count=1` on isolated retry. The separate frontend integration run passed 799 of 804 cases. Five cases exceeded 5 seconds; both affected files passed all 41 tests with one worker and a 30-second case timeout. All 231 frontend unit tests and the separate multi-platform `make build` passed.

The session ledger contains the one expected Codex root. No child agents were launched. The source reported an open turn at extraction; final duration and cost are unavailable.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 11
- **Active Agent Duration:** 45m 3s (recorded duration including 646s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 139.9k tokens (31% of 450k healthy ceiling, 0 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 1,721 reasoning tokens (5.6% of generated output)
- **Tool Health & Self-Correction:** 101 tool calls (9 self-corrections, 91.1% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 277,761 | 29,261 | 12,760,320 | 97.9% | 139.9k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **277,761** | **29,261** | **12,760,320** | **97.9%** | **139.9k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.2-astra** | 1 | 277,761 | 29,261 | 12,760,320 | 97.9% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **277,761** | **29,261** | **12,760,320** | **97.9%** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |
<!-- agentic-lens:metrics:end -->
