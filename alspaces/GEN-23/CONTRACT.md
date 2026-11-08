# GEN-23: Browse the roadmap on GitHub Pages

Publish a small, read-only website with a site-root welcome page linking to `/roadmap/`, where the existing ticket roadmap links to readable specification pages. Markdown remains the source of truth; contributors continue editing the configured workspace's `plan/TICKETS.md` and ticket specifications.

The user approved this scope with “implement the simple tickets.md page” after the recommendation to publish the roadmap, clickable specifications and navigation back to the roadmap, updating automatically after merges to `main`. GitHub Pages is the requested host. The repository is public. GitHub Pages is configured for workflow publication; PR #131 delivered the preceding implementation and is merged; the welcome-page change requires a new branch from current `main` and a new pull request.

Acceptance:

- The site root displays a simple welcome page with an ordinary link to `/roadmap/`, respecting the GitHub project-site prefix. Only the trusted site scaffold may produce this root index; ticket front matter must not escape `/roadmap/` or overwrite the welcome page.
- The `/roadmap/` page contains the current `TICKETS.md` content, including Ready and Blocked sections and their ticket links. For this project the public URL is `/AgenticLens/roadmap/`; ticket pages and back navigation stay beneath that route.
- Remove the ticket website/roadmap/publishing sentence previously added to `README.md`; no README link is part of this change or the welcome-page addition.
- Ticket Markdown under the configured plan directory renders as static HTML with a clear route back to the roadmap. Links containing encoded spaces work under a GitHub project-site URL.
- Links to supporting files outside the published plan directory remain usable through the repository's GitHub source view. External URLs and document anchors retain their meaning.
- A GitHub Actions workflow validates pull requests and publishes the generated artifact on `main` or explicit workflow dispatch. Deployment permissions remain limited to the deployment job.
- The build reads `workspaces.dir` from shared configuration, falling back to `workspaces` when absent. It generates only the intended documentation site into a separate output directory without changing Markdown sources or publishing telemetry, source sessions or unrelated workspace artifacts.
- The page remains readable on narrow screens and uses ordinary accessible links. It requires no backend or client-side application runtime.

Implementation uses native Jekyll and established GitHub Pages plugins as build-only dependencies. The user superseded the Python implementation choice with “if we can get rid of this python and use the native Jekyll I prefer”; remove the custom Python builder, its dependency file and its builder-specific tests. Keep custom integration limited to configuration/staging and any small Jekyll hook required to preserve supporting-file links; do not replace the renderer with another bespoke generator. The user subsequently requested a root welcome page linking to the roadmap and a new branch from `main` because the preceding work had merged. This supersedes the earlier prohibition on a root index only for that trusted welcome page. The roadmap route, ticket publication boundaries and README exclusion remain accepted. Search, boards, ticket editing, synchronization with GitHub Issues and changes to the Go/TypeScript application are outside this ticket.

Engineering and verification follow [TECHNICAL_GUIDELINES.md](../../TECHNICAL_GUIDELINES.md). Production hosting activation and deployment must be distinguished from a verified local build; any unavailable remote check remains explicit at handoff.
