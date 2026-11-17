# AL-97: Decide whether to switch ordinary backfill defaults

## Goal

Give operators an explicit, evidence-backed product decision about the ordinary backfill default and a migration path if it changes.

## Context

[AL-65 discovery, PR #145](https://github.com/NiltonDuarte/AgenticLens/pull/145) recommends deferring a switch until unchanged repeats are write-free and operators can distinguish retained history from verified recovery. External workflow compatibility and acceptable batch cost still need product judgment.

## Acceptance criteria

- Review preview results against agreed batch size, runtime and read budgets; include exhaustion, repeat stability for both files, summary-only repair, partial and unavailable sources, mixed-harness preservation, and retained-owner diagnostics. Identify unverified workloads and external-script assumptions.
- Record an explicit go, defer or retain-current-default decision with rationale. Deferral or retention is a valid outcome. Preview completion, roadmap readiness and ticket creation confer no authority to switch.
- Before approving a switch, agree release timing, announcement, maintenance-only invocation, Force support or deprecation window, failure/exit compatibility and a way to restore prior behavior. Show migration examples for legacy batch, Git-only, explicit source and combined calls.
- Any rollout requires explicit product approval and separately authorized implementation. Preserve explicit harness and Git ticket/ref selection, existing recovery safeguards and opt-in-only heuristics throughout.
