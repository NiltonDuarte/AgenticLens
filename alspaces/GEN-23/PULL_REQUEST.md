# feat(docs): add a welcome page linking to the roadmap (GEN-23)

Add a small welcome page at the GitHub Pages project root, `/AgenticLens/`, with a visible link to `/AgenticLens/roadmap/`. It reuses the existing Jekyll layout and styling.

The build stages only the trusted welcome scaffold alongside the roadmap. Ticket pages remain confined to roadmap destinations, so front matter cannot overwrite the welcome page. The README is unchanged.

Validation: seven native Jekyll tests with 133 assertions; a 102-page real build, 232 internal link/anchor checks, full integrated `make`, and standalone Linux binary execution. The overwrite regression fails when the provenance guard is removed in a disposable copy. Integrated checks and review evidence are recorded in [SUMMARY.md](alspaces/GEN-23/SUMMARY.md).

The welcome page becomes live after merge and successful Pages deployment.
