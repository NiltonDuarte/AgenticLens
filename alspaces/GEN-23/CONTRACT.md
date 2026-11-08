# GEN-23: Browse the roadmap on GitHub Pages

Publish a small, read-only website whose `/roadmap/` page renders the existing ticket roadmap and whose ticket links open readable specification pages. Markdown remains the source of truth; contributors continue editing the configured workspace's `plan/TICKETS.md` and ticket specifications.

The user approved this scope with “implement the simple tickets.md page” after the recommendation to publish the roadmap, clickable specifications and navigation back to the roadmap, updating automatically after merges to `main`. GitHub Pages is the requested host. The repository is public. GitHub Pages is configured for workflow publication; the existing GEN-23 pull request remains the delivery vehicle.

Acceptance:

- The `/roadmap/` page contains the current `TICKETS.md` content, including Ready and Blocked sections and their ticket links. For this project the public URL is `/AgenticLens/roadmap/`; ticket pages and back navigation stay beneath that route. Do not introduce a site-root homepage.
- Remove the ticket website/roadmap/publishing sentence previously added to `README.md`; no README link is part of this change.
- Ticket Markdown under the configured plan directory renders as static HTML with a clear route back to the roadmap. Links containing encoded spaces work under a GitHub project-site URL.
- Links to supporting files outside the published plan directory remain usable through the repository's GitHub source view. External URLs and document anchors retain their meaning.
- A GitHub Actions workflow validates pull requests and publishes the generated artifact on `main` or explicit workflow dispatch. Deployment permissions remain limited to the deployment job.
- The build reads `workspaces.dir` from shared configuration, falling back to `workspaces` when absent. It generates only the intended documentation site into a separate output directory without changing Markdown sources or publishing telemetry, source sessions or unrelated workspace artifacts.
- The page remains readable on narrow screens and uses ordinary accessible links. It requires no backend or client-side application runtime.

Implementation uses native Jekyll and established GitHub Pages plugins as build-only dependencies. The user superseded the Python implementation choice with “if we can get rid of this python and use the native Jekyll I prefer”; remove the custom Python builder, its dependency file and its builder-specific tests. Keep custom integration limited to configuration/staging and any small Jekyll hook required to preserve supporting-file links; do not replace the renderer with another bespoke generator. The website behavior and publication boundaries above remain accepted. Search, boards, ticket editing, synchronization with GitHub Issues and changes to the Go/TypeScript application are outside this ticket.

Engineering and verification follow [TECHNICAL_GUIDELINES.md](../../TECHNICAL_GUIDELINES.md). Production hosting activation and deployment must be distinguished from a verified local build; any unavailable remote check remains explicit at handoff.
