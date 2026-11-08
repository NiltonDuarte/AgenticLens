# C5 behavior correction review: CHANGES_REQUESTED

Run: `AL79-I03-C5-BEHAVIOR-SOL` (`/root/i03_behavior_sol_c5`), independent correction pass. Requested profile: Codex / gpt-6.1-sol / high; runtime model metadata was not independently exposed. Base `3a62a726e029adb7c92fae26d8fe687339a21440`; frozen 70-path SHA-256 `6e7390d021da02c37ebe5c4213a6afa908bb164e57e8afef87b7f900be7955fb`. `verify_c5.py` matched before and after investigation. Scope: eight C5 correction paths and relevant unchanged public installer, migration and adoption callers. Product files remained read-only.

## New blocking finding

**I03-BEHAVIOR-C5-01 — severity 5: retry from a deleted private owner bypasses unknown-owner preservation.** `backend/internal/adapters/capture/user_install.go:123` creates the current owner directory before discovering the shared migration reference. After an actual `home-hooks` checkpoint interruption, delete that private owner home and retry explicit migration with the same effective Codex home. This mkdir recreates the missing owner before `referencedUserJournal` checks it (`user_install_transaction.go:119–128`). Its absent private journal is consequently interpreted as committed/reference-only state. The public installer returns success and removes the shared reference instead of reporting the missing private authority as unknown and preserving it.

A disposable Go overlay invokes the real production checkpoint and `InstallUserCodex`; observed result: `err=<nil>, recreated owner exists=true, shared reference exists=false`. Existing `TestUserMigrationDeletedOwnerPreservesUnknownSharedReference` passes because it calls recovery directly, bypassing the installer mkdir. Preserve the missing-owner fact through the public installer entry path; add that regression before approval. Probe and raw output: `/tmp/al79-c5-behavior-sol-b3fze750/overlay.json` and `missing-owner-probe.log`.

## Other hypotheses and evidence

The owner/checksum reference and private-journal commit sequence prevent another home from retaining rollback authority after migration commits. Stale references do not enroll a newer transaction at another coordinator. Fresh closure checks reject newly introduced unleased resources; ordinary repository installation checks pending migration before treating removed callbacks as absence. The protocol-2 receipt reader now matches the actual writer, and its consumer enables migrated adoption through an inactive legacy manifest without depending on the historical wrapper.

Executed focused race tests for current receipt install/removal, cross-home recovery, stale references, previous-home/resource leases and fresh discovery: PASS, capture package 3.177s. Executed actual CLI inactive-legacy binding case: PASS, 3.499s. The overlay public-installer counterexample FAILS as expected while its direct-recovery control passes (combined 0.558s). Reused supplied acknowledged-worker installer and coordinator-prefix evidence after inspecting their production boundaries and assertions; no broad matrix rerun.

Manager selection, argv, normal trust, collector and portable-ledger changes are outside C5 and retain parent-owned revision-bounded approvals. Full make and actual Linux installation remain parent gates. Windows runtime and power-loss durability are not certified. One new finding; no independent discovery credit claimed for supplied C5 defects.
