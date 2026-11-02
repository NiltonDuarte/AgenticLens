# GEN-15 consolidated review

Verdict: APPROVED for the corrected installation/distribution scope.

## Revision and ownership

Base: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`. Initial candidate: `d7a07eb40f1a66b98da7f6bd9252e5a8c434f1ad`. Corrected and integrated candidate: `2d540946c01851face786f15894b1e00b2451b19`. Both main integrations were already up to date. Later changes are parent-owned delivery artifacts and telemetry only; no product or test edits after the final review candidate.

The user approved the corrected contract, explicitly retaining owner control of Git tracking/ignore choices. Planner, coder, tester and reviewers had separate writable paths. The approved narrow shell-installation exception combines behavior, failure handling, dependency architecture and test effectiveness; no domain, persistence, extraction, API or UI behavior changed. Parent owns consolidation, integration and publication.

## Review roster and attribution

| Run | Reviewer profile | Phase / revision | Verdict | Evidence |
| --- | --- | --- | --- | --- |
| SOL-01 | reviewer / gpt-6.1-sol, high | Discovery / d7a07eb | CHANGES_REQUESTED | [Report](REVIEW_SOL.md) |
| LUNA-01 | reviewer / gpt-6-luna, medium | Discovery / d7a07eb | APPROVED, artifact correction | [Report](REVIEW_LUNA.md) |
| SOL-02 | reviewer / gpt-6.1-sol, high | Re-review / 2d54094 | APPROVED | [Report](REVIEW_SOL_R2.md) |
| LUNA-02 | reviewer / gpt-6-luna, medium | Re-review / 2d54094 | APPROVED | [Report](REVIEW_LUNA_R2.md) |

All four runs used fresh native-agent contexts with explicit model/effort settings. The parent independently checked their real Codex session metadata and first turn contexts: observed model/effort matches every requested profile. Individual reviewers correctly distinguish their own unavailable metadata from requested settings; the source-backed attestation belongs to the parent's [run records](REVIEW_RUNS.json). Per-run cost/end times were not inferred. Raw transcripts remain outside Git.

## Findings and disposition

- GEN15-01, P2, discovered by SOL-01 / Sol 6.1 high: a retained non-executable binary made the recommended repair falsely succeed. Fixed by restoring and checking executable permission without replacing bytes. The exact recommended-repair regression passes; SOL-02 additionally injected a failing chmod and confirmed failure is surfaced.
- GEN15-02, P2, discovered by SOL-01 / Sol 6.1 high: an owner-excluded release marker made a new worktree silently select latest for pinned scripts. Fixed by requiring explicit remote release selection before writes/network when the marker is unavailable. Offline supplied-prebuilt repair remains supported; the contributor entrypoint retains its explicit historical default. Real-worktree and override-precedence probes pass.
- GEN15-DOC1, low, discovered by LUNA-01 / Luna 6 medium: stale pending-approval wording contradicted the current approval record. Parent corrected it; both re-review profiles confirmed the current plan.

Both runtime regressions fail against the initial candidate in disposable bundles and pass after correction. Both re-review profiles verified the fixes; no unresolved candidates or additional confirmed regressions remain. Discovery credit remains separate from verification credit in [FINDINGS.json](FINDINGS.json). These few scoped runs do not establish comparative model quality or recall.

## Verification

- Parent `make install`: passed; no dependency or lockfile changes. Existing npm audit advisories are outside this ticket.
- Parent `make backend-check`: passed, including canonical asset preparation, Go vet and race tests.
- Parent integrated `make`: passed after each integration; final run targets 2d54094. Final evidence: six release-workflow tests, twelve installation tests, Go vet/race suites, frontend typechecking/lint, 35 frontend test files / 781 tests, canonical frontend build/embed preparation and Make's platform builds. No test stages skipped. Raw local log: `/tmp/gen15-make-final.log` (outside Git).
- Changed shell syntax checks and `git diff --check`: passed. [Tester report](TEST_REPORT.md) records boundary assertions and before/after controls.
- Parent native smoke: installed the complete local bundle and actual freshly compiled Linux amd64 binary in a separate consumer directory containing spaces. The new launcher returned commit `2d54094` via version JSON; no config or ignore files were created. This is local-build integration evidence, not an official release download.
- Generated frontend and embedded index contents match. Generated assets and binaries remain outside this ticket's tracked deliverables according to this source repository's existing Git policy.
- No GEN-15 specification or index entry existed under alspaces/plan; no synthetic lifecycle entry was created.

## Limits and delivery gate

Linux amd64 is the executed host. Native macOS, alternate CPU, Windows/Git Bash and PowerShell execution were not performed. Platform URL/dry-run checks and cross-compilation do not establish those native executions. No live release download or checksum verification is claimed. Bootstrap requires a release containing the new installer; this PR does not publish a release. There are no Go/telemetry extraction changes in this ticket.

Final Codex telemetry extraction and source/child completeness verification are recorded in SUMMARY.md at delivery. This approval permits publication under repository policy, not merging.
