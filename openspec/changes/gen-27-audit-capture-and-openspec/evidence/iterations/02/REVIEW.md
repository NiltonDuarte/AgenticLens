# Codex binding inbox review

The candidate at `c67f1528` requires corrections. The preceding audit and its passing CI cover an earlier revision. This record recovers the reviewers’ returned evidence from the original GEN-27 thread. The temporary individual reports were unavailable in the resumed workspace.

| Finding | Discovery | Evidence |
| --- | --- | --- |
| A response arriving after the worker exits for inactivity waits for another callback. | Behavior, data and architecture, gpt-6.2-sol/high | Disposable workers and CLI runs. A later callback consumed the response. |
| A missing Codex inbox prevents OpenCode registration. | Behavior and architecture, gpt-6.2-sol/high | Disposable mixed-source configuration. Restoring the inbox allowed registration. |
| Rollback can delete an empty inbox created by another writer. | Data, gpt-6.2-sol/high | Transaction checkpoint after the competing directory creation. |
| A config edit racing with install or uninstall can be overwritten. | Data and architecture, gpt-6.2-sol/high | Source inspection of the last validation and replacement steps. |
| The installer may crash on an empty sandbox settings table. | Data, gpt-6-luna/medium | Refuted on this revision. Independent installer and test fixtures parsed an empty table at end of file and preserved the setting when adding the inbox. The installer check also covered inline and adjacent tables. |

Behavior review under gpt-6-luna/medium returned a scoped approval from inspection and reported that its test did not finish. The first Luna behavior and data runs ended at ticket-binding questions. The original thread stopped before completing the architecture, test and security roster or issuing a consolidated verdict.

## Corrected candidate at 3eb30200

The focused race tests passed with a native idle-worker wake fixture, mixed-source publication, concurrent settings saves, and an actual six-to-seven group upgrade. Integrated make reached the distribution tests. Four assertions across three test methods still expected the previous six callback groups.

| Focus | gpt-6.2-sol/high | gpt-6-luna/medium |
| --- | --- | --- |
| Behavior | A late eligible registration can starve when each bounded worker starts its scan at zero. | Changes requested: recheck current inbox and private-state separation before acknowledging submission. |
| Data | Changes requested: cursor starvation and source verification that can exceed its per-pass allowance. | Approved in scope by inspection. |
| Architecture | The separation check omits the selected manager and native engine locations. | Approved in scope by inspection. |
| Tests | Pending on the next candidate. | Incomplete: verify replacement-entry preservation and worker restart. |
| Security | Pending on the next candidate. | Approved in scope by static inspection. |

The coordinator confirmed a cleanup race by source inspection. After accepting a response, the worker unlinks its original pathname, which can now name a newer entry. The existing process fixture waits for the initial worker to release its lease, submits a response, then verifies that the manager wakes a new worker and publishes. That establishes a process restart for the immediate submission case. The cursor finding identifies a distinct case with multiple registrations and one slot per worker pass.

### Live installation boundary

On this worktree the installed manager’s `--al-resolve` selected `bin/agentic-lens-engine-linux-amd64` inside the repository. A permission-only probe in a native Codex workspace-write sandbox reported write access to both that engine and its parent directory. The trusted callback uses the manager’s selected engine. Protecting the dedicated inbox alone therefore cannot establish executable isolation on this installation. Live installation awaits a decision on using an operator-controlled engine for trusted callbacks. The check inspected permissions without changing the executable or running an attack.
