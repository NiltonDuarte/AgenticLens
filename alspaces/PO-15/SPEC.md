# Feature Specification: Cryptographic Signing and Verification for Telemetry Ledgers

## Problem Statement

AgenticLens writes telemetry ledgers as JSON files, but currently has no platform capability for authenticating their origin or detecting later modification. Developers have asked whether the ledger can be signed with SSH keys and whether a platform-managed key pair could be used. The trust, key custody, verification, and signature-file lifecycle decisions are unresolved.

## User Stories

- As a developer sharing a telemetry ledger, I want recipients to verify its integrity and signer, so that changes after signing are detectable and the signer identity is explicit.
- As a project maintainer, I want a documented trust and key-custody model, so that ledger signatures do not imply more assurance than the signing process provides.

## Functional Scope

- Create AL2-9 as a time-boxed product and technical discovery spike.
- Evaluate detached signatures for ledger files, including SSH signing where appropriate.
- Compare per-developer signing keys with a platform-managed signing key, including private-key custody and signer attribution.
- Define how trusted public keys or signer identities are distributed, rotated, and revoked.
- Recommend verification behavior, signature naming and storage, and how signatures relate to generated or subsequently edited ledgers.
- Record a threat model, operational trade-offs, and any implementation tickets needed after the spike.

## Acceptance Constraints

- The spike ends with a written recommendation and rationale, including whether signatures are detached files or another representation.
- The recommendation distinguishes integrity verification from proof that telemetry facts are accurate.
- The recommendation states who controls signing keys, who can verify, and how trust changes are handled.
- No private key or secret is committed to the repository or embedded in generated ledgers.
- The spike does not implement signing or verification in AgenticLens.

## Out of Scope

- Implementing ledger signing, verification, key generation, or key distribution.
- Claiming that a valid signature proves the ledger's contents are factually correct.

## Planned Ticket

- **AL2-9**: Evaluate Cryptographic Signing and Verification for Telemetry Ledgers.
