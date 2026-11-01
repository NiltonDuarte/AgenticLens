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
