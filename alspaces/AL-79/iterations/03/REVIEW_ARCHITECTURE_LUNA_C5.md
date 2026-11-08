# AL-79 I03 C5 architecture review — gpt-6-luna/medium

**Outcome: APPROVED for the assigned architecture scope.** Run: AL79 I03 C5. Candidate inventory `CANDIDATE_C5.json` identifies base `3a62a726e029adb7c92fae26d8fe687339a21440` and frozen digest `6e7390d021da02c37ebe5c4213a6afa908bb164e57e8afef87b7f900be7955fb`; `verify_c5.py` matched before and after review. The correction scope is the three user-install/adoption adapters and associated tests/evidence. No product files were written.

## Findings

No blocking architecture finding. The change puts coordinator/reference parsing, private journal validation, lease-resource expansion, rollback and commit in the capture adapter that owns installation side effects. `user_adoption.go` delegates legacy checkout discovery and uses the same protocol constant as the writer; it does not become a second journal implementation. Domain and manager boundaries are untouched. The shared coordinator record contains owner/coordinator identity and a checksum reference; private hook configuration and undo images remain in the owner's home. Recovery validates known target paths and legacy state ownership before restoring bytes.

## Hypotheses investigated

- **Shared reference broadens rollback authority or exposes private undo data:** source trace shows recovery requires a matching owner journal and checksum, validates its exact legacy coordinator/state and restricts writes to known hook/receipt/manifest targets. Stale references do not enroll an unrelated journal. A missing owner remains an error; deletion does not authorize cleanup or recreate the checkout.
- **Resource discovery permits recovery before all relevant leases are held:** `installUserCodexWithCheckpoint` expands pending and current legacy resources, takes leases, rediscovers pending resources under those leases, checks fresh lease coverage, validates all journals, and only then recovers. The new repository callbacks are separately rejected when that checkout has a pending migration.
- **Protocol ownership is duplicated or mismatched:** writer and adoption reader both use `userCaptureOwnershipProtocol` (2), removing the observed literal drift.
- **New adapter state crosses the portable ledger boundary:** no capture ledger, source extraction, or persisted telemetry path changed. The portable ledger contract is therefore unaffected by this correction.

## Verification and limits

`python3 /tmp/al79-native-entrypoint-0ta97kas/verify_c5.py` matched the frozen digest at review start and end. `go test -race ./internal/adapters/capture -run 'TestUserMigration(Coordinator|DifferentHome|StaleReference|Deleted)|TestUserRecoveryRejectsResourcesAddedAfterDiscovery' -count=1` passed (5.160s). The recorded C5 adversarial evidence exercises actual durable prefixes, cross-home recovery, stale references, deletion and tamper controls. Full `make`, real machine installation, actual Codex home hooks, Windows runtime behavior and power-loss durability remain outside this review; no whole-repository or integration approval is implied.
