# PO-20: Machine-wide automatic capture across Git worktrees

## Problem statement

A developer installed AgenticLens capture in a linked worktree, but native Codex did not discover those project hooks. Inspection of Codex 0.167.0 showed that linked-worktree project hooks were loaded from the original checkout. A temporary user-home Python launcher demonstrated discovery, but still required per-worktree installation manifests and lacked the requested configuration behavior.

The integration must be installed once per developer/harness and operate automatically from configured base checkouts and worktrees. Manual CLI generation and automatic capture must obey the same project and metrics rules. The closed user decision review is retained in [CAPTURE_VISION_AND_GAPS.md](CAPTURE_VISION_AND_GAPS.md).

This is the same initiative originally tracked as GEN-22, explicitly reclassified by the user as PO-20. No new discovery interview or implementation approval is implied. The prior discussion decisions remain authoritative; this specification organizes them for separate development sessions.

## Target personas and use cases

- Developers and agents working in disposable Git worktrees who need automatic ticket metrics without per-worktree hook installation.
- Developers using native Codex or OpenCode who expect the integration to ask only for information not already supplied.
- AgenticLens contributors who need capture before a local binary has been built and automatic preference for their completed development build afterward.

## User stories

1. As a developer, I want to install the integration in my harness user directory once, so that configured worktrees are captured without copying untracked hooks.
2. As a worktree user, I want a whole local-file override or base-local fallback, so that machine configuration is predictable without per-setting inheritance.
3. As an agent, I want to bind my exact session using the user's known ticket or request missing information, so that metrics have explicit ownership without repeated confirmation.
4. As an AgenticLens contributor, I want the pinned installed version to work before my first build and a valid local build to be preferred afterward, so that capture does not depend on publishing development releases.
5. As a metrics consumer, I want automatic updates that preserve unavailable history and identify incomplete reads, so that source delays or other developers' private data do not become false zeroes or lost records.

## Functional requirements

### F1. Shared configuration and whole-file local selection

- Resolve the active project from validated harness session context, including supported subdirectory invocations, without redirecting capture to a different callback-process cwd.
- Require the active checkout's `.agenticlens-config.yaml`; if absent, automatic capture silently skips that checkout. Never substitute base-checkout shared YAML.
- Select the entire worktree `.agenticlens-config.local.yaml` when present; otherwise select the base checkout local file identified through Git metadata. Never inherit individual missing settings from the other local file.
- Shared YAML present with neither local file available is a configuration error. A malformed/unreadable selected file, wrong value type or missing required setting produces an actionable diagnostic; do not repair it or select another file.
- Resolve allowed relative local paths against the selected file's directory. Keep absolute and home-relative semantics. A valid worktree-local file does not require the base local file; inaccessible required fallback is an error.
- Local YAML owns capture enablement, source locations, capture state, timing/limits and machine integrations. Shared YAML owns project/ticket/workspace/classification/metrics policy and binary selection. Preserve existing explicit CLI override semantics where applicable; automatic capture does not fabricate overrides to hide configuration disagreement.
- Require an explicit local capture-enable choice. Missing is an error; false disables automatic capture without disabling manual generation or making valid stored ledger reads depend on capture installation.
- Validate configuration before applying disablement. Once disabled, skip source access, capture-only readiness work, registration and worker launch. A missing executable entrypoint remains a harness launch error.

### F2. User-level native integration and adoption

- Install additive owned Codex hooks in its effective user home and an OpenCode plugin in its effective user config/plugin directory. Respect configured locations and preserve unrelated integrations and user trust.
- Replace project-bound callbacks and the temporary Python workaround with native routing plus only harness-required glue. No project paths, private source paths or state directories are frozen into callbacks that serve multiple projects.
- A valid, enabled worktree requires no separate capture install, enrollment or per-worktree installation manifest. Create its capture state on its first real event; discovery and status inspection stay read-only and do not initialize siblings.
- Hooks identify/register/update native sessions and notify; the existing worker collects authoritative source snapshots. Do not introduce an event-payload replay queue, another model invocation or a separately installed service just to handle source delays.
- Treat session identity as harness/source-store/native-session identity for local capture routing. A Codex subagent hook's parent ID is not the child ID. Preserve the canonical portable ledger's exact-session refresh rules without adding machine/checkout receipts as historical admission requirements.
- A reliably established configured-versus-runtime source-location mismatch reports both paths and stops capture until corrected. Guessed defaults do not establish a mismatch. Temporary unreadability in the correct source remains pending collection, not verified empty data.
- Installation/removal and legacy transition preserve pending state and unrelated hooks. Remove/disable only unchanged owned legacy callbacks through explicit migration, avoiding duplicate capture when user and project integrations coexist. Worktree manifest files must not remain a requirement of the new model.

### F3. Binary selection and local development

- Shared YAML always requires an exact `version`; local YAML cannot override executable selection. CLI and capture use the same selection rules.
- Use a versioned user-level native cache shared by projects and both harnesses. Do not depend on untracked per-worktree binaries for ordinary release use.
- Support a separate optional development binary-location field, such as `binary_dir`; do not overload `version` with a path. Resolve it from the active shared file and select the executing platform's native executable.
- Prefer completed valid development output and report its actual identity without relabeling it as the pinned release. No shared-pin edit is required after every build.
- When development output is absent, use the exact pinned cached version. If that required cache entry is missing, report the exact version and require explicit installation.
- Present-but-invalid development output produces an error without execution or fallback. A valid development build having a different identity from the fallback is expected.
- CLI/callback execution never downloads, changes pins or selects arbitrary/latest cache entries. Explicit installation may retrieve exact verified precompiled artifacts; the build/install system supplies unpublished development output without GitHub publication.
- A running worker and its collection/publication subprocesses retain the selected build until normal exit. New CLI invocations and new workers use the preferred build. Rebuilding a mutable output path must not silently change an active worker's later subprocesses. Status distinguishes active and preferred builds.

### F4. Setup boundaries

- Initial base-checkout local configuration generation asks whether to enable capture, recommends enabling it and saves the explicit choice.
- Preserve valid existing settings and enablement, show retained values and ask only for missing settings. Invalid configuration errors rather than being replaced.
- Worktree-local configuration is user-managed: no guided generation, base-value copying or automatic edits by setup skills/commands. Runtime may use base-local fallback without creating a local file.
- Worktree setup must not modify base configuration. Changes to base-local settings are made from the base checkout; existing worktree-local files still win.
- Do not invent an enable default for noninteractive use. When required choices cannot be obtained from existing valid configuration or explicit supplied input, report what is missing rather than guessing or mutating another checkout.

### F5. Explicit binding and all-session collection

- Use a clearly explicit current-task ticket to answer an exact-session request without redundant confirmation. Ask for missing/ambiguous required ticket information; do not infer from branches, paths, roles, timing or incidental mentions.
- Collect authoritative facts locally while unassigned. Publish no ticket ledger or summary and create no placeholder ticket until binding succeeds.
- A source-proven child in the same repository/task inherits a bound parent's ticket unless explicitly bound already. Preserve the child's native identity and actual worktree scope. Unproven ancestry and unbound parents do not establish inheritance; independent roots bind independently.
- A session that cannot answer remains visibly unassigned with collected facts retained. Shutdown alone cannot obtain missing information.
- Target every exposed supported root/child path. Verify them through development tests and genuine native evidence; do not add production certification/enrollment gates, synthetic production callbacks or claims that unobserved sessions were captured.

### F6. Autonomous publication, recovery and retention

- Publish available bound-session facts during activity and at turn/session checkpoints; coalesce notifications rather than writing per event. Distinguish in-progress values from finalized intervals.
- Temporary source/publication failures retry with backoff within a bounded local budget. On exhaustion preserve known facts, published history and unresolved work, report failure and stop that retry cycle. Later normal updates, notifications or explicit retries may resume it.
- Configuration errors and authoritative contradictions remain errors, not delayed-success cases. Missing/failed/partial source reads must not erase valid facts or become successful zeroes.
- Serialize publishers for the same scope and preserve exact-worktree ownership, multiple bound contributors and other developers'/harnesses' records. No duplicate sessions or accumulation of repeated snapshots.
- Retain all existing capture state after worktree deletion, complete and incomplete, until explicit cleanup. Do not recreate the checkout, redirect publication or delete even fully published/inactive state automatically. Report missing publication destinations.
- Automatic publication does not imply automatic Git commits. Binary-cache cleanup is a separate concern; do not introduce unrequested cleanup behavior.

### F7. Ad hoc manual generation without capture-state dependency

- Manual `generate` reads the selected harness source, project/metrics configuration and existing output ledger without reading, creating or modifying the AgenticLens capture registry. No capture installation, enablement or capture-only state configuration is required.
- Explicit manual session/ticket selection does not mutate automatic capture bindings. Stale/conflicting bindings, absent/corrupt registries and unavailable capture-state paths do not block manual generation.
- Preserve source/project validation, root/child extraction, output validation, atomic publication, shared writer serialization and the portable ledger contract. Unavailable/partial source facts are not fabricated or treated as zero.
- OpenCode's source SQLite database remains a legitimate telemetry input; Codex uses its source rollout files. Registry independence concerns AgenticLens's operational database only.
- Automatic worker publication still enforces enabled configuration, registry ownership and ticket binding. Separating manual generation must not provide an automatic-publication bypass.
- Deliver this correction independently in AL-81, coordinating shared publication boundaries with AL-79/AL-78 without artificial dependencies.

## Layer scope and presentation

- **Go domain:** pure eligibility/selection/binding/publication decisions and existing owner/snapshot semantics; no I/O, clocks or process execution in the functional core.
- **Go adapters and CLI:** Git/config/source resolution, user-level installation, native callback/worker composition, cache/development output selection and truthful status/error presentation.
- **TypeScript:** the OpenCode plugin is part of the second vertical slice. Keep it thin around the shared native behavior. No new web-dashboard feature is required; preserve existing ledger/API/dashboard consumers and field availability.
- **Documentation/setup distribution:** update canonical setup/operator/developer instructions and shipped equivalents together. Describe machine/user installation, manual worktree-local configuration, trust, explicit installation and migration limits accurately.

## Acceptance and compatibility constraints

Follow [TECHNICAL_GUIDELINES.md](../../TECHNICAL_GUIDELINES.md) and the [boundary regression checklist](../../docs/REGRESSION_PREVENTION.md). Go remains FCIS and value-oriented; TypeScript remains strict. Zero any/interface{}, zero emojis, no domain mocks, self-contained fixtures and proportionate tests.

Preserve the [portable ledger contract](../../TECHNICAL_GUIDELINES.md#portable-ledger-contract): a Git-shared valid ledger remains usable on another machine/checkout/harness without the producer's private capture database, receipts or logs. Complete successful local snapshots replace only their exact refreshed owners, including fewer/zero invocations; unavailable/partial owners and other-source history remain preserved. Source refresh does not require a new authenticity policy or bulk schema rewrite.

Demonstrate an ordinary checkout and two linked worktrees, nested invocation, conflicting base/worktree local values, whole-file omission behavior, relative-path origin, disabled capture with no source I/O, and no writes on discovery. Add two-session/two-ticket and concurrent same-ticket cases that reveal wrong-worktree writes and lost updates. Exercise missing versus invalid development output, missing pinned cache, active-worker stability across rebuild, failed publication, interruption and deleted destinations.

Native integration evidence must name the actual harness/version/interface and source session identities, show global discovery, normal human trust, root binding, a source-proven child, source collection and ticket publication. Fixture payloads and installed files are not proof of live operation. Evidence gaps remain incomplete rather than silently passing. No test-only activation requirement is shipped.

## Capture integration decomposition and dependencies

Two end-to-end tickets are the smallest practical batch that keeps each session bounded and each delivery independently usable. One ticket per layer would fragment ownership; one ticket for both harness integrations would combine distinct runtime APIs, trust/binding flows and real-source verification into an unnecessarily broad delivery.

1. [AL-79: Native user-level Codex capture with worktree configuration and development builds](../plan/to-do/AL-79%20capture%20codex%20sessions%20across%20worktrees.md) — Ready; no ticket prerequisite. Delivers the common engine/configuration/binary/lifecycle behavior and a real Codex path end to end.
2. [AL-80: Global OpenCode capture using the shared worktree runtime](../plan/to-do/AL-80%20capture%20opencode%20sessions%20across%20worktrees.md) — Blocked by AL-79. Adds the TypeScript global adapter, runtime binding and native OpenCode parity using the same core rules.

Related existing work remains separately owned: AL-70 configuration/path propagation corrections, AL-73 release/checksum reliability, AL-78 open-turn feedback effectiveness, and AL-63 source capability audit. They are coordination points, not artificial prerequisites. Revalidate their current state and avoid duplicating their unrelated scope.

## Additional manual-generation correction

[AL-81: Make manual generate independent of the capture registry](../plan/to-do/AL-81%20generate%20metrics%20without%20capture%20registry%20access.md) is Ready with no ticket prerequisite. This separately requested correction addresses F7; the two-ticket capture integration sequence remains AL-79 then AL-80. The user requested ad hoc generation independent of the AgenticLens registry, while retaining harness source access and automatic capture safeguards.

## Handoff limits

This PO session authors requirements and tickets only. No developer workspace, CONTRACT.md, application/test/runtime implementation or production migration is created. Development starts in separate agentic-lens-dev sessions with their own approved contracts. Numerical timing defaults, exact API/field names and persistence mechanics are engineering details; any new user-visible policy must return for discussion.
