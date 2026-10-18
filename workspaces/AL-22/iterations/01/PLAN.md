# Implementation Plan: AL-22 Render Rich Markdown for Workspace Summary and Contract Artifacts

## 1. Goal and Iteration Scope

Upgrade the web dashboard artifact viewer component in web/src/components/ArtifactViewer.tsx from raw preformatted text blocks to a rich Markdown presentation component with table styling, task list support, view mode toggling, and clipboard copying. This iteration delivers the complete scope of ticket AL-22.

---

## 2. Grounded Reference Architecture and Concrete Paths

The implementation maps to the following existing packages and modules:

### Presentation Layer Components (web/src/components/)
- web/src/components/ArtifactViewer.tsx: Refactor to introduce view mode state, copy feedback state, rich Markdown rendering with custom element mappings, and toolbar controls.
- web/src/components/ArtifactViewer.test.tsx: Dedicated unit and integration tests verifying rendering modes, table structures, task list checkboxes, empty states, and user interactions.

### Presentation Layer Styles (web/src/index.css)
- web/src/index.css: Add styling rules for markdown elements within the artifact viewer, including .artifact-markdown-container, styled table hierarchies, code blocks, blockquotes, view mode toggle buttons, and copy feedback indicators.

### Presentation Package Manifest (web/package.json)
- web/package.json: Add strictly-typed Markdown rendering dependencies (react-markdown and remark-gfm) ensuring compatibility with React 18 and TypeScript 5.

### Regression Test Suite (web/src/components/Components.test.tsx)
- web/src/components/Components.test.tsx: Verify and adjust existing ArtifactViewer test assertions to remain compatible with the enhanced markup.

---

## 3. Phased Implementation Tasks

### Phase 1: Package Dependencies and Setup (web/)
- Task 1.1: Inspect dependency graph in web/package.json.
- Task 1.2: Add react-markdown (version 9.x or matching React 18) and remark-gfm (version 4.x) to web/package.json dependencies.
- Task 1.3: Verify dependency resolution and clean TypeScript compilation with zero type errors.

### Phase 2: Component Architecture and Refactoring (web/src/components/ArtifactViewer.tsx)
- Task 2.1: Define component interfaces in web/src/components/ArtifactViewer.tsx:
  - Define type ArtifactTab as union of 'contract' and 'summary'.
  - Define type ViewMode as union of 'rich' and 'raw'.
  - Define interface ArtifactViewerProps with readonly artifacts property of type TicketArtifacts or null or undefined.
- Task 2.2: Implement local component state:
  - activeArtifactTab initialized to 'contract'.
  - viewMode initialized to 'rich'.
  - copyStatus initialized to boolean false or 'idle'.
- Task 2.3: Implement clipboard copy helper:
  - Extract active document content based on activeArtifactTab.
  - Invoke navigator.clipboard.writeText with fallback for non-secure contexts.
  - Set temporary confirmation state for 2000 milliseconds before reverting.
- Task 2.4: Build toolbar controls in header:
  - Preserve tab buttons for CONTRACT.md and SUMMARY.md.
  - Add segmented button group for switching between Rich Document and Raw Source.
  - Add Copy Content button with dynamic label reflecting copy status.
- Task 2.5: Implement rich Markdown rendering body:
  - When viewMode is 'raw': render standard preformatted code block.
  - When viewMode is 'rich': render ReactMarkdown component configured with remarkGfm plugin.
  - Supply custom component mappings:
    - table: render with responsive scroll wrapper and .markdown-table class.
    - th: render with .markdown-th class.
    - td: render with .markdown-td class.
    - input: render checkboxes with .markdown-checkbox class and disabled attribute.
    - code: distinguish inline code badges from block code wrappers.
    - blockquote: render with .markdown-blockquote class.
    - a: render links with target="_blank" and rel="noopener noreferrer".
- Task 2.6: Defensively guard empty content:
  - Render .artifact-empty-note when content is null, undefined, or empty string.

### Phase 3: Typography and Dark Theme Styling (web/src/index.css)
- Task 3.1: Add toolbar layout styles:
  - Style .artifact-toolbar container to align tabs, view mode controls, and copy action.
  - Style .view-mode-toggle button group with active state highlighting.
  - Style .artifact-copy-btn with subtle border, hover transition, and success accent.
- Task 3.2: Add markdown container styling:
  - Style .artifact-markdown-container with typography matching dashboard font scale.
  - Set spacing for paragraphs, lists, and headings.
- Task 3.3: Add table styling rules:
  - Style .markdown-table-wrapper with overflow-x auto.
  - Style table with border-collapse collapse, width 100%, and margin 12px 0.
  - Style th with text-align left, background var(--bg-surface-elevated), border 1px solid var(--border-color), and padding 8px 12px.
  - Style td with border 1px solid var(--border-color) and padding 8px 12px.
  - Style tr:nth-child(even) with background var(--bg-surface).
- Task 3.4: Add task list and code styling rules:
  - Style .markdown-task-list-item to remove default bullets and align checkbox with label text.
  - Style inline code with background var(--bg-surface-elevated), border 1px solid var(--border-color), and border-radius 4px.
  - Style pre code blocks with background var(--bg-surface-elevated), padding 12px, and rounded corners.

### Phase 4: Automated Testing and Regression Verification (web/src/)
- Task 4.1: Create dedicated unit test suite in web/src/components/ArtifactViewer.test.tsx:
  - Test rich document rendering with headers, paragraphs, and list elements.
  - Test table rendering asserting table, th, and td elements are present in the DOM.
  - Test task lists asserting input checkboxes render with checked and unchecked attributes.
  - Test view mode toggling: click Raw Source button and assert pre and code elements appear; click Rich Document button and assert formatted container returns.
  - Test tab switching between CONTRACT.md and SUMMARY.md.
  - Test copy button action: mock navigator.clipboard.writeText and assert invocation with document string.
  - Test empty state rendering when artifacts prop is null, undefined, or contains empty strings.
- Task 4.2: Update web/src/components/Components.test.tsx to ensure existing smoke tests remain green.

### Phase 5: Build Verification and Quality Gates
- Task 5.1: Run TypeScript typechecker: npm run typecheck in web/. Assert zero errors.
- Task 5.2: Run ESLint: npm run lint in web/. Assert zero warnings.
- Task 5.3: Run frontend test suite: npm test in web/. Assert all tests pass.
- Task 5.4: Run repository build: make build from project root. Verify single-binary compilation with embedded web assets.
- Task 5.5: Verify zero emojis across all modified files, CSS class names, test files, and UI labels.

---

## 4. Execution Dependencies and Delivery Sequence

1. Phase 1 (Package Dependencies) -> unblocks Phase 2 component imports.
2. Phase 2 (ArtifactViewer Refactoring) -> unblocks Phase 3 styling classes and Phase 4 test assertions.
3. Phase 3 (CSS Theme Integration) -> establishes visual parity with dashboard design system.
4. Phase 4 (Test Suites) -> validates functional correctness, accessibility attributes, and regressions.
5. Phase 5 (Quality Gates & Single Binary) -> verifies production readiness and complete ticket delivery.
