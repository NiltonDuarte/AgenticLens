# feat: filter and compare metrics by ticket tags

Add durable ticket tags from shared/local YAML and repeatable CLI overrides, preserving user identity and portable untagged history. Shared defaults also record dev, PO and review skill versions (`0.1.1` each).

Dashboard tags now use one compact key/value control inside each existing filter bar. Switch keys without losing other clauses and select multiple values or Missing. An active-filter list shows each key and its values with an X to remove that clause; one quiet Clear all action replaces the old clear buttons. Refresh and empty results keep controls usable. Optional comparison colors preserve coordinates, completion details and exclusions.

Validated on `1655ba45`: canonical `make` passed in an isolated network namespace after the host's occupied port 8085 blocked an existing test. Go race coverage, 230 frontend units, 781 frontend integrations and multi-platform builds passed; host binary and standalone embedded-dashboard checks passed with no Node on PATH. Paired behavior/test reviews approved the active-list refinement with no new findings; unchanged architecture/data boundaries retain prior review. Focus and positioning regressions are fixed with discovery credit preserved.

[Contract](https://github.com/NiltonDuarte/AgenticLens/blob/AL-82/filter-compare-metrics-by-tags/alspaces/AL-82/CONTRACT.md) and [summary](https://github.com/NiltonDuarte/AgenticLens/blob/AL-82/filter-compare-metrics-by-tags/alspaces/AL-82/SUMMARY.md) retain scope, findings, actors and verification limits. Live-stream filtering, deletion and bulk retagging remain outside scope.
