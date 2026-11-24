## ADDED Requirements

### Requirement: Pages publishes only the main portal

The Pages workflow SHALL publish only the current main revision, on a push to main or a manual dispatch from main. Pull request events SHALL NOT run it. A main publication SHALL drop any PR previews retained from earlier publications after validating the retained state.

#### Scenario: A pull request is opened or updated

- **WHEN** a pull request targeting main is opened, updated, edited, reopened or closed
- **THEN** the Pages workflow does not run and the published site is unchanged

#### Scenario: A dispatch from another ref

- **WHEN** the workflow is dispatched from a ref other than main
- **THEN** it fails before building or changing the publication state

#### Scenario: Retained previews from earlier publications

- **WHEN** main publishes over a retained state that lists PR previews
- **THEN** the published site and the retained state contain only the main portal, and an invalid retained state still fails the publication before any change

### Requirement: Bounded readability wait and no re-run publication

Before recording the pending transaction, the publisher SHALL read the same Pages artifact up to three times over about 80 seconds without uploading another artifact. The publish job SHALL refuse a re-run attempt before it reads or changes the publication state.

#### Scenario: The artifact is briefly unreadable

- **WHEN** the first read of a finalized Pages artifact fails and a later bounded read succeeds
- **THEN** the publisher records the transaction and deploys

#### Scenario: The artifact stays unreadable

- **WHEN** every bounded read fails
- **THEN** the publisher fails before recording a pending transaction, and a new dispatch can publish from the previous state

#### Scenario: A failed publication is re-run

- **WHEN** an operator re-runs the publish job
- **THEN** it fails before loading the state, uploading an artifact or recording a pending transaction
