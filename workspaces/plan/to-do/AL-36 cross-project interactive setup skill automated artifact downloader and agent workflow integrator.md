h3. Title: AL-36: Cross-Project Interactive Setup Skill, Automated Artifact Downloader, and Agent Workflow Integrator

*TARGET*: AgenticLens (skills | cli | adapters | scripts | tests)

*CONTEXT*

Adopting AgenticLens in external repositories currently requires manual binary retrieval, manual YAML authoring, and manual editing of agent guidelines. To achieve zero-friction onboarding, an AI agent in any target repository must be able to self-bootstrap AgenticLens when instructed by a user. The agent requires a dedicated setup skill (`agentic-lens-setup`) that coordinates release binary retrieval, co-locates the executable cleanly with a protective `.gitignore`, guides an interactive user interview, writes `.agenticlens-config.yaml` at the project root, presents platform capabilities, and inspects repository agent files (`AGENTS.md`), strictly requiring user confirmation before injecting any instructions.

*GOAL*

Build and package the interactive `agentic-lens-setup` skill and downloader engine enabling an autonomous agent to retrieve the precompiled AgenticLens release binary into a self-contained, git-ignored skill folder, configure project parameters via interactive interview into root `.agenticlens-config.yaml`, explain platform features, and propose telemetry integration for `AGENTS.md` with mandatory user approval before injection.

*VERTICAL SLICE SCOPE*

* Skill Manifest & Core Logic (`.opencode/skills/agentic-lens-setup/SKILL.md`):
  - Author the skill specification defining the complete setup workflow: binary download, interactive interview, config generation, capability explainer, and agent file review.
  - Sized for single-session autonomous execution by any OpenCode or harness agent.
* Automated Artifact Downloader (Release-Only, Zero Compilation):
  - Detect host operating system (`linux`, `darwin`, `windows`) and architecture (`amd64`, `arm64`).
  - Retrieve the appropriate precompiled release archive from GitHub Releases (`https://github.com/<owner>/AgenticLens/releases/latest/download/...` or via `gh release download`).
  - Strict release-only mandate: never ask the user to compile the platform from source; zero build toolchain dependencies required.
  - Binary Co-Location & Protective `.gitignore`:
    - Co-locate the downloaded binary inside the skill directory (e.g. `.opencode/skills/agentic-lens-setup/bin/agentic-lens`).
    - Co-locate a `.gitignore` inside the skill binary directory containing `*` (or `agentic-lens*`), ensuring the binary is never accidentally committed to the target project's Git history.
    - Set executable permissions (`chmod +x`).
* Interactive Interview & Root Config Generator:
  - Structure interactive questions: project name, ticket prefixes with conventional commit types, workspaces directory (default: `workspaces`), source code folders to observe, SQLite database path (default: `~/.local/share/opencode/opencode.db`), and agent role mappings.
  - Generate a formatted, commented `.agenticlens-config.yaml` located at the project root.
* Platform Capability Explainer:
  - Formulate structured documentation within the skill explaining:
    - CLI subcommands: `generate` (telemetry extraction), `scan` (workspace audit), `server` (embedded web dashboard), `version`.
    - Real-time live session tracking, session tree reconstruction, and anomaly alerting.
    - Web dashboard navigation with permanent `AgenticLens` platform branding and observed project context.
* Agent Guidelines Inspection with Mandatory User Approval:
  - Inspect repository for agent instruction files (`AGENTS.md`, `.opencode/agent/*.md`, `CLAUDE.md`).
  - Strict User Confirmation Rule: never modify, inject, or write instructions to `AGENTS.md` without presenting the proposed text and asking the user for explicit permission first.
  - Present the standardized operational policies:
    - Ticket tracking mandate (`<PREFIX>-<ID>`).
    - Dedicated workspace allocation (`<WORKSPACES_DIR>/<TICKET_ID>/`).
    - Delayed telemetry ingestion mandate as the final workflow turn before commit.
* Quality & Verification:
  - Test script covering OS/arch detection, asset unpacking into co-located folder, and root `.agenticlens-config.yaml` generation.
  - Self-test: execute the co-located binary `./.opencode/skills/agentic-lens-setup/bin/agentic-lens version --json` and assert valid JSON.
  - Zero emojis in all scripts, prompts, and output.

*ACCEPTANCE CRITERIA*

* `.opencode/skills/agentic-lens-setup/SKILL.md` authored and registered.
* Downloader fetches precompiled binary release only; zero compilation or build steps requested from user.
* Binary is co-located within the skill directory alongside a `.gitignore` that prevents it from being tracked in git.
* Interactive interview produces valid `.agenticlens-config.yaml` at the project root.
* Agent instructions file (`AGENTS.md`) is inspected and changes are proposed to the user; zero modifications made without explicit user confirmation.
* Platform capabilities and usage instructions clearly formatted and explained.
* Quality gates pass (`make check`, `make test`, `make build`) with zero emojis.

*DEFINITION OF DONE*

* Skill files, downloader scripts, and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, `make build`).
* Zero any and zero emojis across all files.
* Binary execution and telemetry generation tested via CLI.
