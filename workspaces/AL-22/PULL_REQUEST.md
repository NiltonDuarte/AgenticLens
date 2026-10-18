# feat(web): Render rich Markdown for workspace summary and contract artifacts (AL-22)

## Summary of Changes
Upgrades the workspace artifact viewer (`web/src/components/ArtifactViewer.tsx`) in the AgenticLens web dashboard from raw monospace code blocks to rich, styled GitHub Flavored Markdown.

- **Presentation Shell (`web/src/components/ArtifactViewer.tsx`)**:
  - Integrated `react-markdown` (v10.1.0) and `remark-gfm` (v4.0.1) for strictly-typed Markdown parsing and rendering.
  - Added view mode toggle control allowing users to switch between "Rich Document" (default) and "Raw Source" views.
  - Added "Copy Content" quick action with temporary "Copied" feedback state (2-second timeout reset, cleanup on unmount, and clipboard fallback for non-secure contexts).
  - Provided custom styled element mappings for:
    - Tables: responsive overflow wrapper (`.markdown-table-wrapper`), bordered cells, elevated header rows, and alternating row striping.
    - Task lists: read-only disabled checkbox inputs (`.markdown-checkbox`) with clean flex alignment.
    - Code: distinct styling for inline code badges (`.markdown-code-inline`) vs preformatted blocks (`.markdown-pre`, `.markdown-code-block`).
    - Blockquotes: accent-bordered callouts (`.markdown-blockquote`).
    - Headings: proportional typographic scale with subtle bottom dividers for major sections.
    - Links: secure external link attributes (`target="_blank"`, `rel="noopener noreferrer"`).
  - Defensive rendering: graceful fallback on null, undefined, empty, or whitespace documents with `.artifact-empty-note` and disabled copy button.
- **Theme & Typography Styling (`web/src/index.css`)**:
  - Added CSS classes for `.artifact-toolbar`, `.view-mode-toggle`, `.artifact-copy-btn`, and complete `.artifact-markdown-container` hierarchy matching the AgenticLens dark theme.
- **Packaging & Single-Binary Build**:
  - Updated `web/package.json` and bundled production assets to `backend/internal/adapters/web/dist/`.
  - Embedded into Go executable via `embed.FS` with zero Node.js runtime dependency in production.

## References
- Architectural Contract: `workspaces/AL-22/CONTRACT.md`
- Iteration Plan: `workspaces/AL-22/iterations/01/PLAN.md`
- Review Audit: `workspaces/AL-22/iterations/01/REVIEW.md`
- Work & Telemetry Summary: `workspaces/AL-22/SUMMARY.md`
