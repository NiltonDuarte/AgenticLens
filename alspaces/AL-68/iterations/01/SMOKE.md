# AL-68 source-grounded smoke evidence

Date: 2026-11-01. These observations are implementation evidence, not final consolidated approval. Raw transcripts and request tokens remain outside Git.

## Official protocol grounding

- Codex official hooks reference: https://learn.chatgpt.com/docs/hooks. SessionStart/UserPromptSubmit context reaches the active agent; SubagentStart context reaches the child while callback session_id identifies its parent. SessionEnd is advisory. Trust review is required for non-managed hook definitions. All are protocol facts; callback execution remains separately unverified here.
- OpenCode plugin reference: https://opencode.ai/docs/plugins/.
- Published installed-version types fetched from https://registry.npmjs.org/@opencode-ai/plugin/-/plugin-1.19.5.tgz: package/dist/index.d.ts, Hooks.experimental.chat.system.transform (optional input.sessionID and output.system), Hooks.tool and package/dist/tool.d.ts, ToolContext.sessionID. These established the actual feedback/custom-tool channel before implementation. No lifecycle event return value is treated as agent context.

## OpenCode 1.19.5, Linux amd64, CLI run

Disposable repository: /tmp/al68-opencode-smoke-galj6m56. AgenticLens-owned local plugin installed by the actual native handler. No third-party skill/persona or harness trust settings changed. The run used the user's configured default model rather than a fabricated provider. No synthetic provider was used.

First independent session: ses_e67b05d6affe2pB3HDJUcmJvxg. The prompt explicitly supplied AL-68, asked the agent to follow its hook binding request and perform no source edits. The real agent called agenticlens_bind; runtime JSON reported completed. The local registry records that exact ID, ticket AL-68, explicit_agent_response binding and session.idle checkpoint. The response request token is intentionally excluded from this report.

Source records verified the exact root ID and no native parent. Source-derived compact facts: model gpt-5.6-luna; input 8589, output 58, cache-read 3712, reasoning 37. Cost evidence is observed zero because the actual source records supplied zero; this is not a pricing estimate. Effort was not established by the canonical snapshot and is not invented. The output stream reported two source token snapshots whose summed dimensions match these compact facts.

The actual source database has both complete v1 and v2 table families. Initial reconcile correctly failed under the existing general mixed-family detector. Capture-only exact-ID selection was added and fixture-tested; it selected the uniquely owning v1 session row and then reconciliation reported current with the authoritative compact snapshot. Duplicate IDs across complete families remain rejected and prior facts preserved.

Second independent root: ses_e67ab860dffe3Hvybx8GEcvcIp. It completed its own agenticlens_bind call, then used the real task tool to create native child ses_e67ab72cdffev5PqFh1lPm8lp1. The child also completed its binding interaction. The task result supplied sessionId matching the child's source ID. Authoritative session.parent_id exactly equals the second root ID. After reconciliation, both independent roots and the child remain registered, bound to AL-68 and current; the first root snapshot was preserved.

Source-backed second root facts: model gpt-5.6-luna; input 8872, output 111, cache-read 9856, reasoning 34. Child: same source model; input 9633, output 48, cache-read zero, reasoning 20. All three source model JSON records expose variant default; this is recorded as a source string, not reinterpreted as an effort level. Source cost evidence is observed zero. A repeated unchanged reconciliation preserved the exact compact snapshot digest, establishing no double counting on this real source.

Raw OpenCode event streams are at disposable smoke-output.jsonl and native-child-output.jsonl under the temporary directory and remain outside Git. Both real CLI invocations exited zero without stderr. The generated plugin passed strict TypeScript checking against the exact published runtime types; that compile check is separate from the observed agent interaction.

## Codex 0.166.0, Linux amd64

codex features list reports hooks stable/enabled. Isolated concrete configuration for human review is /tmp/al68-codex-smoke-jykimf10/.codex/hooks.json. It installs SessionStart, UserPromptSubmit, SubagentStart, SubagentStop, Stop and SessionEnd commands invoking the native binary's capture receive with explicit source/store/state paths and a two-second harness timeout. The parent requested human project trust and /hooks review of those exact definitions. No auto-trust edit or bypass flag was used. No callback, agent binding, usage or child capture success is claimed pending that action and actual source comparison.

## Remaining compatibility limits

No runtime smoke has established macOS/Windows, other harness versions, OpenCode v2 native runs, current T3 Codex children, resumed-session interaction or real interrupted-hook recovery. Self-contained fixtures cover selected boundaries separately; fixture success does not establish those runtime paths. Missing events/sources or unassigned registrations must remain visible in capture diagnostics and final coverage reports.

## Fresh smoke after review corrections

Corrected native Linux handler and newly rendered OpenCode plugin installed in disposable /tmp/al68-corrected-opencode-mv67uxul. Actual OpenCode 1.19.5 CLI session ses_e67855b34ffej7NeJ33rgNspY0 received a real binding request and completed agenticlens_bind with explicit AL-68. The run exited zero; no synthetic provider or trust bypass was used. Reconciliation of the exact source owner returned current and explicit_agent_response.

Source-backed compact facts: model gpt-5.6-luna; input 8558, output 53, cache-read 3712, reasoning 18; observed source cost zero. No additional effort value is inferred. Per-owner refreshed_at was 2026-11-01T18:59:02.524692035Z and reconciliation completed at 2026-11-01T18:59:02.528670183Z, explicitly describing completion rather than a common as-of source cutoff. Raw event stream and local registration state remain only in the disposable directory.

The corrected installation rendering and native input path therefore preserve the real agent-feedback/tool-binding/source-refresh exchange on this observed interface. This does not remove the outstanding human-trust/Codex runtime evidence gap or extend runtime claims to untested platforms/families.

The final focused corrections change source-checkpoint precedence, actual source-store alias comparison and bounded child-recovery scheduling. They do not change the installed hook/agent-feedback protocol, OpenCode system-context/custom binding tool or native receive payload contract. The recorded actual OpenCode exchanges carry forward for that interaction boundary; deterministic real-filesystem/SQLite fixtures separately verify the corrected algorithms. No new live-model exchange is claimed for pagination or stale-source scenarios, and none of these corrections closes the pending human-trust/Codex runtime gate. Unknown/equal differing source checkpoints now preserve canonical facts with unresolved diagnostics rather than guessing a newer snapshot.

## Trusted Codex runtime evidence after human review

This section supersedes the earlier pending-trust status for the observed path only. The user confirmed normal project trust and `/hooks` review of the six installed definitions in /tmp/al68-codex-smoke-jykimf10, then sent a message. No agent trust edit, bypass flag, third-party skill/persona edit or application/test change was used. Installed Codex is 0.166.0 on Linux amd64; hooks and multi_agent both report stable/enabled. Product remained frozen at c34cc949278608f4c777f7428a620e2ae29542e7fe33875d5b13ae9f0d17e6ec.

The original user-created source session 01a198fd-22ba-77b2-99bb-bbfb546eef11 has source classification vscode. It received actual hook developer context and asked for a ticket; because no explicit ticket had been supplied, its local registration remains unassigned with event Stop. Its directory was never used to infer AL-68 ownership.

Normal `codex exec` with the workspace-write sandbox in the same trusted directory created root 01a19901-011a-7402-bb1e-c4febe37e7e1. An explicit AL-68 smoke instruction and the real hook developer-context request caused the agent to execute the supplied native binding command successfully. Its JSON acknowledgment reported event UserPromptSubmit and binding explicit_agent_response. It later received actual SessionEnd. A separately launched reviewer-context root 01a19901-c38e-7f22-b224-a334910e4d8b completed the same actual request/bind exchange. Both classify as RootAgent in normalized source facts; reviewer context describes the smoke task and is not invented source persona metadata.

The second root used the native spawn_agent tool and awaited native child 01a19901-dda1-7460-a51f-24423a5d19e8. The actual SubagentStop callback durably queued the child transcript. Authoritative session_meta.id and parent_thread_id, including the native thread_spawn metadata, identify that exact child and parent. Reconciliation bound the child through source_parent and captured all three AL-68 sessions as current. No assumption that a generic interface agent identifier equals a harness source ID was needed. The child's rollout includes inherited parent history; the parent's earlier binding call is not attributed as a child action. No independent child binding command is claimed. The child's normalized role is capture_child; native source agent_role is null.

| Lifecycle event | Observation on the tested path |
| --- | --- |
| SessionStart | Root startup delivered AgenticLens model-visible context; two startup/before-prompt requests are present before the first root tool. |
| UserPromptSubmit | Both root binding acknowledgments explicitly carry UserPromptSubmit. |
| SubagentStart | Not observed on the native exec spawn path. No pending-start record existed before reconciliation and no emitted child-start context or SubagentStart failure marker appeared in root/child source. |
| SubagentStop | Actual queued child path, source identity and proven parent were observed. |
| Stop | Observed on the original trusted user session; completed exec registrations retain their later SessionEnd event rather than historical callbacks. |
| SessionEnd | Observed for both independently completed exec roots and the resumed root. |

A bounded feature/source/stderr check found no established SubagentStart handler failure; enabled features alone do not prove delivery of that event. The local TUI log was unavailable. The reason for absent SubagentStart delivery is not established. Running or interrupted native-child capture before its Stop remains degraded/unverified on this observed interface. This smoke therefore establishes root agent interaction and completed native-child source capture, not execution of all six callbacks or complete interrupted-child coverage.

The first root was resumed normally with `codex exec resume` using its exact ID. It preserved its binding, made no second binding command and advanced cumulative source usage. Reconciliation replaced that growing snapshot, preserved the reviewer/child snapshots, and an unchanged repeat preserved the exact compact snapshot digest. The original user-created unassigned session remained local and excluded from ticket outcomes throughout.

| Source session | Model | Effort evidence | Compact input excluding cache | Cache read | Output | Reasoning tokens | Cost |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- |
| 01a19901-011a-7402-bb1e-c4febe37e7e1, after resume | gpt-6.1-sol | Unknown; source settings null | 7266 | 39552 | 188 | 0 | Unavailable |
| 01a19901-c38e-7f22-b224-a334910e4d8b | gpt-6.1-sol | Unknown; source settings null | 8145 | 56448 | 269 | 0 | Unavailable |
| 01a19901-dda1-7460-a51f-24423a5d19e8 | gpt-6.1-sol | Source explicitly reports low in its own context | 908 | 16128 | 27 | 18 | Unavailable |

Raw source input counts include cached input: resumed root 46818, reviewer root 64593 and child 17036. They equal compact uncached input plus cache-read; output/reasoning values match source records. Before resume, the first root had cumulative source input 30175 and output 177; after resume these became 46818 and 188. No model, effort, price or zero cost was fabricated. CODEX_SMOKE.json stores normalized evidence only; raw logs, source transcripts and binding challenges remain outside Git under the source store/disposable smoke directory. Native non-Linux, other Codex versions/interfaces and actual interrupted-before-Stop child behavior remain unverified.

### Parent source-metadata verification

Parent independently read the exact owning session_meta for all three runtime sessions and matched its id to the selected source ID. The child source is a native subagent.thread_spawn with parent_thread_id01a19901-c38e-7f22-b224-a334910e4d8b, matching the registry's proven source_parent binding. CODEX_SMOKE.json originally carried the parent's exec/null-parent normalization in the child row; parent corrected that evidence from the child's own metadata and retained a normalization-correction record. Inherited parent history is not child metadata or a child binding call. The actual source/registration linkage was correct; no product code changed.
