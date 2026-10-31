# fix(ci): build embedded frontend assets before Go checks (FIX-13)

## Summary

Update backend CI to generate frontend assets before Go vet and backend tests. This lets Go resolve the existing `//go:embed all:dist` pattern in a clean checkout.

## Changes

- Provision Node.js 20 and install web dependencies in `backend-check`.
- Create the embed asset destination and run the existing frontend build before Go vet and tests.
- Preserve the current embedded filesystem and runtime asset behavior.

## Context

- Contract: `alspaces/FIX-13/CONTRACT.md`
- Work summary: `alspaces/FIX-13/SUMMARY.md`
- Plan and review: `alspaces/FIX-13/iterations/01/PLAN.md`, `alspaces/FIX-13/iterations/01/REVIEW.md`
