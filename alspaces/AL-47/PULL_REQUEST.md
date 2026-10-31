# fix(live): subscribe before exposing the SSE connection (AL-47)

The SSE handler flushed success headers and loaded its initial tree before registering a subscriber. A real poll during that lookup could publish the first trace before the connection subscribed; unchanged-trace deduplication then suppressed its delivery on later ticks.

Register the subscriber and deferred cleanup before headers and the initial lookup. Send the initial session tree before draining queued events. The existing SQLite-to-SSE-to-REST regression now controls this startup ordering using the concrete service and real SQLite while preserving its in-place-update assertions.

Follow-up to merged #76; this correction has no frontend, schema, cache, or replay-protocol changes. Original feature evidence and the AL-53 failed gate remain recorded.

Contract: [CONTRACT.md](CONTRACT.md). Correction scope: [iterations/01/PLAN.md](iterations/01/PLAN.md). Delivery evidence: [SUMMARY.md](SUMMARY.md).
