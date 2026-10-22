# fix(ci): automate binary cross-compilation and archive attachment in Release Please workflow

## Summary of Changes
- Integrated binary cross-compilation, packaging, and asset uploading into `.github/workflows/release-please.yml` so that all future releases created by Release Please automatically attach multi-platform executables, archives (`.tar.gz`, `.zip`), and `SHA256SUMS`.
- Added `workflow_dispatch` trigger with `force_tag` input to `.github/workflows/release-please.yml` to allow retroactive asset packaging and publishing on demand.
- Updated `.release-please-config.json` with `"include-component-in-tag": false` so tags follow standard `vX.Y.Z` semantics rather than component prefixes.
- Updated the release workflow to maintain the floating `latest` tag and purge stale release candidates upon official releases.

## Verification
- Executed `make check` (Go vet, TypeScript typecheck, ESLint) - passed with zero errors.
- Executed `make test` (backend unit and race detector tests, vitest suite with 728 assertions) - all passed.
- Executed `(cd backend && go test -race ./cmd/agentic-lens/...)` to verify all release workflow contract tests in `release_verification_test.go` succeed.
- Verified local release build, package, and verification commands (`make release-build release-package release-verify VERSION="v1.0.0"`) pass cleanly.
