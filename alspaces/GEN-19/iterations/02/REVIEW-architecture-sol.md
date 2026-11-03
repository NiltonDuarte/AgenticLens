# GEN-19 iteration 02 architecture review

Verdict: APPROVED for the assigned architecture and maintenance scope. No findings requiring changes; no numeric finding severities apply.

Run: rc2_architecture_sol; phase: independent first pass. Requested model/effort: gpt-6.1-sol/high; runtime model metadata was not independently available. Base: e639f3303c9d92236edce85befccda66617d7cb8. Frozen fingerprint: 903c512467c9100cbb6b80756e501280efb9a3ab08f3466c0f4890cae92ece1e. Independently verified every candidate file hash and deletion against CANDIDATE.json.

Reviewed the accepted contract supersession and plan, RC workflow diff, Makefile and contributor guidance, deleted calculator/dependency entry points, replacement tests, and relevant unchanged Release Please workflow/configuration/manifest. Product code remained read-only. No sibling persona or report informed this pass.

The authority boundary is clear: Release Please computes and writes the next root version; the RC workflow only validates that stored value and derives the existing RC suffix. There is no second bump policy or runtime dependency. The generated branch name matches the repository's actual root release PR and current package configuration. Hardcoding that integration identity is understandable within this repository-specific workflow; there is no demonstrated need for another configuration or abstraction.

Investigated three consequential hypotheses:

- Deleting scripts/release might leave active consumers or require its dependencies elsewhere. Targeted searches across workflows, scripts, Makefile, contributor documentation, active agent guidance and docs found no remaining calculator/package references. The existing release test target now directly runs Python.
- A nearby but unrelated PR could become another version authority. Selection requires the same repository, generated branch, main base and pending label; its immutable head must have exactly the source commit as parent. Metadata and final gates remain separate responsibilities. Repeated current-main/stable checks are justified by the build interval, rather than duplicate version calculation.
- Local fixtures could conceal unsupported real gh filters or generated commit structure. Executed the unmodified production metadata shell with real gh 2.45.0/GitHub reads in a disposable local clone at current main 036d808. It exited zero, selected PR 55/head 1300741296d069b1282d89a6192219bf13fbb4f9, and emitted should_publish=true, RC_TARGET=2.1.0 and VERSION=v2.1.0-rc.999999999+036d808. Only the metadata shell ran; no release was published.

Verification: make test-release-workflow passed all 13 tests without skips. Source tracing confirms build metadata and final publication/pruning guards remain connected. Full integrated make is the parent's pending gate. Hosted workflow scheduling, artifact packaging/upload and cross-workflow atomicity were not established by this review; the accepted contract explicitly treats publication checks as best effort.
