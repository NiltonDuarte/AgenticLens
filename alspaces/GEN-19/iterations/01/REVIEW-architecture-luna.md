# Architecture review — GEN-19

**Outcome: APPROVED** for the assigned architecture and maintenance scope. No blocking architecture finding.

## Reviewed scope and identity

- Candidate base: `036d808b0c77bec400f23ca77a5cf056a0b2078b`
- Frozen candidate fingerprint: `3a5ea9f89c48818862332b7da06019145544ec9ab2edcb830cae9a28b752765b`
- Scope: Release Please engine reuse, dependency pin/API compatibility, CI-only boundary, and upkeep.
- Persona: architecture and maintainability reviewer; requested profile `gpt-6-luna / medium`. Runtime model metadata was unavailable, so this records the requested profile only.

## Findings

None. No introduced architecture defect in the assigned scope requires a change.

## Hypotheses investigated

1. **The resolver could drift from stable release policy.** The resolver consumes `.release-please-config.json` and `.release-please-manifest.json` from `main`, then uses `Manifest.fromManifest(...).buildPullRequests()` and takes the structured candidate version. The existing stable workflow uses `googleapis/release-please-action@v4` with those same files. The CI dependency is pinned to `release-please` 17.3.0; its package API exposes both methods, and the `Version` deep import resolves in that pin. The README records the synchronization obligation when the stable action or configuration changes. The checked in code therefore uses a single version authority today. Residual maintenance risk: the stable workflow references a floating action major while the resolver pins a library version, so a future action update could change the stable engine independently. Keep the versions aligned during stable action updates; this is upkeep guidance, not a current blocker.

2. **The Node tooling could leak into production or violate the CI runtime.** The resolver and its exact dependency tree are under `scripts/release`; workflow execution installs them only in the release-candidate job after Node 20 setup. The production build targets and Go embed path are not changed by this dependency. This keeps Node as a CI tool and preserves the self-contained binary boundary.

3. **The dependency/API may be incompatible with the declared runner.** The current dependency is exact-pinned rather than ranged. The parent supplied a Node 20.20.2 run of the real engine tests (13 passed, 0 skipped). I independently installed from the lockfile on local Node 24.22.1, confirmed `Manifest.fromManifest`, `GitHub.create`, and the `release-please/build/src/version` import resolve, and ran the 13 Node tests successfully with zero skips. The API is verified for the pin and declared Node 20 environment by the supplied run.

## Verification and limits

- `npm ci --ignore-scripts --no-audit --no-fund` in `scripts/release`: passed; lockfile installed 177 packages.
- `node --test scripts/release/*.test.cjs`: 13 passed, 0 failed, 0 skipped on Node 24.22.1.
- Supplied verification: same real-engine test suite on Node 20.20.2, 13 passed, 0 skipped.
- Reviewed `.release-please-config.json`, `.release-please-manifest.json`, `.github/workflows/release-please.yml`, `.github/workflows/release-candidate.yml`, `scripts/release/package.json`, `scripts/release/README.md`, and the resolver/dependency changes.
- `git diff --check`: passed.
- Hosted GitHub release publication was not exercised; the review does not certify hosted publication behavior.

This approval is limited to architecture/maintenance for the recorded candidate fingerprint. It does not approve the full change or consolidate other review focuses.
