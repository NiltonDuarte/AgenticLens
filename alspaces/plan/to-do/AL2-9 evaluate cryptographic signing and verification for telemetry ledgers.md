h3. Title: AL2-9: Evaluate Cryptographic Signing and Verification for Telemetry Ledgers

*TARGET*: AgenticLens (Product / Security / CLI / Documentation)

*CONTEXT*

AgenticLens generates JSON telemetry ledgers but does not authenticate their origin or detect modifications after generation. Developers have raised SSH detached signatures and the possibility of a platform-managed key pair. A signature can establish that signed bytes have not changed and that a trusted key signed them; it does not prove the telemetry facts are accurate. The project has not selected who should control signing keys or how verifiers should establish trust.

*GOAL*

Produce a focused discovery recommendation for signing and verifying telemetry ledger files, resolving signer identity, key custody, trust distribution, verification workflow, and signature lifecycle before implementation is proposed.

*VERTICAL SLICE SCOPE*

* Product and security discovery:
  - Compare per-developer signing keys with a platform-managed signing key or signing service.
  - Evaluate SSH signing and other suitable approaches for detached file signatures.
  - Define verifier trust inputs, signer identity, key rotation and revocation, and expected offline behavior.
* CLI and workflow discovery:
  - Recommend how signing is invoked after ledger generation and how verification is performed.
  - Define where detached signatures are stored, how they are named, and what happens when a ledger is regenerated or edited.
* Documentation:
  - Record the threat model, assumptions, operational costs, limitations, and recommended follow-up implementation tickets.

*ACCEPTANCE CRITERIA*

# The spike documents a recommended signing model and compares it with at least one viable alternative.
# The recommendation identifies the signing actor and private-key custody boundary; no private key or secret is stored in the repository or ledger.
# The recommendation explains how verifiers obtain trusted public keys or signer identities, and how key rotation and revocation are handled.
# The recommendation defines the file/signature relationship and the expected signing and verification workflow, including behavior after ledger changes.
# The document explicitly distinguishes file integrity and signer authentication from factual correctness of telemetry.
# Any unresolved decision and the next implementation ticket(s) are recorded.
# No signing or verification implementation is included in this spike.

*DEFINITION OF DONE*

# A reviewed discovery document captures the recommendation, alternatives, threat model, workflow, and open questions.
# Follow-up implementation work is represented by one or more actionable tickets if the recommendation supports proceeding.
# The ticket is marked complete only after its findings are recorded and accepted by the product owner.
