h3. Title: AL-22: Render Rich Markdown for Workspace Summary and Contract Artifacts

*TARGET*: AgenticLens (presentation | web | tests)

*CONTEXT*

Currently, the workspace artifacts viewer (`web/src/components/ArtifactViewer.tsx`) renders the contents of `CONTRACT.md` and `SUMMARY.md` inside raw `<pre><code>` blocks. This forces users to read unformatted markdown strings, making telemetry summary tables, task checkboxes, architectural headers, and section links difficult to navigate and parse visually.

*GOAL*

Upgrade the TypeScript web dashboard artifact viewer to render `CONTRACT.md` and `SUMMARY.md` as rich, formatted Markdown with styled tables, task lists, headers, code blocks with syntax highlighting, and copyable snippets, replacing raw code block strings.

*VERTICAL SLICE SCOPE*

* Presentation & Components (`web/src/components/`):
  - Integrate a strictly-typed Markdown parser/renderer (e.g. `react-markdown` with `remark-gfm`, or a dedicated lightweight typed Markdown component) with zero `any` dependencies.
  - Refactor `ArtifactViewer.tsx`:
    - Replace the `<pre className="artifact-pre"><code>{currentContent}</code></pre>` wrapper with the rich Markdown renderer.
    - Provide custom styled component renderers for tables (`table`, `th`, `td`), headings (`h1`-`h4`), code blocks (`pre`, `code`), blockquotes, and task checkboxes (`input[type="checkbox"]`).
    - Add a toggle button allowing users to switch between "Rich Document" and "Raw Source" views.
    - Add a "Copy Content" quick-action button with temporary copied feedback.
* Styling (`web/src/index.css`):
  - Add markdown typography styling matching the AgenticLens dark/light theme (table borders, cell padding, alternate row shading, blockquote accents, code badge styling).
* Quality:
  - Vitest component tests verifying that markdown headings, tables, task lists, and view mode toggling render correctly.
  - Verification that empty or malformed markdown strings do not crash the view.
  - Strict TypeScript check ensuring zero `any`.

*ACCEPTANCE CRITERIA*

* `CONTRACT.md` and `SUMMARY.md` render as styled rich Markdown rather than raw monospace code blocks.
* Tables (including the AgenticLens summary metrics table) render with formatted columns, aligned headers, and bordered cells.
* Task lists render with interactive/styled checkboxes.
* Users can toggle between "Rich Document" and "Raw Source" mode.
* The "Copy Content" action copies the document text to clipboard.
* All web tests pass (`npm test`) and type check passes with zero errors (`npm run typecheck`).

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make test-frontend`, `npm run typecheck`, `npm run lint`).
* Single-binary build verified with `make build`.
* Zero any and zero emojis across all frontend files.
