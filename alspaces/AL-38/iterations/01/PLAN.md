# Implementation Plan: AL-38 Migrate Configuration Manifest to Root .agenticlens-config.yaml

## 1. Goal and Iteration Scope

The goal of iteration 01 is to migrate the AgenticLens configuration manifest from .agenticlens/config.yaml to root .agenticlens-config.yaml, update the Go loader adapter to prioritize root paths with legacy fallback, ensure workspace directory resolution works accurately with root config paths, migrate documentation and agent instructions, and provide comprehensive unit and integration test coverage.

## 2. Grounded Reference Architecture & File Paths

The implementation touches the following repository components:

- Boundary Configuration Adapter:
  - backend/internal/adapters/config/loader.go: Updates FindConfigFile candidate paths to prioritize root files (.agenticlens-config.yaml, .agenticlens-config.yml, .agenticlens-config.json) over legacy subfolder files (.agenticlens/config.yaml, .agenticlens/config.yml, .agenticlens/config.json).
  - backend/internal/adapters/config/loader_test.go: Adds unit tests for root candidate discovery, precedence hierarchies, legacy fallback, and upward parent traversal.
- Repository Scanner Adapter:
  - backend/internal/adapters/repository.go: Refactors ResolveWorkspacesDir base directory calculation to support both root configuration files and legacy subfolder configuration files.
  - backend/internal/adapters/repository_test.go: Adds regression test asserting workspace resolution when root .agenticlens-config.yaml is present.
- CLI Entrypoint and Tests:
  - backend/cmd/agentic-lens/main_test.go: Adds CLI integration tests validating that agentic-lens scan and agentic-lens generate correctly pick up root .agenticlens-config.yaml.
- Repository Manifest Migration:
  - .agenticlens-config.yaml: New configuration manifest at repository root with all original settings, comments, role mappings, and operational thresholds intact.
  - .agenticlens/config.yaml and .agenticlens/: Removed from repository tree.
- Operational Guidelines & Agent Persona Definitions:
  - AGENTS.md: Update Section 1 Rule 1 configuration path reference.
  - .opencode/agent/orchestrator.md: Update Step 1 kickoff configuration reference.
  - .opencode/agent/planner.md: Update Core References configuration path.
  - .opencode/agent/coder.md: Update Mandatory Guidelines configuration path.
  - .opencode/agent/tester.md: Update Core References configuration path.
  - .opencode/agent/reviewer.md: Update Audit Scope configuration path.
  - .opencode/skills/agentic-lens-dev/SKILL.md: Update Configuration & Parametrization section and subagent prompt templates.
  - .opencode/skills/agentic-lens-po/SKILL.md: Update Configuration & Parametrization section.
  - .opencode/skills/agentic-lens-review/SKILL.md: Update Project Parametrization reference.

## 3. Phased Step-by-Step Implementation Tasks for Coder

### Phase 1: Boundary Adapter Refactoring (backend/internal/adapters/config/loader.go)
- Task 1.1: Update candidate search order in FindConfigFile:
  - In FindConfigFile, replace the candidate loop with an ordered candidate list evaluated at each directory level:
    Item 1: filepath.Join(curr, ".agenticlens-config.yaml")
    Item 2: filepath.Join(curr, ".agenticlens-config.yml")
    Item 3: filepath.Join(curr, ".agenticlens-config.json")
    Item 4: filepath.Join(curr, ".agenticlens", "config.yaml")
    Item 5: filepath.Join(curr, ".agenticlens", "config.yml")
    Item 6: filepath.Join(curr, ".agenticlens", "config.json")
  - For each candidate, check os.Stat and verify !fi.IsDir(). Return the first matching path immediately.
  - Ensure parent directory upward traversal continues seamlessly when none match at the current level.
  - Update docstring on FindConfigFile documenting the prioritized root discovery and legacy fallback behavior.

### Phase 2: Workspace Directory Resolution Fix (backend/internal/adapters/repository.go)
- Task 2.1: Refactor baseDir computation in ResolveWorkspacesDir:
  - Inspect cfgPath returned by config.FindConfigFile("").
  - If cfgPath is non-empty, determine the project base directory:
    Let parentDir = filepath.Dir(cfgPath).
    If filepath.Base(parentDir) == ".agenticlens", baseDir = filepath.Dir(parentDir).
    Otherwise, baseDir = parentDir.
  - Verify that filepath.Join(baseDir, wsDir) correctly locates the workspaces directory when .agenticlens-config.yaml is at the repository root.

### Phase 3: Repository Manifest Migration
- Task 3.1: Move .agenticlens/config.yaml to .agenticlens-config.yaml at repository root:
  - Preserve all content, section headers, explanatory comments, ticket prefixes, database pragmas, live anomaly thresholds, code metrics globs, and role mappings.
- Task 3.2: Remove empty directory .agenticlens/:
  - Delete .agenticlens/ directory and verify no orphaned files remain.

### Phase 4: Operational Guidelines, Agent Instructions, and Skill Manifests
- Task 4.1: Update AGENTS.md:
  - Update Section 1 Rule 1 to cite .agenticlens-config.yaml.
- Task 4.2: Update agent persona manifests:
  - .opencode/agent/orchestrator.md: Change Step 1 configuration path to .agenticlens-config.yaml.
  - .opencode/agent/planner.md: Change Core References to cite .agenticlens-config.yaml.
  - .opencode/agent/coder.md: Change Mandatory Guidelines to cite .agenticlens-config.yaml.
  - .opencode/agent/tester.md: Change Core References to cite .agenticlens-config.yaml.
  - .opencode/agent/reviewer.md: Change audit references to .agenticlens-config.yaml.
- Task 4.3: Update skill manifests:
  - .opencode/skills/agentic-lens-dev/SKILL.md: Update heading, startup inspection instructions, and prompt templates in Phase 1 to cite .agenticlens-config.yaml.
  - .opencode/skills/agentic-lens-po/SKILL.md: Update heading and startup inspection instructions to cite .agenticlens-config.yaml.
  - .opencode/skills/agentic-lens-review/SKILL.md: Update project parametrization reference to cite .agenticlens-config.yaml.

### Phase 5: Verification & Quality Gates
- Task 5.1: Run make check and make test-backend to confirm all checks compile and pass cleanly.
- Task 5.2: Verify zero any and zero emojis across all touched files.

## 4. Adversarial Testing Strategy for Tester

### Unit and Boundary Tests (backend/internal/adapters/config/loader_test.go)
- Test Case 1: Root .agenticlens-config.yaml discovery and parsing:
  - Write .agenticlens-config.yaml to temp directory root. Call config.Load(tempDir) and assert that all fields across project_name, ticket, workspaces, database, live, code_metrics, role_mappings, and role_order are accurately parsed.
- Test Case 2: Root format precedence (YAML over YML over JSON):
  - In a directory containing root .agenticlens-config.yaml, .agenticlens-config.yml, and .agenticlens-config.json with distinct project names, verify config.Load selects .agenticlens-config.yaml.
  - In a directory containing root .agenticlens-config.yml and .agenticlens-config.json, verify config.Load selects .agenticlens-config.yml.
  - In a directory containing only root .agenticlens-config.json, verify config.Load selects .agenticlens-config.json.
- Test Case 3: Root precedence over legacy subfolder files:
  - In a directory containing both root .agenticlens-config.yaml and legacy .agenticlens/config.yaml with distinct project names, verify config.Load selects root .agenticlens-config.yaml.
  - In a directory containing root .agenticlens-config.json and legacy .agenticlens/config.yaml, verify config.Load selects root .agenticlens-config.json over legacy YAML.
- Test Case 4: Legacy subfolder fallback:
  - In a directory containing only legacy .agenticlens/config.yaml, verify config.Load successfully discovers and loads it.
  - In a directory containing only legacy .agenticlens/config.json, verify config.Load successfully discovers and loads it.
- Test Case 5: Missing configuration fallback:
  - In an empty directory, verify config.Load returns domain.DefaultProjectConfig() and nil error.
- Test Case 6: Upward parent directory traversal with root config:
  - Place .agenticlens-config.yaml in a parent directory. Invoke config.FindConfigFile from a deeply nested child folder (sub1/sub2/sub3) and assert that the parent root config path is returned.

### Repository Workspace Resolution Tests (backend/internal/adapters/repository_test.go)
- Test Case 7: ResolveWorkspacesDir with root .agenticlens-config.yaml:
  - Create temporary root directory with .agenticlens-config.yaml specifying workspaces.dir as custom_root_ws. Assert that adapters.ResolveWorkspacesDir("") returns the expected path under the temp directory.
- Test Case 8: ResolveWorkspacesDir with legacy .agenticlens/config.yaml regression:
  - Assert that existing legacy tests continue to resolve workspace paths accurately.

### CLI Integration Tests (backend/cmd/agentic-lens/main_test.go)
- Test Case 9: CLI scan and generate with root .agenticlens-config.yaml:
  - In a test repository with root .agenticlens-config.yaml specifying custom ticket prefixes and ticket types, verify that CLI generate and scan respect the configured options.

### Invariant Checks
- Execute grep or static scan to confirm zero any across backend/internal/adapters/config/loader.go and backend/internal/adapters/repository.go.
- Verify zero emojis across all code, comments, documentation, and configuration files.

## 5. Review Criteria for Reviewer

### Axis 1: Specification and Contract Alignment
- Does FindConfigFile prioritize root .agenticlens-config.yaml -> .agenticlens-config.yml -> .agenticlens-config.json over legacy .agenticlens/ paths?
- Does legacy fallback remain fully operational when only .agenticlens/config.yaml or .agenticlens/config.json is present?
- Is .agenticlens-config.yaml present at the repository root and .agenticlens/ cleanly removed?
- Does ResolveWorkspacesDir safely handle both root and legacy config paths without path corruption?
- Are AGENTS.md, agent personas, and skills completely migrated to reference .agenticlens-config.yaml?

### Axis 2: Engineering Standards and Technical Invariants
- FCIS Discipline: Is domain logic completely free of filesystem knowledge? Does filesystem traversal remain strictly encapsulated in backend/internal/adapters/?
- Strong Typing: Is there zero usage of any or interface{} in Go?
- Error Handling: Are all filesystem errors checked and wrapped with descriptive context?
- Emoji Prohibition: Are all modified and created files 100% free of emojis?
- Quality Gates: Do make check, make test, and make build pass with zero warnings or errors?
