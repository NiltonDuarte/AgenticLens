# AL-57 frontend implementation

Baseline: 45cb679f21475cd14b7dd4cbc25996a39d734403. Scope: web source and relevant existing frontend tests, matching backend cost evidence/summary wire fields. No commit, telemetry or PR work performed by this agent.

Implemented nullable complete amounts, observed/partial/unavailable fact evidence, invocation-count cost summaries, shared known-subtotal/coverage formatting, stable complete-first sorting in either direction, availability-aware averages/shares, frontend bucket/cumulative evidence composition, source-missing chart gaps and omission counts, preserved known zero, task-link missing-cost display and focused API cost runtime validation. Historical numeric values follow ordinary observed behavior; no special history state, counter or label exists.

Existing numeric legacy/in-memory fixtures remain supported through optional evidence/summary type fields and explicit numeric compatibility helpers. Schema-v5 API facts require evidence at runtime; aggregate metadata is validated for amount/state/count consistency. All actual API responses with supplied metadata must honor it. Changed cost fields remain strict number-or-null rather than any.

Verification on final frontend diff:

- npm install: passed, package-lock.json unchanged. Dependency audit reported existing advisories; no dependency upgrades attempted.
- npm run typecheck: passed.
- npm run lint: passed.
- npm test: 35 files, 747 tests passed, no skips. Existing MetricProgression tests emit React act warnings; suite exits successfully.
- npm run build: passed; current JS/CSS/index copied to backend/internal/adapters/web/dist.
- diff -qr web/dist backend/internal/adapters/web/dist: passed; both asset trees ignored under existing tracking policy.
- git diff --check: passed.

Focused regression additions extend existing suites: partial/coalesced subtotal plus cumulative unknown and filter recomputation; complete-first cost sorting including UI clicks; API rejection of contradictory summary metadata; chart gaps and unknown hover copy; scatter omission with preserved zero/non-cost axes. Two prior expectations changed because zero denominators and unavailable pre-load cost now display Unavailable. No tests were removed.

Standalone binary verification and integrated make remain parent-owned after backend completion. No browser geometry claim is made; no browser automation was used.
