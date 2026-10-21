# feat(skills,cli,web): cross-project setup skill, artifact downloader, and dynamic ticket artifacts (AL-36)

## Overview

Implements the cross-project interactive setup skill (`agentic-lens-setup`), automated release artifact downloader engine, configuration generator, and dynamic ticket artifact system for zero-friction adoption of AgenticLens across external repositories. Autonomous agents can self-bootstrap the platform by retrieving precompiled release binaries into a self-contained git-ignored directory, guiding engineers through interactive configuration, scaffolding `.agenticlens-config.yaml` at the project root, and proposing agent guideline policies under a mandatory user confirmation gate.

Additionally decouples ticket artifacts from hardcoded assumptions by introducing configurable `workspaces.artifacts` rendered dynamically in the ticket deep dive dashboard, eliminates production `DefaultProjectConfig` in favor of strict configuration requirements, and targets the canonical release repository.

- Contract: [workspaces/AL-36/CONTRACT.md](workspaces/AL-36/CONTRACT.md)
- Telemetry & Summary: [workspaces/AL-36/SUMMARY.md](workspaces/AL-36/SUMMARY.md)

## Summary of Changes

### Skill Manifest & Downloader Engine (`.opencode/skills/agentic-lens-setup/`, `scripts/`)
- `.opencode/skills/agentic-lens-setup/SKILL.md`: Authored full setup skill specification across four phases: precompiled release retrieval, interactive parameter interview, platform capability explainer, and agent guidelines review with strict user confirmation gate. Configured canonical release repository: `https://github.com/NiltonDuarte/AgenticLens/releases`.
- `scripts/download-agentic-lens.sh` & co-located `download.sh`: POSIX shell script detecting host OS (`linux`, `darwin`, `windows`) and architecture (`amd64`, `arm64`), downloading release archives from GitHub Releases (`gh` CLI or `curl`/`wget` fallbacks), enforcing executable permissions, and generating protective `.gitignore` (`*` and `!.gitignore`) to prevent tracking binaries in git.
- Strict release-only mandate: zero source compilation or Go/Node toolchains required in target repositories.

### Dynamic Workspace Artifacts (`backend/internal/domain/`, `backend/internal/adapters/`, `web/`)
- Domain: Declared `WorkspaceArtifactConfig` and `TicketArtifactItem`. Updated `WorkspacesConfig.Artifacts` and `TicketArtifacts.Items`.
- Adapters: Updated configuration loader and writer to serialize and deserialize `workspaces.artifacts`. Completely excised legacy root fields (`LegacyTicketPrefix`, `LegacyWorkspacesDir`, `LegacySrcFolders`) from loader DTO and merge logic. Updated `ReadTicketArtifacts` with path containment checks preventing directory traversal exploits. Wired artifacts into HTTP detail endpoint.
- Web: Refactored `ArtifactViewer` to dynamically render configured artifact tabs (`CONTRACT.md`, `SUMMARY.md`, `PULL_REQUEST.md`, and custom documents) with markdown/raw views, empty badges, and clipboard copy.

### CLI Application Layer (`backend/cmd/agentic-lens/`)
- `init`: Strictly requires `--project-name`, `--src-folders`, and `--ticket-prefixes`, terminating with code 1 if omitted.
- Peer Ticket Prefixes: Parses comma-separated prefix list via `--ticket-prefixes` without any "primary" concept or automatic injection of `GEN`/`PO`.
- Comma-Separated Source Folders: Accepts comma-separated directory paths via `--src-folders`.
- Missing configuration guardrails: commands (`server`, `generate`, root scanner) check for configuration presence and exit 1 if absent, directing users to run `agentic-lens init`.
- Eliminated all hardcoded synthetic project defaults (`DefaultProjectConfig()`, `DefaultWorkspaceArtifacts()`) from production source code; test fixtures use dedicated test-scoped helpers.
- Strict configuration validation on load: missing `project_name`, `ticket.prefixes`, or `workspaces.src_folders` returns `ConfigValidationError` wrapping `ErrInvalidConfig`.

### Quality & Verification
- Unit and boundary tests passing across Go packages with race detector (`go test -race ./...`).
- Frontend test suite passing across 28 test files and 579 unit tests (`npm test`).
- Verification script `scripts/test-setup.sh` passing all 6 test stages.
- Quality gates pass (`make check`, `make test`, `make build`) with zero race conditions, zero `any`, and zero emojis.
