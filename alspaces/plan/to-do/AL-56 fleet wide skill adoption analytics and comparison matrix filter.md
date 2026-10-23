h3. Title: AL-56: Fleet-Wide Skill Adoption Analytics and Comparison Matrix Filter

*TARGET*: AgenticLens (domain / aggregations / presentation / web / fleet / comparisons / tests)

*CONTEXT*
 
Engineering leadership lacks macro-level visibility into skill adoption across projects and workspaces. Currently, neither `FleetOverview.tsx` nor `TicketComparisons.tsx` surfaces which skills are most frequently used, what failure rates skills experience across tickets, or how skill usage correlates with delivery velocity and code churn.

*GOAL*
 
Introduce domain aggregations for fleet-wide skill statistics, display a "Skill Adoption" breakdown in Fleet Overview, and provide a "Skills Used" filter and table column in the Ticket Comparisons matrix.

*VERTICAL SLICE SCOPE*
 
* Domain Core / Aggregations:
  - Add `SkillUsageSummary` struct to `backend/internal/domain/aggregations.go` aggregating: `Name string`, `TicketsCount int64`, `TotalCalls int64`, `TotalErrors int64`, `ErrorRate float64`.
  - Update `BuildFleetOverview` to compute fleet-wide skill usage summaries from scanned ticket workspaces.
* HTTP Server & DTO:
  - Expose `skills_usage []SkillUsageSummary` in `/api/fleet` JSON payload.
* Presentation / Fleet Overview:
  - Add a "Skill Adoption" card in `FleetOverview.tsx` showing top invoked skills, adoption percentage across tickets, and error rates.
* Presentation / Ticket Comparisons:
  - Add a "Skills Used" column in the comparison matrix table.
  - Add a multi-select dropdown filter (`SkillsFilter`) allowing operators to filter comparison rows by loaded skills.
* Quality:
  - Domain tests in `aggregations_test.go` verifying calculation of skill adoption summaries across multi-ticket fixtures.
  - Component tests in `FleetOverview.test.tsx` and `TicketComparisons.test.tsx` asserting skill card rendering and filter operations.

*ACCEPTANCE CRITERIA*
 
* `BuildFleetOverview` computes accurate skill usage summaries across all valid workspace ledgers.
* Fleet Overview displays top skills with adoption metrics and error rates.
* Ticket Comparison view allows filtering rows by one or more skills.
* Comparison table displays comma-separated or chip-based skill indicators per ticket row.
* `npm test`, `npm run typecheck`, and `go test ./...` pass cleanly.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
