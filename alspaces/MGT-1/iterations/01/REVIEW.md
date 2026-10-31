# MGT-1 Open pull request integration review

Snapshot: 2026-10-31. Main: `85a2af94b5036b0918aaf9a7a89aed3a990f2650`. Initial Git status clean. Audit-only scope; no product fixes, merges, or comments on reviewed PRs.

## Recommendation

Merge **#76 -> #79 -> #75 -> #77 -> #78 -> #55**, updating each remaining branch from main and running `make` after each integration. The first three form a clean textual combination. Resolve #77 navigation conflicts explicitly. Hold #78 for the reproduced diagnostic defects below and the combined trace UI resolution. Regenerate and validate release #55 after the feature set is integrated. Order reduces integration churn; it cannot eliminate conflicts.

## Prioritized findings

### F1 — P1 integration: #75 and #77 implement incompatible child-navigation APIs

[#75 ThreadViewer](https://github.com/NiltonDuarte/AgenticLens/blob/f57dbf87167439313de57a58f4bc0c6f64548b35/web/src/components/ThreadViewer.tsx#L6) introduces `focusSession: {id}`; [#77 ThreadViewer](https://github.com/NiltonDuarte/AgenticLens/blob/57058521c498b9e376bbab5ec2355adec63d7729/web/src/components/ThreadViewer.tsx#L6) introduces `jumpRequest: {sessionId}`. Both change the same component and TicketDeepDive call site. Merge simulation conflicts in both production files. Choosing one side drops either waterfall badge navigation or panel navigation; retaining mismatched props breaks typecheck. Use one state/prop contract, connect both actions, preserve exact-session expansion/focus and repeat activation, and keep both test sets. Status: open; no resolution implemented or tested.

### F2 — P1 integration: #75 and #78 conflict in the live inspector and trace tests/styles

Confirmed conflicts: `LiveTraceInspector.tsx`, `LiveTraceInspector.test.tsx`, `TraceWaterfall.test.tsx`, and `index.css`. The waterfall production file itself merges cleanly, but that does not validate the live conflict resolution. Preserve invocation badges and child navigation together with amber content warnings, red execution failures, failure expansion defaults, copy controls and the #76 reading-anchor behavior. Do not resolve by taking a whole file from either branch. Status: open integration work.

### F3 — P2 defect in #78: hydration discards new content-warning metadata

[#78 event reconciliation](https://github.com/NiltonDuarte/AgenticLens/blob/8c4e2dc7a25586b13af6a7af0c0859b73edef65b/backend/internal/domain/events.go#L118) copies the existing recorded event and enriches payload/duration fields, but never copies `HasContentError`, `ContentSeverity`, or `ContentErrorSnippet`. The frontend [equivalent](https://github.com/NiltonDuarte/AgenticLens/blob/8c4e2dc7a25586b13af6a7af0c0859b73edef65b/web/src/domain/events.ts#L23) has the same omission. A stored event from before this feature, matched by identity to a newly classified hydrated event, therefore stays unflagged and cannot open the content diagnostic drawer. Reproduced through production `domain.ReconcileEvents`: input warning=true, output warning=false, snippet=nil. Add explicit enrichment semantics for the new optional metadata at both boundaries, preserving recorded diagnostics when local enrichment lacks them. Status: open; implementation out of scope.

### F4 — P2 defect in #78: structured Codex output loses semantic text classification

[#78 codexToolOutput](https://github.com/NiltonDuarte/AgenticLens/blob/8c4e2dc7a25586b13af6a7af0c0859b73edef65b/backend/internal/adapters/codex/rollout.go#L751) unwraps string output but returns raw JSON for object output. The content patterns are line anchored, so an accepted structured content result does not expose its text to the classifier. Production `codex.ParseRollout` reproduction with identical `error: compilation failed` content: plain string produces warning=true/count=1; object `{"content":[{"type":"text","text":"error: compilation failed"}],"isError":false}` produces warning=false/count=0. Decode supported text-content envelopes at the adapter boundary and preserve raw output separately. This is a deterministic accepted-input repro, not evidence that every installed Codex version emits this exact envelope. Status: open.

### F5 — P2 integration: planning dependency edits must be combined, not selected by side

`TICKETS.md` conflicts for #79/#78, #79/#77, and #77/#75. #79/#77 additionally conflict in the AL-55 specification. Each branch removes only its own AL-52/53/54 blocker; #75 leaves AL-55 dependency updates for later integration. After #75, #77 and #79 all land, AL-55 must have no remaining blockers, move into the Ready section, and carry Status: Ready in both index and specification. Preserve every completed-ticket removal and done-file move. Choosing either dependency list will resurrect an already satisfied blocker. Status: open integration work.

### F6 — P2 release readiness: #55 has no checks and does not describe the pending feature set

#55 is MERGEABLE but UNSTABLE with zero check runs in the captured GitHub snapshot. It changes only manifest, changelog and web package version. Its changelog omits #75–#79 because they are not on main yet. Release Please runs on main pushes; let it refresh after the selected features merge, inspect the updated head/changelog, and run required checks explicitly if bot-triggered PR CI is absent. The workflow creates release metadata before subsequent checks/build/package/upload steps, so verify all binary assets and checksums after release; a tag alone is insufficient. This audit did not run a release or establish why PR checks are absent. Status: release validation remains open.

## All open PR heads

| PR | Ticket / purpose | Captured head | CI |
|---|---|---|---|
| [#55](https://github.com/NiltonDuarte/AgenticLens/pull/55) | chore(main): release 2.1.0 | `03c5d1e9a1e25a551dd20f6fcc041be144f1416a` | No check runs |
| [#75](https://github.com/NiltonDuarte/AgenticLens/pull/75) | feat(traces): show invocation badges and child session links (AL-52) | `f57dbf87167439313de57a58f4bc0c6f64548b35` | 3 successful checks |
| [#76](https://github.com/NiltonDuarte/AgenticLens/pull/76) | feat(live): refresh active session traces automatically (AL-47) | `21f00442ab7f04da9fabf0ce5736f89a0ee874bf` | 3 successful checks |
| [#77](https://github.com/NiltonDuarte/AgenticLens/pull/77) | feat(web): add skills and subagents autopsy panel (AL-53) | `57058521c498b9e376bbab5ec2355adec63d7729` | 3 successful checks |
| [#78](https://github.com/NiltonDuarte/AgenticLens/pull/78) | feat(telemetry): distinguish semantic tool content errors (AL-43) | `8c4e2dc7a25586b13af6a7af0c0859b73edef65b` | 3 successful checks |
| [#79](https://github.com/NiltonDuarte/AgenticLens/pull/79) | feat(skills): add fleet adoption analytics and comparison filters (AL-54) | `14fe7fd74a5ea82e2ce98b71e8b6aa138b3c3e09` | 3 successful checks |

## Pairwise merge evidence

All 15 unordered pairs were simulated with `git merge-tree --write-tree <headA> <headB>`. Four pairs conflict; eleven merge textually. Full outputs and shared-path inventories are in [merge-evidence.json](../../merge-evidence.json).

| Pair | Conflicting files |
|---|---|
| #79 / #78 | `alspaces/plan/TICKETS.md` |
| #79 / #77 | `alspaces/plan/TICKETS.md`, `alspaces/plan/to-do/AL-55 formal schema v4 evolution migration step and backfill runner.md` |
| #78 / #75 | `web/src/components/LiveTraceInspector.test.tsx`, `web/src/components/LiveTraceInspector.tsx`, `web/src/components/TraceWaterfall.test.tsx`, `web/src/index.css` |
| #77 / #75 | `alspaces/plan/TICKETS.md`, `web/src/components/ThreadViewer.tsx`, `web/src/views/TicketDeepDive.tsx` |

## Verification

- Export A: #75 + #76 + #79, synthetic head `99ef57ee21f89750911f925f2baa3d9a09d2191b`, tree `a8ebeae179f3e45a2df9dd73ec44acccfd254f01`. Textually clean. Frontend typecheck, lint and full suite: 35 files / 736 tests passed. Production frontend build passed.
- Export B: #78 + #76 + #79, synthetic head `60ee050524884096fc2ae030bef3578996845634`. Production code clean; TICKETS.md intentionally retains conflict markers in this disposable export. Typecheck and 13 frontend files / 361 tests passed; canonical frontend build passed. Go race suites passed: domain, live, codex, opencode, HTTP, adapters. Production reproductions: hydrated input warning=true -> output warning=false/snippet=nil; Codex plain string warning=true/count=1 -> structured content warning=false/count=0. This is code-combination evidence, not a resolved branch.
- Export C: #77 + #78 + #76, synthetic head `c5ccc2b9acaced3584923c63873018aae22fad75`, tree `b6f05e250323275118c1af8ed1798c27f6d22343`. Textually clean. Typecheck and 4 frontend files / 191 tests passed.
- Temporary exports are source archives, not linked worktrees or product branches. Synthetic commits are local unreachable audit objects only. No reviewed head was changed.
- Existing MetricProgression act warnings appeared in Export A. No test failures were observed in the completed frontend suites.
- No combined #75/#77/#78 resolution, real browser geometry, live deployment, or release publishing was tested. Green individual CI and clean textual merges do not establish those boundaries.
- Tier 1–3 screening found no new domain I/O, outer-layer import, any/interface{}, or Node production runtime dependency in the reviewed changes. The primary architecture concern is duplicated navigation responsibility. Test performance was not a requested scope and no performance claim is made.

## Final verification facts

Both built integration exports A and B have byte-identical `web/dist` and `backend/internal/adapters/web/dist` trees (`diff -qr` passed). All six reviewed heads were rechecked after testing and remained unchanged. Export B selected Go race suites all passed. Full `make` on the audit branch is tracked separately below; it validates the documentation branch and baseline application, not the unresolved feature combination.

Audit-branch full `make` passed: Go vet and all Go race suites; frontend typecheck, lint, 34 files / 712 tests; canonical frontend build and four platform binaries. The invocation began on baseline 85a2af9; only audit Markdown/JSON changed during execution and origin/main integration was a no-op. Embedded asset trees match. This is not a full make pass for all five features together.
