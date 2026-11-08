# AL-79 test effectiveness correction re-review — C6

## Identity and scope

- Run: `tests-luna-c6`; requested identity `gpt-6-luna/medium`. Runtime effort metadata is unavailable; this records the requested tuple.
- Candidate: `AL79-C6`, HEAD `8b15b2ee3367da07dd6fd13bc29570aed14e90f4`. Candidate source diff SHA-256: `d667888c16db0fd5eb96cf54802b3d3131d027e51e23a4170baf8904ec14a878`. Frozen C5-to-C6 test diff SHA-256: `dcac2faad7614d8a4aac98e20380fd96f866799b7850e914bdaecf396fb6d4e5`; it contains only `capture/reconcile_test.go` and `capture/user_install_transaction_test.go`. The intervening `main` merge is documented as docs-only. No source or test files were changed during this review.
- Read the approved contract and iteration plan, reviewer-tests persona/common procedure, C5 report, C6 evidence, changed tests, and relevant production reconciliation and transaction-journal paths. C4 production and the prior C5 three-test assertion transfers are unchanged and carried forward as assigned; this report makes no sibling-review conclusions.

## Findings

No blocking test-effectiveness gap found in the C6 test correction.

## Oracle review

The former TEST-200 child / TEST-100 parent setup reached the explicit-choice exemption and did not establish the inherited-binding conflict branch. The corrected `TestReconcileSuccessfulReadPreservesParentConflictAndNormalPendingRecovery` establishes a TEST-200 parent, queues the actual child transcript through `QueueChild`, and reconciles it to prove `source_parent` inheritance. It then changes the source parent to an as-yet-unregistered ID, performs a real `SessionStart` registration and successful source refresh, and only afterward registers that parent to TEST-100. This reaches the production branch that rejects contradictory inherited ownership after a complete read. Assertions retain the accepted source parent fact and child attribution/snapshot, independently bound parent state, refresh time and unchanged fingerprint, normal-Pending recovery, partial/failed diagnostics, all historical callback/binding/source/publication flags, degraded health, feedback and advisory. The independent explicit-child control now requires current registration and accepted/completed diagnostics while retaining its exact source ancestry/facts and parent isolation.

The helper correction reads the actual production journal once at `userInstallJournaled`, caches only its ordered change paths, checks applied checkpoint indexes against that list, and still checks the actual pending journal after interruption. Production writes the complete journal before applying indexed changes, so this preserves path-to-prefix selection while avoiding repeated decoding of large before-image/content payloads. The reviewed change does not drop assertions about durable prefixes, before-images, recovery, user edits, SQLite preservation, or locks.

## Verification and limits

Reused the tester's exact-candidate evidence: the two reconciliation tests passed under `go test -race` in 1.880s, with both cases executed and no skips. All five durable-prefix cases passed in 107.90s and user-edit recovery passed in 21.49s. The broader helper diagnostic command was deliberately interrupted during the remaining lock matrix; that matrix is incomplete and is not reported as passing. The source inventory is the stated 314-file hash `f1b674667f1247eeccfe01640f5353d27a27e9794831aa2b99ce6d917bd6195b`. No full-suite or `make` result is claimed here. Parent-owned full `make`/PR CI remain separate gates. Native Windows execution remains unverified, and this test-only review supplies no new Codex runtime/trust evidence.

## Outcome

**APPROVED — scoped C6 test-effectiveness correction re-review.** The corrected oracle reaches the intended inherited-conflict branch and preserves the distinct positive explicit-child path. The journal helper still selects real production durable prefixes using the production journal. Approval covers only these two test-file changes and their reviewed oracles; it does not close the interrupted lock matrix, integrated gate, Windows runtime gap, or parent consolidation.
