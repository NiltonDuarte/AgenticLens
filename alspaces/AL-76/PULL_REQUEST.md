# feat(ledger): compact records and add selective format command (AL-76)

Persist ledger records with a shared readable JSON layout and provide an opt-in `format` command for all, one, or selected ticket ledgers. Preserve original facts, unknown fields and JSON literals while coordinating safe filesystem publication.

Support dry-run and check modes, explicit or configured workspace roots, deterministic target reports and source-independent operation. The formatting policy belongs to the invoked release; no hooks or release-pin updates are installed.

Contract: [CONTRACT.md](CONTRACT.md). Implementation and verification: [SUMMARY.md](SUMMARY.md).
