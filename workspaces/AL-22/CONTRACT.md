# Contract: AL-22 Render Rich Markdown for Workspace Summary and Contract Artifacts

## 1. Problem Statement

In the AgenticLens web dashboard, the artifact viewer component located at web/src/components/ArtifactViewer.tsx renders workspace deliverables such as CONTRACT.md and SUMMARY.md using an unadorned preformatted code block.

This presentation has distinct operational flaws:
1. Architectural sections, tables, lists, and headings are displayed as flat raw text.
2. Formatted telemetry tables generated in SUMMARY.md are difficult to inspect without visual column delineation, header styling, and cell borders.
3. Task checkboxes and acceptance lists lack visual distinction between completed and open items.
4. Developers and reviewers cannot toggle between formatted reading mode and raw markdown source.
5. Users have no quick mechanism to copy the document contents to clipboard for sharing or external auditing.

Ticket AL-22 resolves these issues by upgrading web/src/components/ArtifactViewer.tsx into a rich Markdown presentation component. It renders GitHub Flavored Markdown with styled tables, headings, blockquotes, code blocks, and task lists, provides a view toggle between Rich Document and Raw Source, and adds a clipboard copy action, while maintaining strict typing and zero any.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All changes must satisfy TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. Functional Core:
   - Backend domain logic in internal/domain/ remains untouched.
   - Frontend pure domain types in web/src/domain/types.ts define the immutable data shapes for artifacts and view models.
   - Text parsing and sanitization utilities remain pure functions without DOM side effects.
2. Imperative Shell:
   - Component rendering in web/src/components/ArtifactViewer.tsx handles DOM events, clipboard interaction, and active tab or mode state.
   - Theme styling resides in web/src/index.css using existing CSS custom properties.
   - Zero runtime network side effects or external unbundled CDN asset loading.

### Strict Typing and Zero Any
1. All component props, internal states, and renderer callbacks must use concrete, strictly-typed interfaces.
2. Zero any or unknown casts in component code or test files.
3. View mode states must use a strict Discriminated Union or typed union of literals: 'rich' or 'raw'.
4. Active artifact tab states must use a typed union of literals: 'contract' or 'summary'.

### Single Binary Distribution and Offline Independence
1. All parsing and rendering dependencies must compile cleanly through Vite and TypeScript without requiring Node.js runtime services at presentation time.
2. Build output in web/dist/ must embed directly into the Go executable via embed.FS.
3. Zero external fonts, scripts, or unbundled stylesheets fetched from third-party CDNs.

### Zero-Emoji Mandate
1. Zero emojis across all source files, component templates, button labels, test assertions, CSS classes, and workspace markdown artifacts.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Presentation & Components (web/src/components/):
   - Integrate a strictly typed Markdown parser and renderer supporting GitHub Flavored Markdown (tables, task lists, strikethrough, autolinks).
   - Refactor ArtifactViewer.tsx to render rich document elements by default.
   - Implement custom sub-renderers for tables, headings, code blocks, blockquotes, and checkboxes matching the AgenticLens dark theme.
   - Implement a segmented view mode control allowing users to switch between Rich Document and Raw Source.
   - Implement a Copy Content action button with timed visual feedback (Copied).
   - Defensively handle null, undefined, or empty markdown strings with a clean empty-state placeholder.
2. Styling & Theme System (web/src/index.css):
   - Define CSS rules for markdown typography inside .artifact-body.
   - Style tables with border collapse, padding, alternating row backgrounds, and sticky or distinct header rows.
   - Style headings with proportional hierarchy and subtle bottom dividers for primary sections.
   - Style code blocks and inline code badges using existing monospace variables.
   - Style task list items and disabled checkbox inputs cleanly.
   - Style the view mode switch and copy action button in the artifact header toolbar.
3. Quality & Test Suites (web/src/components/):
   - Unit tests in web/src/components/ArtifactViewer.test.tsx asserting:
     - Correct rendering of headings, paragraphs, and lists.
     - Formatted table rendering with table header and data cell elements.
     - Task list rendering with checkbox elements.
     - Toggling between Rich Document and Raw Source views.
     - Correct behavior when artifact content is null, empty, or whitespace.
     - Copy button interaction and clipboard write invocation.
   - Strict TypeScript validation (npm run typecheck) with zero errors.
   - Linting validation (npm run lint) with zero warnings.

### Out of Scope
1. Backend changes: No modifications to Go domain models, repository scanners, or REST API endpoints.
2. Bidirectional editing: Artifact viewer remains strictly read-only; in-browser document editing or saving is out of scope.
3. Heavy syntax highlighter engines: Large language bundles that bloat the web bundle are avoided in favor of lightweight theme-consistent token styling.

---

## 4. Component Architecture and Behavioral Specification

### Component: ArtifactViewer
1. Inputs:
   - artifacts: TicketArtifacts or null or undefined.
2. State:
   - activeArtifactTab: 'contract' or 'summary', defaulting to 'contract'.
   - viewMode: 'rich' or 'raw', defaulting to 'rich'.
   - copyFeedback: boolean, true for two seconds following a successful clipboard copy.
3. Header Toolbar Controls:
   - Section Title: Workspace Artifacts.
   - Artifact Switcher Tabs: Buttons for CONTRACT.md and SUMMARY.md, displaying an Empty suffix when corresponding content is missing.
   - View Mode Switcher: Segmented button group toggling between Rich Document and Raw Source.
   - Copy Button: Action button labeled Copy Content, switching to Copied upon successful copy.
4. Content Container:
   - When active content is missing or whitespace: render .artifact-empty-note stating that the document was not found.
   - When viewMode is 'raw': render the existing preformatted block with .artifact-pre and code elements.
   - When viewMode is 'rich': render the markdown tree inside .artifact-markdown-container.

### Markdown Rendering Rules
1. Headings:
   - Render levels 1 through 6 with corresponding typography classes.
   - Level 1 and 2 headings include a subtle bottom border using var(--border-color).
2. Tables:
   - Render table elements with responsive overflow container.
   - Headers styled with background var(--bg-surface-elevated) and font-weight 600.
   - Alternating row backgrounds using var(--bg-surface) and var(--bg-primary).
   - Cell borders using var(--border-color) and padding matching dashboard density.
3. Task Lists:
   - List items containing checkbox inputs rendered with custom spacing and read-only checked or unchecked attributes.
4. Code Blocks and Inline Code:
   - Code blocks wrapped in bordered containers with background var(--bg-surface-elevated).
   - Inline code wrapped in monospace badges with padding and subtle border.
5. Blockquotes:
   - Left-bordered with var(--accent-blue) and muted text color var(--text-secondary).

---

## 5. Acceptance Criteria

1. CONTRACT.md and SUMMARY.md render as formatted rich Markdown by default.
2. Markdown tables, including telemetry and metric rollup tables, render with styled headers and borders.
3. Markdown task lists render with checkboxes reflecting checked or unchecked states.
4. Users can toggle between Rich Document and Raw Source modes via a dedicated header control.
5. The Copy Content button copies the currently active document text to the system clipboard.
6. Empty or missing artifacts display the empty notice without crashing or throwing errors.
7. TypeScript compilation passes with zero errors under strict mode.
8. Vitest tests cover rich rendering, raw mode toggling, empty states, and copy actions.
9. Single-binary build via make build succeeds with embedded web assets.
10. Strict zero-emoji policy maintained across code, styling, UI labels, and documentation.
