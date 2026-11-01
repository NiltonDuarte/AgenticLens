# feat(capture): install harness lifecycle hooks for automatic session capture (AL-68)

Add binary-owned Codex/OpenCode lifecycle adapters, durable local registration, explicit agent binding and exact-source reconciliation. Independent sessions remain separate; native parents are accepted only from proven source records. Missing or ambiguous facts stay visible.

Install and upgrade additively with native-handler compatibility checks and rollback, preserve user hooks, and provide diagnostics, recovery/removal and a digest-bound coverage report without changing canonical ledger v5.

Codex0.166.0 and OpenCode1.19.5 Linux runtime paths are documented with bounded coverage. Codex SubagentStart was not observed on the native exec path; running/interrupted child coverage before Stop remains degraded/unverified. Additional platforms/versions/interfaces are not claimed verified.

Scope: [CONTRACT.md](https://github.com/NiltonDuarte/AgenticLens/blob/feat/AL-68-automatic-session-capture-hooks/alspaces/AL-68/CONTRACT.md). Implementation, evidence and limitations: [SUMMARY.md](https://github.com/NiltonDuarte/AgenticLens/blob/feat/AL-68-automatic-session-capture-hooks/alspaces/AL-68/SUMMARY.md).
