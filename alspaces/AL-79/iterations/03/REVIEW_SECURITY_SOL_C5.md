# AL-79 I03 C5 security correction review

APPROVED for the pending-installation reference boundary. New findings: **0**.

Run: I03 C5 SECURITY SOL. Requested reviewer: codex / gpt-6.1-sol / high, generic reviewer persona and shared review context. Runtime model metadata is unavailable. Reviewed base `3a62a726e029adb7c92fae26d8fe687339a21440` plus the eight correction paths in `/tmp/al79-native-entrypoint-0ta97kas/CANDIDATE_C5.json`, including direct reading of the untracked coordinator tests. Frozen 70-path candidate: `6e7390d021da02c37ebe5c4213a6afa908bb164e57e8afef87b7f900be7955fb`; verifier matched before and after review.

The three changed adapters and consequential unchanged migration, path resolution, file replacement and installation-lock helpers support the preserved contract. Publication writes the typed shared reference before the private full journal and business files; removing the private journal commits. Shared references contain owner/coordinator paths and checksums, without hook configuration or before-images. A matching private journal and exact body checksum are required for rollback; stale references for another coordinator neither enroll nor recover that transaction. Private-owner disappearance is an error. Target and resource allowlists are checked before lease enrollment, with fresh validation under all discovered leases before recovery writes. File rollback preserves unknown edits. Ordinary repository installation checks pending migration independently of remaining callback bytes. The protocol-2 adoption reader agrees with the actual writer.

Independent verification used a disposable Go overlay at `/tmp/al79-security-sol-c5-7i1vdpj7`; shared source was unchanged:

`go test -race -overlay=/tmp/al79-security-sol-c5-7i1vdpj7/overlay.json ./internal/adapters/capture -run '^TestSecurityC5StrictReferenceAndScope$' -count=1 -v`

All three probes passed: unknown reference JSON fields, an additional foreign rollback target with coherently recomputed journal/reference checksums, and an additional foreign resource with matching checksums. The real installer rejected every case without altering protected business files, either pending record, or unrelated foreign bytes. The coherent-checksum controls establish that checksum agreement alone cannot authorize an out-of-scope write.

Reused the recorded C5 coordinator crash/privacy/commit/stale-reference/corruption/deletion race evidence and current-shape lease/adoption checks after inspecting their source oracles. These are installation integrity and recovery checks in the trusted user installation context, not portable-ledger authenticity requirements. No source, ledger, executable admission, network maintenance or trust path changed in C5.

Limits: scoped approval does not certify the whole candidate. Full integrated `make`, actual Linux installation and normal human-trust capture evidence remain parent-owned. Native Windows runtime and machine power-loss durability were not exercised; checkpoint probes model process interruption. No actual HOME, credentials, hooks, trust, binaries, commits, pushes or telemetry were changed.
