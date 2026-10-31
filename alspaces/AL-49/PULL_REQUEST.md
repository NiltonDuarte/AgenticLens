# feat(live-inspector): reverse event ordering and add quick navigation (AL-49)

## Summary

- Render live message turns latest-first by default, with an accessible order toggle.
- Preserve source-based logical turn numbering and message attribution.
- Keep the reading position stable during live updates and offer a jump to the latest turn.
- Add direct scroll-to-top and scroll-to-bottom controls.

## Context

- Technical contract: `alspaces/AL-49/CONTRACT.md`
- Work summary: `alspaces/AL-49/SUMMARY.md`

## Scope

Presentation-only changes in the Live Trace Inspector. No backend or API contract changes.
