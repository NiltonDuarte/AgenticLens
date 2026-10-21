h3. Title: AL-13: Implement Semantic Versioning and Automated CI for Pull Requests and Main Merges

*TARGET*: AgenticLens (domain | adapters | presentation | cli | tests)

*CONTEXT*

The AgenticLens project currently lacks automated CI on GitHub Actions for pull requests and main branch merges. Without automated validation running on every PR, regressions in Go race conditions, TypeScript compilation, code formatting, or single-binary packaging can easily slip into `main`. Furthermore, the software has no compiled-in semantic versioning, preventing users and downstream pipelines from querying the active release version via CLI or API.

*GOAL*

Deliver software semantic versioning with build-time `-ldflags` injection, CLI version commands, HTTP version endpoints, and web UI version indicators, backed by a GitHub Actions CI workflow that enforces code checks, tests, and binary build verification on pull requests and pushes to `main`.

*VERTICAL SLICE SCOPE*

* Domain Core: Create `internal/domain/version.go` defining `BuildInfo` struct (`Version`, `GitCommit`, `BuildDate`) and version string formatting functions.
* Boundary Adapter:
  - Expose `/api/version` GET route in `internal/adapters/http/` returning structured JSON `BuildInfoDTO`.
  - Update `cmd/agentic-lens/main.go` to support `agentic-lens version` subcommand and `-v`/`--version` flags.
  - Update `Makefile` to pass `-ldflags "-X ..."` for version, git commit, and build date.
* Presentation:
  - Add version badge/indicator in `web/src/components/Navbar.tsx` displaying the current application version queried from `/api/version`.
  - Provide fallback display when running in offline or embedded preview mode.
* CI Workflow:
  - Create `.github/workflows/ci.yml` running on `pull_request` targeting `main` and `push` to `main`.
  - Matrix / jobs for Go 1.23+ and Node.js 20+: runs `make check`, `make test-backend` (`-race`), `make test-frontend`, and `make build`.
* Quality:
  - Table-driven Go unit tests for version string generation and API response serialization.
  - Vitest component test for Navbar version display.
  - Validation that zero emojis and zero `any` types exist across added code.

*ACCEPTANCE CRITERIA*

* Running `agentic-lens version` or `agentic-lens --version` prints the semantic version, git commit SHA, and build timestamp.
* Querying `GET /api/version` returns status 200 with `{ "version": string, "commit": string, "build_date": string }`.
* The TypeScript web dashboard displays the active semantic version in the navigation header.
* `.github/workflows/ci.yml` is configured and passes all steps: `make check`, `make test`, and `make build`.
* Pull requests targeting `main` and pushes to `main` trigger the CI workflow and require all checks to pass before merging.
* `go test -race ./...` and `go vet ./...` pass with zero errors.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, `make build`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
