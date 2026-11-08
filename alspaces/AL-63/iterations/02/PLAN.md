# AL-63 iteration 02: Execute the capability audit

## Steering and boundaries

Source: user correction, “revert the changes in AL-63 ticket. Your job here is to work on the ticket not to update it. Dispatch specialized agents if needed. Report the capabilities matrix.” Restore the ticket specification byte-for-byte from eb7d169c. Supersede iteration 01's planning-only outcome with execution of the existing discovery ticket. Preserve all original acceptance, portable history and original AL-55 bytes. No implementation, schema/runtime changes or new tests. User explicitly authorizes investigation and specialized agents in this session despite the previous PO-only plan. Application inputs are frozen at f4152225 (same source as eb7d169c); ticket restoration is documentation-only.

## Evidence and coverage

Use independent architecture, behavior, data and test-effectiveness focuses for the bounded capability audit, each under configured Sol 6.1/high and Luna 6/medium profiles. Each owns only its uniquely named report under this iteration; parent owns matrix, consolidation, telemetry, delivery and registry lifecycle. Shared checkout; preserve others' work; probes use private disposable copies. No reviewer may implement fixes or publish. Inspect relevant unchanged producers/consumers and seek native observations where locally available. Distinguish current native source, synthetic runtime tests, historical reused evidence and static support declarations. Unavailable/unexecuted live checks remain explicit; audit completion must not imply capability parity or successful recovery.

Coverage roster (all at f4152225, discovery):
- architecture / sol, luna: ingestion, source declarations, skill/delegation producers and child identity; ARCHITECTURE_SOL.md / ARCHITECTURE_LUNA.md.
- behavior / sol, luna: historical hydration, live/API behavior and error reporting across three families; BEHAVIOR_SOL.md / BEHAVIOR_LUNA.md.
- data / sol, luna: exact-owner recovery, missing/malformed/children preservation, mixed-age/idempotence; DATA_SOL.md / DATA_LUNA.md.
- tests / sol, luna: evidence strength and actual coverage of ticket preservation/refresh scenarios; TESTS_SOL.md / TESTS_LUNA.md.

Parent validates real Codex hydration against disposable AL-55/AL-62 copies, inventories source versions and available OpenCode schemas, preserves original ledger hashes, reconciles findings and produces CAPABILITIES.md as the owning audit deliverable. Reuse successful full make from 0a8cfcee only for unchanged application/test/build inputs; run targeted evidence commands and full make after integration when required. Final generation appends to existing AL-63 ledger with explicit Codex/docs selection and verified owner preservation. Existing PR state must be checked and its title/body rewritten to the executed audit.

Unknown models, native versions, inaccessible source records and unexecuted probes must stay unknown. No advertised feature is marked runtime-verified solely from a source search. Concrete defects are follow-up recommendations under existing owners, not authorization to implement them.

## Executed evidence and material outcomes

Parent HTTP probes: AL-55 recorded Codex complete232/5owners, partial104/1owner, unavailable0/0owners; recorded OpenCode v1/v2 complete27/4events. Detail reads preserve AL-55 five owners/four delegations and AL-62 26owners/three skills/25delegations without enrichment; no ledger writes. Native AL-62 full hydration exceeded one25s client timeout and remains unverified. Source ownership guards rejected copied-session CLI ingestion before parsing; no identity bypass is claimed.

Tests/Sol identified two audit evidence gaps, not product failures: discriminating v2 occurrence preservation/growth/repeat and explicit error-versus-success-text classification. Parent's private v2 CLI probe exercised growth1/1to2/2 skills/tasks, exact unchanged repeat, malformed/missing root, unavailable child/same-harness/other-harness preservation and two reads/one error. All assertions passed with incomplete/mixed-age diagnostics. No durable test or runtime code was added; native-positive v2 cases remain unexecuted. Keep initial INCOMPLETE review attribution while resolving its specific audit evidence gap.

Two initial Luna data/tests dispatches failed before starting with parent_not_active; retries use fresh delegate tasks with the same requested model/effort. No fallback provider/model or fabricated review execution is used. Detailed dispatch records remain private; durable outcome/attribution is in the consolidated review.
