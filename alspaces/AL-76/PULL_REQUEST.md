# feat(ledger): compact records and add selective format command (AL-76)

Ledger writers expanded small records across many lines and had separate layout paths. Use one JSON-aware policy that compacts recognized tool, token, evidence and anomaly records while preserving readable outer structure.

Add opt-in `format` selection for all, one or a set of ticket ledgers, with configured/explicit roots, dry-run/check modes and deterministic reporting. Existing-file formatting preserves exact literals, member order, unknown fields and facts without querying sources or migrating schemas. Preflight and canonical writer ownership protect concurrent publication; changed files publish atomically with permissions preserved.

Integrated main includes AL-60 capture repairs and AL-61 historical timestamps. Final review and verification are recorded in [SUMMARY.md](alspaces/AL-76/SUMMARY.md); acceptance boundaries in [CONTRACT.md](alspaces/AL-76/CONTRACT.md). No hooks, mandatory formatting check, release-pin update or mass reformat is installed.

MGT-2 owns merging. Hosting checks are the merge owner's final gate.
