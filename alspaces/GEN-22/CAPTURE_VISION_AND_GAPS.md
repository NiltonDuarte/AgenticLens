# AgenticLens capture: vision and gaps

Discussion notes for GEN-22, recorded on 2026-11-07. These notes are not an implementation plan or a frozen contract. Proposed choices remain open for discussion.

## User vision

- Capture works from both the base checkout and linked Git worktrees, including agents launched from project subdirectories.
- Each checkout has its shared AgenticLens configuration. Resolve machine-local configuration from the current worktree when present, falling back to the base checkout as needed.
- Machine-local configuration supplies source locations such as the Codex source home.
- Install the harness integration once in the developer's Codex or OpenCode user directory. Do not depend on uncommitted project hook files being replicated into new worktrees.
- Installation provides the required binaries locally or retrieves them from the configured Git distribution source.
- Platform development supports locally built binaries without requiring publication to GitHub. The agent/build system places those binaries in the plugin installation location.
- Once installed, capture operates seamlessly in configured base checkouts and worktrees.
- Capture observes agent sessions, interacts with agents to obtain missing information such as ticket identity, and generates metrics automatically.

## Feasibility and discovery evidence

This model is feasible. Codex supports user-level hooks at `~/.codex/hooks.json`; OpenCode supports global plugins, normally under `~/.config/opencode/plugins/`. Actual locations must respect the harness's configured home/config directory and supported runtime version.

During troubleshooting, Codex 0.167.0 loaded project hooks from the original checkout when the session directory was a linked worktree. Hooks installed only in the worktree were not discovered. Controlled probes compared the reported hook source paths, not just hook counts. The same six hook definitions loaded successfully in an ordinary repository.

Project-level hooks are possible when committed or provisioned separately, but user-level installation fits the intended machine-local model and removes dependence on worktree file replication.

References: [Codex hook locations and trust](https://learn.chatgpt.com/docs/hooks), [OpenCode global plugins](https://docs.opencode.ai/docs/plugins/).

## Gaps and decisions to resolve

### 1. Configuration precedence and base-local fallback

The current implementation loads local configuration beside the selected shared configuration; it does not implement the proposed base-checkout local fallback.

The user described selecting worktree-local configuration when present, otherwise base-local configuration. We must clarify whether fallback selects an entire file or applies separately to missing settings.

A proposed per-setting precedence is:

1. Current worktree local configuration.
2. Base checkout local configuration.
3. Current checkout shared configuration.
4. Supported defaults.

This precedence is a proposal, not an accepted decision. In particular, release selection currently belongs to shared configuration and cannot be overridden by local configuration.

Identify the base checkout through Git metadata. Keep the active worktree's project identity, shared policy and publication destinations distinct from inherited machine settings. Resolve relative paths against the configuration file that supplied them. Define behavior when the original checkout or its local configuration is unavailable.

### 2. Automatic adoption of configured worktrees

A new configured worktree should not require another `capture install` invocation under the stated vision.

The current implementation and temporary Python launcher depend on a per-worktree installation manifest. That conflicts with automatic adoption. Define what configuration makes a project eligible, how a project disables capture, and when local registration/state may be created. Unconfigured projects should produce no capture activity.

### 3. Native user-level callback routing

The CLI already resolves the current project and configured capture state, but `capture receive` requires an explicit source store. It does not provide the complete installation-aware global entrypoint needed here: discover eligibility, obtain effective source settings, skip inactive/unconfigured projects before writes, then process the real event.

That routing belongs in the native AgenticLens implementation. The home integration must not embed an installing checkout's project, binary, source or state paths into callbacks that run across projects. Maintain exact worktree ownership and validate source-session evidence before attributing activity.

### 4. Binary provisioning and development builds

Define the supported binary location and selection rules for distributed installations and platform development.

A versioned user/plugin cache could serve multiple worktrees, with the current checkout selecting the appropriate binary. A binary need not be copied into every worktree. If project-local binaries are required, their provisioning must account for untracked files not being copied into new worktrees.

The proposed separation is: installation or the build system provisions binaries; callbacks use available binaries without downloading during agent execution. Decide when downloads are allowed, their exact source/version/integrity requirements, and what happens when a binary is unavailable.

Local development builds must retain an honest build identity and work without a published release. Different worktrees may need different versions without overwriting binaries used by running workers.

### 5. Configured source home versus actual harness source

Base-local fallback can supply a Codex source home, but the running session may use a different `CODEX_HOME`. OpenCode has corresponding source/configuration overrides.

Detect and explain disagreements. Do not silently read another source store or switch harnesses. Keep the executing harness's configuration directory, its telemetry source location and AgenticLens state as distinct responsibilities.

### 6. Agent interaction and explicit ticket attribution

Automatic capture and metrics publication should not depend on agents manually running generation commands.

When ticket information is missing, deliver an exact session-specific binding request to the agent. Keep the session unassigned until an explicit response supplies the required ticket information. Do not infer tickets from branches or directory names.

Proven child sessions can inherit attribution through authoritative parent relationships. Independent roots need their own binding. Define behavior when the agent cannot receive or answer a request; a shutdown event alone cannot obtain missing information.

### 7. Meaning of capture for every agent

Target every root and child session exposed by supported harness interfaces. An unconditional guarantee also depends on hook/plugin loading, human trust, lifecycle delivery and source-record availability.

Distinguish supported versions/interfaces from observed runtime coverage. Codex requires review/trust of non-managed hooks before execution. File installation and metadata discovery do not prove callback delivery, binding, extraction or publication.

### 8. Autonomous publication and lifecycle behavior

Handle delayed source writes, concurrent agents, interrupted sessions and worktree deletion without losing pending observations or mixing worktree ownership. Preserve recorded metrics when original sources are unavailable, while keeping coverage limitations visible.

Workers should perform collection/publication without recursively invoking a model. Removing a worktree must not silently delete pending evidence or recreate the checkout. Replacing an installation must preserve unrelated hooks and avoid duplicate callbacks from user and project locations.

## Current local setup

- The local AgenticLens binary is installed and its version matches the current worktree's shared pin.
- User-level Codex hooks are installed at `~/.codex/hooks.json` and are discoverable as six user hooks.
- They currently invoke `~/.codex/hooks/agenticlens-capture.py`, a local Python workaround created during troubleshooting, not part of the shipped native product.
- That launcher requires an active per-worktree installation manifest. It does not implement automatic adoption or base-local fallback.
- Installer-managed files remain in the current worktree because the current native receiver validates them. The original-checkout hook workaround was deleted.
- The last native discovery check showed the six home hooks as untrusted. Live callback delivery, explicit binding, extraction and automatic publication have not been verified.

The current setup demonstrates user-level discovery. It does not yet implement the full vision above.
