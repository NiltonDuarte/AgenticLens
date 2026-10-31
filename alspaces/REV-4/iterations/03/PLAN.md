# REV-4 iteration 03 plan

1. Parse every current Go/TypeScript test declaration structurally, with current source line, complete body, assertions, production calls, and compiler directives. Validate counts against iteration02 evidence.
2. Define an anchored 0-10 importance rubric. Inspect low-value and critical clusters against actual test bodies, production code, and overlapping suites. Assign each declaration a score and supporting contract evidence; distinguish inferred importance from proven redundant coverage.
3. Write SCORE_REVIEW.md with prioritization and TEST_SCORES.md with every current declaration. Preserve historical removal ledger rather than rewriting old decisions into artificial scores.
4. Validate complete one-to-one coverage, numeric range, reasons, and source references; profile current suites, integrate main and run required make gate. Record session evidence, generate telemetry last, commit, push, update and link PR.

No test deletion or application edits authorized for this scoring iteration; the user's latest prompt steers the same REV-4 ticket.

Completed: current AST inventory, calibrated scoring, catalog validation, fresh Go profile, main integration and required full make all passed. Final telemetry and PR update follow. No tests changed.
