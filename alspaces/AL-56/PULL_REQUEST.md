# AL-56: Add Codex session discovery and metrics generation

`agentic-lens generate --source codex` discovers local Codex rollouts and writes the existing schema-v3 fact ledger. Resolution supports a literal nonce, an explicit thread ID, or ticket matching, with child threads included through parent metadata. OpenCode remains the default source. The existing dashboard reads the generated ledger; this PR adds no HTTP routes or frontend changes.

Repeated generation replaces the selected snapshot, model changes preserve per-model usage, and event context uses request input counts rather than cumulative usage. CLI persistence uses the existing error/compaction/anomaly-turn pruning policy.

## Privacy review

- Real-session ledgers and generated diagnostic summaries are gitignored and excluded from the PR.
- Published research omits real thread IDs, literal correlation nonces, local home paths, and machine-specific database inventories. Committed rollout fixtures are synthetic.
- Parser tests verify that prompts, command arguments, tool payloads, and working-directory paths are excluded from exported summaries and ledgers. Extraction runs locally without external service calls.
- Generated local ledgers intentionally retain session identity and timing for correlation. They need review before sharing.

## Validation

- `make`: Go vet, race-enabled Go tests, frontend typecheck/lint/tests, frontend bundle, and cross-platform binary builds.
- CLI regression tests run in isolated temporary projects with synthetic ticket prefixes and custom workspace layouts, independent of the checkout configuration and research ticket.
- Regression coverage for discovery, nonce precedence, gzip rollouts, child threads, repeated generation, model switches, token accounting, event retention, and export privacy.
- `git diff --check` and a scan of the publication candidate files for local identifiers, home paths, credentials, and email addresses.

## Limits

Rollout persistence is a private Codex format. Discovery scans files and has no live tailing or automatic harness detection; use `--source codex`. Cost remains unavailable rather than reported as measured billing. The existing npm dependency audit reports six advisories; this change introduces no new dependencies.
