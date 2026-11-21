## Purpose

Let projects choose whether capture reconciliation details appear alongside their generated metrics.

## ADDED Requirements

### Requirement: Capture observation in Markdown is opt-in

Automatic publication SHALL append the capture observation report to the configured Markdown summary only when the project enables `generation.summary.capture_observation`. The ledger and capture diagnostics SHALL remain available under the existing policies.

#### Scenario: Observation enabled
- **WHEN** the project enables capture observation and automatic publication writes the metrics summary
- **THEN** the generated metrics include the observation for that publication

#### Scenario: Observation omitted or disabled
- **WHEN** the option is absent or false and automatic publication writes a metrics summary, including one with an older observation
- **THEN** normal metrics and authored Markdown remain, and the generated section contains no capture observation
