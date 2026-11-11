# AL-70 A12 architecture review

APPROVED for the assigned architecture scope. No findings; candidate IDs: `[]`.

Run `a12_arch_sol`, independent first pass, iteration 03. Profile `gpt-6.1-sol/high` is dispatch-acknowledged; runtime model metadata and usage/cost were not independently observed. Reviewed baseline `54d58be40c74864d645c5168ba23820a21389c01` plus frozen candidate fingerprint `974d1449fc440115b563067461d43b9b6381baf60e178ff45e6854f884b10e6a`. All 29 manifest files matched SHA-256 at entry and exit. Product/tests remained read-only; no sibling conclusions were consulted.

The inspected producer-to-consumer paths preserve the approved ownership boundary:

- OpenCode v1 assistant and both families' metadata-only accumulators use the empty key without dropping rows. Existing accumulator counters and nonempty metadata parsing remain intact. The v2 assistant `model.id` rejection at `schema_v2_session_extraction.go:217` remains unchanged; A12 does not broaden that source family.
- Codex `assignTurnModel` (`rollout.go:537`) distinguishes absent entries from present empty-model turns and reassigns only the latter. Literal `default` and `unknown` remain observed identities. `turnNumber` (`rollout.go:895`) uses presence, retaining earlier-turn references when the model key is empty. Finalization retains the empty accumulator and existing late-context usage/duration semantics.
- `InvocationFactDTO`, `ModelMetricDTO`, live JSON fields, frontend validators and domain grouping preserve raw strings, including empty. No domain, DTO, schema, migration or grouping redesign was introduced. The unchanged session-owned projection copies invocation facts without relabeling historical names.
- `formatModelLabel` and Markdown labeling act only at presentation. Distribution/table keys, cost-summary lookups, sorting and colors still consume raw names. Optional debug-prompt model omissions remain distinct from supplied empty values. No new I/O or outer-layer dependency enters the functional core.

Verification: independently ran the five focused Codex identity/late-attribution/turn-lookup tests with `go test -count=1 ./internal/adapters/codex -run 'TestParseRollout(MissingAndObservedModelIdentity|ObservedNamesAreNotProvisional|LateModelPreservesEarlierUsageAndDuration|UnknownModelTurnLookupUsesPresence|UpdatesModelWhenTurnContextFollowsTurnStart)$'` (PASS), plus `git diff --check` (PASS). Inspected the source/test oracles and reused `/tmp/al70-a12-tester-evidence.json` and raw passing backend/frontend logs: real SQLite/rollout fixtures, historical API byte preservation, 112 formatter/prompt units and 149 consumer integrations. Reused frontend typecheck/lint/build evidence.

Limits: bounded A12 review, not whole-repository certification. No live developer stores or deployed browser were exercised. Parent owns main integration, canonical `make`, standalone verification, consolidation and delivery.
