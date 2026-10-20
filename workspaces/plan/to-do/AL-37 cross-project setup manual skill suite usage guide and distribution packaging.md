h3. Title: AL-37: Cross-Project Setup Manual, Skill Suite Usage Guide, and Distribution Packaging

*TARGET*: AgenticLens (docs | packaging | skills | cli)

*CONTEXT*

While the `agentic-lens-setup` skill automates initial onboarding, engineering organizations, platform architects, and developers require comprehensive documentation and distribution packaging for cross-project adoption. Teams need a definitive manual explaining architecture, configuration schemas, manual and agent-driven setup pathways, telemetry ledger integration into CI/CD, live session tracking, and dashboard navigation. Furthermore, the complete multi-agent skill suite (`agentic-lens-setup`, `agentic-lens-po`, `agentic-lens-dev`, `agentic-lens-review`) must be packaged and distributed so external projects can install and utilize the entire agent workflow seamlessly.

*GOAL*

Author the comprehensive cross-project setup manual (`docs/SETUP_MANUAL.md`), create a portable skill distribution installer, and establish genericized multi-project skill packaging for seamless adoption across diverse engineering repositories.

*VERTICAL SLICE SCOPE*

* Documentation Core (`docs/SETUP_MANUAL.md`):
  - Quickstart Guide: 3-step onboarding overview (install binary, configure YAML, update agent prompt).
  - Setup Pathways: Detailed instructions for both agent-automated setup (via `agentic-lens-setup` skill) and manual developer installation.
  - Configuration Reference: Exhaustive `.agenticlens/config.yaml` schema documentation with typed examples for every section (`project_name`, `ticket`, `workspaces`, `database`, `live`, `code_metrics`, `role_mappings`, `role_order`).
  - Agent Workflow Integration: Complete guide on structuring `AGENTS.md`, enforcing single-ticket exclusivity, allocating workspace artifacts, and executing delayed telemetry ingestion with conventional commit classification.
  - CLI & Dashboard Operations: Comprehensive reference for all CLI commands (`server`, `generate`, `scan`, `version`, `migrate`, `backfill`) and web dashboard navigation (Fleet Overview, Ticket Comparisons, Metric Progression, Live Trace Waterfall).
  - Troubleshooting & Best Practices: Diagnosing SQLite access issues, resolving unmapped ticket types, handling git branch conventions, and managing ledger compactness.
* Skill Suite Packaging & Distribution (`scripts/install-skills.sh`):
  - Create a portable skill installer script that copies and configures the AgenticLens skill suite into a target repository's `.opencode/skills/`:
    - `agentic-lens-setup`: Initial onboarding, interactive interview, and configuration.
    - `agentic-lens-po`: Genericized requirements discovery, specification drafting, and Jira ticket decomposition.
    - `agentic-lens-dev`: Genericized multi-agent orchestrator (`planner` -> `coder` -> `tester` -> `reviewer`).
    - `agentic-lens-review`: Genericized architectural audit, code smell detection, and test profiling.
  - Audit all skill manifests to guarantee they use repository-relative paths and configuration parameters (`.agenticlens/config.yaml`) without hardcoded AgenticLens references.
* Quality:
  - Validate all command snippets, curl examples, and configuration blocks.
  - Ensure zero emojis across all documentation, installer scripts, and skill definitions.

*ACCEPTANCE CRITERIA*

* `docs/SETUP_MANUAL.md` authored, structured, and validated against technical guidelines.
* `scripts/install-skills.sh` installs and configures the four-skill suite into a specified target directory.
* Skill manifests genericized for cross-project compatibility without project-specific path leaks.
* Quality gates pass (`make check`, `make test`, `make build`) with zero emojis.

*DEFINITION OF DONE*

* Documentation, packaging scripts, and skill manifests merged to target branch.
* Quality gates passing (`make check`, `make test`, `make build`).
* Zero any and zero emojis across all files.
