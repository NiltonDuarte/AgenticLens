# AL-79 manager admission and recovery evidence

The corrected manager tests pass on Linux. This covers the uncommitted iteration03 candidate based on `9c39cf9d`, not a consolidated delivery approval. The manager/config/CLI test inventory has SHA-256 `1b7f6dad86d9098f9c8483dfd1d93c225841cafadaac8eb71f65d45c7672cf0f`; the parent retains the private inventory and raw logs under `/tmp/al79-native-entrypoint-0ta97kas` and owns full integration.

`go test -race ./internal/adapters/manager -count=1` passed in 7.440s. These are diagnostic probes; the required unit-first `make` gate remains parent-owned. No Windows runtime was exercised. The native process/admission test file explicitly targets POSIX, matching the approved Windows self-update limitation.

The existing installation tests were handed over by the distribution tester and retained: durable production transaction prefixes, ownership and kernel leases, conflicting recovery and exact HTTP artifact/checksum handling. Recovery now covers both fresh installation and replacement at every real write prefix. Oversized checksum metadata is rejected before an artifact request.

New real manager/engine fixtures verify wrong digests, versions, executable roles and host architectures without destination mutation; idempotent permission repair preserves the cached inode, while differing bytes require explicit force. Standalone companion installation works outside a project, malformed present project configuration cannot fall back, and disabled hooks validate policy without initializing source, state or cache. The actual manager compiler dependency closure excludes capture, SQLite and dashboard implementations. Existing distribution tests own independent manager/engine versions, transport and raw engine behavior.

Disposable Go overlays demonstrate three effective oracles without editing shared production files:

- Bypassing artifact checksum comparison admits a valid engine with the wrong expected digest; the native admission assertion fails.
- Removing the engine-role probe admits a manager as an engine. An isolated configured-project control first proves that manager forwarding of capture/version probes works, preventing missing-project or missing-companion failures from masking the defect.
- Restoring broad `errors.As` handling for child exits suppresses the wrapped admission diagnostic; the explicit role-diagnostic assertion fails.

The last issue is finding `I03-BEH-01`, severity 4: manager CLI exit handling discarded useful installation errors when a native validation subprocess failed. The adversarial tester (`gpt-6.1-sol`, high) discovered it using a real wrong-role candidate; the implementation engineer corrected it by distinguishing direct forwarded exits from wrapped manager failures. The tester verified the corrected behavior and the failing mutation. Executable role confusion was discovered by the implementation engineer; this tester supplied independent verification, not discovery credit.

No product code, actual user installation, hook trust, commits or telemetry were changed by this test assignment.

## C2 review-correction verification

The corrected C2 manager suite passes: `go test -race ./internal/adapters/manager -count=1`, 10.663s. Its final manager/config/CLI inventory has SHA-256 `58129456a84d03af1e2e0b8f7753301bcc033d14b908c964da73398ac276ea91`; the private `MANAGER_TEST_SOURCE_INVENTORY_C2_FINAL.json` and `manager-adversarial-race-c2-physical.log` identify the tested files and result. The C1 inventory above remains evidence of its own revision.

Existing real fixtures now cover the supplied review corrections: a manager installed at a custom path and invoked through a symlink updates that physical installation without `--dest`, leaves the engine and project pin intact, and creates no default-home installation. All three existing version aliases work outside a project. Valid foreign local manager bytes are rejected before cache creation. Orphan and edited checksum sidecars remain unchanged, with no pending journal created. The distribution tester separately owns the real deleted-registry inspection and Windows command-construction controls.

The runtime exact-pin oracle deliberately places valid version-A engine bytes in a version-B cache leaf and exercises ordinary command selection, independently of install-time admission. A disposable overlay removing that runtime comparison selects version A successfully and fails the assertion. A second overlay removing only the sidecar ownership guard fails both orphan-checksum and edited-owned-checksum controls. Both mutations use private overlays and leave shared production source untouched. Their logs are retained under `manager-mutations-c2` in the private evidence root.

These are verification controls for findings supplied by the architecture, behavior, data/failure, security and test-effectiveness reviewers; they do not transfer discovery credit to this tester. The manager implementation engineer fixed the product behavior, and the adversarial tester (`gpt-6.1-sol`, high) verified the corrections. No new product defect was discovered in this correction assignment. Full integration and native Windows execution remain outside this scoped evidence.

The final C2 addition verifies the orchestrator's physical-path discovery: explicit `install --install-root` and `versions --install-root` receive the literal `alias/../native-cache` where `alias` points at a nested physical directory. Installation reaches the physical cache, listing reports that same root, and the prematurely cleaned logical destination remains absent. Restoring `Abs`/`Clean` in a private overlay makes the real manager create the incorrect destination and fails the assertion. Discovery credit belongs to the orchestrator (`gpt-6.1-sol`, high); the manager implementation engineer fixed the helper and this tester verified it. Shared production source was not changed by the probe.
