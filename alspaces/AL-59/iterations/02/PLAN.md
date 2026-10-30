# AL-59 Iteration 02 Plan: Hermetic OpenCode Test Fixtures

## Goal

Repair the three backend CLI tests exposed by PR CI after selected-source errors became correctly fatal. Keep the AL-59 contract unchanged: if the selected OpenCode source is absent or unreadable, the command must fail with a selected-source error. Tests that intend to exercise later CLI behavior must supply their own valid source fixture and must not rely on a developer or runner's OpenCode database under `HOME`.

Grounding: `../../CONTRACT.md` defines no-fallback failure semantics; iteration 01, `../01/PLAN.md`, calls for deterministic CLI tests using temporary filesystem fixtures and real temporary SQLite where appropriate. The CI log at `/tmp/al59-ci-backend-logs.zip` is UTF-8 text (despite the extension) and shows all three commands fail opening `/home/runner/.local/share/opencode/opencode.db` before reaching their intended assertions.

## Focused fixture changes

Update only the three named subtests in `backend/cmd/agentic-lens/main_test.go`:

1. `TestCLI_Invariants/fatal_on_unresolvable_base_branch`
   - Add `--db-path`, `setupIsolatedEmptyDB(t)` to the `generate` argv, before `--ticket-id`.
   - Set the subprocess `HOME` to a fresh temporary directory (for example `filepath.Join(tempDir, "isolated-home")`, created with `os.MkdirAll`) while preserving other environment entries. Keep the existing temporary config, repo fixture, expected base-branch error, and assertion intact.
   - The valid empty SQLite schema lets generation proceed far enough to assert the missing-base-branch error. Do not alter selected-source failure behavior or weaken the expected assertion.

2. `TestCLI_RootConfigurationDiscovery/generate and scan pick up workspaces and prefixes from root .agenticlens-config.yaml`
   - Add `--db-path`, `setupIsolatedEmptyDB(t)` to `genCmd` argv. The fixture remains independent of root-config discovery: the test config intentionally has no telemetry path, so explicit CLI path selection supplies the hermetic source.
   - Give only `genCmd` an isolated temporary `HOME`, preserving other environment entries. `scanCmd` does not invoke telemetry source ingestion and needs no database flag.
   - Preserve the assertions for generated ticket type and root-config workspace discovery.

3. `TestCLI_RootConfigurationDiscovery/root config takes precedence over legacy config folder in CLI`
   - Add `--db-path`, `setupIsolatedEmptyDB(t)` to `genCmd` argv and isolate `genCmd`'s `HOME` in the same way.
   - Preserve both competing config fixtures and the assertions that root config wins for ticket type and workspace destination.

Use a small test-local helper only if needed to set subprocess `HOME` while retaining `os.Environ()`; avoid process-global `os.Setenv` because the tests are parallel. The HOME fixture must be empty of OpenCode files. The explicit database path must point inside `t.TempDir()` and be created by `setupIsolatedEmptyDB(t)`.

## Acceptance checks

- Each named test invokes `generate` with an explicit `--db-path` pointing at its isolated valid SQLite fixture.
- Each affected `generate` subprocess runs with an isolated `HOME`, so neither the current developer's nor CI runner's `~/.local/share/opencode/opencode.db` can satisfy the command accidentally.
- Existing assertions still require the intended base-branch error and root-config precedence/discovery behavior.
- No fallback is introduced, and no test is changed to accept a missing selected source.
- Changes remain limited to the three test invocations and any narrowly scoped test helper in `backend/cmd/agentic-lens/main_test.go`.

## Verification commands

Run focused cases from `backend/`:

- `go test ./cmd/agentic-lens -run '^(TestCLI_Invariants/fatal_on_unresolvable_base_branch|TestCLI_RootConfigurationDiscovery/(generate_and_scan_pick_up_workspaces_and_prefixes_from_root_\.agenticlens-config\.yaml|root_config_takes_precedence_over_legacy_config_folder_in_CLI))$' -count=1`
- `go test ./cmd/agentic-lens -run '^(TestCLI_Invariants|TestCLI_RootConfigurationDiscovery)$' -count=1`

Then run the backend suite and repository checks:

- From `backend/`: `go test ./...`
- From repository root: `make`

The focused and full test runs should pass without a real OpenCode installation or a database anywhere under the invoking user's home. The full `make` command is the repository delivery verification required by the contract; report any unrelated environment-dependent failure separately.
