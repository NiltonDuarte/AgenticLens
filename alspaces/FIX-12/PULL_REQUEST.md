# fix(deep-dive): render ledger waterfall without hydration (FIX-12)

## Summary

Render supported event facts from the committed telemetry ledger independently of local hydration. Hydration now enriches the trace when available while preserving ledger facts, selection state, and local-source precedence.

## Changes

- Preserve nested session events through repository reads and events fallbacks while retaining compact writes and runtime legacy-schema validation.
- Reconcile committed and local event facts without dropping unmatched or ambiguous rows; preserve local enrichment in mixed SQLite/Codex traces.
- Render the waterfall as soon as ticket details load, keep the inspector current after enrichment, and show accurate source and payload availability states.
- Add regression coverage and refresh embedded frontend assets.

## Context

- Contract: `alspaces/FIX-12/CONTRACT.md`
- Work and verification summary: `alspaces/FIX-12/SUMMARY.md`
- Iteration plan and review: `alspaces/FIX-12/iterations/01/PLAN.md`, `alspaces/FIX-12/iterations/01/REVIEW.md`
