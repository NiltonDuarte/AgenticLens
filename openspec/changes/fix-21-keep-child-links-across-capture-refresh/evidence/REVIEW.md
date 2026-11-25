# FIX-21 consolidated review (working record)

**Candidate under review:** `133100bb..0f73f759`, first discovery pass.

**Consolidated verdict:** CHANGES_REQUESTED. Two defects were confirmed with reproductions (DATA-01, BEH-01), and test coverage has gaps (TEST-01 to TEST-04).

## Run roster

The model column shows the profile each run was asked to use. Every runtime reported the model ID; none could confirm the effort level.

| Focus | Model / effort | Verdict | Report |
|---|---|---|---|
| Behavior | claude-opus-5-5 / medium | APPROVED | [behavior-opus](review/behavior-opus.md) |
| Behavior | claude-sonnet-5-5 / high | CHANGES_REQUESTED | [behavior-sonnet](review/behavior-sonnet.md) |
| Data | claude-opus-5-5 / medium | CHANGES_REQUESTED | [data-opus](review/data-opus.md) |
| Data | claude-sonnet-5-5 / high | APPROVED | [data-sonnet](review/data-sonnet.md) |
| Architecture | claude-opus-5-5 / medium | APPROVED | [architecture-opus](review/architecture-opus.md) |
| Architecture | claude-sonnet-5-5 / high | APPROVED | [architecture-sonnet](review/architecture-sonnet.md) |
| Tests | claude-opus-5-5 / medium | CHANGES_REQUESTED | [tests-opus](review/tests-opus.md) |
| Tests | claude-sonnet-5-5 / high | CHANGES_REQUESTED | [tests-sonnet](review/tests-sonnet.md) |

## Canonical findings

Each finding is classified as introduced or pre-existing, has a severity from 0 to 10, and credits the runs that found it.

| ID | Sev | Issue | Discovered by | Disposition |
|---|---|---|---|---|
| DATA-01 | 6 | If a pass is interrupted after the owner snapshot is saved but before the family annotation, the parent's registry snapshot keeps no links. The next pass then has nothing to retain. The window existed before; this change makes it violate the registry-snapshot guarantee. | data-opus (probe), data-sonnet | Confirmed. Coder: apply retention when each owner snapshot is built. Tester: cancellation regression. |
| BEH-01 | 4 | A resumed child sweep whose first candidate cannot be admitted at the identification ceiling is closed with `cursor=''`, not left pending. | behavior-sonnet (probe); resume bookkeeping unasserted per tests-sonnet T2 | Confirmed. Coder: keep the cursor. Tester: resume-at-ceiling test. |
| DATA-02 | 3 | Accepted links the rule refuses to carry (ambiguous key, unknown invoked_at, held by a sibling) are dropped without any incompleteness signal. This contradicts acceptance 3. | data-sonnet, data-opus, behavior-sonnet | Confirmed. Coder: count them as unresolved child evidence and in the stderr line. Planner: restore the spec wording. |
| BEH-02 | 2 | `retainCaptureChildLinks` opens the registry without the busy-aware path. A contended registry then exits 1, not with the busy code, and the worker backs off. | behavior-opus, data-sonnet, behavior-sonnet | Confirmed by code reading. Coder fixes this together with ARCH-03. |
| ARCH-03 | 2 | The CLI drives a public result API in sequence and re-records the diagnostic, so a caller that runs the steps in the wrong order can record "completed". | architecture-opus, architecture-sonnet | Accepted: fold into one capture-side method. |
| ARCH-04 | 2 | For `.gz` rollouts, header admission checks the compressed size but charges decompressed bytes, so a header read can be abandoned partway. The total is never overspent. | architecture-opus, data-opus | Confirmed by code reading. Coder fixes. |
| ARCH-02 | 3 | The ancestry-agreement and required-field rules are duplicated between `headerFromMetadata` and `readSessionContext`. | architecture-opus, architecture-sonnet | Accepted: unify into one helper. |
| ARCH-01 | 3 | The capture exact lookup and walker duplicate the manual and live lookup rules. | architecture-opus, architecture-sonnet | Accepted debt. Unifying them would touch the unchanged manual and live paths. Recorded as a limit with a follow-up. |
| ARCH-05 | 1 | Comment and enum mismatches: the deferral comment, the Codex-only `readMetadata` comment, an empty Evidence on `refused_ambiguous`, test-only exports. | architecture-opus, architecture-sonnet | Accepted. Coder fixes the comments and enum; test-only exports are left as-is with a rationale. |
| DATA-03 | 2 | A queued child path whose child is not yet registered can be counted twice. This only over-reports. | data-opus | Confirmed by code reading. Coder fixes the dedup. |
| DATA-04 | 2 | The CLI helper ignores registry child snapshots as evidence, so in divergent states the older ledger link can win over a locally proven one. | data-sonnet | Coder evaluates and fixes if low-risk. |
| TEST-01 | 5 | The retention call in `regroupArchiveFamily` is unprotected: a deletion mutation survives. | tests-opus, tests-sonnet | Tester. |
| TEST-02 | 4 | The CLI "parent with some children" subcase does not reach fresh-plus-retained links, so the double-count guard goes unexercised. | tests-opus | Tester. |
| TEST-03 | 3 | The non-v7 parent instant fallback is only unit-tested. A +48 h shift mutation survives. | tests-opus | Tester. |
| TEST-04 | 3 | The `RestoredSessions` rewrite in the CLI helper is unprotected. | tests-sonnet, tests-opus (suggestion) | Tester. |
| TEST-05 | 2 | The fixture has no unrelated rollouts later than the root, and its irregular names sit only in pruned directories. | tests-sonnet, tests-opus (suggestion) | Tester. |
| TEST-06 | 1 | Minor test issues: the idempotence test should assert the exact `[C1 C2]`; nothing guards against production imports of `codexstoretest`; the worker test uses a wall-clock bound; manual linking of a missing child is only covered as setup; contradiction through `--capture-auto` is untested. | tests-opus, tests-sonnet, behavior-opus, architecture-sonnet | Tester. |

Accepted limits, recorded in delivery and not blocking:
- A machine that can never re-read the children stays partial, and the "continue reconcile" hint does not help there.
- Coverage can be reported as partial when the parent and a child are read in different rounds of the same pass.
- A very small `max_source_bytes` can stall identification; this shows as pending.
- Capture does not surface the contradiction warning, which is pre-existing behavior.
- OpenCode child pages use session slots, which is pre-existing behavior.

Reporters: behavior-opus O2, data-sonnet F5/F6, architecture-opus S5.

## Round 2 (re_review) at `86eb6dae`

**Corrections reviewed:** delta `0f73f759..86eb6dae`, which also adds ordinal group retention.

**Consolidated verdict:** CHANGES_REQUESTED. All round-1 findings are resolved or dispositioned. R2-01 needed a user decision; on 2026-11-25 the user chose "Alinhar por posicao".

| Focus | Model / effort | Verdict | Report |
|---|---|---|---|
| Behavior | claude-opus-5-5 / medium | CHANGES_REQUESTED | [behavior-opus-r2](review/behavior-opus-r2.md) |
| Behavior | claude-sonnet-5-5 / high | APPROVED | [behavior-sonnet-r2](review/behavior-sonnet-r2.md) |
| Data | claude-opus-5-5 / medium | CHANGES_REQUESTED | [data-opus-r2](review/data-opus-r2.md) |
| Data | claude-sonnet-5-5 / high | APPROVED | [data-sonnet-r2](review/data-sonnet-r2.md) |
| Architecture | claude-opus-5-5 / medium | CHANGES_REQUESTED | [architecture-opus-r2](review/architecture-opus-r2.md) |
| Architecture | claude-sonnet-5-5 / high | APPROVED | [architecture-sonnet-r2](review/architecture-sonnet-r2.md) |
| Tests | claude-opus-5-5 / medium | CHANGES_REQUESTED | [tests-opus-r2](review/tests-opus-r2.md) |
| Tests | claude-sonnet-5-5 / high | CHANGES_REQUESTED | [tests-sonnet-r2](review/tests-sonnet-r2.md) |

| ID | Sev | Issue | Reported by | Disposition |
|---|---|---|---|---|
| R2-01 | 4 | A partly re-read ambiguous group is refused as a whole. Its unread links are reported once and then lost from the registry and the ledger. This also happens in the CLI stage after the registry has already filled part of the group. | all six behavior, data and architecture runs | Confirmed, with reproductions. Decided by the user: align by position, applying the unique-delegation rules at each position. Refuse only when the fresh group is smaller than the accepted one. |
| R2-02 | 3 | An untimed unique delegation lets an unconfirmed different fresh child replace the accepted link. | behavior-opus-r2, architecture-sonnet-r2 | Confirmed. R2-01 fixes it. |
| R2-03 | 3 | An accepted delegation whose key has no fresh counterpart (for example, `invoked_at` was unknown and is now known) is dropped without a report. | behavior-sonnet-r2 N2, architecture-sonnet-r2 | Confirmed. Coder: report it as refused, with a typed unmatched reason, so it counts as unresolved. The later loss is recorded as a limit. |
| R2-04 | 1 | Complete groups emit `refused_ambiguous` decisions, which contradicts the report contract. | architecture-opus-r2, architecture-sonnet-r2, behavior-sonnet-r2, behavior-opus-r2 | Coder fixes it as part of the R2-01 rewrite. |
| R2-05 | 1 | The stderr line says "remains partial" when links were only cleared. | architecture-opus-r2 | Coder fixes the wording. |
| R2-06 | 1 | When the 2 s retention deadline expires during contention, the exit is 1 rather than the busy exit. | behavior-sonnet-r2 | Coder maps deadline-on-contention to busy. |
| R2-07 | 2 | `design.md` and `tasks.md` still name removed APIs and three evidence states. They also omit `Known`, the name-dedup hint and the per-owner application. | architecture-sonnet-r2 | Planner updates them. |
| R2-T1 | 3 | The contention test reuses an almost-spent 2 s context and fails about 1 time in 9 under `-race`. | tests-opus-r2 | Tester gives the retry its own context and reruns capture under `-race`. |
| R2-T2 | 2 | Group position versus absolute index is untested. | tests-opus-r2 | Tester adds interleaved and offset groups. |
| R2-T3 | 2 | The contradicted-child filter in refused groups is unprotected. | tests-opus-r2 | Tester. |
| R2-T4 | 3 | The `RecordDiagnostic` failure path is untested. | tests-sonnet-r2 | Tester. |
| R2-T5 | 3 | The unified conflicting-parent check on the header path is untested. | tests-sonnet-r2 | Tester. |
| R2-T6 | 3 | The queued-path cursor restore at the identification ceiling is unasserted. | tests-sonnet-r2 | Tester. |
| R2-T7 | 2 | The import-guard test runs `go list ./...`, which needs the gitignored `web/dist`. | tests-sonnet-r2 | Tester makes the guard independent of the embed. |
| R2-T8 | 1 | There is no CLI ordinal or alignment end-to-end case; the `delete(refused)` guard is unprotected; the `refused_ambiguous` decision has no assertion. | tests-sonnet-r2, tests-opus-r2 | Tester. |
| R2-P1 | 3 | Pre-existing: when two delegations share a task name and only one child is read, `resolveSourceChildren` points both at that child. | data-opus-r2 | Not introduced by FIX-21. Recorded as a pre-existing limit with a follow-up. |
