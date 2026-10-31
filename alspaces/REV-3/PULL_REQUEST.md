# docs(review): prevent recurring REV and FIX regressions (REV-3)

Review REV-1, REV-2, and FIX-1 through FIX-15 and turn their lessons into boundary-specific planning and verification guidance. The new prevention checklist links to the full evidence inventory; canonical technical rules now cover source fidelity, complete/idempotent refresh, persisted-data availability, configuration propagation, and polling cost.

Update contributor guidance to use the existing backend-check/web-assets entry points and link the checklist from AGENTS.md. This is documentation-only; unresolved REV-2 runtime findings remain open.

Evidence: alspaces/REV-3/iterations/01/REVIEW.md. Verification and telemetry limitations are recorded in alspaces/REV-3/SUMMARY.md.

Validation: full `make` passed after origin/main integration (12 backend race-test packages, 793 frontend tests, lint/typecheck/vet, and four platform builds). Documentation links and whitespace passed. Existing npm audit findings and React test warnings are recorded in the summary.
