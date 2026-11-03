# chore(release): align RC versions with the upcoming release (GEN-19)

After stable v2.0.0, release candidates still reused v2.0.0-rc labels even though the upcoming release was 2.1.0. Calculate the upcoming version using the pinned Release Please engine and existing release configuration, producing v2.1.0-rc.<run>+<sha> for the current release train.

Skip superseded commits and runs without release-worthy changes; reject invalid or already stable targets. Recheck eligibility before publishing while preserving build metadata, prerelease assets and candidate retention. Stable publishing and historical releases are unchanged. GitHub reads and publication remain separate operations; the final guard cannot make them atomic.

See [contract](CONTRACT.md) and [summary](SUMMARY.md).
