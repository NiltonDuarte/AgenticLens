# AL-50 Iteration 01 Plan

## Grounded findings

- `web/src/utils/debugPrompt.ts` exports four layer-specific prompt generators and currently applies `sanitizeTextForPrompt` to raw inputs, outputs, error details, reasoning, message bodies, descriptions, and flagged issues. The sanitizer chains identifier/severity/rule replacement with punctuation and conjunction cleanup.
- `formatCodeBlock` in the same module chooses a fence longer than embedded backtick runs; preserve and reuse this defensive behavior for technical payloads.
- `web/src/utils/debugPrompt.test.ts` has broad generator coverage, including assertions that currently require ticket/branch/severity stripping. Replace those expectations with preservation-focused cases while retaining missing-data, metadata, and fence-boundary coverage.
- The generators are consumed in `web/src/components/LiveTraceInspector.tsx`, `web/src/components/TraceWaterfall.tsx`, `web/src/components/AnomalyDrawer.tsx`, and `web/src/components/FailedToolAlertCard.tsx`. Confirm signatures and rendered copy behavior remain compatible; do not expand component scope unless a compatibility defect is demonstrated.
- `TECHNICAL_GUIDELINES.md` requires strict TypeScript, no `any`, pure deterministic utilities, and graceful handling of missing presentation data.

## Implementation tasks

1. Refactor `web/src/utils/debugPrompt.ts` to remove identifier stripping, severity stripping, embedded rule translation, and regex-based grammar repair from raw payload handling. Keep only narrowly scoped deterministic metadata formatting if needed.
2. Ensure each raw evidence field is passed directly into the defensive code-fence formatter when rendered as a technical payload. Do not trim, normalize, or rewrite its contents; select empty-payload placeholders based on presence without changing included content.
3. Streamline all four prompt templates: concise provenance, descriptive section labels, and short actionable diagnostic guidance without repetitive boilerplate or rhetorical filler.
4. Update `web/src/utils/debugPrompt.test.ts` to cover exact raw payload preservation for ticket identifiers (`AL-42`), branch strings (`fix/FIX-3`), severity labels, rule identifiers, punctuation, whitespace, and embedded backticks. Retain coverage for every generator, optional fields, empty evidence, metadata formatting, and defensive fence selection.
5. Inspect the four established component call sites for signature compatibility and verify resulting Markdown is still passed to their copy controls/cards unchanged.

## Verification

- Run `npm test` for the debug prompt test suite and the frontend suite.
- Run `npm run typecheck` and `npm run lint`.
- Confirm `debugPrompt.ts` is below 250 lines and raw payload evidence is not routed through a destructive sanitizer.
- Run `make` or its applicable repository quality gates before delivery.

## Out of scope

- Go backend, API, event schema, or persistence changes.
- Changes to how anomaly rules are detected or stored.
- UI redesign or changes to copy-control behavior beyond preserving generated prompt output.
