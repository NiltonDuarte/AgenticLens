# chore(release): reuse Release Please PR versions for candidates (GEN-19)

After stable v2.0.0, candidates still reused v2.0.0-rc labels even though Release Please's open PR already targeted 2.1.0. Read the root version from that PR's manifest at its immutable head, producing v2.1.0-rc.<run>+<sha> for the current release train.

Run automatic candidates after Release Please completes successfully. Capture the pending PR's immutable manifest version once, validate it, and append the existing RC suffix. Skip a missing release PR and stop on API or invalid-version errors. Before publishing, check only whether that exact stable tag already exists. Later main or PR updates do not invalidate the built snapshot; the RC tag is pinned to the captured build commit. Preserve build verification, prerelease assets and retention, with no standalone calculator or added package dependencies. The final tag read and publication remain separate operations.

See [contract](https://github.com/NiltonDuarte/AgenticLens/blob/chore/GEN-19-rc-target-version/alspaces/GEN-19/CONTRACT.md) and [summary](https://github.com/NiltonDuarte/AgenticLens/blob/chore/GEN-19-rc-target-version/alspaces/GEN-19/SUMMARY.md).
