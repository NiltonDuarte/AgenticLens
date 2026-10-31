# AL-54: Fleet-Wide Skill Adoption Analytics and Comparison Matrix Filter

- **Subsystem**: Presentation / Web / Fleet / Comparisons / Domain / Tests
- **Phase**: Phase 8 (v2 Observability / Skills)
- **Status**: Ready
- **Blocked by**: None
- **Blocks**: AL-55

## Planning status (2026-10-31)

AL-51 merged in [PR #62](https://github.com/NiltonDuarte/AgenticLens/pull/62), providing optional skill/task facts and live attributes. Its dependency is satisfied; this presentation work remains open. Supported v4 ledgers may omit those optional fields; ordinary runtime rejects v1-v3 ledgers until explicit migration and source-backed backfill complete.

## Context
Engineering managers and lead architects need visibility into modular skill adoption across all tickets: Which skills are leveraged most frequently? Which skills fail or trigger anomalies? Which tickets succeeded without using any specialized skills?

## Goal
Implement pure domain aggregations for skill usage statistics across ticket ledgers, surface a fleet-wide adoption metric card in Fleet Overview, and provide a "Skills Used" filter and column in the Ticket Comparison view.

## Scope Across Layers

### 1. Pure Functional Domain (`internal/domain/aggregations.go`, `web/src/domain/aggregations.ts`)
- Define `SkillAdoptionFact`:
  ```typescript
  export interface SkillAdoptionFact {
    name: string;
    total_invocations: number;
    tickets_count: number;
    failure_count: number;
    failure_rate: number;
  }
  ```
- Implement pure aggregation function:
  ```go
  func ComputeSkillAdoption(ledgers []AgenticFactLedger) []SkillAdoptionFact
  ```
- Defensive invariant: gracefully skip or treat unpopulated `Skills` slices as empty, ensuring legacy ledgers do not cause nil panics or runtime exceptions.

### 2. Fleet Overview View (`web/src/views/FleetOverview.tsx`)
- Add a "Skill Adoption" summary card displaying:
  - Total unique skills utilized across the fleet.
  - Top 3 most frequently invoked skills.
  - Overall skill success rate.

### 3. Ticket Comparisons View (`web/src/views/TicketComparisons.tsx`)
- Add "Skills Used" column to the comparative ticket table.
- Add "Filter by Skill" multi-select dropdown in comparison controls.
- Update scatter plot tooltips to include list of skills leveraged by each ticket.

### 4. Automated Tests
- Domain aggregation tests in Go (`internal/domain/aggregations_test.go`) and TypeScript (`web/src/domain/aggregations.test.ts`).
- Vitest tests in `web/src/views/FleetOverview.test.tsx` and `web/src/views/TicketComparisons.test.tsx`.

## Acceptance Criteria
1. Pure aggregation accurately computes invocation counts, ticket counts, and failure rates per skill.
2. Fleet Overview displays top skills and fleet-wide adoption metrics.
3. Ticket Comparisons table allows filtering tickets by specific skill names.
4. Legacy ledgers without skill data are handled safely without breaking fleet metrics.

## Definition of Done
- Pure aggregation functions covered with unit tests across empty, legacy, and modern ledgers.
- View components tested with Vitest.
- All Go and TypeScript tests pass (`make check`).
- Zero emojis across all code, tests, and documentation.
