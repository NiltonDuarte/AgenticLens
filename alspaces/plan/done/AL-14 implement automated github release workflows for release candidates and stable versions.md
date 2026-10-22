h3. Title: AL-14: Implement Automated GitHub Release Workflows for Release Candidates and Stable Versions

*TARGET*: AgenticLens (cli | adapters | build | tests)

*CONTEXT*

After merging code into `main`, developers need automated distribution of release candidates (RC) to test changes before tagging a stable release. When a tag `vX.Y.Z` is pushed, a stable release must be published automatically, accompanied by compiled binary artifacts for target platforms, and a pointer to the `latest` stable release must be updated. Currently, release generation and distribution are completely manual.

*GOAL*

Implement automated GitHub Actions workflows to publish Release Candidate pre-releases on every successful commit to `main`, publish official Stable releases on `v*.*.*` tags, compile multi-platform single binaries, and maintain an updated `latest` release tag and release pointer.

*VERTICAL SLICE SCOPE*

* Boundary Adapter & Build:
  - Add cross-compilation target in `Makefile` (e.g. `make release-build` compiling for Linux amd64, Linux arm64, macOS darwin amd64/arm64) embedding `web/dist` assets.
  - Script or action configuration to package checksums (`sha256sum`) and release tarballs/zip archives.
* Workflows:
  - Create `.github/workflows/release-candidate.yml`:
    - Triggered on push to `main` following successful CI run (or workflow run dependency).
    - Generates pre-release candidate tag (e.g. `v<version>-rc.<run_number>+<short_sha>`).
    - Builds single-binary artifacts and creates a GitHub Pre-release marked as `prerelease: true`.
  - Create `.github/workflows/release-stable.yml`:
    - Triggered on tag push matching `v*.*.*`.
    - Validates that the tag conforms to Semantic Versioning (`vMAJOR.MINOR.PATCH`).
    - Builds multi-platform release binaries embedding the exact tag version.
    - Creates GitHub Release with auto-generated release notes.
    - Updates the floating `latest` git tag and points `latest` release assets to this new stable release.
* Quality:
  - Local verification script or dry-run test asserting build commands and binary artifact packaging without error.
  - Validation that zero emojis exist in workflow names, commit logs, release titles, and documentation.

*ACCEPTANCE CRITERIA*

* Every commit to `main` produces an automated Release Candidate pre-release with attached binary assets.
* Pushing a tag `vX.Y.Z` triggers the stable release workflow, publishing a GitHub Release with compiled binaries.
* The floating tag `latest` is automatically updated on every stable release to reference the newest stable release commit.
* The latest stable release is marked as the official latest release on GitHub Releases.
* Release binaries correctly report their version via `./agentic-lens version` without requiring Node.js at runtime.

*DEFINITION OF DONE*

* Workflows merged to target branch.
* Quality gates passing (`make check`, `make test`, `make build`).
* Workflow syntax verified with GitHub Actions linters/schema.
* Telemetry tested with `agentic-lens` CLI.
