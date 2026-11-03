---
name: agentic-lens-setup
description: Autonomous setup skill to bootstrap AgenticLens in OpenCode and Codex repositories, downloading precompiled release binaries, configuring project parameters, explaining telemetry features, and reviewing agent guidelines with mandatory user confirmation.
---

# AgenticLens Cross-Project Autonomous Setup Skill

Guides autonomous bootstrapping and adoption of AgenticLens in Codex and OpenCode software repositories using precompiled release artifacts from GitHub Releases (https://github.com/NiltonDuarte/AgenticLens/releases).

AgenticLens is an observability and developer telemetry platform for OpenCode and Codex workflows. OpenCode traces come from its local SQLite database; Codex traces come from rollout files under the Codex home directory.

This skill follows the autonomous agentic distribution model: it is completely self-contained and executable by any autonomous coding agent reading this manifest directly via shared repository URL or local path.

## Operational Mandates

- Install official precompiled releases only in adopting projects. No Go, Node, Python or yq installation is needed. OpenCode glue uses its existing harness runtime.
- Shared `agenticlens.version` is the sole release authority. Ordinary commands and every integration use the stable user-level common wrapper, which resolves the project and executing OS/architecture offline. Installation is explicit; never download from a callback or silently select latest.
- Present exact proposed files, destinations, source URLs and existing-file diffs before changing an adopting repository. Existing explicit approval for those changes satisfies the confirmation gate; do not ask again. Preserve locally edited files, unrelated PATH entries and Git policy. Never grant harness trust or change execution policy automatically.
- Agent instruction edits require the exact proposed text and affirmative user confirmation unless that exact scope is already approved. Do not modify third-party skills/personas. No staging, commits or pushes during setup.
- Zero emojis in generated text or code.

## Step 1: Install Complete Skills and the User Entrypoint

Select one exact published release before installation. Install both setup and operator skills, not just this manifest. The bundle includes `SKILL.md`, setup `scripts/install.sh`, `download.sh`, `download.ps1`, `agentic-lens-wrapper.sh`, `agentic-lens-wrapper.ps1`, an exact `RELEASE_VERSION` record, operator `SKILL.md`, `references/{cli,configuration,dashboard}.md` and the compatibility launcher. An installed release record identifies the skill bundle; it cannot override a project's YAML pin.

When scripts are absent, retrieve `scripts/install.sh` from the selected tag, inspect it and run it with `--version <EXACT_TAG> --dest <PROJECT>`. For local/offline bundle installation use `--source-dir <SOURCE_CHECKOUT>`. Existing differing bundle files cause a preflight failure before any bundle file changes. Native Windows users can retrieve the same manifest from that tag, preserving local edits, then use the native `download.ps1`; no Bash or policy bypass is required.

For an existing valid shared project pin:

```sh
bash .agents/skills/agentic-lens-setup/scripts/download.sh --project . --dry-run
bash .agents/skills/agentic-lens-setup/scripts/download.sh --project .
```

The default POSIX stable wrapper is `~/.local/bin/agentic-lens`; the native Windows stable entrypoint is `~/.agenticlens/agentic-lens.ps1`. Native artifacts coexist under `~/.agenticlens/bin/<version>/<os-arch>/<artifact>`. Local `agenticlens.install_root` changes the cache root independently of capture state. `--dest` (`-Dest` in PowerShell) changes only the stable wrapper directory. Preview the required per-user PATH entry. Persistent profile/user-PATH edits require explicit separate authorization; the installer never makes them. Hooks use the resolved absolute stable wrapper path so restricted environments do not depend on interactive PATH.

```powershell
powershell.exe -NoLogo -NoProfile -NonInteractive -File .\.agents\skills\agentic-lens-setup\scripts\download.ps1 -Project . -DryRun
powershell.exe -NoLogo -NoProfile -NonInteractive -File .\.agents\skills\agentic-lens-setup\scripts\download.ps1 -Project .
```

Installation stages exact-version artifacts, verifies SHA256SUMS and native version, checks hook protocol 2 when enabling capture or repairing a capture-aware leaf, then activates with rollback. A version/platform leaf is immutable; explicit `--force`/`-Force` repairs only that leaf. Other versions remain intact. Edited/unowned common wrappers are preserved. An interrupted installation lock requires inspection before retrying. Offline local installation requires both `--local-bin <NATIVE>` and explicit `--sha256 <EXPECTED_HASH>` (PowerShell `-LocalBin`/`-Sha256`). Never pass a shell wrapper as the native artifact. Foreign platform downloads can be staged with `--include-host OS-ARCH` or `--all-platforms`; this does not verify their runtime behavior.

For a new project without a pin, `download.sh --project . --init-config --version <EXACT_TAG>` (PowerShell `-InitConfig -Version`) explicitly creates a minimal pin/capture scaffold and installs the binary. This is binary readiness only. Complete the full project configuration in Step 2 before installing capture or claiming publication readiness. A legacy selected release can still run ordinary commands; if it lacks hook protocol 2/guarded init, setup reports that limit and requires a reviewed exact supported pin before adopting those new surfaces. No automatic upgrade occurs. Missing existing configuration in a worktree is a diagnosis task, not permission to reinitialize it.

## Step 2: Interview and Complete Project Configuration

Inspect existing shared/local files first. Ask only for unavailable project facts: human-readable project name, code source folders, ticket prefixes, workspace root/artifacts and the selected harness's actual local source store. Shared policy belongs in `.agenticlens-config.yaml`; machine-local source/state/cache paths belong in `.agenticlens-config.local.yaml` according to the owner's Git policy. Preserve existing settings; never run unconditional `init --force` to change one key.

Use the selected native `init` through the common wrapper with explicit `--agenticlens-version <EXACT_TAG>`, `--project-name`, `--src-folders`, `--ticket-prefixes` and configured `--workspaces` values. Review `--dry-run` output. When a bootstrap scaffold is present, use `init --complete-scaffold --agenticlens-version <EXACT_TAG>` with the full explicit project flags, as printed by a supporting installer; an arbitrary existing shared/local file must remain untouched. `init` generates the full schema, while the pin-only scaffold does not define ticket discovery or publication destinations. Distinguish this incomplete state explicitly.

Read [configuration and bootstrap syntax](../agentic-lens-operator/references/configuration.md#shared-release-authority-and-user-entrypoint). Exactly one plain root `agenticlens:` block with two-space children is supported. Literal plain/single/double-quoted exact release pins, whitespace/comments and LF/CRLF/BOM are supported. Aliases, merges, tags, quoted keys, flow pin mappings, block scalars and duplicate definitions fail with guidance; do not normalize them silently. Local version keys are rejected; environment/version flags cannot change shared selection. Legacy `.yml` uses the same profile; a nearer JSON configuration requires explicit migration to shared YAML rather than parent fallback.

## Step 2.5: Additive Lifecycle Setup and Agent Binding

Binary installation, complete configuration and runtime capture verification are separate outcomes. Use the exact source store and state path after inspecting harness configuration. Preview the common-wrapper command:

```sh
agentic-lens --al-offline --al-project <PROJECT> capture install --repo <PROJECT> --source <codex-or-opencode> --store <ABSOLUTE_SOURCE_STORE> --dry-run
```

Repeat without `--dry-run` when the reviewed changes are already authorized. `--wrapper <ABSOLUTE_STABLE_ENTRYPOINT>` identifies a nondefault user wrapper; `--binary` must identify the exact shared-pin native cache leaf. Never freeze that native leaf in installed callbacks. Codex and OpenCode glue enter the same common wrapper offline; selected native workers retain their executable for a pass. Installation records canonical source store, selected version, wrapper/template provenance and deterministic ownership. A project pin/template switch or incompatible/unowned wrapper change needs an explicit owned setup refresh and a new live round trip before health can be claimed.

Review Codex `.codex/hooks.json` or OpenCode `.opencode/plugins/agenticlens-capture.ts`, ownership/discovery manifests, local state and any exact state ignore entry. Preserve other integrations and edited owned files; conflicts require explicit resolution. Human trust/plugin enablement remains outside installer authority.

Verify the actual harness version/interface and normal trust, real callback registration, exact agent session, binding request/explicit response and source extraction. `capture status`, `capture diagnostics` and `capture verify` distinguish installed files from observed verification. A disposable self-test, version output or installed files cannot prove callback delivery. Unknown/unobserved children, unavailable source and platform gaps remain visible. AgenticLens does not invent ticket ownership from checkout, branch, role, timing or tool task IDs.

Codex sends a request token and exact wrapper-based `capture bind` command. Supply only the user-provided ticket/title and optional conventional type. OpenCode exposes `agenticlens_bind`, which gets the source session ID from the runtime. After a material setup refresh, reaffirm explicit binding when requested; an old binding remains preserved but does not establish the new installation's verified round trip. Independent roots need their own explicit bindings; source-proven parents may supply inherited binding.

### Optional commit warning

Only explicit `capture install-check --repo <PROJECT> --wrapper <ENTRYPOINT> --dry-run` / installation enables the advisory pre-commit integration; `uninstall-check` removes unchanged owned content. It honors effective `core.hooksPath`, derives the actual committing worktree at runtime and never changes CI/global Git settings. Existing arbitrary hooks are preserved and installation refuses unsafe automatic composition; give manual-composition guidance. AgenticLens diagnostics do not block commits. Local `telemetry.capture.commit_warning: false` disables its output. A pending marker or ordinary active tail alone is not a missing-metrics warning.

## Step 3: Explain Operations and Recovery

The embedded dashboard needs no Node runtime: `agentic-lens --al-project <PROJECT> server --port 8085`. Use [CLI reference](../agentic-lens-operator/references/cli.md) for scanning, generation, migration and backfill. Agents only answer binding requests; bounded native workers publish ledger and SUMMARY.md automatically. Manual `generate` is recovery, not mandatory work for every contributor. No capture_report.json is produced; deprecated `detailed_report` is accepted with a visible no-effect warning. Existing sidecars remain obsolete until the owner explicitly removes them.

Inspect durable local database diagnostics after source/publication errors. Available accepted facts survive failed refreshes; refreshed and retained owners remain distinguishable. Read-only status/check commands do not create registries, scan sources or clear pending work. Deleting local capture state can lose bindings and unprocessed observations; moving binaries into the home cache does not move capture data there or change retention.

For upgrades, explicitly change the shared exact pin as a reviewed project configuration change, then install that release and preview owned capture refresh using the original source/state context. Existing workers may finish on the native version they already selected. Do not overwrite a foreign pinned leaf or automatically change trust. Legacy repository-local wrappers/binaries are preserved; migrate owned callbacks to the stable entrypoint instead of using their updater behavior. Runtime `update` no longer installs releases.

Native Windows local PowerShell 5.1 entrypoint/PE dispatch and installation have real execution evidence. UNC script execution was rejected by Windows AuthorizationManager in the observed WSL path; do not advertise it as supported or bypass policy. Generic Windows console cancellation and untested architectures/harness paths remain unverified; check current release evidence.

## Step 4: Agent Guidelines Review and Confirmation

Present the exact proposed adopting-repository block before writing it, unless that exact text/scope was already approved. Keep it small:

```markdown
## AgenticLens session binding and publication

- Respond to the hook's session-specific binding request with the explicitly supplied ticket/title and request token, using its common-wrapper command or runtime binding tool. If ownership is unknown, ask; never infer it from checkout, branch, role or timing.
- Verify automatic ledger/SUMMARY.md publication and its coverage limits. Use manual generate only for recovery. Local diagnostics explain failures; keep raw transcripts and capture state out of Git.
```

Complete handoff states the stable wrapper/PATH requirement, exact shared pin/platform/native, complete configuration status, actual runtime verification evidence, optional Git installation state and remaining limitations. No capture-health claim follows from binary availability alone.

### Windows lifecycle transport budgets

The native registration/request/binding handoff retains its two-second default. Native Windows Codex hook declarations allow five seconds for PowerShell startup plus that handoff; POSIX declarations retain two seconds. The OpenCode Windows transport owns a five-second deadline, limits each output stream to 64 KiB, closes owned pipes on failure and requests termination of only its spawned Windows process tree through SystemRoot taskkill. Cleanup has its own one-second bound and reports failure explicitly. POSIX OpenCode transport retains its existing 1.5-second bound. Process-transport fixtures establish these measured boundaries; actual Windows Codex/OpenCode callback delivery and generic console cancellation still require separate harness evidence. Existing OpenCode supplies its runtime; no adopting-project Node installation is introduced.

### Explicit installation consistency and completion scope

Validated byte-identical native cache leaves retain their filesystem identity, including force-identical installation and in-place permission repair, so another project's admitted worker is not retired. A changed-content repair remains explicit. Pre-apply validation/activation failures roll back owned changes. Once external capture apply may have committed, an incomplete result, output-limit or timeout failure retains validated native/wrapper files and recoverable prior backups; it reports incomplete setup and directs status/owned reinstall recovery rather than claiming complete rollback or removing handlers referenced by hooks. Resolve those diagnostics before declaring setup healthy.

Printed scaffold completion runs inside the selected project and restores the caller's location. The common wrapper's --al-project selects release authority; it does not change ordinary native working-directory semantics. Keep the printed scoped command when supplying project facts, so shared and local configuration writes cannot target another same-version scaffold in the caller's directory.
