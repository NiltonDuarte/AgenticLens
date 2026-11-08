# feat(docs): publish the ticket roadmap on GitHub Pages (GEN-23)

Make the existing roadmap and ticket specifications readable as a small static website. Markdown remains the source of truth: the builder reads the configured plan root, preserves ticket links and GitHub section anchors, and provides navigation back to the roadmap. Supporting documents open at the matching GitHub source revision.

The dedicated workflow validates pull requests and publishes only from main with deployment permissions confined to the deploy job. GitHub Pages is enabled for this repository; the first live deployment follows merge.

Validation: six real-renderer fixture tests, a 101-page build with 229 internal links/anchors checked, full integrated `make`, and standalone Linux binary execution. Independent review and the integrated repository gate are recorded in [SUMMARY.md](alspaces/GEN-23/SUMMARY.md).

Contract: [CONTRACT.md](alspaces/GEN-23/CONTRACT.md). Publishing instructions: [TICKETS_SITE.md](docs/TICKETS_SITE.md).
