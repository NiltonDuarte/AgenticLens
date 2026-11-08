# AL-79 test effectiveness correction re-review

## Identity and scope

- Requested run: `tests-luna-c4`; requested identity `gpt-6-luna/medium`. Runtime effort metadata is unavailable; this report records the requested tuple only.
- Candidate: `AL79-C4`, base `4df45967c801824500ac2ea39a8f52fa9cbdcf66`, source diff SHA-256 `9d1ca49a319643feeb13774a28942037cfff9d175884d53a8e8495660a90e265` (matches `CANDIDATE_C4.json`).
- Focus: correction re-review of TEST-01 and the supplied transaction/partial-analysis oracles; test effectiveness only. I read the approved contract/plan, the prior Luna test report, parent review context and relevant raw evidence, but no sibling reports.
- Source and test files remain frozen/read-only. No product or test files were changed, and no broad suite was repeated.

## Findings

No blocking test-effectiveness gap found in the requested C4 scope.

## Correction oracles inspected

| Requirement / hypothesis | Actual oracle and disposition |
|---|---|
| TEST-01: new root/home migration journal can recover after durable-prefix interruption while retaining pending facts. | `TestUserInstallationDurablePrefixesRecoverExplicitMigration` interrupts the production transaction after journal publication and after actual legacy-hook, dispatcher and home-hook mutations, then invokes normal recovery before a new transaction. It compares restored artifacts to exact before-images and completes the migration. Its fixture includes a real SQLite registration/pending publication and verifies the database bytes survive. `TestInterruptedUserMigrationPreservesEditAndResumesAfterResolution` separately protects a later user edit, durable journal retention and explicit resumption after resolving the test-owned conflict. These exercise the new user transaction journal and recovery path, not the older project-journal assumptions. The focused recovery oracle log records the named mutation causing the test to fail; the unmutated focused run passes.
| INST-01: unknown executable legacy callback must block explicit migration. | The shipped POSIX distribution fixture copies the actual installed `SessionStart` callback, changes only its status text so ownership becomes unknown while its command remains executable, and adds it as a second callback. Explicit migration is attempted for both the original checkout and linked worktree; it must fail without changing owned-path snapshots, deactivating the legacy anchor, or activating home hooks. Removing the unknown callback explicitly then permits migration and preserves the foreign callback and pending state. This is a reachable installer/dispatcher fixture, not a marker-only stub.
| DATA-02/TEST-01: installer resource discovery racing a second installer must not write without the newly required lease and must give a retry oracle. | `TestUserRecoveryRejectsResourcesAddedAfterDiscovery` starts a real second installation inside the first installer's resource-discovery checkpoint. The second invocation uses the first home as its install root and a distinct Codex home, then interrupts after writing that home's actual hooks. The test takes that home's real installation lease and requires the first call to request retry while preserving the actual journal and artifact snapshots byte-for-byte. After lease release, fresh discovery recovers the interrupted before-image and completes installation. The test does not handcraft a policy-shaped journal or replace locking with a mock. The earlier failing mutation log demonstrates the no-retry regression; `focused.log` records the corrected affected-package pass.
| DATA-05: separately published external-parent child must use automatic `--append` and expose correct fresh/cached partial analysis. | `TestAutomaticParentAndExplicitOtherTicketChildPublishIndependently` runs the actual CLI with `--capture-auto` and `--append` for parent and separately bound child ticket outputs. It checks child ownership/parent ID and input facts, byte-preservation of the parent's ledger, and the actual child Markdown's unavailable-context wording with no clean-session claim. `TestExternalParentAnomalyCoverageRemainsPartial` covers fresh, recorded-empty and recorded-warning histories: unavailable-source preservation, independent child analysis, retained group findings and correct completeness/unrecorded-scope state. `TestReadDerivationPreservesUnrecordedAndRecordedEmptyAnalysis` and the formatter control additionally protect unavailable versus recorded-empty interpretation and visibility of cached warnings. The disposable coverage mutation fails the fresh and cached-warning cases; the focused corrected run passes.

## Unchanged-scope carry-forward

The prior Luna first-pass checks remain applicable to unchanged paths: retained worker executable identity across mutable build replacement; user-install ownership/idempotence and source replacement; selected local configuration precedence; and native Codex callback/source-child identity. The C4 correction evidence does not claim new native execution or alter those contracts. C4's added executable residual-callback oracle complements the earlier normal install/removal coverage. Earlier true native Linux/Codex evidence remains revision-bounded to its recorded build.

## Verification and limits

- Inspected reusable evidence: `/tmp/al79-data-sol-c4-probes/focused.log`, `recovery-oracle.log`, and `coverage-oracle.log`. The latter two are expected failing disposable-mutation probes; they are evidence that the named regressions are detected, not candidate test failures. `focused.log` reports passing affected Go packages. The C4 tester records the final targeted package command and result in `TEST_EVIDENCE.md` and `/tmp/al79-tester-c4/final-targeted.log`.
- Inspected distribution fixture source for the exact executable unknown-callback residual case. C3 POSIX capture/installation results and the changed C4 migration oracle are the relevant host-side distribution evidence; I did not rerun broad suites.
- Parent-owned integrated `make`/race remains pending. Native Windows hosting/runtime stage remains unexecuted locally. These stages are not satisfied by compilation or static/helper checks. No new C4 model/native session was run; existing normal-trust native Codex Linux evidence remains bounded to its recorded revision.

## Outcome

**APPROVED — scoped test-effectiveness correction re-review.** The specified recovery, unknown executable hook, real installer discovery race, separate automatic child append, and fresh/cached partial-analysis controls reach their claimed production boundaries and distinguish the corrected behavior from the supplied mutations. This outcome covers only the requested test focus and does not replace parent consolidation, the pending integrated gate, or native Windows runtime verification.
