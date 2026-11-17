# chore(GEN-27): verify capture and the OpenSpec development workflow

## Pull request

This audit exercised pinned OpenSpec 2.0.0 and native Codex 0.169.0 and OpenCode 1.19.5/2.2.0 against fetched main `b6ab0471`. It records the supported CLI sequence, native event-to-ledger evidence, and what happens when source data or a former change path disappears. The audit proposes guidance corrections. No application or runtime change was made.

- **First publication is the operational gap.** An installed, trusted Codex hook recorded a real root and child. Both were bound to the correct ticket, yet the worker could not publish into the newly created OpenSpec change until an explicit generation supplied its first ledger path. The setup manual describes automatic publication without explaining this condition. The setup/operator and workflow owners should document capture pending before a first ledger and include the required human-readable title in first-generation examples. Repository policy reserves final telemetry for delivery. If first-ticket live publication is a product requirement, authorize a separate design for selecting the destination before the final extraction.
- **The CLI can carry the development lifecycle.** Use the creation helper, then pinned status and per-artifact instructions. Read apply instructions and obtain explicit scope approval before coding. Run strict validation alongside application verification and consolidated review. Read archive instructions, invoke the pinned archive command and use the returned location for subsequent operations. The generated propose and archive skills route around repository-specific creation checks and the supported archive operation. The workflow and setup owners should reconcile those entry points.
- **Published history followed the ticket.** Native Codex resume and the v1 OpenCode plugin published to moved homes. Unavailable owners survived, including exact parentage. OpenCode v2 manual extraction succeeded. Global plugin and cross-worktree implementation remain AL-80. A worker-control test twice missed its terminal shutdown deadline after its publication assertions passed. Its cause needs separate investigation by the capture-worker owner.

## Evidence

- [Observed CLI stages, prior delivery evidence and guidance findings](evidence/iterations/01/WORKFLOW.md)
- [Native event, binding, worker and publication results](evidence/iterations/01/NATIVE_CAPTURE.md)
- [Moved-ledger and unavailable-source controls](evidence/iterations/01/PUBLICATION.md)

The native investigators used loopback providers with synthetic usage. They exercised actual client runtimes and source stores. The publication investigator separately used synthetic rollouts with a real writer. These bounded checks do not establish provider pricing, global OpenCode callbacks or every runtime/profile combination. The coordinator consolidated the three investigations as audit evidence. There is no application-change review verdict.

## Integrated checks

Full `make` exited 0 after integrating fetched main into `892a7473`. Application, build and test inputs were identical at the evidence commit `97d2c271`. The run passed Go race and integration packages, 231 frontend unit tests, 804 frontend integration tests, distribution checks and four native build targets. The passing integrated run did not resolve the focused worker shutdown failures. Strict GEN-27 validation and relative evidence links passed. The audit change remains active.

## Audit session telemetry

A private snapshot of the four actual GEN-27 Codex rollouts supplied the final extraction. The ledger contains the coordinator and all three native child IDs with matching parent edges. Costs were unavailable. The source reported the coordinator turn open at extraction and the invoking caller unverified. Disposable probe sessions belong only to the capture evidence.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 3
- **Active Agent Duration:** 39m 31s (recorded duration including 1232s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 101.3k tokens (23% of 450k healthy ceiling, 0 compactions) | Subagents peak 79.4k (Opencode)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Native 1.0, Opencode 1.0, Publication 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 2,612 reasoning tokens (6.5% of generated output)
- **Tool Health & Self-Correction:** 174 tool calls (8 self-corrections, 95.4% success rate)

### Anomalies & Health Flags
- **Status:** 1 execution anomaly detected (1 WARNING).
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: native <-> opencode (3 cycles) — Subagent roles 'native' and 'opencode' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 189,707 | 15,095 | 5,992,704 | 96.9% | 101.3k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **native** | 1 | 82,961 | 7,538 | 2,619,648 | 96.9% | 75.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **opencode** | 1 | 64,050 | 9,170 | 2,752,000 | 97.7% | 79.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **publication** | 1 | 58,636 | 5,900 | 2,597,632 | 97.8% | 75.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **4** | **395,354** | **37,703** | **13,961,984** | **97.2%** | **101.3k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.2-astra** | 4 | 395,354 | 37,703 | 13,961,984 | 97.2% | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **Total** | **4** | **395,354** | **37,703** | **13,961,984** | **97.2%** | **Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations)** |
<!-- agentic-lens:metrics:end -->
