# Data integrity review — AL-73

**Outcome: CHANGES_REQUESTED**

Reviewed the frozen candidate at baseline `b6ab0471224c8bcea7bc00adce867d3ba0de4a37` with the supplied working-diff fingerprint `47fd2495ce`. Scope: `Makefile`, `scripts/release-checksums.sh`, the setup downloader and bundle installer, and their release/installation fixtures. This is the first-pass data focus; no sibling conclusions were read. Requested run identity: Luna / medium (runtime model metadata was not independently available).

## Finding

**[P1] A failed repackage can leave the prior manifest attached to replaced artifacts.** In [Makefile](/home/niltonduarte/.t3/worktrees/AgenticLens/fix-AL-73-verify-releas-09ca2dd540083673/Makefile:201), `release-package` removes and recreates each public archive before checksum generation (lines 219–223). The helper correctly stages a new manifest and preserves the existing `SHA256SUMS` when hashing or sorting fails, but it does not stage the archives or binaries. On a repeated package operation, a nonzero hash producer therefore preserves the old manifest while the public archive and potentially binaries have already changed. Consumers see a manifest that no longer authenticates the published directory; a retry/failure does not preserve a usable prior release set.

I reproduced this in a disposable temp release directory: first `make release-package` succeeded; I changed one selected binary and made `sha256sum` emit normal output then exit 7; the second package exited 2 and left the manifest byte-identical, while `release-verify` then failed for both the changed binary and newly recreated archive. The current `test_failed_manifest_generation_preserves_previous_manifest` asserts only that the manifest bytes survive, so it misses this inconsistent state. Preserve the prior artifact set until the new inventory is completely generated/validated, or invalidate/remove the old manifest when artifacts are replaced and generation fails; the chosen behavior must not leave a success-looking stale receipt.

## Other paths assessed

- Checksum helper checks each inventory path is a readable regular file, stages generation in the release directory, relies on `pipefail`/`set -e` for producer and sort failures, validates exact inventory membership and uniqueness, checks bytes, then renames the staged manifest. A failed hash output followed by exit 7 is propagated.
- Bootstrap digest computation is guarded by `pipefail`; the selected remote manager record requires one matching entry and valid digest syntax before fetching/probing. Local manager and engine hash failures abort before manager operations. The setup flow reports completed engine installs separately from manager replacement, and does not print its final success claim after self-update failure. Existing Go manager replacement transaction writes executable, receipt and checksum sidecar under lock/recovery; this review did not find a new receipt corruption path in the inspected flow.
- Remote repository validation occurs before fetch on both paths. No concern found in the inspected digest, installer receipt, retry, or replacement paths apart from the release package inconsistency above.

## Verification and limits

- Manual isolated probe described above; direct result: failed package preserved `SHA256SUMS`, subsequent checksum verification returned 1 for replaced selected files.
- `python3 scripts/tests/test_platform_support.py` passed (4 cases).
- `python3 scripts/tests/test_installation.py` passed (16 cases).
- No hosted publication, real release binaries, concurrent packaging, or full `make` gate exercised. Verdict is limited to the supplied candidate and assigned data/failure focus.
