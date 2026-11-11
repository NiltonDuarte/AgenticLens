# AL-83: Present AgenticLens on GitHub Pages

## Goal

Give developers and engineering leads a public page that explains AgenticLens, demonstrates its value in project workflows, and leads to installation and documentation.

## Context

Today the README and technical documents contain the necessary pieces, but newcomers must assemble the story themselves. [PO-23](../../PO-23/SPEC.md) sets the direction following [research into comparable project sites](../../PO-23/RESEARCH.md).

## Acceptance criteria

- A public English GitHub Pages presentation opens with a clear purpose and audience, with working setup, documentation and repository links. The README links to it.
- Explain what it is (local multi-agent development telemetry and observability) and is not (an agent executor/orchestrator, coding assistant, hosted monitoring service, or proof of code quality). Explain that anomaly verdicts and test/production ratios support investigation rather than establish coverage or productivity.
- Show a configure/capture → generate → inspect → share workflow. Distinguish OpenCode live monitoring, OpenCode/Codex offline generation, and historical backfill/migration; state evidence-bounded harness support and link current compatibility information.
- Present anomaly inspection, execution/delegation traces, role/model/token/cache analytics, ticket comparison and progression, and Git-shareable ledgers with captioned product visuals. Use current, publishable captures or clearly labeled illustrative examples; expose no private sessions or credentials and invent no results. Explain observed/partial/unavailable costs rather than implying missing cost is zero.
- Include at least three concrete project use cases, each connecting a problem to a view or metric and a human decision: diagnose tool retries, compare ticket effort, and preserve project handoff history.
- Include a readable architecture/data-flow diagram and text equivalent: harness stores/Git, adapters and pure analysis, durable ledgers, CLI/local embedded dashboard; show native manager/engine selection separately. Explain local execution and history portability across machines/harnesses, without a production Node dependency.
- Publish static content at the repository’s Pages path with correct assets, anchor/deep links, mobile layout, keyboard navigation, headings, contrast and image alternatives. Core explanations remain readable without animation. Page operation requires no local engine, account or private API.
- Provide a reproducible Pages build/deployment path and verify the public URL after deployment. If repository Pages settings/permissions block publication, report the exact prerequisite and leave publication acceptance explicitly incomplete. Link release-aware setup; label any future features as planned.

## Implementation notes

Keep the public presentation separate from the embedded dashboard and its runtime/build contract. No new backend capability, hosted demo, analytics service, custom domain or documentation migration is required. Choose the static-site tooling during development; use the research as guidance rather than a copied design.
