# AL-78: Reduce the delay from code commit to settled metrics

## Roadmap ticket

[AL-78 · Make completion feedback effective across harnesses](../../../alspaces/plan/to-do/AL-78%20make%20completion%20feedback%20effective%20across%20harnesses.md)

## Why

An agent can commit the result of a long task while its native interval is still open. A metrics read at that point can contain provisional duration and miss usage or turns that the source has yet to expose. Repeating `generate` inside that interval can publish a newer snapshot, but it cannot settle the open interval. Reduce the time from the code commit to queryable metrics for the completed work on native Codex and OpenCode.

## What changes

- Reproduce the code-commit workflow on both native clients. Measure commit to native completion, completion to source visibility, and source visibility to the first queryable settled publication.
- Reuse automatic lifecycle capture to refresh after natural completion without a metrics-only model turn. Change feedback wording or delivery, or improve the existing refresh path, where the reproduction identifies avoidable delay.
- Make manual `generate` explain what is provisional and how to refresh after completion. Provide an explicit operator fallback where automatic refresh is unavailable.

## Capabilities

Add `completion-metrics-freshness` for the relationship between a code commit, native completion, source availability, and queryable publication in automatic and manual workflows.

## Impact

The change affects CLI feedback, the installed lifecycle path where justified by evidence, and setup, operator, and developer guidance. Preserve the 180-second checkpoint default, exact attribution, the lean ledger, and portable history. Final metrics in Git can be recorded in a later telemetry commit when they become available after the code commit. Installer changes, forced continuation, and automatic Git writes are outside this scope.
