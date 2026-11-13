# docs(audit): verify cross-harness source capabilities (AL-63)

Deliver the original AL-63 capability audit for OpenCode v1, OpenCode v2 and Codex, preserving its unchanged acceptance. The matrix separates recorded local-source observations, constructed fixtures, unsupported behavior, failures and unexecuted cases across eight capability areas.

Integrate main4e5c4583 and refresh the report for AL-69/AL-70/AL-79: configured Codex-home hydration and live child diagnostics are corrected upstream; mixed recovery now rejects unsupported layouts explicitly. Metadata child-link loss and Codex function-output error counts still reproduce. Keep original discovery credit and date earlier native observations rather than relabel them as new verification. No application changes are introduced by this PR.

[Matrix/evidence](alspaces/AL-63/CAPABILITIES.md), [verification/limits](alspaces/AL-63/SUMMARY.md), and [consolidated attribution](alspaces/AL-63/iterations/02/REVIEW.md) retain current versus historical evidence. Current isolated probes and focused tests passed; the integrated gate and final session ledger are recorded in the summary. AL-63 acceptance is archived unchanged; AL-65 and upstream AL-80 remain Ready. This audit does not approve parity, follow-up implementation or a backfill-default change.
