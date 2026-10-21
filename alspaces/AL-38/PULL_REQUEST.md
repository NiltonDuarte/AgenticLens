# refactor(config): migrate configuration manifest to root .agenticlens-config.yaml (AL-38)

## Overview

Migrates the AgenticLens configuration manifest from the dedicated subfolder `.agenticlens/config.yaml` to `.agenticlens-config.yaml` directly at the project root, aligning with modern development tooling conventions while eliminating unnecessary directory nesting. The boundary configuration loader has been updated to evaluate candidates with strict format and location priority (root YAML -> root YML -> root JSON -> legacy YAML -> legacy YML -> legacy JSON), guaranteeing complete backward compatibility for repositories with legacy subfolder manifests. Workspace base directory resolution and agent guidelines across the repository have been updated accordingly.

- Contract: [workspaces/AL-38/CONTRACT.md](workspaces/AL-38/CONTRACT.md)
- Telemetry: [workspaces/AL-38/SUMMARY.md](workspaces/AL-38/SUMMARY.md)

## Summary of Changes

### Adapters & Boundary Shell Layer (`backend/internal/adapters/`)
- `config/loader.go`:
  - Updated `FindConfigFile` candidate search order to prioritize root files (`.agenticlens-config.yaml`, `.agenticlens-config.yml`, `.agenticlens-config.json`) over legacy subfolder files (`.agenticlens/config.yaml`, `.agenticlens/config.yml`, `.agenticlens/config.json`).
  - Preserved upward parent directory traversal and fallback to domain defaults when no configuration is detected.
- `repository.go`:
  - Refactored `ResolveWorkspacesDir` base directory derivation to inspect the resolved configuration path: uses `filepath.Dir(cfgPath)` for root files and `filepath.Dir(filepath.Dir(cfgPath))` when nested under `.agenticlens/`.

### Repository Configuration Manifest
- Migrated `.agenticlens/config.yaml` to `.agenticlens-config.yaml` at the project root with all comments, mappings, and thresholds intact.
- Removed the empty `.agenticlens/` directory.

### Operational Guidelines & Agent Instruction Manifests
- `AGENTS.md`: Updated Section 1 Rule 1 configuration path reference.
- Agent Personas: Updated `.opencode/agent/orchestrator.md`, `planner.md`, `coder.md`, `tester.md`, and `reviewer.md` to reference `.agenticlens-config.yaml`.
- Skill Manifests: Updated `.opencode/skills/agentic-lens-dev/SKILL.md`, `agentic-lens-po/SKILL.md`, and `agentic-lens-review/SKILL.md`.

### Testing & Invariants
- Unit tests in `loader_test.go` covering full root schema parsing, format precedence, root-over-legacy precedence, six-tier elimination sequence, upward parent traversal, hostile filesystem edge cases, and corrupted syntax handling.
- Workspace resolution tests in `repository_test.go` across 9 distinct scenarios.
- CLI integration tests in `main_test.go` verifying flagless discovery with root configuration files.
- Zero `any` or `interface{}`, zero emojis, and zero skipped tests.
