# docs(PO-23): plan the AgenticLens GitHub Pages presentation

Newcomers currently have to combine README screenshots and technical documents to understand AgenticLens and decide whether it fits their workflow.

Create AL-83 for a public GitHub Pages presentation covering purpose and exclusions, illustrated operation, capabilities, project use cases, operating modes, architecture and installation links. Include static publication and accessibility acceptance, preserve source/cost limitations, and separate planned capabilities from available behavior.

[Initiative](https://github.com/NiltonDuarte/AgenticLens/blob/docs/PO-23-github-pages-presentation/alspaces/PO-23/SPEC.md) · [Research](https://github.com/NiltonDuarte/AgenticLens/blob/docs/PO-23-github-pages-presentation/alspaces/PO-23/RESEARCH.md) · [Verification and limits](https://github.com/NiltonDuarte/AgenticLens/blob/docs/PO-23-github-pages-presentation/alspaces/PO-23/SUMMARY.md)

Research compares official Langfuse, OpenTelemetry and Dagger content patterns and GitHub Pages hosting guidance with repository facts. This PR delivers planning only; it creates no application code or website. AL-83 is dependency-ready, with development and publication deferred to its owner.

Validation: planning links, ticket structure and diff checks passed. Full `make` is incomplete: the CLI integration process was stopped at about 150 seconds while a configuration/capture test was active. Other backend integration packages passed; frontend integration and final build targets were not reached. See the summary for exact limits.
