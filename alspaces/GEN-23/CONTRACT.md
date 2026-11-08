# GEN-23: Browse the roadmap on GitHub Pages

Publish a small, read-only website whose homepage renders the existing ticket roadmap and whose ticket links open readable specification pages. Markdown remains the source of truth; contributors continue editing the configured workspace's `plan/TICKETS.md` and ticket specifications.

The user approved this scope with “implement the simple tickets.md page” after the recommendation to publish the roadmap, clickable specifications and navigation back to the roadmap, updating automatically after merges to `main`. GitHub Pages is the requested host. The repository is public; there is no existing Pages site to preserve.

Acceptance:

- The homepage contains the current `TICKETS.md` content, including Ready and Blocked sections and their ticket links.
- Ticket Markdown under the configured plan directory renders as static HTML with a clear route back to the roadmap. Links containing encoded spaces work under a GitHub project-site URL.
- Links to supporting files outside the published plan directory remain usable through the repository's GitHub source view. External URLs and document anchors retain their meaning.
- A GitHub Actions workflow validates pull requests and publishes the generated artifact on `main` or explicit workflow dispatch. Deployment permissions remain limited to the deployment job.
- The build reads `workspaces.dir` from shared configuration, falling back to `workspaces` when absent. It generates only the intended documentation site into a separate output directory without changing Markdown sources or publishing telemetry, source sessions or unrelated workspace artifacts.
- The page remains readable on narrow screens and uses ordinary accessible links. It requires no backend or client-side application runtime.

Implementation may use an existing Markdown renderer and YAML parser as build-only dependencies. The earlier Jekyll suggestion is not an acceptance requirement; a small locally verifiable static builder satisfies the same outcome. Search, boards, ticket editing, synchronization with GitHub Issues and changes to the Go/TypeScript application are outside this ticket.

Engineering and verification follow [TECHNICAL_GUIDELINES.md](../../TECHNICAL_GUIDELINES.md). Production hosting activation and deployment must be distinguished from a verified local build; any unavailable remote check remains explicit at handoff.
