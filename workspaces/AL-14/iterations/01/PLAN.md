# Implementation Plan: AL-14 Automated GitHub Release Workflows

## 1. Goal and Iteration Scope

Implement automated GitHub Actions release workflows for release candidates and official stable versions, accompanied by Makefile cross-compilation and packaging targets, version validation gates, SHA256 checksum generation, floating latest tag management, and repository release documentation. This iteration delivers the full vertical slice for ticket AL-14.

---

## 2. Grounded Reference Architecture and Concrete Paths

The implementation grounds against the following repository components:

### Domain Core Layer (backend/internal/domain/)
- backend/internal/domain/version.go: Reference for package variables Version, GitCommit, BuildDate, and pure functions GetBuildInfo, FormatBuildInfo, FormatShortVersion. Ensure linker flags in Makefile accurately target agenticlens/internal/domain symbols.

### Boundary Shell & CLI (backend/cmd/agentic-lens/, backend/internal/adapters/web/)
- backend/cmd/agentic-lens/main.go: Entrypoint for CLI flag handling and version subcommands.
- backend/internal/adapters/web/embed.go: Static SPA asset embedding root via embed.FS.
- web/package.json & web/src/: Source of frontend bundle compiled into backend/internal/adapters/web/dist/.

### Build Automation Shell (Makefile)
- Makefile: Extend existing build targets (all, check, test, build) with new cross-compilation and distribution targets: release-build, release-package, release-clean.

### GitHub Actions Workflows (.github/workflows/)
- .github/workflows/ci.yml: Reference for runner environments, Go 1.23 setup, Node.js 20 caching, and validation commands.
- .github/workflows/release-candidate.yml: New workflow automating release candidate pre-releases on push to main.
- .github/workflows/release-stable.yml: New workflow automating official stable releases on v*.*.* tag pushes with SemVer validation and floating latest tag update.

### Documentation Layer (README.md)
- README.md: Document automated release mechanics, platform matrix, manual verification steps, and downloading release binaries.

---

## 3. Phased Implementation Tasks

### Phase 1: Build System Cross-Compilation and Packaging (Makefile)
- Task 1.1: Define build parameter defaults in Makefile:
  - Define RELEASE_DIR pointing to dist.
  - Define PLATFORMS list containing: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64.
  - Retain existing LDFLAGS injection targeting agenticlens/internal/domain.Version, GitCommit, and BuildDate.
- Task 1.2: Implement release-build target:
  - Ensure frontend assets are built first if web/node_modules exists.
  - Create output platform directories under RELEASE_DIR/bin.
  - Loop through each platform in PLATFORMS setting GOOS and GOARCH.
  - Enforce CGO_ENABLED=0 for static standalone binary generation.
  - Append .exe extension for Windows target; produce standard executable names for POSIX targets.
- Task 1.3: Implement release-package target:
  - Package each compiled binary into its archive: tar -czf for Linux/Darwin platforms and zip -j for Windows platform.
  - Include LICENSE and README.md inside each archive.
  - Place packaged archives into RELEASE_DIR.
  - Generate SHA256SUMS file containing sha256 digests of all generated archive files.
- Task 1.4: Implement release-clean target:
  - Remove RELEASE_DIR and temporary cross-compilation artifacts.

### Phase 2: Release Candidate Workflow (.github/workflows/release-candidate.yml)
- Task 2.1: Define workflow triggers:
  - Trigger on push to branch main.
  - Set workflow concurrency group to avoid conflicting simultaneous release candidate publishing runs.
- Task 2.2: Define release-candidate job:
  - Runner: ubuntu-latest.
  - Permissions: contents: write (required for creating GitHub releases and pushing tags).
- Task 2.3: Check out repository:
  - Use actions/checkout@v4 with fetch-depth: 0 to obtain full git history and tag metadata.
- Task 2.4: Set up runtime environments:
  - Set up Go 1.23 using actions/setup-go@v5 with backend/go.sum caching.
  - Set up Node.js 20 using actions/setup-node@v4 with web/package-lock.json caching.
- Task 2.5: Install dependencies and verify:
  - Install web dependencies via npm ci in web/.
  - Run quality gates: make check and make test.
- Task 2.6: Compute candidate release metadata:
  - Derive base version from latest semantic tag or fallback to default version v0.1.0.
  - Format release candidate version string: vBASE-rc.GITHUB_RUN_NUMBER+GITHUB_SHA_SHORT.
  - Export computed version, commit SHA, and ISO 8601 build date to GitHub Actions environment.
- Task 2.7: Build and package multi-platform release artifacts:
  - Invoke make release-build followed by make release-package with VERSION, GIT_COMMIT, and BUILD_DATE set.
- Task 2.8: Publish GitHub Pre-release:
  - Use softprops/action-gh-release@v2 (or gh release create CLI).
  - Configure prerelease: true.
  - Set tag_name and name to the computed release candidate version string.
  - Enable automatic release note generation (generate_release_notes: true).
  - Attach all archives (.tar.gz, .zip) and the SHA256SUMS manifest.

### Phase 3: Stable Release Workflow (.github/workflows/release-stable.yml)
- Task 3.1: Define workflow triggers:
  - Trigger on push of tags matching v*.*.*.
- Task 3.2: Define release-stable job:
  - Runner: ubuntu-latest.
  - Permissions: contents: write.
- Task 3.3: Semantic Versioning Validation Step:
  - Inspect GITHUB_REF_NAME against strict regex pattern: ^v[0-9]+\.[0-9]+\.[0-9]+$.
  - If validation fails, log informative error message to standard error and exit with non-zero status.
- Task 3.4: Set up Go and Node.js environments with dependency caching.
- Task 3.5: Execute quality gates:
  - Run make check.
  - Run make test.
- Task 3.6: Compile and package release artifacts:
  - Set VERSION to the exact tag name (GITHUB_REF_NAME).
  - Invoke make release-build and make release-package.
- Task 3.7: Publish Stable GitHub Release:
  - Use softprops/action-gh-release@v2 (or gh release create).
  - Configure prerelease: false.
  - Configure make_latest: true.
  - Set release name to GITHUB_REF_NAME.
  - Enable automatic release note generation.
  - Attach all platform archives and the SHA256SUMS manifest.
- Task 3.8: Floating latest tag synchronization:
  - Configure git identity in runner.
  - Force update local tag latest to point to GITHUB_SHA.
  - Force push tag latest to origin.

### Phase 4: Local Verification and Dry-Run Tooling
- Task 4.1: Add dry-run target or verification script:
  - Ensure make release-build and make release-package succeed locally.
  - Verify that produced binaries run and print expected version metadata via ./dist/bin/agentic-lens-linux-amd64 version --json.
  - Verify SHA256SUMS validity via sha256sum -c SHA256SUMS.
- Task 4.2: Ensure clean execution of make release-clean to reset workspace state.

### Phase 5: Documentation Update (README.md)
- Task 5.1: Document release distribution channels in README.md:
  - Release Candidates: Generated automatically on every merge to main.
  - Stable Releases: Published when semantic tags (vX.Y.Z) are pushed.
  - Latest Release Pointer: The latest tag and GitHub release always reflect the most recent stable version.
- Task 5.2: Document cross-platform installation and verification:
  - List of supported architectures (Linux amd64/arm64, macOS amd64/arm64, Windows amd64).
  - Instructions for verifying integrity using SHA256SUMS.
  - Zero-emoji enforcement across all added documentation sections.

---

## 4. Execution Dependencies and Delivery Sequence

1. Phase 1 (Makefile cross-compilation and packaging targets) establishes the local build foundation.
2. Phase 2 (.github/workflows/release-candidate.yml) implements the automated pre-release pipeline for main branch merges.
3. Phase 3 (.github/workflows/release-stable.yml) implements the official tagged release pipeline with SemVer validation and floating tag management.
4. Phase 4 (Local verification and dry-run validation) verifies packaging integrity, binary execution, and checksum calculation.
5. Phase 5 (Documentation in README.md) updates operational instructions for users and contributors.

---

## 5. Automated Verification, Dry-Run Testing, and Quality Strategy

- Makefile verification:
  - Execute make release-build and assert creation of five platform binaries.
  - Execute make release-package and assert creation of archive files and SHA256SUMS.
  - Run sha256sum -c dist/SHA256SUMS and assert all files pass integrity checks.
  - Execute make release-clean and assert removal of dist directory.
- Binary version verification:
  - Execute built binary with version, version --short, and version --json to verify correct metadata injection.
- Workflow linting and syntax validation:
  - Validate release-candidate.yml and release-stable.yml syntax against GitHub Actions schemas.
  - Verify permissions, trigger conditions, step ordering, and environment variable bindings.
- Invariant audits:
  - Zero emojis across workflow files, step names, commit messages, and documentation.
  - Zero CGO dependencies (CGO_ENABLED=0 enforced on all targets).
  - Single self-contained binary packaging with embedded web/dist assets and zero Node.js runtime requirement.
