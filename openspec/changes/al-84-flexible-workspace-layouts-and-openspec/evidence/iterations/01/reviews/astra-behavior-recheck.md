# AL-84 iteration 01 behavior correction recheck

APPROVED within the correction scope. No new blocking findings; no fresh discovery credit for verification of supplied findings.

Reviewer: reviewer-behavior, gpt-6-astra/high (requested and runtime identity). Run: astra-behavior-recheck. Published baseline: `2b4118736b5bde24d17cc382da7445c8046ce884`. Corrected candidate: `build/al84-astra-review/CANDIDATE.json`, manifest identity `a10e916ca8c02c8552ff8287690c3e9fc6de2ec9531b2197462e77dc16366f20`. All 538 listed path hashes matched before investigation and again before this report. No other current recheck reports were read.

## Behavioral assessment

- **ASTRA-BEH-01 verified:** `workspace_locations.go` now treats operational read failures as incomplete identity enumeration. A pending sibling cannot establish implicit uniqueness or absence for first generation. Definite validation failures remain excluded individually; healthy explicit selection and partial fleet listing still work. Maintenance retains its recorded-identity decoder, including historical-schema selection, and malformed identity is classified as invalid rather than an operational failure.
- **ASTRA-DATA-01 verified:** traced CLI write-intent loading through `LoadWorkspaceArtifactSelection`, independent directory reference retention, writer admission, descriptor-relative stat, recovery, fact reread and publication. The retained inode is compared before recovery can touch a replacement. Public projection fields do not determine write identity. The resolver closes read leases before returning the independent handle; loader errors close retained handles; successful CLI return closes its selection. CLI error exits release descriptors through process termination. Stable publication while the selection remains open passed, refuting a retained-read-lock writer deadlock on this path.
- **ASTRA-ARCH-01 verified:** discovery queue admission and visitation measure physical relative depth. A shallow alias cannot advertise an excluded deeper home; independently configured coverage restores inclusion. Per-root traversal remains independent, and global physical-location deduplication preserves union semantics.

Inspected unchanged read/publication dependencies and CLI recovery ordering. Explicit recovery before read remains necessary for interrupted selected artifact sets; subsequent publication recovery is behind physical admission. Raw facts are reread under ownership and preserved while computed exports are derived. The new selection is transient, creates no persisted identity, and does not add private-state requirements to ordinary completed-ledger reads.

## Verification and limits

Executed in `backend`: `go test ./internal/adapters -run 'TestWorkspace(PendingSibling|AliasUses|MaintenanceInvalid|ArtifactSelection)|TestPublishWorkspaceNested' -count=1 -v` — PASS, 0.273s, including stable, replacement and pending-replacement subtests.

Reused matching tester source receipt and inspected affected CLI race, broad capture/CLI rerun and recovery-order mutation logs under `build/al84-astra-review/tester`. CLI affected race passed (17.491s); broad capture and CLI passed (102.221s/232.015s). Moving admission below recovery failed because the replacement pending marker disappeared, supporting the test's recovery-order oracle.

The earlier broad callback deadline/SQLite rollback failure remains an unexplained reliability observation. The correction does not change callback persistence or its handler deadline; six targeted passes and a broad pass reduce reproducibility concerns but do not prove unrelatedness or a fix. No new evidence establishes it as a correction blocker. Parent must retain the observation and complete integrated make/consolidation. This review does not certify native macOS behavior, exhaustive descriptor leaks, physical crashes, frontend behavior, or whole-repository delivery.

PR https://github.com/NiltonDuarte/AgenticLens/pull/144, published head `2b4118736b5bde24d17cc382da7445c8046ce884`, corrected manifest `a10e916ca8c02c8552ff8287690c3e9fc6de2ec9531b2197462e77dc16366f20`, focused checks PASS, scoped behavior review APPROVED; integrated gate pending parent.
