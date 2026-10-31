# AL-43 Iteration 02 independent testing

## Tested scope

Tested the uncommitted application diff against baseline 9a3796ebadad7f5eb002b764b43c28e0526efb86 on chore/AL-43-revert-content-classification. SHA-256 of `git diff -- backend web/src` was c5fbb4c7413639afa95054d1bfc1bf39e6c7f19e161589af5fb06933a400b307. Parent owns later integration, telemetry, and delivery.

Read the superseding contract, iteration plan, implementation rationale, strict schema boundary, adapted Codex evidence tests, live parity test, and mixed LiveTraceInspector interaction test. Reviewed the removed assertions against the rationale and existing recipient tests. Content-specific classifier, warning propagation, counter, badge and filter assertions are intentionally retired. Ordinary raw-output, source/session identity, execution status/count, enrichment, and interaction oracles remain covered by the recipients named in IMPLEMENTATION.md. No blocking missing retained oracle was found.

Execution classification is restored to pre-PR #78 behavior. In particular the pre-feature classifier retains crash-signature heuristics; this revert does not promise that every completed invocation with arbitrary output text is successful. The added Java/Python/TypeScript cases protect successful execution for those diagnostic-looking outputs, while existing crash/exit/explicit-failure cases protect the restored execution rules.

## Test change

Strengthened the existing `TestOfflineAndLiveToolExecutionParity` rather than adding another fixture or test. It now independently expects three live calls, one failed invocation, completed status for the diagnostic-looking output and grep no-match call, and error status for the explicit failure. Equality alone could previously pass if both extraction paths regressed identically. The existing real temporary SQLite fixture and production live/offline parsing paths are preserved.

The strict retired-ledger compatibility regression exercises metadata validation, nested decoding, DTO validation, domain conversion and both direct/canonical serialization. It retains supported owner/source/execution facts while discarding retired fields, and rejects unrelated unknown fields in both nested tool and event objects. Its fixture is self-contained and requires no live ticket workspace or source hydration.

## Verification

- `go test -race ./internal/adapters/live -run TestOfflineAndLiveToolExecutionParity -count=1` from backend: passed after strengthening its independent expectations.
- `make`: passed, exit 0. All stages executed: dependency installation, frontend production asset preparation, Go vet, frontend typecheck and ESLint, all Go race-tested packages, frontend Vitest (35 files, 741 tests), frontend production build, and Linux/macOS amd64/arm64 binary builds.
- `diff -qr web/dist backend/internal/adapters/web/dist`: passed after the final build. Index, JS and CSS asset trees are identical. `git check-ignore` confirms both generated index paths remain ignored. The full Go race suite includes `adapters/web`'s `TestGetFS` and HTTP static/SPA fallback coverage.
- `./bin/agentic-lens-linux-amd64 version --json` and `./bin/agentic-lens-linux-amd64 help`: passed for the newly built standalone host binary.
- `git diff --check`: passed.
- Targeted search in backend and web/src for classifier/content-error identifiers: only decode-only schema compatibility placeholders and their regression fixture/assertions remain.

Coder focused Go and frontend results recorded in IMPLEMENTATION.md were inspected and reused; the full gate above independently ran all packages and frontend files. No additional application changes were needed. Build output stayed ignored and raw test logs stayed outside Git.

## Limits and existing diagnostics

No verification stage was skipped. Browser geometry, manual browser interaction, and serving production assets from a running standalone process were not tested here; DOM interaction tests, embedded-FS/static-serving tests, identical generated asset trees, and standalone version/help execution establish their stated boundaries. Parent owns any additional standalone serving verification and the required gate after upstream integration.

Dependency installation reports 21 existing npm vulnerabilities (18 moderate, one high, two critical). The frontend suite reports non-failing React act warnings in MetricProgression. Neither required dependency or unrelated product changes for this revert.

## Parent final integration verification

Full `make` passed at integrated commit 4af99a8 after merging origin/main 3857fa2. All Go race suites/vet, frontend typecheck/lint, 741 frontend tests and four platform builds executed successfully. Generated asset trees are identical and ignored. The standalone host binary served byte-identical index, JS and CSS; its temporary server was stopped. Version/help and diff checks passed. This completes the previously pending parent integration gate; browser geometry remains untested. No application changes followed the reviewed integrated diff.
