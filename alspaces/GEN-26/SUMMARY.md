# GEN-26: OpenSpec compatibility assessment

AgenticLens is partly ready to use OpenSpec folders directly: setting `workspaces.dir: openspec/changes` discovers existing ledgers in active change folders, with the change slug as ticket ID. Standard proposal/design/tasks artifacts can be configured in place. A reliable full lifecycle needs archive-aware detail and writer resolution; a separate AgenticLens workspace is not the intended solution. The [assessment and gap map](ASSESSMENT.md) records supported primitives and the remaining work.

The user clarified the direct-storage requirement in [iteration 02](iterations/02/PLAN.md), superseding the earlier separate-folder pilot recommendation. The assessment ticket remains under the repository's existing configured root. Application code, tests, runtime configuration and OpenSpec installation are untouched; implementation remains unauthorized. This was a coordinator-led compatibility study without independent review or whole-application approval claims.

The [iteration 03 design assessment](iterations/03/PLAN.md) recommends explicit active/archive roots with immediate-child discovery and one locator shared by all readers and writers. Multiple roots alone do not prevent recreation: publication must require a unique existing change directory and must not synthesize a missing active path. Stable ticket/session binding can span those roots while retaining existing worktree boundaries. OpenSpec already accepts lowercase ticket-prefixed names such as `al-123-add-login`; no upstream change is needed to supply one. Canonical `AL-123` identity should be separate from the mutable folder name, with explicit handling of missing/duplicate identities. Configuration syntax and implementation remain unapproved.

Assessed AgenticLens `460c511ef96f9e42f2ecbfdba451828b52d3ce91` and OpenSpec `d12c2817776779e06891d68230773b3bb673b106` (source package version 2.0.0). Initial status was clean. GEN-26 was allocated after checking workspaces, index and fetched refs; the initial generated GEN-13 branch name belonged to an existing unrelated ticket, and GEN-25 also had an active branch.

Verification: current-source native build passed with `make build PLATFORMS=linux/amd64`; generate help confirmed supported flags. Disposable fixture probes confirmed Markdown-only changes are not discovered, slug IDs can load, and recursive scan/root CLI filtering retain archived ledgers. Static HTTP detail tracing identified the direct-directory limitation after archive. Copied fixture history was not treated as a fresh extraction. The full `make` gate passed (exit 0) after integration reported already up to date with origin/main: Go vet and race/integration suites, frontend unit/integration tests, typecheck/lint, release/install/capture/orchestration checks, and native builds. It started against unchanged baseline application source and finished on documentation commit `73ef749e`; no source changes occurred during verification. `git diff --check` and local Markdown-link resolution also passed. Only the GEN-26 artifact directory differs from origin/main.

Limits: no actual OpenSpec assisted workflow, browser interaction, live HTTP archive request, store integration or new harness adapter was exercised. Installed manager rejected the shared configuration at line 100 before help; current-source binaries were used without editing configuration. OpenSpec source revision is pinned; package publication of that SHA was not verified. Findings remain compatibility gaps and proposed work, not implemented fixes.

Delivery uses the freshly built manager and engine because the installed manager cannot parse the shared configuration. An initial Codex checkpoint was generated before the user clarified direct storage. It contains one native root session, source codex, and no delegated children. The final successful refresh retained that same native owner and included the direct-storage clarification; no other-harness history was present. Source-reported cost/caller/provisional limits remain explicit.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 4
- **Active Agent Duration:** 15m 2s (recorded duration including 124s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 132.2k tokens (29% of 450k healthy ceiling, 0 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 5,680 reasoning tokens (19.9% of generated output)
- **Tool Health & Self-Correction:** 57 tool calls (3 self-corrections, 94.7% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 117,531 | 22,802 | 5,378,816 | 97.9% | 132.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **117,531** | **22,802** | **5,378,816** | **97.9%** | **132.2k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 1 | 117,531 | 22,802 | 5,378,816 | 97.9% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **117,531** | **22,802** | **5,378,816** | **97.9%** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

Capture observation finalized: 2026-11-13T13:02:45.364235286Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 1; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 1; reserved bytes: 1551863. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1d4ea-9e1b-7242-909c-42d7410438f3: current; last successful source read 2026-11-13T13:02:45.351687647Z; accepted source checkpoint 2026-11-13T13:02:45.253Z.
<!-- agentic-lens:metrics:end -->

Iteration 02 verification: a self-contained disposable fixture with the actual `workspaces.dir: openspec/changes` configuration confirmed active and archived recursive discovery, without a CLI workspace override. After the archive-layout move, the direct detail path no longer existed. No live capture or HTTP/browser claim is made. The full make result remains valid because only ticket documents changed.

Final telemetry verification: the iteration 03 refresh exited 0 and retained the expected native Codex root `01a1d4ea-9e1b-7242-909c-42d7410438f3`, source codex, no parent and no delegations. No child agents were launched and no other-harness history was present. Generated provisional-duration, caller-verification and cost-availability fields are authoritative for this checkpoint; it does not claim a final closed-session measurement. The prior checkpoint owner was preserved.
