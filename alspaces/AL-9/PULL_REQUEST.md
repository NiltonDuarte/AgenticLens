## feat(presentation): visual trace waterfall and anomaly inspector (AL-9)

### Architectural Context
- Contract: [CONTRACT.md](workspaces/AL-9/CONTRACT.md)
- Work Summary: [SUMMARY.md](workspaces/AL-9/SUMMARY.md)

### Overview
Delivers the complete visual execution trace waterfall, anomaly inspector drawer, and fleet-wide anomaly overview for the AgenticLens dashboard, with full fidelity error payloads, decoupled diagnostic inspection, smooth timeline interaction, and clear discovery affordances.

### Changes by Layer

#### Domain Core & Contracts
- `web/src/types/telemetry.ts`: Added canonical domain models for `EventFact`, `EventType`, `Anomaly`, `AnomalySeverity`, `AnomalyRuleID`, `AnomalyContext`, `BrittleToolSummary`, and `FleetAnomalySummary`.
- `web/src/domain/types.ts`: Re-exported and consolidated telemetry types across dashboard models with strict `readonly` immutability, zero `any`, and full `exactOptionalPropertyTypes` compliance.

#### Adapters & Network Boundary
- `backend/internal/adapters/opencode/extractor.go`: Raised tool error truncation bound from 200 to 4096 characters, capturing complete multi-line tracebacks and failure payloads without premature clipping.
- `web/src/api/client.ts`: Added `fetchWorkspaceEvents` (calling `/api/workspaces/{ticketId}/events`) and `fetchWorkspaceAnomalies` with fallback handling.
- `web/src/api/mockData.ts`: Added chronological event fixtures for ticket `AL-2` and maintained legacy fallback status for `AL-1`.
- `backend/internal/adapters/web/dist/`: Synchronized production build assets for single-binary embedding via Go `embed.FS`.

#### Presentation & UI Components
- `web/src/components/TraceWaterfall.tsx`: Chronological waterfall timeline defaulted to "Errors and Anomalies Only". Decoupled timeline node selection from drawer opening: unflagged events highlight without opening the drawer; error events and anomaly badges open the inspector. Provides an explicit `Inspect Failure →` action button on error turns, contextual tooltips, dynamic row classes, and an inline header guidance hint.
- `web/src/components/AnomalyDrawer.tsx`: Dynamic drawer titling distinguishing genuine domain anomalies (`anomaly.title`) from tool execution failures ("Tool Execution Failure") and execution details. Conditionally renders Rule Diagnosis strictly when domain rules exist, completely eliminating synthetic fallback strings.
- `web/src/components/AnomalyBadge.tsx`: Severity badge component (`WARNING` / `CRITICAL`), count display, accessible role switching, click event propagation control, and tooltip support (`title="Click to open Anomaly Inspector"`).
- `web/src/components/FleetAnomalyCard.tsx`: Fleet overview anomaly and reliability card with KPI cards, rule distribution counts, brittle tools ranking table, and clean zero-anomaly state.
- `web/src/views/TicketDeepDive.tsx`: Integrated `TraceWaterfall` with decoupled selection logic, legacy v1 fallback notification banner, and connected `AnomalyDrawer`.
- `web/src/views/FleetOverview.tsx`: Integrated `FleetAnomalyCard` into the fleet reliability layout.
- `web/src/index.css`: Eliminated selection flash and layout jump on non-issue rows with transparent baseline borders and targeted color transitions. Established neutral cursors for benign rows and active pointer cursors for issue rows. Styled `.btn-inspect-failure` action pill, `.waterfall-hint`, and 380px scrollable error containers.
