# AL-77 Work Summary

## Baseline and scope

- Initial revision: `a84adbe5be46c2ef52ede0987a417562715d38e1`.
- Initial working tree: clean; original branch: `AL-77/warn-duplicate-sessions-across-tickets`.
- Implementation branch: `feat/AL-77-warn-duplicate-sessions`.
- Request: implement the existing AL-77 specification for read-only warnings about session IDs shared across tickets.
- Workspace root: configured `alspaces`.

## Current acceptance and authorization

The user explicitly instructed: "Do not fill missing id from directory name. Ticket id is a required field on the data". [Iteration 02](iterations/02/PLAN.md) records this approved correction: require a nonblank persisted ticket ID through canonical validation and remove directory/request fallback in ordinary ledger reads. The earlier proposed recorded_ticket_id metadata solution and its pending approval request are superseded. No new metadata or API shape is planned. Missing-ID data is rejected with diagnostics; valid siblings and their source-independent recorded facts remain usable. No automatic repair or mutation is authorized.

## Iteration 01

Planning completed by the planner (gpt-6.1-sol, high; inherited runtime settings). See [CONTRACT.md](CONTRACT.md) and [iteration 01 plan](iterations/01/PLAN.md). Existing unfiltered ledger responses provide the required IDs; proposed production scope is frontend only, with a Fleet Overview warning and existing ticket navigation. Independent review is planned across four core focuses plus frontend async/accessibility, each under both configured model profiles.

User explicitly approved the contract and plan on 2026-11-04 by selecting "Approve implementation". No application code has changed and no checks have been run. Telemetry is deferred until the final delivery step; no ledger has been generated.

## Candidate 1 implementation and verification

Frontend overlap projection, disclosure, navigation and refresh are implemented. Coder and tester (gpt-6.1-sol/high, inherited harness settings) passed focused frontend checks (23 tests), full frontend suite (833 tests across 37 files), final strengthened domain test, typecheck, lint and focused HTTP race test. Production frontend build and mirrored asset comparison passed. Native summary keyboard activation and assistive technology remain untested in a real browser; jsdom tests cover semantics/focus/click and keyboard ticket activation. Existing unrelated MetricProgression tests emit act warnings. Full integrated make and standalone asset runtime verification remain pending.

Candidate fingerprint: `AL77-candidate-1-4a5906d5ba2f237a`, baseline `a84adbe5be46c2ef52ede0987a417562715d38e1`. Review remains open. Behavior reviewer (gpt-6-luna/medium, dispatch acknowledged) found BEH-01, severity 6: scanner directory fallback can cause the warning to treat missing recorded ticket IDs as real membership. Parent verified the scanner-to-DTO-to-UI path. Fix and affected verification remain pending; planner is preparing an iteration 02 delta because correcting this read-boundary distinction may require a minimal backend projection beyond the approved frontend-only scope. Other reviews remain active; no consolidated approval or publication.

[Iteration 02 correction plan](iterations/02/PLAN.md) is ready for renewed approval: expose the original nullable recorded ticket ID as nonpersisted read metadata in the existing HTTP response, and let only overlap inspection use it. Preserve effective ticket IDs, nested facts, existing fallback/navigation, totals and persisted schema. No correction code has been written. Missing review cells remain queued until the corrected candidate is frozen; in-flight candidate 1 reviews retain their own revision scope.

Candidate 1 review update: Behavior reviewers gpt-6.1-sol/high and gpt-6-luna/medium, and data integrity reviewer gpt-6.1-sol/high (all dispatch acknowledged) independently reported the same severity-6 BEH-01. The two Sol reviewers additionally reproduced the missing-recorded-ID API response using disposable HTTP overlay probes. All three discovery runs returned CHANGES_REQUESTED; no additional concrete findings. Shared discovery credit is retained. Other seven initial discovery cells remain queued, and affected reviews will target the corrected frozen revision after approval/implementation. Iteration 02 approval is still pending; completion notifications do not approve its scope.

## Iteration 02 execution

Planner (gpt-6.1-sol/high, inherited runtime settings) revised the current correction plan under the explicit user requirement. Coder (same inherited model/effort) owns canonical validator, removal of the two read fallbacks, schema contract documentation and relevant regression fixtures. Implementation is active; no further approval is pending. Prior frontend evidence is retained for unchanged behavior, while required validation/load/diagnostic checks and all final delivery gates remain pending.

Iteration 02 coder implemented the required ticket-ID boundary and removed both read fallbacks. Full backend `go test -race ./...` passed on the final production diff; final assertion-only strengthening reran focused schema tests successfully. HTTP tests exercise all four missing-ID forms, all-invalid and mixed sibling listings, unchanged valid totals, repeated reads without file changes, and a valid recorded ID differing from its directory. Tester review is active. No API metadata field, response shape change, schema bump or live ledger edits were introduced.

Corrected candidate `AL77-candidate-2-3871cfb9066dccfb` completed all ten focus/model cells: behavior, data/failure, architecture, test effectiveness and frontend async/accessibility under gpt-6.1-sol/high and gpt-6-luna/medium (dispatch acknowledged). Nine returned scoped APPROVED. Architecture reviewer gpt-6-luna/medium claimed a completed-migration cleanup bypass (severity 5); parent gpt-6.1-sol/high refuted it by tracing runtime validation into the canonical DTO check and executing a disposable overlay probe: absent/null/empty/whitespace IDs reject before publication, valid completed-migration control passes. Its CHANGES_REQUESTED is retained with that refuted disposition, not relabeled as child approval. No unresolved introduced blocker remains; integrated gates are pending.

Preexisting limitations retained: behavior and data reviewers gpt-6.1-sol/high independently noted recorded-ID/directory divergence can fail existing detail navigation (severity 4); test-effectiveness reviewer gpt-6.1-sol/high identified an unchanged nested-session test masked by an earlier schema error (severity 2). Repository maintainers own separate identity-location and test-fixture maintenance follow-ups; neither changes AL-77 acceptance or authorizes path-based identity inference. Browser/assistive-technology and narrow-viewport geometry remain unperformed, with native semantics/source inspection and jsdom focus/click/ticket-keyboard coverage distinguished from browser proof.
