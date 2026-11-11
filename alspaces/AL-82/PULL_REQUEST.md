# feat: filter and compare metrics by ticket tags

Add durable ticket-level tags for cohort analysis without changing session user identity or untagged historical metrics. Shared defaults, private local overrides and repeatable `generate --tag KEY=VALUE` merge per key; absent inputs preserve recorded tags. Effective private tags become shared ledger metadata. Shared project tags record the dev, PO and review skill versions (`0.1.1` each), inherited by ticket generation.

Dashboard filters select whole tickets with OR values, AND keys and a distinct Missing choice. Optional comparison coloring preserves point coordinates, completion details and existing exclusions. Refreshed ticket values update the catalog and legend without erasing other cohorts.

Validation passed on `57b1a45c`: full `make` including Go race coverage, 230 frontend unit and 768 frontend integration tests; host `make verify-build`; standalone Linux HTTP/embedded-assets verification with no Node on PATH. Independent architecture, behavior, data and test reviews on both configured profiles approved the corrected source. Three confirmed defects have regressions: Unicode acceptance, late-publication rollback and refreshed tag catalogs.

[Contract](https://github.com/NiltonDuarte/AgenticLens/blob/AL-82/filter-compare-metrics-by-tags/alspaces/AL-82/CONTRACT.md) and [summary](https://github.com/NiltonDuarte/AgenticLens/blob/AL-82/filter-compare-metrics-by-tags/alspaces/AL-82/SUMMARY.md) retain scope, review credit and verification limits. Live-stream filtering, deletion and bulk retagging remain outside scope.
