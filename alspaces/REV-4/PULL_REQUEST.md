# test(REV-4): remove 50 low-value tests with individual review rationale

Remove six Go and 44 TypeScript test declarations that only confirm local fixtures/copied implementations or duplicate stronger surviving cases. Delete two files made entirely of type-cast JSON round trips and clean up unused test-local setup/imports. Production behavior is unchanged.

REVIEW.md records the rationale and retained protection for every removal. TEST_INVENTORY.md preserves historical entries and distinguishes removed tests from retained and conditional candidates. Corrected baseline inventory: 1,253 declarations; reduced suite: 1,203. Boundary, migration, ownership, hydration, security, and negative compiler contracts remain.

Scope and implementation boundaries: [CONTRACT.md](alspaces/REV-4/CONTRACT.md), [iteration 02 contract](alspaces/REV-4/iterations/02/CONTRACT.md). Verification and telemetry provenance: [SUMMARY.md](alspaces/REV-4/SUMMARY.md). No controlled speedup or mutation-equivalence claim is made.
