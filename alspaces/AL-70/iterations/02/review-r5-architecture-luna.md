# AL-70 R5 architecture review — Luna profile

**Outcome: APPROVED (scoped).** Candidate AL70-R5 at base revision `f63e3db6b2d4baa8e22452e4ec94eff566867721`; requested profile `gpt-6-luna/medium`, run `r5-architecture-luna`. This review covers UI-05 formatter/caller ownership and direct consumers only. Entry and exit checks matched all 274 manifest inputs with fingerprint `53216836c57b58aea99790ea61b6556e92a7bdc485bcda24907b4545b975da57`.

## Findings

No blocking architecture or maintainability finding in the assigned scope.

`formatAnomaly` accepts a small role-label function, keeping canonical anomaly data independent of project presentation. `TraceWaterfall` owns raw source anomaly/event rendering and supplies `sourceLabel`; the explicit recorded-summary section supplies the identity `projectedLabel`, then records `roleOrigin: 'projected'` in its selected entry. `AnomalyDrawer` chooses one label policy from that provenance. Raw trace entries remain tagged `source`; `selectWorkspaceAnomalyEntries` and `recordedAnomalyEntries` do not alter their anomaly objects. The role provenance is local UI state, not added to DTOs or persisted facts.

The drawer's generated title/description use the selected presentation function, while prompt-copy generation deliberately calls `formatAnomaly(anomaly)` with its default raw formatter. Structured facts and evidence are read from the original anomaly/event values. Missing anomaly fields still take the existing generic description branch. The trace summary button exposes the generated description as its native `title`, alongside the display title and accessible label. This preserves a clear owner for display aliases without introducing a second mutable representation.

## Hypotheses investigated

- **Double application or remapping of projected labels:** not found. Raw source anomaly contexts are mapped at display time; recorded summary strings already supplied by the projected detail response use identity display in the drawer. Selection carries that distinction through `PresentedAnomaly`.
- **Display aliases leaking into copied diagnostics or source facts:** not found. The prompt path continues to format without the project callback and copies original diagnostic context; no formatter mutation or API/schema change is present in the inspected path.
- **Unclear ownership across direct consumers:** not found. Direct `formatAnomaly` consumers are `TraceWaterfall`, `TicketDeepDive`, and `AnomalyDrawer`; each selects source/projected behavior at its presentation boundary. The live inspector's role badge uses projected labels independently and does not pass through anomaly formatting.

## Verification and limits

Static source/diff trace covered `web/src/domain/anomalies.ts`, `web/src/domain/events.ts`, `web/src/components/TraceWaterfall.tsx`, `web/src/views/TicketDeepDive.tsx`, `web/src/components/AnomalyDrawer.tsx`, and `web/src/contexts/ProjectConfigContext.tsx`, including direct formatter consumers. Reused the supplied R5 evidence: focused canonical tests (218 passed) and affected integrations (173 passed); a disposable raw-only mutation failed the raw-display guards while the projected control passed. No new test execution was performed. Data/test R4 approvals remain carried only for unchanged identity/order/wire/harness scopes as parent specified; this scoped review does not claim those executions or approve the complete change. Runtime requested model/profile is recorded as supplied; no usage or cost is asserted.
