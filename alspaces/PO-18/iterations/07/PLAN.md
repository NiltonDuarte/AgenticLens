# PO-18 iteration 07 plan

Baseline ec16b60; initial Git status clean; PR #90 open. Source: user requests a broader repository interface/injection architecture review across OpenCode v1/v2 and Codex.

Acceptance delta: add Ready AL-68 for architecture review, as-is contracts/wiring, common/capability-specific operations, schema/harness selection, mixed-source injection, failure/availability semantics, evidence and proposed design. Review-only; no automatic refactor implementation. AL-67 remains retired and existing toggle ticket is AL-66; AL-68 is the next unreserved current ID after checking workspace/registry/remote heads. Supersedes seven-active-ticket count with eight. Preserve all existing ticket definitions/dependencies, AL-55 evidence and PO-only execution.

Inspect repository/service/config/composition entry points read-only to ground the review ticket, without performing the review itself. Write template-based AL-68 and update registry/SPEC/PR/summary. Validate uniqueness, template/links, Ready placement, no dependency cycles, unchanged existing definitions and AL-55 hash. Reuse prior make PASS for unchanged code; integrate main and reassess if code changes. Refresh Codex telemetry as final step with literal nonce, verified root/source and session preservation, commit/push/update PR #90.

Applicable regression boundaries: source/schema identity, injected dependency ownership, unavailable/partial versus empty, persisted baseline without hydration, mixed-harness retention, evidence vs architectural recommendation. No source parity or runtime architecture correctness is claimed by this PO update.

Review correction after 87fea2c: remove redundant Phase metadata from the current batch, per user feedback. Remain in iteration 07 because scope/acceptance/dependencies do not change. Verify content changes are metadata-only against baseline, preserve older/historical phase records and AL-55 hash, reuse unchanged code checks, integrate main and refresh final telemetry before delivery.
