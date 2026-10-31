# GEN-14 iteration 02: manager as a user proxy

## Assessment

The hello-world source evidence matches the requested isolation model: two independent native roots, a user-like destination message, and no metadata parent relationship. The manager's selected root cannot discover the target through that send. This is intended behavior.

Read the source identities, T3 links and delivery timestamps in the historical iteration 01 review. The user clarification supersedes its suggested aggregation/human-origin requirements. No application implementation was requested or performed.

## Ownership contract

- Manager ticket: own workspace/ledger, root-session reasoning, context/token usage and control calls.
- Managed ticket: own workspace/ledger, destination reasoning/usage and user-like steering from the manager. Native subagents belonging to that ticket stay beneath its root.
- Message provenance: T3 truthfully records agent/mcp sender metadata; Codex receives the hello-world text as a user-role message. User-like interaction does not require falsifying origin.
- Results/status text read by the manager becomes manager context and incurs manager processing. That is local manager usage; it does not justify copying the target's usage counters into the manager ledger.
- Separate code attribution requires extraction in each ticket's checkout or with its intended code ref. A remote control message does not switch the CLI's current checkout.

## Operational approach

Use existing independent top-level threads and t3_thread_send. Default auto starts an idle thread, steers a running turn or queues as appropriate; use queue for a separate next turn when needed. Do not use native subagent spawning or delegation as a substitute for controlling an existing ticket root. This audit did not exercise every T3 creation/fork/delegation path.

Each owning ticket thread performs its final telemetry command using its own unique literal nonce or exact native Codex session ID and its own output paths. Use --source codex explicitly in this setup. Do not send the manager's final nonce to managed threads. Do not merge managed-ticket roots into the manager's ledger. Never select all sessions by shared project path/title.

The relevant implementation is metadata-only tree discovery in backend/internal/adapters/codex/discovery.go, parent persistence in rollout.go:839, user-role steering handling in rollout.go:608, and identity-based refresh merge in backend/cmd/agentic-lens/main.go:1461. Reading a T3 response does not cause its native session metadata to become the manager's session metadata.

## Remaining limitations

The controller rollout has completed McpToolCall records for T3 reads/search/send, but rollout.go:665 handles only CommandExecution completed items and skips exec wrappers. Tool operation counts/events can therefore omit the manager's T3 control calls even while token/turn metrics are present. This is the concrete instrumentation gap to fix later; fixing it must preserve root isolation.

The target has two real runs but three user-role records because injected instructions/environment are also represented as a user message. The current parser counts those records as steering. Counting the manager's hello-world as destination steering is desired; counting setup content may inflate the number of actual interactions. No claim that destination steering exclusively counts human-authored messages is intended.

This audit covers local Codex 0.166.0 and the actual hello-world exchange. It does not establish behavior for remote stores, other providers, new T3 thread creation, fork inheritance, every existing project ledger, or ticket threads whose own native parent trees differ. No browser, cross-provider or remote-machine tests were run.

## Verification

Baseline source 85a2af9; documentation commit 2eaf4b9. Initial working tree clean. Fetch/merge origin/main returned already up to date, so application code is unchanged. The first make run passed Go race tests, frontend type/lint checks and 712 frontend tests, but was interrupted during multi-platform builds when the user clarified scope. A fresh post-sync full make run is being completed; final status is recorded below. No tests were added or removed. No source sessions or other-ticket ledgers were modified.

### Actual CLI diagnostic

Built source CLI `bin/agentic-lens-linux-amd64 generate --help` confirmed source/session/nonce/type/output flags. Extracted the target's exact native session into `/tmp/agenticlens-gen14-audit-xz0pgmxu/target.json`, labelled as GEN-14 diagnostic data rather than modifying its ticket ledger. Exit 0; exactly one Codex session, target ID 01a1921d-776d-7ed3-8aa9-3e79df5490f5, no parent and no manager session. The target had no tools. Its code.sdlc_iterations was 3, corroborating that setup plus initial prompt plus hello-world are counted rather than two actual run boundaries.

Passing the target T3 UUID to --session-id returned nonzero “Codex session ... was not found”, as expected. The T3/native identity distinction is an observed CLI boundary, not a guess. No target conversation messages were sent during this audit and its raw store was opened read-only.
