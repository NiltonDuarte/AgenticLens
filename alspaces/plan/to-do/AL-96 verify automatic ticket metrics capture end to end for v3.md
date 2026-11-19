# AL-96: Verify automatic ticket metrics capture end to end for v3

## Goal

Ship v3 only after a repeatable end-to-end test proves that the installed capture integration automatically records a development session and every same-ticket descendant in the ticket’s `agentic_metrics.json`.

## Context

Automatic capture publishes metrics after explicit ticket binding. The GEN-27 audit observed a fresh Codex ticket whose root and child were bound but whose first ledger required manual generation. AL-80 demonstrated native OpenCode capture on a supported version. The release gate must verify the path a user takes with the actual v3 candidate, including a ticket’s first publication. [Setup manual](../../../docs/SETUP_MANUAL.md#automatic-publication-policy) · [GEN-27 audit](https://github.com/NiltonDuarte/AgenticLens/pull/146) · [AL-80](../done/AL-80%20capture%20opencode%20sessions%20across%20worktrees.md)

## Acceptance criteria

- In an isolated, normally configured Git project, the test installs the v3 candidate through the supported setup path and launches a real harness session. The root receives an explicit ticket ID and title, responds to the normal binding request, and delegates distinct implementation and verification tasks on a small code change. The test waits for both child tasks and the root to finish. Include a nested delegation wherever the supported harness permits it. A deterministic provider may drive the real clients and tools.
- Starting with an explicitly selected ticket destination and no ledger, normal lifecycle callbacks and the worker create and finalize the ledger within a stated deadline. The run uses no manual generation, backfill, seeded ledger or direct worker invocation. Unresolved pending or partial publication fails the check.
- Compare the native source and the published ledger by exact ID. The root and every same-ticket descendant appear exactly once, with the correct harness and parent links. Tool activity and available token usage agree with the native source. An unrelated session is absent. After another normal turn and publication, earlier session records survive and cumulative metrics are counted once.
- Run the check for Codex and OpenCode versions covered by v3’s automatic-capture support. Record exact candidate and harness versions, the expected and observed session trees, and the result of each run. Explicitly report combinations that only support manual extraction. Reproduce and fix capture defects or gaps in the supported path, then rerun the gate on the corrected candidate before shipping v3.
