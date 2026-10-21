# Contract: AL-14 Automated GitHub Release Workflows for Release Candidates and Stable Versions

## 1. Problem Statement

AgenticLens currently builds and verifies single binaries in continuous integration via .github/workflows/ci.yml, but release artifact distribution remains entirely manual. Development iterations merged to the main branch lack automated pre-release binary distribution. Downstream users and test environments cannot obtain pre-compiled release candidates (RC) to validate functionality prior to an official release.

Furthermore, pushing release tags requires manual cross-compilation across operating systems and architectures, manual checksum calculations, and manual GitHub Release drafting. There is no automated mechanism to enforce Semantic Versioning on pushed tags, compile cross-platform single binaries embedding web/dist assets, publish release archives, or update the floating latest release pointer.

Ticket AL-14 delivers an automated GitHub Actions release infrastructure. Every merge to the main branch publishes an automated Release Candidate pre-release with multi-platform binary archives. Every valid semantic tag push (vX.Y.Z) publishes an official Stable GitHub Release, cross-compiles production binaries, generates cryptographic checksum manifests, and updates the floating latest tag.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All changes adhere strictly to the boundaries established in TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. GitHub Actions Workflows (.github/workflows/):
   - Act as the outermost automation shell triggering on repository events (push to main, tag creation).
   - Invoke standard build targets defined in Makefile and execute environment validation.
   - Delegate version formatting logic to the Go domain and build flags.
2. Build Automation Shell (Makefile):
   - Coordinates frontend asset pre-compilation (npm run build in web/) and Go compilation.
   - Drives multi-platform cross-compilation using environment variables (GOOS, GOARCH, CGO_ENABLED=0).
   - Injects build-time release variables into the Go domain package through linker flags (-ldflags).
   - Packages distribution archives (.tar.gz and .zip) and generates cryptographic SHA256 digest files.
3. Functional Core Domain Layer (backend/internal/domain/version.go):
   - Remains completely free of network, filesystem, or operating system dependencies.
   - Retains immutable BuildInfo value objects and pure string formatting functions: GetBuildInfo, FormatBuildInfo, and FormatShortVersion.
   - Receives build values injected at compile time into domain.Version, domain.GitCommit, and domain.BuildDate.
4. Imperative Boundary Layer (backend/internal/adapters/web/embed.go & backend/cmd/agentic-lens/main.go):
   - Statically embeds web/dist assets into all target platform binaries via embed.FS.
   - CLI version subcommands and flags (-v, --version, version --short, version --json) report exact build values without external dependencies or Node.js runtime presence.

### Platform and Binary Invariants
1. Pure Go Cross-Compilation:
   - Modern SQLite driver (modernc.org/sqlite) operates without CGO. All target builds run with CGO_ENABLED=0, permitting seamless cross-compilation across Linux, macOS (Darwin), and Windows from a single runner.
2. Target Platform Matrix:
   - linux/amd64: agentic-lens-linux-amd64
   - linux/arm64: agentic-lens-linux-arm64
   - darwin/amd64: agentic-lens-darwin-amd64
   - darwin/arm64: agentic-lens-darwin-arm64
   - windows/amd64: agentic-lens-windows-amd64.exe
3. Single Binary Embedding:
   - All cross-compiled binaries must contain the compiled SPA bundle embedded from backend/internal/adapters/web/dist/.
   - The web build (npm ci && npm run build) must precede all Go compilation steps.
4. Cryptographic Verification:
   - Every release artifact must be cataloged in a SHA256SUMS checksum file generated during packaging.
5. Zero Emojis Mandate:
   - Workflow names, step titles, commit logs, release titles, and release descriptions must be 100% emoji-free.

---

## 3. Scope Boundaries: Go Backend, CI/CD Workflows, and TypeScript Frontend

### In Scope
1. Build Target Modernization (Makefile):
   - Add cross-compilation targets: release-build, release-package, and release-clean.
   - Support parameterized target directory, version injection, and platform matrices.
   - Archive generation for Unix tarballs (.tar.gz) and Windows zip files (.zip).
   - Checksum generation producing SHA256SUMS.
2. Release Candidate Workflow (.github/workflows/release-candidate.yml):
   - Trigger on push to the main branch after successful CI checks.
   - Determine dynamic pre-release version string incorporating semantic base, run number, and short commit hash.
   - Build multi-platform binaries with injected version information.
   - Create and publish GitHub Pre-release marked as prerelease=true.
   - Attach compressed archives and SHA256SUMS manifest.
3. Stable Release Workflow (.github/workflows/release-stable.yml):
   - Trigger on push of tags matching pattern v*.*.*.
   - Validate tag format against strict Semantic Versioning syntax (vMAJOR.MINOR.PATCH).
   - Run verification checks (make check, make test).
   - Build multi-platform production binaries embedding the exact tag version.
   - Create official GitHub Release with auto-generated release notes, marked as prerelease=false and latest=true.
   - Atomically update floating git tag latest to point to the stable release commit.
4. Release Verification and Dry-Run Testing:
   - Unit and shell verification asserting Makefile targets execute cleanly in dry-run/local mode.
   - Ensure compiled binaries accurately report their version via ./bin/agentic-lens version.
5. Repository Documentation (README.md):
   - Document release tiers (Release Candidates vs. Stable Releases).
   - Document release asset structure, cross-platform matrix, and checksum verification steps.

### Out of Scope
1. Container image publishing (Docker Hub or GHCR), which will be addressed in containerization tickets.
2. Homebrew tap or package manager formula updates.
3. Upstream modifications to telemetry schema or UI presentation components.

---

## 4. Release Pipeline Specifications

### Release Candidate Workflow Specification
- Trigger: Push to main branch.
- Version Schema: vX.Y.Z-rc.RUN_NUMBER+SHORT_SHA (e.g. v0.1.0-rc.42+a1b2c3d).
- Prerelease Flag: true.
- Assets:
  - agentic-lens-vX.Y.Z-rc.RUN_NUMBER+SHORT_SHA-linux-amd64.tar.gz
  - agentic-lens-vX.Y.Z-rc.RUN_NUMBER+SHORT_SHA-linux-arm64.tar.gz
  - agentic-lens-vX.Y.Z-rc.RUN_NUMBER+SHORT_SHA-darwin-amd64.tar.gz
  - agentic-lens-vX.Y.Z-rc.RUN_NUMBER+SHORT_SHA-darwin-arm64.tar.gz
  - agentic-lens-vX.Y.Z-rc.RUN_NUMBER+SHORT_SHA-windows-amd64.zip
  - SHA256SUMS
- Release Notes: Automatically generated changelog from commits since prior tag.

### Stable Release Workflow Specification
- Trigger: Push tag matching v*.*.*.
- Validation: Tag must strictly match regular expression ^v[0-9]+\.[0-9]+\.[0-9]+$. Non-conforming tags fail immediately.
- Prerelease Flag: false.
- Make Latest: true.
- Floating Tag: Force update git tag latest to the tagged commit and push to remote.
- Assets:
  - agentic-lens-vX.Y.Z-linux-amd64.tar.gz
  - agentic-lens-vX.Y.Z-linux-arm64.tar.gz
  - agentic-lens-vX.Y.Z-darwin-amd64.tar.gz
  - agentic-lens-vX.Y.Z-darwin-arm64.tar.gz
  - agentic-lens-vX.Y.Z-windows-amd64.zip
  - SHA256SUMS

---

## 5. Acceptance Criteria

1. Merges to the main branch automatically invoke the release candidate workflow and publish a pre-release on GitHub.
2. Pushing a tag vX.Y.Z triggers the stable release workflow, publishing an official GitHub release marked as latest.
3. Pushing an invalid tag format (e.g. v1.0 or v1.0.0-beta) aborts the stable release workflow before artifact publication.
4. The floating latest git tag points to the commit of the most recent stable release.
5. All release archives package single self-contained binaries that execute without external Node.js dependencies.
6. The CLI command agentic-lens version outputs the exact release version, commit hash, and build timestamp injected during compilation.
7. A SHA256SUMS file accompanies every release containing valid cryptographic hashes for all published archives.
8. Makefile includes functional release-build, release-package, and release-clean targets.
9. All workflows, titles, documentation, and logs adhere strictly to the zero-emoji mandate.
10. All existing quality checks (make check, make test, make build) pass without warnings or regressions.
