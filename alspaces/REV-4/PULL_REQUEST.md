# test(REV-4): prune low-value tests and consolidate role ordering

Reduce the suite from 1,253 to 1,197 top-level declarations: delete 55 low-value tests and consolidate one role-ordering wrapper into the existing enum table. Net reduction: 1,896 test lines. No production behavior changes.

Remove fixture-only JSON/type checks, copied release logic, duplicated component cases, the remaining release-verification file, and a redundant date-preset smoke. Release workflow/help strings and missing-input guards did not establish successful CI release/publishing; deleting those guards does not claim to improve release validation. The canonical role cases now share the existing 19-row enum table with both ordering entrypoints, with no new tests or rows.

REVIEW.md records removal reasons and retained protection. SCORE_REVIEW.md explains the 0-10 rubric and historical findings; TEST_SCORES.md rates all 1,197 current declarations, with evidence and confidence labels. Scores are not an automatic deletion threshold or mutation-equivalence proof.

Scope: [CONTRACT.md](alspaces/REV-4/CONTRACT.md) and iteration 02/03/04 deltas. Verification and telemetry: [SUMMARY.md](alspaces/REV-4/SUMMARY.md).
