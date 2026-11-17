# release-artifact-verification Specification

## Purpose
Ensure a release manifest covers the selected build and an installer can account for every binary it admits from that release.

## Requirements

### Requirement: Complete release inventory

Packaging and independent release verification SHALL require exactly one valid SHA-256 record per configured platform archive, raw manager and raw engine for the selected version.

#### Scenario: A manifest or input is incomplete

- **WHEN** a required record is missing, duplicated, malformed, unexpected or names an unreadable or missing input, or a recorded digest disagrees with the bytes
- **THEN** independent release verification exits unsuccessfully before executing a host artifact.

#### Scenario: A manifest producer fails or emits incomplete output

- **WHEN** hashing, sorting or validating the candidate manifest fails, including a hash command that emits a valid first record then fails or returns success with partial coverage
- **THEN** packaging exits unsuccessfully without installing a partial final manifest.

#### Scenario: The complete release verifies

- **WHEN** all selected files match their records
- **THEN** verification succeeds after the existing manager and engine host probes succeed if the host platform was selected. Files outside the selected inventory are never executed.

### Requirement: Verified installation and truthful failure state

Each installer SHALL require successful checksum verification before executing or replacing a fetched binary. Bootstrap shall accept only an unambiguous, well-formed digest record for the selected manager. Remote requests shall use the exact selected release for each component. An explicit local artifact shall require its supplied digest.

#### Scenario: Remote source selection

- **WHEN** the bootstrap or skill bundle installer uses a repository override for a network fetch
- **THEN** it accepts a valid owner/repository pair, constructs requests under the component’s exact selected release or source tag, and rejects malformed pairs before fetching.

#### Scenario: Verification fails before installation

- **WHEN** a download or hash command fails, even after emitting plausible output, or the artifact is empty, corrupt, of the wrong supported identity or lacks a valid matching digest
- **THEN** installation returns nonzero, preserves the existing binary and associated version state for the failed component, and emits no completion claim. Integrity failure prevents probing those fetched bytes.

#### Scenario: An earlier component succeeds

- **WHEN** an engine installation succeeds and a later platform or manager installation fails
- **THEN** the command returns nonzero, retains each completed engine installation and the previous state of the failed component, and reports only completed installations as successful.

#### Scenario: An exact release is installed

- **WHEN** a verified replacement passes the applicable platform and native identity checks
- **THEN** the command installs it atomically at the selected destination, with manager receipt and checksum matching the manager bytes, and honors the existing project engine pin. A foreign engine's runtime identity check remains deferred to its target.
