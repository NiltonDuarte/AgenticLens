## ADDED Requirements

### Requirement: Readiness before recording a Pages deployment

The Pages publisher SHALL confirm it can read the named Pages artifact from the current run before recording the pending deployment transaction. It SHALL record that transaction before invoking deployment and retain the composed site only after deployment succeeds.

#### Scenario: The uploaded artifact cannot be read

- **WHEN** the Pages artifact upload or the pre-deployment read fails
- **THEN** the publisher fails before recording a pending transaction and a subsequent publication can start from the previous durable state

#### Scenario: The uploaded artifact is readable

- **WHEN** the Pages artifact is readable
- **THEN** the publisher records the transaction before invoking deployment and retains the successfully published composition after deployment
