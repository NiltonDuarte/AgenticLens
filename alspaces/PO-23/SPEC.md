# AgenticLens project presentation on GitHub Pages

## Goal

Help developers and engineering leads decide whether AgenticLens fits their projects, understand how it works, and reach a trustworthy installation path from a public presentation page.

## Context

The repository README already describes telemetry and includes dashboard captures, but newcomers must assemble the product story from several technical documents. A visitor should be able to move from “What does this do for my coding workflow?” to a concrete investigation example and then setup, without installing the dashboard first.

The intended audience is individual developers adopting coding agents and teams reviewing multi-agent engineering work. The page should communicate the value of local observation and durable project history, while explaining the evidence limits that matter to adoption. Its public content must not depend on access to a visitor’s repository or session stores.

## Expected behavior and acceptance

[AL-83](../plan/to-do/AL-83%20present%20agenticlens%20on%20github%20pages.md) owns the observable acceptance criteria for this single deliverable: an English project page and its GitHub Pages publication path. The planned reading sequence is product purpose and boundaries, illustrated workflow, features and practical uses, operating modes, architecture, then setup and deeper documentation. This order is a content direction, not a prescribed visual layout or framework.

The evidence-based positioning is a local telemetry and observability tool for multi-agent software development. OpenCode live observation and OpenCode/Codex retrospective generation are distinct capabilities. A static public page must not imply it is running the local engine or observing the visitor’s sessions. The product story should explain how humans can use metrics to investigate effort and execution problems without promising automatic quality assessment or guaranteed savings.

## Delivery

One independently useful ticket covers content, presentation and static publication. No application feature dependency is required: unimplemented roadmap items can be omitted. AL-63’s pending source-capability audit is useful future evidence, not a prerequisite or permission to advertise broader support. Developer implementation and publication happen in a separate session.

## Implementation notes

[Research](RESEARCH.md) records primary-source examples and the rationale for the content direction. Repository facts come from README, architecture, setup and cost documentation at `f63e3db6`. Recheck those facts and the published release at implementation time. Preserve the canonical [native architecture](../../TECHNICAL_GUIDELINES.md#5-native-manager-engine--static-asset-bundling) and [portable ledger contract](../../TECHNICAL_GUIDELINES.md#portable-ledger-contract); this initiative changes their explanation, not their behavior.
