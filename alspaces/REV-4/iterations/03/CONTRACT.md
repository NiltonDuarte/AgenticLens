# REV-4 iteration 03 contract: score remaining tests

Baseline: `2e87ab15ffee382b1685bf38f233eca3d76fe72e`, clean working tree. Existing PR72 is open. Configuration workspace root: `alspaces`.

User requests another review expressed as a 0-10 score for each test, replacing binary dispositions in this new review. Zero means definitely expendable; ten means critical protection that must survive. Review all current executable declarations (505 Go and 698 TypeScript, 1,203 total), with evidence and reasons. Parameterized declarations receive a score for their combined contract, not a claim about every expanded table row.

Documentation-only scope: create scored review artifacts, preserve iteration01/02 findings and removal history, and update summary/PR context. No additional tests or production changes. Score prioritizes demonstrated regression value, oracle strength, and marginal protection. Explicitly state uncertainty and overlap proof limitations. Reuse valid unchanged-code verification; record profiling separately from safety proof. No mutation-equivalence or controlled speedup claims.
