# Presentation publication

The static presentation and roadmap are published to GitHub Pages from `main`
only. A push to main publishes the current main revision. Pull requests do not
run this workflow and have no Pages preview. To republish main manually, or to
publish again after recovering an interrupted publication, dispatch the workflow
from main:

```sh
gh workflow run tickets-pages.yml --ref main
```

A dispatch from any other ref fails before building. The publishing job checks
out trusted workflow scripts, validates the static artifact and current main
revision, and never executes code from a build artifact. If main advances during
a build, the publication fails; the run for the newer push publishes it.

Earlier versions also published PR previews beneath `/AgenticLens/preview/<PR-number>/`.
The first main publication after their removal validates the retained previews
and then drops them, so the published site and the state branch keep only the
main portal.

## Durable state and interrupted publication

The `portal-pages-state` branch contains only the composed static site and a
small publication manifest. It is not a source branch. No raw sessions or ledgers
are copied. A shared publication lock covers reading state, composing, deploying
and saving state. It queues up to 100 pending publishers so consecutive main
publications do not replace one another in the pending queue. The provider cancels
additional runs when that queue is full; dispatch a new publication from main if one was canceled. This branch has no expiration; per-run build artifacts expire
after seven days and recovery artifacts after ninety days.

The publisher uploads the Pages artifact and checks that it can read it from
the same run. The artifact service can briefly report a finalized upload as
missing, so the check reads the same artifact up to three times over about 80
seconds. If the upload or every read fails, dispatch a fresh publication from
main. Do not use **Re-run jobs**: a re-run uploads a second `github-pages`
artifact, which the deploy action refuses after the pending marker is recorded.
The publish job therefore fails immediately on any re-run attempt. After the check
and before invoking deployment, the publisher commits
`.pending-publication.json` with the run ID and recovery artifact name. After Pages succeeds it commits the published
site and removes the pending marker. If deployment or final retention fails,
subsequent publication fails closed rather than overwriting the retained state with an
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
   dispatch a new publication from main. Verify the production URL.

Never infer a successful empty publication from an API, artifact or Git failure.

## Local verification

```sh
BUNDLE_GEMFILE=.github/pages/Gemfile bundle exec ruby .github/pages/tests/site_test.rb
python3 .github/pages/tests/preview_test.py
```

The Ruby suite exercises the production Jekyll entry point with temporary
repositories; the Python suite exercises static composition and failure paths.
