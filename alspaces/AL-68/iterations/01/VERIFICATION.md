# AL-68 verification

## Frozen initial candidate

Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`. Candidate fingerprint: `c7dade3313a1d65a4269d595c782750655be8a044c020dc1f3ac502464cf325a` (CANDIDATE.json). Parent ran `make` on 2026-11-01; exit 0. Content fingerprint unchanged afterward.

The gate installed dependencies, built embedded assets, ran Go vet, frontend strict typechecking/lint, release revision tests (6), capture distribution tests (8), all Go package race tests, all frontend tests (35 files, 781 tests), frontend build and Linux/Darwin native cross-builds. Existing npm audit, bundle-size and React test act warnings were emitted; no failing stage or silently skipped frontend stage.

This pre-integration result does not establish correctness: independent reviewers have reproduced unresolved preservation/ownership/publication defects. Affected checks must run after fixes, and the required make gate must run after main integration. Runtime/platform limits remain in SMOKE.md.

## Corrected candidate 02

Fingerprint `17fdae376119de798b382af56952351564389a69a174e5b0049259d387151cbc`: parent full `make` PASS (exit 0), including all Go race/vet packages, frontend typecheck/lint, 35 frontend files/781 tests, 6 release revision tests, 14 capture distribution tests, frontend embedding/build and Linux/Darwin native builds. Fingerprint unchanged afterward.

Compared every web/dist asset with backend/internal/adapters/web/dist byte-for-byte and verified its exact bytes occur in the built Linux amd64 executable: index.html, assets/index-Xjo2aE-x.css, assets/index-C6E9fpfD.js. This establishes the generated asset tree embedded by this build. Windows amd64 and Darwin amd64 corrected lock cross-builds passed per IMPLEMENTATION.md; native non-Linux execution remains unverified.

This is pre-integration evidence, reused by corrected-focus reviewers. Required main integration/make and Codex real trust/callback/agent evidence remain pending.

## Final correction candidate 03

Parent `make` PASS (exit 0); unchanged source fingerprint `26898644f86547fb9013c161120b291c1c6ed7f130307f214351b25c1741534d`. All Go vet/race packages, 35 frontend files/781 tests, frontend typecheck/lint/build, 6 release-revision tests, 14 capture-distribution tests and native Linux/Darwin builds passed. Raw command log remains outside Git at /tmp/al68-candidate03-make.log. Earlier source model/runtime limits remain in SMOKE.md.

Fetched origin/main after this check; it has no commits beyond original baseline HEAD. No integration or source mutation performed by that fetch; feature-branch merge/check phase remains recorded separately when performed. Codex live human-trust smoke is still pending, so this build result is not final delivery approval.
