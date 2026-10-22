# Contract: AL-38 Migrate Configuration Manifest to Root .agenticlens-config.yaml

## 1. Problem Statement

AgenticLens currently loads project configuration from a dedicated subfolder manifest: .agenticlens/config.yaml (with fallbacks to .agenticlens/config.yml and legacy .agenticlens/config.json). Maintaining a dedicated hidden directory .agenticlens/ solely to house a single YAML configuration file introduces unnecessary filesystem nesting and departs from standard repository ergonomics observed in modern developer tools (such as .golangci.yaml, .prettierrc.yaml, and .release-please-config.json).

Relocating the primary configuration file to .agenticlens-config.yaml directly in the project root streamlines project setup, enhances visibility, and eliminates redundant directory structures. To ensure zero disruption across existing projects, workspaces, and test harnesses during migration, the configuration loader adapter must prioritize root .agenticlens-config.yaml while guaranteeing backward-compatible fallback to legacy .agenticlens/ paths.

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All changes must strictly adhere to TECHNICAL_GUIDELINES.md and AGENTS.md.

### Inward Dependency Flow
1. Functional Core Domain Layer (backend/internal/domain/config.go):
   - Houses pure configuration models (ProjectConfig, TicketConfig, WorkspacesConfig, DatabaseConfig, LiveConfig, AnomalyThresholdsConfig, CodeMetricsConfig, RoleMappings) and pure transformation functions (NormalizeTicketID, DefaultProjectConfig).
   - Invariant: Zero external dependencies. Zero filesystem traversal, zero file reading, zero knowledge of disk paths (.agenticlens-config.yaml or .agenticlens/).
2. Boundary Adapter Layer (backend/internal/adapters/config/loader.go):
   - Encapsulates all filesystem I/O, parent directory traversals, format detection, YAML/JSON deserialization, and fallback merging.
   - Updates candidate search paths and exposes Load(repoRoot) and FindConfigFile(startDir).
   - Invariant: Only adapters touch the filesystem. Adapters return pure domain models.
3. Consuming Adapters and CLI Layer:
   - backend/internal/adapters/repository.go (ResolveWorkspacesDir): Discovers workspaces relative to the repository root determined by FindConfigFile. It must safely handle both root configuration paths and legacy nested subfolder paths without path traversal errors.
   - backend/internal/adapters/git/classifier.go (LoadConfig): Delegates directly to config.Load(repoRoot).
   - backend/cmd/agentic-lens/main.go (runGenerate, runScan, runServer): Discovers project configuration through config.Load("").
4. Presentation Layer (web/):
   - Pure presentation. The web SPA consumes configuration aggregated and served by the Go HTTP server. Zero direct file dependencies.

### Strong Typing and Zero Any Invariant
- Zero usage of any or interface{} across all Go code.
- Typed structs and immutable value semantics for all configuration representations.

### Zero Emojis Mandate
- Absolute prohibition of emojis across code, comments, docstrings, configuration manifests, CLI messages, logs, test outputs, and workspace artifacts.

## 3. Scope Boundaries

### In Scope for AL-38
1. Configuration Loader Search Hierarchy Update (backend/internal/adapters/config/loader.go):
   - Update FindConfigFile to search at each directory level in strict order:
     Priority 1: .agenticlens-config.yaml (repo root)
     Priority 2: .agenticlens-config.yml (repo root)
     Priority 3: .agenticlens-config.json (repo root)
     Priority 4: .agenticlens/config.yaml (legacy subfolder)
     Priority 5: .agenticlens/config.yml (legacy subfolder)
     Priority 6: .agenticlens/config.json (legacy subfolder)
   - Preserve existing upward traversal logic to parent directories until the filesystem root boundary.
2. Workspace Directory Resolution Adaptation (backend/internal/adapters/repository.go):
   - Refactor base directory derivation in ResolveWorkspacesDir. When cfgPath points directly to a root config file (.agenticlens-config.yaml), the repository base directory is filepath.Dir(cfgPath). When cfgPath points to a legacy subfolder file (.agenticlens/config.yaml), the base directory is filepath.Dir(filepath.Dir(cfgPath)).
3. Repository Configuration Manifest Relocation:
   - Move .agenticlens/config.yaml to .agenticlens-config.yaml at the repository root.
   - Remove the empty .agenticlens/ directory.
4. Operational Guidelines, Agent Instructions, and Skill Manifests:
   - Update configuration file references from .agenticlens/config.yaml to .agenticlens-config.yaml in AGENTS.md, .opencode/agent/orchestrator.md, .opencode/agent/planner.md, .opencode/agent/coder.md, .opencode/agent/tester.md, .opencode/agent/reviewer.md, and skills (.opencode/skills/agentic-lens-dev/SKILL.md, .opencode/skills/agentic-lens-po/SKILL.md, .opencode/skills/agentic-lens-review/SKILL.md).
5. Comprehensive Test Suite Updates:
   - Table-driven unit tests in backend/internal/adapters/config/loader_test.go covering root discovery, root format precedence, root-over-legacy precedence, legacy fallback, and missing file fallback.
   - Repository resolution tests in backend/internal/adapters/repository_test.go.
   - CLI command tests in backend/cmd/agentic-lens/main_test.go verifying that scan and generate parse .agenticlens-config.yaml at project root.

### Out of Scope for AL-38
1. Modifying domain configuration models or adding new schema keys in backend/internal/domain/config.go.
2. Altering web dashboard UI layout or TypeScript types in web/.
3. Implementing cross-project setup skill automated downloader (governed downstream by AL-36).
4. Authoring cross-project setup documentation (governed downstream by AL-37).

## 4. Architectural Contracts and Discovery Precedence

### Discovery Contract in backend/internal/adapters/config/loader.go

Function FindConfigFile(startDir string) (string, error):
- Evaluates startDir; if blank, resolves current working directory.
- For each directory step from startDir upwards to filesystem boundary:
  Checks candidates in this exact priority sequence:
  1. filepath.Join(curr, ".agenticlens-config.yaml")
  2. filepath.Join(curr, ".agenticlens-config.yml")
  3. filepath.Join(curr, ".agenticlens-config.json")
  4. filepath.Join(curr, ".agenticlens", "config.yaml")
  5. filepath.Join(curr, ".agenticlens", "config.yml")
  6. filepath.Join(curr, ".agenticlens", "config.json")
- The first existing non-directory path encountered is returned immediately as the resolved configuration path with nil error.
- If the filesystem root is reached without any match, returns empty string and nil error.

Function LoadFromFile(filePath string) (domain.ProjectConfig, error):
- Reads target file path.
- Resolves file extension: .yaml and .yml use yaml.Unmarshal; .json uses json.Unmarshal; unrecognized extensions attempt YAML unmarshal followed by JSON unmarshal fallback.
- Merges unmarshaled values on top of domain.DefaultProjectConfig().

Function Load(repoRoot string) (domain.ProjectConfig, error):
- Calls FindConfigFile(repoRoot). If error occurs, returns DefaultProjectConfig() and error.
- If no file is found, returns DefaultProjectConfig() and nil error.
- If file is found, delegates to LoadFromFile.

### Base Directory Derivation Contract in backend/internal/adapters/repository.go

In ResolveWorkspacesDir(overridePath string) string:
- When candidate configuration file is located at cfgPath:
  Derive base directory:
  If filepath.Base(filepath.Dir(cfgPath)) equals ".agenticlens", baseDir is filepath.Dir(filepath.Dir(cfgPath)).
  Otherwise, baseDir is filepath.Dir(cfgPath).
- Resolves candidate workspace folder at filepath.Join(baseDir, wsDir).

## 5. Backward Compatibility & Fallback Guarantees

1. Precedence Guarantee:
   If both .agenticlens-config.yaml and legacy .agenticlens/config.yaml exist in the same repository tree, .agenticlens-config.yaml is loaded, and legacy files are ignored.
2. Legacy Fallback Guarantee:
   If a repository only defines .agenticlens/config.yaml, .agenticlens/config.yml, or .agenticlens/config.json, the loader successfully discovers and loads it without errors.
3. Absence Resilience:
   If neither root nor legacy configuration files exist, Load returns domain.DefaultProjectConfig() with a nil error.
4. Schema Parity:
   All YAML/JSON fields, comments, and structure in .agenticlens/config.yaml remain identical when relocated to root .agenticlens-config.yaml.

## 6. Acceptance Criteria

1. .agenticlens-config.yaml resides at repository root and is successfully discovered by FindConfigFile and Load.
2. Legacy .agenticlens/config.yaml continues to load seamlessly if the root file is absent.
3. Search order strictly prioritizes root YAML -> root YML -> root JSON -> legacy YAML -> legacy YML -> legacy JSON.
4. ResolveWorkspacesDir correctly derives root directories whether cfgPath points to root .agenticlens-config.yaml or legacy .agenticlens/config.yaml.
5. Empty .agenticlens/ directory is removed from the repository.
6. References across AGENTS.md, .opencode/agent/*.md, and .opencode/skills/*.md cite .agenticlens-config.yaml.
7. Unit, boundary, and CLI integration tests pass cleanly with race detection enabled.
8. Zero any and zero emojis across all repository files.
9. make check and make test pass with zero errors.
