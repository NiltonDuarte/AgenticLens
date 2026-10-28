# Iteration 04 Plan

1. Purge the OpenCode user directories for the installed CLI, data, config, cache, and state as requested.
2. Reinstall OpenCode from the official v1 installer.
3. Verify the installed version is below 2 and the v1 CLI initializes a clean database.
4. Update GEN-9 documentation, summary, and PR context with the reset and fresh-install result.

## Acceptance
- OpenCode's prior local data and config directories are removed.
- The reinstalled command reports a v1 release.
- OpenCode starts against a fresh v1 database without the missing `session` table error.
