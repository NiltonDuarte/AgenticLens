# AL-61 Technical Contract

The approved ticket is `alspaces/plan/to-do/AL-61 show full local timestamps in ticket deep dive.md`. User approval: “its pre-approved. full send it” on 2026-11-03. Initial revision: 122e983f8c365a2aa248b1b6d492252555e922ae; working tree initially clean. Initial branch: AL-61/show-full-local-timestamps-in-ticket-deep-dive; parent allocated feat/AL-61-full-local-timestamps for delivery. Workspace root resolves to alspaces from .agenticlens-config.yaml.

## Accepted behavior

Historical timestamps in Ticket Deep Dive show the viewer-local calendar date, hour, minute and second without milliseconds in visible text. Preserve timezone context through existing full tooltip and global timezone badge; the tooltip may retain milliseconds and the instant-specific UTC offset. Missing values retain explicit N/A, In Progress, or Unavailable states according to ownership. Invalid values must remain visibly unavailable or preserve existing raw invalid text, never fabricate a date.

## Label inventory

- TicketDeepDive.tsx: workspace Started, Completed and Updated labels, currently formatLocalTimeWithMs; missing completion remains In Progress and missing start remains N/A.
- TraceWaterfall.tsx: chronological event timestamp, currently formatLocalTimeWithMs.
- SkillsSubagentsPanel.tsx: recorded skill Loaded At timestamp, currently formatLocalTimeWithMs. Delegation rows have no timestamp label and need no new column.
- AnomalyDrawer.tsx: selected historical event Timestamp, currently raw ISO text; render the same local date and second precision while keeping source instant available as machine-readable datetime and supplemental tooltip.
- ThreadViewer.tsx has no structured timestamp labels. Raw thread/source diagnostic content is evidence and must remain verbatim.

## Boundaries

Changes belong to TypeScript presentation/formatters and narrowly scoped CSS. Reuse existing string instants from WorkspaceDetail and EventFact without modifying API schemas, validators, Go models, transport, persistence, sorting, duration calculations, chart axes or live inspector. Do not change shared formatLocalTimeWithMs semantics, since live callers rely on its precision. Existing API decoding preserves timestamp strings; inspect that path rather than add transport changes.

Use strict TypeScript with no any and no emojis. Date formatting converts the represented instant using the client timezone and must handle local date rollover and daylight-saving offsets. Keep visible values accessible as text, preserve existing controls and keyboard behavior, and allow full values to wrap or remain reachable through existing keyboard-scrollable tables at narrow widths. Keep alspaces/AL-55/agentic_metrics.json untouched.

## Acceptance evidence

Extend coherent formatter and view/component tests using self-contained fixtures. Cover different calendar dates, a non-UTC local rollover, DST instants, millisecond omission, missing and invalid values, integration of each label owner, preserved source datetime/tooltips and existing navigation. Complete canonical frontend build, inspect matching web/dist and backend/internal/adapters/web/dist trees, verify embedded assets/standalone binary, and pass make after integration. Tests of DOM/CSS alone do not prove browser geometry; state any browser-validation limit.
