# Test effectiveness review

**Outcome: APPROVED (scoped).** Reviewed candidate fingerprint `7655ae2f7e75ee3c9615525029fa65a95bc356d2156dc8a6832ea30e4dab26e2` against base `d29d834`. Requested reviewer identity: gpt-6-luna, medium; coordinator run ID was not supplied.

The retained tests exercise the production shell at the relevant boundaries. `test_manifest_selection_writes_real_metadata_and_uses_immutable_pr_head` checks the manifest request pins `ref` to the captured PR head and checks resulting RC metadata/suffix. The raw-record filter test executes captured production `gh` filters with real `gh` against local JSON, including same-repository, branch, label, base and malformed manifest cases. This detects filter drift that the canned `gh` fixture alone would miss.

The exact stable-tag test runs the actual `curl` step against a local HTTP server and checks the exact `/git/ref/tags/v2.1.0` path. Its 404, 200, 403, 500 and dropped-connection cases distinguish allow, skip and blocking failure outcomes. The publication wiring test asserts `target_commitish: ${{ env.EXPECTED_SHA }}`; removing or changing that value fails the assertion. This protects tag-to-binary source identity.

Deletion of parent-provenance, main-recheck, historical stable-scan and second-PR-identity assertions matches the explicitly relaxed iteration03 contract. No numeric findings.

Reused supplied evidence: 14 release workflow tests passed with zero skips; supplied disposable wrong-version, allow-on-200 and wrong-tag-URL mutations failed. Independently checked candidate file hashes against `CANDIDATE.json`; no test rerun was needed. Limits: these tests do not establish hosted scheduling, release upload/pruning behavior or atomicity between tag lookup and publication; integrated `make` remains the parent gate. No blocking test-effectiveness gap found in this scope.
