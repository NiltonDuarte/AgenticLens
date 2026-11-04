h3. Title: REV-6: Review test performance and unit/integration boundaries

*TARGET*: AgenticLens (Test architecture / Capture CLI and adapters / CI sequencing)

*Status*: Done
*Work type*: review
*Blocked by*: None
*Blocks*: None
*Origin*: User requested a review after the MGT-2 batch, citing slow capture tests and requiring fast, cheap unit checks before slower integration checks. MGT-2 is ending at the user's request; AL-62 will restart separately. Creating this ticket does not claim its old PR merged or start this review.
*Related*: AL-74 owns its existing scoped test-reliability/unused-code fixes. Reuse its findings where relevant; do not duplicate that implementation scope.

*CONTEXT*

Capture/CLI verification is expensive. A reported duration near 43 seconds needs measurement: a single test, a whole package, subprocess builds, startup, waits and race instrumentation have different causes. Genuine integration coverage must remain meaningful, while the unit suite should be fast and inexpensive to run.

*GOAL*

Produce a measured review and bounded follow-up recommendations for fast unit checks followed by slower integration checks. This is a review ticket, not authorization to refactor tests, change timeouts or weaken coverage.

*SCOPE AND ACCEPTANCE*

* Measure per-test and per-package duration, subprocess/build overhead, waits and resource cost, beginning with capture. State cache, race-detector, concurrency, hardware/OS and source revision assumptions so timings are comparable.
* Classify tests by the boundaries actually exercised. Native CLI processes, lifecycle hooks, external harnesses and genuine filesystem/SQLite/Git behavior may require integration tests; renaming a slow test does not solve its cost.
* Identify slow unit work that can be made deterministic/cheaper and genuine integration tests that should run separately. Preserve independent assertions, native ownership/pending/ACK guarantees and realistic boundary fixtures.
* Recommend practical budgets, commands and CI dependencies so slower integration execution starts only after fast unit checks pass. Account for local and hosted workflows, shared fixture builds and platform-specific coverage.
* Distinguish individual-test slowness from package/suite duration. Do not use blanket timeout inflation, silent skips, arbitrary coverage removal or test-count targets as a fix.
* Deliver concise measured findings with actionable scoped follow-ups and review limits. Use separate FIX tickets for unrelated concrete defects under the management routing rule; do not implement them in this review session.
* Inspect the then-current main and any delivered AL-62 changes at review start. This ticket makes no claim about the unfinished/closed AL-62 candidate.

*DEFINITION OF DONE*

* Measured review delivered with clear unit/integration classification, recommended sequencing and prioritized follow-ups.
* Existing coverage preserved during the review; any later implementation has its own approved ticket and verification.
* No review/development thread is launched by this planning-only ticket creation.

*REFERENCES*

* [Test purpose and proportionality](../../../TECHNICAL_GUIDELINES.md#test-purpose-and-proportionality)
* [Test review checklist](../../../docs/REGRESSION_PREVENTION.md#adding-reviewing-or-reducing-tests)
