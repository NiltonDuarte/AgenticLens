# AL-79 iteration 03 C3 build/distribution correction review

**APPROVED within this narrow focus; new findings: 0.** Run `i03_build_luna_c3`; requested profile `codex/gpt-6-luna/medium` (runtime metadata unavailable). Initial C3 hash: `428b192a7ea05cea99fb12944db0772092e975d142e5670ba642a944b878336e`; final amended freeze: `1731b30ea56e2aa378aa6b592f49dac29376c75ebca6328a09d3aca6751aff12`; C2: `d3a52d7385dae492e3c224f76b890234cc673987db543ee0cffefae58b910afe`; baseline: `9c39cf9d4f44bdb968cb92b0212dccf577d10255`. The verifier matched the initial C3 on the first pass and the final amended freeze at re-review. C3→final manifests differ only at `.agents/skills/agentic-lens-operator/references/configuration.md`; requirements and all application, runtime, test and build paths remain unchanged.

## C2 finding disposition

The C2 finding about stale wrapper/protocol/rollback instructions is corrected in the requested paragraphs. The guide now distinguishes `agentic-lens manager protocol` and `--manager-version` from `engine protocol` and `capture protocol`, and instructs operators to probe engine protocols on the selected engine executable. This matches `backend/internal/adapters/manager/run.go`: the manager answers `manager protocol`, rejects an engine-protocol probe as the wrong executable role, and exposes manager maintenance help separately.

The guide accurately separates explicit `install EXACT` engine provisioning from `self-update --version EXACT` manager replacement. It states that a later bootstrap failure does not roll back an already completed operation, consistent with the plan's per-operation transaction boundary. It limits foreign release caching to named/checksummed artifacts and defers native interface admission until target-host use, while preserving the separate local target-host admission rule. The paragraph on compatible owned manager updates preserves the stable command path and does not promise that a manager version check establishes capture health; ownership conflicts and separate engine/configuration checks remain explicit.

The neighboring historical Linux runtime coverage table remains an observation bounded by harness/version and degraded paths; it does not claim newly certified runtime coverage for C3. No issue in the corrected text requires widening this documentation-only pass.

The final amendment also changed the two remaining operator-guide references to schema v5 (binding-title requirement and ledger schema) to v6. Both now agree with the already-corrected v6 publication paragraph, the unchanged v6 ledger contract and the approved C2 implementation evidence. This is documentation-only consistency; it adds no runtime certification claim or product-path change.

## Evidence and limits

Reused the approved C2 build Luna/Sol evidence and unchanged C2 core/security/test approvals as authorized by the iteration plan. This pass did not rerun suites, integration `make`, machine installation, remote release download, or Windows runtime checks. Those remain parent-owned/pending; Windows runtime remains explicitly uncertified. This is a scoped documentation correction approval, not whole-ticket or release approval.
