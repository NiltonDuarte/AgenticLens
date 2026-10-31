# docs(PO-18): plan post-v2.0 follow-up tickets

Create eight active tickets from the post-v2.0 review and user corrections:

- AL-60: Search tickets within the comparison exclusion panel.
- AL-61: Show full local historical timestamps.
- AL-62: Fix missing Codex skill/delegation facts with granular ownership, safe migration and all consumers.
- AL-63: Audit OpenCode v1/v2 and Codex capabilities; depends on AL-62.
- AL-64: Investigate active-turn duration across all three sources; fix the diagnosed Codex bug and confirmed OpenCode defects.
- AL-65: Research safe source-backed backfill defaults and Force semantics; depends on AL-63.
- AL-66: Configure optional generate output for skills, task_links and anomalies.
- AL-68: Review shared repository contracts and correct source injection, with as-is evidence and a proposed architecture.

Consolidate PO-18 into one current SPEC.md, a factual SUMMARY.md, this PR description and the session telemetry ledger. Remove development-style iterations/ plans and self-review files at the user's request; author checks are not independent review or plan approval. Keep individual AL specifications, acceptance boundaries and registry dependencies intact. AL-67's granular structure scope remains merged into AL-62. The standalone live walkthrough was removed and redundant phase metadata dropped.

Planning only: no application/test/runtime/config changes. Preserve original AL-55 evidence and existing AL-57/AL2 plans. Implementation, backfill behavior changes and architecture refactors require their separate development boundaries.

Validation: document structure and links, reciprocal dependencies, unchanged ticket acceptance, no dangling iteration references and unchanged AL-55 SHA256. SUMMARY.md records author verification and source-grounded telemetry, including reused full make PASS for unchanged code (Go race tests, frontend typecheck/lint, 741 frontend tests, canonical assets/binary builds). No planned application fix is claimed as delivered.
