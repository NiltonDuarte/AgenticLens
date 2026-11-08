h3. Title: AL-81: Make manual generate independent of the capture registry

*GOAL*

Deliver manual generate as a source-backed ad hoc operation that does not read, create or mutate the AgenticLens capture registry, while retaining source validation, portable ledger preservation and serialized publication. Automatic capture publication continues to enforce its registry ownership and lifecycle rules.

*CONTEXT*

Manual Codex generation currently acquires capture configuration/state, creates the AgenticLens registry when absent, validates selected sessions against its bindings and records manual bindings. A stale local binding can therefore block an explicit session/ticket extraction even when the authoritative transcripts and existing ledger are available. This makes ad hoc generation depend on machine-local capture history. The database at issue is AgenticLens's capture registry, not OpenCode's source database.

*ACCEPTANCE CRITERIA*

* Manual generate with explicit ticket and supported session/nonce selection succeeds from valid project/metrics/source configuration without capture installation, capture enablement or capture-only state settings. Do not weaken validation of settings actually required by manual generation.
* A missing registry stays missing; an existing registry is neither read nor modified. A stale/conflicting ticket binding, corrupt registry or inaccessible capture-state location cannot block manual extraction. Test these cases with isolated paths and prove no registry/state initialization side effects.
* Codex generation reads authoritative rollout files. OpenCode generation still reads its configured source SQLite database through the supported source adapters. Missing, unreadable or invalid source data remains an actionable error, not empty successful telemetry.
* Preserve supported session/child discovery and source/project validation. An explicit manual ticket selection is authoritative for the requested output and does not rebind automatic capture sessions. This ticket adds no general reticket command, automatic cross-ticket cleanup or silent ticket inference.
* Refresh only exact source-complete owners, including legitimate lower/zero snapshots; preserve unavailable/partial owners, other harnesses and other contributors under the portable ledger contract. Do not require private capture receipts for supported Git-shared history or accumulate duplicate snapshots.
* Keep output destination validation, interruption-safe writes and shared publication serialization between manual and automatic writers without making manual publication depend on the registry. Test concurrent publication preserves updates rather than losing or duplicating sessions.
* Automatic publication, including the worker's generate invocation, still requires enabled configuration, valid registry ownership and ticket binding. Conflicting binding remains an error on that path; unassigned sessions do not publish. Verify separating manual admission does not turn automatic invocations into an ownership bypass.
* Demonstrate the same real source root/child extraction with and without a local capture registry and record source IDs, output facts and limits. Synthetic fixtures do not establish native runtime capture coverage.

*Status*: Ready
*Blocked by*: None
*Specification*: [PO-20 SPEC.md](../../PO-20/SPEC.md), F7.
*Work type*: fix
*Priority*: P1

*IMPLEMENTATION NOTES*

*Related*: Coordinate shared CLI/publication changes with AL-79 and AL-78; neither is a prerequisite. AL-80 consumes the automatic runtime and must retain its safeguards.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [relevant regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. Historical findings require revalidation; this rewrite is not runtime verification or implementation approval.
