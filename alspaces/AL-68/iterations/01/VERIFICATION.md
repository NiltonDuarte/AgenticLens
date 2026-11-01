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

## Integrated final code checkpoint

Implementation commit `9db4672a971709a7433eeee4e11a5d7d0817cbbe` on feat/AL-68-automatic-session-capture-hooks. Working tree was clean before `git fetch origin main && git merge origin/main --no-edit`; integration reported Already up to date, with no source changes. Required post-integration `make` PASS (exit 0). All Go vet/race packages, frontend typecheck/lint/build, 35 frontend files/781 tests, 6 release revision tests and14capture distribution tests passed; Linux/Darwin artifacts built. Source fingerprint remains `c34cc949278608f4c777f7428a620e2ae29542e7fe33875d5b13ae9f0d17e6ec`. Every web asset matches the embedded tree and exact bytes in the standalone Linux binary. Native handler help/protocol were checked. Raw log remains outsideGit at /tmp/al68-integrated-make.log.

All twelve final focused profile cells approve their corrected code scopes; no code finding remains open. Final overall acceptance remains INCOMPLETE solely for required live Codex trust/callback/agent-binding/native-child/source evidence. No PR publication or remote merge was performed.

## Final telemetry metadata fidelity

Final native source generation captured all35expected Codex sessions and preserved4registered OpenCode smoke sessions. The existing generate command populated ticket completed_at from an invocation/publication timestamp even though acceptance remains pending. Parent corrected only the ticket-owned metadata to null, preserved all source session/invocation facts byte-semantically, updated the report digest/limits, and rerendered the persisted ledger using the native inspection command. No application code or canonical schema changed. This existing default remains a limitation; it is not claimed fixed by AL-68. Independent reviewer session endings do not establish ticket completion.

## User-trusted runtime delivery checkpoint

After the user's normal project/hook trust action, live Codex exec/resume verified independent roots, actual agent binding, native-child source-parent capture, same-ID resumption/growing usage and unchanged snapshot preservation. Five event observations are scoped exactly in CODEX_SMOKE; SubagentStart was not observed and running/interrupted-before-Stop child coverage remains degraded/unverified. Parent verified actual owning source metadata and corrected the evidence child's summary classification/parent only. No program/test or trust-setting changes occurred.

Commit `ba0b50a0d7f64561b21a1d11cd5bf3d909dfe5d7` contains factual support guidance, runtime evidence and existing specification/index lifecycle. Clean-branch main integration reported Already up to date; required `make` PASS (exit0). All Go vet/race, frontend typecheck/lint/build,35files/781tests,6release and14distribution tests, Linux/Darwin builds passed. Runtime code/tests unchanged from reviewed source; only canonical operator documentation differs (DELIVERY_REVISION). Raw log /tmp/al68-trusted-runtime-make.log remains outsideGit. Final telemetry is next and last before final evidence commit.

Final trusted-run telemetry verified42sessions:38Codex including3actual bound smoke sessions, plus4preserved OpenCode sessions. New smoke parent/model/usage exactly match authoritative source and owning metadata; unrelated unassigned session excluded. Source facts unchanged by clearing only legacy inferred ticket completion. Persisted native CLI validation/render, ledger-report digest and payload/live-challenge privacy checks pass. No generation source warnings. Root remains active at snapshot; subsequent finalization calls are outside that cutoff.
