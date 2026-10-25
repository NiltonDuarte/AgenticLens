# AL-56: Add Codex session discovery and metrics generation

`agentic-lens generate --source codex` discovers local Codex rollouts and writes the existing schema-v3 fact ledger. Resolution supports a literal nonce, an explicit thread ID, or ticket matching, with child threads included through parent metadata. OpenCode remains the default source. The existing dashboard reads the generated ledger; this PR adds no HTTP routes or frontend changes.

`--code-ref HEAD` supports code attribution for branches without ticket IDs, including renamed tickets. It measures committed changes against the configured base and respects source/classification settings. Codex SDLC iterations derive from root-session user messages, without duplicate event/response counting or child delegation.

Test/production classification now comes from project configuration rather than hardcoded directory names or extensions. Recursive globs respect directory boundaries; configured test paths outside source roots are included, and both Git discovery modes apply source filtering.

Repeated generation replaces the selected snapshot, model changes preserve per-model usage, and event context uses request input counts rather than cumulative usage. CLI persistence uses the existing error/compaction/anomaly-turn pruning policy.

## Privacy review

- The real-session `agentic_metrics.json` and generated metrics summary in `SUMMARY.md` are committed lifecycle deliverables.
- The ledger retains real thread IDs and timing for correlation. Raw rollouts, local home paths, and machine-specific database inventories are excluded. Test fixtures are synthetic.
- Parser tests verify that prompts, command arguments, tool payloads, and working-directory paths are excluded from exported summaries and ledgers. Extraction runs locally without external service calls.
- Metrics are generated from this conversation using nonce-based session discovery.

## Validation

- `make`: Go vet, race-enabled Go tests, frontend typecheck/lint/tests, frontend bundle, and cross-platform binary builds.
- CLI regression tests run in isolated temporary projects with synthetic ticket prefixes and custom workspace layouts, independent of the checkout configuration and research ticket.
- Regression coverage for discovery, nonce precedence, gzip rollouts, child identity/parent linkage, inherited replay exclusion, missing resume markers, configured persona role mapping, path privacy, repeated generation, model switches, token accounting, event retention, and export privacy.
- `git diff --check`; generated facts omit transcript content, command payloads, and working-directory paths.

## Limits

Rollout persistence is a private Codex format. Discovery scans files and has no live tailing or automatic harness detection; use `--source codex`. Cost remains unavailable rather than reported as measured billing. The existing npm dependency audit reports six advisories; this change introduces no new dependencies.
