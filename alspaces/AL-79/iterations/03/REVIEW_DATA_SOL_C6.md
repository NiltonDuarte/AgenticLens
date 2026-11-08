# AL-79 I03 C6 data and failure correction review

APPROVED within the C5-to-C6 correction scope. New findings: 0.

Run: `AL79/I03/C6/DATA_SOL`. Requested profile: `gpt-6.1-sol/high`; actual runtime model, usage and cost were not independently observed. Base: `3a62a726e029adb7c92fae26d8fe687339a21440`. The supplied 70-path candidate inventory matched SHA-256 `09809118d0960798320244143f64ec67644627820c4c064d59a7b63ed9416283` before and after investigation. Comparing C5/C6 inventories identified only `user_install.go` and `user_install_coordinator_test.go` as changed. No sibling C6 report was read.

The supplied deleted-owner finding is corrected. In `backend/internal/adapters/capture/user_install.go:122`, recursive readonly pending-resource discovery now precedes home creation and the machine lease. A surviving shared reference reaches `referencedUserJournal` (`user_install_transaction.go:117`), which rejects an unavailable owner before `MkdirAll` can make its missing private journal resemble a completed transaction. Both interrupted and committed-prefix public-entrypoint cases assert an error, no recreated home, and unchanged shared reference, legacy files and registry facts. Discovery credit for the supplied severity-5 defect remains with the C5 behavior reviewer (`gpt-6.1-sol/high`); this review verifies its correction.

Three consequential hypotheses were assessed:

- **Earlier discovery could authorize recovery from stale state:** refuted within the inspected path. After leases are acquired, `user_install.go:190` rediscovers every held resource and requires the full fresh resource set; validation precedes recovery writes at lines 215–221. New resources force retry. Moving discovery does not change reference identity, target whitelist or private-journal commit authority.
- **An absent ordinary home could become permanently un-installable:** refuted by the fresh-home public control. Without pending authority, installation creates six owned hooks and a current receipt, removes transaction records and preserves registry bytes.
- **The fix could weaken prior recovery guarantees:** no changed authority or recovery implementation was found. Carry the unchanged C5 data evidence for private undo, cross-home recovery, stale-reference isolation, edited/corrupt-record refusal, ordinary-adoption refusal and receipt protocol alignment. The earlier deleted-owner conclusion is specifically superseded by the corrected public-entrypoint evidence here. Manager, collector and portable-ledger paths were unchanged and were not re-audited.

Executed from `backend/`: `go test -race ./internal/adapters/capture -run '^(TestUserMigrationPublicInstallPreservesDeletedOwner|TestUserRecoveryRejectsResourcesAddedAfterDiscovery|TestRootRecoveryLocks.*)$' -count=1`: PASS, 2.735s. Inspected the real checkpoint fixture, snapshot assertions and fresh-resource scheduling oracle. Reused the supplied C6 coordinator race pass (5.264s), C5 adversarial/commit-authority evidence and unchanged C5 focused results.

Full integrated `make`, actual machine installation and delivery remain parent-owned gates. No Windows runtime, power-loss durability or new live Codex trust certification is claimed. Actual home was untouched; only this assigned evidence report was written.
