# Source capability evidence

## ADDED Requirements

### Requirement: Versioned cross-harness capability report

The report SHALL distinguish observed support, demonstrated failure, unsupported behavior, unavailable evidence and unexecuted checks for every source and capability in the [roadmap acceptance](../../../../../alspaces/plan/done/AL-63%20verify%20source%20capabilities%20across%20harnesses.md).

#### Scenario: Selecting a source for a historical or live trace

- **WHEN** an operator checks OpenCode v1, OpenCode v2 or Codex
- **THEN** the matrix identifies its ingestion, hydration, live access, skills, delegations, child navigation, execution errors and recovery behavior with source layout, version and evidence provenance.

#### Scenario: Enrichment or recovery cannot read an owner

- **WHEN** records are absent, malformed or readable for only some owners
- **THEN** the report gives the observed coverage, diagnostics and retained facts, including exact identities, occurrence growth and repeated refresh, and identifies the cases supported only by fixtures.
