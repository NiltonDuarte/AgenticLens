# docs(PO-18): plan Phase 1 post-v2.0 follow-up tickets

Create seven active Phase 1 tickets, numbered AL-60 through AL-66:

- AL-60: Search tickets within the comparison exclusion panel.
- AL-61: Show full local historical timestamps.
- AL-62: Fix Codex skill/delegation facts with granular session/invocation ownership, compatible migration, consumers and regression coverage.
- AL-63: Audit OpenCode v1/v2 and Codex capability parity; depends on AL-62.
- AL-64: Investigate active-turn duration across Codex, OpenCode v1 and OpenCode v2; fix the diagnosed Codex bug and any confirmed OpenCode defects with provisional snapshot/replacement semantics.
- AL-65: Research safe default source-backed backfill and Force semantics; depends on AL-63.
- AL-66: Configure optional generate output for skills, task_links and anomalies.

Remove the standalone active OpenCode walkthrough and renumber later active tickets consecutively. The parity audit retains its original capability-matrix scope and explicitly reports unexecuted checks; the removed walkthrough is not reassigned. Former AL-67's granular structure scope remains merged into AL-62, with its historical definition in plan/discarded. Update all active dependencies, filenames and references; historical iteration records retain their original IDs.

Planning documents only. Preserve AL-55 evidence and existing AL-57/AL2 plans. No application, tests, runtime scripts or real configuration changes; development runs in separate agentic-lens-dev sessions. Backfill default changes remain research-only.

Validation: ticket numbering/template completeness, roadmap links and reciprocal acyclic dependencies, unchanged remaining scope and AL-55 SHA256. Reuse prior full make PASS for unchanged code (Go race tests, frontend typecheck/lint, 741 frontend tests, canonical assets/binary builds). Iteration reviews record planning verification and source-grounded telemetry; no application fix is claimed by this PR.

AL-64 requires independent source/version evidence and regression coverage for generation during unfinished activity, refresh after completion and stable repeat backfill. OpenCode defects are investigated, not assumed; unsupported or unexecuted evidence stays explicit. This does not restore the removed live-inspector walkthrough.
