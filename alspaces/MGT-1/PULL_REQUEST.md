# chore(management): coordinate AL ticket threads and safe PR integration (MGT-1)

Add the MGT ticket category and agentic-lens-management skill for independent ticket ownership, dependency-aware dispatch and exact-head merge gates. Release/autorelease PRs are strictly human controlled.

Preserve the original review under MGT-1 and record verified merges of #76, #79, #75, corrective #82, #77 and #78. The controller routed a reproduced SSE startup race to its existing AL-47 owner, then required the corrected base and fresh full gates before continuing. #81 was human-closed without merge. User-requested branch rename auto-closed original management #80; this replacement retains its work and later records.

New AL-55 development planning and PO-17 review use verified isolated T3/Git worktrees with separate telemetry. AL-55 code remains held for explicit acceptance; AL-57 costs are deferred and AL2 development remains queued per user steering. CONTROL.md records current heads, merge SHAs, questions and owners.

Management config/guidance/artifacts only. See iterations/03/REVIEW.md and iterations/04/REVIEW.md for evidence and limits. Shared-branch wording was clarified after independent-worktree verification. Ongoing backlog supervision does not imply approval of new product scope.
