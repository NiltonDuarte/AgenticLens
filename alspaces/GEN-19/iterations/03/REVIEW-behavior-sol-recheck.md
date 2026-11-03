# Behavior publication-binding recheck

APPROVED for the supplied REL-02 correction; no new finding.

Run: rc3_behavior_sol_recheck; phase: supplied-fix verification. Requested profile: gpt-6.1-sol/high; runtime model metadata unavailable. Base: d29d8347245b357e0a64c3fee2c3bbed4276e3c9. Reviewed fingerprint: 7655ae2f7e75ee3c9615525029fa65a95bc356d2156dc8a6832ea30e4dab26e2; all three file hashes verified.

The supported operation creates a new RC tag after building the captured commit. Workflow lines 11, 29, 84 and 109 bind checkouts and metadata to EXPECTED_SHA; lines 124–150 derive and build its short SHA. Publication line 182 now supplies that same full SHA, preserving workflow_run.head_sha precedence over github.sha. Advancing main therefore cannot redirect a newly created tag through the API default. The [action input](https://raw.githubusercontent.com/softprops/action-gh-release/v2/action.yml) supports commit SHAs; its [creation path](https://raw.githubusercontent.com/softprops/action-gh-release/v2/src/github.ts) forwards the parameter. [GitHub documents](https://docs.github.com/en/rest/releases/releases#create-a-release) that it controls new-tag creation.

Verification: existing publication wiring test passed; an in-memory removal of the input failed its new assertion. Removing exactly the supplied workflow line and test assertion reproduces both initial candidate hashes, proving no removed guard or dependency returned. Reuse the supplied valid 14-test evidence for unchanged behavior; no broad suite repeated.

Limits: static/API-contract and local assertion evidence; no hosted publication or token-permission probe. Existing tags are not retargeted by this API parameter. Final integrated make remains the parent's gate. This verifies the supplied discovery rather than claiming independent discovery credit.
