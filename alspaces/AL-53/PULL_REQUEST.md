# feat(web): add skills and subagents autopsy panel (AL-53)

Integrated the verified AL-47 startup correction from PR #82. Combined navigation review remains valid; manager retains final merge authority. Historical integration evidence is preserved in SUMMARY.md and iterations/01/REVIEW.md.


Ticket Deep Dive now exposes recorded modular skills and subagent delegations in a dedicated collapsible Skills & Subagents card. Operators can inspect invocation turns, roles, source errors, task descriptions, and recorded measurements without searching individual tool spans.

Waterfall badges and delegation rows share a single child-session request contract. Available child-session actions open and focus the exact execution thread, including repeated activation. Missing child threads and unavailable measurements are identified explicitly. Supported v4 ledgers without optional skill/task facts show an empty state, and persisted panel data stays visible when optional trace enrichment fails.

Integrated with current main after AL-47, AL-54 and AL-52; their live refresh, reading anchors and fleet adoption/filter behavior are preserved. AL-55 now has no blockers and is Ready; its remaining scope is unchanged.

The change is presentation-only and uses existing fact contracts. Tables and controls support keyboard navigation and preserve full source text.

Contract: [CONTRACT.md](CONTRACT.md). Implementation and verification details: [SUMMARY.md](SUMMARY.md).
