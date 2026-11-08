# AL-79 I03 C5 security correction review

**APPROVED** for the pending-installation reference and recovery boundary. New findings: **0**.

Run: I03 C5 SECURITY LUNA. Requested identity: Codex / gpt-6-luna / medium; runtime model metadata is unavailable. Reviewed base `3a62a726e029adb7c92fae26d8fe687339a21440` and the eight correction paths listed in `/tmp/al79-native-entrypoint-0ta97kas/CANDIDATE_C5.json`; directly inspected the new coordinator test. Frozen 70-path candidate: `6e7390d021da02c37ebe5c4213a6afa908bb164e57e8afef87b7f900be7955fb`. The required start verifier matched.

The source trace supports the stated ownership boundary: migration references are written before the private canonical journal and contain no before-images; the private journal checksum binds the reference; removing the private journal is the commit point. Recovery checks the coordinator's legacy receipt, journal resources and each unique rollback target against narrow allowlists before writes. The installer rediscovers resources and validates them under the acquired leases before recovery. A stale reference whose owner journal belongs to another coordinator does not enroll or recover that journal. Missing private owner state remains an error, while unknown edits to transaction targets are preserved. The adoption reader now accepts the protocol emitted by the actual installer.

Inspected the changed adapters, unchanged path canonicalization, migration ownership, atomic replacement and lease call paths, plus the new crash-prefix/privacy, cross-home, stale-reference, corruption and deleted-owner tests. Reused the recorded C5 focused race and installation/distribution checks; did not rerun tests. This assessment concerns installation recovery integrity and does not impose a private receipt requirement on portable ledger history.

Process note: I inadvertently opened `REVIEW_SECURITY_SOL_C5.md` while checking report-path state, after first reading the contract, plan and production diff. I then inspected the raw implementation and test oracles directly. This does not change the findings, but the parent should account for that exposure when assessing first-pass independence.

Limits: full integrated `make`, live Linux installation and native Codex trust evidence remain parent-owned. Windows runtime and power-loss durability were not exercised. The required end verifier also matched the frozen candidate.
