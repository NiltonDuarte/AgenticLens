# GEN-22: Capture installation and worktree troubleshooting

## Delivered repository changes

- Recorded the requested capture vision, current gaps and unresolved decisions in [CAPTURE_VISION_AND_GAPS.md](CAPTURE_VISION_AND_GAPS.md). Proposed configuration precedence is explicitly distinguished from accepted requirements.
- Updated shared `agenticlens.version` to the existing local build's exact version, `v3.0.0-rc.131+03f5b05-2-g57a3503e`, with explicit user authorization. The earlier pin had no matching published release or local cache leaf.
- Prepared a draft PR. No application code, installer fix or distributed integration is implemented by these changes.
- The user requested deleting formal plans and continuing troubleshooting together. Both temporary plans and empty iteration directories were removed; no contract was created.

## Environment and investigation evidence

Initial Git status was clean at `57a3503e3024f3993ce9391fee604aad0bcc44d5`; the initial branch was literally `GEN-<ID>/install-capture-plugin`. The delivery branch is `chore/GEN-22-capture-worktree-findings`.

The environment is Linux amd64 on WSL2, with Codex CLI 0.167.0 and source records under `~/.codex`. Configured project is AgenticLens, workspace root is `alspaces`, and shared capture state is `~/.agenticlens/state`.

`make verify-build PLATFORMS=linux/amd64` passed, including the canonical frontend build, embedded asset synchronization, host-native compilation and version-command checks. The local binary identifies commit `57a3503e` and hook protocol 3. Its SHA256 is `e16b0c141e2729654fbb0df107f7d00ee5e0a81a2eac0ccba61d8b77061b6e1a`.

Installed the native build offline using the existing `--local-bin` installer and the explicit checksum after matching the shared pin. This proved that no new development installation feature or published release was needed for this installation. Selection independent of the shared pin remains a separate discussion point.

Codex showed zero hooks despite consistent worktree-local installation files. Adding trust for this exact worktree, explicitly authorized by the user, did not resolve discovery. Source-path comparisons in isolated ordinary repositories and attached/detached worktrees established the cause: Codex 0.167.0 reads linked-worktree project hooks from the original checkout. The actual original checkout lacked hooks. The exact six AgenticLens groups load correctly in an ordinary repository. Both inspected Codex 0.166.0 and 0.167.0 returned zero hooks on the actual worktree before the operational repair.

The parent and a planning investigator, both gpt-6.1-sol with high reasoning effort, established and checked the discovery evidence. Isolated controls ruled out hook syntax, worktree trust alone, stale user configuration, branch spelling and generic ~/.t3 ancestry. An initially misleading positive worktree control was corrected by checking returned sourcePath rather than counts. These were troubleshooting investigations, not a formal product-code review.

## Machine-local operations and current limits

The following are not repository deliverables:

- Stable wrapper at `~/.local/bin/agentic-lens` and the matching native version under `~/.agenticlens/bin/`.
- Fresh capture state under `~/.agenticlens/state/`. The user explicitly authorized deleting legacy databases and then all old state/recovery folders while preserving bin; capture was reinstalled afterward. No raw Codex source sessions were deleted.
- The approved worktree trust entry in `~/.codex/config.toml`. Individual hook trust was not granted automatically.
- A guarded original-checkout hook file briefly demonstrated native discovery. It was subsequently deleted at the user's request.
- Six user-level hooks at `~/.codex/hooks.json`, invoking a local launcher at `~/.codex/hooks/agenticlens-capture.py`. The Python file was created during troubleshooting; it is not shipped AgenticLens code. The user's target is native routing without this workaround.
- Worktree `.codex/hooks.json` and `.codex/agenticlens-installation.json`, retained because current receiver/launcher validation depends on them. Both were unstaged and excluded through repository-local `.git/info/exclude` at the user's request. No shared ignore policy was changed.

The home launcher selects an existing active worktree installation and its pin, source and state. Read-only checks selected this worktree and a subdirectory, skipped uninstalled/inactive projects and rejected a mismatched-worktree fixture. Native Codex 0.167.0 subsequently listed exactly six user-source hooks with no warnings/errors or duplicates. All six were untrusted at the last discovery check.

Live callback delivery, an explicit session-binding exchange, source extraction and automatic publication remain unverified. No fabricated callback or session was used to claim those outcomes. The local setup does not implement automatic adoption of new configured worktrees or base-checkout local-configuration fallback; those gaps are documented for further collaboration.

## Delivery verification and telemetry

Repository changes are limited to the approved pin and ticket documentation. Delivery integration/check results and session extraction facts are recorded below when available. No implementation review or platform-wide compatibility claim is made.

Delivery checks on `1ce7cc25` after fetching and merging `origin/main`: upstream was already current, and the complete `make` gate passed. This covered Go/frontend units (190 frontend unit cases), vet, TypeScript checking/lint, release/install/capture-distribution/orchestration checks, Go domain/adapter/CLI race suites, 678 frontend integration cases and the multi-platform build. Frontend output included non-failing React act warnings and the bundle-size advisory. Generated assets/binaries remain ignored.

Final telemetry source is Codex at the normal `~/.codex` home. The root rollout metadata identifies session `01a1b268-f31e-7503-a954-4f4fd77370b2`, this exact worktree, source `vscode` and originator T3 Code. Source metadata proves one planner child, `01a1b291-2ff1-7830-81d8-7d0dff387ee3`, through its parent_thread_id. No prior GEN-22 ledger or other-harness sessions exist to preserve. The independent diagnostic CLI roots are not inferred to be children. Extraction is performed immediately before the final telemetry commit and checked against these source identities; source availability is distinct from live hook verification.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 35
- **Active Agent Duration:** 49m 24s (recorded duration including 491s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 256.2k tokens (57% of 450k healthy ceiling, 0 compactions) | Subagents peak 208.5k (Planner)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Planner 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 38,197 reasoning tokens (29.9% of generated output)
- **Tool Health & Self-Correction:** 152 tool calls (11 self-corrections, 92.8% success rate)

### Anomalies & Health Flags
- **Status:** 2 execution anomalies detected (1 CRITICAL, 1 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+163.4k tokens) — Context tokens surged by 163.4k to 163.4k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (254.0k tokens) — Context tokens reached 254.0k, breaching the 250k warning ceiling.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 388,552 | 16,674 | 5,996,160 | 93.9% | 208.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 469,961 | 73,000 | 28,805,632 | 98.4% | 256.2k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **2** | **858,513** | **89,674** | **34,801,792** | **97.6%** | **256.2k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 2 | 858,513 | 89,674 | 34,801,792 | 97.6% | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **2** | **858,513** | **89,674** | **34,801,792** | **97.6%** | **Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations)** |

Capture observation finalized: 2026-11-07T13:06:40.604259309Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 0; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 0; reserved bytes: 0. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.

Ledger validation confirmed schema 6, ticket GEN-22/type chore, exactly the expected Codex root and planner child, and the source-proven parent link. The generator reports an open/provisional root interval and unknown root role; those source limitations are retained rather than treated as finalized execution or fabricated role data. This is a manual source snapshot, not proof that home hooks ran. Raw transcripts remain outside Git.
