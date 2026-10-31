# docs(REV-4): audit test value and identify 50 removal candidates

The suite contains tests that assert test-local literals or copied implementations without exercising application behavior. This audit recommends 50 specific removals, explains each rationale and retained protection, and records conservative dispositions and assertion evidence for all 1,252 test declarations across 92 test/support files.

Includes separate conditional consolidation findings so unique regression cases and compiler-negative contracts survive future reductions. This PR changes documentation only; no tests have been removed, and no reduced-suite equivalence or speedup is claimed.

Validation: full `make` passed on the baseline, including Go race tests and 793 frontend cases; uncached Go and Vitest JSON profiles were inspected. Inventory counts and all 50 removal entries were checked. Integration verification and telemetry provenance are recorded in SUMMARY.md.
