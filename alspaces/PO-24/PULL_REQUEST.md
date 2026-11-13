# docs(PO-24): plan OpenSpec adoption and multi-root workspaces

Adopt OpenSpec as the home for future AgenticLens work, with telemetry inside ticket-prefixed change folders and stable access after archive. The product spec defines one end-to-end implementation slice, AL-84, covering multi-root reads, safe publication/binding, OpenSpec setup, agent/skill guidance and the live repository/backlog cutover. The user explicitly requested a single slice; partial platform delivery does not satisfy it.

This replaces the GEN-26 assessment PR. The assessment and source-session ledger are relocated to PO-24; iteration plans and the old assessment contract are removed. Existing roadmap work retains its IDs and ownership. Historical alspaces ledgers remain supported through an explicit legacy root rather than being duplicated or rewritten.

Only planning documents and telemetry are delivered. No OpenSpec initialization, application implementation, runtime configuration or live workflow migration occurs in this PO session. [SPEC.md](SPEC.md) owns the initiative; [SUMMARY.md](SUMMARY.md) records verification, source limits and reclassification. Development must verify the OpenSpec release/tooling and freeze the concrete configuration/resolver contracts.
