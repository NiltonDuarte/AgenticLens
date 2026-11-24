# GEN-32: Publish Pages only from main

## Roadmap ticket

GEN-32 is an ad hoc maintenance ticket with no roadmap entry. It narrows the Pages publication delivered by [GEN-24](../../../alspaces/GEN-24/SUMMARY.md) and repaired by [GEN-28](../gen-28-restore-pages-publication-and-close-al-83/design.md).

## Why

Per-PR roadmap previews kept failing the shared publication job. Each PR push queued a serialized publication behind every other open PR, so one failure delayed or disturbed the rest. Re-running a failed job could also leave a pending marker that blocked all later publications.

## What changes

- The Pages workflow runs only on pushes to main and on a manual dispatch from main. Pull request events no longer run it, and the dispatch has no PR input.
- The PR candidate build, preview composition and preview reporting are removed.
- The first main publication after this change drops the previews retained from earlier publications.
- The pre-deployment readability check reads the same Pages artifact up to three times over about 80 seconds.
- The publish job fails at its first step on a re-run attempt and tells the operator to dispatch from main.

PR previews at `/AgenticLens/preview/<PR-number>/` stop working once this merges.

## Capabilities

Modify the [ticket Pages publication](specs/ticket-pages-publication/spec.md) capability.

## Impact

Production keeps its base path, environment, serialized lock, pending marker and recovery procedure. The `github-pages-preview` environment no longer has a caller.
