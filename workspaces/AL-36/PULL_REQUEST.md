# feat(skills,cli): cross-project setup skill, artifact downloader, and config generator (AL-36)

## Overview

Implements the cross-project interactive setup skill (`agentic-lens-setup`), release artifact downloader engine, configuration generator, and CLI `init` command for zero-friction adoption of AgenticLens across external repositories. Autonomous agents can now self-bootstrap the platform by retrieving precompiled release binaries into a self-contained git-ignored directory, guiding engineers through interactive configuration, scaffolding `.agenticlens-config.yaml` at the project root, and proposing agent guideline policies under a mandatory user confirmation gate.

- Contract: [workspaces/AL-36/CONTRACT.md](workspaces/AL-36/CONTRACT.md)
- Telemetry & Summary: [workspaces/AL-36/SUMMARY.md](workspaces/AL-36/SUMMARY.md)

## Summary of Changes

### Skill Manifest & Downloader Engine (`.opencode/skills/agentic-lens-setup/`, `scripts/`)
- `.opencode/skills/agentic-lens-setup/SKILL.md`: Authored full setup skill specification across four phases: precompiled release retrieval, interactive parameter interview, platform capability explainer, and agent guidelines review with strict user confirmation gate.
- `scripts/download-agentic-lens.sh` & co-located `download.sh`: POSIX shell script detecting host OS (`linux`, `darwin`, `windows`) and architecture (`amd64`, `arm64`), downloading release archives from GitHub Releases (`gh` CLI or `curl`/`wget` fallbacks), enforcing executable permissions, and generating protective `.gitignore` (`*` and `!.gitignore`) to prevent tracking binaries in git.
- Strict release-only mandate: zero source compilation or Go/Node toolchains required in target repositories.

### Adapters & CLI Layer (`backend/internal/adapters/config/`, `backend/cmd/agentic-lens/`)
- `writer.go`: Implemented `GenerateConfigYAML` producing commented `.agenticlens-config.yaml` manifests and `WriteConfigFile` ensuring atomic writes via temporary files and typed collision protection.
- `main.go`: Implemented `init` CLI subcommand supporting `--project-name`, `--workspaces`, `--db-path`, `--src-folders`, `--ticket-prefix`, `--output`, `--force`, and `--dry-run`.

### Quality & Verification
- Unit tests in `writer_test.go` and `main_test.go` covering YAML generation, round-trip unmarshaling, collisions, `--force`, `--dry-run`, spaces in paths, and permission boundaries.
- Verification script `scripts/test-setup.sh` passing all 6 test stages.
- Quality gates pass (`make check`, `make test`, `make build`) with zero race conditions, zero `any`, and zero emojis.
