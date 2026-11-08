# C6 behavior correction review: APPROVED

Run `AL79-I03-C6-BEHAVIOR-SOL` (`/root/i03_behavior_sol_c6`), bounded correction re-review. Requested profile: Codex / gpt-6.1-sol / high; runtime model metadata was not independently exposed. Base `3a62a726e029adb7c92fae26d8fe687339a21440`; frozen 70-path SHA-256 `09809118d0960798320244143f64ec67644627820c4c064d59a7b63ed9416283`. `verify_c6.py` matched at start and end. C5/C6 inventory comparison confirms only `user_install.go` and `user_install_coordinator_test.go` changed. Product/tests remained read-only; no sibling C6 conclusions were consulted.

**Resolved finding I03-BEHAVIOR-C5-01, severity 5.** Public installation now completes readonly pending-resource discovery before `MkdirAll` (`backend/internal/adapters/capture/user_install.go:123`). A shared reference to a deleted private owner reaches `referencedUserJournal` while the owner is still absent, so retry errors without manufacturing completed-transaction evidence. I reran the original C5 production overlay against C6:

`go test -overlay /tmp/al79-c5-behavior-sol-b3fze750/overlay.json ./internal/adapters/capture -run '^TestBehaviorC5DeletedCurrentOwnerRemainsUnknown$' -count=1 -v`

PASS, 0.082s: owner-unavailable error, recreated owner `false`, surviving shared reference `true`; protected recovery/business bytes remain unchanged. This directly reverses the supplied C5 counterexample. No new findings or discovery credit.

The new public regression exercises real interrupted and committed prefixes, owner deletion, public retry, directory absence and retained-reference/fact snapshots. Its fresh-home control has no pending authority and verifies successful migration, six owned native hooks, protocol-2 receipt, unchanged capture database and transaction cleanup. Earlier discovery does not authorize recovery: the existing fresh resource checks under leases, required-lease checks and validation-before-recovery remain at `user_install.go:188`. Resource changes during preparation therefore require safe retry; another home's interrupted transaction remains leased and recovered before new business writes.

Reused tester raw PASS (5.264s) for public controls, coordinator prefixes, different-home recovery, stale references and deletion cases; its six-file source/test inventory matches C6. Also retained prior unchanged C5 lease, late-resource, idempotence, adoption and CLI evidence after tracing their assertions. No broad suite was repeated.

This approval covers the C6 ordering correction and preserves the other C5 behavior conclusions. Full `make` and actual machine installation remain parent delivery gates. Normal home trust was untouched; no live trust exchange, Windows runtime/CI certification or machine power-loss durability is claimed. Manager transport, collector and portable-ledger scopes retain their separately bounded evidence.
