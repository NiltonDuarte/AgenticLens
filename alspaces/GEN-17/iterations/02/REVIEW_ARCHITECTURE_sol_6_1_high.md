# GEN-16 iteration 02 architecture review

Outcome: CHANGES_REQUESTED for this architecture assignment.

Run: architecture_sol_6_1_high, discovery. Requested and dispatch-configured identity: Codex gpt-6.1-sol, high reasoning effort. Runtime source-observed identity, token usage and cost are unavailable; the configured identity is not a claim of independently observed identity. No sibling persona or review conclusions were read.

Reviewed base: 99fbcd7be0d0ed05639e4a0cb3be8744a3d5371e. Reviewed source fingerprint: 785b903e14bb4af8923e054f2d8d5fb69f44210150b172d431d75aa5416d6f9a, independently recomputed using the REVIEW_RUNS.json method. Scope: backend/web diff plus untracked domain/persona.go, requirements in CONTRACT.md as superseded by iterations/02/PLAN.md, and consequential unchanged live SQLite, HTTP hydration, domain mapping, event reconciliation and hierarchy consumers. Product and test files remained read-only.

## Finding A-SOL-01: Return persona evidence with its trace snapshot

Priority: P2. Classification: introduced observable correctness defect and approved-contract violation; cache state ownership and snapshot consistency.

Location: backend/internal/adapters/live/service.go:335-344, with cache invalidation at :406-408 and evidence publication at :589-591.

Trigger: an active session has no session.agent but has a consistent explicit assistant persona. A live tree read overlaps another trace request and a committed source update, including an update in a different session. Background polling and HTTP/SSE trace readers make these concurrent service calls reachable.

The tree first obtains its trace through GetSessionTrace, which releases traceMu before returning. It then reacquires traceMu and separately looks up tracePersonas[rec.ID]. Between those operations another sessionTraceSnapshot can observe a new SQLite data_version, replace the entire tracePersonas map, and populate only its own requested session. The tree then resolves the missing map entry's zero value against its blank session.agent. It emits Agent="" and Role="", although the returned trace was read successfully and the source still contains a stable named persona. The UI renders Unknown and can subsequently flip back on another poll. Each individual map access is synchronized, so the Go race detector does not prevent this logical race.

Proof: a disposable Go test overlay used the real live service and a temporary WAL SQLite database with two active sessions, each containing an assistant message with agent="planner" and blank session.agent. A writer changed only the other message's time_updated while three readers requested GetSessionTrace("other"); the foreground repeatedly requested GetSessionTree and asserted the target remained planner. The first ordinary run failed at poll 16 with `stable assistant persona lost at poll 16: agent="" role=""`. Three race-enabled repetitions all reproduced the same failure at poll 0, with no data-race diagnostic. An otherwise identical control with no source commits passed 200 polls under the race detector. The persona content never changed in either fixture.

Required correction: return the relevant PersonaEvidence together with the trace from one immutable snapshot obtained under traceMu, and have GetSessionTree use that returned evidence. Avoid obtaining a related snapshot component from the mutable cache after the snapshot call returns. Keeping both values in one cached entry is one possible implementation. Preserve the existing source revision, error and eviction behavior, and add an isolated concurrent invalidation regression. This concerns transient live display only; no persisted-role corruption was demonstrated.

## Other consequential hypotheses assessed

- FCIS: PersonaEvidence imports only strings, returns new value structs, retains conflicts, and gives explicit nonblank session metadata precedence. JSON decoding and source access stay in adapters. Both OpenCode schema-family extractors resolve once per session and apply the result to invocation and event facts; transport roles do not become persona candidates.
- Raw versus presentation roles: extraction and hydration use source spelling after trimming. RoleMappings remains a display projection in DeriveMetricsLedger and live node Role; DeriveMetricsLedger returns original Facts. The new Unknown formatters operate at presentation boundaries rather than changing persisted roles. Hydration tests explicitly cover a raw Build persona with configured mappings and nonmutation of stored facts.
- Unknown grouping and ancestry: OrderRoles/OrderSubagentRoles now retain the blank bucket, and their aggregation callers group by the same trimmed key. Workflow child effort, peaks and velocity use ParentID rather than persona. Existing BuildSessionTree topology is unchanged; no persisted topology field was introduced. The intentional blank-role exclusion from role-alternation anomaly evidence does not manufacture a persona.
- Live work bound: assistant evidence adds one O(message count) pass when an existing trace read occurs; unchanged source revisions reuse the evidence and trace, and both maps share active-session plus one on-demand-session eviction. This introduces no independent repeated full-history lookup on unchanged polls. Global revision invalidation and full trace reconstruction on committed changes predate this change; no performance benchmark was claimed.
- Baseline compatibility: legacy constants remain readable. Hydration returns persisted baseline events when source access fails or produces no events; successful enrichment reconciles with the baseline and does not write it. Known stored personas remain the fallback when optional persona evidence is absent. Existing exact-identity reconciliation and source-owned refresh behavior were not modified by this diff.

## Verification and limits

Reused iterations/02/TEST_REPORT.md at the reviewed fingerprint: all 12 backend packages passed go test -race ./...; frontend 35 files/783 tests passed; final typecheck, lint, selected changed-fixture rerun and diff check passed. These results did not cover the interleaving above. No broad suite was repeated.

Executed from backend:

- `go test -overlay /tmp/gen16-arch-probe-35lcvwcn/overlay.json ./internal/adapters/live -run '^TestArchitectureConcurrentPersonaSnapshotProbe$' -count=1 -timeout=40s`: FAIL, reproduced finding.
- Same command with `-race -count=3`: FAIL in all three repetitions, reproduced finding without race-detector warnings.
- `go test -race -overlay /tmp/gen16-arch-probe-35lcvwcn/control_overlay.json ./internal/adapters/live -run '^TestArchitectureConcurrentPersonaSnapshotProbe$' -count=1 -timeout=40s`: PASS for unchanged-source control.

Probe code and overlays are disposable files outside Git; no production/test source was edited. Its synthetic source records establish this concurrency behavior, not this session's actual persona. The report records the fixture and call sequence so the parent/tester can turn it into a durable self-contained regression. No fix was implemented or verified here. Final integrated make, embedded asset rebuilding/standalone verification, telemetry, publication and cross-review consolidation remain parent-owned gates. This scoped outcome does not certify the entire repository.
