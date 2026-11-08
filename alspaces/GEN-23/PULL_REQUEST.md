# feat(docs): publish the ticket roadmap with Jekyll (GEN-23)

Publish the existing ticket roadmap and specifications at `/AgenticLens/roadmap/`. Native Jekyll renders the Markdown through standard optional-front-matter and relative-link plugins, with a shared readable layout and navigation back to the roadmap. Supporting documents open at the matching GitHub source revision.

The staging adapter reads the configured workspace, includes only plan Markdown and presentation files, and validates that generated pages stay under `roadmap/`. The workflow tests pull requests and deploys only main, with Pages/OIDC permissions confined to deployment. Ruby is needed only during the build; the published site is static.

Validation: seven native Jekyll tests with 109 assertions, a real 101-page build, 229 internal link/anchor checks, full integrated `make`, and standalone Linux binary execution. The integrated repository gate and retained review findings are recorded in [SUMMARY.md](alspaces/GEN-23/SUMMARY.md).

The new route becomes live after merge and successful deployment. [Publishing guide](docs/TICKETS_SITE.md); [contract](alspaces/GEN-23/CONTRACT.md).
