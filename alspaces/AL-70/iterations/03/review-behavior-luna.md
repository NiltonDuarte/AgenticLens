# Behavior review: AL-70 iteration 03 (A12)

**Outcome: APPROVED** for the assigned behavior scope. This is a scoped review, not a whole-change approval.

## Reviewed boundary

Compared baseline `54d58be40c74864d645c5168ba23820a21389c01` with the frozen candidate diff (`974d1449fc440115b563067461d43b9b6381baf60e178ff45e6854f884b10e6a`). Inspected OpenCode v1/v2 extraction, Codex rollout attribution, live model resolution, aggregate and Markdown output, and frontend/API human-label consumers. Requirements are A12 in `CONTRACT.md` and iteration 03 plan: empty string is the missing identity, displayed as “Unknown”; observed names including literal `default` and `unknown` remain identities; aggregate rows/counters and historical ledger bytes remain intact; v2 assistant-model validation remains required.

## Findings

No blocking behavior finding. No changed path I inspected rewrites retained ledger data or changes aggregation identity. Empty string remains a valid group key in unchanged domain aggregation and map-keyed cost summaries; formatting occurs at the Markdown/UI boundary. The rendered frontend assets are part of the frozen diff, alongside source changes.

## Hypotheses investigated

- **Missing values collapse into observed `default`/`unknown`.** Refuted: OpenCode v1 assistant records and metadata-only fallbacks, v2 metadata-only fallback, Codex model usage, and live resolver now retain `""`. Nonempty model values are separately grouped. v2's required assistant `model.id` guard is unchanged.
- **Late Codex model context loses provisional usage, duration, or turns, or reassigns an observed sentinel-like name.** Refuted by the new map-presence checks in `turnNumber` and `assignTurnModel`: only a present empty-key turn is eligible for late attribution. `default`/`unknown` are nonempty and remain fixed. The reassignment transfers provisional turn membership and start time while preserving unknown-key usage/duration; targeted tests cover multiple turns and late attribution.
- **Human labels change keys, row counts, ordering, or totals.** Refuted by display-only `formatModelLabel` and local Markdown label conversion. Grouping, sorting, color selection, React row keys, cost lookup and counters continue to use raw values. Empty and literal `Unknown` may display the same label, but are still separate groups/rows as required.
- **Historical reads relabel or rewrite stored values.** No read-side transformation or persistence change was introduced. Existing API returns raw model strings; UI formats at render time. Tester reports a repeated historical API byte-preservation check.

## Verification and limits

Reused the tester’s frozen-candidate evidence: backend targeted adapter/domain/API command passed; frontend unit (112) and consumer integration (149) checks passed; `git diff --check` passed. Relevant tests include native SQLite/rollout missing, blank, default and unknown models; metadata-only rows/counters; live HTTP empty-string values; raw API/UI keys and row counts; Markdown labels; and Codex late attribution/map presence. I did not rerun suites. Parent still owns integration, canonical `make`, standalone binary checks and delivery. This review does not claim broader regression coverage. Coordinator run ID and requested model/effort were not supplied in the delegated brief; runtime model identity is not independently asserted here.
