# test(REV-4): prune low-value tests and consolidate coverage

Reduce the suite from 1,253 to 1,155 top-level declarations, with 2,430 net test lines removed. Delete fixture-only checks, copied release logic, duplicate smoke cases and unnecessary wrappers; preserve meaningful role-ordering and UI/boundary contracts in existing tests. Production behavior is unchanged.

The 0-10 review covers the current suite. A direct follow-up audit examined all 47 declarations scoring at most four: seven redundant/overconstrained cases deleted, 35 donor cases consolidated, and five independent guards retained at their original scores. No new top-level tests, helper suites or parameterized declarations were added. Scores are not treated as an automatic deletion threshold, and no mutation-equivalence guarantee is claimed.

CSS guards, compiler narrowing, metadata precedence, tooltip/header wiring and existing callbacks survive their named transfers. REVIEW.md retains earlier removal history; SCORE_REVIEW.md explains scoring and the confidence limits; TEST_SCORES.md lists all 1,155 current entries in ascending order. iterations/05/CANDIDATES.md records every low-score decision and recipient.

Scope: [CONTRACT.md](alspaces/REV-4/CONTRACT.md) and iteration deltas. Verification and session provenance: [SUMMARY.md](alspaces/REV-4/SUMMARY.md).
