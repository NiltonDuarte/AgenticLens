# ci(GEN-32): publish Pages only from main

## Outcome

The Ticket roadmap pages workflow now publishes only the main portal. It runs on pushes to main and on a manual dispatch from main. Pull request events no longer run it, the dispatch has no `preview_pr` input, and the PR candidate build and preview reporting are gone. The main build, the serialized publication lock, the recovery artifact, the pending marker, retention and the documented recovery are unchanged. The user directed the change; the decisions are in [design.md](design.md).

- **Retained previews.** A main publication already pruned previews of closed or retargeted PRs, but kept previews of open PRs. With no more preview publications, those would stay frozen until their PRs closed. The first main publication after merge now validates the retained state, then drops every preview from the site and from `portal-pages-state`. The state branch currently lists previews 160, 162 and 163. Production never depended on them.
- **Readability retry.** The race also affects main: main-only [run 45068922715](https://github.com/NiltonDuarte/AgenticLens/actions/runs/45068922715) failed the readability check with "No artifacts found" 0.3 seconds after upload, and three main-ref dispatches on 2026-11-23 failed the same way. The check now reads the same artifact up to three times, immediately, after 20 seconds and after another 60 seconds. It never uploads another artifact, and it still fails before the pending marker if every read fails.
- **Re-run guard.** The publish job fails at its first step on any re-run attempt and tells the operator to dispatch from main. That prevents the second `github-pages` artifact which stranded a pending marker after AL-97's re-run.
- `preview.py resolve` refuses any ref other than main before calling the API. The open-PR listing and the candidate and preview arguments are removed.

## Verification and review

Candidate `248f8920`, integrated with main `9eeabb94` in merge `c3083986`. Main's changes since the branch point do not touch `.github/pages` or the workflows.

- `python3 .github/pages/tests/preview_test.py`: 8 tests passed. They cover main-only output, dropping legacy previews, rejecting invalid or unsafe legacy state, rejecting an unexpected production payload, protected outputs, refusing a non-main ref, and the real `load-state` CLI against a temporary Git remote with a pending marker. A mutation that carried previews forward failed the preview-drop test.
- A copy of the live `portal-pages-state` tree, with previews 160, 162 and 163, composed against its own production files. `read_state` accepted the input, the output had no `preview/` directory and an empty preview set, and the output validated as a retained state.
- The workflow parsed as YAML with only `push` and `workflow_dispatch` triggers and no dispatch inputs. Every `needs`, job output and step reference resolved, and no preview, candidate or `pull_request_target` reference remains.
- actionlint 1.7.12 passed apart from one existing error, which it also reports on main: it does not recognize `concurrency.queue`, which GitHub supports and the workflow already used. shellcheck was not installed, so actionlint did not lint the inline scripts.
- `git diff --check`, `scripts/openspec.sh validate gen-32-publish-pages-only-from-main --strict` and `python3 scripts/tests/openspec_workflow_test.py` (3 tests) passed.
- The integrated `GOFLAGS=-timeout=35m GOMAXPROCS=6 VITEST_MAX_FORKS=2 VITEST_MIN_FORKS=1 make -j2` exited 0 on `c3083986`. It covered Go race integration, 231 frontend unit tests, 819 frontend integration tests, the Python boundary suites and the builds. The first run without `GOFLAGS` failed because `agenticlens/cmd/agentic-lens` exceeded Go's default 10-minute package timeout (607s) while another worktree ran its own `make`. Every other package passed in that run, and the same package passed in 540s on the rerun. This change does not touch Go code.
- No Make target covers Pages. Ruby is not installed locally, so `site_test.rb` did not run; this change does not modify it or `build.rb`, and the production job runs it on main.

This session reviewed the diff directly. No independent reviewer ran.

## Limits

- The new trigger set, retry and re-run guard run on GitHub only after merge. `pull_request_target` workflows run from main, so this PR still triggers the old preview workflow once.
- A main publication can still fail if the artifact stays unreadable for about 80 seconds. It fails before the pending marker, and a dispatch from main publishes again.
- The unused `github-pages-preview` environment remains in the repository settings for an administrator to remove.
- This session ran in Claude Code, which AgenticLens cannot extract yet, so this change has no `agentic_metrics.json`. No ledger was generated from another harness's session.

## Pull request

The Ticket roadmap pages workflow ran on every same-repository PR to publish a roadmap preview. Those runs kept failing the shared publish job, and each push queued another serialized publication, so one failure delayed or disturbed the others. This change publishes Pages only from main: on pushes to main and on a manual dispatch from main.

- Removes the `pull_request_target` trigger, the `preview_pr` dispatch input, the PR candidate build and preview reporting.
- The first main publication after merge validates and then drops the previews still on `portal-pages-state` (160, 162 and 163). Previews at `/AgenticLens/preview/<PR-number>/` stop working.
- The pre-deployment readability check reads the Pages artifact up to three times over about 80 seconds, because the same race also failed a main-only run. It still fails before the pending marker.
- The publish job refuses re-run attempts, which upload a second `github-pages` artifact that deploy-pages refuses after the marker is recorded. Dispatch from main instead.
- The Pages README describes main-only publication and recovery by dispatch.

Verification: the Pages composition suite (8 tests), a composition against a copy of the live state branch, a YAML and reference check, actionlint (apart from its existing `concurrency.queue` gap), strict OpenSpec validation, `git diff --check` and the integrated `make`. Telemetry is unavailable because this session ran in Claude Code.

Because `pull_request_target` workflows run from main, this PR still triggers the old preview workflow once. The change takes effect after merge; the first main run should be checked for a successful deployment and an empty preview set.
