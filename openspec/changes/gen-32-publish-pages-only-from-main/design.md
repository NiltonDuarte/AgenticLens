# GEN-32: Technical design

## Scope and boundaries

The Pages workflow publishes only the main portal. It runs on pushes to main and on a manual dispatch from main, which serves for republication and recovery. The `pull_request_target` trigger, the `preview_pr` dispatch input, the PR candidate build job, preview composition and preview reporting are removed. Everything that serves the main publication stays: the main build, the serialized `portal-pages-publication` lock with its 100-run queue, the recovery artifact, the readability check, the pending marker before deployment, retention after deployment and the documented recovery.

The `portal-pages-state` branch is not edited by hand in this change. The `github-pages-preview` environment and any repository settings stay as they are; removing the unused environment is an administrator follow-up.

## Decisions and ownership

The GEN-32 implementation session owns `.github/workflows/tickets-pages.yml`, `.github/pages/preview.py`, `.github/pages/tests/preview_test.py`, `.github/pages/README.md` and this change.

**Retained previews are dropped.** Before this change, `compose` kept a retained preview only while its PR was open and targeted main, and dropped it otherwise. A main publication therefore already pruned previews of closed or retargeted PRs. It kept previews of still-open PRs, and with no preview publications left those would stay frozen at their last head until the PR closed. A PR that stays open indefinitely would keep a stale preview indefinitely. Now `compose` validates the retained state with the same checks as before, including every legacy preview, then publishes only the production build and records an empty preview set. The first main publication after merge removes every preview from the site and the state branch. Production is safe because it never depended on preview content: the composed site is a copy of the validated production build. An invalid retained state still fails before any change. `read_state` still accepts a non-empty preview set so the first publication can read the current state branch. `compose` no longer needs the open-PR listing, so `open_prs` is removed.

**Bounded readability retry.** The retry is needed for main publication, not only for PRs. Main-only [run 45068922715](https://github.com/NiltonDuarte/AgenticLens/actions/runs/45068922715) failed the check with "No artifacts found" 0.3 seconds after the upload finalized, with no candidate build. Three main-ref dispatches between 23:00 and 23:03 UTC on 2026-11-23 failed the same way. Without the check, GEN-28 showed that `deploy-pages` hits the same lag after the marker is recorded, which blocks every later publication. The check therefore stays before the marker and reads the same artifact up to three times: immediately, after 20 seconds and after another 60 seconds. Each read only downloads, so no second `github-pages` artifact is created. If all three fail, the job fails before the marker, as before.

**Re-run guard.** "Re-run failed jobs" uploads a second `github-pages` artifact in the same run. `deploy-pages` refuses it after the marker is recorded, as AL-97's re-run of run 45099570741 did. The publish job's first step now fails on any run attempt other than the first, before it loads state or uploads anything, and tells the operator to dispatch from main. The README describes dispatch from main as the only way to publish again.

**Ref guard.** `preview.py resolve` refuses any ref other than `refs/heads/main` before calling the API. Previously a dispatch from a feature ref without a PR number was rejected the same way. The `github-pages` environment stays restricted to main as a second barrier.

The script and test file keep their `preview` names to limit churn and keep historical links working. Their docstrings now describe main publication. `site_test.rb` still builds at a nested `/FixtureRepo/preview/24` base path. That covers the generic `--base-path` support in `build.rb`, which this change does not touch.

## Implementation approval

The user directed this change: "Can we remove this workflow from the PR? Keep it only on merges to main, it's causing too much problem." The orchestrating session passed the request on with the scope above, the instruction not to touch `portal-pages-state`, and the instruction to add a retry only if main publication needed it. That covers the implementation recorded here.

## Implementation approach

Remove the PR trigger, dispatch input, candidate job, preview environment selection and preview report from the workflow. Simplify `resolve`, `prepare` and `compose` to main-only. Replace the candidate composition tests with tests for main-only output, dropping legacy previews, validating legacy state, rejecting unsafe input and refusing non-main refs. Keep the state-branch load test. Verify the workflow with a YAML parse, a reference check and actionlint, then run the Pages suite and the integrated `make`.

Because `pull_request_target` workflows run from the base branch, this PR still triggers the old preview workflow from main once. The change takes effect after merge.
