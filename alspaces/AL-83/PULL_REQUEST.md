# feat(AL-83): present AgenticLens on GitHub Pages

Replace the minimal welcome page with an English presentation for developers and engineering leads. Explain local telemetry, investigation workflows, practical use cases, operating modes and architecture, with clearly labeled illustrative product visuals and release-aware setup links. Preserve the existing roadmap and single native Jekyll deployment path.

The presentation remains separate from the embedded local dashboard and publishes no session sources or telemetry payloads. See the [contract](CONTRACT.md) and [summary](SUMMARY.md) for boundaries and evidence.

Local validation passed: 8 Jekyll tests/335 assertions; project/root builds and all internal links; responsive keyboard/asset checks and zero axe violations at four viewport widths. Integrated make and final consolidation are pending. Public deployment remains incomplete until this change is merged and the main Pages workflow deploys successfully; this PR does not authorize merging.
