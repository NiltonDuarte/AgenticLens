# AL-71: Keep dashboard requests and live data reliable

## Roadmap ticket

[AL-71: Keep dashboard requests and live data reliable](../../../alspaces/plan/to-do/AL-71%20keep%20dashboard%20requests%20and%20live%20data%20reliable.md)

## Why

A late filter response can overwrite the current dashboard selection. Progression also switches date anchors when a type or prefix filter is applied, which can erase a chart of historical activity. Live traces can accept malformed data as a successful read, and slow live polling requests can accumulate.

## What changes

- Make the current Comparisons and Progression request own its results, diagnostics, loading and errors. Preserve the same guarantee in Fleet.
- Use the newest ledger in the selected results as the reference for Progression presets, and show the effective date range.
- Validate live payloads at the REST and SSE boundaries. Preserve the last valid observation and surface a diagnostic when a read is malformed.
- Bound fallback polling to one in-flight poll per subscription and restore healthy polling status after an accepted successful read.

## Capabilities

- Add `dashboard-query-reliability` for request ownership and Progression windows.
- Extend `live-source-reporting` with browser-side payload validation and fallback polling recovery.

## Impact

The changes affect Fleet, Comparisons, Progression and Live Observability. They preserve the comparison exclusion search, configured ticket matching, selected-session trace ownership, historical enrichment and existing source diagnostics. Benchmark arithmetic and server-wide date semantics stay under their existing contracts. The ledger format and live endpoints remain compatible with their documented producers.
