# feat(generate): configure optional metric output (AL-66)

Add independent enabled-by-default project settings for generation.metrics.skills, generation.metrics.delegations and generation.metrics.anomalies. Disabled categories are omitted from generated JSON and managed Markdown, including facts retained from existing ledgers; invalid settings fail before output replacement.

Keep JSON lean through field presence: omitted means not recorded, [] means recorded-empty, and nonempty arrays contain recorded facts. New output contains no generation_policy or anomaly_availability metadata. Existing transitional ledgers remain readable; explicit rewrites use the lean representation.

Preserve core metrics, child relationships, enabled history, known facts under partial coverage, and independent source traces. Generation settings remain in configuration and runtime context; persisted data does not retain why a category was not recorded. Complete source-backed recovery can restore facts, while unavailable sources cannot recreate intentionally omitted evidence. Backfill preserves these evidence distinctions without applying generation settings globally.

Document configuration precedence, diagnostics, field-presence semantics and recovery limits.

- [Technical contract](CONTRACT.md)
- [Iteration 02 plan](iterations/02/PLAN.md)
- [Work summary and verification](SUMMARY.md)
