# Implementation Plan: AL-9 Visual Trace Waterfall and Anomaly Inspector (Iteration 03 Delta)

## 1. Iteration Goal and Steering Feedback Scope

This iteration incorporates follow-up steering feedback for ticket AL-9 to eliminate visual glitches during timeline interaction and provide explicit discoverability affordances for diagnostic failure inspection.

Iteration 03 targets two concrete areas:
1. Elimination of Flash Animation and Layout Jump on Non-Issue Rows:
   - Eliminate the visual flashing and layout recalculation when clicking or selecting a non-issue (normal, unflagged) row in the Trace Waterfall.
   - Refactor the timeline item styling in the stylesheet to establish a baseline 1px transparent border, preventing dimension changes when the active selection border is applied.
   - Replace indiscriminate transition declarations with targeted color transitions to eliminate animation jitter on non-color properties.
   - Establish clear cursor and interaction semantics, distinguishing clickable issue rows (errors and anomalies) with an active pointer from non-issue rows with neutral cursor behavior.
2. Visual Affordance for Tool Execution Failure Window and Anomaly Inspector:
   - Provide an explicit visual action cue ("Inspect Failure ->") on rows with tool execution errors, accompanied by hover styling and an accessible title tooltip.
   - Enhance the interactive affordance, hover state, and accessible title tooltip on the AnomalyBadge component.
   - Add a subtle, inline guidance hint in the trace waterfall header/toolbar indicating that clicking any error or anomaly opens detailed failure inspection.

This plan adheres strictly to the Zero-Code Mandate (100% descriptive prose, architectural specifications, and structured task lists) and the Zero-Emoji Mandate.

---

## 2. Grounded Reference Architecture and Concrete File Paths

### Presentation Layer Components
- File: web/src/components/TraceWaterfall.tsx
  - Component: TraceWaterfall
  - Modifications:
    - Determine issue state per row: evaluate whether a row has an error status or a correlated anomaly.
    - Dynamically apply CSS classes distinguishing issue rows (item-interactive, item-anomaly, item-error) from unflagged rows.
    - For tool execution errors, render an explicit action button or pill labeled "Inspect Failure ->" with an accessible title attribute and click handler routing directly to event selection.
    - Pass an explicit title tooltip attribute ("Click to open Anomaly Inspector") to rendered AnomalyBadge instances.
    - In the waterfall header/toolbar, render a subtle inline text hint: "Click any error or anomaly to inspect failure details".
- File: web/src/components/AnomalyBadge.tsx
  - Component: AnomalyBadge
  - Modifications:
    - Expand the component props interface to accept an optional title string property.
    - Apply the title attribute to both the interactive button variant and the static status span variant, defaulting clickable badges to "Click to open Anomaly Inspector".

### Stylesheet and Design Tokens
- File: web/src/index.css
  - Selectors: .timeline-item, .timeline-item.selected, .timeline-item:hover, .timeline-item.item-error, .timeline-item.item-anomaly, .timeline-item.item-interactive, .btn-inspect-failure, .waterfall-hint, .anomaly-badge.clickable
  - Modifications:
    - On .timeline-item: Define a baseline border of 1px solid transparent and set default cursor to default. Replace transition declaration with targeted background-color and border-color transitions over 0.12 seconds ease.
    - On .timeline-item.selected: Update border-color to the primary blue accent token and background-color to the elevated surface token without altering border width or box sizing.
    - On .timeline-item.item-interactive, .timeline-item.item-error, and .timeline-item.item-anomaly: Set cursor to pointer.
    - On .timeline-item.item-error: Align border rules to prevent horizontal layout shift between error and non-error states.
    - Author .btn-inspect-failure styling: Compact, semi-bold monospace pill with subtle red background, red border, red text, pointer cursor, and high-contrast hover feedback.
    - Author .waterfall-hint styling: Subtle, muted informational text with small font size positioned responsively in the toolbar/header.
    - Refine .anomaly-badge.clickable: Enhance hover feedback with interactive glow, elevated brightness, and clear affordance.

### Test Suites
- File: web/src/components/TraceWaterfall.test.tsx
  - Add tests verifying the elimination of interactive classes on unflagged rows, presence of the "Inspect Failure ->" button on error rows, trigger of event selection upon clicking the inspect failure button, presence of tooltip attributes, and rendering of the header hint.
- File: web/src/components/AnomalyBadge.test.tsx
  - Add tests verifying acceptance of the optional title prop and default title attribute on clickable anomaly badges.
- File: web/src/views/TicketDeepDive.test.tsx
  - Verify seamless integration when clicking the inspect failure button to open the diagnostic drawer.

---

## 3. Phased Implementation Tasks for Coder Subagent

### Phase 1: Trace Waterfall Component and Anomaly Badge Enhancements

Task 1.1: Refactor Row Interactivity and Class Mapping in web/src/components/TraceWaterfall.tsx
- In the filteredEvents mapping loop, evaluate whether the event is an issue:
  - An event is considered an issue if its status equals 'error' or if a matching anomaly is returned by findMatchingAnomaly.
- Build the className string for the timeline item container:
  - Base class: 'timeline-item'.
  - If the item matches activeSelectedEvent, append 'selected'.
  - If the item has status 'error', append 'item-error'.
  - If the item has a correlated anomaly, append 'item-anomaly'.
  - If the item is an issue (has error or anomaly), append 'item-interactive'.
  - When the item has neither an error nor an anomaly, do not append 'item-interactive', 'item-error', or 'item-anomaly'.

Task 1.2: Add Failure Inspection Affordance in web/src/components/TraceWaterfall.tsx
- For events of type 'tool_call' that have status 'error' (or any row with status 'error'):
  - Render an explicit action button or pill labeled "Inspect Failure ->" inside the tool details group (or adjacent to the status pill).
  - Assign CSS class name 'btn-inspect-failure'.
  - Set the HTML title attribute to "Click to open Tool Execution Failure window".
  - Set an accessible aria-label attribute such as "Inspect failure for turn <turn_number>".
  - Attach an onClick handler that calls stopPropagation on the synthetic event, then invokes handleItemSelect passing the current event and matching anomaly.
- On the error message preview box (.timeline-error-preview), add the HTML title attribute "Click to open Tool Execution Failure window".

Task 1.3: Enhance AnomalyBadge Tooltip and Interactive State
- In web/src/components/AnomalyBadge.tsx:
  - Update AnomalyBadgeProps interface to include an optional title string property.
  - On both the button element (when onClick is provided) and the span element (when onClick is omitted), bind the title attribute.
  - When title is omitted in props, default the title to "Click to open Anomaly Inspector" if onClick is defined, or undefined if non-clickable.
- In web/src/components/TraceWaterfall.tsx:
  - Where AnomalyBadge is rendered in the timeline top row, explicitly pass title="Click to open Anomaly Inspector".

Task 1.4: Add Inline Guidance Hint in web/src/components/TraceWaterfall.tsx
- In the waterfall-header section, adjacent to the event counter or within the filter toolbar:
  - Render a span element with className 'waterfall-hint'.
  - Content: "Click any error or anomaly to inspect failure details".
  - Ensure the hint is positioned gracefully and remains readable across both wide and mobile viewport widths.

### Phase 2: CSS Stylesheet Layout and Transition Refactoring

Task 2.1: Eliminate Layout Shifting and Flashing on .timeline-item in web/src/index.css
- Locate .timeline-item rule:
  - Add baseline border: 1px solid transparent.
  - Set default cursor: default.
  - Remove transition: all 0.15s ease.
  - Declare specific transitions: transition: background-color 0.12s ease, border-color 0.12s ease. Do not transition layout or dimension properties.
- Locate .timeline-item.selected rule:
  - Set border-color: var(--accent-blue).
  - Set background-color: var(--bg-surface-elevated).
  - Ensure no border-width changes occur upon selection.
- Locate .timeline-item:hover rule:
  - Maintain background-color: var(--bg-surface-hover).

Task 2.2: Differentiate Issue and Interactive Row Styles in web/src/index.css
- Define .timeline-item.item-interactive:
  - Set cursor: pointer.
- Define .timeline-item.item-anomaly:
  - Set cursor: pointer.
- Update .timeline-item.item-error:
  - Set cursor: pointer.
  - Set border-left: 3px solid var(--status-red).
  - To prevent horizontal layout jumping when an item has a 3px left border while baseline has 1px, ensure border-left-width is 3px and other borders remain 1px solid transparent (or border-color adjusted accordingly).
  - When .timeline-item.item-error.selected is active, preserve the red left border (border-left-color: var(--status-red)) while applying the accent blue to the top, right, and bottom borders.

Task 2.3: Author Failure Inspection Button Styles in web/src/index.css
- Author .btn-inspect-failure:
  - Display: inline-flex with align-items center.
  - Font styling: monospace font family, font-size 0.75rem, font-weight 600, line-height 1.2.
  - Color styling: text color var(--status-red), background color rgba(239, 68, 68, 0.12), border 1px solid rgba(239, 68, 68, 0.35).
  - Spacing: padding 2px 8px, border-radius 4px.
  - Interactivity: cursor pointer, transition background-color 0.12s ease, border-color 0.12s ease, color 0.12s ease.
- Author .btn-inspect-failure:hover and .btn-inspect-failure:focus:
  - Background color rgba(239, 68, 68, 0.22).
  - Border color var(--status-red).
  - Text color #ffffff.
  - Outline: none.

Task 2.4: Author Waterfall Toolbar Hint Styles in web/src/index.css
- Author .waterfall-hint:
  - Font size: 0.8rem.
  - Text color: var(--text-muted).
  - Font weight: 400.
  - Display: inline-flex with align-items center.

Task 2.5: Enhance AnomalyBadge Interactive Hover State in web/src/index.css
- Locate .anomaly-badge.clickable:hover:
  - Enhance visual feedback: elevated brightness (filter: brightness(1.25)), subtle box-shadow glow matching the severity token (rgba(245, 158, 11, 0.3) for warning, rgba(239, 68, 68, 0.35) for critical).
  - Ensure cursor is pointer.

---

## 4. Adversarial Testing Plan for Tester Subagent

The tester subagent must verify all changes with zero silent skips (zero test.skip, zero it.skip, zero t.Skip).

### Suite 1: TraceWaterfall Component Tests (web/src/components/TraceWaterfall.test.tsx)
1. Cursor and Class Differentiation:
   - Render TraceWaterfall with a mixed list of successful turns, tool error turns, and turns with correlated anomalies.
   - Assert that an unflagged successful turn (Turn 1) has the base timeline-item class, but does NOT possess the item-interactive, item-error, or item-anomaly classes.
   - Assert that a tool error turn has both item-error and item-interactive classes.
   - Assert that a turn with a correlated anomaly has both item-anomaly and item-interactive classes.
2. Selection Stability:
   - Select an unflagged turn. Assert that it receives the selected class without triggering drawer-opening intentions.
   - Verify that clicking an unflagged turn triggers onSelectEvent with undefined anomaly.
3. Inspect Failure Button:
   - On a tool error turn, assert that a button with text "Inspect Failure ->" is present in the DOM.
   - Assert that the button has title="Click to open Tool Execution Failure window".
   - Fire a click event on the "Inspect Failure ->" button. Assert that onSelectEvent is called with the error event and stopPropagation prevents duplicate bubbling.
4. Anomaly Badge Affordance:
   - On an anomalous turn, assert that the rendered AnomalyBadge has the title attribute "Click to open Anomaly Inspector".
5. Waterfall Toolbar Guidance Hint:
   - Assert that visible text "Click any error or anomaly to inspect failure details" is present in the rendered waterfall component.

### Suite 2: AnomalyBadge Component Tests (web/src/components/AnomalyBadge.test.tsx)
1. Default Clickable Title:
   - Render AnomalyBadge with onClick handler provided and no explicit title prop.
   - Assert that the rendered button element has title="Click to open Anomaly Inspector".
2. Custom Title Prop:
   - Render AnomalyBadge with an explicit title prop (e.g. "Custom Inspection Tooltip").
   - Assert that the rendered element possesses the title="Custom Inspection Tooltip".
3. Non-Clickable Badge Title:
   - Render AnomalyBadge without onClick and without title.
   - Assert that the title attribute is not set or empty.

### Suite 3: TicketDeepDive View Integration Tests (web/src/views/TicketDeepDive.test.tsx)
1. Inspect Failure Action Triggers Drawer:
   - Render TicketDeepDive with v2 ticket AL-2.
   - Locate the "Inspect Failure ->" button on an error turn.
   - Fire click event.
   - Assert that the diagnostic drawer opens with the title "Tool Execution Failure" and displays the full error traceback.
2. Anomaly Badge Action Triggers Drawer:
   - Click the anomaly badge on an anomalous turn.
   - Assert that the diagnostic drawer opens with the genuine anomaly title and Rule Diagnosis section.

### Boundary and Quality Gates
- Typecheck: npm run typecheck across web/ must complete with zero errors and zero any.
- Unit and Component Tests: npm test must pass 100% of tests with zero skips.
- Asset Build: npm run build must compile web/dist cleanly without warnings.
- Backend Verification: make check and make test must pass cleanly.

---

## 5. Reviewer Audit Checklist

The reviewer subagent must audit the resulting git diff along two strict axes:

### Axis 1: Spec and Steering Feedback Alignment
- Visual flash / layout jump eliminated: .timeline-item defines a baseline border of 1px solid transparent and transitions only background-color and border-color over 0.12s.
- Transition declaration transition: all 0.15s ease is completely removed from .timeline-item.
- Unflagged non-issue rows have neutral cursor (cursor: default) and do not have item-interactive, item-error, or item-anomaly classes.
- Issue rows (errors or anomalies) have cursor: pointer and item-interactive class.
- Explicit visual cue button "Inspect Failure ->" is rendered on rows with tool execution errors.
- "Inspect Failure ->" button has title="Click to open Tool Execution Failure window" and opens the failure drawer when clicked.
- AnomalyBadge instances have title="Click to open Anomaly Inspector" and interactive hover affordance.
- Trace waterfall header/toolbar displays the subtle inline hint "Click any error or anomaly to inspect failure details".
- Zero silent test skips across all test suites.

### Axis 2: Engineering Standards (TECHNICAL_GUIDELINES.md)
- FCIS boundary discipline: All changes are strictly localized to the TypeScript presentation layer (web/src/components/ and web/src/index.css). Backend Go domain and adapters remain untouched.
- Static typing: Zero occurrences of any or unsafe type assertions across all modified TypeScript files.
- Immutability: Readonly component props and event structures are strictly preserved.
- Zero emojis: Zero emojis permitted across all code, CSS, tests, attributes, and user-facing text strings.
- Single binary parity: web/dist compiles cleanly for embedding into the standalone Go executable.
