# fix(telemetry): persist lean skills and delegations (AL-62)

Codex previously omitted skill/delegation facts, while ticket-wide collections obscured their owner. The final schema stores lean skills and delegations within their source session: activation name/caller/time, and delegated agent/description/time/proven child session.

Schema 6 removes occurrence success/outcome/provenance metadata, duplicated delegation costs and persisted events. Traces hydrate separately from native sources with explicit complete, partial or unavailable coverage; they never enter ledger serialization. Activation counts and filters remain, while skill success-rate analytics are retired.

The direct 5-to-6 migration converts older ledgers, audits removed payloads and preserves known owners, multiplicity, child links and unrelated financial/anomaly facts. Current reads require the new schema. Partial refresh preserves unavailable owners; complete exact snapshots replace their own facts.

Contract: [CONTRACT.md](alspaces/AL-62/CONTRACT.md). Delivery details: [SUMMARY.md](alspaces/AL-62/SUMMARY.md).
