h3. Title: AL-79: Deliver native user-level Codex capture across worktrees

*TARGET*: AgenticLens (Domain / Configuration / Git / Codex / Native CLI / Worker / Setup Distribution / Tests)

*Status*: Ready
*Work type*: feat
*Priority*: P1
*Blocked by*: None
*Blocks*: AL-80
*Origin*: PO-20, reclassified from GEN-22 after the user closed the capture design review.
*Specification*: [PO-20 SPEC.md](../../PO-20/SPEC.md), F1-F6; [confirmed discussion](../../PO-20/CAPTURE_VISION_AND_GAPS.md).
*Related*: Coordinate shared modules with AL-70 and installer boundaries with AL-73. AL-78's open-turn feedback and AL-63's broader capability audit remain independent; do not absorb unrelated acceptance.

*CONTEXT*

The inspected Codex runtime reads linked-worktree project hooks from the original checkout. Untracked worktree-only hooks are therefore insufficient. The local home-hook workaround uses Python and per-worktree manifests, neither of which meets the user's machine-wide native adoption model. Existing source extraction, binding and worker behavior are foundations to preserve and reuse, not a mandate to rewrite the platform.

*GOAL*

Install once in Codex user home and capture configured base/worktree sessions through the native runtime, with correct file-based configuration fallback, explicit binding, version/development selection and autonomous publication. Deliver a usable Codex slice and the common runtime consumed by AL-80; no OpenCode API implementation is required here.

*VERTICAL SLICE SCOPE*

* Domain Core: pure configuration/enablement and executable-selection decisions; exact session/parent binding and snapshot/publication semantics.
* Boundary Adapters: Git-aware whole-file local selection, source agreement, additive user hooks, native dispatcher, cache/development executable handling, source reads and existing worker/persistence integration.
* Presentation: CLI/setup/status/errors expose selected configuration and its origin, disabled/unassigned/failed publication, and active versus preferred build without introducing production verification gates.
* Quality: deterministic self-contained Go/config/Git/process tests, existing distribution checks and genuine native Codex root/child integration evidence. Preserve existing dashboard/API consumers; no new web UI is requested.

*ACCEPTANCE CRITERIA*

h4. Configuration and automatic adoption

* Enforce PO-20 F1/F2: require active shared YAML or silently skip; select one whole worktree-local file or base-local file; no key inheritance or base-shared fallback. Resolve local relative paths from the selected file. Missing/invalid required local configuration errors without mutation or fallback.
* Move capture runtime/enable/source/state settings to the selected local policy while keeping shared project/metrics/executable rules consistent between CLI and capture. Explicit local enablement is required; false preserves manual generation and skips capture-only I/O after configuration validation.
* A configured enabled worktree needs no install, enrollment or manifest. Initialize on its first actual event; discovery/status stay read-only. Test ordinary checkout, two linked worktrees, nested cwd, custom workspace/prefix/source roots and unavailable base fallback.
* Native routing follows validated session context even if callback cwd differs. Reliably established configured/runtime source mismatch reports both locations and stops capture; source delay is not a mismatch or authoritative empty data.

h4. Native Codex integration and setup

* Own additive hooks in the effective Codex user home. Preserve other hooks and trust; remove only unchanged owned legacy callbacks through explicit migration, with no duplicate invocation. Python and per-worktree manifest files cease to be requirements of the shipped path.
* Base setup preserves valid values, asks only for missing choices and recommends enabling capture. Worktree-local configuration remains user-managed: no guided creation/copying and no writes to base from a worktree. Do not guess required choices in noninteractive setup.
* Use hooks as exact-session refresh signals and the existing collector/worker, not an event-replay queue, another service/model or a per-worktree certification feature. Parent callback IDs must not be relabeled as children.

h4. Versioned cache and development output

* Shared version pin is always required and local YAML cannot override selection. CLI and capture use the same resolver. Provision exact published binaries only through explicit installation; no runtime download, latest fallback or pin changes.
* A separate optional shared development-location field prefers valid host-platform output. Absent local output uses the exact pinned cache; missing required cache errors. Present-but-invalid output errors without fallback. Report actual development identity; no per-build pin edit or publication is required.
* A running worker and its collection/publication children retain their selected build until normal exit, despite rebuilds at the same output path. New invocations/workers use the preferred build. Verify retained pending work, single-publisher scope and active/preferred status.

h4. Binding, publication and durable recovery

* Bind from explicit current-task context without redundant confirmation. Missing/ambiguous context stays unassigned and requests information. Collect locally before binding, publish only after it; no placeholder tickets or inferred ownership.
* Source-proven same-task children inherit a bound parent unless explicitly bound already; keep exact child identity/worktree scope. Independent roots bind independently.
* Publish during activity and turn/session checkpoints with coalescing and truthful provisional/unknown values. Temporary failures back off within a bounded local budget; preserve facts and unresolved work on exhaustion and resume on later triggers. Never fabricate empty success or erase unavailable owners.
* Retain all complete/incomplete state after worktree deletion until explicit cleanup. Do not recreate destinations or publish into another checkout. Preserve unrelated contributors and serialize publication; repeated snapshots do not accumulate or duplicate.

h4. Evidence and portable history

* Prove the pre-fix native discovery case through sourcePath and the resulting global-hook path with normal trust, actual root binding, a source-proven child and source-backed publication. Record actual native version/interface and source IDs. Do not treat fixtures or installed files as runtime proof.
* Exercise PO-20's cross-machine/other-harness ledger preservation, unavailable enrichment and complete-versus-partial snapshot cases. Capture state/installation receipts are not prerequisites for reading supported Git-shared history. No schema expansion or bulk migration is assumed.
* Keep fixtures independent of PO/ticket workspace paths and isolate every writable HOME/config/cache/state destination. Retain meaningful existing assertions and update canonical guidance/shipped copies without general audit cleanup.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with approved developer contract; no implementation is authorized by this PO file alone.
* All Codex/common-runtime acceptance is delivered end to end; AL-80 can reuse the shared behavior without another policy implementation. Do not call the combined two-harness initiative complete before AL-80.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md: FCIS, strict typing, zero any/interface{}, zero emojis, zero domain mocks. Pass affected checks, focused independent review and required make. Rebuild/verify embedded assets if affected; preserve existing UI consumers.
* Publish revision-specific validation, source-grounded telemetry and implementation PR. Runtime evidence gaps are explicitly incomplete, not silent skips. Coordinate related tickets instead of duplicating their unrelated work.
