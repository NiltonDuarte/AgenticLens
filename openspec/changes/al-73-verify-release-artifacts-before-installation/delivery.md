# fix(AL-73): verify the complete release inventory before installation

## Pull request

Release packaging and verification now derive the required files from the selected version and platforms. Each archive, manager and engine needs one well-formed SHA-256 entry. Packaging validates the staged manifest and file contents before replacing the final manifest. Verification repeats those checks before invoking a selected host binary.

The setup downloader requires a successful digest computation and one valid record for the selected manager. The bundle installer and binary bootstrap validate remote repository overrides before fetching from the exact tag. A later component failure exits unsuccessfully while preserving previously completed engine installs and the failed component’s prior state.

Local fixtures exercised partial manifests, hash and sort failures, stale host files, malformed manager records, exact URLs and total and partial installation failure. Receipt and binary checks covered a successful retry.

## Verification

At baseline `b6ab0471` plus product/test diff `47fd2495ce`, the tester exercised all 16 installation cases, four platform/packaging cases, 23 capture distribution cases and the native manager Go suite on Linux. The behavior reviewer also passed an end-to-end Make-to-installed-bundle probe with separately selected engine and manager versions. A nonprivileged fixture using an unreadable regular binary made independent verification exit 2 before any host probe.

## Limits

Local verification exercises the host binaries on Linux. Foreign artifacts receive checksum and platform checks. Runtime identity is deferred to their target host. The manifest establishes integrity relative to the selected release publisher. Hosted publication and tag mutation are outside this delivery.
