# Presentation and PR previews

The static presentation and roadmap share GitHub Pages. The root always builds
from the current `main` revision; an open same-repository PR targeting main is
published at `/AgenticLens/preview/<PR-number>/`. All local links and assets use
that prefix. Fork PRs are not published.

The trusted `pull_request_target` workflow resolves current API facts. PR code
runs in a separate read-only hosted job, without secrets, persisted checkout
credentials, or shared dependency caches. The publishing job checks out trusted
workflow scripts, validates static artifacts and current revisions, and never
executes PR code. The native Pages preview feature is not used.

## First PR bootstrap

Before this workflow reaches main, an administrator creates the
`github-pages-preview` environment restricted to the approved feature branch.
Keep the existing `github-pages` environment restricted to main. Open a PR, then
run the existing workflow on the PR's latest head:

```sh
gh workflow run tickets-pages.yml --ref GEN-24/telemetry-portal-preview-colors -f preview_pr=NUMBER
```

A feature-ref dispatch must match that PR's current head. The deployment summary
reports the preview URL and revision. Once merged, same-repository PR
open/update/reopen events publish automatically; closing or retargeting a PR prunes it on the
next successful publication. Dispatching from main without a PR also rebuilds
production and prunes closed previews. Stale builds fail; rerun to build current
main/PR revisions. Existing successful open-PR previews remain until replaced or
closed, including while a newer candidate is building or has failed.

## Durable state and interrupted publication

The `portal-pages-state` branch contains only the composed static site and a
small publication manifest. It is not a source branch. No raw sessions or ledgers
are copied. A shared publication lock covers reading state, composing, deploying
and saving state. It queues up to 100 pending publishers so independent PR
updates do not replace one another in the pending queue. The provider cancels
additional runs when that queue is full; rerun a canceled publication. This branch has no expiration; per-run build artifacts expire
after seven days and recovery artifacts after ninety days.

Before deploying, the publisher commits `.pending-publication.json` with the
run ID and recovery artifact name. After Pages succeeds it commits the published
site and removes the pending marker. If deployment or final retention fails,
subsequent publication fails closed rather than overwriting previews with an
older snapshot. Do not delete the state branch to bypass this check.

An administrator recovers the recorded run as follows, using a separate temporary
clone of `portal-pages-state` (never the application checkout):

1. Inspect the recorded run's `publish` job and the **Deploy ticket pages** step.
   A successful build alone does not prove a deployment.
2. If deployment succeeded, download that exact run's
   `portal-publication-recovery` artifact and validate its `.portal-state.json`,
   repository, revisions and static trees with `read_state` in `preview.py`.
   Replace only `site/` with that composition. If the artifact is unavailable,
   stop and reconstruct the deployed state before any new publication.
3. If deployment failed or never started, retain the previous `site/` unchanged.
   On the initial publication there may be no `site/`; remove the pending-only
   branch after confirming that no deployment occurred.
4. Remove `.pending-publication.json`, commit and push the recovered state, then
   rerun the intended publication. Verify the production and preview URLs.

Never infer a successful empty publication from an API, artifact or Git failure.

## Local verification

```sh
BUNDLE_GEMFILE=.github/pages/Gemfile bundle exec ruby .github/pages/tests/site_test.rb
python3 .github/pages/tests/preview_test.py
```

The Ruby suite exercises the production Jekyll entry point with temporary
repositories; the Python suite exercises static composition and failure paths.
