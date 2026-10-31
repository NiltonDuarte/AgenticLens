# feat(web): add skills and subagents autopsy panel (AL-53)

Ticket Deep Dive now exposes recorded modular skills and subagent delegations in a dedicated collapsible Skills & Subagents card. Operators can inspect invocation turns, roles, source errors, task descriptions, and recorded measurements without searching individual tool spans.

Available child-session actions open and focus the exact execution thread. Missing child threads and unavailable measurements are identified explicitly. Supported v4 ledgers without optional skill/task facts show an empty state, and persisted panel data stays visible when optional trace enrichment fails.

The change is presentation-only and uses existing fact contracts. Tables and controls support keyboard navigation and preserve full source text.

Contract: [CONTRACT.md](CONTRACT.md). Implementation and verification details: [SUMMARY.md](SUMMARY.md).
