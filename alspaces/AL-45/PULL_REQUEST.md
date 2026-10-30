# fix(presentation): contain and middle-truncate debug identifiers (AL-45)

## Summary

- Middle-truncate long message and part debug identifiers while preserving the full ID in each token title.
- Keep copied debug prompts sourced from full, untruncated IDs.
- Constrain message metadata and token flex layouts to prevent horizontal overflow in narrow inspector panes.
- Add formatter and component regression coverage.

## Context

- Contract: `alspaces/AL-45/CONTRACT.md`
- Work summary: `alspaces/AL-45/SUMMARY.md`
