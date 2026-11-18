# Protected Codex callback selection

The approved iteration 03 contract covers protected operator installation, callbacks, worker lifetimes and mixed-source coordination. The implementation is at `aa47dc49`; test additions are at `7fae9556` and `63aa4125`. Focused native tests reached the normal hook and worker, late source availability, corrupt progress recovery, and preservation of pending work with a missing or modified engine.

At this revision the architecture and permission reviewers using gpt-6.2-sol/high independently identified the same OpenCode regression. The machine-level Codex inbox receipt changes dispatch for a separate OpenCode-only worktree. A protected worktree needs to retain its admission independently of sandbox-editable project settings. Its admission must also leave a separate worktree’s supported OpenCode installation usable. The coordinator assigned the correction and a two-worktree regression test.

The behavior reviewer using gpt-6-luna/medium confirmed bounded source retries and worker handoff from source inspection. The permission reviewer using gpt-6-luna/medium approved the inspected executable and private-state path checks. Both noted that native hook payloads do not report additional per-launch grants. Operators control those grants. The actual installation requires a native sandbox permission check and fresh-root acceptance.

## Worktree admission at `3fde4c04`

The manager now dispatches ordinary OpenCode callbacks according to their worktree. A separate linked worktree retained its own engine when the Codex installation was unavailable, while nested projects inside an admitted worktree retained protected routing. Targeted native tests passed under the race detector and across three repeats. The incompatible-engine fixture preserved both the queued response and the ledger. Two distinct worker PIDs carried the same protected build and completed the bounded binding sweep.

The behavior, data, architecture, test and permission reviewers using gpt-6-luna/medium approved their scoped inspections. The architecture reviewer using gpt-6.2-sol/high confirmed the original OpenCode regression was fixed. The following observations from the gpt-6.2-sol/high verification pass remain open until corrected.

| Discovery | Impact |
| --- | --- |
| Permission | The worker checks publication state before launching a writer that reads the path again. A later configuration edit can put operational files in a writable directory. |
| Behavior and architecture | Unrelated worktrees contend for a single first-admission lease. An OpenCode lifecycle callback can fail before registration and the plugin does not retry that event. |
| Behavior | Explicit reconciliation can skip a queued response in a worktree without an admission record after an upgrade. |
| Data | An inactive Codex receipt can be visible before the uninstall journal commits. Routing can release an admission that recovery still has to restore. |
| Tests | The independent-worktree native manager fixture reaches OpenCode registration. It needs an assertion on source-backed usage in the published ledger. |

The earlier gpt-6.2-sol/high and gpt-6-luna/medium test coverage objections about successor identity and incompatible protocol retention were resolved at this revision. Native fixtures use isolated installations. The operator rollout requires a separate fresh-root result.

## Corrected writer, routing and recovery at `8f400b9c`

The exact protected writer requires worker protocol 2 and validates its admitted project, Codex home and physical publication state at acquisition. Data review verified that interrupted uninstall retains worktree admission until commit. Behavior review confirmed that scoped leases allow sibling callbacks during first admission, and explicit reconcile admits an enabled upgraded worktree with a previously queued response. The tester verified independent OpenCode publication against an actual SQLite message with 47 input tokens, in its own ledger while the other worktree's protected engine was unavailable.

The gpt-6.2-sol/high permission, behavior, data and test reviewers closed the corresponding findings. The gpt-6-luna/medium correction reviewer approved the changed behavior, data, architecture, permissions and test boundaries and reran their targeted races. A protected writer regression covered a private-to-private and private-to-worktree state change, prior ledger bytes and pending work. The changed-path fixture does not inject interruptions into the publication journal. Earlier coverage of unrelated code remains scoped to its recorded revision.

The operator approved repair of a modified live manager. After backing up the live program and its sidecars, the coordinator verified the previous program against the receipt, restored it and upgraded at the same path through `self-update`. The protected manager and engine match release `v3.0.0-rc.150+0e9c8b3-20-g8f400b9c`. Installation changed only the inbox, its grant, the hook file and its ownership receipt. A second preview was clean. An OS permission check under native Codex workspace-write policy found the worktree and inbox writable and the manager, manager directory, exact engine, capture state and publication state read-only. This check applies to this installation's grants.

The user reviewed and trusted the new PostToolUse entry. A fresh native Codex root submitted its hook-issued response in the original worktree. Status identified the correct GEN-27 title and the protected executable. The existing ledger initially held ten owners. The worker reported conflicting explicit titles on one older registration and withheld publication. The operator authorized using the ledger title, but the normal bind command declined because the stored active rollout path is absent. The exact archived rollout has the original session ID and worktree. The historical registration and archive are unchanged pending approval for temporary native unarchive. This is separate from the takeover conversation’s starting worktree, where the GEN-27 ledger is absent.

Full `make -j2` passed at `8552a8c7` after integration with main `9308ac77`. Its runtime is identical to the live `8f400b9c` build. Only two test fixtures changed to provision their intended ordinary and protected native installations. Both retained their assertions and passed twice with the race detector. Strict OpenSpec validation passed. Fresh-root publication remains open, so there is no consolidated delivery approval.
