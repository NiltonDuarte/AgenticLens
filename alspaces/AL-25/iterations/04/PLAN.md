# Iteration 04 Implementation Plan: AL-25 Horizontal Table Overflow Containment and Responsive Scrolling

## 1. Goal and Iteration Scope

Iteration 04 addresses operator feedback regarding tables projecting out of their visual containers instead of establishing horizontal scrollbars. Specifically, the Role Resource Consumption table, Model Resource Consumption table, and Session Execution Threads breakdown bleed past card borders on constrained viewports.

This iteration delivers:
1. Addition of the missing global `.table-responsive` utility class definition in `web/src/index.css` specifying full width, horizontal auto-scrolling, and WebKit momentum scrolling.
2. Minimum width and cell white-space constraints for `.data-table` and `.thread-models-table` in `web/src/index.css` to safeguard column readability and trigger horizontal scrolling when container space is constrained.
3. Verification of coordinate space isolation between card header tooltips (situated outside `.table-responsive` with visible overflow) and column header tooltips (situated inside `.table-responsive` with downward placement).
4. Structural and behavioral test coverage additions in `web/src/views/TicketDeepDive.test.tsx` and `web/src/components/Components.test.tsx`, followed by complete typecheck, lint, and build verification.

---

## 2. Root Cause Analysis

### Root Cause 1: Missing Class Definition in Global Stylesheet (`web/src/index.css`)
In `web/src/views/TicketDeepDive.tsx`, `web/src/components/ThreadViewer.tsx`, and `web/src/views/TicketComparisons.tsx`, table elements are wrapped in a div with class `table-responsive`.

However, `web/src/index.css` contains zero CSS rules for `.table-responsive`. The browser user agent assigns this element default block styles: automatic width, visible overflow, and no scroll container formation. As a result, the wrapper div provided no horizontal containment.

### Root Cause 2: Overflow Bleed Unlocked by Iteration 03 Unclipping
In Iteration 03, `.table-card` and `.thread-accordion-item` were updated from `overflow: hidden` to `overflow: visible` to prevent clipping of card header `InfoTooltip` popovers.

Prior to Iteration 03, the outer `.table-card` container clipped any child elements extending beyond its boundaries. While this clipping erroneously cut off header tooltips, it also masked the fact that wide tables had no horizontal scroll wrapper. Once outer card clipping was removed, wide tables expanded beyond the padding box of their parent cards and bled directly into adjacent layout columns, accordion borders, and viewport margins.

### Root Cause 3: Column Count and Dimension Constraints
1. Role and Model Resource Consumption Tables (`TicketDeepDive.tsx`):
   - Arranged side by side in `.tables-grid` using a CSS grid template repeating auto-fit columns with a 400px minimum threshold (`repeat(auto-fit, minmax(400px, 1fr))`).
   - Accounting for 20px of card padding on each side, the usable interior width of `.table-card` at the 400px breakpoint is only 360px.
   - The five columns (Role/Model, Invocations, Tokens, Cache Hit, Cost) include header text and `InfoTooltip` trigger buttons. Because `.data-table th` specifies `white-space: nowrap` and `.th-content` uses inline flex, the natural minimum content width of these tables is roughly 480px to 520px. Without `.table-responsive` enforcing horizontal scrolling, the table expands to 480px+ and projects 120px+ past the card's right boundary.
2. Session Execution Threads Table (`ThreadViewer.tsx`):
   - `.thread-models-table` contains 10 distinct analytical columns: Role, Model, Invocations, Input, Output, Cache Read, Cache Hit, Cost, Peak Context, and Context Health.
   - Each column header includes non-wrapping label text and an information tooltip. Ten columns require an absolute minimum of 850px to 950px to display data without truncating metrics or causing extreme cell distortion.
   - Without a defined horizontal scrolling wrapper, this 10-column table projects out of `.thread-body` and `.thread-accordion-item`, overflowing the entire dashboard container.

---

## 3. Delta Architecture and Layout Specifications

The changes reside strictly within the imperative presentation shell. Pure Go domain models, boundary JSON schemas, and TypeScript domain formatters remain untouched.

### Specification 1: Responsive Wrapper Definition (`web/src/index.css`)
Define the class `.table-responsive` in `web/src/index.css`:
- `width`: 100% of the parent container to fill available card or accordion space.
- `overflow-x`: `auto` to create a dedicated horizontal scrolling context whenever table width exceeds container width.
- `overflow-y`: `visible` (browser computes to auto internally under CSS Overflow specifications, ensuring content is not vertically truncated).
- `-webkit-overflow-scrolling`: `touch` for smooth momentum-based panning on touch interfaces.

### Specification 2: Data Table Minimum Sizing (`web/src/index.css`)
Update `.data-table` in `web/src/index.css`:
- Retain existing width of 100% and border collapse.
- Add an explicit `min-width` of 480px (or 500px). This prevents columns from collapsing below legible proportions and reliably triggers `.table-responsive` horizontal scrolling whenever `.table-card` is rendered near or below 480px width.
- Add `white-space: nowrap` to `.data-table td` so numeric values, token counts, percentages, and dollar amounts never wrap across multiple lines.

### Specification 3: Thread Models Table Minimum Sizing (`web/src/index.css`)
Update `.thread-models-table` in `web/src/index.css`:
- Retain existing width of 100% and border collapse.
- Add an explicit `min-width` of 850px (or 900px). This guarantees that all 10 columns maintain proportional spacing and forces `.table-responsive` to scroll horizontally on viewports narrower than 900px.
- Add `white-space: nowrap` to `.thread-models-table td` to ensure session metrics and health status badges remain on a single horizontal row.

---

## 4. Coordinate Space and Tooltip Interaction Invariants

Establishing horizontal scrolling on `.table-responsive` requires strict boundary verification to ensure `InfoTooltip` instances remain visible and unclipped.

### Card Header Tooltips (Outside `.table-responsive`)
1. Role Resource Consumption and Model Resource Consumption headers in `web/src/views/TicketDeepDive.tsx`:
   - The header container `.table-card-header` resides directly inside `.table-card`, prior to and sibling to `<div className="table-responsive">`.
   - Outer `.table-card` enforces `overflow: visible`.
   - The header tooltips use `placement="left"`.
   - Because `.table-card-header` is outside `.table-responsive`, it is completely insulated from horizontal scroll clipping. The tooltip bubble projects leftward across card space without being cut off or generating unwanted scrollbars.
2. Session Execution Threads section header in `web/src/components/ThreadViewer.tsx`:
   - The header container `.thread-viewer-header` resides directly inside `.thread-viewer-card`, outside `.threads-accordion-list` and outside `.table-responsive`.
   - Its tooltip uses `placement="right"` and renders with complete visibility.

### Column Header Tooltips (Inside `.table-responsive`)
1. Table column headers in `.data-table` and `.thread-models-table`:
   - Tooltip triggers reside in `th` elements inside the scrolling table.
   - All column header tooltips use `placement="bottom"`.
   - Downward placement projects the tooltip bubble directly into the vertical space occupied by table body rows. Because tables contain multiple rows of data, the tooltip bubble overlays table rows without extending beyond the vertical footprint.
   - When the user scrolls the table horizontally, the column headers, labels, and tooltip triggers scroll synchronously together, preserving visual alignment between column headers and their contextual explanations.

---

## 5. Grounded References and Concrete Paths

### Presentation Styles (`web/src/index.css`)
- `web/src/index.css`:
  - Declare `.table-responsive` with full width, horizontal auto overflow, and WebKit touch scrolling.
  - Update `.data-table` with explicit minimum width and non-wrapping cell content.
  - Update `.thread-models-table` with explicit minimum width and non-wrapping cell content.

### Presentation Components and Views (`web/src/`)
- `web/src/views/TicketDeepDive.tsx`:
  - Verify that `.table-card-header` remains outside `.table-responsive`.
  - Verify that `<div className="table-responsive">` cleanly wraps `.data-table` for both Role Resource Consumption and Model Resource Consumption.
- `web/src/components/ThreadViewer.tsx`:
  - Verify that `.thread-viewer-header` remains outside `.table-responsive`.
  - Verify that `<div className="table-responsive">` cleanly wraps `.thread-models-table` inside `.thread-body`.
- `web/src/views/TicketComparisons.tsx`:
  - Verify that `<div className="table-responsive">` wraps `.comparison-matrix-table`.

### Automated Test Suites (`web/src/`)
- `web/src/views/TicketDeepDive.test.tsx`:
  - Add assertions verifying that the Role Resource Consumption and Model Resource Consumption tables are wrapped in elements with the `.table-responsive` class.
  - Verify that card header tooltips retain `placement-left` orientation.
  - Verify that table column headers retain `placement-bottom` orientation.
- `web/src/components/Components.test.tsx`:
  - In `ThreadViewer` test suites, add assertions verifying that `.thread-models-table` is wrapped in `.table-responsive`.
  - Verify that all 10 column header tooltips retain class `placement-bottom`.
  - Ensure accordion toggle interactions continue to pass without error.

---

## 6. Phased Implementation Tasks

### Phase 1: CSS Responsive Table Utility and Sizing (`web/src/index.css`)
- Task 1.1: Declare `.table-responsive` in `web/src/index.css`:
  - Set width to 100%.
  - Set overflow-x to auto.
  - Set -webkit-overflow-scrolling to touch.
- Task 1.2: Update `.data-table` styling in `web/src/index.css`:
  - Set min-width to 480px.
  - Set white-space to nowrap on `.data-table td`.
- Task 1.3: Update `.thread-models-table` styling in `web/src/index.css`:
  - Set min-width to 850px.
  - Set white-space to nowrap on `.thread-models-table td`.

### Phase 2: Structural Verification of Views and Components (`web/src/`)
- Task 2.1: In `web/src/views/TicketDeepDive.tsx`:
  - Confirm Role Resource Consumption table is wrapped by `<div className="table-responsive">` as sibling to `.table-card-header`.
  - Confirm Model Resource Consumption table is wrapped by `<div className="table-responsive">` as sibling to `.table-card-header`.
  - Confirm header tooltips maintain `placement="left"`.
- Task 2.2: In `web/src/components/ThreadViewer.tsx`:
  - Confirm `.thread-models-table` is wrapped by `<div className="table-responsive">` within `.thread-body`.
  - Confirm thread title group tooltip maintains `placement="right"`.
- Task 2.3: In `web/src/views/TicketComparisons.tsx`:
  - Confirm `.comparison-matrix-table` is wrapped by `<div className="table-responsive">`.

### Phase 3: Automated Test Suite Updates (`web/src/`)
- Task 3.1: In `web/src/views/TicketDeepDive.test.tsx`:
  - Write test asserting that the Role Resource Consumption table resides within a `.table-responsive` container.
  - Write test asserting that the Model Resource Consumption table resides within a `.table-responsive` container.
  - Assert that card header tooltips have class `placement-left`.
  - Assert that column header tooltips have class `placement-bottom`.
- Task 3.2: In `web/src/components/Components.test.tsx`:
  - Write test asserting that `ThreadViewer` renders `.thread-models-table` inside a `.table-responsive` container when expanded.
  - Verify that all 10 column header tooltips retain class `placement-bottom`.
  - Ensure accordion toggle interactions continue to pass without error.

### Phase 4: Quality Verification and Build Packaging
- Task 4.1: Run `npm run typecheck` in `web/` to confirm strict typing with zero `any` or `unknown` casts.
- Task 4.2: Run `npm run lint` in `web/` to verify zero ESLint errors or warnings.
- Task 4.3: Run `npm test` in `web/` to ensure all existing and updated unit/view tests pass.
- Task 4.4: Verify complete absence of emojis across all modified code, CSS, markup, and test files.
- Task 4.5: Execute root `make build` to verify Vite compilation and single-binary Go embedding via `backend/internal/adapters/web/embed.go`.

---

## 7. Execution Dependencies and Delivery Sequence

1. Phase 1 (`web/src/index.css`) provides the foundational CSS layout rules for `.table-responsive`, `.data-table`, and `.thread-models-table`.
2. Phase 2 verifies markup hierarchy across `TicketDeepDive.tsx`, `ThreadViewer.tsx`, and `TicketComparisons.tsx`.
3. Phase 3 updates test assertions in `TicketDeepDive.test.tsx` and `Components.test.tsx` to lock in structural regression guards.
4. Phase 4 runs end-to-end typecheck, lint, test, and binary compilation checks.

---

## 8. Acceptance Criteria

1. `.table-responsive` is declared in `web/src/index.css` with width: 100%, overflow-x: auto, and -webkit-overflow-scrolling: touch.
2. In `TicketDeepDive`, the Role Resource Consumption and Model Resource Consumption tables scroll horizontally within their card boundary when container width is below the table minimum width (480px), never overflowing card edges.
3. In `ThreadViewer`, the Session Execution Threads table (`.thread-models-table`) scrolls horizontally within `.thread-body` when viewport width is below 850px, never overflowing accordion borders or the dashboard viewport.
4. Tooltip bubbles on card headers for Role Resource Consumption and Model Resource Consumption remain completely unclipped, floating to the left of their trigger buttons outside the `.table-responsive` scroll container.
5. Tooltip bubbles on column headers for all tables project downward (`placement="bottom"`), scrolling synchronously with column headers inside `.table-responsive`.
6. Unit tests in `web/src/views/TicketDeepDive.test.tsx` and `web/src/components/Components.test.tsx` verify the presence of `.table-responsive` wrappers and proper placement classes.
7. Zero emojis exist in code, markup, CSS, comments, or test assertions.
8. `npm run typecheck`, `npm run lint`, and `npm test` pass with zero errors and zero warnings.
9. `make build` completes cleanly, packaging the embedded dashboard into the single binary.
