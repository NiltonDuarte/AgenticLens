# chore(release): reuse Release Please PR versions for candidates (GEN-19)

After stable v2.0.0, candidates still reused v2.0.0-rc labels even though Release Please's open PR already targeted 2.1.0. Read the root version from that PR's manifest at its immutable head, producing v2.1.0-rc.<run>+<sha> for the current release train.

Run automatic candidates after Release Please completes successfully. Check the candidate commit and generated PR provenance; skip missing/stale PRs and superseded commits, reject invalid/already stable targets, and recheck eligibility before publishing. Preserve build metadata, prerelease assets and retention. Remove the duplicate version engine and dependency tree. GitHub reads and publication remain separate operations; the final guard cannot make them atomic.

See [contract](CONTRACT.md) and [summary](SUMMARY.md).
