# GEN-24 — Presentation portal alignment

Status: approved by the user with “$agentic-lens-dev approved”, after removal of the origin story. Source: the user's GEN-24 request, limited to the presentation portal.

AgenticLens should introduce itself as telemetry for teams developing with agents. The portal will lead with understanding skill use, agent and role activity, model use, and context window saturation. Its next emphasis will be project evolution and ticket comparisons: lines of code, recorded cost, token use, cache use, elapsed effort and developer steering. Questions about changes to AGENTS.md, skills, guidance and specifications should frame investigation over time; the copy must not claim that correlations establish causality or promise productivity gains. Anomaly detection remains a supporting diagnostic capability.

The presentation will explain why ledgers live with the Git project: teammates, team leads and engineering managers can inspect recorded history across machines and compare development as tools and practices evolve. Preserve source and cost limitations, meaningful links, accessible navigation, responsive layout, and clear labels distinguishing illustrations from measured results.

The presentation's colors, including its SVG illustrations, will align with the platform's navy surfaces and blue accents in `web/src/index.css`. No platform stylesheet or dashboard behavior changes are included.

An open same-repository PR must have a working public presentation URL beneath `/AgenticLens/preview/<PR-number>/`, before merge. Assets and navigation must work under that prefix. Publishing a preview must preserve the production presentation and roadmap from main and other retained active previews. Updates replace the matching preview; closing a PR removes it on the next successful publication. Publication must serialize shared-site changes and reject stale or unrelated build artifacts. Untrusted PR code must never execute in a job with publishing credentials.

Initial delivery includes bootstrapping this PR's preview using the existing dispatchable workflow and a dedicated, narrowly scoped preview publishing environment; production environment restrictions remain intact. Automatic lifecycle handling becomes available after its workflow reaches the default branch. External-fork preview hosting and any hosting-provider migration are excluded.

Scope includes `.github/pages/`, necessary Pages workflows, preview hosting configuration, and GEN-24 artifacts. Application code, telemetry calculation, ledger schemas, dashboard functionality, roadmap content and other tickets are excluded. Engineering rules remain in [TECHNICAL_GUIDELINES.md](../../TECHNICAL_GUIDELINES.md).
