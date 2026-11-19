# Archived Codex rollout verification

The tests use isolated Codex stores, real SQLite registries and the native CLI. Production source rollouts remain outside the regression fixtures.

- Both native entry points failed on baseline `95c130ed` after a registered rollout moved from `sessions/` to `archived_sessions/`. The corrected tests verify updated facts, repaired registry paths, unchanged attribution and stable session and invocation counts on repeat.
- Negative cases exercise absent, ambiguous, foreign, unreadable, partially parsed and over-budget sources. They require prior accepted facts to survive and independently valid siblings to keep refreshing. One-byte healthy-source retries remain pending without spending source bytes.
- The accepted-root and stale-annotation tests failed on `fdfac845`. The corrected tests reject changed ancestry and superseded evidence while preserving the newer registry checkpoint.
- A separate regression failed on `27835ea8` after the final annotation write. The newer child observation remained available, but its parent's event and delegation still referenced the superseded read. The final test verifies exclusion from the actual publication projection, retains an independent sibling link and preserves refreshed numeric facts.
- Focused race tests passed for the corrected capture adapter and native CLI. The full build on `27835ea8` also passed after restoring the metadata-first budget check for healthy paths.

The final integrated build passed at `c1b87c6e` with `make -j 3` and Vitest concurrency limited to two workers. It ran the full Go race suites, 231 frontend unit tests, 819 frontend integration tests, distribution checks and all four platform builds. The standalone Linux engine reported `c1b87c6e`; its embedded assets match the frontend build. Strict OpenSpec validation passed. Native macOS execution and a live retry of the six archived REV-8 sessions remain outside this verification.
