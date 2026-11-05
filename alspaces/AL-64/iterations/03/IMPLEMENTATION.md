# AL-64 iteration 03 implementation

Implementation baseline: c68fbc9079d2f24c624d5de0fe87d313ed095b19. Coder: gpt-6.1-sol/high, inherited runtime. Scope follows the approved numeric-only iteration 03 plan. Parent-owned planning/summary changes and independent tester changes are preserved. Product edits are frozen for verification; this report does not constitute consolidated approval or ticket completion.

## Canonical behavior

Invocation facts now contain the existing recorded active_duration_seconds total and optional flat provisional_seconds. The provisional portion is included in the total, not added to it. Canonical writes omit zero; input accepts zero, rejects negative/noninteger/null/unsafe values and rejects a provisional portion greater than its total. Workflow, fleet, comparison and progression projections expose only the same flat numeric provisional subtotal. The canonical domain no longer has a rich duration-evidence model or a parallel mandatory coverage graph.

The pure core keeps interval arithmetic, transient caller-supplied cutoff, established invocation rounding and reconciling numeric portions. Source adapters retain Codex turn versus OpenCode assistant-message semantics, current native activity markers, explicit terminal handling and all six corrected boundaries. Malformed Codex lifecycle boundaries retain parser-known numeric intervals but produce parse warnings, which strict replacement callers reject; invalid OpenCode interval boundaries return contextual extraction errors. Unsupported missing activity proof does not create elapsed time. Numeric zero or absence is never described as completed source coverage.

No liveness/caller-proof mechanism, source wakeup or session message is introduced. Open source state accrues provisionally as the accepted estimate, including the acknowledged possibility that a process stopped without recording a native terminal marker.

## Backend compatibility and ownership

The private draft input shape is confined to backend migration/input compatibility. Shared validation accepts original v6, draft evidence, lean scalar and matching dual forms. It validates legacy state/component/cutoff consistency before projecting the numeric portion. Raw member presence distinguishes omitted values from explicit zero and null; inconsistent dual values are rejected. Duplicate invocation/draft JSON members are rejected before normalization can hide them. Non-duration fields remain for the owning strict reader to validate.

Adapter invocation DTOs, historical migration decoding, current-schema explicit migration and registered snapshot input all normalize at their input boundaries. Retained history DTOs are already validated numeric records; exact owner/numeric equality is checked before output projection. Canonical target writes omit retired duration metadata even for unavailable retained owners, without changing source facts, the caller-owned baseline or unrelated files. Capture reads normalize returned values while preserving original registry bytes; unsupported retired trace/occurrence cache remains subject to its existing opaque/re-extraction rules.

An early native copied-ledger backfill and the independent tester exposed a current-schema migration path that decoded domain structs directly and bypassed the old invocation-input adapter. That path now performs the same duration normalization/validation before strict current-domain decoding. The original failed attempt is retained in the native evidence report. This is an input compatibility expansion, with no source/project identity guard bypass.

AL-66 optional skills/delegations/anomalies presence, configuration and final projection remain unchanged. AL-77 shared-session warnings and their complete population/request ownership remain unchanged. The canonical duration scalar is independent of those optional output categories.

## Presentation

CLI/Markdown describe recorded duration and only a positive numeric provisional portion. The dashboard and API use the scalar directly; no frontend draft object or cutoff/count interpretation remains. Ticket/fleet/comparison/scatter displays identify positive provisional amounts without making completion claims for zero/absent values. The average comparison KPI computes its provisional portion using the same per-ticket denominator and rounding as its average total. Progression calculations carry the numeric provisional subtotal. Existing primary KPI numbers and other product capabilities are preserved.

The owning schema guidance is updated in docs/OWNED_OCCURRENCES.md. Reader-first rollout limitations remain explicit: a pre-expansion strict reader may reject new scalar output; the change does not authorize distributing or releasing binaries.

## Implementation paths

- Domain: duration.go, models.go and aggregations.go; migration duration_decode.go, cost_decode.go, engine.go and invariant validator.
- Boundaries: invocation and fact-ledger DTOs, workflow/HTTP comparison/fleet/progression DTOs, capture/store.go, Codex rollout and OpenCode v1/v2 extraction/accumulation, CLI list summary and Markdown formatter.
- Frontend: numeric domain types/duration helper/progression/formatter, API validator, duration consumers in ticket/fleet/list/comparison/scatter views and components.
- Tests: exclusively owned by the independent tester after parallel coordination. Coder did not modify test sources in this iteration.

## Coder verification and limits

Production Go build passed across backend packages. Focused DTO/migration/retained-draft duration tests passed after the current-schema migration correction. Focused Codex, OpenCode and backfill packages passed on the frozen production candidate. Full affected-package outcomes are appended below when complete. Frontend typecheck/lint and canonical make web-assets passed. The ignored embedded asset tree contains index-zSnCKFwx.js and index-BGH-Pi_l.css.

The runtime tester independently verified a genuine copied 19-owner draft ledger: force-only normalization preserves source-owned numeric facts, followed by real Codex source-backed recomputation with authoritative new-child discovery, positive-only scalar output and no lost/duplicate owners. See NATIVE_REFRESH_EVIDENCE.md for exact binary, commands, source identities, failed/corrected attempts and limitations. This coder did not mutate any live source, ticket ledger or workspace.

Independent full Go race/frontend verification, focused review, required integrated make, final standalone asset verification, telemetry and PR delivery remain parent/tester-owned. No staging, commit, telemetry generation, push or PR action was performed by the coder.

## Final boundary correction before review

Coder source tracing identified a concrete equality bypass before review: draft evidence with zero provisional seconds plus a noncanonical Provisional_Seconds value of seven could evade exact raw-key checks and then populate the canonical struct through Go's case-insensitive JSON matching. The duration input normalizer now rejects case aliases of active_duration_seconds, provisional_seconds and duration_evidence before projection. This is narrowly scoped to duration members; unrelated JSON-policy behavior is unchanged. Independent tester owns regression assertions at runtime validation, DTO and current-migration boundaries. Broad suites started before this correction remain evidence for the prior revision, and affected checks must run after the fix. The earlier native-probe binary is preserved; the separate final binary is rebuilt for the corrected candidate.

Final ordinary focused Go command passed Codex, OpenCode, backfill, capture, HTTP, domain and migration packages. The independent tester also reported green scalar compatibility/retained-history/registered-cache checks before the final case-alias correction. Frontend typecheck/lint and canonical assets passed with no frontend changes following that build. Full required make and correction-aware review remain parent-owned.

## First-review corrections

LEAN-WRITE-01: the completed-migration cleanup path had validated the numeric domain projection but published its original metadata-free raw input, retaining draft duration metadata. The cleanup candidate now applies the shared duration-only input/output normalization before removing migration metadata. All other supported raw fields, optional collection presence and row multiplicity remain independent of that projection; no unrelated file or caller-owned baseline is changed.

LEAN-READ-01: historical v1-v5 typed decoding could collapse a canonical scalar and a case-variant scalar before old/new equality validation. A raw invocation guard now validates the known flat and session-owned invocation collections before historical typed decode. It reuses shared duration-member validation without rewriting the historical input or adding a custom decoder to the embedded invocation type, preserving flat-row identity and migration ownership behavior.

Legacy nested-null numeric interpretation is deliberately unchanged, matching the verified baseline-compatible interpretation documented as LEAN-READ-02. No new legacy-null policy, output option or liveness machinery is introduced. The independent tester owns correction oracles and original-code overlay evidence; parent owns finding attribution, dispositions and final approval.

The existing raw cleanup regression checks also showed that unconditional map serialization reordered already-canonical session/invocation members. Shared duration normalization now returns original bytes whenever no draft member or zero scalar needs projection, propagating that no-change result through owner and ledger inputs. Original wire validation still runs before that return. This preserves established metadata-only cleanup oracles and avoids touching unaffected raw fields merely to remove migration metadata.

The historical raw guard is limited to duration names and values. It preserves the historical decoder's unrelated-field and unrelated-duplicate tolerance rather than applying the canonical object's broader duplicate rejection to older input. Duration members remain duplicate/alias checked before typed matching, and shared legacy numeric validation remains unchanged. This scope check was raised by the independent tester and retained as a control requirement, not used to broaden historical JSON policy.
