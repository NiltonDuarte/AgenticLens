# AL-64 architecture review: CHANGES_REQUESTED

Run: reviewer-architecture, discovery, AL64-candidate-01. Model/effort: gpt-6.1-sol/high, inherited dispatch identity; no source-session identity or usage inspection performed. Baseline a84adbe5be46c2ef52ede0987a417562715d38e1. Candidate fingerprint /tmp/al64-review-01/fingerprint.json SHA-256 6aab550b873d4b5de1242b889c26bb3b7c6838d4900ecae6d418b426c2a35695; all 46 recorded application/test files matched at final verification.

## Findings

### ARCH-SOL-01: Invalid Codex completion is emitted as verified complete duration

Severity: 5. Classification: introduced evidence misclassification around an existing invalid-interval omission. Location: backend/internal/adapters/codex/rollout.go:440-448 and finalizeFacts at 931.

A valid task_started at 00:00:10 followed by a matching task_complete at 00:00:05 reaches the completion branch. Its date guard rejects the arithmetic, but the branch still deletes the pending start and does not record a duration limitation or warning. finalizeFacts subsequently emits active_duration_seconds 0 with state complete, completed_seconds 0, and no limitation. Invalid timestamps must be handled explicitly and preserve honest unsupported coverage; they must not become a completed zero measurement.

Evidence: an isolated Go overlay invokes the actual public ParseRolloutAt parser at cutoff 00:00:20 on those records. It prints seconds=0, State:complete, Limitation empty, warnings empty. The assertion requiring an explicit non-complete result fails. The arbitrary valid session ID does not bypass or mock the parser.

Architecture implication: the shared pure DurationInterval helper correctly handles reversed completion, while the separate Codex completion path omits equivalent limitation handling. Correct the concrete branch and regression; no broad adapter unification is required.

### ARCH-SOL-02: v1 update timestamp gate discards an explicitly open interval through the cutoff

Severity: 6. Classification: introduced accounting defect. Location: backend/internal/adapters/opencode/schema_v1_session_extraction.go:195-201, especially p.time_updated<=cutoff.

An assistant and a native running tool have established start 1760000001000. The tool remains running with state.time.start at that same instant, but its row has a later progress/update time 1760000020000. At explicit cutoff 1760000011000, extraction discards the running proof entirely because its update timestamp is after cutoff. The actual adapter emits zero seconds with partial coverage instead of the ten supported seconds through the cutoff. A source row updated after operation cutoff need not represent a future activity start; this can occur when a tool/text record progresses while generation is reading it.

Evidence: an isolated overlay uses the existing SQLite fixture helper only for schema construction, inserts the described source rows, and invokes ExtractSessionsStrictWithSourceAt. It logs seconds=0, State:partial and limitation unfinished activity lacks reliable current execution evidence. The assertion expecting ten provisional seconds fails. Known start and current open lifecycle establish this interval independently of whole-session wall time or missing completion alone.

Use native activity start/end boundaries to decide whether the interval covers cutoff; do not treat the generic last-update timestamp as its start or completion. Preserve conservative handling for cases lacking those boundaries. The related latest-message SQL predicate also has no cutoff bound, so its treatment of future records deserves an affected-scope check during the fix.

## Verification and hypotheses assessed

Command executed from backend: go test -overlay /tmp/al64-arch-sol-probe/overlay.json ./internal/adapters/codex ./internal/adapters/opencode -run TestAL64Arch -v. Both counterexample assertions fail with the concrete outputs above. Probe files and overlay live only under /tmp/al64-arch-sol-probe; shared source/tests were not modified. An initial overlay had a duplicated backend path, ran no tests and established nothing; the corrected overlay ran both named tests.

Read responsibilities: pure duration model/calculation/validation, extraction cutoff flow, capture and backfill request-local ownership, durable invocation DTO/migration copies, aggregate/API/progression/UI evidence propagation, strict schema detection and related unchanged compact fact equality. The domain remains free of clocks/I/O. Operation-specific cutoff is passed from shell boundaries; backfill clones its engine for request-local cutoff state. Legacy no-evidence facts are preserved as unknown. Duration evidence is carried by DTO and migration copies and aggregate projections; no raw event journal was added. These refute the main FCIS, shared-engine mutation and accidental copy-loss hypotheses in the inspected paths.

Reader-first optional v6 compatibility is explicitly bounded: expanded readers preserve old valid facts, while old strict readers reject new emitted evidence. No transparent old-binary forward compatibility or release rollout approval is inferred. This accepted deployment limitation is not reported as a new defect.

Shared verification supplied by parent/tester: targeted all-family fixed-cutoff owner-retention/DTO/migration/UI regressions pass, while full Go race/frontend suites are in progress. They are not substitutes for the isolated counterexamples. This reviewer did not repeat full suites or independently verify the final embedded binary.

## Remaining limits and disposition

Product semantics for historical open source state versus proof of process liveness remain pending user clarification. This review assesses the current source-state provisional candidate; it does not approve a policy change or claim open native state proves a live process. Native v1 cancellation can leave source state open, as separately recorded in RUNTIME_EVIDENCE.md. Final consolidated approval still requires that clarification, fixes, affected verification/re-review and integrated make. No sibling conclusions were consulted during discovery. Parent owns validation, fixes, consolidation and publication.
