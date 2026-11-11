# feat: filter and compare metrics by ticket tags

Implementation in progress for AL-82.

Add a durable ticket-level tag map with per-key CLI-over-local-over-shared precedence, dashboard filters and optional tag colors in comparisons. Preserve user attribution and portable untagged history. Filters include or exclude complete tickets and their totals. Comparison coloring preserves one point per ticket, coordinates and existing exclusions, with a distinct Missing identity and accessible completion details.

The [contract](CONTRACT.md) defines accepted scope, including the user's ticket-level ownership instruction superseding the original session-level proposal. The [summary](SUMMARY.md) records verification and delivery status. Planned checks cover configuration, generation and capture publication, cross-machine/harness persistence and refresh, whole-ticket filter totals, comparison accessibility, integrated `make`, and standalone embedded-dashboard verification. Execution is pending.
