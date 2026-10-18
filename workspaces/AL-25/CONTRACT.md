# Contract: AL-25 Contextual Hover Information Icons for Cards and Tables

## 1. Problem Statement

The AgenticLens web dashboard aggregates dense operational telemetry, including context window saturation against model ceilings, prompt cache hit percentages, test-to-production code churn ratios, subagent turn velocities, and tool error rates.

Currently, the interface lacks consistent, accessible contextual explanations for these metrics:
1. The primary metric component at `web/src/components/KPICard.tsx` relies exclusively on the browser native HTML `title` attribute on the outer card container. This lacks visual affordance, offers no indication that an explanation exists, fails keyboard accessibility standards, introduces arbitrary browser render delays, and cannot be styled to match the dark theme.
2. The distribution component at `web/src/components/DistributionCard.tsx` has no explanatory tooltip property or visual indicator for role or model distributions.
3. Tabular headers in `web/src/components/TicketTable.tsx`, the quick-launch directory in `web/src/views/LandingPage.tsx`, and the comparison matrix in `web/src/views/TicketComparisons.tsx` display technical column labels without explaining underlying computation formulas.
4. Pulse summary cards and analytical feature navigation cards on the landing page in `web/src/views/LandingPage.tsx` lack explanatory context for new operators unfamiliar with multi-agent telemetry concepts.

Ticket AL-25 resolves these limitations by creating a reusable, accessible, zero-emoji contextual information component (`InfoTooltip` / `InfoIcon`) and integrating it across cards and tables throughout the web presentation layer.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All changes must satisfy `TECHNICAL_GUIDELINES.md`:

### Inward Dependency Flow
1. Functional Core:
   - Backend Go domain logic in `backend/internal/domain/` remains untouched. Pure mathematical rollups and constants (such as `HealthyCeilingTokens = 200,000`, `accumulateInvocations`, and `ComputeTotals`) serve as the authoritative semantic source for all tooltip explanatory text.
   - Frontend domain models in `web/src/domain/types.ts` and pure formatters in `web/src/domain/formatters.ts` remain untouched.
2. Imperative Shell:
   - Component rendering in `web/src/components/` handles DOM presentation, hover states, keyboard focus events, and positioning.
   - Component styles reside in `web/src/index.css` using existing CSS custom properties (such as `--bg-surface-elevated`, `--border-color`, `--text-primary`, `--text-secondary`, and `--font-sans`).
   - Zero external runtime script downloads or CDN stylesheet dependencies.

### Strict Typing and Zero Any
1. All component props, state variables, and placement options must use concrete, strictly-typed TypeScript definitions.
2. Zero `any` or `unknown` casts in component code or test files.
3. Tooltip placement must be modeled as a strict union of string literals: `'top' | 'bottom' | 'left' | 'right'`.

### Single Binary Distribution and Offline Independence
1. Tooltip rendering must rely on standard React and CSS layout without external third-party floating libraries, ensuring zero bundle bloat and zero Node.js runtime dependency in production.
2. Build output in `web/dist/` compiles cleanly via Vite and embeds directly into the Go executable via `embed.FS` in `backend/internal/adapters/web/embed.go`.

### Zero-Emoji Mandate
1. Zero emojis permitted in icon glyphs, tooltip text, button labels, ARIA tags, CSS classes, test assertions, or documentation.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Presentation Components (`web/src/components/`):
   - Create `web/src/components/InfoTooltip.tsx` exporting `InfoTooltip` and `InfoIcon` alias.
   - Render a high-contrast, accessible circular information glyph using an inline SVG icon (circle with centered lowercase `i` stroke) with `aria-hidden="true"`.
   - Render a keyboard-accessible trigger button with `type="button"`, `aria-label`, and `aria-describedby` referencing the tooltip container.
   - Render the tooltip popover container with `role="tooltip"`, unique DOM id, and placement class.
   - Support configurable placement: `'top'`, `'bottom'`, `'left'`, `'right'`, defaulting to `'top'`.
   - Refactor `web/src/components/KPICard.tsx`: replace outer `title` attribute reliance with `InfoTooltip` inside the `.kpi-header` row when `tooltip` prop is present.
   - Refactor `web/src/components/DistributionCard.tsx`: add optional `tooltip` prop and render `InfoTooltip` beside the title in a header container.
   - Refactor `web/src/components/TicketTable.tsx`: add `InfoTooltip` to all sortable and non-sortable column headers with `placement="bottom"`.
2. Views Integration (`web/src/views/`):
   - `web/src/views/LandingPage.tsx`:
     - Add `InfoTooltip` to all four pulse metric cards.
     - Add `InfoTooltip` to all four analytical feature navigation cards.
     - Add `InfoTooltip` to all column headers in the Quick-Launch workspaces table.
   - `web/src/views/FleetOverview.tsx`:
     - Supply explicit explanatory tooltip strings to Role Cost Distribution and Model Cost Distribution cards.
   - `web/src/views/TicketComparisons.tsx`:
     - Add `InfoTooltip` to all 15 comparison matrix table column headers with `placement="bottom"`.
3. Styling & Theme System (`web/src/index.css`):
   - Define styles for `.info-tooltip-container` (inline-flex, relative positioning).
   - Define styles for `.info-tooltip-trigger` (borderless transparent button, muted text color, smooth hover transition to text-primary, focus-visible outline).
   - Define styles for `.info-tooltip-glyph` (14px by 14px SVG dimensions).
   - Define styles for `.info-tooltip-bubble` (absolute positioning, z-index 100, max-width 260px, background `var(--bg-surface-elevated)`, border `1px solid var(--border-color)`, border-radius 6px, padding 6px 10px, font-size 0.75rem, line-height 1.4, pointer-events none, smooth opacity and visibility transitions).
   - Define placement offsets for `.placement-top`, `.placement-bottom`, `.placement-left`, and `.placement-right`.
   - Define caret pointer styles for each placement direction.
   - Ensure card headers and table headers accommodate the icon with proper flex alignment and no broken text wrapping.
4. Quality & Test Suites (`web/src/components/`):
   - Unit tests in `web/src/components/Components.test.tsx` covering `InfoTooltip`:
     - Renders trigger button with accessible ARIA attributes.
     - Renders tooltip content in the DOM with `role="tooltip"`.
     - Displays tooltip content on mouse hover and keyboard focus.
     - Applies designated placement modifier classes.
     - Contains zero emojis.
   - Regression verification across all existing test suites:
     - `web/src/components/TicketTable.test.tsx`
     - `web/src/views/LandingPage.test.tsx`
     - `web/src/views/FleetOverview.test.tsx`
     - `web/src/views/TicketDeepDive.test.tsx`
     - `web/src/views/TicketComparisons.test.tsx`
     - `web/src/views/MetricProgression.test.tsx`

### Out of Scope
1. Go Backend: No modifications to Go domain models, repository scanner, or HTTP handlers.
2. Third-Party Libraries: No external tooltip or Popper libraries added to `web/package.json`. Pure CSS and React implementation avoids dependency bloat.
3. Chart Element Tooltips: SVG chart tooltips in `TrendChart.tsx` already have dedicated SVG rendering and remain unchanged.

---

## 4. Component Architecture and Behavioral Specification

### Component: InfoTooltip
1. Props Interface:
   - `content`: string, required explanatory text.
   - `placement`: optional union of `'top' | 'bottom' | 'left' | 'right'`, default `'top'`.
   - `ariaLabel`: optional string, default `'Information'`.
   - `className`: optional string for additional container classes.
2. Elements and Accessibility Tree:
   - Outer wrapper: `span` element with class `.info-tooltip-container`.
   - Trigger element: `button` element with `type="button"`, class `.info-tooltip-trigger`, `aria-label` set to `ariaLabel`, and `aria-describedby` pointing to the bubble id.
   - Icon glyph: inline SVG with `viewBox="0 0 24 24"`, stroke `currentColor`, fill `none`, strokeWidth `2`, `aria-hidden="true"`. Renders circle at (12, 12) radius 10, line from (12, 16) to (12, 12), and dot at (12, 8).
   - Bubble element: `span` element with `role="tooltip"`, id generated or derived, class `.info-tooltip-bubble` and placement modifier, containing the text content.
3. Visibility and Interaction:
   - Hovering over the trigger transitions bubble opacity from 0 to 1 and visibility from hidden to visible.
   - Focusing the trigger via Tab key activates the bubble, ensuring complete keyboard accessibility.
   - Pressing Escape or blurring the trigger dismisses the bubble.

### Table Header Integration Pattern
1. Table headers with sort buttons or plain titles:
   - Header `th` element contains a flex container or inline-flex layout with header text, optional sort indicator, and `InfoTooltip`.
   - Stop click propagation on the `InfoTooltip` trigger button to ensure clicking the information icon does not trigger column sorting.

---

## 5. Acceptance Criteria

1. A reusable `InfoTooltip` component exists in `web/src/components/InfoTooltip.tsx` and is exported for application-wide consumption.
2. Every `KPICard` across Fleet Overview, Ticket Deep Dive, Metric Progression, and Ticket Comparisons displays an information icon when `tooltip` prop is provided.
3. Both distribution cards (Role Cost Distribution, Model Cost Distribution) in Fleet Overview display an information icon in the card header.
4. All four pulse metric cards and four feature navigation cards on the Landing Page display an information icon.
5. All table column headers in `TicketTable`, LandingPage Quick-Launch table, and `TicketComparisons` matrix table display an information icon describing the column metric.
6. Hovering or focusing on any information icon displays the styled tooltip bubble with high-contrast background and clear typography.
7. Information icon clicks do not trigger parent row navigation or column sorting.
8. Zero emojis appear in icon glyphs, tooltip copy, markup, CSS, or test assertions.
9. All unit and view tests pass cleanly with `npm test`.
10. TypeScript typechecking (`npm run typecheck`) and linting (`npm run lint`) pass with zero errors and zero `any`.
11. Single binary build via `make build` completes successfully.
