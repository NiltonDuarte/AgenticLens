# Iteration 06 Review

## Verdict: APPROVED

- The previous iteration's overbroad ad hoc exception is corrected: all tasks require a ticket ID and isolated workspace.
- Ad hoc work receives a `GEN-#` ID when none is provided, without requiring a specification or index entry under `plan/`.
- Lifecycle instructions still move an existing matching specification to `plan/done/` and remove an existing index entry when present.
- `AGENTS.md`, `CONTRIBUTING.md`, and the setup manual policy block agree.
- `git diff --check` passed; targeted search found no contradictory mandatory-ID wording.

No application or test code changed.
