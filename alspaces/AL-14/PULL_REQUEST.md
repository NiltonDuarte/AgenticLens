# Pull Request: AL-14 Automated GitHub Release Workflows for Release Candidates and Stable Versions

## Description
This pull request implements ticket AL-14, establishing automated GitHub Actions release workflows for release candidates on `main` merges and official stable releases on `v*.*.*` tags, accompanied by multi-platform single-binary cross-compilation in `Makefile`, cryptographic checksum verification, SemVer validation gates, and floating `latest` tag management.

## References
- Architectural Contract: `workspaces/AL-14/CONTRACT.md`
- Detailed Delivery Summary: `workspaces/AL-14/SUMMARY.md`

## Summary of Changes
- **Build System (`Makefile`)**: Added `release-build`, `release-package`, `release-verify`, and `release-clean` targets supporting cross-platform compilation (`CGO_ENABLED=0`) across Linux, macOS, and Windows (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`), automated archive packaging (`.tar.gz`, `.zip`), and `SHA256SUMS` generation.
- **Release Candidate Workflow (`.github/workflows/release-candidate.yml`)**: Automated GitHub pre-releases on pushes to `main` with dynamic SemVer-compliant version string (`vBASE-rc.RUN_NUMBER+SHORT_SHA`), multi-platform binary archives, and release notes.
- **Stable Release Workflow (`.github/workflows/release-stable.yml`)**: Automated official GitHub Releases on `v*.*.*` tag pushes with strict SemVer validation (`^v[0-9]+\.[0-9]+\.[0-9]+$`), release asset publication, and floating `latest` git tag synchronization.
- **Test Infrastructure (`backend/cmd/agentic-lens/release_verification_test.go`)**: Added adversarial tests asserting SemVer regex validation, RC tag syntax, base tag derivation, workflow structure and zero emoji rules, packaging error resilience, and SHA256 checksum verification.
- **Documentation (`README.md`)**: Documented release distribution tiers, platform matrix, checksum verification instructions, and local build commands.
