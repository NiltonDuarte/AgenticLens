# feat(telemetry): distinguish semantic tool content errors (AL-43)

Completed tool calls can emit compiler errors, tracebacks or assertion failures while their process succeeds. Track these as independent content warnings so execution reliability remains accurate and diagnostic evidence stays visible.

Adds pure content classification, canonical session persistence and counts, supported OpenCode/Codex and live extraction, and API projections. The dashboard shows amber content warnings alongside red execution failures, separate live counters, and Deep Dive content-error totals. Legacy ledgers retain zero defaults.

Contract: [CONTRACT.md](CONTRACT.md)
Implementation and verification: [SUMMARY.md](SUMMARY.md)
