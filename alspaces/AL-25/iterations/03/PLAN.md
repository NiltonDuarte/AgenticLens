# Iteration 03 Implementation Plan: AL-25 Unclipping Role and Model Resource Consumption Tooltips

## 1. Goal and Iteration Scope

Iteration 03 resolves operator feedback regarding contextual tooltips on the Role Resource Consumption and Model Resource Consumption cards in `web/src/views/TicketDeepDive.tsx`. In the current interface, the tooltip bubbles for these two card headers are trapped inside the card boundary, rendering only a truncated fragment of the tooltip window.

This iteration delivers:
1. Root-cause elimination of CSS container clipping in `web/src/index.css` by updating `.table-card` to `overflow: visible` and verifying related accordion and table wrappers.
2. Placement optimization for Role Resource Consumption and Model Resource Consumption header tooltips in `web/src/views/TicketDeepDive.tsx` using `placement="left"` to maximize readability and prevent boundary collisions with adjacent cards.
3. Test suite updates in `web/src/views/TicketDeepDive.test.tsx` and `web/src/components/Components.test.tsx`, followed by full typecheck, lint, and build verification.

---

## 2. Root Cause Analysis

### Root Cause 1: Container Overflow Clipping (`web/src/index.css`)
In `web/src/index.css`, rule `.table-card` explicitly sets `overflow: hidden`. 

In CSS layout, `overflow: hidden` establishes a strict rectangular clipping boundary. Any descendant element positioned absolutely—such as `.info-tooltip-bubble`—that projects beyond the outer padding box of `.table-card` is cut off by the browser rendering engine.

Other card containers across the application—including `.kpi-card`, `.distribution-card`, `.turn-velocity-card`, and `.trend-card`—do not specify `overflow: hidden`, allowing their contextual tooltips to render unhindered across layout boundaries. The presence of `overflow: hidden` on `.table-card` is an isolated anomaly.

### Root Cause 2: Header Proximity and Default Tooltip Placement (`web/src/views/TicketDeepDive.tsx`)
In `web/src/views/TicketDeepDive.tsx`:
1. The `InfoTooltip` instances in the headers of Role Resource Consumption and Model Resource Consumption do not pass an explicit `placement` prop, defaulting to `placement="top"`.
2. The card header `.table-card-header` is located at the top of `.table-card`, separated from the outer card border only by 20 pixels of container padding.
3. With `placement="top"`, `.info-tooltip-bubble` is positioned via `bottom: 100%` and `transform: translateY(-6px)`. A multi-line tooltip bubble of height 45 to 65 pixels extends 30 to 45 pixels above the top border of `.table-card`.
4. Because `.table-card` enforces `overflow: hidden`, every pixel extending above the top edge is clipped. Only the bottom sliver of the tooltip bubble remains visible inside the 20-pixel card padding area, appearing "trapped inside the box".
5. Furthermore, because `.table-card-header` uses `justify-content: space-between`, the trigger button sits flush against the right padding edge. Any centered upward or downward tooltip (`transform: translateX(-50%)`) risks horizontal clipping against the card's right boundary.

---

## 3. Placement Ergonomics Evaluation

In `web/src/views/TicketDeepDive.tsx`, the two cards are arranged side by side in `.tables-grid` using `grid-template-columns: repeat(auto-fit, minmax(400px, 1fr))` with a 20-pixel column gap.

### Evaluation of `placement="top"`
- Pros: Conventional default.
- Cons: Even with `overflow: visible`, upward placement projects the bubble over the bottom boundary of the preceding Subagent Turn Velocity card. If the viewport or scroll offset changes, top-positioned tooltips may collide with above elements.

### Evaluation of `placement="bottom"`
- Pros: Projects downward into the table space where vertical room exists.
- Cons: Because the trigger button is anchored to the far right edge of the card header, a bottom-centered tooltip (`left: 50%`, `transform: translateX(-50%)`) extends half of its 260-pixel width (130 pixels) to the right of the trigger. Since the trigger is only 20 pixels from the right border, 110 pixels of the tooltip bubble will spill over into the grid gap or directly overlap the neighboring Model Resource Consumption card in multi-column layouts.

### Evaluation of `placement="left"`
- Pros:
  - Projects directly to the left of the trigger icon (`right: 100%`, `top: 50%`, `transform: translateY(-50%) translateX(-6px)`).
  - The card header spans at least 400 pixels in width, with the title text occupying only the left half. The wide horizontal space to the left of the trigger provides unobstructed clearance.
  - The tooltip bubble remains fully contained within the horizontal bounds of its parent card, completely avoiding overlap with adjacent grid columns.
  - The arrow pointer (`.placement-left::after`) points directly to the right toward the trigger glyph, providing clear visual orientation.
- Cons: None for wide card headers.

### Architectural Decision
Configure both Role Resource Consumption and Model Resource Consumption header `InfoTooltip` instances with `placement="left"`. Column headers inside the tables will retain `placement="bottom"`.

---

## 4. Grounded Reference Architecture and Concrete Paths

The changes strictly touch the presentation shell without mutating Go domain models or TypeScript domain formatters:

### Presentation Styles (`web/src/index.css`)
- `web/src/index.css`:
  - Update `.table-card`: Change `overflow: hidden` to `overflow: visible` (or remove `overflow: hidden`).
  - Audit and adjust `.thread-accordion-item`: Update from `overflow: hidden` to `overflow: visible` so table headers and cells inside expanded session threads do not clip tooltips. Ensure border radius on `.thread-header-btn` and `.thread-body` preserves rounded card geometry without requiring container clipping.
  - Verify `.table-responsive`: Confirm default `overflow: visible` behavior is preserved so table column headers do not suffer secondary vertical clipping.

### Presentation Views (`web/src/views/`)
- `web/src/views/TicketDeepDive.tsx`:
  - Set `placement="left"` on the Role Resource Consumption header `InfoTooltip`.
  - Set `placement="left"` on the Model Resource Consumption header `InfoTooltip`.
  - Retain `placement="bottom"` on all column header tooltips (`Role`, `Model`, `Invocations`, `Tokens`, `Cache Hit`, and `Cost`).

### Automated Test Suites (`web/src/`)
- `web/src/views/TicketDeepDive.test.tsx`: Update and add assertions verifying `placement-left` class on the Role and Model table header tooltips, and verify that column headers retain `placement-bottom`.
- `web/src/components/Components.test.tsx`: Verify that `InfoTooltip` placement modifiers and `ThreadViewer` header and column tooltips remain unclipped and properly attributed.

---

## 5. Phased Implementation Tasks

### Phase 1: CSS Overflow Unclipping (`web/src/index.css`)
- Task 1.1: Locate `.table-card` rule in `web/src/index.css`.
  - Replace `overflow: hidden;` with `overflow: visible;` (or remove the property to rely on standard CSS visible overflow).
  - Verify container borders, padding, and dark theme background colors remain visually identical.
- Task 1.2: Locate `.thread-accordion-item` rule in `web/src/index.css`.
  - Update `overflow: hidden;` to `overflow: visible;`.
  - Ensure `.thread-header-btn` defines top-left and top-right border radii matching the item border radius (6px), and full border radius when collapsed.
  - Ensure `.thread-body` defines bottom-left and bottom-right border radii matching the item border radius (6px).
- Task 1.3: Audit `.table-responsive` and `.data-table` wrappers.
  - Ensure no wrapper sets `overflow: hidden` or creates an accidental clipping context that truncates column header tooltips.

### Phase 2: View Tooltip Placement Optimization (`web/src/views/TicketDeepDive.tsx`)
- Task 2.1: In Role Resource Consumption card header:
  - Add `placement="left"` to `InfoTooltip`.
  - Maintain existing explanatory content and accessible `ariaLabel`.
- Task 2.2: In Model Resource Consumption card header:
  - Add `placement="left"` to `InfoTooltip`.
  - Maintain existing explanatory content and accessible `ariaLabel`.
- Task 2.3: Verify that column headers in both tables retain `placement="bottom"`.

### Phase 3: Automated Test Updates (`web/src/`)
- Task 3.1: In `web/src/views/TicketDeepDive.test.tsx`:
  - In test suite checking header tooltips, assert that the tooltip bubble for Role Resource Consumption has class `placement-left`.
  - Assert that the tooltip bubble for Model Resource Consumption has class `placement-left`.
  - Assert that the column header tooltip bubbles continue to have class `placement-bottom`.
  - Confirm edge cases (zero roles, zero models) continue to pass all tooltip assertions.
- Task 3.2: In `web/src/components/Components.test.tsx`:
  - Confirm that placement class coverage for `'left'` remains verified.
  - Verify that `ThreadViewer` renders without layout regressions.

### Phase 4: Quality Gates and Verification
- Task 4.1: Run `npm run typecheck` in `web/` to confirm strict TypeScript typing with zero `any`.
- Task 4.2: Run `npm run lint` in `web/` to verify zero ESLint errors or warnings.
- Task 4.3: Run `npm test` in `web/` to ensure all unit and view test suites pass.
- Task 4.4: Verify complete absence of emojis across all code, markup, CSS, and test assertions.
- Task 4.5: Run `make build` from workspace root to verify Vite build packaging and Go single-binary embedding.

---

## 6. Execution Dependencies and Delivery Sequence

1. Phase 1 (CSS updates in `web/src/index.css`) eliminates the root-cause container clipping boundary.
2. Phase 2 (`TicketDeepDive.tsx` placement update) optimizes horizontal placement to prevent inter-card overlap.
3. Phase 3 (Test suite updates) ensures regression coverage for `placement-left` on table headers.
4. Phase 4 (Quality gates) validates end-to-end type safety, linting, and single-binary packaging.

---

## 7. Acceptance Criteria

1. Tooltip bubbles for Role Resource Consumption and Model Resource Consumption headers in `TicketDeepDive` render completely without top, bottom, or side clipping.
2. The tooltip bubbles for Role Resource Consumption and Model Resource Consumption headers display to the left of their trigger buttons (`placement="left"`), avoiding spillover into adjacent cards.
3. `.table-card` in `web/src/index.css` has `overflow: visible`, matching the behavior of all other dashboard cards.
4. `.thread-accordion-item` in `web/src/index.css` does not clip child table tooltips.
5. All column header tooltips across Role Resource Consumption and Model Resource Consumption tables remain intact with `placement="bottom"`.
6. Unit tests in `TicketDeepDive.test.tsx` and `Components.test.tsx` verify the updated placement and unobstructed rendering.
7. Zero emojis exist in code, CSS, markup, or test files.
8. `npm run typecheck`, `npm run lint`, and `npm test` execute with zero errors.
9. Single-binary build via `make build` compiles cleanly.
