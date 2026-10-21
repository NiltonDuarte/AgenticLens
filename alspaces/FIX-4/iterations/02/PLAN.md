# Iteration 02 Plan: Elimination of Hardcoded Workspaces Fallbacks and Obsolete Directory Logic

## Objective
Eliminate all hardcoded fallback mechanisms that default to the string "workspaces", enforcing that workspace locations are either explicitly supplied through command-line flags or declared in the project configuration file. Purge all remaining references, path traversals, and discovery rules associated with the obsolete and removed ".agenticlens" directory across server startup, repository resolution, configuration loading, and test suites.

## Grounded Context References
- backend/cmd/agentic-lens/main.go:
  - runServer function: CLI flag evaluation, configuration loading, workspaces directory resolution, and server initialization.
  - CLI root execution: list and summary mode workspace resolution and execution flow.
- backend/internal/adapters/repository.go:
  - ResolveWorkspacesDir function: resolution hierarchy across override arguments, volume mounts, configuration lookup, and parent directory traversals.
  - ScanWorkspaces function: target directory validation and ledger scanning.
- backend/internal/adapters/config/loader.go:
  - mergeConfig function: ProjectConfig initialization and field defaults.
  - FindConfigFile function: upward directory traversal candidate evaluation.
  - ValidateConfig function: configuration field completeness enforcement.
- Test suites:
  - backend/internal/adapters/repository_test.go: unit tests for workspace directory resolution.
  - backend/internal/adapters/config/loader_test.go: unit tests for configuration discovery and parsing.
  - backend/cmd/agentic-lens/main_test.go: integration tests for CLI commands and server startup behavior.

## Problem Analysis and Scope Boundaries

### Problem 1: Hardcoded Fallback to "workspaces"
1. In backend/cmd/agentic-lens/main.go (runServer), when the --workspaces flag is omitted, the code inspects projCfg.Workspaces.Dir. If the configured directory is empty, it currently assigns "workspaces" as a fallback. This violates the configuration contract: workspace directory must be explicitly configured or passed via flag. If unconfigured and unsupplied, startup must fail immediately with an actionable error.
2. In backend/internal/adapters/config/loader.go (mergeConfig), domain.WorkspacesConfig.Dir is initialized with the default string "workspaces". As a consequence, even when a configuration file completely omits the workspaces.dir field, projCfg.Workspaces.Dir receives "workspaces", preventing downstream components from distinguishing between an explicit user configuration and an unintended default.
3. In backend/internal/adapters/repository.go (ResolveWorkspacesDir), line 30 initializes wsDir to "workspaces", and line 72 returns wsDir as the final fallback. Furthermore, line 23 checks an arbitrary container mount at /app/workspaces. When neither an override path nor a configured workspace directory exists, the function must not fabricate a path named "workspaces". It must return an empty string, allowing callers to handle the missing configuration explicitly.

### Problem 2: Obsolete ".agenticlens" Directory Logic
1. In backend/cmd/agentic-lens/main.go (runServer), the path resolution checks if the base directory of the configuration file parent equals ".agenticlens" and steps up one additional directory level. The ".agenticlens" directory has been decommissioned and removed from the platform. Configuration files reside exclusively at project root (.agenticlens-config.yaml, .agenticlens-config.yml, .agenticlens-config.json). The directory containing the configuration file is already the project root.
2. In backend/internal/adapters/repository.go (ResolveWorkspacesDir), line 38 performs the identical check against ".agenticlens". This must be removed so the base directory is derived directly from filepath.Dir of the configuration file path.
3. In backend/internal/adapters/config/loader.go (FindConfigFile), lines 422 to 424 search for config.yaml, config.yml, and config.json nested inside a ".agenticlens" subfolder. Because this directory is invalid and no longer part of the platform, these candidates must be removed from the candidate evaluation list, and the associated docstrings must be updated.
4. In test suites (repository_test.go, loader_test.go, main_test.go), several test cases construct ".agenticlens" directories to test legacy discovery. These tests must be updated to align with the platform contract, using root configuration files instead.

## Phased Implementation Tasks

### Phase 1: Configuration Loader Cleanup (backend/internal/adapters/config/loader.go)
1. In mergeConfig:
   - Change the initialization of Workspaces.Dir in domain.WorkspacesConfig from "workspaces" to an empty string.
   - Retain assignment of cfg.Workspaces.Dir from dto.Workspaces.Dir only when non-nil and trimmed non-empty.
2. In FindConfigFile:
   - Remove the three legacy candidates searching within the ".agenticlens" subfolder.
   - Restrict candidate lookup at each directory level exclusively to root configuration files: .agenticlens-config.yaml, .agenticlens-config.yml, and .agenticlens-config.json.
   - Update documentation comments to remove all references to legacy ".agenticlens" subfolder resolution.

### Phase 2: Repository Adapter Workspace Resolution (backend/internal/adapters/repository.go)
1. In ResolveWorkspacesDir:
   - If overridePath is non-empty, return overridePath immediately.
   - Remove the hardcoded container path probe for /app/workspaces.
   - Query the centralized configuration loader via config.Load.
   - Extract the configured workspace directory string from cfg.Workspaces.Dir.
   - If loading configuration fails or if cfg.Workspaces.Dir is empty:
     - Return an empty string immediately. Do not fall back to "workspaces".
   - If cfg.Workspaces.Dir is non-empty:
     - If the configured path is already absolute, return it directly.
     - Locate the configuration file path via config.FindConfigFile.
     - If a configuration file is found, set baseDir to filepath.Dir(cfgPath) directly. Remove the obsolete conditional check inspecting whether filepath.Base(parentDir) equals ".agenticlens".
     - If no configuration file path is found, fall back to current working directory as baseDir.
     - Construct candidate path by joining baseDir and the configured directory.
     - Check existence on disk. If directory exists, return the candidate.
     - Check current working directory and parent directory traversals using only the configured directory name.
     - If not found on disk during traversal, return the resolved candidate path under baseDir.
2. In ScanWorkspaces:
   - When targetDir resolved by ResolveWorkspacesDir is empty, ensure the function records an explicit scan error indicating that the workspace directory is unconfigured or not found, rather than attempting to stat an empty path.

### Phase 3: Server Startup Enforcement (backend/cmd/agentic-lens/main.go)
1. In runServer:
   - Evaluate rawFlag from *workspacesFlag.
   - If rawFlag is provided and non-empty:
     - Assign workspacesDir to rawFlag.
     - Synchronize projCfg.Workspaces.Dir to workspacesDir.
   - If rawFlag is empty:
     - Retrieve configuredDir from strings.TrimSpace(projCfg.Workspaces.Dir).
     - If configuredDir is empty:
       - Print the following error message to os.Stderr: "Error: workspaces.dir is empty or unconfigured: configure 'workspaces.dir' in .agenticlens-config.yaml or provide --workspaces flag".
       - Terminate process with os.Exit(1).
     - If configuredDir is absolute:
       - Assign workspacesDir to configuredDir.
     - If configuredDir is relative:
       - Locate configuration file path via adaptersconfig.FindConfigFile.
       - If configuration file is found, derive baseDir directly as filepath.Dir(cfgPath). Remove the obsolete check inspecting whether filepath.Base(parentDir) equals ".agenticlens".
       - If configuration file is not found, derive baseDir from os.Getwd.
       - Join baseDir with configuredDir and assign to workspacesDir.
2. In CLI root execution (main function):
   - When *workspacesFlag is empty:
     - Check projCfg.Workspaces.Dir. If empty, print error "Error: workspaces.dir is empty or unconfigured: configure 'workspaces.dir' in .agenticlens-config.yaml or provide --workspaces flag" and exit with code 1.

### Phase 4: Test Suite Alignment and Regression Verification
1. In backend/internal/adapters/repository_test.go:
   - Update TestResolveWorkspacesDir:
     - Remove tests asserting resolution from legacy ".agenticlens" subdirectories or replace them with root .agenticlens-config.yaml configurations.
     - Add test asserting that when overridePath is empty and no configuration file exists, ResolveWorkspacesDir returns an empty string without falling back to "workspaces".
     - Add test asserting that when configuration has an empty workspaces.dir, ResolveWorkspacesDir returns an empty string.
     - Verify tests with root YAML, YML, JSON, and nested execution pass.
2. In backend/internal/adapters/config/loader_test.go:
   - Update tests asserting that unspecified operational sections retain a default workspaces.dir of "workspaces" so that they now assert Workspaces.Dir is empty.
   - Migrate tests that set up legacy ".agenticlens" directories to root configuration files, or update tests that assert legacy files are no longer recognized as configuration candidates.
3. In backend/cmd/agentic-lens/main_test.go:
   - Add integration test for server startup when configuration file exists but workspaces.dir is omitted, asserting the process exits with status 1 and outputs the exact unconfigured error message.
   - Add test verifying server startup succeeds when --workspaces flag is passed even if workspaces.dir is empty in configuration.
   - Verify existing server startup tests pass using root configuration and configured workspace directory (alspaces).

### Phase 5: Verification and Quality Gate
1. Execute backend unit test suite: go test ./backend/internal/...
2. Execute command line integration test suite: go test ./backend/cmd/agentic-lens/...
3. Verify zero occurrences of the string ".agenticlens" used as a directory lookup in active code paths.
4. Verify that running the server without flags in a repository configured with alspaces launches correctly without references to "workspaces".
