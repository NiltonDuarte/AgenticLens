h3. Title: AL-38: Migrate Configuration Manifest from .agenticlens/config.yaml to Root .agenticlens-config.yaml

*TARGET*: AgenticLens (domain | adapters | cli | config | tests)

*CONTEXT*

AgenticLens currently loads project configuration from `.agenticlens/config.yaml` (or `.agenticlens/config.json`). Creating a subfolder `.agenticlens/` solely to hold a single YAML configuration file adds unnecessary repository nesting. Relocating the primary configuration file to `.agenticlens-config.yaml` directly in the project root aligns with standard modern development tooling conventions (such as `.golangci.yaml`, `.prettierrc.yaml`, `.release-please-config.json`). The configuration loader must prioritize `.agenticlens-config.yaml` in the root while retaining backward-compatible fallback to legacy `.agenticlens/config.yaml` during transition.

*GOAL*

Migrate the project configuration manifest to `.agenticlens-config.yaml` located at the project root, update the Go configuration loader adapter to discover root `.agenticlens-config.yaml` first with legacy fallback, and update all references across adapters, CLI commands, agent instructions (`AGENTS.md`), and test suites.

*VERTICAL SLICE SCOPE*

* Boundary Adapter (`backend/internal/adapters/config/loader.go`):
  - Update candidate search order in `FindConfigFile` to prioritize:
    1. `.agenticlens-config.yaml` (repo root)
    2. `.agenticlens-config.yml` (repo root)
    3. `.agenticlens-config.json` (repo root)
    4. `.agenticlens/config.yaml` (legacy subfolder)
    5. `.agenticlens/config.yml` (legacy subfolder)
    6. `.agenticlens/config.json` (legacy subfolder)
  - Ensure `Load(repoRoot)` seamlessly resolves root `.agenticlens-config.yaml`.
* Repository Migration:
  - Move `.agenticlens/config.yaml` to `.agenticlens-config.yaml` at project root.
  - Remove empty `.agenticlens/` directory if no other files reside within it.
* System Wiring & Agent Guidelines:
  - Update references across `backend/internal/adapters/repository.go`, `git/classifier.go`, `cmd/agentic-lens/main.go`.
  - Update `AGENTS.md` and skill manifests to reference `.agenticlens-config.yaml`.
* Quality:
  - Go table-driven unit tests in `loader_test.go` verifying root `.agenticlens-config.yaml` discovery, precedence over legacy paths, and fallback behavior when only legacy files exist.
  - Integration tests verifying `agentic-lens scan` and `agentic-lens generate` parse `.agenticlens-config.yaml` correctly.

*ACCEPTANCE CRITERIA*

* `.agenticlens-config.yaml` at project root is discovered and loaded as the primary configuration file.
* Backward compatibility preserved: legacy `.agenticlens/config.yaml` continues to load if root file is absent.
* Existing repository `.agenticlens/config.yaml` successfully migrated to root `.agenticlens-config.yaml`.
* All references across documentation, `AGENTS.md`, and test fixtures updated to `.agenticlens-config.yaml`.
* `make check` and `make test` pass cleanly with zero errors.

*DEFINITION OF DONE*

* Code, configuration, and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, `make build`).
* Zero any and zero emojis across all files.
* CLI commands tested with `.agenticlens-config.yaml`.
