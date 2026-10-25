# AL-56: Add Codex session discovery and metrics generation

`agentic-lens generate --source codex` discovers local Codex rollouts and writes the existing schema-v3 fact ledger. Resolution supports a literal nonce, an explicit thread ID, or ticket matching, with child threads included through parent metadata. OpenCode remains the default source.

The research adapter also provides a dashboard file picker for in-memory extraction, preview, and JSON download. Repeated generation replaces the selected snapshot, model changes preserve per-model usage, and event context uses request input counts rather than cumulative usage. CLI persistence and downloaded ledgers share the existing error/compaction/anomaly-turn pruning policy.

## Privacy review

- Real-session ledgers and generated diagnostic summaries are gitignored and excluded from the PR.
- Published research omits real thread IDs, literal correlation nonces, local home paths, and machine-specific database inventories. Committed rollout fixtures are synthetic.
- Parser and HTTP tests verify that prompts, command arguments, tool payloads, and uploaded working-directory paths are excluded from responses. The browser sends the selected file only to the current AgenticLens server; extraction does not persist the uploaded transcript or call an external service.
- Generated local ledgers intentionally retain session identity and timing for correlation. They need review before sharing.

## Validation

- `make`: Go vet, race-enabled Go tests, frontend typecheck/lint/tests, frontend bundle, and cross-platform binary builds.
- Regression coverage for discovery, nonce precedence, gzip rollouts, child threads, repeated generation, model switches, token accounting, event retention, and response privacy.
- `git diff --check` and a scan of the publication candidate files for local identifiers, home paths, credentials, and email addresses.

## Limits

Rollout persistence is a private Codex format. Discovery scans files and has no live tailing or automatic harness detection; use `--source codex`. Cost remains unavailable rather than reported as measured billing. The upload endpoint accepts one uncompressed JSONL file up to 64 MiB. The existing npm dependency audit reports six advisories; this change introduces no new dependencies.
