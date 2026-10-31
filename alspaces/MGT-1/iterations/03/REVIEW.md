# MGT-1 Iteration 03 verification and supervision

Skill quick_validate passed. Behavioral scenario inspection: a ready feature requires current exact-head gates; a release stays human_only even when all CI passes; missing owner remains unmapped; a changed head invalidates readiness; independent owners retain telemetry; sequential integration follows verified merges. No product code/tests/scripts were changed by the manager.

Canonical AGENTS.md now names the management role and release exclusion, linking the skill rather than duplicating its workflow. Earlier no-message/no-merge audit boundary was explicitly superseded by the user's latest request in CONTRACT.md. The management skill applies to this iteration.

Actual PR/task mappings, instructions and merge evidence are maintained in CONTROL.md. PR #81's original thread read returned thread_not_found; neither a successful owner lookup nor a delegated task is fabricated. Repository merge policy allows squash only. #76 has no GitHub review/comments at the first check; prior independent code review and full make evidence are in its ticket thread/workspace. Fresh checks and unresolved discussions still gate any merge.
