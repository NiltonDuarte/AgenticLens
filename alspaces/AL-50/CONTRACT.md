# AL-50 Technical Contract

## Problem

`web/src/utils/debugPrompt.ts` currently removes ticket identifiers, branch names, severity labels, anomaly rule identifiers, and selected punctuation from text that may be raw runtime evidence. It also trims and rewrites payload text before placing it into prompts. Those transformations can corrupt commands, compiler output, stack traces, and agent messages. Its prompt templates repeat provenance boilerplate and use verbose checklists.

## Goal

Simplify debug prompt generation while preserving technical evidence exactly. Apply any normalization only to generated display metadata such as summary titles, never to raw diagnostic payloads. Keep all four prompt layers concise, factual, and useful to an AI assistant.

## Scope and boundaries

- Presentation utility: `web/src/utils/debugPrompt.ts`, including the four generators `generateToolExecutionPrompt`, `generateExecutionIssuePrompt`, `generateMessageTurnPrompt`, and `generateSessionOverviewPrompt`.
- Utility tests: `web/src/utils/debugPrompt.test.ts`.
- Consumers to verify for compatibility: `web/src/components/LiveTraceInspector.tsx`, `web/src/components/TraceWaterfall.tsx`, `web/src/components/AnomalyDrawer.tsx`, and `web/src/components/FailedToolAlertCard.tsx`.
- No Go domain, adapter, API, schema, or backend changes.

## Invariants

- Raw input, output, error detail, reasoning, message body, and flagged issue evidence must not undergo destructive replacement, whitespace trimming, punctuation repair, ticket/branch removal, severity removal, or rule-name translation.
- Raw technical content included as a code payload must be enclosed in a defensive Markdown fence whose delimiter is longer than any backtick run in the content. The payload bytes/characters between the fence boundaries remain unchanged.
- Metadata may be normalized only where required for safe, readable prompt headings. Keep this transformation deterministic and narrowly scoped.
- Preserve the existing exported TypeScript parameter types and generator function signatures unless a concrete compile-time incompatibility requires a narrowly justified adjustment.
- Keep TypeScript strict and avoid `any`; keep output free of emojis.
- Keep all work within the specified web utility, its tests, and AL-50 workspace artifacts.

## Acceptance boundaries

- Remove fragile regex grammar repair and destructive sanitization of raw evidence.
- Reduce `debugPrompt.ts` to under 250 lines, targeting under 200 where maintainable.
- Make the four prompt templates concise and remove repetitive database provenance boilerplate and rhetorical checklist items.
- Add or revise tests proving exact preservation of ticket IDs, branch names, severity labels, rule identifiers, punctuation, whitespace, and embedded backticks in applicable raw payloads.
- Preserve consumer integration and keep output well-formed when optional evidence is absent.
- `npm run typecheck`, `npm run lint`, and `npm test` pass; repository quality gates are verified during delivery.
