# AL-68 behavior review: behavior-sol-02

Scoped code verdict: CHANGES_REQUESTED. CAP-01, CAP-03, CAP-04 and the original CAP-11 isolation trigger are corrected, but their bounded child-recovery interaction introduces the reproducible P2 defect below. Overall outcome: CHANGES_REQUESTED; required real Codex evidence also remains INCOMPLETE. This is not support or delivery approval.

## Revision and identity

- Phase: revision-aware re-review, behavior/integration focus, not broad discovery. Requested profile: codex/gpt-6.1-sol/high. Actual source-record model, effort, usage and cost were not independently established; requested settings are not claimed as measured metadata.
- Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Reviewed corrected-02 fingerprint: `17fdae376119de798b382af56952351564389a69a174e5b0049259d387151cbc`. Independently recomputed from sorted tracked/nonignored untracked regular repository files outside alspaces using path/NUL/content/NUL; exact match with CANDIDATE-02.json after isolated probing.
- Requirements: CONTRACT.md and PLAN.md, including the user's actual hook-to-agent interaction/no-guessing clarification. Reconciled CORRECTIONS.md, FINDINGS.json, prior own behavior-sol report, implementation/test/setup reports and raw smoke evidence. No sibling persona or sibling reviewer conclusion loaded.
- Product and tests remained read-only. Only this assigned report is written in the checkout. Probe additions used disposable Go overlays outside the checkout. No children, full-suite duplication, staging, commits, telemetry or publication.

## New correction-interaction finding

### B-SOL-02-1: OpenCode bounded child discovery never advances past already known children (P2, introduced by correction)

Locations: `backend/internal/adapters/capture/children.go:139`, `:146`; `backend/internal/adapters/opencode/capture_snapshot.go:126`.

Supported operation: a bound OpenCode root has two source-proven native children that have delivered their own lifecycle registrations but remain unassigned; reconcile the parent's ticket repeatedly with `--max-sessions 1`. The source query orders children by ID and applies LIMIT before the caller discards already known ticket children. There is no discovery continuation or exclusion of those known children. After child-a is inherited, every later pass returns child-a again; child-z can never reach inheritance or extraction. The normal owner-refresh cursor rotates only registrations already bound to the ticket and cannot repair this omission. The fixed-point propagation cannot repair it either: child-z has neither an extracted snapshot nor ParentID because unassigned owners are intentionally excluded from ordinary extraction.

A read-only Go overlay exercised Store.Reconcile with a real temporary SQLite source containing root, child-a and child-z, authoritative parent_id=root for both children, an explicitly bound root, and both children locally registered/unassigned. Eight passes with session budget 1 and the normal byte budget produced:

```text
pass=1 registrations=[root:pending child-a:pending] pending_children=0 attempts=1
pass=2 registrations=[root:current child-a:pending] pending_children=0 attempts=1
pass=3 registrations=[root:pending child-a:current] pending_children=0 attempts=1
...
pass=8 registrations=[root:current child-a:pending] pending_children=0 attempts=1
later source-proven child permanently unassigned: id=child-z ticket="" state=unassigned parent=""
```

The query/guard establish that further unchanged passes cannot progress. A control changing only the budget to 2 passes, inherits both children in the first pass, and subsequently extracts both. This is a continuation defect, not unavailable source evidence or guessed ownership. It also occurs whenever a parent's child count exceeds its available per-pass child limit, including larger default-budget cases. PendingChildren remains zero for the omitted OpenCode child, so the ticket result does not disclose this missing inherited owner.

Expected: bounded repeated passes eventually visit every authoritative child while preserving unrelated-root isolation; unresolved child work remains explicit. Required correction: persist a parent-scoped discovery continuation or otherwise exclude/advance beyond already handled source child IDs, with a regression where the child set exceeds one pass's allowance. Preserve exact source-family selection, conflict rejection and byte/session limits. Parent owns consolidation and final finding ID assignment.

Probe command from backend:

`go test -overlay /tmp/al68-behavior-sol02-z35o3b79/overlay.json ./internal/adapters/capture -run '^TestReviewOpenCodeChildDiscoveryContinuesAcrossBoundedPasses$' -count=1 -v`

Observed exit 1, expected assertion failure. Control command uses `/tmp/al68-behavior-sol02-z35o3b79/control-overlay.json`; observed exit 0. Probe source and overlays remain outside Git. Neither run used the race detector; the defect is deterministic query/continuation behavior, not a race allegation.

## Requested correction dispositions

| ID | Disposition in this scope | Evidence and consequential path |
| --- | --- | --- |
| CAP-01 | Fixed for reviewed stale-retention triggers | Reconcile marks RefreshedSessions only after successful accepted extraction. main.go:336 fills missing canonical owners from retained local fallback, merges accepted registry refreshes, then lets later explicit extraction win. Existing mergeSessions replaces complete owner values rather than taking fieldwise maxima. `TestCaptureGenerationRetainedRegistryCannotOverrideNewerCanonicalOrExplicitSnapshot` reaches unequal 10/30 registry/canonical facts, missing registry source, archived-path explicit success, other-harness preservation and repeated AL-62 row preservation. Its corrected pass is recorded in TEST_REPORT.md and full make. |
| CAP-03 | Fixed for known exact-binding selector trigger | Both Codex and OpenCode successful selection paths call validateCaptureSelection before merge/publication. The guard compares exact source, absolute store and every selected session ID against known nonempty other-ticket bindings, including selected descendants. `TestCaptureGenerationRejectsExplicitSelectionOfAnotherTicketOwner` asserts rejection and absence of a published conflicting ledger. No claim is made that bindings from a deliberately different/unavailable registry are known. |
| CAP-04 | Fixed for start-without-stop trigger | ReceiveCodex persists opaque SubagentStart agent_id/parent/store in pending_starts. recoverChildren reads bounded authoritative session_meta and requires actual ParentThread equality before inheritChild; it never invents a source ID from opaque agent_id. `TestReconcileRecoversSourceProvenChildAfterStartWithoutStop` uses different interface/source IDs and no Stop. Unresolved opaque correlation remains separately pending, as documented; finding B-SOL-02-1 concerns the adjacent OpenCode route. |
| CAP-11 | Original unassigned starvation/report leak fixed; new bounded-child regression blocks code approval | Reconcile indexes/extracts only ticket-bound owners; report projection filters the ticket again; local List preserves all unassigned registrations. `TestReconcileUnassignedRootsDoNotConsumeTicketWorkOrPublishIdentities` forces four unrelated earlier registrations and a one-owner/one-source budget, requires the bound owner current, excludes unrelated report identities and retains all five local rows. The required source-proven inheritance route has the new continuation defect above. |

## Verification reused and limits

Independently inspected the affected CLI selection/merge/report path; Reconcile's refreshed/fallback projection, ticket filtering and unchanged mergeSessions consumer; hooks and registration/binding; Codex metadata child recovery; OpenCode exact-family child query and its caller. Challenged the interaction of excluding unassigned owners with bounded inherited-child recovery, which produced B-SOL-02-1. No other new blocking defect was established in this bounded pass.

Reused TEST_REPORT.md's coordinated corrected six-package race checks and durable correction oracles. Parent VERIFICATION.md records full corrected `make` exit 0 with unchanged reviewed fingerprint: Go vet/race, frontend typecheck/lint/781 tests/build, release checks, 14 distribution checks and native builds. Parent also verified embedded asset tree/binary bytes. These passing checks do not exercise or refute the over-limit child-discovery probe. They remain pre-main-integration evidence; parent owns integration and its required make gate.

SMOKE.md records a fresh actual OpenCode 1.19.5 Linux CLI binding/tool/source round trip after corrections, plus earlier independent-root/native-child runtime evidence. Reused this as runtime evidence rather than running another live model interaction. Source model and observed-zero cost facts in that report remain source claims, not reviewer model attribution. SETUP_IMPLEMENTATION.md records corrected compatibility/rollback and manual guidance evidence; this behavior review does not certify its release/platform scope.

The required real Codex human trust/callback/agent-binding/native-child source evidence remains missing. No fixture, source trace, installed hooks or enabled feature substitutes for it. No trust bypass or additional human request was made by this reviewer. Windows/macOS native behavior, OpenCode v2 native smoke, resumed runtime and current T3 Codex children retain the documented limits. Resolving the code finding alone would still leave the overall acceptance evidence INCOMPLETE until the authorized runtime boundary is met or explicitly changed by the user.
