# AL-64 iteration 04 native runtime capabilities

Status: bounded investigation completed on 2026-11-06 local time (2026-11-05 23:08 UTC). Planning baseline: `d8a7c14cfd9f0ee626fedc33c70f92bd928d9192`. This report owns no application or test changes. No new agent session, lifecycle event, queued message, source edit, telemetry generation or continuation action was executed.

## Decision for this implementation

No installed harness has a verified transport giving a generic `generate` invocation both independently trustworthy executing-caller identity and that caller's current native interval ID/start. Consequently, production F01/F02 must not be enabled by trusting an environment variable, extraction root, nonce, source path or registry receipt. This is a limit on stronger advice, not a reason to reject extraction or change duration/accounting or portable-history policy.

Source evidence can identify an open native interval independently of who invoked generation. Use F03 when such exact evidence is available and eligible, and F04 when the approved subtotal-only conditions apply. Several individually known short intervals remain a no-feedback case; caller uncertainty cannot restore the removed aggregation-only case.

Continuation command/API surfaces exist, but a real completion followed by a new turn and refresh has not been independently verified here. Mark automatic next-turn continuation **unverified**, rather than claiming the harness has no continuation feature. F05 remains the approved notice appended once after an eligible F01; its absence when no F01 can be established is intentional. No continuation action is added by this investigation.

## Per-harness matrix

| Harness/source family | Observed installed version | Caller transport into generic generate | Native interval evidence | Completion / next-turn capability | F01/F02 decision |
| --- | --- | --- | --- | --- | --- |
| Codex | CLI `0.166.0`, T3 Code originator | Shell exposes `CODEX_THREAD_ID`, `CODEX_SESSION_ID`, `CODEX_TOOL_CALL_ID`. Child and parent values differ. Source metadata corroborates their different owners, but a trusted invocation-bound transport is **unverified**. | A genuine current child rollout contains `task_started` with native `turn_id` and timestamp. This verifies source-open evidence, not independently authenticated executing caller/current-turn context. | Ordinary terminal records exist in parent history. `queue --resume`, interactive `resume`, and `exec resume` are advertised. A fresh natural-completion -> later turn -> telemetry refresh cycle is **unexecuted/unverified**. | Unavailable in production without additional independently verified context. Exact source-open evidence can support F03. |
| OpenCode schema family v1 | CLI, plugin and SDK `1.19.5`; local DB has `session/message/part` tables | Plugin tool context declares native `sessionID` and `messageID`; tool hooks also expose session/call IDs. Existing AgenticLens plugin uses that context for binding only. No installed generate wrapper/transport was verified. Generic-generate caller is **unverified**. | Adapter reads native assistant message IDs/creation and running-tool or streaming-part timing. This is assistant-message evidence, not a whole harness turn. Source tracing is verified; an authentic executing OpenCode generate cycle is **unexecuted**. | CLI `run --session/--continue` and SDK message/prompt APIs are advertised. Actual completion and next-turn refresh through the generic command are **unverified**. | Unavailable. F03 can use properly extracted source-open assistant-message evidence; otherwise eligible F04. |
| OpenCode schema family v2 | Same installed CLI/plugin/SDK `1.19.5`; local DB also has `session_v2/session_message` tables | Same plugin declarations do not establish a separate family-v2 generate caller transport. Native callback execution and its handoff into generate are **unverified**. | Adapter uses native assistant-message IDs/timing, running tool state and optional `session_v2.time_idle`, respecting superseding messages and cutoff. Source tracing is verified; authentic current family-v2 execution is **unexecuted**. | SDK/CLI surfaces do not establish family-specific after-completion semantics. Native family-v2 completion -> next-turn -> refresh is **unverified**. | Unavailable. Qualified source-open/subtotal cases remain supported within the approved evidence boundaries. |

The OpenCode adapter family labels describe storage schemas, not installed CLI major versions or SDK import directories. The presence of both schemas, or a `version` field in a source row, does not certify that a corresponding native runtime cycle occurred.

## Evidence and limits

### E1: installed commands, help and generic invocation

Read-only commands executed successfully:

- `codex --version`, `codex --help`, `codex exec --help`, `codex app-server --help`, `codex queue --help`, `codex agents --help`.
- `opencode --version`, `opencode --help`, `opencode run --help`.
- `bin/agentic-lens-linux-amd64 generate --help`.

Resolved Codex executable: `/home/niltonduarte/.codex/packages/standalone/releases/0.166.0-x86_64-unknown-linux-musl/bin/codex`. OpenCode executable: `/home/niltonduarte/.opencode/bin/opencode`. Installed `@opencode-ai/plugin` and `@opencode-ai/sdk` package metadata both report `1.19.5`.

The existing generate help labels `--session-id` as an explicit **root** and `--nonce` as session-resolution input. It does not document an authenticated current-caller or current-native-interval handoff. `backend/cmd/agentic-lens/main.go`, `runGenerate`, confirms those extraction inputs. That checked binary is the existing baseline host build; this read-only inspection does not certify a later rebuilt implementation/help output.

`codex queue --help` advertises `--thread`, `--message` and optional `--resume`. Its documented resume failure can produce a warning while the command still succeeds. A successful queue exit alone therefore cannot prove continuation, and help does not prove execution specifically after the current turn completes. Nothing was queued or resumed.

OpenCode `run --help` advertises session continuation and attaching to a server. The installed SDK declares `session.prompt` and `session.promptAsync`; the latter is described as creating/sending a message, starting if needed, and returning immediately. These establish available API declarations, not observed next-turn scheduling semantics or successful telemetry refresh. No server was started or contacted and no prompt API was invoked.

### E2: real Codex child and parent metadata, without transcript disclosure

A filtered environment probe printed only relevant identifiers/version/originator. Other environment entries, credentials, prompts, tool arguments and transcript bodies were not printed. Read-only SQLite queries used URI `mode=ro`; rollout inspection reduced records to session metadata and native lifecycle IDs/timestamps.

Observed environment:

- `CODEX_THREAD_ID=01a1adf6-782f-7b92-a0ea-39407c971b2f`.
- `CODEX_SESSION_ID=01a1a7d8-b9d2-7383-a6f7-3c7002ba88e6`.
- `CODEX_VERSION=0.166.0`; originator `T3 Code`.

`~/.codex/state_5.sqlite` and the matching `session_meta` records independently identify the first UUID as this investigator's child session and the second as its parent. The child's native `source.subagent.thread_spawn.parent_thread_id` equals the parent UUID. Thus `CODEX_THREAD_ID` happened to identify this child in the observed transport, while `CODEX_SESSION_ID` was inherited parent context. Neither variable name alone certifies generic caller identity, and this observation must not be generalized across harnesses/launch modes.

The child's real rollout contains an open native turn:

- Owner session: `01a1adf6-782f-7b92-a0ea-39407c971b2f`.
- Native `turn_id`: `01a1adf6-783b-77d2-9cca-df802f373c28`.
- Native start timestamp: `2026-11-05T23:06:13.447Z`.
- No matching terminal event was present at the bounded read.

This is authentic **source-open** evidence. Its lack of a terminal event does not prove process liveness, a trustworthy invoking caller or a completed future cycle. The environment's `CODEX_TOOL_CALL_ID` had no exact `function_call`/`custom_tool_call` `call_id` match in either inspected rollout at that read. That absence is not proof of unsupported transport, source completeness or an error; it simply supplies no independent invocation binding. Source metadata plus an environment claim is insufficient for stronger advice under the approved plan.

Parent source metadata includes genuine `task_complete` records followed by later task starts. Those historical lifecycle records establish that native completion events exist, but were not correlated to an independently proved generate invocation and subsequent refresh. They are not presented as this iteration's end-to-end continuation verification.

### E3: existing hooks and plugin do not convey equivalent generate context

`backend/internal/adapters/capture/hooks.go` parses native lifecycle hook payloads containing `session_id`, event and transcript pointers. It explicitly handles Codex `SubagentStart` as carrying the parent's session ID, rather than silently attributing it to a child. The accepted event set includes `SessionStart`, `UserPromptSubmit`, `SubagentStart`, `SubagentStop`, `Stop` and `SessionEnd`. This lifecycle receiver is separate from generic `generate`.

`backend/internal/adapters/capture/plugin.ts` forwards OpenCode lifecycle session IDs to `capture receive`; its system transform calls `capture request`. Its `agenticlens_bind` custom tool takes `context.sessionID` from runtime context and forwards it to `capture bind`. It contains no generate custom tool, `shell.env` integration for generation, interval handoff or continuation scheduling. These are source-level facts, not executed plugin callback proof.

The installed plugin declarations in `~/.config/opencode/node_modules/@opencode-ai/plugin/dist/tool.d.ts` define `ToolContext.sessionID` and `messageID`. `dist/index.d.ts` declares `tool.execute.before/after` session and call IDs, and optional `shell.env` session/call IDs. Their existence suggests possible integration points for separately approved work, but does not prove that the current command receives them or establish trustworthy interval start/time semantics. No such integration is proposed here.

The current checkout has no `.codex/hooks.json`, `.codex/agenticlens-installation.json`, `.opencode/plugins/agenticlens-capture.ts` or `.opencode/agenticlens-installation.json`. This is only a local file-presence observation; it does not establish global harness configuration or absence of all runtime mechanisms. No private configuration/authentication contents were read.

### E4: interval source traces and schema metadata

`backend/internal/adapters/codex/rollout.go` reads `task_started`/`turn_started` native turn IDs and timestamps, matching completion and interruption records. It groups numeric duration by model later. Exact interval evidence must retain the native identity before that grouping, rather than reconstructing an interval from an invocation subtotal or model row.

`backend/internal/adapters/opencode/schema_v1_session_extraction.go` selects running message IDs using native tool state/timing or open text/reasoning-part timing and rejects superseded evidence at the cutoff. `schema_v2_session_extraction.go` reads its schema's native idle state and tool content, and bounds assistant-message activity by later records. Neither parser establishes which process invoked generation; an assistant-message interval must keep that truthful label.

Read-only `PRAGMA table_info` confirmed local v1 and v2 table shapes and v2 `time_idle`. Metadata-only queries confirmed timestamp field shapes. No message content was emitted. Those source records may include historical or fixture-generated data; table presence, row version labels and timestamp field presence are **schema evidence only**, not authentic OpenCode runtime proof. No current OpenCode session was independently identified or exercised.

## Verification boundary retained

This investigation did not generate deterministic fixtures or run tests. Existing tests can validate selector behavior for verified/unknown caller values, exact intervals, F05 and fallback wording, but cannot certify an installed native transport. Installed help, installed type declarations, authentic source-open records and a verified executing-caller handoff remain distinct evidence classes.

Outstanding native cases are explicitly unexecuted: trustworthy generic caller/current-interval handoff for Codex and both OpenCode families; independently observed natural completion followed by a later turn and telemetry refresh; and any automatic next-turn operation. Their absence does not block supported F03/F04 behavior, alter the previously accepted numeric producer, or create a new liveness/accounting/authenticity requirement.
