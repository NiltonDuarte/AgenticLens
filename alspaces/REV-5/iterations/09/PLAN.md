# REV-5 iteration 09: ticket all open findings and recover reviewer telemetry

Source: user requested, "Create tickets for the open findings, add the two sessions to the ledger" after the merge-readiness check.

Baseline cd3a880; initial working tree clean. Continue REV-5 ownership. User authorizes backlog specifications and real-source ledger repair, not application remediation or PR merge. Use the PO ticket conventions without creating new active developer workspaces. Current origin/main fetched; existing AL-60 through AL-68 inspected before allocating AL-69 through AL-93.

Changed acceptance: every open iteration 03 catalog entry has an explicit owner; preserve fixed ADP-01/D3 and original iteration 02 fixes, consolidate overlaps, and reuse AL-62 for CC-02. Record unnumbered dependency/toolchain observations as bounded investigation, not confirmed vulnerabilities. Preserve local-only ADP-05 scope and all previous preservation/identity guarantees.

Recover the two reviewer session IDs from actual Codex files using explicit --session-id selection into a temporary candidate copied from the existing ledger. Verify cumulative usage/model against source, preservation of all 25 prior sessions, timestamp envelope, unknown cost and replay idempotence before installing the candidate. Do not fabricate native parent links or change source to T3. Final root-nonce refresh must retain both imports and precede final telemetry commit. Historical incomplete snapshots remain historical; record current recovery separately.

Parent owns all ticket/index/mapping artifacts, ledger, source verification, integration and publication. This is backlog recording plus an existing CLI operation, not new product code or workflow-policy implementation. Validate ticket coverage, references, acceptance boundaries and ledger invariants directly. No new parallel product review is warranted; no application/test/runtime changes are authorized. Existing application make and exact-head prior CI evidence remain applicable to unchanged code.
