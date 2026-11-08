h3. Title: AL2-9: Evaluate optional ledger signatures

*GOAL*

Decide whether and how optional ledger signatures help users verify signer identity and file integrity.

*CONTEXT*

Users raised SSH detached signatures and platform-managed keys. A signature can establish signed-byte integrity and trusted signer identity; it cannot prove factual telemetry accuracy. Key ownership, trust distribution and signature lifecycle remain undecided.

*ACCEPTANCE CRITERIA*

* Deliver a recommendation comparing per-developer keys with platform-managed key/service options, including SSH detached signing and at least one viable alternative. Explain threat model, costs, offline behavior and limits.
* Identify signing actors, private-key custody, verifier trust inputs, public-key distribution, rotation and revocation. No private keys or secrets enter the repository or ledger.
* Define when signing occurs, detached signature naming/storage and how regeneration, editing, migration and multi-contributor updates affect signatures. Distinguish unsigned, invalid, unknown signer and trusted-valid outcomes.
* Preserve baseline portable ledger access without private receipts or mandatory authenticity gates. Any proposal to change that product policy requires an explicit decision rather than being smuggled into verification behavior.
* Record open decisions and actionable bounded implementation recommendations. Discovery includes no signing/verification implementation, schema changes or new application tests; product acceptance of findings precedes implementation.

*Status*: Ready
*Blocked by*: None
*Work type*: docs

*IMPLEMENTATION NOTES*

This is product/security discovery, not proof of a production exploit. Keep factual correctness, transport integrity and signer authentication separate.

Follow [engineering rules](../../../TECHNICAL_GUIDELINES.md) and [regression boundaries](../../../docs/REGRESSION_PREVENTION.md). [PO-19](../../PO-19/SPEC.md) preserves the earlier development reconciliation; historical observations require revalidation, not an assumption of current runtime support.
