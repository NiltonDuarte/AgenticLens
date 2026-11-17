# AL-73 technical design

## Scope and baseline

At remote main `b6ab0471`, the native manager's `downloadVerified` checks a single selected artifact against its exact release manifest. Host installations additionally validate native identity. The canonical setup downloader verifies the bootstrap manager, but its shell comparisons and self-update argument use hash command output without requiring a successful exit. Two isolated baseline runs with inert bytes reached the native-format rejection after a hash utility printed the expected digest and exited 7, and after a selected manifest record contained a trailing field. Require both successful hashing and valid output at each admission boundary. SHA-256 authenticates bytes relative to the manifest. The publisher and delivery of that manifest remain trusted.

REV-5 D1's legacy success cache and D6's copied wrappers have been retired. The contributor script delegates to the installed skill's downloader. The setup bundle fetches its scripts from the selected tag. Its installer and bootstrap interpolate the repository override without validating its shape. Require the native manager’s owner/repository grammar on remote paths so a URL delimiter cannot consume the exact-tag suffix. The manager and engine may have distinct exact versions. Bootstrap installs each engine before updating the manager and exits on a failing subcommand. Retain this per-component commit model, exact-tag provenance, explicit offline sources and existing destinations and defaults. Keep capture handlers, optional hooks and owner-edited files intact.

For D5, Make already preserves the hash producer's exit status. It sorts directly into the public manifest and independent verification checks only listed records. Use the existing Make inventory as the authority. Validate a staged manifest before atomically replacing SHA256SUMS. Independently check inventory and bytes before host probes, executing only selected host files. A one-entry fixture passed current verification. A hash utility that emitted one entry and exited 7 caused packaging to exit 2 without a final manifest in a fresh directory. See the [acceptance scenarios](specs/release-artifact-verification/spec.md) and the [original findings](../../../alspaces/REV-5/iterations/03/DELIVERY_REPORT.md).

## Ownership and approval

The planner owns proposal, specification, design and tasks through scope review. The user explicitly approved this contract and verification plan with “$agentic-lens-dev approved, continue” after reviewing the proposal. The coder owns the Make release recipes, any necessary release inventory helper and the canonical setup downloader and the bundle installer’s remote repository check. The tester owns associated tests in scripts/tests and any necessary manager test cases. Each reviewer owns an assigned evidence file. The parent owns roadmap links, integrated verification, delivery, telemetry and the normal development PR.

## Iteration 01

Exercise actual Make and installer entry points with isolated releases, home directories and benign native fixtures. Review inventory semantics, failure propagation, admission order and receipts against on-disk state. Hosted publication, tag changes, signing and release PR merges are outside scope.
