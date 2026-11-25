# FIX-21 test-effectiveness re-review: tests-sonnet-r3

- Run: FIX-21 re_review round 3, run ID tests-sonnet-r3, focus reviewer-tests, requested profile Claude Sonnet 5.5 / high (runtime model self-report: claude-sonnet-5-5; effort not independently verifiable).
- Reviewed revision: 5d4df7ee (frozen head). Corrections delta `86eb6dae..5d4df7ee`, base 133100bb. Working tree clean at start and at end.
- Independence: no other `*-r3.md` report was read. The brief, persona, shared context, REVIEW.md (rounds 1 and 2), my round-2 report, TESTS.md and the delta were read.
- Write scope: this report only. Every probe and mutation ran in scratch copies of `backend/` under `.../scratchpad/tests-sonnet-r3/` (checkout, `~/.codex`, `~/.agenticlens` and live ledgers untouched).

## Verdict: APPROVED for the test-effectiveness scope, with two low-severity gaps for the parent to close or accept

All of R2-T1 to R2-T8 are closed by tests that fail when their fix is removed. I found no product defect in the rewritten position-alignment rule: a property probe over 20,000 random append-only transcripts agrees with the rule. The two new gaps (T3-1, T3-2) are surviving mutants, severity 3 and 2. Neither hides a current defect. I recommend T3-1 (two table rows) before delivery and a conscious acceptance of T3-2.

## Round-2 test findings: verification at 5d4df7ee

| ID | Status | Evidence (mutations are mine, in scratch copies, unless noted) |
|---|---|---|
| R2-T1 contention test flake | Closed | The retry now has its own 2 s context (`retryCtx`). `-race -count=20` passed 20 of 20. A harder run passed 36 of 36: six parallel processes, a prebuilt `-race` binary, `-test.cpu=1`, on a host with load average about 11. Together with `TestRetainAcceptedChildLinksReturnsDiagnosticWriteFailure`. The first, contended call still relies on SQLite returning BUSY before its 2 s context expires, which held in every run. |
| R2-T2 group position vs absolute index | Closed | Mutation "absolute accepted index selects the fresh delegation" fails both interleaved subtests (`a group interleaved with another key and offset between sides`, `two interleaved groups each keep their own order`). |
| R2-T3 contradicted member of a refused group | Closed | Mutation "contradicted filter removed in `refusedChildLinks`" (M24) fails `a contradicted member of a refused group is not reported`. |
| R2-T4 `RecordDiagnostic` failure | Closed | `RecordDiagnostic` error discarded (M28, valid form with `_ =`) fails `TestRetainAcceptedChildLinksReturnsDiagnosticWriteFailure`. The test uses a registry trigger on the diagnostic insert, so it reaches the real write. |
| R2-T5 header conflicting ancestry | Closed | `header.ConflictingParent = false` (M29) fails the `direct and nested parents disagree` subtest. My round-2 mutant G1 is now killed. |
| R2-T6 queued cursor at the ceiling | Closed | Skipping the restore (`_ = previous`, Q1/M30) fails `TestQueuedChildPathsResumeAtTheRefusedPath`. The test asserts the cursor value and that the next one-read pass reads the refused path, not the one after it. |
| R2-T7 import guard needs `web/dist` | Closed | On a scratch copy without `web/dist` the guard passes. With a production import of `codexstoretest` added to the capture package, it still fails and lists the dependent production packages (cmd, adapters, backfill, capture, http). |
| R2-T8 e2e alignment, `delete(refused)`, refusal decision | Closed | (a) CLI `TestCaptureAutoAlignsAPartlyResolvedGroupWithTheCommittedLedger`: reinstating whole-group refusal (M20-style, in the domain, run through the CLI) fails it. (b) Removing `for link := range kept { delete(refused, link) }` (M31) fails `...ReportsALinkKeptAndRefusedAsKept`; my round-2 S2 is now killed. (c) `refused_ambiguous` and `ChildParentNotEvaluated` no longer exist (`rg` finds none in backend), so the missing decision assertion is moot. |

R2-T4/N1 from my round 2 also covers M10c, which is now killed. Round-2 N2 (decision assertion) is removed with the dead decision. Round-2 N3, N4, N5 map to R2-T7, R2-T5, R2-T6 above.

## The rewritten domain rule: oracles and mutation evidence

I read the replaced oracles against the "Position alignment" and "R2 implementation sync" steering. Two refusal-table cases changed by design (`an unconfirmed fresh link in a group keeps the accepted link at its position`, `a partly re-read group keeps the unread position`); both now assert the exact links, an empty `Refused`, and the unresolved list. The smaller-group and unmatched cases keep independent rows. No assertion was weakened without cause. The facts-only second-application check in the refusal table is still justified: a rule output fed back in as fresh has all links set, so nothing is retained and `UnresolvedLinks` is legitimately empty. The alignment test adds the stronger `f(f(x))` and "output fed back as accepted" checks.

Domain mutations against `go test ./internal/domain/` (all in a scratch copy):

| Mutation | Result |
|---|---|
| M20: whole-group refusal when any fresh link is set | killed (16 failures) |
| Absolute index in place of group position | killed (both interleaved subtests) |
| M22: untimed unique delegation accepts an unconfirmed child | killed (6) |
| M23: unmatched key not reported | killed |
| M25: complete positions decided (equal link not skipped) | killed (6) |
| M24: contradicted filter removed in refusals | killed |
| `grouped` always false (outcome `retained` instead of `retained_by_position`) | killed (9) |
| `grouped := key.timed` (inverted) | killed (11) |
| `<` to `<=` on the timed and untimed smaller-group cases | killed (38 and 42) |
| Timed and untimed refusal reasons swapped | killed (5) |
| `!linked[...]` check removed in `refusedChildLinks` | killed (4) |
| Held-by-sibling case disabled | killed (5) |
| Held-count guard (`held[match.accepted] > 0`) removed | killed (5) |
| Outcome switch order swapped (grouped before pending confirmation) | killed |
| `replaced` branch disabled | killed (10) |
| Key without description, or without agent type (presence and value) | killed (2 each) |
| `Known` evidence precedence swapped with `Accepted` | killed |
| Accepted record of another harness ignored (`prior.Source`) | killed |
| Agent-type value or invoked-at value dropped from the key | SURVIVED (T3-1) |
| Empty-string `InvokedAt` treated as a known time | SURVIVED (T3-1, minor) |

Property probe (scratch only, not committed): 20,000 random append-only transcripts (1 to 7 delegations over four keys, including untimed and untyped), an accepted prefix with about one in five links empty, and a fresh read that sets a random third of links. For every position the result equals the true child when the fresh read or the accepted prefix has it and is empty otherwise; the rule is idempotent on facts; every carried link is reported unresolved and no re-read link is; `Refused` stays empty. This supports the tester's alignment table through an independent oracle and found no invented or misplaced link.

## Capture, CLI and codex mutations

| Mutation | Result |
|---|---|
| M31, M28, M30 (queued cursor), M29 (header conflict) | killed, see table above |
| M26: stderr claims partial always | killed by `TestCaptureAutoReportsOnlyClearedLinksWithoutPartial` |
| M26b: partial claimed when `Added == 0` | killed |
| Early-return condition ignores `Cleared` | killed |
| M20 through the CLI (`TestCaptureAutoAligns...`) | killed |
| M26c: "remains partial" never printed | SURVIVED (T3-2) |
| Contention/deadline exit mapping removed (both forms) | SURVIVED (T3-2) |

## Findings

### T3-1. The key's agent-type value and invoked-at value are unprotected (severity 3, introduced; the key code was rewritten in this delta)
- Location: `backend/internal/domain/child_links.go`, `delegationLinkKey`; tests `child_links_test.go`.
- Expectation: the spec groups delegations by agent type, description and invocation time. A delegation whose agent type or recorded time differs is another delegation, so its accepted link has no fresh counterpart and is reported as unmatched, not carried.
- Trigger: accepted `tester/d@t1 -> C1`; fresh `worker/d@t1` (or `tester/d@t2`) with no link.
- Consequence: a mutation that drops either value from the key (keeping only the `typed` and `timed` flags) passes the domain package, the capture retention and reconcile tests, and the CLI tests. The mutation would carry `C1` onto a different delegation, so a link could be placed on a delegation that changed its agent type or time. The current code is correct: my probe asserts `[""]` for both and passes unmodified. The existing case `a different description or agent type is another delegation` changes only the description and the agent-type presence.
- Evidence: scratch mutants K-agent-value and K-invoked-value, each run against domain, capture and cmd focused tests: all survived. With both mutations and the two probe cases added, the probe fails, so two table rows would kill them.
- Suggested evidence: two rows in `TestRetainAcceptedChildLinksReportsRefusedLinks` (agent type value differs; known invocation time differs), both expecting an empty link, `unmatched` and the child listed as unresolved. Also one row with a pointer to an empty string for `InvokedAt`, which the key treats as unknown.

### T3-2. The CLI exit-code classification and the positive "remains partial" wording have no assertion (severity 2, introduced; the exit mapping is a recorded limit)
- Location: `backend/cmd/agentic-lens/capture.go`, `retainCaptureChildLinks`.
- Expectation: contention or a deadline on the 2 s retention context ends generation with the busy exit (BEH-02, R2-06). The line says "remains partial" exactly when `Added > 0` (R2-05).
- Consequence: removing the whole `IsRegistryContention || DeadlineExceeded` branch, or only the `DeadlineExceeded` clause, passes every cmd test, so a contended registry would exit 1 and the worker would stop backing off. Dropping the "remains partial" text altogether (`if false`) also passes: only the negative case (cleared-only) is asserted, so no test reads the positive claim.
- Evidence: scratch mutants "contention branch removed", "R2-06 clause removed" and M26c all survived `TestCaptureAuto*`, `TestManualGenerate*`, `TestCaptureWorkerIsolates*`. TESTS.md records the exit mapping as verified by reading only; the product has no seam between reconciliation and retention.
- Suggested evidence: the exit-versus-error classification is a two-term boolean, so extracting it as a pure function and testing it in the `main` package is a small seam that needs no live contention. For the wording, one assertion in the existing alignment case that the stderr line contains ", so the reconciliation remains partial." when child 3 is kept.

### T3-3. Stale test comments after the alignment rewrite (severity 1)
- Location: `domain/child_links_test.go`, lines 535 to 537 ("an ordinal group the rule filled is then partly linked, so it is refused as a whole"), 448 to 449 and 476 to 478 ("within an ordinal group"), and the header of `TestRetainAcceptedChildLinksOrdinalGroups` ("when every fresh link of the group is empty and the fresh group is at least as large").
- Consequence: no behavior. Whole-group refusal on a partly linked group no longer exists, and the carry no longer requires an all-empty group, so a reader will take the wrong rule from these comments. The facts-only second-application check is still correct, for the reason given above.
- Suggested evidence: reword the comments; no assertion change.

## Suggestions (severity 1)
- `TestCaptureAutoAlignsAPartlyResolvedGroupWithTheCommittedLedger` and the domain `fedBack` check pass `tc.accepted[1:]` after the output; that works but reads oddly. No action needed.
- Design.md still carries the superseded ordinal Decision paragraph (line 218) beside the later Steering entries. It is history and is clearly superseded in lines 235 and 245.

## Flakiness and hygiene
- No `t.Skip`, `testing.Short`, domain mocks, sleeps or wall-clock assertions in the added tests; all fixtures are under `t.TempDir()`.
- Under `-race`: domain, codex and codexstoretest pass (one run); `TestRetainAcceptedChildLinks*`, `TestQueuedChildPathsResume*`, `TestResumedSweep*`, `TestBrandNewSweep*` and `TestReconcileAmbiguousGroup*` pass `-count=5`; `TestCaptureAuto*` and `TestManualGenerate*` pass `-count=3` (50 s); the contention test passes 20 of 20 and, under heavier load, 36 of 36.
- `go vet` of domain, capture and cmd packages: clean.

## Verification performed
- Read: the brief, persona, shared context, REVIEW.md, my round-2 report, TESTS.md, the full delta for the domain rule, `cmd/agentic-lens/capture.go`, `capture/reconcile.go` (`RetainAcceptedChildLinks`), `children.go` (queued loop), and every test file changed in the delta.
- Mutation runs: 40 mutants in scratch copies (listed above), each restored from the checkout afterwards. Several first attempts did not compile (unused variable or an inert regex); those were redone in a valid form and only the valid result is reported.
- Probes: random append-only property test, key-value probe, no-`web/dist` import guard copy.
- Not done: whole-package suites for capture and cmd, the incident-scale divisor 1 run and the integrated `make` gate. I reuse TESTS.md for them. I did not independently re-run author mutations M1 to M6, M8, M9, M12, M13 to M19, M21b-through-M25 in their original form (I ran equivalents), nor the large-store oracles.

## Gaps
- No effort level was verifiable at runtime.
- Whether the parent wants to accept T3-2 as a documented limit (it is already listed for the exit mapping, but not for the positive wording) is a parent decision.
