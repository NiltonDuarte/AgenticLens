# MGT-1 Iteration 03 verification and supervision

Skill quick_validate passed. Behavioral scenario inspection: a ready feature requires current exact-head gates; a release stays human_only even when all CI passes; missing owner remains unmapped; a changed head invalidates readiness; independent owners retain telemetry; sequential integration follows verified merges. No product code/tests/scripts were changed by the manager.

Canonical AGENTS.md now names the management role and release exclusion, linking the skill rather than duplicating its workflow. Earlier no-message/no-merge audit boundary was explicitly superseded by the user's latest request in CONTRACT.md. The management skill applies to this iteration.

Actual PR/task mappings, instructions and merge evidence are maintained in CONTROL.md. PR #81's original thread read returned thread_not_found; neither a successful owner lookup nor a delegated task is fabricated. Repository merge policy allows squash only. #76 has no GitHub review/comments at the first check; prior independent code review and full make evidence are in its ticket thread/workspace. Fresh checks and unresolved discussions still gate any merge.

#76 merged after fresh exact-head checks and no-review-thread evidence; merge commit is in CONTROL.md. #79 integration released, #78 review repairs active with upstream hold. #81 is documentation-only and its current content has been reviewed; the manager requested a guarded update from main, leaving new-head CI pending. Release #55 remains unchanged and human_only.

Supervision checkpoint verification: full make exited 0 at integrated head 7b8833d (main #76 included), with Go vet/race suites, frontend typecheck/lint, all 722 frontend tests and four platform builds. Embedded assets match. Skill quick_validate and diff whitespace checks pass. Only management records changed while verification ran. Completion remains pending for #79/#75/#77/#78/#81/#80. Telemetry at this checkpoint includes only the management native root; managed independent roots are excluded.

Supervision uncovered a deterministically reproduced HTTP subscription startup race inherited from AL-47, not AL-53 frontend integration. Original failure retained. Existing AL-47 owner resumed for a separate fix PR before #77; manager changed no product code and did not waive the gate. See CONTROL.md for source proof and instruction IDs.

Inspected #82 real fix/test diff at c7a36d0, including deterministic subscription order, initial-tree ordering, mutation/REST retention and concrete cleanup/error cases. Owner verified old code fails; full local gate/review pass. Exact-head GitHub CI pending, so manager did not merge. #77 must integrate this verified correction and re-run its full gate before readiness.

Original product merge sequence completed through verified #78 mergea77695c. Startup regression fixed in separate#82 before allowing#77/#78 integration. All six product/corrective PRs merged behind exact-head gates; human-closed#81 not counted and human-only release#55 untouched. Replacement management#83 pending its own final sync/checks.
