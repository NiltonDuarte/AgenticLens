# docs(architecture): assess telemetry sources and plan corrective tickets (AL-67)

Document actual repository/service boundaries and source injection across OpenCode v1/v2 and Codex, with current/proposed dependency maps, an operation matrix, twelve findings and bounded recommendations.

Create four actionable follow-ups with plain-language examples:
- FIX-18: report unsupported live schemas accurately.
- FIX-19: reject ambiguous Codex owner snapshots.
- FIX-20: cancel native source reads with their callers.
- REV-7: investigate shared read, snapshot and ordering guarantees.

Keep related corrections in AL-70 (configuration and selection tests), AL-69 (child diagnostics) and AL-74 (column-validation oracle). Link every finding to one owner, preserve audit attribution and clarify the accessor's current production use. All fixes remain unimplemented; this PR contains assessment and planning documentation only.

Context: `alspaces/AL-67/ARCHITECTURE_ASSESSMENT.md`, `alspaces/AL-67/CONTRACT.md`, `alspaces/AL-67/SUMMARY.md`, and `alspaces/AL-67/iterations/02/FOLLOW_UP_TICKETS.md`.
