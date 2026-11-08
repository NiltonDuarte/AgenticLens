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

## Configuration discussion decision

User selected whole-file local fallback: worktree local file wins in full, base local is considered only if that file is absent, and individual settings never inherit from the base local file. The active checkout must have .agenticlens-config.yaml; if missing, silently skip capture without using the base shared file. Recorded in CAPTURE_VISION_AND_GAPS.md, superseding the prior per-setting proposal. The user confirmed the recorded interpretation is clear and correct and authorized proceeding. Configuration edge cases remain open. No application code was changed.

User selected a missing-local-configuration error when the shared YAML exists but neither worktree nor base local YAML exists. Capture must not proceed in that case. This differs from a missing shared YAML, which remains a silent skip. Recorded the decision; the user confirmed it is clear and correct and authorized continuing. No implementation was changed.

User selected resolving supported relative paths against the selected local YAML file's directory: base-local paths remain based in the base checkout, worktree-local paths in that worktree. Recorded this decision; the user confirmed it is clear and correct and authorized continuing. No runtime behavior or application code was changed.

User selected a clear configuration error for invalid selected configuration. Stop capture, identify the file and relevant setting/expected format, preserve configuration and do not substitute another local file or edit automatically. Omitted optional settings are not inherently invalid. Recorded the decision; the user confirmed it is clear and correct and authorized continuing. No application behavior was changed.

Configuration-topic checkpoint: proposed the direct consequence that base-local availability is required only when falling back to that file. A valid worktree-local file stands on its own; inability to obtain the needed base local file is a configuration error. The user confirmed the combined rules are complete and clear and authorized moving to automatic worktree adoption.

User selected explicit shared enablement for automatic capture: valid configuration plus an affirmative enable setting in the active checkout's shared YAML is required; no per-worktree installation manifest or enrollment command is required. The setting name and local disable behavior remain undecided. Eligibility does not infer ticket identity. Recorded this decision; the user confirmed it is clear and correct and authorized continuing. No application code was changed.

User selected shared-only enablement: local YAML cannot enable or disable capture, including no local disable-only override. Machine-local source/path settings remain separate from project eligibility. Recorded this decision; clarity/progression confirmation is pending. No application behavior was changed.

Steering correction: the user changed the prior shared-only enablement choice to allow a local disable-only override. Shared YAML must explicitly enable capture; the selected local YAML may veto it but cannot force-enable capture. With no local disable, shared enablement applies. Only the selected whole local file contributes this override; missing worktree-local settings do not inherit from the base. This supersedes the preceding shared-only decision. Revised clarity/progression confirmation is pending; no runtime implementation changed.

Open ownership discussion: before confirming the local-veto rule, the user proposed making capture configuration entirely local, noting that machine-installed capture and manual CLI use should follow the same repository rules. Evaluating local capture-runtime ownership versus shared project/metrics policy; no selection or configuration migration has been recorded yet.

User selected local ownership of automatic capture and machine-specific capture settings, with shared project/metrics policy common to manual CLI and plugin execution. This supersedes shared capture enablement and the shared-enable/local-veto model. Disabling automatic capture locally does not disable manual generation. Whole-file local selection, required active shared YAML and existing missing/invalid-file rules remain unchanged; binary-version policy is separate. Updated CAPTURE_VISION_AND_GAPS.md to reflect the current decision rather than conflicting earlier alternatives. The user confirmed this revised boundary is clear and correct and authorized continuing to the local enable default; no runtime behavior was changed.

User selected a configuration error when the selected local YAML omits the capture-enable setting: no implicit true/false default. Installation must generate local configuration, ask whether to enable capture and recommend enabling it, then persist the explicit user choice. Existing-file handling and noninteractive installation are still open. Recorded this decision; the user confirmed it is clear and correct and authorized continuing. No installer or runtime code was changed.

User selected preserving valid existing local settings during installation, including explicit capture enablement and unrelated/custom source settings. Ask only for missing settings; if enablement is absent, recommend enabling and save the explicit choice. Invalid existing configuration errors instead of being silently replaced. Installation write location and noninteractive behavior remain open. Recorded the decision; the user confirmed it is clear and correct and authorized continuing. No installer/runtime code was changed.

User selected a strict setup write boundary: setup running in a worktree may modify only that worktree's configuration; changes to base-local settings must be made from the base checkout. Existing worktree-local configurations still win under whole-file selection, so base changes affect only worktrees that select the base local file. New-file initialization and noninteractive setup remain open. Recorded the decision; the user confirmed it is clear and correct and authorized continuing. No setup/runtime implementation changed.

User rejected guided worktree-local configuration setup and directed leaving it to the user. Worktree-local files must not be automatically generated, copied from base or guided through a setup flow; the user manages them independently. Absence continues to use base-local fallback, while a user-created worktree-local file is selected in full and must satisfy required settings. This supersedes the pending copy-versus-fresh initialization question. The clarification is specific to worktree-local setup; base initial-setup discussion remains unchanged, and the user confirmed the revised boundary is clear and correct and authorized continuing. No implementation changed.

Clarification prompted by the user's missing-binary question: a missing plugin entrypoint cannot inspect capture enablement; startup fails at the harness. Revised the pending disabled-state question to concern source access and capture-only dependencies after successful startup/configuration validation. Do not assume an extra bootstrap executable or an optional selected-version binary before the separate binary-architecture decision. No disabled-state policy was selected yet.

User selected skipping capture-only readiness checks while disabled, after successful entrypoint startup and configuration validation. Disabled callbacks do not inspect source access or capture-only dependencies, collect sessions or start capture workers. Invalid/missing required configuration still errors; a missing entrypoint remains a harness launch failure. Recorded the corrected policy; the user confirmed it is clear and correct and authorized continuing. No runtime implementation changed.

User selected first-real-event initialization for a valid, explicitly enabled worktree. Configuration discovery and status inspection remain read-only; no proactive sibling initialization or per-worktree enrollment is required. Real event evidence supplies session identity, not initialization itself. Recorded the decision; the user confirmed it is clear and correct and authorized continuing to native callback routing. No implementation changed.

User selected routing by validated harness session/event context. A differing callback-process cwd must not select a different project; authoritative source conflicts still require explicit handling. Clarified that hook-file location does not determine callback cwd and that normal Codex callbacks run from the session directory. Recorded the decision; the user confirmed it is clear and correct and authorized continuing. No runtime implementation changed.

User requested simpler source-delay handling and asked whether session ID is available. Official Codex hook documentation confirms common session_id and that subagent hooks use the parent session ID. Proposed reusing existing session registration/worker refresh, treating callbacks as refresh signals rather than persisting a separate event replay queue. Proposal not yet selected; source availability does not prove metrics or child identity. No runtime implementation changed.

User selected the simplified principle: hooks notify; the worker collects from the authoritative source. Callbacks register/update identified sessions and signal a collection pass; use the existing session registry/worker instead of a separate event replay queue. Missing source data is awaiting collection, not verified zero; existing facts are preserved. Timing and retry limits remain undecided. Recorded the decision; the user confirmed it is clear and correct and authorized continuing to binary provisioning. No implementation changed.

User selected a versioned user-level binary cache shared by projects/worktrees and Codex/OpenCode integrations. Different release and honestly identified development builds may coexist outside repositories; no per-worktree binary copy is required. Version selection, downloads, entrypoint architecture and cleanup remain separate decisions. Recorded the storage choice; the user confirmed it is clear and correct and authorized continuing. No binary installation or application implementation changed.

User selected shared-YAML-only version pinning: the active checkout's shared configuration selects the exact release/development build for both CLI and automatic capture, with no local override or base-shared fallback. This is shared engine compatibility policy, separate from local capture control. Different development-build identities require matching active shared pins but do not require GitHub publication. Storage remains a versioned user-level cache; downloads/provisioning remain open. Recorded the decision; the user confirmed it is clear and correct and authorized continuing. No configuration or runtime implementation was changed.

User selected reporting an exact missing shared-pinned version/build and requiring explicit installation. Normal CLI execution and callbacks must not download, substitute a cached version or change the pin. Published artifacts can be obtained by explicit installation; unpublished development builds are supplied locally by the build/install process. Recorded the decision; the user confirmed it is clear and correct and authorized continuing. No implementation changed.

Open user proposal: let shared configuration optionally point to locally generated development binaries, avoiding a shared-pin edit for every build. Discussing an explicit binary-directory/location field versus overloading version with a path. This is not selected or implemented; it may introduce a development exception to exact-version-only selection while preserving shared authority and normal release caching. Development pin-synchronization choice is on hold.

User selected a separate shared binary-location field (such as binary_dir) for development output rather than overloading version with a path. This replaces repeated development pin synchronization with an explicit development-location mode; release caching and shared-only selection authority remain. Exact field spelling, both-fields behavior and worker executable stability remain open. Recorded the decision; the user confirmed it is clear and correct and authorized continuing. No native configuration parser or runtime was changed.

User proposed a clean-worktree handoff: use the exact shared-pinned cached release while the local development binary is absent, then prefer the local build once available. Explained that this differs from the previously offered development-precedence/no-fallback option and would revise that absence-as-error policy. Proposal remains open; no arbitrary-version fallback or runtime download is implied. Running processes require a safe handoff boundary rather than changing executable because a file appeared. No implementation changed.

2026-11-08 decision: user selected automatic local-build preference with exact pinned-cache fallback. With both shared selectors, absent development output uses the exact cached pin; a completed valid local build is preferred when available and reports its actual identity without requiring a pin edit for each build. If local output is absent and the needed cached version is missing, explicit installation is required. No runtime downloads, arbitrary-version fallback or pin mutation. Consolidated binary notes to remove superseded exact-pin-for-every-development-build and missing-local-output error claims. Present-but-invalid output and live worker handoff details remain open. The user confirmed the recorded handoff policy is clear and correct and authorized continuing; no implementation changed.

User selected a clear error without execution or cached fallback when present local development output fails executable/platform/required-interface validation. This is distinct from absent output, which still uses the exact pinned cache. A valid build's different identity is expected, not a validation failure. Recorded the decision; the user confirmed it is clear and correct and authorized continuing. No runtime implementation changed.

User selected reporting a reliably established mismatch between configured and actual harness source locations and stopping capture until configuration is corrected. Report both locations; do not override settings, edit files or read an unrelated store. A guessed default is not proof of disagreement; temporary metadata unavailability in the correct store remains the separate background-collection case. Recorded the decision; the user confirmed it is clear and correct and authorized continuing. No runtime implementation changed.

User selected answering an exact session-specific binding request from explicit current-task ticket context without redundant user confirmation. Missing/ambiguous ticket context still requires asking, with the session unassigned meanwhile; background/dependency mentions do not establish current-task ownership. Source-proven parent/child behavior remains distinct. Recorded the decision; the user confirmed it is clear and correct and authorized continuing. No runtime implementation changed.

User selected collecting authoritative session facts locally while waiting for ticket binding. Keep the session visibly unassigned; do not create placeholder tickets or publish ticket ledger/summary until binding succeeds. Previously collected facts can be used after binding. Unresolved-session retention remains open. Recorded the decision; the user confirmed it is clear and correct and authorized continuing. No collection/publication implementation changed.

User selected automatic ticket inheritance for a source-proven child in the same repository/task when its parent is bound and the child has no explicit binding. Preserve the child's native identity, metrics and worktree scope; do not overwrite an explicit child binding or infer parentage from roles/paths/timing. Unbound/unproven ancestry remains unresolved. Recorded the decision; the user confirmed it is clear and correct and authorized continuing to capture coverage. No runtime implementation changed.

Scope clarification after the user's concern about development testing in production: withdrew the unselected preactivation coverage-verification gate. Root/child integration verification belongs in development tests, not per-worktree certification or production test callbacks. Production keeps only the already agreed configuration checks, actual identity/source validation, missing-data diagnostics and collection retries. No production code was changed by this design discussion. The user confirmed the scope clarification is clear and correct and authorized continuing to publication/lifecycle.

User selected background publication during activity and at turn/session checkpoints for bound sessions with available source facts. Periodic and checkpoint signals may be coalesced; avoid a write for every event and keep timing local. In-progress metrics remain distinguishable from finalized totals, unassigned sessions stay unpublished, and generation invokes no model. Recorded the decision; the user confirmed it is clear and correct and authorized continuing. No worker/publication implementation changed.

User selected bounded retries with backoff for temporary collection/publication failures, using the existing worker. Preserve known facts, published metrics and unresolved work after budget exhaustion; show the failure and stop that retry cycle. Later regular updates, notifications or explicit retries can resume it. No additional service/event queue or empty-result overwrite. Exact local timing/budgets remain open. Recorded the decision; the user confirmed it is clear and correct and authorized continuing. No implementation changed.

User selected retaining each running worker's selected build until normal exit. New workers and CLI invocations use the preferred valid build, but a rebuild does not force restart. Preserve pending work and single-publisher scope; keep an active worker's collection/publication subprocesses on its original build despite mutable development output paths. Status must distinguish active-worker build from preferred selection. Recorded the decision; the user confirmed it is clear and correct and authorized continuing. No worker implementation changed.

User selected retaining all existing machine-local capture state for deleted worktrees, including completed/published and incomplete records, until explicit cleanup. Do not automatically delete inactive state, recreate a checkout or redirect publication. Preserve registrations, bindings, collected facts, pending work and diagnostics; this does not duplicate whole harness logs and is separate from binary-cache cleanup. Recorded the decision; clarity/progression confirmation is pending before final consistency review. No retention implementation changed.

The user confirmed full retention is clear and correct and authorized final consistency review. Consolidated CAPTURE_VISION_AND_GAPS.md into current decisions, removing stale guided-worktree-setup wording and superseded binary/enablement proposals while preserving their discussion history here. No plan, contract or runtime implementation was created. Review identified one remaining selection-policy detail: whether a shared version pin remains mandatory alongside a development binary location. Exact interfaces, timing defaults and persistence/executable mechanics are implementation details; any newly discovered user-visible policy change must return for discussion.

Final consistency decision: the user selected always requiring a valid exact shared version pin, with binary_dir only an optional development-output preference. No development-location-only mode without fallback is intended. Valid local output is preferred; absence uses the exact pinned cache, invalid local output errors without fallback, and execution still does not download or mutate pins. Recorded the final policy choice; the user confirmed it is clear and correct and explicitly closed the decision review. Discussion notes remain non-implementation artifacts; no plan, contract or code was created.

Decision review closed by explicit user confirmation. CAPTURE_VISION_AND_GAPS.md now consolidates the agreed behavior. Engineering details remain identified separately; no implementation, plan or contract is implied by closing this discussion. The user subsequently requested committing and pushing the confirmed notes to draft PR #124; delivery evidence follows.
