# Delivery, configuration and fixture audit

Baseline d5227812b09a17ff15d7e6ffe058e96ecf1616eb. All five shell scripts, Makefile, Dockerfile, four workflows, Python compatibility provider and tracked runtime/build/agent configuration were read. Wrapper copies were compared byte-for-byte (identical). Supporting data fixtures were parsed/structurally inspected and their consumers traced; historical fixture prose is data, not endorsed current guidance. No source/test changes.

## D1 - P2: update reports success when every artifact download fails

Locations: scripts/agentic-lens-wrapper.sh:147-175 and identical setup-skill wrapper. Download failures only emit warnings; the final path always stores the target version and exits zero. A failed update therefore appears successful to automation and writes misleading update state.

Confirmed in an isolated directory with an unchanged wrapper, benign existing version fixture and failing curl stub: exit0, "Successfully updated 0 binary(ies)", target cache v99.0.0 while the old binary remains. See DELIVERY_REPRODUCTIONS.json. No network calls or installed binaries touched. Fix by preserving failed/partial update outcome, returning nonzero for failures, and updating state only after successful replacement. Test total failure, partial failure and successful atomic replacement. Reuse the installer failure accounting rather than duplicate another updater policy.

## D2 - P2: downloader claims verification without checking artifact integrity

Locations: scripts/agentic-lens-wrapper.sh:73-87,228-232; scripts/download-agentic-lens.sh:280-300 and their setup-skill copies. Success means HTTP utility returned zero and a file exists, followed by chmod/execute; no checksum, size or executable-format verification occurs. The wrapper explicitly says "verified". Releases already publish SHA256SUMS, but installers never consume it.

A benign arbitrary shell payload returned by a local curl stub was installed and executed as the platform binary; captured output is in DELIVERY_REPRODUCTIONS.json. This demonstrates absent validation, not exploitation of GitHub/TLS or a real compromise. Verify the selected release artifact against its published digest before replacement and preserve the old binary on mismatch; do not claim this removes trust in the publisher. Test empty/wrong/mismatched artifacts and failed verification. Wrapper fallback from repository main is also unpinned relative to an explicit binary version; keep wrapper and artifact provenance aligned.

## D3 - P1: force_tag release rebuild can publish HEAD code under an older tag

Locations: .github/workflows/release-please.yml:31-35,66-87,97-103. workflow_dispatch accepts a force_tag to attach assets to an existing release, but checkout does not select that tag. Metadata names the requested tag while GIT_COMMIT is taken from the checked-out workflow revision. Dispatching on main with an older force_tag can build current main and upload it under that older version.

This is source-level control-flow evidence; no release or GitHub workflow was executed. Resolve and checkout the requested tag before tests/build, validate that the built commit matches it, and avoid the fallback invented tag. Verify dispatch against two distinct fixture commits/tags before any real release publication. The normal tag-triggered stable workflow has a separate event path; do not infer it shares this exact mismatch.

## D4 - P2: compatibility harness depends on historical ticket artifacts

Locations: scripts/opencode-compat.sh:20,25,27 and alspaces/AL-58/evidence/mock_openai_server.py. Maintained Make targets execute a provider and configs inside alspaces/AL-58. Renaming/removing archival ticket evidence breaks a supported runtime check, contrary to workspace independence. Fixed ports plus checking only TCP accept can also mistake another listener for the provider; readiness exhaustion does not fail before launching the runtime.

Move provider/config fixtures to an owned stable test-fixture directory, parameterize a free port and verify the launched child/expected protocol. Keep AL-58 historical evidence unchanged. This is source/path evidence; actual OpenCode v1/v2 binaries and provider sessions were not run during this audit. The mock provider is intentionally test-only code, not an unused production helper.

## D5 - P2: checksum failures can leave a successful partial release manifest

Locations: Makefile:164,172-183. SHA256SUMS generation pipes sha256sum into sort, so the pipeline returns sort's success even when hashing fails. release-verify checks only entries present in the manifest and executes the host binary; it does not compare manifest coverage against expected release artifacts.

Fault-injection reproduction on an unchanged Makefile in a disposable directory: hash stub emits the first valid archive hash then exits1; release-package exits0 and release-verify also exits0, although the raw binary has no manifest entry. See PACKAGING_REPRODUCTION.json. This proves the failure boundary, not actual hosted release corruption. Check the hashing command exit status and expected artifact inventory before writing/publishing a manifest. Verify a missing/read-failed member fails packaging and verification. An initial reproduction using an absolute RELEASE_DIR hit the existing './absolute-path' host-run issue; the final reproduction uses the supported relative directory and isolates the checksum defect.

## D6 - P3: duplicate installer implementations require fixes in four files

Locations: scripts/agentic-lens-wrapper.sh, scripts/download-agentic-lens.sh, .agents/skills/agentic-lens-setup/scripts/{agentic-lens-wrapper.sh,download.sh}. Wrapper copies are identical; downloaders duplicate nearly all platform/option/download logic with different path defaults. D1/D2 require coordinated edits across copies. Both installation destinations are legitimate; deleting one deployment surface would break standalone skill installation.

Use one canonical implementation and generate/package the destination copy, or check deterministic equivalence for the intentionally different defaults. This is maintainability debt, not a separate runtime failure. No generalized downloader framework is needed.

## D7 - P2: configured telemetry patterns count frontend fixtures as production

Location: .agenticlens-config.yaml:108-125, consumed by backend/internal/adapters/git/classifier.go:129-147. Explicit tests cover test suffixes, testdata and __tests__, but omit web/src/test/**. The production web/src/**/*.ts pattern then classifies mockData.ts and setup.ts as production.

Confirmed using unchanged copied production ClassifyFile/LoadConfig with an isolated copy of repository configuration: mockData.ts=prod, setup.ts=prod, client.test.ts=test. See CLASSIFICATION_REPRODUCTION.json. These are legitimate test support files; the issue is metric classification, not a reason to delete them. Add the intended fixture/setup patterns and assert non-default configured classification through the actual classifier. Other repositories keep their explicit pattern choices.

## Additional bounded observations

CI/release workflows install Go1.23 while go.mod requires Go1.26; automatic toolchain selection can make them pass, but labels/cache setup obscure the compiler actually used. Align workflow setup with the module toolchain rather than claim a tested Go1.23 minimum. Docker uses Go1.26 and has no production Node runtime. Its npm-ci-or-install fallback can conceal lockfile drift; prefer failing reproducibly on lockfile mismatch. These are delivery hygiene concerns, not observed build failures.

Workflow source review does not prove release publication/download/host execution. No remote workflow, release, or compatibility-provider session was executed. Existing full make at49cf350 passed781frontend tests with identical application source. Shell syntax, Python AST, JSON/YAML/TOML parsing and the targeted local reproductions passed. No dependency vulnerability database or performance profile was consulted; no unsupported current-version/security-advisory claims are made.
