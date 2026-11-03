# Behavior recheck — REL-02 publication binding

**Outcome: APPROVED (scoped).** Reviewed candidate fingerprint `7655ae2f7e75ee3c9615525029fa65a95bc356d2156dc8a6832ea30e4dab26e2`; requested identity: gpt-6-luna / medium. This verifies the supplied correction, not a new discovery or whole-workflow approval.

`EXPECTED_SHA` is derived from the completed Release Please run's `head_sha` (or manual dispatch `github.sha`). Both checkout steps pin `ref` to it; the build job checks `git rev-parse HEAD` against it before creating metadata, and `GIT_COMMIT` is derived from that checked-out `HEAD`. The publish action now receives `target_commitish: ${{ env.EXPECTED_SHA }}` alongside the RC `tag_name`. Thus a newly created tag targets the captured commit associated with the built artifacts, even if the default branch advances. Existing test `test_candidate_build_and_publication_require_selection_and_final_eligibility` asserts this exact action input in the publication step; deleting the binding causes the assertion to fail.

The correction adds no runtime guard, API request, dependency, or freshness/provenance/history check. Inspection confirms the intentionally removed main-head, parent, stable-history, and PR recheck code remains absent, consistent with accepted iteration 03 scope.

**Verification reused:** parent-provided valid 14-test/evidence result (not rerun). **Limits:** this is source and test-oracle evidence; no hosted workflow or GitHub release was executed. Parent's integrated `make` gate remains outstanding.
