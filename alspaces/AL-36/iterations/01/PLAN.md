# Implementation Plan: AL-36 Cross-Project Interactive Setup Skill, Automated Artifact Downloader, and Agent Workflow Integrator

## 1. Goal and Iteration Scope

Deliver the complete interactive setup skill, precompiled artifact downloader, configuration generator, and agent workflow integrator for cross-project adoption of AgenticLens:
- Author the .opencode/skills/agentic-lens-setup/SKILL.md skill manifest governing the entire bootstrapping lifecycle.
- Implement the release artifact downloader script with multi-platform detection, GitHub Releases retrieval, archive extraction, binary co-location, and protective gitignore generation.
- Implement configuration generation in the Go backend via adapter writer functions and the agentic-lens init CLI subcommand.
- Implement platform capability explainer and agent guidelines reviewer within the setup skill with a strict user confirmation gate.
- Author an automated verification script and unit tests verifying zero-compilation release download, root configuration generation, and binary execution.

---

## 2. Grounded Reference Architecture & File Paths

The implementation grounds against concrete existing files, directories, and symbols:

### Skills and Agent Architecture (.opencode/skills/ and .opencode/agent/)
- .opencode/skills/agentic-lens-setup/SKILL.md:
  - Complete skill manifest defining name, description, parameters, and step-by-step interactive workflow.
- .opencode/skills/agentic-lens-setup/bin/.gitignore:
  - Protective gitignore co-located in the skill binary folder to prevent committing downloaded executables.
- .opencode/skills/agentic-lens-setup/scripts/download.sh:
  - Skill-local copy or execution symlink of the downloader script.

### Shell Scripts (scripts/)
- scripts/download-agentic-lens.sh:
  - Standalone POSIX-compliant shell script detecting OS (linux, darwin, windows) and architecture (amd64, arm64).
  - Downloads release artifacts from GitHub Releases (gh CLI or curl fallback), extracts binary to destination directory, writes .gitignore, and sets executable permissions.
- scripts/test-setup.sh:
  - Automated verification test suite validating OS/arch detection, extraction, config generation, and binary self-test execution.

### Go Functional Core & Configuration Adapters (backend/internal/domain/ and backend/internal/adapters/config/)
- backend/internal/domain/config.go:
  - Contains domain.ProjectConfig, domain.TicketConfig, domain.DefaultProjectConfig().
  - Ground for default configuration properties and validation rules.
- backend/internal/adapters/config/loader.go:
  - Ground for configuration file discovery and YAML unmarshaling.
- backend/internal/adapters/config/writer.go:
  - New adapter containing GenerateConfigYAML(cfg domain.ProjectConfig) ([]byte, error) and WriteConfigFile(targetPath string, cfg domain.ProjectConfig, force bool) error.
  - Generates cleanly formatted, richly commented YAML matching root .agenticlens-config.yaml conventions.
- backend/internal/adapters/config/writer_test.go:
  - New unit tests verifying YAML formatting, comment preservation, round-trip unmarshaling with adaptersconfig.Load, and file write safety.

### Go CLI Application Layer (backend/cmd/agentic-lens/)
- backend/cmd/agentic-lens/main.go:
  - Add case "init": in subcommand routing.
  - Implement runInit(args []string) parsing flags (--project-name, --workspaces, --db-path, --src-folders, --force, --dry-run, --output).
  - Update root command usage text and help documentation.
- backend/cmd/agentic-lens/main_test.go:
  - Test cases for runInit verifying flag parsing, file creation, force overwriting, and dry-run execution.

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Go Backend Configuration Generator and CLI Init Subcommand
- Task 1.1: Create backend/internal/adapters/config/writer.go.
  - Implement GenerateDefaultYAML(projectName string, workspacesDir string, dbPath string, srcFolders []string, prefixes []domain.TicketPrefix) ([]byte, error).
  - Ensure the output YAML includes structured comments for all sections: project name, ticket identification, workspaces, database connectivity, live observability thresholds, code metrics patterns, role mappings, and role display order.
  - Implement WriteConfigFile(path string, data []byte, force bool) error. If the target file already exists and force is false, return an explicit typed error indicating file exists.
- Task 1.2: Implement runInit in backend/cmd/agentic-lens/main.go.
  - Wire init subcommand into main().
  - Parse CLI flags: --project-name (defaults to current directory name), --workspaces (defaults to "workspaces"), --db-path (defaults to "~/.local/share/opencode/opencode.db"), --src-folders (comma-separated list), --force (overwrite existing), --dry-run (print to stdout without writing), and --output (defaults to ".agenticlens-config.yaml").
  - On execution, call writer.WriteConfigFile or print to stdout if dry-run.
  - Output descriptive success message with project name and configuration path upon completion.
- Task 1.3: Author unit tests in backend/internal/adapters/config/writer_test.go and backend/cmd/agentic-lens/main_test.go.
  - Assert generated YAML parses cleanly back through adaptersconfig.Load.
  - Assert that attempting to write to an existing file without --force fails safely.
  - Assert that --force successfully overwrites.

### Phase 2: Automated Artifact Downloader Script
- Task 2.1: Create scripts/download-agentic-lens.sh.
  - Enforce strict shell error handling (set -e).
  - Detect operating system using uname -s. Normalize Linux to linux, Darwin to darwin, MINGW/MSYS/CYGWIN to windows.
  - Detect architecture using uname -m. Normalize x86_64 and amd64 to amd64, arm64 and aarch64 to arm64.
  - Support configurable target version (defaults to latest published release).
  - Support configurable output binary directory (defaults to .opencode/skills/agentic-lens-setup/bin).
  - Retrieve precompiled archive or binary from GitHub Releases:
    - If gh CLI tool is available, use gh release download.
    - If gh is unavailable, use curl -fsSL or wget -qO- targeting GitHub Releases download endpoints.
  - Unpack archive into temporary location and move the compiled executable to the destination directory.
  - Ensure executable permissions (chmod +x).
  - Automatically create the destination directory if missing and write .gitignore containing wildcard pattern ignoring all binaries while preserving .gitignore.
- Task 2.2: Ensure local copy exists at .opencode/skills/agentic-lens-setup/scripts/download.sh and make executable.

### Phase 3: Setup Skill Manifest Authoring
- Task 3.1: Create .opencode/skills/agentic-lens-setup/SKILL.md.
  - Author YAML frontmatter with name: agentic-lens-setup and clear description.
  - Document Step 1: Precompiled Release Retrieval.
    - Instruct the agent to run the downloader script to fetch the latest precompiled binary into .opencode/skills/agentic-lens-setup/bin/agentic-lens.
    - Explicitly prohibit building from source or requesting Go/Node toolchain installation.
    - Instruct verification of downloaded binary via ./.opencode/skills/agentic-lens-setup/bin/agentic-lens version --json.
  - Document Step 2: Interactive Interview & Configuration Scaffolding.
    - Present structured interview questions: project name, ticket prefixes with conventional commit types, workspaces directory, source code folders, SQLite database path, and agent role mappings.
    - Instruct the agent to write .agenticlens-config.yaml at project root using the interview answers or by invoking ./.opencode/skills/agentic-lens-setup/bin/agentic-lens init with the collected flags.
  - Document Step 3: Platform Capability Walkthrough.
    - Document CLI subcommands: generate, scan, server, version, init.
    - Document live session observability, SQLite hooks, and anomaly alerting.
    - Document web dashboard capabilities, view modes, and navigation.
  - Document Step 4: Agent Guidelines Review & Mandatory Confirmation Gate.
    - Instruct the agent to inspect repository files (AGENTS.md, .opencode/agent/*.md, CLAUDE.md).
    - Enforce Mandatory User Confirmation Gate: agent must present the exact markdown diff or block to the user and request explicit confirmation before writing.
    - Document the standard operational policies: ticket tracking mandate, dedicated workspace layout, and delayed telemetry ingestion mandate.

### Phase 4: Automated Verification Test Script and Self-Test
- Task 4.1: Create scripts/test-setup.sh.
  - Test OS and architecture detection against various environment strings.
  - Test protective .gitignore creation and ensure binary files are ignored by git.
  - Test root .agenticlens-config.yaml creation and validate it with backend/cmd/agentic-lens scan or config loader.
  - Test execution of the co-located binary: run version check with --json flag and verify valid JSON structure.
  - Assert zero emojis across all script outputs.
- Task 4.2: Execute make check, make test, and make build to ensure the entire repository passes all quality gates.

---

## 4. Verification & Testing Strategy for Tester

The tester subagent must execute adversarial verification covering:
1. Go Adapter & CLI Tests:
   - writer_test.go: Verify YAML string generation, check that all expected top-level keys exist (project_name, ticket, workspaces, database, live, code_metrics, role_mappings, role_order), and verify round-trip unmarshaling with adaptersconfig.Load.
   - main_test.go: Execute agentic-lens init tests verifying --force, --dry-run, custom --project-name, and default .agenticlens-config.yaml location.
2. Downloader Script Tests:
   - Execute scripts/download-agentic-lens.sh in dry-run or mock mode to verify architecture and OS resolution logic across linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64.
   - Verify .gitignore exists at .opencode/skills/agentic-lens-setup/bin/.gitignore and contains the protective wildcard pattern.
   - Run git status --porcelain to verify that any binary inside the skill directory remains untracked.
3. Skill Self-Test:
   - Execute the binary in the skill directory: ./.opencode/skills/agentic-lens-setup/bin/agentic-lens version --json and assert zero exit status and valid JSON containing version, git_commit, and build_date.
4. Zero Emojis and Quality Gates:
   - Scan new files for any unicode emojis.
   - Run make check, make test, and make build.

---

## 5. Audit & Review Criteria for Reviewer

The reviewer subagent must audit changes along two strict axes:
- Axis 1: Spec Alignment
  - Is .opencode/skills/agentic-lens-setup/SKILL.md present and registered?
  - Does the downloader fetch precompiled release artifacts only, strictly avoiding source compilation or toolchain dependencies?
  - Is the binary co-located within the skill directory alongside a protective .gitignore preventing it from being committed?
  - Does the interactive interview produce a valid, commented .agenticlens-config.yaml at project root?
  - Does the skill enforce a strict user confirmation gate before modifying AGENTS.md?
  - Are platform capabilities clearly explained?
- Axis 2: Engineering Standards
  - Are Go domain layers free of filesystem I/O?
  - Is boundary I/O strictly encapsulated in internal/adapters/?
  - Are there zero any or interface{} types across Go and TypeScript?
  - Are all scripts and outputs completely free of emojis?
