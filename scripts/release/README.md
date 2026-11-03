# Release candidate version resolution

This directory contains CI-only Node tooling. It adds no runtime dependency to the distributed AgenticLens binary.

`rc-version.cjs` uses Release Please's structured `Manifest.fromManifest(...).buildPullRequests()` API with the repository's existing configuration and last-released-version manifest. It never opens or updates a pull request. Release Please selects patch, minor, major and explicit release versions; Git tags and open PR titles do not choose the upcoming version.

The dependency is pinned to `release-please` 17.3.0, matching the engine in the official `googleapis/release-please-action` v4 package-lock when this resolver was introduced. When updating the stable action or its versioning configuration, check that this pin and its regression fixtures still match the stable engine. The public `buildPullRequests` candidate version and internal `Version` import are verified against this pin; do not upgrade without running the release tests.

Install with `npm ci --prefix scripts/release --ignore-scripts`, then run `make test-release-workflow`. The tests exercise the real pinned versioning engine with isolated GitHub read fixtures, and separately test API/publishing guards. They do not publish a release.

The workflow invokes `node scripts/release/rc-version.cjs resolve` with GitHub's normal repository, run number, output and environment variables, plus `GH_TOKEN` and `EXPECTED_SHA`. The token requires contents and pull-request read access for version calculation; only the existing publish step needs contents write. `EXPECTED_SHA` must be the full checked-out commit SHA. Resolve writes the validated RC metadata to `GITHUB_ENV` and `should_publish` to `GITHUB_OUTPUT`.

Immediately before publication, `node scripts/release/rc-version.cjs verify` rechecks `main` and stable releases using the saved numeric `RC_TARGET`. Both commands only perform GitHub reads. A superseded commit or no release-worthy changes skips explicitly. Invalid configuration/version output, unavailable API data, ambiguous proposals, or a target already published as stable fails the job before publication. No guessed next-version fallback exists.

Checks around branch calculation and before publication narrow the window for concurrent changes. GitHub provides no transaction covering an independent stable release workflow and prerelease publication, so the final read cannot make those operations atomic.
