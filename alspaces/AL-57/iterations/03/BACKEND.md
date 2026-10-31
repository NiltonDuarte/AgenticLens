# AL-57 iteration 03 adapter implementation

Baseline: ee7558749c7101ddc4134cbdd22961bcba0a5a09. Ownership: adapters and shared pure domain validator only; domain/migration is owned by its separate coder.

## Concrete boundaries

- Invocation/task DTOs now carry nullable costs directly. Their known-location codecs retain required-member presence and use domain.ValidateCostEvidence. Canonical v4 and v5 ledger decode families select legacy compatibility versus strict required evidence. Older canonical output remains decodable for explicit migration tooling; ordinary runtime gating still accepts only canonical v4/v5.
- Declared output DTOs cover grouped/role/model/workflow metrics, thread records, HTTP facts/session/ledger, fleet maps, comparisons and time buckets. Every complete-cost field is nullable before serialization; non-cost members copy directly. The existing compact CLI export remains separate from HTTP facts/threads.
- HTTP's closed ResponsePayload union contains concrete projected response types. Error, build/config, event and live families pass through without cost logic. Fact and metrics persistence marshal their final DTO once with existing indentation/newline.
- Removed cost_json.go and ordered_cost_json.go; no byte-kind dispatch, arbitrary-tree field guessing, reflection, untyped values or ordered-object framework replaces them. Struct declarations preserve generated key order; real maps retain standard deterministic encoding.

## Assertion transfer

Removed only the obsolete arbitrary-tree normalization subtest in TestCostEvidenceBoundaryCompatibility as authorized by PLAN.md. The existing v4 numeric zero/nonzero/null/omitted and v5 contradictory-evidence tables remain. Actual SaveFactLedger/SaveMetricsLedger serializer guards retain nullable partial/subtotal/count/determinism/order assertions. Independent tester owns additional concrete Fleet/Workflow/map/task/pass-through cases. HTTP test compilation changed a detail fixture to its typed projection; the recorded-event preservation test independently decodes actual JSON to its domain ledger oracle and retains every prior assertion.

## Verification

Passed: all adapter package and CLI Go tests; go vet ./...; domain/adapter/HTTP/backfill/CLI focused repeat; git diff --check. Embedded assets/dependencies were already prepared. No application tests skipped. Independent tester/reviewer and parent full integrated make remain separate gates. No commits, pushes, telemetry or live ledger migrations by this coder.
