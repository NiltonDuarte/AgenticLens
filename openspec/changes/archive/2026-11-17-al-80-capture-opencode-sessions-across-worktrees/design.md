# AL-80 technical design

## Scope and approval

The user approved the AL-80 opening summary with “$agentic-lens-dev approved”. The [accepted capability](specs/global-opencode-capture/spec.md) covers a global OpenCode plugin, exact-session binding, publication and native verification. The initial checkout was clean at `b6ab0471`.

## Runtime and ownership

The coder owns application and operator documentation changes. Extend the user installer in `backend/internal/adapters/capture` and the engine's `capture install-user` surface to install and remove the plugin additively. Preserve existing Codex defaults. Inspect the actual effective OpenCode plugin loader, paths and legacy ownership evidence before choosing installation locations or migration actions.

Replace the project-bound behavior of `plugin.ts` with thin runtime notifications. Verify directory and exact ID on each entry point: lifecycle event, system transform request and binding tool execution. Where an event only carries an ID, resolve its native session context through the supported interface. Cross-check source project evidence and known runtime database location. Keep unavailable source data pending and surface contradictions. The binding tool accepts ticket fields and the native request challenge. Runtime context supplies session identity.

Extend the manager in `backend/internal/adapters/manager` and the engine composition alongside `capture_user_hook.go` to route the actual worktree. Share AL-79's admission, binary resolution and user-level registration with existing receive/request/bind and worker paths. Keep collection in the native adapter, using OpenCode's schema-family dispatch, exact-owner snapshot and source parent proof. Preserve installation-independent manual generation, per-project state and shared writer serialization.

Binary selection follows the later [AL-70 contract](../../../../alspaces/AL-70/CONTRACT.md). With a configured development directory, missing or invalid host output is an error. Omitting the directory selects the exact installed pin. OpenCode must exercise the same resolver as Codex and the manual CLI.

## Iteration 01 verification

The tester owns tests and verification evidence. Exercise the normal user install, strict TypeScript plugin, isolated native OpenCode server, real callbacks and publication into disposable configured repositories. Record runtime version/interface, database family, root/child IDs, binding and output evidence. Check schema families independently and report unobserved native behavior explicitly. Compare existing Codex, manager, project-proof, retry, concurrency and portable-ledger coverage against each changed path. Independent reviewers investigate behavior, architecture, data integrity and test effectiveness at the frozen candidate. The orchestrator owns integration, consolidated review, final telemetry and publication.
