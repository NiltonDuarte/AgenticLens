# GEN-19 iteration 03: Simplify candidate eligibility

Status: Approved. The user's last two messages and explicit “make it simplier” instruction accept the assistant's concrete simplification proposal; no renewed approval is needed.

Baseline d29d834; PR #98 remains open; initial working tree clean. This delta supersedes iteration 02 freshness and historical stable-version requirements, preserving its version authority and workflow ordering.

Keep successful Release Please workflow_run ordering, explicit candidate-SHA checkouts, and manual dispatch. Discover the pending, same-repository root release PR, capture its immutable head, read its manifest, validate the stable version, and append the existing RC suffix. Preserve API-error and ambiguous-selection failure behavior and explicit no-PR skipping.

Immediately before publishing, perform one exact target stable-version/tag existence check. Remove generated-commit parent validation, repeated main-head/currentness checks, paginated historical stable-release lists, and the second PR identity check. A newer main commit or later release-PR update no longer invalidates a valid captured and built candidate. These deliberately relaxed guarantees are accepted scope, not unresolved review findings. Preserve build, archive verification, publication, pruning, and suffix behavior. Add no standalone script or dependency.

Coder owns the RC workflow and relevant CONTRIBUTING guidance. Tester owns existing release-workflow tests: retain meaningful version/output, PR filtering, API-error, and exact stable-guard evidence; delete only superseded freshness/history assertions and record their disposition. Both preserve shared-checkout edits.

Review all four core responsibilities under both configured model profiles because accepted release behavior changes. Parent owns consolidation, integration, required make verification, telemetry, commits, and PR delivery. Reuse valid unchanged evidence and rerun affected checks. Record hosted scheduling/publication limits without manufacturing execution evidence. No stable release or release-PR merge is authorized.
