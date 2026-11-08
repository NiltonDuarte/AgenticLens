# AL-63 iteration 01: Capability audit planning

## Scope and source

The user invoked agentic-lens-po with the existing AL-63 specification. Refine that single ticket for planning delivery; retain AL-63 rather than allocate a duplicate PO identifier. Initial Git status was clean on AL-63/verify-source-capabilities-across-harnesses at eb7d169c. Configured workspace root is alspaces. This workspace holds PO planning artifacts only; no developer contract or implementation is authorized.

## Acceptance and decisions

Keep the ticket's original discovery scope and all eight comparison rows. PO-18 explicitly identifies the three columns as OpenCode v1, OpenCode v2 and Codex. Clarify evidence fields and native versus fixture versus historical proof, without assuming capability parity. Make canonical persisted-fact, complete-empty and exact-owner replacement guarantees explicit. Preserve removed-walkthrough exclusion, original AL-55 data and historical findings. Existing FIX-18 and REV-7 retain their boundaries.

The specification remains the single acceptance owner. The registry remains unchanged: AL-63 Ready, AL-65 blocked by AL-63. Planning delivery does not complete the audit, move its specification to done or unblock AL-65. No new tickets or parallel specification are needed. No subagents: this is documentation-only planning.

## Verification and delivery

Inspect current source declarations/dispatch and prior AL-62/PO-18/PO-19 evidence without claiming runtime compatibility. Validate local links, comparison coverage, unchanged dependency metadata, whitespace and documentation-only diff. Commit owned planning files, integrate origin/main and run required make. Record actual results and limits in SUMMARY.md. Generate Codex telemetry as the final content step using docs classification, verify real root/source coverage, then commit and deliver a planning PR. No audit probes or new native provider sessions are part of this PO session.

Selected regression boundaries: declared versus observed compatibility; portable ledgers with unavailable enrichment; exact identity and complete-empty authority; configured workspace paths; honest incomplete evidence. Canonical rules remain authoritative; this plan does not introduce a new authenticity policy.

Author verification passed on planning commit 0a8cfcee: link/coverage/scope checks and unchanged registry. Integration with origin/main was a no-op. Full make exited 0 with all required lanes executed; detailed results and environmental warnings are in SUMMARY.md. No source probes or independent code review were performed.
