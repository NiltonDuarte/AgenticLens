# MGT-1 Review contract

Review every open AgenticLens pull request at its captured head, focusing on cross-PR conflicts, semantic interference, release readiness, and merge order. User requested a review and recommendations; do not merge or modify application code or post reviews to other PRs. Audit artifacts only. Initial Git status was clean.

## Accepted management workflow classification (iteration 02)

The user explicitly superseded the original GEN-14 classification with a new management ticket category. MGT-1 is the same active task, relocated without changing application code or tests. MGT is configured for management coordination, reviews and delivery planning, with conventional type chore. Current audit findings and verification remain accepted; implementation of fixes remains outside this task. Historical iteration 01 work was performed under GEN-14. The existing PR remains the delivery vehicle.

## Accepted supervision scope (iteration 03)

The user now explicitly authorizes user-proxy control of existing ticket threads and supervision/merge of all eligible non-release open PRs. This supersedes the earlier audit-only prohibition on messages and merges. Create a reusable management skill first. The manager owns coordination artifacts and merge gates; ticket owners retain application code, review corrections, independent roots and telemetry. Release PRs, including autorelease/release-please, are strictly human controlled: the manager and all managed agents must never merge or enable auto-merge for them. They are excluded from the agent-merge completion condition.
