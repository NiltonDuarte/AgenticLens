h3. Title: AL-36: Cross-Project Interactive Setup Skill, Automated Artifact Downloader, and Agent Workflow Integrator

*TARGET*: AgenticLens (skills | cli | adapters | scripts | tests)

*CONTEXT*

Adopting AgenticLens in external repositories currently requires manual binary compilation, manual YAML authoring, and manual editing of agent system prompts. To achieve zero-friction onboarding, an AI agent in any target repository must be able to self-bootstrap AgenticLens when instructed by a user. The agent requires a dedicated setup skill (`agentic-lens-setup`) that coordinates release binary retrieval, guides an interactive user interview, writes `.agenticlens/config.yaml`, presents platform features, and updates repository agent instruction files (`AGENTS.md`) with mandatory telemetry extraction commands.

*GOAL*

Build and package the interactive `agentic-lens-setup` skill and downloader engine enabling an autonomous agent or developer to install the AgenticLens binary artifact, configure project parameters via interactive interview, explain platform functionalities, and verify telemetry integration in `AGENTS.md`.

*VERTICAL SLICE SCOPE*

* Skill Manifest & Core Logic (`.opencode/skills/agentic-lens-setup/SKILL.md`):
  - Author the skill specification defining the complete setup workflow, interview prompts, capability explainer, and verification checkpoints.
  - Sized for single-session autonomous execution by any OpenCode or harness agent.
* Artifact Downloader (`scripts/download-agentic-lens.sh` or Go helper):
  - Detect host operating system (`linux`, `darwin`, `windows`) and architecture (`amd64`, `arm64`).
  - Retrieve the appropriate release archive from GitHub Releases (`gh release download` or direct `curl` from release assets), extract the single binary to `./bin/agentic-lens`, and set executable permissions (`chmod +x`).
  - Provide graceful local compilation fallback (`make build`) if running inside the source repository or when network release assets are inaccessible.
* Interactive Interview & Config Generator:
  - Structure interactive questions: project name, ticket prefixes with conventional commit types, workspaces directory, source code folders, SQLite database path, and agent role mappings.
  - Generate a formatted, commented `.agenticlens/config.yaml` with the user's selected parameters.
* Platform Capability Explainer:
  - Formulate structured documentation within the skill explaining:
    - CLI subcommands: `generate` (telemetry extraction), `scan` (workspace audit), `server` (embedded web dashboard), `version`.
    - Real-time live session tracking, session tree reconstruction, and anomaly alerting.
    - Web dashboard navigation and KPI analysis.
* Agent Guidelines File Integration (`AGENTS.md`):
  - Audit existing agent guidelines (`AGENTS.md`, `.opencode/agent/*.md`, `CLAUDE.md`).
  - If absent, generate baseline `AGENTS.md`.
  - If present, inject the standardized AgenticLens operational policies:
    - Ticket tracking mandate (`<PREFIX>-<ID>`).
    - Dedicated workspace allocation (`<WORKSPACES_DIR>/<TICKET_ID>/`).
    - Delayed telemetry ingestion mandate as the final workflow turn before commit.
* Quality & Verification:
  - Test script covering OS/arch detection, asset unpacking, and config validation.
  - Self-test command: execute `./bin/agentic-lens version --json` and assert valid JSON.
  - Zero emojis in all scripts, prompts, and output.

*ACCEPTANCE CRITERIA*

* `.opencode/skills/agentic-lens-setup/SKILL.md` authored and registered.
* Downloader script reliably fetches and installs `./bin/agentic-lens` for Linux and macOS architectures.
* Interactive interview generates valid `.agenticlens/config.yaml` correctly parsed by `./bin/agentic-lens scan`.
* Agent instructions file (`AGENTS.md`) correctly audited and updated with delayed telemetry command.
* Platform capabilities and usage instructions clearly formatted and explained.
* Quality gates pass (`make check`, `make test`, `make build`) with zero emojis.

*DEFINITION OF DONE*

* Skill files, downloader scripts, and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, `make build`).
* Zero any and zero emojis across all files.
* Binary execution and telemetry generation tested via CLI.
