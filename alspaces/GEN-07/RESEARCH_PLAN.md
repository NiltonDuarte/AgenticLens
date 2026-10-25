# GEN-07 Codex metrics research

Find the most complete, accurate, and privacy-conscious way for AgenticLens to extract Codex session metrics, comparable to its OpenCode extraction and live observability.

Research and validate three Codex surfaces:

- **Session history:** rollout JSONL under `CODEX_HOME`, including compressed/archived sessions and multi-agent threads.
- **Opt-in diagnostic traces:** `CODEX_ROLLOUT_TRACE_ROOT` bundles and the reduced trace graph; assess whether this is the highest-fidelity source for inference, tools, compactions, and agent relationships.
- **Harness-managed live runs:** `codex exec --json` event stream and process lifecycle.
- **Interactive Codex clients:** app-server JSON-RPC events and the limits on observing sessions owned by another client.
- **Local persistence:** read-only inspection of CODEX_HOME rollout files and the state, thread-history, logs, goals, memories, and queue SQLite databases; classify canonical, derived, diagnostic, and unrelated data.

Compare each surface with AgenticLens facts and computed metrics: invocation/thread hierarchy, role/model, turns and timing, token/cache/reasoning use, context/compaction, tool counts/errors/duration, anomalies, cost, and Git-derived code facts. Implement a fixture-backed probe where it resolves a real mapping question, then produce a schema-v3 `agentic_metrics.json` session ledger extracted from a real local Codex rollout.

Keep findings, code probes, fixtures, and session metrics in `alspaces/GEN-07`. Distinguish observed values from unavailable or inferred values; never invent cost or retain content-bearing transcript/tool payloads by default. Do not include the raw rollout in the repository. This remains a research ticket: no iterations or production-readiness claim are required.

Publication: keep real-session telemetry and diagnostic summaries gitignored. Commit only synthetic fixtures, source changes, and research notes without local identifiers or machine-specific inventories.
