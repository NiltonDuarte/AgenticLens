# GEN-21 iteration 02: Codex source backfill

Steering source: user instruction "backfill codex ledgers" following the completed migration (2026-11-04).

Changed acceptance: refresh exact Codex-owned sessions from this machine's /home/niltonduarte/.codex source store across all existing workspace ledgers, including nested history. User explicitly authorizes this source-backed refresh. Preserve unavailable/partial sessions and all other harness owners; no heuristic matching. Retain schema 6 and recompute aggregate views through the CLI. Complete snapshots may replace older facts with fewer invocations. Refresh summaries through the supported CLI option.

Preserved guarantees: portable history, session identity, explicit uncertainty and no capture-registry reset, application or configuration edits. Supersedes iteration 01's exclusion of source backfill and unchanged-AL-62 acceptance for this new iteration; migration hashes remain historical before-backfill evidence.

Verification impact: retain separate pre-backfill originals outside Git, dry-run using verified published rc.125, apply backfill --source codex --recompute-db --update-summaries, compare exact preserved-owner snapshots, verify source availability diagnostics and current-schema reads. Repeat only where required by findings. Existing make pass remains valid for unchanged application code; run the required integration gate if upstream code changes. Final telemetry may remain blocked by previously recorded capture admission incompatibility.
