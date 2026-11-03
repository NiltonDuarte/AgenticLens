# GEN-19 iteration 02 behavior/integration review

Verdict: APPROVED within assigned scope. Findings: none; no severity assigned.

Run identity: rc2_behavior_sol, independent first pass. Requested model/effort: gpt-6.1-sol/high; actual runtime model metadata is unavailable. Base: e639f3303c9d92236edce85befccda66617d7cb8. Candidate fingerprint: 903c512467c9100cbb6b80756e501280efb9a3ab08f3466c0f4890cae92ece1e. Independently checked every listed file hash/deletion against CANDIDATE.json: all matched.

Reviewed RC trigger, checkout/source detection, manifest selection/provenance, metadata emission, eligibility and build/publication/prune wiring; traced unchanged Release Please configuration/workflow, stable release workflow and Makefile consumers. Applied accepted iteration 02 supersession. No sibling reports were read; product/tests remained read-only.

Consequential hypotheses investigated:

- Wrong source after workflow chaining: workflow name matches the existing producer; completion, main-branch filter, success and same-repository condition gate automatic execution. Both checkouts explicitly use upstream head_sha rather than workflow_run's default-branch github.sha. Manual execution uses its selected SHA, then requires current main. Release-commit skipping now reads the checked-out commit. These semantics agree with [GitHub's workflow_run documentation](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#workflow_run).
- Stale or unrelated manifest: selection requires exactly one same-repository pending root release PR, expected branch/base, a single parent equal to the source, and manifest retrieval at immutable PR_HEAD. Invalid identities/versions/API failures fail closed; absent/stale candidates skip explicitly. Published stable versions are compared numerically, independently of prior RC tags.
- Incorrect output or side effects after rejection: metadata writes VERSION, GIT_COMMIT, BUILD_DATE and immutable PR identity only after selection guards; Makefile receives all build values. Build/package/verify require selection success. Final eligibility checks source main, PR state/head/repository/base/branch/label and stable versions; publication and pruning require that result. Existing retention and archive wiring remain unchanged. No live references to the removed engine remain in inspected consumers.

Executed read-only integration probe: temporary local clone detached at actual main 036d808b0c77bec400f23ca77a5cf056a0b2078b; extracted and executed the complete production metadata and eligibility shells with installed gh 2.45.0 and real GitHub --jq processing. Both exited 0 and emitted should_publish=true. Selected PR55/head1300741296d069b1282d89a6192219bf13fbb4f9, root target2.1.0, VERSION=v2.1.0-rc.114+036d808, GIT_COMMIT=036d808 and UTC BUILD_DATE. Separate API reads confirmed sole parent036d808 and successful producer run41612956703 on that SHA. Run number114 was probe input, not a claim of a hosted RC run.

Reused supplied 13-test pass for rejection/guard fixtures; inspected their shell execution and output assertions. Limits: no hosted scheduling, build/archive/upload or deletion executed; full make belongs to parent integration. GitHub reads/publication remain best effort as accepted. Scoped approval does not authorize publishing or certify those unexecuted boundaries.
