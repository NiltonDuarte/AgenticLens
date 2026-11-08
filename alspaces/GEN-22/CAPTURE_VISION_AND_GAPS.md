# AgenticLens capture: vision and decisions

GEN-22 discussion notes. Started 2026-11-07; consolidated 2026-11-08. The user confirmed the final consistency rule and closed the decision review. This document is not an implementation plan or a frozen contract. Confirmed decisions describe the intended behavior, not functionality already implemented. Superseded choices remain traceable in the session discussion and SUMMARY.md.

## Vision

Install the integration once in the developer's Codex or OpenCode user directory. It must operate from configured base checkouts and linked worktrees, including project subdirectories, without copying untracked hooks or requiring per-worktree enrollment.

Capture interacts with agents only as needed for explicit ticket information. Hooks notify; the background worker reads authoritative session sources and generates metrics automatically. Manual CLI and automatic capture use the same project and metrics rules.

## 1. Configuration selection

Confirmed by the user:

- The active checkout must contain `.agenticlens-config.yaml`. If absent, automatic capture silently skips that checkout. Never use the base checkout's shared YAML as a fallback.
- Select one whole `.agenticlens-config.local.yaml`: use the worktree-local file when present, otherwise the base-local file. Do not inherit individual missing settings from the other local file.
- If shared YAML exists but neither local file is available, report a missing-local-configuration error and do not capture.
- Invalid selected configuration produces a clear file/setting diagnostic. Preserve the file; do not substitute another local file or repair it automatically. Omitted optional settings are not inherently invalid.
- Resolve supported relative local paths against the directory containing the selected local file. Absolute and `~/` paths retain their normal meanings.
- A valid worktree-local file does not require the base-local file to be accessible. If fallback is needed but the base-local file cannot be obtained, report the configuration error.
- Identify the base checkout through Git metadata when fallback is needed. Preserve the active worktree's project identity and publication destinations.

A user-created worktree-local file replaces the base-local file in full. For example, omitting its Codex source-home setting does not inherit that setting from the base local file.

## 2. Configuration ownership, enablement and setup

Confirmed by the user:

- Selected local YAML owns automatic capture and machine-specific operation: source locations, capture state, timing/limits and machine integrations.
- Shared YAML owns project identity, ticket conventions, workspace/output conventions, code classification, metrics-generation policy and executable selection. Common settings have the same meaning for manual CLI and capture.
- There is no shared capture-enable prerequisite or shared-enable/local-veto combination. These earlier proposals were superseded.
- The selected local file must explicitly enable or disable automatic capture. Omission is a configuration error; there is no implicit default.
- Disabling automatic capture does not prohibit manual generation.
- The entrypoint must start and configuration must validate before the disable decision. Disabled callbacks then stop without checking source access/capture-only dependencies, collecting sessions or starting capture workers. A missing entrypoint remains a harness launch failure.
- Initialize worktree capture state on its first real event, after valid configuration and explicit enablement. Configuration discovery and status inspection stay read-only. Do not proactively initialize siblings or invent session identities.
- No per-worktree installation or installation manifest is required by the target design.

### Base-checkout setup

Initial local configuration generation asks whether to enable capture, recommends enabling it and saves the explicit choice. Preserve valid existing settings and capture enablement; show retained choices and ask only for missing settings. Invalid existing configuration errors instead of being silently replaced.

### Worktree-local configuration

Worktree-local configuration is user-managed. Neither the setup skill nor a binary setup operation should automatically create/copy that local file or guide its creation. The user creates or edits it independently. If absent, runtime selects base-local configuration under section 1.

Setup run from a worktree must not modify base configuration. Changes to the base-local file must be made from the base checkout and affect only worktrees selecting that file. The rejected copy-versus-fresh initialization proposal is not part of this design.

## 3. Native callback routing and collection

Confirmed by the user:

- Select the project/worktree using validated harness session context. A different callback-process cwd must not redirect the event or automatically reject otherwise valid session context.
- Validate context and configuration boundaries. Conflicting authoritative session/source evidence is not silently accepted.
- The hook configuration file's location does not determine the callback's cwd. Normal Codex command hooks execute with the session working directory.
- Hooks identify/register/update sessions and notify. The worker collects authoritative source snapshots. Reuse session registration and worker scheduling instead of a separate event-payload replay queue.
- Identity includes harness, source store and native session ID. Codex subagent hook session_id identifies the parent; do not reuse it as the child's identity.
- Temporarily unavailable source data remains awaiting collection with visible diagnostics. Preserve known facts; do not claim a successful empty/zero snapshot.
- The shared user-level integration must not freeze an installing worktree's project, source or state paths into callbacks for other projects.

Native routing should replace the temporary Python launcher. The exact command/API is an implementation detail, not an approved new command name.

## 4. Binary provisioning and development output

Confirmed by the user:

- Use a versioned user-level cache shared by projects, worktrees, Codex and OpenCode for release/provisioned native binaries. Multiple builds/platforms can coexist.
- Shared YAML alone controls selection for both manual CLI and capture. Local YAML cannot override it; there is no base-shared fallback.
- Keep `version` as an exact identifier. Add a separate shared development-location field, such as the proposed `agenticlens.binary_dir: ./bin`; do not overload version with a path.
- Development output resolves relative to the active shared configuration and selects the executing platform's binary. This mode permits untracked build output in the development checkout; new worktrees need not contain it initially.
- With both selectors configured, prefer a completed valid local build. While it is absent, use the exact shared-pinned version already installed in the cache.
- A usable development build reports its actual identity; it need not match the fallback release identity, and no pin edit is required after every build.
- Present-but-invalid local output is an error: do not execute it or fall back. This includes executable/platform/required-interface validation failures.
- If local output is absent and the required cached version is missing, report the exact missing version/build and require explicit installation.
- Normal CLI execution and capture callbacks do not download, change the pin or choose an arbitrary cached/latest version.
- Explicit installation may retrieve an exact published precompiled artifact. The local build/install process supplies unpublished development output; no GitHub publication is required.
- A missing installed entrypoint cannot inspect configuration; the harness reports launch failure.

### Running workers

A running worker and its collection/publication subprocesses retain their selected build until normal exit. A newly preferred build does not force restart. New CLI invocations and new workers use the preferred valid build.

Preserve pending work and prevent competing publishers for the same scope. Status distinguishes the active worker build from the preferred selection. Mutable development output must not accidentally change the executable used by an existing worker's later subprocesses.

### Required fallback pin

**Confirmed user decision: always require shared `version`; the development location is an optional preference.** The active shared YAML must contain a valid exact version pin even when `binary_dir` is configured and local output is available. Do not support a development-location-only configuration with no defined fallback. The pinned artifact is used when local output is absent; a completed valid local build is preferred when present. Required configuration does not mean executing or downloading both choices on every callback.

The user confirmed this final selection rule is clear and correct and closed the decision review. Missing required cached output, present-but-invalid local output, explicit installation and worker build lifetime retain their previously agreed behavior.

Exact field spelling, platform filenames, stable-executable retention, cache layout and installer integrity mechanics remain implementation details to define consistently with these rules.

## 5. Source-home agreement

Confirmed by the user:

- If the integration reliably knows the actual harness source location differs from selected local configuration, report both locations and stop capture until configuration is corrected.
- Do not override or edit configuration, silently use another store or switch harnesses.
- A guessed default is not proof of a mismatch.
- Temporary unreadability in the correct store is separate and follows background collection/retry behavior.
- Keep harness configuration location, telemetry source location and AgenticLens state as distinct responsibilities.

## 6. Ticket binding and agent interaction

Confirmed by the user:

- Use an explicitly supplied current-task ticket to answer the exact session's binding request directly, without redundant user confirmation.
- Background examples, dependencies or incidental ticket mentions do not establish the active task's ticket.
- Ask when ticket context is missing or ambiguous. Leave the session unassigned; do not infer from branches, paths, roles or timing.
- Collect authoritative facts locally while unassigned, but do not create a placeholder ticket or publish a ticket ledger/SUMMARY.md until binding succeeds.
- A source-proven child in the same repository/task inherits a bound parent's ticket unless the child already has an explicit binding. Preserve its own native ID, metrics and worktree scope.
- An unbound parent or unproven ancestry does not establish inherited attribution. Independent roots require their own exact-session binding.

If an agent cannot answer a binding request, collection remains local and unassigned; a shutdown event alone cannot obtain missing information. Binding and background collection do not require pausing the agent's task or invoking another model.

## 7. Capture coverage and development verification

The user confirmed that development verification is not a production activation gate:

- Exercise real root/child capture in development and integration tests.
- Do not add user-facing certification, per-worktree verification enrollment or production test callbacks.
- Production follows configuration/enablement rules, validates actual source/session identities, reports real failures and retries as agreed.
- Do not claim unobserved sessions were captured or turn missing evidence into a successful zero.

Target every root and child session exposed by supported harness interfaces. Hook/plugin trust, lifecycle delivery and source availability still affect what can be collected. Installation/discovery alone does not prove actual callback delivery or publication.

## 8. Publication, recovery and retention

Confirmed by the user:

- Publish available bound-session facts during activity and at turn/session checkpoints. Coalesce notifications instead of writing for every event.
- Timing and limits belong to local capture configuration. In-progress values must not be represented as finalized session totals.
- Temporary source-read/publication failures use backoff within a bounded retry budget in the existing worker. Do not add another service or event queue.
- On budget exhaustion, retain known facts, published metrics and unresolved work; report failure and stop that retry cycle. Later normal updates, notifications or explicit retries may resume it.
- Configuration errors and authoritative source contradictions retain their stop/error behavior.
- After worktree deletion, retain all existing capture state, complete and incomplete, until explicit cleanup. This includes registrations, bindings, collected facts, pending work and diagnostics.
- Do not automatically delete even fully published/inactive state, recreate the checkout or redirect publication to the base or another worktree.
- Report an unavailable publication destination. Retention does not mean duplicating whole harness source logs; binary-cache cleanup is separate.

The user confirmed retention and authorized this consistency review. Concurrent publication must preserve known session facts and worktree ownership. No automatic Git commit or model invocation is implied by publication.

## Implementation details still to define

These are not additional selected policies or implementation work:

- Exact configuration field names and command/API names.
- Supported adapters' session-context extraction and Git base-checkout discovery.
- Stable executable retention when a development binary is rebuilt.
- Serialized publication, atomic writes and preservation of unrelated/persisted session history.
- Local timing defaults, retry budgets, normal worker-exit limits and duplicate-notification handling.
- Noninteractive base setup: preserving the required explicit enable choice without assuming a default.
- Exact installation/migration/removal behavior for existing project hooks and manifests, user-hook trust and the temporary Python launcher.
- Explicit cleanup/recovery commands and version-cache maintenance.

Changes to user-visible policy discovered while resolving these details must be brought back for discussion. No implementation plan or contract has been created.

## Current local setup and known implementation gaps

These are the last observed local facts, not a new live runtime verification:

- A local native build was installed and matched this worktree's shared version pin.
- Six user-level Codex hooks at `~/.codex/hooks.json` invoke `~/.codex/hooks/agenticlens-capture.py`, a local workaround outside the shipped product.
- That launcher requires a per-worktree installation manifest and does not implement automatic adoption or base-local fallback.
- Worktree-local installer files remain for current validation, excluded from Git. The original-checkout hook workaround was deleted.
- The last native discovery check found six untrusted home hooks. Live callback delivery, binding, extraction and automatic publication have not been verified here.
- Native routing, development-location selection and the configuration behavior above remain intended changes, not features implemented by these discussion notes.

## Evidence and references

During troubleshooting, native Codex 0.167.0 metadata probes showed that linked-worktree project hooks were read from the original checkout. Worktree-only hook files were invisible. Isolated probes compared sourcePath as well as counts; the exact six hook groups worked in an ordinary repository. User-level installation avoids relying on uncommitted project hook replication.

References: [Codex hooks](https://learn.chatgpt.com/docs/hooks), [OpenCode global plugins](https://docs.opencode.ai/docs/plugins/).
