h3. Title: AL-68: Install harness lifecycle hooks for automatic session capture

*TARGET*: AgenticLens (Installation / Setup / Domain / Codex and OpenCode Adapters / CLI / Persistence / Tests)

*Status*: Done
*Work type*: feat
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 follow-up; user requested a ticket to add lifecycle hooks to the installation script after two independently launched reviewers were omitted from the ticket ledger.
*Related*: AL-62 owns granular skill/delegation facts; AL-63 audits harness capability parity. Reuse their delivered contracts when available; this ticket owns installation and lifecycle capture, not their broader migrations or analytics.

*CONTEXT*

REV-5 recorded two Codex reviewer sessions with verified models and usage, but the final root-session extraction omitted them because their source records contained no native parent link. The interface launched separate harness sessions. T3 is an interface, not a telemetry source. Client repositories may use skills/personas that they do not own; editing those files or relying on reviewers to remember a final telemetry command is not an acceptable integration mechanism.

Current setup installs a release binary/operator skill and proposes final-generation instructions in AGENTS.md. It does not install lifecycle capture or reconciliation. Relevant delivery surfaces include scripts/download-agentic-lens.sh, the packaged setup scripts under .agents/skills/agentic-lens-setup/scripts/, the setup skill, and the CLI installation/update paths. Determine the canonical installation entrypoint and keep all shipped paths consistent; do not assume the downloader alone performs repository setup.

*GOAL*

As an adopting repository owner, I want installation to configure supported Codex and OpenCode lifecycle hooks once, so existing agent workflows automatically register and capture sessions and expose capture gaps without changing user-owned skills or personas. Deliver an end-to-end install, capture, reconcile and diagnose slice. Hook configuration pointing to an absent handler does not satisfy this ticket.

*VERTICAL SLICE SCOPE*

* Installation and Setup: Install AgenticLens-owned hook adapters and release-compatible handlers, capability-check the selected harness/version, and present the additive changes through the existing setup approval flow. Support reinstallation, upgrades, disabling and removal while preserving unrelated configuration. Retain release-only installation and the existing standalone binary distribution; do not require a new Go/Node toolchain in client repositories. OpenCode glue executes within the existing harness runtime.
* Domain Core: Pure session identity, explicit ticket/parent association, idempotent snapshot reconciliation and completeness rules. Keep source-derived facts distinct from registration/correlation metadata and preserve unknown values.
* Boundary Adapters: Normalize supported harness lifecycle events into a durable local registration/capture path. Use authoritative source records for model, effort, usage and cost when available. Handle concurrently active roots and children, resumptions, interrupted writes and unavailable source stores.
* CLI and Persistence: Provide inspectable installation health, registered-session and reconciliation results. Integrate final ledger generation with registered sessions, preserving existing other-source facts. Keep intermediate state local and ignored; retain final ledger/summary publication at workflow completion.
* Quality: Self-contained installer and source fixtures plus real harness smoke evidence. CLI diagnostics are the presentation surface; no new dashboard feature is required.

*ACCEPTANCE CRITERIA*

* Supported installation paths configure the selected harness without editing third-party skills/personas or overwriting existing hooks. Keep any AGENTS.md block small and limited to ticket binding/final publication guidance. Automatic collection must not depend on a reviewer reading that block.
* Inspect actual harness capabilities during implementation. For Codex, evaluate SessionStart, SubagentStart, SubagentStop, Stop and SessionEnd; for OpenCode, evaluate session creation/update/status/idle/error events through its plugin API, separately for supported schema/version families. Record tested versions and actual payload semantics. Unsupported, disabled, untrusted or unverified hooks produce an actionable degraded status, never a claim of complete capture. Respect harness trust controls; never auto-approve or bypass them.
* Bind each session to a ticket through an explicit session-specific association, inherited proven parent binding or equivalent unambiguous registration. Do not infer ownership solely from directory, branch, role, model or temporal proximity. Multiple tickets/roots sharing a checkout cannot steal each other's sessions. Unbound sessions remain discoverable and explicitly unassigned; document the minimal binding mechanism when no authoritative ticket context exists.
* Capture independent sessions as well as native children. Persist native parent links only when proven. Delegation tool events may supply explicit correlation when verifiable identifiers are present; do not assume an interface task/thread ID equals a harness session ID. Missing parent evidence remains unresolved, distinct from missing session capture. No T3 telemetry source or dependency on T3 is introduced.
* Event handlers perform bounded durable handoff rather than long extraction inside shutdown hooks. Repeated/out-of-order events, duplicate registration, resumed sessions and growing cumulative usage must not double count facts or regress a newer snapshot. Recover pending work after a crash/restart; prevent hook recursion and concurrent-writer corruption. Idle/turn completion is a checkpoint, not proof of ticket completion; do not rely solely on shutdown callbacks.
* Reconciliation refreshes registered sessions from their actual source and reports pending, failed, unsupported and unresolved states separately from known zero activity. Preserve already stored facts and other-harness sessions when enrichment/source access fails or records are partial. Preserve unavailable model/effort/cost as unknown instead of fabricating values. Any required schema changes use the established migration/preservation contract and coordinate with AL-62.
* Installation is idempotent and upgrade-safe. Dry run exposes planned changes; conflicting local edits are preserved and explained. Failed setup rolls back only AgenticLens-owned changes. Disable/uninstall removes only owned entries/files. Support configured installation/workspace/source paths, paths containing spaces and the platforms advertised by the existing installer. Hook failures leave coding usable while diagnostics visibly report degraded capture.
* Raw transcripts, secrets and intermediate capture state remain outside Git; hook payloads are treated as data, never interpolated as executable shell input. Collection remains local. Final generation publishes a reconciled ticket ledger/summary with a documented snapshot cutoff and explicit coverage limits, rather than rewriting committed artifacts on every event.
* Verify fresh install, repeated install, upgrade, existing user hooks, conflicting edits, rollback/removal, absent binaries and unsupported/untrusted harnesses. Exercise native parent/child capture, independent roots, concurrent tickets, duplicate/reordered events, resumed sessions, late source flush, crashes and missing source data with self-contained fixtures.
* Run a real parent/child smoke check through each claimed supported harness/interface path. Verify IDs, available model/effort, usage, preserved sessions and unresolved links against source records. A config file or successful hook exit alone is not evidence of capture. Test independent reviewer sessions without modifying their skills. Report unavailable environments explicitly; do not claim unverified compatibility.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session. This ticket creation authorizes planning only; the developer records the concrete implementation contract and scope approval before code changes.
* Ship the installer integration, working handlers, reconciliation, CLI health diagnostics and operator/setup documentation together. Document capabilities, supported versions, upgrade/removal and recovery procedures at their canonical sources.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md: FCIS, strict typing, zero any/interface{}, zero emojis, no domain mocks and self-contained fixtures. Pass applicable regression checks, independent focused review and the required make gate.
* Publish source-grounded smoke evidence and the reviewed implementation PR. Preserve historical REV-5 evidence; tests must not depend on its live workspace or private source files.

*RESEARCH REFERENCES*

* Codex lifecycle events, trust, identities and tool coverage: https://learn.chatgpt.com/docs/hooks
* OpenCode project plugins and lifecycle events: https://opencode.ai/docs/plugins/
* These sources informed the proposal on 2026-11-01; validate the installed versions during implementation. The REV-5 host reported codex-cli 0.166.0 with hooks enabled; this was a capability observation, not a hook capture test.
