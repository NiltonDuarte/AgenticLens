# AL-53 Iteration 01 Testing

## Tested boundary and ownership

Baseline: `85a2af94b5036b0918aaf9a7a89aed3a990f2650`. Reviewed production component/view/formatter/thread changes and the existing view enrichment lifecycle tests before adding coverage. Test changes are confined to `web/src/views/TicketDeepDive.test.tsx`; no application files, existing assertions, or tests were removed. Fixtures are synthetic, self-contained variants of the existing frontend fixture module; no live ticket workspace or source database is read.

Working test diff SHA-256: `ab10e8abe1ce74a26ec11a73b10bc1e0b48729690d0ef95b6277ef6a8e2470db`. New panel source SHA-256: `703fcb04abf33f4652f6a8f49d9058b1f30d41f9f248f9e9c7aa4592e39c9d47`.

## Regression coverage

Added five runtime cases in the existing view suite, exercising the actual view-to-panel-to-thread integration:

- Two parameterized supported-v4 cases cover omitted and empty optional fact collections, exact empty copy, absence of tables, no console errors, accessible collapse association/state, and Enter/Space collapse/reopen.
- One skills case guards invocation multiplicity, recorded boolean status despite error-like text, known/unknown/empty role labels, turns and timestamp precision/tooltip, full multiline diagnostic preservation, and native Tab focus across rows.
- One delegations-only case guards independent rendering, recorded prompt preservation, caller/turn labels, positive duration/cost, measured zero versus absent measurements, repeated delegation rows, absent/unavailable children, and exact child identity against a prefix-sharing sibling. Enter/Space jumps focus and expand the correct child, preserve the initially open root, permit manual closing, and reopen on a second jump.
- One deferred enrichment case proves persisted skills and delegations render before optional enrichment settles and remain visible after its failure.

Existing view lifecycle coverage already exercises successful/empty/failed enrichment, stale ticket requests, trace inspection, and surrounding baseline functionality. Separate literal formatter/accessor tests would duplicate production-path assertions without protecting a distinct boundary, so none were added.

## Commands and results

From `web/`:

- `npm test -- src/views/TicketDeepDive.test.tsx`: 62 tests passed, no skips.
- `npm test`: 34 files, 717 tests passed, no skips; 32.68 seconds reported. Output stored outside Git at `/tmp/al53-frontend-tests.log`.
- `npm test -- src/views/TicketDeepDive.test.tsx -t 'Skills and subagents autopsy'`: all 5 selected cases passed after adding row Tab-focus assertions; 57 unrelated existing cases explicitly filtered by the command. No tests have skip declarations added.
- `npm run typecheck`, `npm run lint`, `git diff --check`: passed after the final test edit.

No production defects were found. Parent owns the required integrated `make`, Go race suite, asset comparison and standalone binary checks after upstream synchronization. DOM interaction evidence establishes keyboard semantics, focus, and content visibility; no browser geometry or narrow-viewport visual claims are made.

## Routine main integration: MGT-1 F1 (2026-10-31)

Compared prior AL-53 head `57058521c498b9e376bbab5ec2355adec63d7729` with actual incoming main `6bc1064943ae22e1b6586a6d145543c6a9cd23f2` and read the management F1/F5 findings. No new product acceptance criteria or production edits were introduced by the tester. The merged production state has one `focusSession: { id }` contract and one view handler for both navigation entry points. Test diff SHA-256 at verification: `6349e4178834f2e8d41983b947d89564812504b92306412fdb4810ad4d351cbe`; staged production navigation diff SHA-256: `c04147a6178adbb46bf74a6ae4cb2cce8f0db67f5702933a8e1662680922ccdd`.

Extended the existing upstream view test for committed task-badge navigation instead of adding a redundant standalone case. Its initial collapsed-child, expanded-child, focus, visible-body, and absent-dialog assertions all remain in the same test under its new combined-navigation title. The initial broad child-ID regexp now uses the exact thread-header accessible name, avoiding collision with the panel's new jump button. A long full child ID and a prefix-sharing sibling prove exact identity. The fixture now uses an error task so accidental parent click propagation would open a visible failure inspector; optional enrichment is deferred to expose the persisted baseline and later stale-response handling.

The combined case exercises both entry points in the same mounted view with Enter and Space, closes the child between invocations, and verifies repeat reopening, focus, scroll requests, unchanged root expansion and sibling collapse, and no unintended parent selection. It then leaves a child request and parent inspector active, switches to a ticket containing the same child ID, and checks that child focus/expansion and the inspector reset. A late old-ticket enrichment response must not restore the old parent trace or badge. Existing AL-53 and upstream live/fleet test bodies and assertions remain intact.

The initial view run reported 62 passing tests and one test-only failure: an assertion looked for the collapsed-only event counter after expanding the new ticket's trace. Corrected that assertion to the existing expanded count (`Showing 1 of 1 events`); no application fix was required. Typecheck and lint passed after adding the combined case. Parent owns the required full `make`, Go race tests and embedded/standalone checks on the final merge commit. DOM and method-call assertions establish interactions and scroll dispatch, not physical browser geometry.

Final focused verification: `npm test -- src/views/TicketDeepDive.test.tsx src/components/TraceWaterfall.test.tsx` passed both files and all 146 tests, with no skipped tests (20.60 seconds). Full view suite includes 63 passing tests and preserves all five original AL-53 runtime cases. `npm run typecheck`, `npm run lint`, and final `git diff --check` passed. Focused output is outside Git at `/tmp/al53-integration-focused.log`.
