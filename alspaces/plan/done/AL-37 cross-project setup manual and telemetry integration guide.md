h3. Title: AL-37: Cross-Project Platform Setup Manual and Telemetry Integration Guide

*TARGET*: AgenticLens (docs | packaging | cli)

*CONTEXT*

External engineering organizations, platform architects, and developers require comprehensive documentation to onboard and operate AgenticLens across diverse repositories. Crucially, the internal skills used to build this repository (`agentic-lens-dev`, `agentic-lens-po`, and `agentic-lens-review`) are part of this repository's own git project structure, NOT parts of the general AgenticLens product platform. External projects adopting AgenticLens must define their own ways of working, agent personas, and delivery workflows rather than copying the multi-agent conventions of this repository. They need a definitive manual explaining platform installation, root configuration schema, CLI flags, web dashboard navigation, and how to integrate AgenticLens telemetry generation into their own custom agent workflows.

*GOAL*

Author the comprehensive cross-project setup manual (`docs/SETUP_MANUAL.md`) and package the standalone `agentic-lens-setup` skill for external distribution, establishing a clear separation between the AgenticLens product platform and project-specific development methodologies.

*VERTICAL SLICE SCOPE*

* Documentation Core (`docs/SETUP_MANUAL.md`):
  - Platform Overview & Value: Zero-dependency single-binary architecture, local telemetry ledgers, live observability, and context health analytics.
  - Setup Pathways:
    - Agent-Driven Onboarding: Running `agentic-lens-setup` to self-install the precompiled binary into a co-located git-ignored folder, configure `.agenticlens-config.yaml`, and confirm agent instructions.
    - Manual Developer Installation: Downloading precompiled binaries from GitHub Releases, making executable, and generating `.agenticlens-config.yaml`.
  - Configuration Reference: Exhaustive `.agenticlens-config.yaml` schema reference for root configuration (`project_name`, `ticket`, `workspaces`, `database`, `live`, `code_metrics`, `role_mappings`, `role_order`).
  - Agnostic Agent Workflow Integration:
    - Guide for any agent setup (single-agent, multi-agent, custom harnesses) on how to invoke `./bin/agentic-lens generate ...` as the final turn before committing work.
    - Demonstrating that projects define their own workflows, personas, and commit strategies, integrating the telemetry ledger without inheriting AgenticLens-specific roles.
  - CLI & Web Dashboard Reference: Full operational guide for `server`, `generate`, `scan`, `version`, `migrate`, and `backfill`.
  - Troubleshooting & Best Practices: Database locks, permissions, SQLite paths, and git branch inference.
* Standalone Setup Skill Packaging:
  - Provide a clean distribution mechanism for the single platform onboarding skill (`agentic-lens-setup` in `.opencode/skills/agentic-lens-setup/`).
  - Ensure zero export or bundling of internal repository skills (`agentic-lens-dev`, `agentic-lens-po`, `agentic-lens-review`).
* Quality:
  - Verify all CLI examples, config snippets, and command flags.
  - Ensure zero emojis across all documentation and guides.

*ACCEPTANCE CRITERIA*

* `docs/SETUP_MANUAL.md` authored, structured, and validated against technical guidelines.
* Explicitly distinguishes the AgenticLens product platform from project-specific internal skills.
* Clear instructions for both agent-driven setup and manual installation.
* Exhaustive `.agenticlens-config.yaml` reference included with typed examples.
* Only the platform setup skill `agentic-lens-setup` is packaged for distribution; internal repo skills are excluded.
* Quality gates pass (`make check`, `make test`, `make build`) with zero emojis.

*DEFINITION OF DONE*

* Documentation and skill distribution files merged to target branch.
* Quality gates passing (`make check`, `make test`, `make build`).
* Zero any and zero emojis across all files.
