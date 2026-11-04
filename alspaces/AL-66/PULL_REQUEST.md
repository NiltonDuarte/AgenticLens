# feat(generate): configure optional metric output (AL-66)

Add independent enabled-by-default project settings for generation.metrics.skills, generation.metrics.delegations and generation.metrics.anomalies. Disabled categories are omitted from generated JSON and managed Markdown sections, including facts retained from existing ledgers; invalid settings fail before output replacement.

Record optional schema-6 generation policy and unavailable-analysis metadata so readers distinguish intentional omission, missing evidence and measured zero. Preserve that uncertainty through later backfill. Preserve core session/invocation metrics, child relationships, enabled history, live/backfill defaults and independent source traces. Re-enabling recovers facts only when source data remains available.

Document configuration precedence, anomaly diagnostic behavior and recovery limits.

- [Technical contract](CONTRACT.md)
- [Work summary and verification](SUMMARY.md)
