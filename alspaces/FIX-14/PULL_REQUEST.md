# fix(build): create embedded asset destination before copy (FIX-14)

## Summary

Create the embedded frontend asset staging directory before refreshing it with the Vite build output. This fixes `make build` failures on clean checkouts where the destination does not exist.

## Context

- Contract: `alspaces/FIX-14/CONTRACT.md`
- Work summary: `alspaces/FIX-14/SUMMARY.md`
- Plans and reviews: `alspaces/FIX-14/iterations/01/PLAN.md`, `alspaces/FIX-14/iterations/01/REVIEW.md`, `alspaces/FIX-14/iterations/02/PLAN.md`, `alspaces/FIX-14/iterations/02/REVIEW.md`
