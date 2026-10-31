# chore(build): create embedded asset destination before copy (GEN-12)

## Summary

Create the embedded frontend asset staging directory before refreshing it with the Vite build output. This fixes `make build` failures on clean checkouts where the destination does not exist.

## Context

- Contract: `alspaces/GEN-12/CONTRACT.md`
- Work summary: `alspaces/GEN-12/SUMMARY.md`
- Plan and review: `alspaces/GEN-12/iterations/01/PLAN.md`, `alspaces/GEN-12/iterations/01/REVIEW.md`
