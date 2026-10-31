# chore(management): supervise ticket PRs with human-only releases (MGT-1)

Add the MGT management-ticket category and reusable agentic-lens-management skill. The manager maps existing independent ticket threads, sends authorized review/integration instructions, and serializes non-release merges behind exact-head verification. Release and autorelease PRs remain strictly human controlled.

Move the original review from GEN-14 to MGT-1 and preserve its evidence. The review covers cross-PR navigation/trace conflicts, two semantic-warning defects in #78, planning dependency reconciliation, and release readiness. Current supervision progress, actual merge SHAs and blockers are in CONTROL.md; managed ticket telemetry remains separate.

Management guidance/configuration/artifacts only. Each product owner retains implementation and its existing development workflow. See iterations/03/PLAN.md for the user-authorized scope and iterations/03/REVIEW.md for verification. Supervision is ongoing until every eligible non-release PR is verified merged; release PRs are excluded.
