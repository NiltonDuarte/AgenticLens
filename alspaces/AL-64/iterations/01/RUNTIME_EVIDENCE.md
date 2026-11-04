# AL-64 native OpenCode activity observations

Observed 2026-11-04 by the runtime test engineer, gpt-6.1-sol/high (inherited dispatch settings). Scope: native source boundary evidence, not application review or complete generation/backfill acceptance.

## Isolation and method

Used installed OpenCode v1.19.5 at /home/niltonduarte/.opencode/bin/opencode and OpenCode v2.2.0 at /home/niltonduarte/.local/bin/opencode2. Each run used a unique disposable project, HOME, XDG data/config/state directories, absolute OPENCODE_DB and PWD. No original database, historical ticket ledger or source application file was modified. Each local deterministic HTTP provider listened on a different OS-assigned loopback port, with no external provider credentials. Existing AL-58 provider examples were read as references without modification.

A disposable Python driver returned either an initial text stream chunk and withheld completion, or a native bash/shell tool call executing sleep 8 before returning a final text response. While the native runtime was blocked, the driver inspected SQLite read-only. Tool-run snapshots used SQLite backup to disposable active.db files. Completion runs then released the stream or waited for the tool to finish. Separate cancellation runs sent SIGINT only to the probe process group. All probe processes and local servers exited; a final /proc environment check found no surviving process using these probe database paths.

The first attempt preserved inherited PWD while setting subprocess cwd; OpenCode selected this checkout's directory/configuration and the custom model was unavailable. Both disposable databases recorded failure before any model request. The retry explicitly set PWD to the disposable project and succeeded. This correction matters for reproducing isolation with these runtimes. The first v2 failure also produced an idle record with outcome failed. No failed attempt is reported as active evidence.

Private experiment drivers, logs and raw databases remain outside Git under /tmp/al64_runtime_* and /tmp/al64-runtime-*. The source observations below contain IDs and structural/timing evidence only, not transcripts. Numeric timestamps are native epoch milliseconds. Labels in the private driver's JSON named all initial snapshots streaming_without_running_tool; the tool experiments did contain running tools as described here.

## v1.19.5 text streaming, without a tool

Session ses_e58108e74ffeXNNxTduSa9hG8Q; assistant msg_1a7ef7e5f0012nMWmsGcoiucpw.

- During the blocked model stream, assistant payload time.created was 1793818852959, with no completed field. A native step-start part existed. Its text part carried time.start 1793818856862 and no end.
- Observation cutoff was 1793818858804. Source assistant creation to observation was 5845 ms; this is source arithmetic, not a generated ledger assertion.
- Once released, native text timing gained end 1793818858864, step-finish appeared, and assistant time.completed became 1793818858921 with finish stop. The measured assistant interval was 5962 ms. CLI exited 0.
- This verifies a real unfinished streaming shape without a running tool. Restricting all supported v1 activity to running tools would exclude this shape. Assistant creation alone and an open timed text/step lifecycle are distinct evidence.

## v2.2.0 text streaming, without a tool

Session ses_e581054c4ffeCHdXcCPH3jUsLp; assistant msg_1a7efbe6c001qqqydGvq2IHIui.

- During the blocked stream, assistant time.created was 1793818869506 without completed. Native text content had type and text; the completed payload confirms no text-level timing/state structure. Initial summary sanitization did not retain arbitrary text-item fields, so absence of every hypothetical initial field is not independently asserted.
- The session had time_idle null and time_suspended 1793818869160, earlier than assistant creation. No idle message was present. Observation cutoff was 1793818871537, 2031 ms after assistant creation.
- After release, assistant time included streamed 1793818871556 and completed 1793818871561, finish stop. Session time_idle became 1793818871576 and time_suspended became null; an idle message at that timestamp had outcome succeeded. CLI exited 0.
- A nonnull session time_suspended is therefore not, by itself, proof that this native runtime has stopped working: it was waiting on the observed provider stream. Reliable inference from only persisted fields must be established separately; this experiment establishes live observation, not a universal historical liveness rule.

## v1.19.5 running tool and completion

Session ses_e580f7012ffe8b6MVhH7acUurI; assistant msg_1a7f092f40010Iq1Ww1MgzCnz7.

- During sleep 8, assistant time.created was 1793818923764 without completed. The tool part had state.status running and state.time.start 1793818924441, without end.
- Snapshot cutoff 1793818926407 was 2643 ms after assistant creation.
- Completion changed tool state to completed and added state.time.end 1793818932556; assistant time.completed became 1793818932583 with finish tool-calls. The next assistant completed its own separate interval 1793818932589 through 1793818932690. CLI exited 0.
- Preserve assistant-message accounting: the first message includes tool execution, while the next assistant remains a separate measured interval. Do not add tool duration to assistant duration again.

## v2.2.0 running tool and completion

Session ses_e580f456fffee1dlB5bKAdcthk; assistant msg_1a7f0c362001eVjH6jaqfxLy6t.

- During sleep 8, assistant time contained created 1793818936177 and streamed 1793818936229, without completed. The content tool item had type, id, name, executed, state and time.
- Native timing is on the tool content item's top-level time, not state.time: created 1793818936214 and ran 1793818936224. Its state contained status running.
- Snapshot cutoff 1793818938185 was 2008 ms after assistant creation.
- Completion added content-item time.completed 1793818944340 and changed state.status to completed. Assistant time.completed became 1793818944354 with finish tool-calls. The next assistant interval was 1793818944434 through 1793818944492, followed by idle succeeded at 1793818944507. CLI exited 0.
- The pre-change adapter's nested state.time shape does not read this native tool timing. The implementation owner was notified; this is native shape evidence, not a claim that every adapter defect is already fixed.

Private active/final snapshot locations for implementation probes: /tmp/al64-runtime-gtnq_h32/v1/{active.db,opencode.db} and the corresponding v2 directory. They are disposable evidence, not permanent test prerequisites.

## Cancellation and stale-state limitation

The v1 cancellation session was ses_e580ec224ffeT6Bozt8Vaz4HG7, assistant msg_1a7f140fa001gBvb2a2djkSZvD. Before SIGINT, assistant created 1793818968314 and running tool time.start 1793818968931 were present; observation cutoff 1793818970899. The process exited from SIGINT (-2). Its database still contained the identical unfinished assistant and running tool without a terminal/completion marker immediately after exit. This confirms that persisted running state can outlive this process. A historical v1 database alone cannot distinguish this terminated instance from the earlier active snapshot using only those fields. Do not claim process liveness, or unbounded stale-duration certainty, from the marker alone.

The v2 cancellation session was ses_e580eb086ffevVMqheoFc7LjjH, assistant msg_1a7f15883001YTjG75Xzl9MVsW. Initial assistant created was 1793818974348, content-item tool time.ran 1793818974388, and observation cutoff 1793818976354. SIGINT exited 130. Native cleanup changed the tool status to error and added content-item time.completed 1793818976517; assistant gained completed 1793818976524 and finish error. The session's idle timestamp became 1793818976531 and suspended became null. An idle record at that timestamp had outcome interrupted. This is an authoritative terminal shape for this observed graceful interruption, not proof every crash writes it.

Cancellation snapshots remain under /tmp/al64-runtime-qdcg1bx7/{v1,v2}/{active.db,opencode.db} for disposable follow-up.

## What this verifies and what remains open

Verified independently for both pinned runtimes: provider-connected unfinished text activity, running tool shape, subsequent native completion, and SIGINT outcomes. Source schema family was verified from the actual populated tables: v1 session/message/part; v2 session_v2/session_message. Version strings alone were not used to select the family.

Confirmed source-path risks: v1 completed-only accounting omits native unfinished streaming/tool intervals; v2's update-time fallback is not an explicit completion measurement, and pre-change strict running-tool rejection prevents a truthful active snapshot. Their corrected generated/persisted values still require adapter/CLI checks. This record does not itself establish final implementation correctness.

Not executed in this runtime probe: application generate output, --recompute-db source refresh, repeated backfill, root/child preservation, missing-source retention, frontend projection or full make. No built application binary was available when the source probes completed. Deterministic regression work and final verification belong to the implementation/test/review workflow. No synthetic fixture result is represented here as native runtime evidence, and no native source duration arithmetic is represented as an emitted ledger value.

## Live source to generated ledger and source refresh follow-up

The later candidate binary /tmp/al64-agentic-lens became available. Its root, generate and backfill help were inspected. The direct built binary rejects wrapper-only --al-offline/--al-project flags (exit 2), so experiments invoked it from the disposable project's explicit cwd/PWD with local capture state configured. The native database paths remained isolated. Config validation initially rejected missing required anomaly ceiling fields; the fixture configuration was corrected. Generation then required an explicit ticket title, which was supplied. These rejected setup attempts produced no ledger and are not passing application checks.

Each successful experiment used a real native sleep 20 tool, not a replayed historical open snapshot. Commands were run in the disposable project:

- /tmp/al64-agentic-lens generate --source opencode --session-id <native root> --ticket-id AL-64 --ticket-type fix --ticket-title 'AL-64 isolated runtime duration probe' --db-path <isolated native database> --json-output <isolated project>/alspaces/AL-64/agentic_metrics.json
- /tmp/al64-agentic-lens backfill --source opencode --recompute-db --workspaces <isolated project>/alspaces --ticket-id AL-64 --db-path <isolated native database>

Backfill ran while the tool remained blocked, after native completion, and once more unchanged. --force was never used. Durable output is a lean schema-v6 ledger, so this probe reads invocation facts; it does not claim a persisted workflow_efficiency aggregate exists in that file.

### v2.2.0 emitted values

Session ses_e5805e24affetudRSatF5jmxof, initial assistant msg_1a7fa24a6001pwIb0033D4N3jP. Real assistant creation was 2026-11-04T19:12:30.907Z; native tool timing had top-level ran 1793819551149. The final native assistant-message intervals were 1793819550907 through 1793819571307 and 1793819571347 through 1793819571399; existing per-invocation rounding yields 20 seconds.

| Operation | Stored total seconds | Completed seconds | Provisional seconds | Evidence state / cutoff |
| --- | ---: | ---: | ---: | --- |
| Generate during native tool | 2 | 0 | 2 | provisional / 2026-11-04T19:12:33.013270596Z |
| Source refresh during same tool | 3 | 0 | 3 | provisional / 2026-11-04T19:12:34.206751782Z |
| Source refresh after native completion | 20 | 20 | 0 | complete / cutoff absent |
| Unchanged repeated source refresh | 20 | 20 | 0 | complete / cutoff absent |

All four application operations exited 0 and native OpenCode exited 0. Duration basis was opencode:v2-assistant-message. Invocation completed_at was null during activity and 2026-11-04T19:12:51.399000+00:00 after completion. The same single session and invocation were retained across snapshots. The private generated/persisted ledgers and command logs are under /tmp/al64-runtime-o27v2fde/v2/. This is actual live generation/refresh evidence, not a synthetic fixture or arithmetic-only claim.

### Native v1 source-layout gate found during follow-up

The initial native v1.19.5 application attempt was blocked before duration extraction: its fresh database contained the complete session/message/part family plus an empty session_message table, without session_v2. The candidate rejected this as an unsupported mixed layout missing v2.session_v2. This reproduces the empty auxiliary table shape already described by historical AL-58 evidence. No table was deleted or reshaped to mask the failure. Source failure was sent to the orchestrator and implementation owner with /tmp/al64-runtime-o27v2fde/v1/opencode.db and generate_active.log. Subsequent backfills found no ledger because generation had correctly failed; their successful empty scan exit codes are not duration acceptance evidence.

The implementation owner made a narrow detector correction to recognize complete v1 plus a correctly shaped empty auxiliary session_message; populated, malformed and competing layouts remain subject to failure checks. Native v1 live verification against that corrected binary follows below. This report records discovery; the independent tester owns permanent regression assertions.

### v1.19.5 emitted values after source-layout correction

The corrected binary was rebuilt before this retry. Native session ses_e58049704ffe4IkT88o63Un8oW used initial assistant msg_1a7fb6c33001hC1KBhVzv26XEO. Native assistant-message intervals after completion were 1793819634739 through 1793819655820 and 1793819655826 through 1793819655957; existing per-invocation rounding yields 21 seconds. The native database's empty auxiliary session_message table was retained unchanged.

| Operation | Stored total seconds | Completed seconds | Provisional seconds | Evidence state / cutoff |
| --- | ---: | ---: | ---: | --- |
| Generate during native tool | 3 | 0 | 3 | provisional / 2026-11-04T19:13:57.611150217Z |
| Source refresh during same tool | 4 | 0 | 4 | provisional / 2026-11-04T19:13:58.732821151Z |
| Source refresh after native completion | 21 | 21 | 0 | complete / cutoff absent |
| Unchanged repeated source refresh | 21 | 21 | 0 | complete / cutoff absent |

All four application operations exited 0 and native OpenCode exited 0. Duration basis was opencode:v1-assistant-message. Invocation started_at was 2026-11-04T19:13:54.739000+00:00; completed_at remained null while unfinished and became 2026-11-04T19:14:15.957000+00:00 after completion. The same single session and invocation were retained. Private ledgers, native database and logs are under /tmp/al64-runtime-7xgvlk2v/v1/.

For both successful native families, complete and unchanged-repeat sessions arrays were compared structurally and were exactly equal. This proves stable persisted session facts for those unchanged repeats; it does not assert whole-file byte equality because operation metadata may change. No completion contribution was added to its prior provisional snapshot.

These live follow-ups supersede the earlier source-probe-only limitation for native OpenCode generation and --recompute-db refresh. They do not decide the pending historical source-state versus process-liveness policy, nor establish root/child, other-harness, missing-source, Codex, API/UI or full-suite verification; those remain separately owned checks. Probe warnings about unconfigured code classification patterns concern this minimal disposable fixture and do not invalidate the source duration observations.

## Native Codex lifecycle follow-up and fresh-generation limit

A bounded read-only follow-up inspected Codex CLI version 0.166.0 source records for the current ticket. The parent root is 01a1a7d8-b9d2-7383-a6f7-3c7002ba88e6. Native child rollouts can begin with their own session metadata and subsequently replay parent metadata. The first metadata record, filename and source.subagent.thread_spawn.agent_path were checked together; a later inherited session_meta.id was not treated as the child's identity.

The implementation child is owner 01a1a7e7-42a1-78e1-85ac-c58550281d6c, with parent_thread_id 01a1a7d8-b9d2-7383-a6f7-3c7002ba88e6 and agent path /root/implement_al64 in the native first metadata. Its first metadata session_id carries the parent root while id carries the actual child, and forked_from_id names the parent. The parser's existing inherited-replay handling and first identity are therefore material. Its observed native tail included task_started 2026-11-04T19:12:40.137Z, matching task_complete 19:13:46.207Z, then a new task_started 19:28:48.315Z with turn_id 01a1a809-0eb6-71d3-9da0-2efe19904229. The new turn was unfinished when the source was observed. No lifecycle record was fabricated, appended or truncated to create this observation.

The tester child was independently identified from its first native metadata as owner 01a1a7f7-6142-7c31-a6d3-7d1e90305294, agent path /root/test_al64, with the same parent root. Its observed tail contained task_started 2026-11-04T19:09:29.920Z and matching task_complete 19:17:54.318Z, then task_started 19:28:57.562Z with turn_id 01a1a809-32d4-76b0-8013-d788fe46193b. This is native lifecycle shape evidence, not emitted ledger accounting.

A complete current copy of the implementation child's rollout was placed unchanged in a disposable CODEX_HOME, outside Git. A fresh disposable project used the same minimal configuration/isolated capture-state pattern as the successful OpenCode probes. The direct candidate command was generate --source codex --session-id 01a1a7e7-42a1-78e1-85ac-c58550281d6c --ticket-id AL-64 --ticket-type fix --ticket-title 'AL-64 isolated native Codex duration probe' --codex-home <disposable store> --json-output <disposable project>/alspaces/AL-64/agentic_metrics.json.

The command exited 1 before publication with: Capture ownership conflict: selected source owner 01a1a7e7-42a1-78e1-85ac-c58550281d6c: source project directory conflicts with configured project. The original native source correctly retains this repository's cwd, which differs from the disposable project's cwd. This is an expected ownership-boundary rejection, not an observed duration-accounting failure. No ledger was emitted and no active/completion/backfill numerical result is claimed for this Codex native probe. Exact command log and unchanged private copied source are retained under /tmp/al64-codex-native-yhead9fh/.

The orchestrator explicitly directed that no synthetic identity baseline be seeded to bypass ownership validation, no native metadata be rewritten, and no actual ticket telemetry or current-project capture state be generated early. The probe stopped at that boundary. Native Codex source lifecycle observations and the supplied AL-55 diagnosis remain separate from deterministic end-to-end regression evidence. Live native Codex CLI generation, active source refresh, completion refresh and unchanged repeat remain unverified by this runtime probe; they must not be reported as passing based on the OpenCode checks. The single CLI process exited, no background process was created, and live sources were unchanged.
