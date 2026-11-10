# Presentation-site research

Reviewed on 2026-11-10 using official public pages. These are content and information-structure observations from fetched text, not a visual/browser usability audit or feature comparison. Recommendations below are planning judgments; reuse patterns, not third-party copy, branding or unsupported product claims.

## Useful patterns

- [Langfuse homepage](https://langfuse.com/) leads with product purpose, groups capabilities, links use cases and documentation, and offers a clear starting action. Apply a concise purpose statement, annotated AgenticLens captures and examples tied to decisions. Its cloud service, integrations, evaluation and prompt-management features do not establish AgenticLens capabilities.
- [OpenTelemetry’s product explanation](https://opentelemetry.io/docs/what-is-opentelemetry/) explicitly defines its scope and leaves storage/visualization to other tools. Apply that clarity to “what AgenticLens is and is not.” Its [demo architecture](https://opentelemetry.io/docs/demo/architecture/) connects a diagram to explanatory detail; use a smaller AgenticLens data-flow diagram with a text equivalent, rather than reproducing its service topology.
- [Dagger introduction](https://docs.dagger.io/getting-started/introduction/) moves from purpose to a short “how it works” explanation and installation. Apply the progression to configure/capture, generate, inspect and share. Dagger’s CI execution model is not AgenticLens’s role.
- [GitHub Pages documentation](https://docs.github.com/en/pages/getting-started-with-github-pages/what-is-github-pages) establishes static HTML/CSS/JavaScript hosting and project paths under the repository name. The presentation can be public without a running engine; assets and navigation need to work under `/AgenticLens/`.

## Repository grounding and content direction

- [README](../../README.md): live OpenCode observation, offline OpenCode/Codex generation, anomaly investigation, role/model analytics, comparison and progression views. Existing `docs/images/` captures are candidates, not evidence of current release behavior or publishable data by themselves.
- [Architecture](../../ARCHITECTURE.md) and [technical guidelines](../../TECHNICAL_GUIDELINES.md#5-native-manager-engine--static-asset-bundling): native installation/router manager selects a pinned engine; the engine owns telemetry and embeds the dashboard. Explain source stores and Git feeding adapters/domain analysis, persisted ledgers and CLI/local dashboard. Keep installation/selection separate from telemetry flow.
- [Setup manual](../../docs/SETUP_MANUAL.md): use the approved exact project release pin or latest stable for fresh installation. Link the maintained instructions instead of inventing a parallel installation process.
- [Cost semantics](../../docs/COST_SEMANTICS.md): observed, partial and unavailable costs differ; current Codex extraction does not establish measured USD cost. Missing data must not become zero or a claimed saving.
- [Portable history](../../TECHNICAL_GUIDELINES.md#portable-ledger-contract): shared ledgers remain useful across machines and harnesses when original sources are absent. Local source hydration can add detail without invalidating recorded history.
- Active roadmap items in [the index](../plan/TICKETS.md), including remote multi-project workflows and API-equivalent cost simulation, are not available-feature claims. Compatibility is evidence-bounded, not universal support for arbitrary harnesses or releases.

Recommended examples: investigate repeated tool failures in a run; compare role effort and token/cache use between tickets; inspect delegation and context pressure during OpenCode live work; share recorded session facts in a project handoff. Each example should identify the relevant view and a plausible next human action, without fabricated customer outcomes or numerical productivity claims.
