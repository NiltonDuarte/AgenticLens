# AL-68 data integrity final bounded re-review — data-luna-04

## Assessment

- **Scoped code verdict: APPROVED** for the requested data-integrity contracts on candidate 04. No substantiated introduced regression remains in the named cases.
- **Overall acceptance: INCOMPLETE.** Real Codex trust, callback, and agent-binding/source smoke remains unavailable, as recorded in `SMOKE.md`. No trust bypass or user-controlled callback was attempted. This evidence gap does not negate the scoped code result.
- Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Candidate: `CANDIDATE-04.json`; independently recomputed sorted path/NUL/content/NUL digest over its 375 listed paths: `c34cc949278608f4c777f7428a620e2ae29542e7fe33875d5b13ae9f0d17e6ec` (exact match).
- Requested profile: `codex / gpt-6-luna / medium`, run `data-luna-04`. The observed interface is Codex on the shared Linux checkout. Requested model/effort are not treated as independently observed runtime metadata.
- Reviewed the data reviewer instructions and common procedure, approved contract and plan including the actual hook-to-agent protocol clarification, correction/findings records, initial data review and candidate-02/03 data findings, implementation/test/smoke/setup/verification evidence, and relevant source/tests. Did not read the Sol-04 closing reports. Product and test files remained read-only; this report is the only path written.

## Findings and corrected IDs

No blocking data-integrity finding reproduced or remained in the bounded scope.

| ID | Disposition | Evidence assessment |
| --- | --- | --- |
| CAP01 | Fixed for the reported counterexamples | Registered refresh is filtered against the canonical ledger's source-owned completion checkpoint; explicit reads use the same whole-owner comparison. Equal or unknown differing facts remain unresolved and preserve the accepted owner. No token-field maximum is used. This addresses the exact older-source replay and equal/unknown checkpoint cases recorded in prior findings. |
| CAP02 | Fixed for the reported publication ordering case | `runGenerate` acquires output ownership before loading the baseline or reading sources, and retains it through ledger/report/summary publication. Explicit and registry-backed generations share the lock. Staged ledger/report replacement and process-owned kernel locking support the recorded crash/recovery boundary. No affected lock path changed in candidate 04. |
| CAP03 | Fixed for reported actual-store aliases | Registration canonicalizes existing store paths; selection checks stored identities with `SameStore`, including filesystem identity for aliases, and the report records symlink/hardlink rejection plus distinct-store controls. The former lexical-path bypass is no longer supported by the inspected path. |
| CAP05 | Fixed for failed-read accounting | Source attempts are charged before parse success is known. Codex failed parses reserve the source size; OpenCode preflights its estimated snapshot size before extraction; child metadata reserves its bounded read before parsing. Failed work remains subject to the same per-pass session/byte limits. The candidate-04 fairness correction changes scheduling, not these charging rules. |
| CAP10 | Fixed for declared cutoff semantics | `Cutoff` is assigned on reconciliation return after work and is explicitly described as completion, not a common historical source instant. Successful `refreshed_at` values remain per-owner checkpoints; failed/pending owners do not receive a fabricated successful checkpoint. |
| CAP11 | Fixed for ticket isolation and source-proven inheritance | Ticket reconciliation selects bound/source-proven owners; unassigned roots remain in local diagnostics without consuming that ticket's source refresh allowance or appearing in its final report. Child identity and parent come from authoritative source metadata. Relevant existing regression and candidate reports cover this boundary. |
| CAP12 | Fixed for the reported bounded starvation schedules | A durable phase rotates first priority across ordinary refresh, Codex stop queue, Codex start scan, and OpenCode child discovery. The first eligible class gets the configured remaining session and byte allowance; other work uses only leftovers. Existing cursors advance within parent/child queues and unresolved evidence remains. This directly covers persistent malformed-stop starvation of a start-only child and failed metadata consuming bytes needed by a valid owner, without increasing limits or dropping failures. OpenCode source-ID paging and queued-stop continuation remain intact. |

## Verification and limits

- Independently inspected the candidate fingerprint and source paths for checkpoint extraction/precedence, lock ownership and publication, actual-store identity, parse/metadata budget accounting, cutoff reporting, ticket selection, and rotating scheduler/cursor behavior.
- Reused the recorded candidate-04 affected race results rather than duplicating them: scheduler-targeted capture race tests and full domain/capture/CLI race packages passed with matching before/after source/test fingerprint `1a4a6c3ab0138374161d9ff51f6ffbc108cede7034978f034399f83432a174b1`. The reported full `make` result on candidate 03 is not represented here as candidate-04 integration evidence; parent owns the later integrated gate.
- Actual OpenCode 1.19.5 Linux binding/tool/source interaction and source-backed native-child observations remain valid within that runtime scope. They do not prove Codex callback delivery, human project trust, current T3 child behavior, or untested platforms/schema families.
- No new live-model exchange was performed for scheduler-only corrections. Codex trust/callback/agent smoke remains an external evidence gate; fixtures and static source tracing do not close it.

**Final scoped result: APPROVED. Overall acceptance remains INCOMPLETE pending the unavailable real Codex trust/callback/agent-binding evidence and parent-owned integrated verification.**
