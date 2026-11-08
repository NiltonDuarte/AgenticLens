# AL-79 native evidence: discovery preparation

Status: native Codex acceptance PASSED for the final affected Linux candidate below; historical gaps and corrections remain revision-scoped. Prepared by the adversarial test engineer using the installed native Codex, independently of candidate implementation. No application/test changes, credential reads, hook/config/auth/trust mutations or bypass flags were performed. The first pass isolated all writable destinations; the follow-up below used the normal user home only for explicitly authorized native operational state writes during inspection.

## Observed interface and environment

- Native version: `codex-cli 0.167.0`, executable resolved to `/home/niltonduarte/.codex/packages/standalone/releases/0.167.0-x86_64-unknown-linux-musl/bin/codex`.
- Disposable evidence root: `/tmp/al79-native-89u4rvf4`. Raw JSON, generated protocol types, probe script and stderr remain there outside Git.
- All process-writable HOME, CODEX_HOME, XDG config/cache/state destinations were explicitly set below that disposable root. Git fixtures comprise an ordinary base checkout, two linked worktrees and nested cwd. Hook command `true` was used only as a discovery marker; it is not an AgenticLens runtime capture test.
- `codex app-server generate-ts --experimental --out <temporary schema>` succeeded. Generated protocol exposes experimental `hooks/list`, accepts `cwds` or `threadId`, and returns hook `sourcePath`, `source`, `enabled`, `currentHash` and `trustStatus`.
- Actual process interface: `codex app-server --stdio`; JSON-line initialize request with clientInfo and experimentalApi capability, initialized notification, then `hooks/list` over base, worktree-one, worktree-one/nested and worktree-two. Processes were terminated after inspection; no model session/turn was started.

## Results

1. With only an untracked worktree `.codex/hooks.json`, all four cwd results contained zero hooks and no hook errors/warnings.
2. Adding base `.codex/hooks.json` still returned zero hooks in this disposable setup. Project trust/config-layer admission was not established, so these results do not prove the original-checkout-versus-worktree sourcePath behavior. Neither fixture contained a `.codex/config.toml`; a future probe must distinguish missing config-layer discovery from project trust before diagnosing the cause. The earlier PO observation is not substituted as fresh evidence.
3. Adding `$CODEX_HOME/hooks.json` produced exactly the marker SessionStart hook for all four cwd values. Native metadata reported `sourcePath: /tmp/al79-native-89u4rvf4/codex-home/hooks.json`, `source: user`, `enabled: true`, `trustStatus: untrusted`, with the same hook key/hash across scopes. This proves native home discovery across those checkouts and nested cwd; it proves no hook execution or capture.
4. Starting normal TUI with `codex --no-daemon --no-alt-screen -C /tmp/al79-native-89u4rvf4/base` under the same isolated destinations displayed the normal Codex welcome/authentication choices (ChatGPT, device code, API key/other provider). The probe exited with Ctrl-C without choosing authentication or starting a turn. This is an isolated-home authentication gap, not evidence that the machine lacks usable existing authentication.
5. Read-only filtering of existing user config found legitimate trusted AgenticLens project entries, including the original checkout. No auth values or hook trust records were read or copied. Starting an app-server against the real CODEX_HOME can write native operational state, so it was not attempted without parent authorization.

## Smallest remaining normal interaction

The native binary advertises `/hooks` to view/manage lifecycle hooks. The generated client request schema exposes hooks/list but no dedicated hook-trust mutation RPC. Do not synthesize trust records or use `--dangerously-bypass-hook-trust`.

Two supported paths remain for the parent:

- Keep isolated acceptance: a human signs in normally in the isolated Codex home, accepts ordinary project trust if prompted, and reviews/trusts the candidate owned hook commands through normal `/hooks`. Auth/trust cannot be fabricated by the probe.
- Reuse the existing authenticated user home: parent explicitly authorizes candidate additive home-hook installation and native operation there, preserving all unrelated hooks and current trust. A human reviews/trusts the new candidate hook commands via normal `/hooks` if their hashes are untrusted/modified. Existing authentication may avoid a new login; no credential extraction is needed. Hook installation still does not establish trust or runtime proof.

Which path is smallest depends on existing hook hashes/trust and authenticated runtime access. Isolated no-auth is not an absolute runtime blocker. Normal human hook review is the only identified interaction that cannot be inferred from approved implementation scope or replaced with an automated trust write.

## Candidate acceptance recipe

1. Finish a fresh pre-fix native discovery probe with an admitted project config layer and ordinary project trust; record sourcePath for base-only and worktree-only hook files. Reuse the actual hooks/list interface, without assuming the earlier zero-hook result identifies its cause.
2. Explicitly install the built candidate user-home integration in the chosen approved environment; inspect all six owned groups, unrelated groups, sourcePath, duplicates and normal trust status. Record exact candidate build/version/interface.
3. Start a real Codex root with explicit AL-79 current-task context. Answer only its exact binding request using the supplied request/session IDs. Request one supported same-task native child, inspect actual rollout metadata, and retain distinct parent/child IDs plus parent linkage and worktree context. A callback's parent ID is not a child ID.
4. Observe candidate worker collection and autonomous publication during activity/checkpoints; compare source-backed root/child facts against the actual ledger and summary. Record output destination, source completeness and active/preferred build. Verify two worktrees/nested context and disabled no-I/O behavior through the appropriate native and deterministic evidence, keeping results distinct.
5. Record remaining runtime access/trust/auth or child-exposure gaps as INCOMPLETE. Do not unblock AL-80 or claim AL-79 completion from the current discovery-only probe.

No root/child source IDs, binding exchange, callback execution, candidate publication or complete native acceptance is claimed by this preparation report.

## Follow-up: admitted normal-home inspection

The parent explicitly authorized normal existing Codex home app-server operational writes needed for hooks/list inspection, while prohibiting hook/config/auth/trust mutation and model turns. Candidate installation/execution remained deferred.

- Added empty `.codex/config.toml` files only in disposable base and worktree-one. Repeating isolated native hooks/list still showed only the global marker. A missing config file alone does not explain the original zero-project-hook result.
- Started a fresh normal-home `codex app-server --stdio` and used the same initialize/hooks/list interface, then `config/read` with `includeLayers: true`. Inspected actual AgenticLens base, this worktree, its backend subdirectory and disposable checkouts. Filtered output retained only metadata/layer origins; no credential values or full user configuration were read or printed. Processes were terminated after inspection.
- Normal-home native hooks/list reports six existing user hooks, all enabled and all `trustStatus: untrusted`, from `/home/niltonduarte/.codex/hooks.json`. This is an actual existing-home hook-trust gap; it is not merely the isolated-home authentication condition. No callback execution is implied.
- Native config/read reports enabled project layers for `/home/niltonduarte/workspace/projects/agenticlens/AgenticLens/.codex` and this worktree's `.codex`. File existence checks confirm neither existing checkout has `.codex/hooks.json` or `.codex/config.toml`, so read-only inspection there cannot produce project-hook sourcePath evidence.
- For `/tmp/al79-native-89u4rvf4/base`, the native project layer gives this explicit disabledReason: `To load project-local config, hooks, and exec policies, add /tmp/al79-native-89u4rvf4/base as a trusted project in /home/niltonduarte/.codex/config.toml.` This establishes the disposable project admission gap without guessing from hook counts. Trust was not written or overridden.

Raw filtered follow-up data and the probe remain under the disposable evidence root: `real-home-hooks-filtered.json`, `real-home-layers-3.json`, `real-home-layers-4.json`, `real-home-layers-5.json`, and `probe-real-home.py`.

Smallest remaining pre-fix action under the read-only existing-checkout constraint: a human admits the disposable base as a trusted project using the normal Codex project trust flow. The app-server diagnostic names the exact project and user config. Once legitimately admitted, the existing disposable base/worktree marker files can be inspected for sourcePath without starting a model turn or executing hooks. Separately, candidate owned user hooks require normal `/hooks` human review/trust after installation if untrusted; old-hook trust would not approve changed candidate hashes.

An alternative would be parent authorization for temporary project marker files in an already legitimately trusted checkout and a linked worktree, preserving/restoring prior files. That exceeds this follow-up's explicit read-only existing-checkout boundary and was not attempted. No fresh pre-fix project sourcePath proof is claimed yet.

## Completed pre-fix sourcePath proof in legitimately trusted checkouts

The parent subsequently authorized only temporary `.codex/hooks.json` discovery markers in this AL-79 checkout and its legitimately trusted original base. Each path was first confirmed absent with lstat (including symlinks); `.codex` parents were confirmed normal directories. Marker creation used exclusive-create mode, command `true`, and a single SessionStart group. Only hooks/list was invoked; no session/model turn or hook execution occurred. No user config or trust records changed. A finally block removed only byte-identical owned markers and would preserve an unexpectedly edited file.

Actual native interface remained Codex 0.167.0 experimental stdio hooks/list. Inspected cwd values were the original base, AL-79 worktree and AL-79 `backend` subdirectory:

- Base cwd: `/home/niltonduarte/workspace/projects/agenticlens/AgenticLens`.
- Worktree cwd: `/home/niltonduarte/.t3/worktrees/AgenticLens/t3-350b0421`.
- Nested cwd: `/home/niltonduarte/.t3/worktrees/AgenticLens/t3-350b0421/backend`.

| Marker files present | Native project hooks at all three cwd values | Native sourcePath |
| --- | --- | --- |
| Worktree marker only | Zero; only six unchanged user hooks appeared | No project sourcePath |
| Both base and worktree markers | One SessionStart project hook, enabled, trustStatus modified | `/home/niltonduarte/workspace/projects/agenticlens/AgenticLens/.codex/hooks.json` |
| Base marker only | Same one SessionStart project hook, enabled, trustStatus modified | `/home/niltonduarte/workspace/projects/agenticlens/AgenticLens/.codex/hooks.json` |

No inspected cwd returned the worktree marker path as a project sourcePath. These native metadata observations reproduce the pre-fix dependency on the original checkout's project hook file. The worktree-only untracked hook is invisible to this interface; home hooks remain visible from all scopes. This proof supersedes the earlier pre-fix discovery gap. It proves discovery behavior, not callback execution or trust approval.

Cleanup confirmed both temporary `.codex/hooks.json` paths absent, including symlinks. No created directory was removed unless empty; existing directories/files and all unrelated application edits were preserved. The six existing user hooks retained their source, enablement and untrusted statuses throughout all three inspections. Raw filtered native results and safe probe script remain outside Git under `trusted-worktree-only.json`, `trusted-base-and-worktree.json`, `trusted-base-only.json` and `probe-trusted-markers.py` in the disposable evidence root.

Remaining native acceptance gaps are candidate global installation/discovery, normal human trust of candidate hashes, exact root binding, a source-proven child and source-backed automatic publication. The pre-fix sourcePath proof now needs no additional human action; candidate hook trust remains a later human gate.

## Candidate isolated installation and native discovery

The parent refroze production after a worker correction and rebuilt the canonical Linux amd64 candidate with Makefile build metadata. Evidence below is scoped to that binary; later candidate/source changes require an impact assessment and affected re-verification.

- Git HEAD: `4df45967c801824500ac2ea39a8f52fa9cbdcf66`, with uncommitted AL-79 implementation.
- Actual candidate version JSON: version `v3.0.0-rc.133+57a3503-1-g4df45967-dirty`, commit `4df45967`, build_date `2026-11-08T15:55:37Z`.
- Candidate path: `/home/niltonduarte/.t3/worktrees/AgenticLens/t3-350b0421/bin/agentic-lens-linux-amd64`.
- Candidate and installed dispatcher SHA-256: `0f308b80be2a0deb3ce72a39ca5ab692b7a89304413a55c9930f5d345cd257e7`.
- Source diff fingerprint at setup: `cb00b4fab202eb3fd633adeb746e8958c50a3bb79a2b68b5897a667a894e2204`. Input was git diff HEAD for backend/scripts/.agents/docs/README.md, followed by sorted untracked files under backend/scripts/.agents/docs, each path, NUL, file bytes, NUL. This identifies the inspected working candidate; it is not a claim that the fingerprint alone proves binary/source provenance.

Fresh disposable environment: `/tmp/al79-candidate-f5tgsdwu`. HOME, CODEX_HOME, XDG config/cache/state/data, install root and capture state root are all explicit children of that root. Ordinary Git base and two linked worktrees share a committed self-contained shared YAML. The shared file includes the required exact fallback version `v3.0.0-rc.131+03f5b05-2-g57a3503e`, `binary_dir` pointing to the host bin directory, custom `evidence-records` workspace, AL prefix and sample-src classification. One ignored base-local file explicitly enables capture, selects codex/default source, sets the isolated source home/state/install root. Worktree-local files are absent, exercising whole-file base fallback. No actual user-home changes or credentials were read/copied.

Candidate commands used the absolute built binary: `capture install-user --repo /tmp/al79-candidate-f5tgsdwu/base --binary <absolute candidate> --dry-run`, followed by the same command without dry-run. Both exited zero. Dry-run created no paths (excluding its explicit external evidence output). Apply installed only the native dispatcher/checksum, additive user hooks and machine-wide ownership metadata. A pre-existing unrelated SessionStart marker group was preserved byte-for-byte structurally; no per-worktree manifest/install was performed.

Actual Codex 0.167.0 stdio hooks/list inspected:

- `/tmp/al79-candidate-f5tgsdwu/base`.
- `/tmp/al79-candidate-f5tgsdwu/worktree-one`.
- `/tmp/al79-candidate-f5tgsdwu/worktree-one/nested`.
- `/tmp/al79-candidate-f5tgsdwu/worktree-two`.

Every scope returned seven hooks: six owned events (sessionStart, sessionEnd, userPromptSubmit, subagentStart, subagentStop, stop) plus the unrelated fixture hook. Owned metadata has source `user`, sourcePath `/tmp/al79-candidate-f5tgsdwu/codex-home/hooks.json`, enabled true and trustStatus untrusted. No hook warnings/errors appeared. All six callbacks use only `'/tmp/al79-candidate-f5tgsdwu/native-cache/dispatcher' 'dispatch' 'codex-hook'`; no project/source/state path is frozen into them.

Candidate capture status exited zero for all four cwd values, reported the active worktree identity, preferred actual development build and selected base-local origin for linked/nested scopes. It reported verification_pending, no registrations/callback/binding/source/publication observation and no installation issues. The capture-state root remained absent after installation, discovery and status. No worktree-local config was created. Native execution/publication is still explicitly INCOMPLETE.

Raw evidence is outside Git in the candidate disposable root: `acceptance-manifest.json`, `install-dry-run.json`, `install-applied.json`, `candidate-hooks-list.json`, `candidate-discovery-summary.json`, `candidate-status.json`, and `inspect.py`.

### Concrete normal human acceptance step

Run `sh /tmp/al79-candidate-f5tgsdwu/open-codex.sh` in an interactive terminal. This launcher sets every isolated HOME/CODEX_HOME/XDG destination explicitly and opens normal Codex in worktree-one without a shared daemon. Complete ordinary authentication if prompted, accept ordinary project trust if prompted, then use `/hooks` to review and trust the six `AgenticLens native user capture v1` commands. Do not trust unrelated groups merely to complete this probe. Do not use hook-trust bypass flags or write trust receipts manually. Candidate hashes need their own normal review; trust for other commands/builds is insufficient.

No model prompt was submitted by this preparation pass. After normal authentication/trust is established, the parent can run the approved AL-79 root/child binding/publication scenario and record actual source IDs. The isolated base/worktree files and installed candidate remain available for that step. Existing real-home legacy conflict remains untouched.

## Genuine authenticated native root/child capture and automatic publication

The user confirmed normal authentication and trust of all six candidate hooks. Before starting this probe, fresh actual Codex hooks/list responses at base/two worktrees/nested cwd confirmed every candidate event `trustStatus: trusted`. No trust receipts were synthesized and no trust bypass was used. The current installed/retained executable SHA-256 was `1731a3722b49e74fd35ca8eb28c5fd3f3bd97542d6ffabaaf7be9ebece1caca0`, with actual version `v3.0.0-rc.133+57a3503-1-g4df45967-dirty`, commit `4df45967`, build_date `2026-11-08T16:02:46Z`.

Started a new native `codex app-server --stdio` under the same isolated paths, preserving the human TUI and its independent context. A normal `thread/start` plus `turn/start` requested explicit task AL-79, feat, title Native user-level Codex capture, direct exact-session binding, one genuine same-task native child reading fixture text, native wait and final pwd checkpoint. Workspace-write permission was limited to the disposable acceptance root. No manual generation, reconciliation or publication command was submitted. Root completed successfully in about 29 seconds; child completed its read in about eight seconds. Actual root model was gpt-6.1-sol with effort unset in start response; native child activity reported gpt-6.1-sol/low.

| Evidence | Actual native value |
| --- | --- |
| Root ID | `01a1bbe8-0244-70c3-a80a-4f7df5c0eec2` |
| Root turn ID | `01a1bbe8-0272-7853-8f24-beb6bc9972fc` |
| Child ID | `01a1bbe8-3c56-7e93-ae52-a64951c695e7` |
| Child role/path | fixture_read; native /root/fixture_read |
| Actual cwd, both source metadata records | `/tmp/al79-candidate-f5tgsdwu/worktree-one` |
| Source store | `/tmp/al79-candidate-f5tgsdwu/codex-home` |
| Root binding | explicit_agent_response; AL-79/feat/title supplied directly; binding command exit zero |
| Child binding | source_parent; same AL-79; exact parent ID retained |
| Latest publication observed | generation 6, acknowledged 6; diagnostics sequence 19 committed at 2026-11-08T16:05:55.15695255Z |

Root binding used the hook-supplied actual native session/request token without another confirmation; token remains only in private raw evidence. The child did not invent a binding request or run a manual binding command. Its actual source metadata contains id equal to the child ID, parent_thread_id equal to the root ID, source.subagent.thread_spawn parent/depth/path, and forked_from_id. Its session_id field shares the root's session tree identity, so that field was not relabeled as its distinct child ID. Both durable registrations reached verified project proof and current source-snapshot state.

The worker automatically published `worktree-one/evidence-records/AL-79/agentic_metrics.json` and SUMMARY.md. The ledger contains exactly two distinct session owners, the root delegation names the child's native ID, and the child parent_id names the root. Both source files have actual task-start/task-complete and cumulative token evidence. Assertions comparing source token_count total_token_usage to persisted invocation values passed: input is source input minus cached input, cache_read is cached input, and output/reasoning match directly. Observed values:

| Owner | Source input | Source cached input | Ledger uncached input | Output | Reasoning |
| --- | --- | --- | --- | --- | --- |
| Root | 82,621 | 73,472 | 9,149 | 519 | 0 |
| Child | 37,788 | 35,072 | 2,716 | 129 | 14 |

Cost remains explicitly unavailable with null cost_usd and unavailable cost_evidence. Re-reading after later worker refresh showed the same two unique sessions, one invocation each and unchanged token totals. Publication diagnostics name both successfully refreshed owners with exact source checkpoints, no pending children/unresolved parent links, and no manual publisher involvement. The retained worker executable path includes the observed SHA and status reports its actual active build.

The human TUI was not interrupted, bound, driven or inspected for credentials. Only the two probe rollout files existed in this isolated sessions directory at inspection; no source-backed human root ID or unassigned human registration was available to assert. Its open context remained untouched. This is preservation of the existing UI context, not a fabricated claim that a separate human source session was captured.

### Actual defects discovered in the live run

- Status observation gap: despite durable callback/binding timestamps, current source registrations and committed publication diagnostics, top-level capture status reported verification_pending and false for all four observation flags, with installation empty. Sent actual raw evidence to the implementation owner; no synthetic certification is requested. Native trust is established independently by hooks/list. The coder reported a read-only aggregation fix; actual rebuilt-candidate status verification is pending.
- Child SubagentStart callback failure: actual native hook/completed for child turn `01a1bbe8-3c84-7f43-b90f-524a2ca977e7` reports user sourcePath, status failed, duration 13 ms, error `hook exited with code 1` at 2026-11-08T16:05:22.681Z. The app-server stderr contained no detailed callback error. Root SessionStart/UserPromptSubmit/Stop and child SubagentStop completed normally. Child recovery, inherited binding and automatic publication subsequently succeeded, so this failure did not erase source facts, but clean native callback coverage is not yet proven. Sent the exact event/source/log paths to the coder and parent for investigation. A source-context race or mismatch is a hypothesis, not an established cause.

Raw transcripts remain outside Git: live-native-events.jsonl, live-native-stderr.log, live-root-start.json, live-root-turn-completed.json, live-root-read.json, live-capture-status.json, trusted-hooks-metadata.json, and sanitized live-source-facts-filtered.json under the isolated candidate root. Source rollout basenames identify the actual root and child IDs above. No further model/session probe was run after discovering these issues; affected verification must follow the corrections.

Verdict at that checkpoint: root/child source-backed binding and autonomous publication PROVEN for that candidate; native coverage was INCOMPLETE pending the recorded status/callback issues and revision-aware verification. This does not claim full AL-79 acceptance or AL-80 readiness.

Bounded callback diagnosis inspected the isolated native logs_2.sqlite read-only around the actual 16:05:22 failure. No hook-specific tracing target, stdin payload/transcript_path or detailed dispatcher/context-mismatch error was recorded; app-server stderr likewise lacked detail. The exact failed run ID is `subagent-start:4:/tmp/al79-candidate-f5tgsdwu/codex-home/hooks.json`, emitted 1794153922667 to 1794153922681 ms. Root/child first metadata records were 22,574/23,003 bytes, below SessionContext's 64 KiB read limit. This narrows hypotheses without claiming an unobserved callback payload. No model/session rerun or event replay was performed during diagnosis.

## Verified correction: recorded stage status and active/preferred builds

Read-only `capture status --repo /tmp/al79-candidate-f5tgsdwu/worktree-one` using the rebuilt canonical host binary exited zero with empty stderr. Actual rebuilt identity: version `v3.0.0-rc.133+57a3503-1-g4df45967-dirty`, commit `4df45967`, build_date `2026-11-08T16:08:28Z`; binary SHA-256 `52c0a68688a2662dd8dcc2897f35ff8576604e65f2933cd5fff4ae08a4c6b1fe`. No dispatcher reinstall, trust change, model/session turn, manual generation/reconciliation or source/state removal occurred.

The corrected output now reports verification state verified and callback_observed, binding_accepted, source_accepted and publication_accepted all true. Assertions checked these four actual flags and that registrations still contain exactly the original real root/child IDs, preserving explicit root binding and source-parent child binding. Status correctly limits its claim to recorded sessions and explicitly excludes harness trust/unobserved sessions from those stage observations. This resolves the observed status aggregation defect for the rebuilt candidate.

Active/preferred separation is truthful in the same output: preferred executable remains the mutable host build path with build_date 16:08:28Z, while the recorded worker retains executable SHA `1731a3722b49e74fd35ca8eb28c5fd3f3bd97542d6ffabaaf7be9ebece1caca0` and its admitted build_date 16:02:46Z. Generation/acknowledgement remain 6/6. Worker state is recorded_running with the explicit read-only limitation that heartbeat alone does not verify live process ownership. This observation proves the displayed identity distinction; it does not alone prove a newly launched publication child after this rebuild.

Corrected raw status is outside Git at `/tmp/al79-candidate-f5tgsdwu/live-capture-status-corrected.json`. The original failure remains retained against its earlier build. Remaining native acceptance issue: the actual SubagentStart exit-one event still needs diagnosis, correction and affected native verification; the status correction does not refute that separate callback failure.

## Confirmed callback root cause through one native diagnostic probe

To resolve the missing stdin evidence, the parent explicitly authorized one temporary nonproduction diagnostic wrapper at the isolated dispatcher command path. Ownership was checked first: native bytes matched receipt SHA `1731a3722b49e74fd35ca8eb28c5fd3f3bd97542d6ffabaaf7be9ebece1caca0`, mode 0755. Exact native bytes/mode were backed up privately. The temporary shell wrapper recorded actual stdin and stderr, executed that exact native backup with original arguments/stdin/stdout, forwarded stderr, and made no hook/trust/auth/config changes. No payload was synthesized or replayed. Installer activity was excluded while instrumented.

One bounded real native root/child turn ran in worktree-one (already started before an optional suggestion to use worktree-two, so no restart or extra probe was added). Root `01a1bbef-0511-75c2-a30f-01b5093dbdf9`, turn `01a1bbef-0545-7702-bc08-15e34e6211b7`, bound explicit AL-79 and spawned one read-only fixture child `01a1bbef-3c81-7392-a8d1-925c06dc0d04`. Both completed; the root turn took about 31.8 seconds. No manual generation/reconciliation/publication ran.

Actual SubagentStart stdin:

- session_id: `01a1bbef-0511-75c2-a30f-01b5093dbdf9` (parent).
- agent_id: `01a1bbef-3c81-7392-a8d1-925c06dc0d04` (child).
- transcript_path: `/tmp/al79-candidate-f5tgsdwu/codex-home/sessions/2026/11/08/rollout-2026-11-08T17-13-00-01a1bbef-3c81-7392-a8d1-925c06dc0d04.jsonl` (child transcript).
- cwd: `/tmp/al79-candidate-f5tgsdwu/worktree-one`.
- agent_transcript_path: absent.

Exact stderr: `AgenticLens dispatcher: native session context mismatch: callback 01a1bbef-0511-75c2-a30f-01b5093dbdf9, source 01a1bbef-3c81-7392-a8d1-925c06dc0d04`.

The child's real source metadata independently records id equal to the child, session_id equal to the root tree, parent_thread_id equal to that root, and source.subagent.thread_spawn with matching parent/depth/path. Thus strict callback-session equality rejects a genuine source-proven parent/child event. The correction must validate that relationship while preserving parent callback identity; changing the callback ID to the child would violate the contract.

Actual SubagentStop differs: session_id and transcript_path both identify the parent, while agent_id and agent_transcript_path identify the child. Its stderr was empty. The four ordinary root lifecycle callbacks likewise had empty stderr. These observations explain why child-start failed while stop/worker recovery still succeeded.

In finally, the exact original native bytes/mode were restored atomically. Post-restoration SHA matched the receipt, mode remained 0755, and both hook JSON and installation receipt were byte-identical to their pre-probe values. No diagnostic wrapper remains at the installed path. Private evidence/backup remains outside Git under `/tmp/al79-candidate-f5tgsdwu/native-callback-diagnostic-1kxi0955`; failed stdin is `input.qF1PpQI2.json`, matching stderr is `input.qF1PpQI2.stderr`, stop stdin is `input.VvgoFd43.json`, and `filtered-callback-results.json` summarizes all six actual callbacks. Native stream/readback logs are `diagnostic-native-events.jsonl`, `diagnostic-native-stderr.log`, `diagnostic-root-start.json`, `diagnostic-root-turn-completed.json` and `diagnostic-root-read.json` in the candidate fixture root.

Both parent and coder received this concrete reproduction. Callback correctness remains open until the implementation correction and affected normal native re-verification; this diagnostic is discovery evidence, not a passing re-check.

## Final affected native re-check: PASSED

After the actual-shape regression and callback correction, the parent rebuilt the canonical host candidate and installed the uninstrumented native dispatcher at the same owned path. Both files were independently hash-checked against SHA-256 `38281506d34e1172100d8f2ff03fd49b4941c2b78bb1d3bdf295288a05c6d1d2`. Actual version: `v3.0.0-rc.133+57a3503-1-g4df45967-dirty`, commit `4df45967`, build_date `2026-11-08T16:16:46Z`. Existing normally trusted hook commands remained unchanged. No diagnostic wrapper, source payload replay, manual generate/reconcile or trust bypass was used.

One clean native root/child turn ran in worktree-two, preserving the other worktree's source/ledger history. Explicit current task AL-79/feat/title supplied the root binding without another confirmation. Root `01a1bbf4-1977-79c1-a8dc-5c8b7da3e5d3`, turn `01a1bbf4-19d0-7401-87ea-90c5752f2067`, spawned exactly one real native child `01a1bbf4-4f75-7241-842d-32abc66c0120`, waited for the read-only fixture result and executed the final pwd checkpoint. Root turn completed in 26.8 seconds. Both actual metadata records identify cwd `/tmp/al79-candidate-f5tgsdwu/worktree-two`; child id and parent_thread_id/source.subagent.thread_spawn prove the distinct child and exact parent relationship.

All observed owned native hook runs completed normally:

| Event | Actual completed duration |
| --- | --- |
| SessionStart | 151 ms |
| UserPromptSubmit | 133 ms |
| SubagentStart | 161 ms |
| SubagentStop | 185 ms |
| Stop | 172 ms |

The corrected SubagentStart returned context explicitly saying the callback identifies the parent and forbidding binding with the parent's ID as the child's own identity. The child performed no manual binding; authoritative source-parent inheritance bound it to AL-79. The durable root registration subsequently records SessionEnd, proving its shutdown callback reached capture; app-server termination does not expose that final callback as a retained hook/completed event, so no unobserved SessionEnd run status is invented.

Read-only final status reports verified project proof, current root/child snapshots, explicit_agent_response for root, source_parent for child and all four recorded stage flags true. The automatic worker published ledger plus SUMMARY.md under `worktree-two/evidence-records/AL-79/`, generation 7 acknowledged 7. Exactly the two worktree-two owners appear once each, with one invocation each, correct delegation/parent ID and source-matched token totals:

| Owner | Source input | Cached input | Ledger uncached input | Output | Reasoning |
| --- | --- | --- | --- | --- | --- |
| Root | 82,674 | 73,728 | 8,946 | 530 | 0 |
| Child | 37,936 | 35,328 | 2,608 | 110 | 0 |

Assertions compare actual source token_count totals to ledger input/cache/output/reasoning and require source-proven parent linkage. Costs remain unavailable rather than measured zero. Final preferred and admitted worker identities both name the corrected 16:16:46Z build; retained worker path contains the verified SHA. The earlier worktree-one worker/source/history remains separate. A saved before/after comparison confirmed all four prior worktree-one session IDs, parent IDs and complete invocation values preserved exactly, with no worktree-two owner redirected into that ledger.

Scoped verdict: both native-discovered defects are fixed and verified on the affected candidate. Genuine normal-trust global hook discovery, explicit root binding, source-proven child inheritance, checkpoint callbacks, automatic source-backed publication and separate linked-worktree destinations passed on Codex 0.167.0/Linux amd64. Native all-event proof has the explicit SessionEnd visibility limit above. This does not replace independent code review, integrated make, deterministic disabled/error/portable-history coverage, other-platform verification or AL-80 OpenCode acceptance. The open human TUI/auth/trust was preserved; no captured human source session is asserted.

Final raw evidence stays outside Git under the isolated candidate root: `final-candidate-version.json`, `final-native-events.jsonl`, `final-root-start.json`, `final-root-turn-completed.json`, `final-root-read.json`, `final-capture-status.json`, `final-source-facts-filtered.json`, `final-hook-statuses.json` and `worktree-one-ledger-before-final.json`. The source-backed final assertions are in `/tmp/al79-final-evidence-check.py`. Earlier failing evidence is retained against its earlier binary, preserving discovery versus correction/verification.

## C3 installation/read-only carry-forward check

The parent built the frozen C3 source (diff SHA256baf424ef84db886d4d8d1de9f5c271c3d4955c554714d400df4397a5d5f42e0c) with canonical `make build PLATFORMS=linux/amd64`. Host binary SHA2565ed6315774a0d25c50cc71a1e700a0b6ff7839fb6c50f1f12854b943d3d402ce, versionv3.0.0-rc.133+57a3503-1-g4df45967-dirty, commit4df45967, build2026-11-08T17:34:18Z. Explicit owned installation into the same isolated candidate home succeeded, including upgrade from the earlier home-owned metadata to the current root ownership model. Hook configuration bytes were unchanged; no trust/auth values were edited or copied.

Both worktree status commands exited0. Worktree-one's four-owner ledger and worktree-two's two-owner ledger remained byte-identical through installation/read-only inspection. Full filtered result is outside Git at /tmp/al79-candidate-f5tgsdwu/c3-readonly-carryforward.json. No new model turn, callback, manual generation or reconciliation was used. This verifies C3 installation/read compatibility and preservation of recorded facts; it does not claim a new native source observation. Earlier normal-trust root/child execution and source-token comparisons remain their own revision-scoped evidence. Subsequent C4 corrections require explicit impact assessment before carrying this check forward.
