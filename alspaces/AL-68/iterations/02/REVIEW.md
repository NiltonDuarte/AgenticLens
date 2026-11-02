# AL-68 iteration02 consolidated review

Current scope verdict: **APPROVED; final integration gate pending**. All required persona/profile cells are complete or explicitly carried from candidate06 to candidate07. No confirmed blocking finding remains. This does not yet approve delivery or merge.

Candidate: candidate-02-07, based on de31cb1. All58 frozen source/test/docs paths match the reviewed snapshot. Candidate06->07 changes only installer store canonicalization, raw source-prefix expansion and the native AUTO21 oracle. Behavior/data/architecture/security approvals carry forward because their actual source guards, physical resolver, scheduler, state ownership, locks, journal, writers and schema are unchanged; paired release/test reviews verify the changed integration.

Model policy: S = Codex gpt-6.1-sol/high; L = Codex gpt-6-luna/medium. Requested and acknowledged settings are separate from source observations in [REVIEW_RUNS.json](REVIEW_RUNS.json). Observed entries exist only where exact source ID/first task/turn_context were independently matched. No review usage/cost is invented.

## Current gate roster

| Focus | S run | L run | Revision and result |
| --- | --- | --- | --- |
| behavior | iter02-behavior-sol-06 | iter02-behavior-luna-06 | candidate-02-06; APPROVED; unchanged-boundary carry to07 |
| data | iter02-data-sol-06 | iter02-data-luna-06 | candidate-02-06; APPROVED; unchanged-boundary carry to07 |
| architecture | iter02-architecture-sol-06 | iter02-architecture-luna-06 | candidate-02-06; APPROVED; unchanged-boundary carry to07 |
| tests | iter02-tests-sol-07 | iter02-tests-luna-07 | candidate-02-07; APPROVED |
| security | iter02-security-sol-06 | iter02-security-luna-06 | candidate-02-06; APPROVED; unchanged-boundary carry to07 |
| release | iter02-release-sol-07 | iter02-release-luna-07 | candidate-02-07; APPROVED |

Complete original reports are embedded verbatim at `archive.reports[run_id].content`; exact candidate diffs and supporting records are in `archive.documents`. Every original discovery, correction, refutation, cancelled/not-dispatched run and no-findings result is retained. [FINDINGS.json](FINDINGS.json) stores local IDs, raw claims, shared discovery/model credit and verification links.

## Findings and attribution

Nineteen confirmed defects/oracle gaps are corrected. AUTO05 and AUTO10 remain refuted with their original reports intact. Discoveries during re-review retain that phase; confirmation of supplied findings does not gain discovery credit.

| ID | Disposition | Discovering focus/profile | Finding |
| --- | --- | --- | --- |
| AUTO-01 | fixed | behavior/S, data/S, security/S | Retained baseline owners are not checked against the requested ticket |
| AUTO-02 | fixed | behavior/S | Continuous callbacks indefinitely reset publication debounce |
| AUTO-03 | fixed | behavior/S, behavior/L, architecture/S, release/S, release/L | Existing automatic worker continues writing after disable |
| AUTO-04 | fixed | behavior/S, architecture/S | Status fails for a valid registry without publication_request table |
| AUTO-05 | refuted | behavior/L | Automatic path publishes first ledger with missing explicit title |
| AUTO-06 | fixed | data/S | Missing-leaf aliases allow ledger and summary destinations to overlap |
| AUTO-07 | fixed | tests/S | In-flight regression accepts acknowledgement of unseen work |
| AUTO-08 | fixed | architecture/S | Installed state paths silently diverge after YAML path changes |
| AUTO-09 | fixed | architecture/S | Subdirectory generation selects process cwd instead of owning configuration root |
| AUTO-10 | refuted | architecture/L | Configurable work budgets have no practical upper ceilings |
| AUTO-11 | fixed | security/S, security/L | Publication recovery trusts arbitrary journal artifact destinations |
| AUTO-12 | fixed | security/S | Direct callback transcript path bypasses selected Codex source-store proof |
| AUTO-13 | fixed | release/L | Setup guidance contradicts automatic publication and optional reports |
| AUTO-14 | fixed | release/S | True project aliases cannot reinstall or remove owned installation |
| AUTO-15 | fixed | release/S | Installer resolves omitted source store from caller cwd instead of selected project |
| AUTO-16 | fixed | security/S | Dangling output symlink bypasses project destination containment |
| AUTO-17 | fixed | security/S | Lexical cleaning precedes physical symlink traversal |
| AUTO-18 | fixed | data/S | Dangling ledger leaf alias acquires independent publication lock |
| AUTO-19 | fixed | security/S, data/S | Atomic artifact writers replace dangling aliases and strand publication identity |
| AUTO-20 | fixed | security/S | Raw source symlink-dotdot traversal bypasses selected Codex store membership |
| AUTO-21 | fixed | release/S | Installer changes physical source-home semantics before emitting hooks |

Nonblocking suggestion from test L candidate07: AUTO21 does not separately assert registry state after callback exit and before manual generation. This does not defeat its demonstrated old/new installer counterexample. Persisted receive/binding and real hook-driven publication have complementary evidence. The limitation and optional strengthening remain recorded, rather than being relabeled a production defect or silently erased.

## Verification and runtime boundary

Candidate06 full `go test -race ./...` and `go vet ./...` passed; candidate07 targeted installer/config/capture race checks and independent old-installer/old-config overlays pass/fail as expected. The final integrated `make` must run after the implementation commit and upstream integration. Earlier make01-04 outputs are retained as evidence of their own candidates, not current delivery approval.

Corrected-handler real runtime evidence: Linux Codex0.166, three independent roots and two completed native children; Linux OpenCode1.19.5, two same-ticket/store directories with disjoint12/3owners, including five children in the first. Exact source project/session/native-parent/model/usage and available costs were compared; automatic publication required no contributing-agent generate. Human Codex trust was approved by the user; hook definition bytes were not bypassed. Current path/source changes are covered by native synthetic regressions and bounded unchanged existing-directory carry-forward.

Unverified broader cases remain explicit: Codex SubagentStart delivery, active/interrupted children before Stop, non-Linux runtime, actual forty-minute elapsed operation and physical power-loss behavior. Synthetic long-growth fixtures, cross-builds, native crash and persisted-journal recovery do not prove those claims. The accepted supported boundary does not promise universal harness/platform capture or hostile concurrent filesystem mutation. Top-level completion is observability finalization, not ticket completion.

## Historical run roster

Every row links by run ID to the full report/archive record. Cancelled not-dispatched cells are not approvals; interrupted obsolete reviews are retained and replaced by later bounded runs.

| Run | Focus/profile | Phase | Candidate | Execution / verdict |
| --- | --- | --- | --- | --- |
| iter02-behavior-sol-01 | behavior/S | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-behavior-luna-01 | behavior/L | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-data-sol-01 | data/S | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-data-luna-01 | data/L | discovery | candidate-02-01 | completed / APPROVED |
| iter02-architecture-sol-01 | architecture/S | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-architecture-luna-01 | architecture/L | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-tests-sol-01 | tests/S | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-tests-luna-01 | tests/L | discovery | candidate-02-01 | completed / APPROVED |
| iter02-security-sol-01 | security/S | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-security-luna-01 | security/L | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-release-sol-01 | release/S | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-release-luna-01 | release/L | discovery | candidate-02-01 | completed / CHANGES_REQUESTED |
| iter02-behavior-sol-02 | behavior/S | re_review | candidate-02-02 | completed / APPROVED |
| iter02-behavior-luna-02 | behavior/L | re_review | candidate-02-02 | completed / APPROVED |
| iter02-data-sol-02 | data/S | re_review | candidate-02-02 | cancelled / none |
| iter02-data-luna-02 | data/L | re_review | candidate-02-02 | cancelled / none |
| iter02-architecture-sol-02 | architecture/S | re_review | candidate-02-02 | cancelled / none |
| iter02-architecture-luna-02 | architecture/L | re_review | candidate-02-02 | cancelled / none |
| iter02-tests-sol-02 | tests/S | re_review | candidate-02-02 | cancelled / none |
| iter02-tests-luna-02 | tests/L | re_review | candidate-02-02 | cancelled / none |
| iter02-security-sol-02 | security/S | re_review | candidate-02-02 | completed / CHANGES_REQUESTED |
| iter02-security-luna-02 | security/L | re_review | candidate-02-02 | completed / APPROVED |
| iter02-release-sol-02 | release/S | re_review | candidate-02-02 | cancelled / none |
| iter02-release-luna-02 | release/L | re_review | candidate-02-02 | cancelled / none |
| iter02-data-sol-03 | data/S | re_review | candidate-02-03 | completed / CHANGES_REQUESTED |
| iter02-data-luna-03 | data/L | re_review | candidate-02-03 | completed / APPROVED |
| iter02-architecture-sol-03 | architecture/S | re_review | candidate-02-03 | cancelled / none |
| iter02-architecture-luna-03 | architecture/L | re_review | candidate-02-03 | cancelled / none |
| iter02-tests-sol-03 | tests/S | re_review | candidate-02-03 | cancelled / none |
| iter02-tests-luna-03 | tests/L | re_review | candidate-02-03 | cancelled / none |
| iter02-security-sol-03 | security/S | re_review | candidate-02-03 | completed / CHANGES_REQUESTED |
| iter02-security-luna-03 | security/L | re_review | candidate-02-03 | completed / APPROVED |
| iter02-release-sol-03 | release/S | re_review | candidate-02-03 | cancelled / none |
| iter02-release-luna-03 | release/L | re_review | candidate-02-03 | cancelled / none |
| iter02-data-sol-04 | data/S | re_review | candidate-02-04 | completed / CHANGES_REQUESTED |
| iter02-data-luna-04 | data/L | re_review | candidate-02-04 | cancelled / none |
| iter02-architecture-sol-04 | architecture/S | re_review | candidate-02-04 | cancelled / none |
| iter02-architecture-luna-04 | architecture/L | re_review | candidate-02-04 | cancelled / none |
| iter02-tests-sol-04 | tests/S | re_review | candidate-02-04 | completed / APPROVED |
| iter02-tests-luna-04 | tests/L | re_review | candidate-02-04 | cancelled / none |
| iter02-security-sol-04 | security/S | re_review | candidate-02-04 | completed / CHANGES_REQUESTED |
| iter02-security-luna-04 | security/L | re_review | candidate-02-04 | cancelled / none |
| iter02-release-sol-04 | release/S | re_review | candidate-02-04 | cancelled / none |
| iter02-release-luna-04 | release/L | re_review | candidate-02-04 | cancelled / none |
| iter02-data-sol-05 | data/S | re_review | candidate-02-05 | completed / APPROVED |
| iter02-data-luna-05 | data/L | re_review | candidate-02-05 | cancelled / none |
| iter02-architecture-sol-05 | architecture/S | re_review | candidate-02-05 | completed / APPROVED |
| iter02-architecture-luna-05 | architecture/L | re_review | candidate-02-05 | cancelled / none |
| iter02-tests-sol-05 | tests/S | re_review | candidate-02-05 | cancelled / none |
| iter02-tests-luna-05 | tests/L | re_review | candidate-02-05 | cancelled / none |
| iter02-security-sol-05 | security/S | re_review | candidate-02-05 | completed / CHANGES_REQUESTED |
| iter02-security-luna-05 | security/L | re_review | candidate-02-05 | cancelled / none |
| iter02-release-sol-05 | release/S | re_review | candidate-02-05 | completed / APPROVED |
| iter02-release-luna-05 | release/L | re_review | candidate-02-05 | cancelled / none |
| iter02-behavior-sol-06 | behavior/S | re_review | candidate-02-06 | completed / APPROVED |
| iter02-behavior-luna-06 | behavior/L | re_review | candidate-02-06 | completed / APPROVED |
| iter02-data-sol-06 | data/S | re_review | candidate-02-06 | completed / APPROVED |
| iter02-data-luna-06 | data/L | re_review | candidate-02-06 | completed / APPROVED |
| iter02-architecture-sol-06 | architecture/S | re_review | candidate-02-06 | completed / APPROVED |
| iter02-architecture-luna-06 | architecture/L | re_review | candidate-02-06 | completed / APPROVED |
| iter02-tests-sol-06 | tests/S | re_review | candidate-02-06 | completed / APPROVED |
| iter02-tests-luna-06 | tests/L | re_review | candidate-02-06 | completed / APPROVED |
| iter02-security-sol-06 | security/S | re_review | candidate-02-06 | completed / APPROVED |
| iter02-security-luna-06 | security/L | re_review | candidate-02-06 | completed / APPROVED |
| iter02-release-sol-06 | release/S | re_review | candidate-02-06 | completed / CHANGES_REQUESTED |
| iter02-release-luna-06 | release/L | re_review | candidate-02-06 | cancelled / none |
| iter02-tests-sol-07 | tests/S | re_review | candidate-02-07 | completed / APPROVED |
| iter02-tests-luna-07 | tests/L | re_review | candidate-02-07 | completed / APPROVED |
| iter02-release-sol-07 | release/S | re_review | candidate-02-07 | completed / APPROVED |
| iter02-release-luna-07 | release/L | re_review | candidate-02-07 | completed / APPROVED |
