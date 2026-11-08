# Behavior review C6 (gpt-6-luna, medium)

**Outcome: APPROVED for the assigned behavior/integration scope.** No new finding; numeric findings: 0. This is a bounded correction review, not approval of the full ticket.

## Scope and evidence

Reviewed candidate based on `3a62a726e029adb7c92fae26d8fe687339a21440`; frozen 70-path candidate SHA-256: `09809118d0960798320244143f64ec67644627820c4c064d59a7b63ed9416283`. Assigned run ID was not supplied. The start/end freeze check passed: `python3 /tmp/al79-native-entrypoint-0ta97kas/verify_c6.py` reported `MATCH` for that hash.

The supported operation is public user-hook installation for a selected repository. Its setup now enumerates affected legacy `.codex` directories and calls readonly `pendingUserHomes` before `MkdirAll(root)`. A shared reference whose private owner was deleted therefore returns an owner-unavailable error before the current home can be created. After lock acquisition, pending resources are rediscovered and lease-checked, and recovery authority is validated under the leases before any recovery writes. The C6 public-entrypoint test reaches this path for both an interrupted `home-hooks` prefix and a `committed` prefix; it asserts error, no recreated owner, and unchanged reference/capture artifacts. Its fresh-home control proves normal first installation still creates all six owned hooks without changing retained capture facts.

## Hypotheses and verification

- **C5 failure could remain through the public installer:** not reproduced after C6; the public test reaches the previously missing preparation boundary and verifies preservation.
- **Moving discovery earlier could reject valid first install:** the fresh-home control succeeds.
- **Cross-home recovery or lease guards could regress:** the supplied C5 coordinator, stale-reference and different-home evidence remains applicable to unchanged transaction logic. The focused race command covering `Coordinator|DifferentHome|StaleReference|Deleted|PublicInstall` passed locally: 5.415s. `git diff --check` passed.

The carried C5 behavior finding (I03-BEHAVIOR-C5-01) remains corrected: prior ordering let `MkdirAll` recreate the deleted owner, making the missing private journal appear to authorize shared-reference cleanup. C6 orders readonly discovery before creation and adds public-path coverage.

Full `make`, machine installation, native Codex trust/lifecycle and Windows runtime certification remain outside this review; parent delivery gates are still required.
