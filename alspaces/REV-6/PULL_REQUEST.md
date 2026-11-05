# test(testing): gate integration on successful unit tests (REV-6)

Verification previously mixed pure tests with native/storage/DOM checks, allowing expensive integration work to start before unit failures were known. Two ledger-boundary fixtures also omitted required ticket identity and rejected otherwise-valid controls.

Separate Go core packages and pure frontend tests from complementary integration suites. Make and CI require both language unit gates before repository integration; standalone npm commands require frontend units first. Preserve race coverage, shared embedded-asset preparation, native/platform fixtures and build-only commands. Repair fixture identities and strengthen intended validation-cause assertions; add real failure-injection and partition checks.

- [Current contract delta](alspaces/REV-6/iterations/02/CONTRACT_DELTA.md)
- [Implementation plan](alspaces/REV-6/iterations/02/PLAN.md)
- [Summary, verification and historical review](alspaces/REV-6/SUMMARY.md)
