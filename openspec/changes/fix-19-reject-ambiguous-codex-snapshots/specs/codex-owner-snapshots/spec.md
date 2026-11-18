## Purpose

Ensure Codex owner facts come from an unambiguous native snapshot while preserving durable accepted history and genuine occurrence multiplicity.

## ADDED Requirements

### Requirement: Ambiguous native owner snapshots are rejected
The system SHALL reject identical or divergent readable copies of one native Codex owner before combining cumulative facts or declaring that owner complete, with explicit ambiguity and incomplete diagnostics and no filename, timestamp or directory authority rule.

#### Scenario: Active and archived copies
- **WHEN** selected input includes JSONL and/or gzip copies of the same native owner, including a warning-bearing copy
- **THEN** generation cannot publish doubled facts and the owner is not certified complete

#### Scenario: Duplicate child relationships
- **WHEN** copies of a native child contain matching or conflicting parent relationships
- **THEN** selection and grouping reject ambiguity rather than resolving a copy by order

### Requirement: Failed source refresh preserves accepted history
The system SHALL retain accepted facts outside successful complete refreshes and keep unaffected owners usable with truthful diagnostics and coverage.

#### Scenario: Refresh or backfill sees ambiguity
- **WHEN** an accepted owner has duplicate source snapshots
- **THEN** its accepted history and unavailable same-harness and other-harness facts survive and the failure reason is reported

#### Scenario: Historical hydration sees ambiguity
- **WHEN** an owner has duplicate source snapshots and other owners have unambiguous available sources
- **THEN** hydration does not choose an arbitrary duplicate, exposes incomplete coverage and keeps unaffected source evidence available without rewriting the ledger

### Requirement: Genuine occurrence multiplicity survives
The system SHALL preserve repeated legitimate invocations, skills and delegations within one unambiguous transcript and preserve exact-discovery ambiguity rejection.

#### Scenario: Stable repeated runs
- **WHEN** unambiguous inputs containing genuine repeated occurrences are consumed repeatedly
- **THEN** occurrence counts remain stable and no broad row deduplication removes legitimate facts
