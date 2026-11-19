# AL-96: Verify automatic ticket metrics capture end to end for v3

## Goal

Ship v3 only after a repeatable end-to-end test proves that the installed capture integration automatically records a development session and every same-ticket descendant in the ticket’s `agentic_metrics.json`.

## Context

Automatic capture follows explicit ticket binding. In GEN-27, Codex bound a root and child, but their first ledger required manual generation. AL-80 demonstrated native OpenCode capture on a supported version. V3 needs evidence from the packaged candidate, including first publication. [Setup manual](../../../docs/SETUP_MANUAL.md#automatic-publication-policy) · [GEN-27 audit](https://github.com/NiltonDuarte/AgenticLens/pull/146) · [AL-80](../done/AL-80%20capture%20opencode%20sessions%20across%20worktrees.md)

## Acceptance criteria

- Install the v3 candidate through the supported setup path in an isolated Git project. Launch a real harness session. The root binds to an explicit ticket ID and title, delegates implementation and verification of a small code change, and waits for both tasks before finishing. Include a nested delegation where supported. A deterministic provider may drive real clients and tools.
- Starting with an explicitly selected ticket destination and no ledger, normal lifecycle callbacks and the worker create and finalize the ledger within a stated deadline. The run uses no manual generation, backfill, seeded ledger or direct worker invocation. Unresolved pending or partial publication fails the check.
- Compare the native source and the published ledger by exact ID. The root and every same-ticket descendant appear exactly once, with the correct harness and parent links. Tool activity and available token usage agree with the native source. An unrelated session is absent. After another normal turn and publication, earlier session records survive and cumulative metrics are counted once.
- Measure per-session and ticket `provisional_seconds` before and after native completion. Within the stated deadline, the first automatic refresh that sees completion must report zero provisional seconds for completed sessions and preserve their full measured duration exactly once. Record the delay from native completion to that refresh and fix any avoidable delay. Keep genuinely open or unconfirmed duration provisional. [Duration semantics](../done/AL-64%20investigate%20and%20fix%20active%20turn%20duration%20across%20codex%20and%20opencode.md)
- Run the check for Codex and OpenCode versions covered by v3’s automatic-capture support. Record exact candidate and harness versions, the expected and observed session trees, and the result of each run. Explicitly report combinations that only support manual extraction. Reproduce and fix capture defects or gaps in the supported path, then rerun the gate on the corrected candidate before shipping v3.
