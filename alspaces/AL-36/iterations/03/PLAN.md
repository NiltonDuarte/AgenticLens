# Implementation Plan: AL-36 Iteration 03 - Peer Ticket Prefixes, Comma-Separated Inputs, Zero Artifact Defaults, and Complete Removal of Project Fallback Defaults

## 1. Goal and Iteration 03 Scope

This delta plan addresses user steering directions for Ticket AL-36 across Go domain modeling, configuration adapters, CLI initialization ergonomics, repository artifact scanning, and test suite fixtures:

1. Comma-Separated Peer Ticket Prefixes and Source Folders:
   - Support comma-separated strings for ticket prefixes via `--ticket-prefixes` (with `--ticket-prefix` maintained as a backward-compatible alias) to accept any number of prefixes in a single argument.
   - Support comma-separated strings for source folders via `--src-folders`.
   - Completely eliminate the concept of a primary prefix across CLI flags, domain logic, setup skills, documentation, help text, and error messages. Treat all user-specified prefixes as equal peers without injecting unrequested prefixes.

2. Zero Workspace Artifact Defaults:
   - Treat workspace documentation artifacts as strictly opt-in. Neither CONTRACT.md, SUMMARY.md, nor PULL_REQUEST.md may serve as implicit platform defaults.
   - Delete DefaultWorkspaceArtifacts from domain and adapters entirely.
   - Ensure that when workspaces.artifacts is empty or omitted from configuration, writer.go serializes an empty artifact list, loader.go retains an empty slice, and repository.go returns an empty item slice.

3. Complete Removal of Hardcoded Project Defaults from loader.go:
   - Remove the inlined default ProjectConfig struct in mergeConfig within loader.go, eliminating hardcoded fallbacks for project name, ticket prefixes, and source folders.
   - Require explicit configuration loaded from a user-provided manifest; return a typed validation error if project_name, prefixes, or src_folders are missing or empty.
   - Retain standard technical defaults solely for non-project operational thresholds such as SQLite timeouts and live monitoring polling intervals.

4. Complete Removal of Fake Fallbacks from writer.go:
   - Ensure writer.go requires valid input parameters from callers rather than substituting synthetic project defaults.
   - Eliminate all synthetic fallbacks in GenerateConfigYAML and GenerateDefaultYAML.

5. Test Suite Realignment and Adversarial Fixture Updates:
   - Retain testProjectConfig strictly as a local fixture generator within test files.
   - Update test suites across domain, adapters, and CLI entrypoint to assert that configuration missing project_name, prefixes, or src_folders fails with explicit validation errors.

---

## 2. Grounded Reference Architecture & File Paths

### Go Functional Core (backend/internal/domain/)
- backend/internal/domain/config.go:
  - ProjectConfig, TicketConfig, WorkspacesConfig, WorkspaceArtifactConfig domain models.
  - Verification that no DefaultWorkspaceArtifacts or DefaultProjectConfig functions exist in production domain code.

### Go Imperative Boundary Shell: Configuration Adapters (backend/internal/adapters/config/)
- backend/internal/adapters/config/loader.go:
  - Delete DefaultWorkspaceArtifacts function.
  - Define ErrInvalidConfig sentinel error and ConfigValidationError struct detailing missing required fields.
  - In mergeConfig, eliminate the inlined fallback struct containing AgenticLens, AL, GEN, PO, and web/backend.
  - Enforce strict validation: if project_name is blank, ticket.prefixes is empty, or workspaces.src_folders is empty, return ConfigValidationError.
  - Preserve empty slice for Workspaces.Artifacts when not configured by user.
- backend/internal/adapters/config/writer.go:
  - Remove all references to DefaultWorkspaceArtifacts.
  - In GenerateConfigYAML: remove fallback substitution of AgenticLens for ProjectName; require non-empty project_name, prefixes, and src_folders; serialize empty artifact slice without injecting unconfigured artifacts.
  - In GenerateDefaultYAML: reject calls with empty project name, empty source folders, or empty prefixes with an explicit error instead of falling back to synthetic defaults; do not populate default artifacts.
- backend/internal/adapters/config/loader_test.go:
  - Update testProjectConfig to set Artifacts to an empty slice.
  - Update tests that previously relied on mergeConfig injecting default project names, prefixes, or source folders to supply explicit valid fixtures.
  - Add dedicated test cases verifying that configuration files missing project_name, prefixes, or src_folders return ConfigValidationError.
- backend/internal/adapters/config/writer_test.go:
  - Update tests to reflect removal of DefaultWorkspaceArtifacts and synthetic fallbacks.
  - Add test verifying that GenerateDefaultYAML returns an error when required parameters are empty.
  - Add test verifying that an empty artifacts slice serializes without injecting default artifacts.

### Go Imperative Boundary Shell: Repository & HTTP API (backend/internal/adapters/)
- backend/internal/adapters/repository.go:
  - In ReadTicketArtifacts, remove fallback to DefaultWorkspaceArtifacts.
  - When configuredArtifacts is empty and no artifacts are defined in project configuration, return a TicketArtifacts struct with an empty Items slice.
- backend/internal/adapters/http/server.go:
  - In NewServerWithConfig, remove fallback to DefaultWorkspaceArtifacts.
  - Pass configured artifact slice directly to WorkspaceDetailHandler.
- backend/internal/adapters/http/server_test.go:
  - Update test fixtures to use testProjectConfig with empty or explicitly configured artifacts.
- backend/internal/adapters/repository_test.go:
  - Verify that ReadTicketArtifacts returns empty Items when no artifacts are configured.

### Go CLI Application Layer (backend/cmd/agentic-lens/)
- backend/cmd/agentic-lens/main.go:
  - In runInit:
    - Register `--ticket-prefixes` flag alongside `--ticket-prefix` alias, documented as a comma-separated list of peer ticket prefix identifiers.
    - Parse comma-separated prefixes, split on commas, trim whitespace, and discard empty tokens.
    - If no non-empty prefixes are provided, fail with exit code 1 and output an error stating that ticket prefixes are required.
    - Construct TicketPrefix entries for all user-provided prefixes as equal peers; do NOT automatically inject GEN or PO prefixes unless explicitly provided by the user.
    - Parse comma-separated source folders from `--src-folders`, split on commas, trim whitespace, and discard empty tokens.
    - If no non-empty folders remain, fail with exit code 1.
    - Remove all usage and help text mentioning a primary ticket prefix.
    - Pass empty artifact slice into configuration generation unless custom artifacts are explicitly specified.
- backend/cmd/agentic-lens/main_test.go:
  - Update init tests to test comma-separated `--ticket-prefixes` and `--ticket-prefix` flags with peer prefixes.
  - Add tests verifying that init fails when `--ticket-prefixes` or `--src-folders` contain only commas or whitespace.
  - Add test verifying that generated YAML contains only the user-specified prefixes and zero default artifacts.

### Setup Skill Manifest & Downloader Automation (.opencode/skills/agentic-lens-setup/ and scripts/)
- .opencode/skills/agentic-lens-setup/SKILL.md:
  - Update documentation to describe comma-separated peer ticket prefixes and source folders.
  - Remove all mentions of primary and secondary prefixes.
  - Explain that documentation artifacts are optional and not populated by default.
- scripts/test-setup.sh:
  - Update init invocations to pass peer prefixes using comma-separated notation.

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Zero Artifact Defaults and Removal of DefaultWorkspaceArtifacts
- Task 1.1: Delete DefaultWorkspaceArtifacts from backend/internal/adapters/config/loader.go and ensure no occurrence exists in backend/internal/domain/config.go.
- Task 1.2: In backend/internal/adapters/config/writer.go, remove DefaultWorkspaceArtifacts invocations. When cfg.Workspaces.Artifacts is empty, serialize workspaces.artifacts as an empty YAML list without injecting synthetic artifact items.
- Task 1.3: In backend/internal/adapters/config/writer.go inside GenerateDefaultYAML, remove default artifact population. If customArtifacts is omitted or empty, assign an empty slice to cfg.Workspaces.Artifacts.
- Task 1.4: In backend/internal/adapters/repository.go inside ReadTicketArtifacts, remove the fallback that called DefaultWorkspaceArtifacts. If neither configuredArtifacts nor project config specifies artifacts, set configs to an empty slice and return empty Items.
- Task 1.5: In backend/internal/adapters/http/server.go inside NewServerWithConfig, remove the fallback that populated DefaultWorkspaceArtifacts when cfg.Workspaces.Artifacts was empty.

### Phase 2: Elimination of Fake Project Defaults in Configuration Loader and Strict Validation
- Task 2.1: In backend/internal/adapters/config/loader.go, define ErrInvalidConfig sentinel error and ConfigValidationError struct implementing the error interface, formatting a description of all missing required fields.
- Task 2.2: In backend/internal/adapters/config/loader.go inside mergeConfig, remove lines 171-229 that initialized an inlined ProjectConfig with hardcoded values for AgenticLens, AL, GEN, PO, and web/backend.
- Task 2.3: In mergeConfig, construct ProjectConfig using zero values for project-specific fields, applying defaults only to non-project operational infrastructure parameters (database timeouts, live anomaly thresholds, role mappings, and role order).
- Task 2.4: In mergeConfig or LoadFromFile, validate that ProjectName is non-empty, Ticket.Prefixes has at least one valid non-empty prefix, and Workspaces.SrcFolders has at least one valid non-empty folder. If any are missing, return a ConfigValidationError.
- Task 2.5: In backend/internal/adapters/config/loader.go inside LoadFromFile, ensure validation errors are returned directly to the caller, preventing invalid configurations from being loaded into memory.

### Phase 3: Elimination of Fake Fallbacks in Configuration Writer
- Task 3.1: In backend/internal/adapters/config/writer.go inside GenerateConfigYAML, remove the fallback that set projName to AgenticLens when cfg.ProjectName was empty. Validate that ProjectName, Ticket.Prefixes, and Workspaces.SrcFolders are populated; return an error if any required field is empty.
- Task 3.2: In backend/internal/adapters/config/writer.go inside GenerateDefaultYAML, validate that projectName is non-empty, srcFolders contains at least one non-empty string, and prefixes contains at least one non-empty prefix. Return an error immediately if any mandatory parameter is missing instead of substituting synthetic values like AgenticLens, src, or AL/GEN/PO.
- Task 3.3: In GenerateDefaultYAML, remove hardcoded role mapping overrides or ensure they remain pure defaults without fabricating project identity.

### Phase 4: Peer Ticket Prefixes and Comma-Separated Flag Parsing in CLI Init
- Task 4.1: In backend/cmd/agentic-lens/main.go inside runInit, define ticketPrefixesFlag (`--ticket-prefixes`) and ticketPrefixFlag (`--ticket-prefix`) as peer aliases. Document both as a comma-separated list of ticket prefix identifiers.
- Task 4.2: In runInit, read the provided flag value, split on commas, trim whitespace from each token, and filter out empty strings. If no non-empty prefixes remain, print an error message stating that `--ticket-prefixes` (or `--ticket-prefix`) is required and exit with status 1.
- Task 4.3: In runInit, construct a domain.TicketPrefix for each parsed prefix token. Assign a neutral description and omit hardcoded conventional commit types unless explicitly known. Do NOT append GEN or PO unless the user explicitly included them in the flag argument.
- Task 4.4: In runInit, parse `--src-folders` by splitting on commas, trimming each token, and filtering out empty strings. If no non-empty folders remain, print an error message stating that `--src-folders` is required and exit with status 1.
- Task 4.5: In runInit, remove all log, help, and error strings that reference a primary ticket prefix. Update usage documentation to describe peer prefixes.

### Phase 5: Repository Artifact Reader and HTTP Server Alignment
- Task 5.1: In backend/internal/adapters/repository.go, verify that ReadTicketArtifacts handles an empty slice of artifact configurations cleanly, returning a TicketArtifacts instance where Items is an empty slice and legacy pointer fields are nil.
- Task 5.2: In backend/internal/adapters/http/server.go, verify that WorkspaceDetailHandler receives an empty slice when no artifacts are configured, and that the GET /api/workspaces/{ticketId} endpoint returns an artifacts object with an empty items array.

### Phase 6: Skill Documentation and Shell Automation Alignment
- Task 6.1: In .opencode/skills/agentic-lens-setup/SKILL.md, update Step 2 interview and execution examples to demonstrate comma-separated ticket prefixes using `--ticket-prefixes` and source folders using `--src-folders`.
- Task 6.2: In SKILL.md, remove all references to primary and secondary prefixes. Emphasize that all prefixes specified are equal peers, and clarify that workspace documentation artifacts are not configured by default.
- Task 6.3: In scripts/test-setup.sh, update invocations of agentic-lens init to supply comma-separated prefixes and source folders matching the updated CLI contract.

### Phase 7: Test Suite Realignment and Adversarial Test Expansion
- Task 7.1: In backend/internal/adapters/config/loader_test.go, update testProjectConfig to set Artifacts to an empty slice.
- Task 7.2: In loader_test.go, update tests that loaded minimal YAML snippets expecting synthetic defaults: ensure tests testing general features provide project_name, prefixes, and src_folders so they pass required field validation.
- Task 7.3: In loader_test.go, author new adversarial unit tests verifying that loading YAML missing project_name returns a ConfigValidationError, missing prefixes returns a ConfigValidationError, and missing src_folders returns a ConfigValidationError.
- Task 7.4: In backend/internal/adapters/config/writer_test.go, author tests verifying that GenerateDefaultYAML fails when project name, source folders, or prefixes are missing, and that GenerateConfigYAML serializes zero artifacts when Artifacts is empty.
- Task 7.5: In backend/cmd/agentic-lens/main_test.go, author tests verifying that runInit parses comma-separated prefixes into peer entries without injecting GEN or PO, parses comma-separated source folders, and rejects empty comma strings with exit code 1.

---

## 4. Verification & Testing Strategy for Tester

The tester subagent must verify all boundary behaviors and negative test conditions:

1. CLI Flag Parsing and Peer Prefix Tests:
   - Run `agentic-lens init --project-name Test --src-folders web,backend --ticket-prefixes AL,GEN,PO`: assert exit code 0, and verify generated YAML contains AL, GEN, and PO as peers with no other prefixes.
   - Run `agentic-lens init --project-name Test --src-folders src --ticket-prefix CUSTOM`: assert exit code 0, and verify generated YAML contains only CUSTOM (ensure GEN and PO are NOT present).
   - Run `agentic-lens init --project-name Test --src-folders web,backend --ticket-prefixes ",  , "`: assert non-zero exit code and error on stderr.
   - Run `agentic-lens init --project-name Test --src-folders ",  , " --ticket-prefixes AL`: assert non-zero exit code and error on stderr.
   - Verify that CLI help output contains zero occurrences of the word primary in relation to ticket prefixes.

2. Zero Default Artifacts Tests:
   - Inspect YAML generated by `agentic-lens init`: assert that workspaces.artifacts is an empty list or absent.
   - Stat a ticket workspace directory: call ReadTicketArtifacts with empty configuration; assert that returned Items slice is empty and legacy contract/summary pointers are nil.
   - Verify that DefaultWorkspaceArtifacts does not exist in any compiled package.

3. Strict Configuration Loading Validation Tests:
   - Attempt to load YAML with missing `project_name`: assert that Load returns a validation error indicating missing project_name.
   - Attempt to load YAML with missing `ticket.prefixes`: assert that Load returns a validation error indicating missing ticket prefixes.
   - Attempt to load YAML with missing `workspaces.src_folders`: assert that Load returns a validation error indicating missing source folders.
   - Verify that when a valid configuration file is loaded without artifacts, Workspaces.Artifacts is an empty slice.

4. Writer Validation Tests:
   - Call GenerateDefaultYAML with empty project name: assert error returned.
   - Call GenerateDefaultYAML with empty prefixes: assert error returned.
   - Call GenerateDefaultYAML with empty source folders: assert error returned.

5. Regression and Quality Gate Verification:
   - Run `make check`, `make test`, and `make build` to confirm that all unit, adapter, and CLI test suites pass with zero warnings or errors.
   - Scan all modified files to ensure zero emojis are present.

---

## 5. Audit & Review Criteria for Reviewer

The reviewer subagent must audit the implementation against two strict axes:

- Axis 1: Spec and User Steering Alignment
  - Are ticket prefixes parsed from comma-separated strings via `--ticket-prefixes` and `--ticket-prefix`?
  - Are source folders parsed from comma-separated strings via `--src-folders`?
  - Is the concept of a primary prefix completely removed from flags, domain models, help text, error messages, and skills?
  - Are GEN and PO excluded from generated configurations unless explicitly provided by the user?
  - Is DefaultWorkspaceArtifacts completely deleted from domain and adapters?
  - Do writer, loader, and repository handle empty artifact configurations without injecting CONTRACT.md, SUMMARY.md, or PULL_REQUEST.md?
  - Are hardcoded project defaults removed from loader.go, requiring valid project_name, prefixes, and src_folders in loaded configuration files?
  - Are fake fallbacks removed from writer.go?
  - Are testProjectConfig helpers confined strictly to test files?

- Axis 2: Engineering Standards and Invariants
  - Does the functional core remain pure with zero I/O?
  - Are boundary validations strictly encapsulated in internal/adapters/?
  - Are Go and TypeScript 100% free of any or interface{} types?
  - Are all code files, comments, docstrings, shell scripts, CLI outputs, and workspace artifacts completely free of emojis?
  - Do all tests pass deterministically without race conditions or flaky behaviors?
