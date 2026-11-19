# docs(PO-26): define the v3 automatic capture release gate

## Outcome

Added [AL-96](../../../alspaces/plan/to-do/AL-96%20verify%20automatic%20ticket%20metrics%20capture%20end%20to%20end%20for%20v3.md) to the [roadmap](../../../alspaces/plan/TICKETS.md) as Ready and blocking the v3 stable release. The test begins with an empty destination, exercises supported native capture integrations, and compares the published session tree against the harness source. Capture defects found by the test must be fixed and the corrected release candidate must pass.

## Verification and limits

PO-26 delivers the acceptance criteria. A separate development session owns the test and any fixes. V3 capture behavior remains unverified by this planning change.

## Pull request

Add AL-96 as a v3 release gate. A real agent and its delegated sessions must publish to a fresh ticket ledger through the installed Codex and supported OpenCode integrations. The gate checks source-backed session identity and parentage, first publication without manual generation, and preservation without duplicate metrics on refresh. Fix and retest any capture gaps before shipping v3.
