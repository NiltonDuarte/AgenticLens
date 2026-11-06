# AL-64 iteration 03 native Codex refresh evidence

Runtime verification owner: test engineer, gpt-6.1-sol/high (inherited execution settings). Scope: authentic recorded portable history read by the numeric-only candidate and refreshed from the native Codex source. No application/test changes, real-ticket telemetry generation or liveness checks are part of this probe.

## Baseline and isolation

Prepared a byte-identical copy of the actual recorded alspaces/AL-64/agentic_metrics.json in a disposable configured Git project under /tmp/al64-iter03-native-refresh-2umb5n6b/project. The original remained untouched. Baseline SHA256 was 80f7eb327d3ec55e85f391a6aee02b1e9120079ece983574444406d3e54a3ec2, schema v6, with 19 native Codex owners: root 01a1a7d8-b9d2-7383-a6f7-3c7002ba88e6 and 18 recorded children. The root carried active_duration_seconds 5953 and legacy duration_evidence with completed_seconds 5603, provisional_seconds 350, cutoff 2026-11-05T10:36:46.850981184Z and basis codex:lifecycle-turn. All recorded children had zero provisional contribution.

This baseline is genuine recorded history rather than a synthetic owner seed. The earlier fresh copied-source generation failed the expected project ownership boundary; that result is preserved in iteration 01. This probe tests whether the supported portable-record refresh path can use authentic existing history from a different project directory. Any ownership rejection remains a limit; no metadata or identity bypass is permitted.

The disposable project has its own configuration and capture state, explicit cwd/PWD, and a copied ledger. Source selection uses the actual Codex store read-only; raw transcripts remain outside Git. Existing incomplete T3 source discovery is not repaired by guessing owners or attaching sessions. Only identity proven by native source metadata can affect refreshed facts. The original AL-55 and AL-64 ledgers are never experimental outputs.

## Independent native lifecycle observations

Read-only inspection of the root's first metadata established id and session_id 01a1a7d8-b9d2-7383-a6f7-3c7002ba88e6, no parent, cli_version 0.166.0 and source vscode. The native root rollout is /home/niltonduarte/.codex/sessions/2026/11/04/rollout-2026-11-04T19-36-00-01a1a7d8-b9d2-7383-a6f7-3c7002ba88e6.jsonl. Later child replay metadata is not used as child identity.

The turn open at the stored cutoff was 01a1ab43-01e5-71f3-bd8e-82b863d5bf83. Native task_started was 2026-11-05T10:30:57.773Z and matching task_complete was 2026-11-05T10:40:53.859Z. Direct timestamp subtraction gives a completed interval of 596.086 seconds; the earlier cutoff was 349.077981 seconds after that start. The stored provisional integer need not equal independently rounding that one interval because component totals retain the existing invocation-level rounding remainder. These are independent event arithmetic observations, not a duplicate parser used as an oracle for all historical totals.

The latest observed root lifecycle before binary readiness included task_started 2026-11-05T18:53:00.422Z, turn_aborted 18:53:12.295Z for that same turn, then task_started 18:56:17.659Z for turn 01a1ad11-a6ee-7870-a861-ad289609fda4. The current open interval is a provisional source estimate under the approved iteration 03 policy; no process-liveness inference, poke or extra status protocol is introduced.

## Candidate execution

The implementation owner confirmed /tmp/al64-iteration03 as an intermediate numeric-only candidate. Root and backfill help were inspected before testing; no new reasoning options were present. The initial executable SHA256 was 4363c35137b89cffe970140fcc21ee4852af098e0f726cb3704008d3e775fecc. This identifies the tested intermediate executable, not a final reviewed revision.


### Initial compatibility failure

From the disposable project, the command was /tmp/al64-iteration03 backfill --source codex --recompute-db --workspaces /tmp/al64-iter03-native-refresh-2umb5n6b/project/alspaces --ticket-id AL-64 --codex-home /home/niltonduarte/.codex. No --force, source event modification or seeded identity was used.

The first operation ran 2026-11-05T19:04:42.932148Z through 19:04:42.957333Z; an unchanged repeat ran 19:04:44.159066Z through 19:04:44.174394Z. Both processes exited 0, but both explicitly reported Updated: 0 with a per-workspace error: validating current ledger before backfill: decoding current ledger: decoding current schema: json: unknown field "duration_evidence". Neither source refresh nor canonical numeric writing succeeded. The 19-owner legacy file remained unchanged, including root total 5953 and its recorded provisional contribution 350. Absence of a flat scalar in this untouched old file is not measured provisional zero.

The failure is before source extraction, not the earlier project-ownership limit. It was reported immediately to the orchestrator and implementation owner. The implementation owner confirmed a current-schema migration normalization bypass and is rebuilding a correction; the initial failure remains part of this evidence. Command logs and unchanged before/after copies are private under the disposable probe directory. A successful process exit alone is not treated as completed extraction.

### Corrected candidate: representation-only rewrite

The implementation owner rebuilt /tmp/al64-iteration03 after correcting current-schema input normalization. The tested corrected executable SHA256 was 40ccdaf97bf7127b2fd9305bb8d3a95bfda22c71af3d92421c11633a713b8b59. This remains an intermediate implementation candidate; final revision review and integrated verification are separate.

First ran the same backfill command with --force in place of --recompute-db. Recompute Source Facts was explicitly false. The operation ran 2026-11-05T19:06:45.352322Z through 19:06:45.383531Z, exited 0 and reported Scanned: 1, Updated: 1, Completed: 1. All 19 recorded owner identities, parent links, source labels, invocation counts and numeric facts remained intact. Root total stayed 5953 with flat provisional_seconds 350. Closed children omitted the optional zero-valued scalar.

The entire resulting sessions array was structurally equal to the authentic baseline after only this independently specified wire projection: remove each legacy duration_evidence object, and carry its provisional_seconds into the invocation only when positive. No additional session/invocation difference existed. Thus this stage verifies pure representation normalization independently of source growth, rather than treating later recomputation changes as preservation evidence.

Canonical output contained no duration_evidence, duration_reasoning, completed_seconds, duration_coverage, provisional_invocations, unknown_invocations or cutoffs fields. Every present provisional_seconds was an integer greater than zero and no greater than the recorded active_duration_seconds. This was a whole-output recursive audit plus per-invocation scalar checks, not an assertion based only on a process exit code.

### Corrected candidate: authentic source refresh and repeat

Next ran backfill --source codex --recompute-db against the same copied history and the actual native source store, followed by a second source recomputation. Neither used --force. Both exited 0 and explicitly reported Updated: 1, Completed: 1, Incomplete: 0. No project-ownership rejection occurred for this authentic portable-history path.

| Operation | Root total seconds | Root provisional seconds | Derived nonprovisional difference | Owner count |
| --- | ---: | ---: | ---: | ---: |
| Authentic legacy baseline | 5953 | 350 | 5603 | 19 |
| Force-only canonical rewrite | 5953 | 350 | 5603 | 19 |
| Native source refresh 1 | 7784 | 629 | 7155 | 20 |
| Native source refresh 2 | 7791 | 636 | 7155 | 20 |

The nonprovisional difference is arithmetic on total minus provisional; it is not a newly serialized completed_seconds field or a claim that absent provisional values prove complete source coverage. Root invocation count remained one. The previously open morning turn had the actual completion recorded above; later completed root turns and the current open evening turn were source facts, not fabricated replay records. Source refresh replaced the root snapshot rather than adding the old total or provisional contribution to it.

Refresh 1 ran 2026-11-05T19:06:46.585107Z through 19:06:52.770926Z. Refresh 2 ran 19:06:53.977004Z through 19:06:56.284387Z. The current root turn began 18:56:17.659Z; direct subtraction to the respective operation starts is approximately 628.926 and 636.318 seconds, consistent with the stored 629 and 636 provisional values and transient operation cutoff semantics. The scalar representation intentionally does not persist a new cutoff object. This arithmetic is an independent check of the current interval, not a reimplementation of all extraction/rounding logic.

The owner increase was legitimate native discovery of 01a1ad14-a2c4-7d83-b609-e7404fc60e23. Its first native metadata identifies that exact own id, parent_thread_id 01a1a7d8-b9d2-7383-a6f7-3c7002ba88e6, cli_version 0.166.0 and source agent path /root/test_lean_duration. Its native task_started was 2026-11-05T18:59:33.249Z. This was a proven child of the recorded root, not an inferred association for missing T3 sessions. Both refreshes retained all original 19 owners, had exactly 20 unique session IDs and one invocation per owner; the second added no duplicate owner or invocation.

Of the original 18 children, 15 unchanged numeric snapshots stayed unchanged on the first source refresh. Three legitimately advanced: the coder and runtime engineer had new open source turns, while the planner increased 392 to 523 seconds after a native completed turn from 2026-11-05T18:26:41.857Z to 18:28:52.494Z (130.637 seconds, rounding to 131). Child identities were established from first native metadata, not inherited root replay metadata.

On the second refresh, the full session objects for all 16 closed owners were structurally identical to refresh 1. Only four source-open owners changed their duration numerics:

| Owner | Refresh 1 total / provisional | Refresh 2 total / provisional | Stable nonprovisional difference |
| --- | ---: | ---: | ---: |
| Root 01a1a7d8-b9d2-7383-a6f7-3c7002ba88e6 | 7784 / 629 | 7791 / 636 | 7155 |
| Coder 01a1a7e7-42a1-78e1-85ac-c58550281d6c | 2180 / 583 | 2188 / 591 | 1597 |
| Runtime engineer 01a1a7ed-cac0-7930-86d6-0a006e4712fa | 1177 / 310 | 1185 / 318 | 867 |
| New tester 01a1ad14-a2c4-7d83-b609-e7404fc60e23 | 433 / 433 | 441 / 441 | 0 |

The native new open starts were coder 18:57:03.312Z, runtime engineer 19:01:36.801Z and new tester 18:59:33.249Z. The four advancing source estimates are expected when a new cutoff is supplied; this is not an unchanged fixed-cutoff experiment. Completed portions did not accumulate again. All three corrected outputs passed the no-retired-fields and positive-only scalar audits.

## Result and limits

Verified against the authentic real-source history: old rich v6 input normalization, full-session numeric/identity preservation during force-only canonicalization, native Codex source refresh after the formerly provisional turn completed, new open source estimates, legitimate child discovery and repeated source refresh without duplicate completed duration or sessions. Source-state estimates use the accepted iteration 03 semantics; no liveness protocol or additional session work was introduced.

The initial compatibility rejection was a real failed check and remains recorded despite its correction. All corrected command logs, byte-preserved baseline and post-operation ledgers remain outside Git under /tmp/al64-iter03-native-refresh-2umb5n6b/. Native IDs, versions and selected event timing evidence are retained here without raw reasoning/transcripts. At the final integrity check, the actual repository AL-64 ledger still had the original baseline SHA256; none of these experiments wrote it. All invoked CLI processes exited and no background process or server was started.

Not established by this bounded probe: final reviewed revision approval, full make, unavailable-source retention, other harness replacement, fixed-cutoff idempotence, frontend/API rendering or a new native OpenCode run under the scalar representation. Those require separate deterministic/integration evidence; unchanged prior OpenCode native boundary observations remain in iteration 01. Existing missing T3 source facts remain explicitly incomplete and were not filled by guessing associations.
