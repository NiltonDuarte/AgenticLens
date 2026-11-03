# GEN-19 architecture review

Verdict: APPROVED for the assigned architecture and maintenance scope. No blocking correctness or contract finding. One optional maintenance observation, severity 2/10, is recorded below.

## Identity and reviewed candidate

- Phase: independent first pass; local run label architecture-sol. The dispatch did not supply another run ID.
- Responsibility: stable engine reuse, dependency/API compatibility, CI responsibility boundaries, implementation size and upkeep.
- Requested reviewer identity: gpt-6.1-sol / high. Runtime model and reasoning effort are not independently exposed by the inspected evidence; requested identity is not a measured runtime assertion.
- Base: 036d808b0c77bec400f23ca77a5cf056a0b2078b.
- Candidate fingerprint: 3a5ea9f89c48818862332b7da06019145544ec9ab2edcb830cae9a28b752765b. SHA-256 comparison confirmed every file listed in CANDIDATE.json matches its recorded hash.
- Read the assigned persona, shared context, review skill, technical guidelines, release regression checklist, contract, plan and frozen brief. No sibling conclusions were read. Product and tests remained read-only; this report is the only owned write.

## Findings

No blocking findings.

Optional maintenance observation A-SOL-1: align the declared runtime support of the locked graph, severity 2/10, high confidence in metadata and limited confidence about unexercised runtime paths. Locations: scripts/release/package-lock.json:47 and .github/workflows/release-candidate.yml:85. The lock installs @google-automations/git-file-utils 3.2.1 with engines.node >=22; the workflow selects Node 20. Release Please imports this package through its GitHub adapter, so it is a real resolver dependency. The current official action lock uses version 3.0.0 of that transitive dependency with Node >=18. Thus matching the direct engine version does not establish identical transitive runtime support. The supplied Node 20.20.2 test run passed all 13 tests with zero skips, and no concrete supported resolver failure was established; this is not a blocker. On the next dependency update, align the runner or choose a supported transitive version and retain the real-engine Node-version verification. The practical consequence today is unsupported declared runtime metadata and installation warnings, not proven incorrect RC versions. This observation does not request changing stable release policy.

## Consequential hypotheses investigated

- A second version increment implementation could diverge from stable releases. Refuted: rc-version.cjs:50-62 calls Manifest.fromManifest and buildPullRequests with the existing config/manifest and consumes the structured candidate version. Version.parse/compare at lines 11-26 validates and orders versions; it does not choose the bump. Real engine fixtures exercise feature, patch, breaking, commit Release-As and configured release-as behavior. The current [official action v4 lock](https://raw.githubusercontent.com/googleapis/release-please-action/v4/package-lock.json) also resolves release-please 17.3.0. Its [entrypoint](https://raw.githubusercontent.com/googleapis/release-please-action/v4/src/index.ts) loads the same manifest API for configured repositories.
- The library call might mutate release PRs or releases. Refuted for the supported current configuration by inspecting installed release-please/build/src/manifest.js:127-139, 191-420 and 471 onward: candidate construction returns proposals; remote PR creation is a separate createPullRequests path. The resolver never calls it. The current root Go configuration has no custom plugins. HTTP helpers introduced here issue reads only.
- A SHA passed as a branch could produce an unsupported history lookup. Refuted: installed github.js:514 onward sends targetBranch to GraphQL ref(qualifiedName: ...); the resolver supplies main, with explicit checked-out SHA checks and remote-head guards. It does not assume an arbitrary SHA is a valid branch parameter. The documented final guard remains a best-effort read, not an atomic publication transaction.
- New Node tooling could enter the shipped runtime or duplicate build responsibilities. Refuted: the dependency and lock live under scripts/release with ignored node_modules; the workflow installs them only for version resolution. Existing release-build, release-package and release-verify targets remain the build/package authority, and archives contain binaries, README and optional LICENSE. No Go, frontend, stable workflow or production dependency files changed.
- Extra wrappers or repeated validation could create an unnecessary second policy layer. No required simplification found. The helper has a compact boundary: GitHub reads, engine invocation, guards, metadata emission. The initial raw-file checks reject missing/malformed root inputs before library defaults can conceal them; parsed-config checks enforce the supported one-root unprefixed-tag contract. Exported functions are actual resolver/adapter responsibilities used by the CLI and tests, rather than production code existing only for tests. The pinned internal Version import is an intentional upkeep coupling documented in scripts/release/README.md:7.

## Verification and limits

- Executed: Python SHA-256 comparison of all frozen candidate files; every file matched.
- Executed: node --test scripts/release/rc-version.test.cjs on Node v24.22.1; 13 passed, zero failed/skipped.
- Inspected: git diff for workflow/Makefile/Python tests; all new release-tooling files; unchanged stable workflow/config/manifest; installed pinned Manifest, GitHub and dependency metadata; Makefile release packaging and verification consumers.
- Reused parent/tester evidence: make test-release-workflow passed 8 Python tests and 13 Node tests with zero skips; separate Node v20.20.2 real-engine run passed all 13 tests; actual read-only pinned-engine calculation returned 2.1.0 at main/036d808; make build completed with embedded assets. These are supplied results, not runs performed by this reviewer.
- Full integrated make was pending in the frozen brief and remains the parent's delivery gate. No hosted release publication was exercised or authorized. This scoped approval does not certify unrelated behavior or authorize merge/publication. Dependency pins require reevaluation if the floating stable v4 action or release configuration changes; the README records that maintenance responsibility.
