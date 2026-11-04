# AL-62 backend execution

Historical iteration 01 working record. The accepted iteration 02 contract, implementation and final SUMMARY.md supersede unfinished progress and compatibility assumptions below.

Baseline 19d70574; approved contract and plan. Backend coder owns domain/adapters/CLI and backend schema docs. Shared checkout, preserve frontend/parent changes. No commits or telemetry.

Representation declared: schema 6; SessionFact owns skills/task_links; FactProvenance retains event_id/message_id/call_id/turn_id. SkillFact preserves legacy success plus explicit outcome unknown/succeeded/failed and evidence_kind/path. Flattened API facts add inherited session_id/source. unresolved_facts preserves ownerless legacy history. Native occurrence identity alone deduplicates explicit transitional projections.

Implemented so far: domain types/projections, DTO shape/nullable cost carriage, v5->v6 transition and registry, migration nested decode/preservation, anomaly compaction retains owned facts, aggregation known/unknown outcome counters. Not yet compiled. Codex producer, OpenCode ownership and CLI/capture/backfill wiring in progress. Existing assertions require compatibility updates only where semantics change; meaningful new regressions follow.
