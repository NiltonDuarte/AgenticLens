# AL-82 iteration 03: Compact integrated ticket-tag filter

## Steering and approval

The user instructed: “ticket tags filter must be a single "thing" where you can select a key and values. It can take that much space. it should be integrated with filters already existing in the page.” This directly authorizes the UI refinement. The space concern is interpreted as a compact closed control. Baseline is `b10a4daa8ceb1d9ad963544d1d204d8b24809319`, initially clean; PR 141 remains open. The [contract](../../CONTRACT.md) and iteration 02 semantics remain authoritative.

Changed acceptance: replace `TagFilters`' full-width fieldset and one always-visible dropdown per key with one Ticket tags trigger, visually consistent with adjacent filters. A bounded popover contains a key selector, that key's multi-value choices, active-clause identification and clear actions. Selecting a key alone changes editor focus, not matching. Editing values preserves clauses for other keys; clearing one key removes only that clause, while clear-all restores no filter. Show a concise active count/summary when closed. Missing remains distinct from literal values. Empty catalogs/results, refreshed choices and long arbitrary keys must not strand existing selections or recovery controls. Support keyboard navigation, Escape, outside dismissal and predictable focus return.

This supersedes only iteration 02's presentation. Preserve whole-ticket OR-within-key/AND-across-keys matching, API transport, totals, shared defaults, recorded metadata, pre-filter catalog reconciliation and comparison color/point identities. No backend, configuration or domain-matching changes are authorized.

## Implementation and ownership

The coder owns `web/**`, initially including affected tests. Rework `components/TagFilters.tsx` and local styles in `index.css`; reuse existing filter styling without changing unrelated `MultiSelectDropdown` behavior. Place the control inside Fleet's existing date/type/prefix filter section, Comparison's `table-controls-bar`, Progression's `controls-filters`, and Deep Dive's ticket-selector bar in both matching and mismatch branches. Remove the separate tag fieldset rows. Closed footprint must remain independent of the number of available keys; narrow layouts wrap existing controls normally and constrain the editor within the viewport.

After coder handoff, the tester owns missing adversarial coverage. Parent owns integration, telemetry, artifacts and publication. All share the checkout and preserve others' changes.

## Verification and review

Add focused `TagFilters.test.tsx` interaction coverage; adapt existing Fleet/Comparison/Progression tag tests rather than losing their totals/refetch oracles. Cover key switching with retained clauses, OR/AND/Missing, clear-one/all, empty recovery, changed catalogs, long text, dismissal/focus and toolbar placement including Deep Dive mismatch. Use actual component state and existing API fixtures. Run frontend units, affected DOM suites, typecheck/lint, canonical asset build, integrated `make` and standalone embedding verification; record actual revisions/results.

Select independent behavior, architecture and test-effectiveness focuses, each on gpt-6.1-sol/high and gpt-6-luna/medium. Behavior covers editor accessibility, responsive containment and page integration; architecture covers local UI-state ownership and unchanged matching/transport boundaries; tests cover meaningful state and retained view oracles. Omit data-integrity focus because persisted facts, generation/configuration, codecs, domain predicates and numerical aggregation remain byte-identical; behavior/architecture verify this scope boundary. Freeze the candidate and assign separate private report paths before discovery. Prior backend/persistence evidence carries only while those paths remain unchanged.

## Completed focus/profile roster

The parent froze the candidate after tester handoff. All six current cells approved corrected source `650d370c`; initial Sol behavior/architecture findings required correction and re-review. [Retained findings](REVIEW_SUMMARY.json) preserve all reporting actors; the ticket summary records scope, outcomes and integration evidence.

| Focus | Profile | Final outcome |
| --- | --- | --- |
| Behavior | gpt-6.1-sol/high | APPROVED |
| Behavior | gpt-6-luna/medium | APPROVED |
| Architecture | gpt-6.1-sol/high | APPROVED |
| Architecture | gpt-6-luna/medium | APPROVED |
| Test effectiveness | gpt-6.1-sol/high | APPROVED |
| Test effectiveness | gpt-6-luna/medium | APPROVED |

Full integrated checks and standalone embedding verification passed on `650d370c`; the first host attempt hit an occupied-port environment conflict, resolved with an isolated network namespace and the same UID. The summary retains commands/results and limits. Completed private reports were compacted after approval. Iteration 02 evidence remains retained for unchanged backend/data boundaries.
