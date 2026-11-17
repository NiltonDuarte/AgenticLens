# AL-73 candidate review

All four focuses reviewed b6ab0471 plus working diff 47fd2495ce on gpt-6.1-sol/high and gpt-6-luna/medium. Both architecture, behavior and test reviewers approved. The Sol data reviewer approved. The Luna data reviewer requested changes after reproducing an old manifest beside changed build inputs following a failed repackage.

The parent verified the manifest was already stale when the raw binary changed before the packaging attempt. Packaging failed, preserved the previous manifest, and independent verification rejected the directory before probing. Each shipped workflow gates upload on successful packaging and verification. The approved transaction covers publishing the complete manifest. Restoring the entire previous build directory would require retaining inputs from before the build. The observed local limitation is retained for release operators. It does not permit upload or installation of mismatched bytes.

The Luna tests reviewer identified missing direct evidence for unreadable regular files. The parent ran the actual release-verify target as uid 1000 after removing read permissions from a selected regular binary. It returned 2 with the expected missing/unreadable diagnostic. No source changes were required. Final consolidated approval awaits integration.
