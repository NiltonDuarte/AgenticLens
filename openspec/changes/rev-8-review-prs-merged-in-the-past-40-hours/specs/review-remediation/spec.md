## ADDED Requirements

### Requirement: Publish against the selected workspace directory

Generation, capture publication, backfill, and migration SHALL preserve each selected workspace’s physical directory through write admission.

#### Scenario: The selected ledger directory is replaced before admission

- **WHEN** the directory selected for a write moves before admission and its former path names a distinct directory, including a same-ticket copy
- **THEN** the writer reports a retryable location change before recovering or writing any replacement artifact. Batch operations retain valid sibling work and report the changed target as incomplete.

#### Scenario: A moved or copied ledger is freshly selected

- **WHEN** an operator selects a supported ledger at its current location and some prior session sources are unavailable
- **THEN** normal admission and refresh or migration continue with established source ownership, recovery rules, and preservation of recorded facts.

### Requirement: Recover a saturated live stream

The live stream SHALL recover an existing client that misses events because of subscriber backpressure.

#### Scenario: A slow subscriber misses source recovery

- **WHEN** subscriber backpressure prevents delivery of a successful read after a source failure
- **THEN** the client can resynchronize automatically and clear the source diagnostic using the fresh state, even when the payload equals the last good observation. Responsive subscribers continue receiving live events independently.

### Requirement: Select an installable exact engine release

The shared project policy SHALL select a published compatible native engine release by its exact version.

#### Scenario: An operator installs the configured engine

- **WHEN** an operator explicitly installs the pinned version on a supported platform
- **THEN** the release supplies its matching native engine and checksum. Normal dispatch uses that exact verified engine and surfaces any missing installation.
