# GEN-14 iteration 01: T3 thread control and AgenticLens CLI

## Scope and evidence

Documentation-only audit of baseline 85a2af9. No application, test, schema, or runtime script changes. Source: Codex 0.166.0 rollouts on this machine and T3 thread APIs, inspected 2026-10-31. No raw transcripts are stored here. Session start filenames use local clock formatting; source timestamps and T3 run timestamps were inspected separately.

The user-authorized hello-world message was delivered at 13:20:08 UTC; the target replied “Hello, world!” at 13:20:10 UTC. T3 records the incoming item as `createdBy: agent`, `creationSource: mcp`, `type: user_message`, with the controlling thread in `senderThreadId`. The target thread has two completed runs and no T3 parent.

Controller: <t3-thread://v2/ccbaed6c-5335-443d-9aaa-69948fbd15d1/7b8bd224-24ce-4495-bb4c-39a605235612>

Target: <t3-thread://v2/ccbaed6c-5335-443d-9aaa-69948fbd15d1/d70b23b4-2e55-4b42-b0c3-7e3e8164c333>

Native Codex controller ID: `01a1921c-e1ff-7783-87b3-edfd59fa6aae`.
Native Codex target ID: `01a1921d-776d-7ed3-8aa9-3e79df5490f5`.

These exact rollout identities were read from session_meta, corroborated by the matching worktree and hello-world source messages. A working-directory match alone is not a supported discovery selector. Both native metadata records report `source: vscode` and no `parent_thread_id` or agent role. Both target turns remain in one native rollout.

## Findings

### High: cross-thread control does not establish telemetry hierarchy

`backend/internal/adapters/codex/discovery.go:36` selects a native root by nonce or native session ID and includes only that root and its metadata descendants. `rollout.go:115` understands native structured subagent provenance. It does not query T3 or infer relationships from calls/messages. `rollout.go:839` persists parent IDs only from native metadata.

Consequently, generating the controller's ledger excludes this independently created target. Sending or restarting a top-level thread is not equivalent to spawning a native Codex subagent. Shared project/ticket text does not change that result. A single session generation cannot prove all cooperating top-level threads were captured.

Native subagents with source-backed parents are supported by existing tests. T3 delegate_task, fork, launch, cross-provider child mapping and remote machines were not exercised; no claim that their native provenance always has or lacks parent metadata is justified by this test.

### High: completed T3 MCP calls are omitted in this rollout format

The real controller rollout contains completed `McpToolCall` items for `t3_thread_search`, `t3_thread_read`, and `t3_thread_send`. `rollout.go:665` accepts only `CommandExecution` completed items. The `exec` wrapper is deliberately excluded at `rollout.go:626`. These records have no corresponding named function-call records or legacy MCP begin/end events in the observed source.

Thus the source contains evidence of the send, but this parser drops that call from per-tool facts/events. Command calls can still be counted, and turn/token totals still have their own source records. This is a parser gap, not proof that T3 failed to log the operation. Legacy MCP lifecycle records are separately supported at `rollout.go:422`.

### Medium: agent-sent input is counted as steering

The target rollout represents hello-world as a user-role message. Its passthrough metadata labels it `user.text` with `provenance.type: user`; it contains no T3 senderThreadId. `rollout.go:608` increments SteeringTurns for every user-role response message, with semantic event counts preferred when present. `CombineTranscripts` takes steering from the selected root.

The source snapshot contains three user-role records in the target: injected instruction/environment content, the user's initial test prompt, and agent-sent hello-world. There are two actual runs. The existing parser cannot distinguish human feedback from agent messaging or injected setup at this boundary. Do not interpret this count as a verified number of human interventions.

### Operational: use native identifiers and explicit source

The configuration has no machine-independent default selecting Codex; the CLI defaults to OpenCode absent a machine-local override. Pass `--source codex`. `--session-id` expects the native Codex ID above, not a T3 URL/UUID, a T3 run ID, or the branch ticket number. Codex home defaults to configured local settings, CODEX_HOME, then ~/.codex. No override is needed for the observed normal store.

A literal unique nonce in the current source transcript is an alternative. Copying one nonce between independent roots makes selection ambiguous. In this version adding a session ID does not resolve a multi-root nonce collision: discovery checks nonce ambiguity first. Do not propagate the controller's final extraction nonce to other roots.

## Supported operating workflow

1. Generate each ticket's telemetry in its owning thread/worktree at that thread's final step. Select the actual harness and native session identity or a fresh literal nonce.
2. Keep unrelated tickets in separate ledgers. Do not gather all open project threads into one ticket merely because one thread controls them.
3. If several independent roots truly belong to one ticket, repeated explicit native-root generations into the same ledger merge sessions by identity; this retains roots rather than inventing a parent relationship. Cross-thread code/Git attribution and root steering are separate concerns: generation uses the invoking checkout/code ref, and code SDLC iterations sum persisted UserIterations across root-session invocations (main.go:363), including any agent-sent/setup records the parser classified as steering.
4. Re-running a selected root replaces that root's cumulative snapshot. Existing tests cover idempotent refresh, model changes, child identity, and preservation of unrelated roots/harnesses. Verify actual session IDs and source labels after each extraction.
5. The dashboard's historical workspace events can hydrate stored Codex sessions from local rollouts (`live/service.go:1066`). Live polling/session-tree control remains SQLite-backed in this baseline. This is not a T3 thread management API or a guarantee that every open T3 thread is visible before ledger ingestion.

The CLI does not control T3 conversations; T3 APIs perform sends, steering, interruption and organization. AgenticLens consumes resulting native telemetry. T3 ownership/delegation links and AgenticLens session parent links have distinct meanings. Existing task_links production here is OpenCode-specific (`main.go:283`), not a T3-message bridge.

## Recommendations (not implemented)

- Add McpToolCall lifecycle decoding with the same ownership, error, duration and deduplication guarantees as existing tool facts; prove it using a minimized real-format fixture.
- Define a T3 bridge contract mapping T3 thread/run IDs to native provider session IDs and typed message/delegation relationships. Preserve independent root identity; do not rewrite session parents from arbitrary messages. Keep provider/remote provenance explicit and avoid importing prompts/results.
- Preserve agent-versus-human provenance if T3 can export it. If native source cannot establish it, label the metric as input/steering records rather than claiming measured human interventions. Separate injected setup from actual prompts.

## Verification

Initial Git status was clean; only GEN-14 documentation is owned by this audit. No fixture/test additions are warranted for documentation-only findings. Full make and final CLI checks will be recorded below. No browser verification, remote-provider experiment, source mutation, or other-ticket ledger update is authorized or claimed.
