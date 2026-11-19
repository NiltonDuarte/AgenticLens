# feat(GEN-27): bind Codex sessions through an isolated inbox

## Pull request

A fresh Codex workspace-write session can submit its ticket through a dedicated inbox. The installer grants access to that directory once. A trusted worker checks the response and verifies the native session before binding the ticket and publishing its ledger and summary. The `codex-binding-wake` PostToolUse callback wakes an idle worker after a submission.

Trusted Codex callbacks now select a checksum-verified operator-installed engine outside the worktree. Development worktrees use the stable manager’s matching companion. The worker retains this selection through restart and publication. Independent OpenCode worktrees keep their configured engine. Operators must review the new hook in Codex and keep private installation and state paths outside any extra launcher grants.

## Application verification

Runtime revision `8f400b9c` includes the protected manager, engine and guarded publication path. The focused race checks passed for installer recovery, same-worktree admission, independent OpenCode publication with a real SQLite source, late binding, incompatible engines, distinct worker PIDs and guarded publication state. Two older fixtures needed explicit selection for their temporary installations. They passed twice with the race detector at `8552a8c7`, whose changes are confined to test setup. Full `make -j2` passed on an isolated checkout of `8552a8c7` after integrating main `9308ac77`. It covered Go race tests, 231 frontend unit tests, 819 frontend integration tests, distribution checks and platform builds. Strict OpenSpec validation passed. See [inbox review](evidence/iterations/02/REVIEW.md) and [protected engine review](evidence/iterations/03/REVIEW.md).

The operator approved repair of the stable manager and trusted the new hook. The updated manager and engine match their installation records. Repeat installation preview was clean. An actual workspace-write sandbox had access to the inbox and worktree and lacked write access to the manager, protected engine, capture registry and publication state. A fresh native Codex session submitted its own response and the protected worker bound it to GEN-27. The original publication stopped on two display titles for the same ticket ID. The correction at `ee83fc5b` preserves the existing ledger title and lets the worker publish sessions with different display hints. The native regression passed at `dd02db8b` and reproduced the original error against `7cb700ec`. Full `make -j2` passed on the integrated `dd02db8b` snapshot.

The protected build at `ee83fc5b` was installed through the stable manager and retried against the actual GEN-27 worktree. The worker reached generation and attempted to write both artifacts. Three previously cached session snapshots have a blank required `user` field. The writer rejected the proposed ledger and restored the accepted ten-session ledger and summary. The already verified fresh binding remains pending. The archived conversation and its registered display hint were preserved.

## Original audit

This audit exercised pinned OpenSpec 2.0.0 and native Codex 0.169.0 and OpenCode 1.19.5/2.2.0 against fetched main `b6ab0471`. It records the supported CLI sequence, native event-to-ledger evidence, and what happens when source data or a former change path disappears. The audit proposes guidance corrections. No application or runtime change was made.

- **First publication is the operational gap.** An installed, trusted Codex hook recorded a real root and child. Both were bound to the correct ticket, yet the worker could not publish into the newly created OpenSpec change until an explicit generation supplied its first ledger path. The setup manual describes automatic publication without explaining this condition. The setup/operator and workflow owners should document capture pending before a first ledger and include the required human-readable title in first-generation examples. Repository policy reserves final telemetry for delivery. If first-ticket live publication is a product requirement, authorize a separate design for selecting the destination before the final extraction.
- **The CLI can carry the development lifecycle.** Use the creation helper, then pinned status and per-artifact instructions. Read apply instructions and obtain explicit scope approval before coding. Run strict validation alongside application verification and consolidated review. Read archive instructions, invoke the pinned archive command and use the returned location for subsequent operations. The generated propose and archive skills route around repository-specific creation checks and the supported archive operation. The workflow and setup owners should reconcile those entry points.
- **Published history followed the ticket.** Native Codex resume and the v1 OpenCode plugin published to moved homes. Unavailable owners survived, including exact parentage. OpenCode v2 manual extraction succeeded. Global plugin and cross-worktree implementation remain AL-80. A worker-control test twice missed its terminal shutdown deadline after its publication assertions passed. Its cause needs separate investigation by the capture-worker owner.

## Evidence

- [Observed CLI stages, prior delivery evidence and guidance findings](evidence/iterations/01/WORKFLOW.md)
- [Native event, binding, worker and publication results](evidence/iterations/01/NATIVE_CAPTURE.md)
- [Moved-ledger and unavailable-source controls](evidence/iterations/01/PUBLICATION.md)

The native investigators used loopback providers with synthetic usage. They exercised actual client runtimes and source stores. The publication investigator separately used synthetic rollouts with a real writer. These bounded checks do not establish provider pricing, global OpenCode callbacks or every runtime/profile combination. The coordinator consolidated the three investigations as audit evidence. There is no application-change review verdict.

## Original audit checks

Full `make` exited 0 after integrating fetched main into `892a7473`. Application, build and test inputs were identical at the evidence commit `97d2c271`. The run passed Go race and integration packages, 231 frontend unit tests, 804 frontend integration tests, distribution checks and four native build targets. The passing integrated run did not resolve the focused worker shutdown failures. Strict GEN-27 validation and relative evidence links passed. The audit change remains active.

## Audit session telemetry

A private snapshot of the four actual GEN-27 Codex rollouts supplied the final extraction. That snapshot recorded the coordinator and all three native child IDs with matching parent edges. Costs were unavailable. The source reported the coordinator turn open at extraction and the invoking caller unverified. Disposable probe sessions belong only to the capture evidence.

The accepted telemetry baseline below holds ten Codex sessions and preserves the original audit owners and parent links. At that checkpoint, five sessions retained older facts and child discovery was still pending. Cost is unavailable. The fresh binding and this handoff remain outside that checkpoint while live publication is pending.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 23
- **Active Agent Duration:** 1h 43m (recorded duration including 462s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 230.5k tokens (51% of 450k healthy ceiling, 0 compactions) | Subagents peak 243.7k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.6 avg turns/run (peak: 3 turns) | Planner 3.0, Coder 1.5, Tester 2.0, Native 1.0, Opencode 1.0, Publication 1.0
- **Test-to-Production Code Ratio:** 0.5:1 (+430 test lines / +923 prod lines)
- **Code Delivered:** +1,353 / -51 lines across 16 code files (excluding docs)
- **Reasoning Effort:** 17,242 reasoning tokens (13.6% of generated output)
- **Tool Health & Self-Correction:** 402 tool calls (16 self-corrections, 96.0% success rate)

### Anomalies & Health Flags
- **Status:** 1 execution anomaly detected (1 WARNING).
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: native <-> opencode (3 cycles) — Subagent roles 'native' and 'opencode' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 202,216 | 5,551 | 3,967,232 | 95.2% | 198.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 2 | 145,384 | 24,888 | 15,741,696 | 99.1% | 243.7k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **tester** | 1 | 70,186 | 11,376 | 8,834,048 | 99.2% | 235.7k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 3 | 509,259 | 45,524 | 31,006,080 | 98.4% | 230.5k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **native** | 1 | 82,961 | 7,538 | 2,619,648 | 96.9% | 75.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **opencode** | 1 | 64,050 | 9,170 | 2,752,000 | 97.7% | 79.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **publication** | 1 | 58,636 | 5,900 | 2,597,632 | 97.8% | 75.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **10** | **1,132,692** | **109,947** | **67,518,336** | **98.4%** | **243.7k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.2-astra** | 8 | 1,093,238 | 109,216 | 67,466,240 | 98.4% | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **gpt-6.2-sol** | 2 | 39,454 | 731 | 52,096 | 56.9% | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **10** | **1,132,692** | **109,947** | **67,518,336** | **98.4%** | **Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations)** |

Capture observation finalized: 2026-11-18T20:21:37.255178861Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 5; retained older facts: 5; pending: 5; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 3; parent links: 0; discovery continuation: true. Source attempts: 10; reserved bytes: 536870912. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1eb9b-c88f-79b0-9cbf-15289f0cc44c: pending; last successful source read 2026-11-18T20:21:00.901783891Z; accepted source checkpoint 2026-11-18T20:21:00.283Z.
Owner codex/01a1eb9c-5404-7450-95f7-3e6f179abcc8: pending; last successful source read 2026-11-18T20:21:02.208670043Z; accepted source checkpoint 2026-11-17T22:29:59.244Z.
Owner codex/01a1f001-a475-7f02-a49a-455007a397d2: pending; last successful source read 2026-11-18T20:21:04.149076835Z; accepted source checkpoint 2026-11-18T20:20:51.638Z.
Owner codex/01a1f04b-532f-77c1-9d1a-86b03ec2748f: pending; last successful source read 2026-11-18T20:21:05.407573936Z; accepted source checkpoint 2026-11-18T20:20:29.556Z.
Owner codex/01a1ef94-247e-73d0-b3b5-1586cbc07510: current; last successful source read 2026-11-18T20:21:31.452088594Z; accepted source checkpoint 2026-11-18T16:54:00.436Z.
Owner codex/01a1f001-1c5c-7c23-8bd5-c71743d5fc7a: current; last successful source read 2026-11-18T20:21:33.06241345Z; accepted source checkpoint 2026-11-18T18:53:01.406Z.
Owner codex/01a1f000-f0fa-7d33-bbc4-5f23c0060caa: current; last successful source read 2026-11-18T20:21:34.190067603Z; accepted source checkpoint 2026-11-18T20:15:18.263Z.
Owner codex/01a1eb9c-40c0-76a1-9878-d2ab29a44496: current; last successful source read 2026-11-18T20:21:35.489832176Z; accepted source checkpoint 2026-11-17T22:30:59.89Z.
Owner codex/01a1eb9c-2fe2-78f2-92bf-7be89cbcfd02: current; last successful source read 2026-11-18T20:21:36.832173525Z; accepted source checkpoint 2026-11-17T22:30:26.32Z.
Owner codex/01a1f001-8aec-7eb1-8609-5f2e0104c4ba: pending; last successful source read 2026-11-18T20:20:59.265345552Z; accepted source checkpoint 2026-11-18T20:20:17.36Z.

Retained recorded history: 5 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
