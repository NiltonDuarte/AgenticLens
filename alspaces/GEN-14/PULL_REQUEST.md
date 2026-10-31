# chore(GEN-14): audit independent manager and ticket telemetry

Document the user-proxy workflow: a manager ticket sends user-like instructions to independent top-level ticket threads while each ticket retains its own workspace, root and telemetry ledger. The real hello-world exchange has separate Codex roots and a user-role destination message.

Record the remaining manager instrumentation gap: completed T3 McpToolCall records are omitted by the current Codex decoder. No application code is changed. Iteration 02 is authoritative; iteration 01 is retained as superseded audit history.

Validation: source-backed T3/Codex inspection and the full make gate; exact final CLI results are recorded in iterations/02/REVIEW.md. Raw transcripts stay outside Git.
