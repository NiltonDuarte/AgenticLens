# fix: [AL-81] generate manual metrics without capture registry access

Manual `generate` could be blocked by machine-local capture bindings even when authoritative source data and the requested ledger were available. It now uses project/metrics/source configuration without reading or initializing the capture registry. Capture-only settings are unnecessary, and explicit manual attribution leaves automatic bindings unchanged.

Source/project validation, portable owner replacement/preservation, shared publication locking and interruption recovery remain intact. `--capture-dir` is accepted but ignored manually. Automatic worker publication retains enabled-policy and registered-owner admission; session/nonce selectors cannot bypass it. Grouping warnings and identifiable selected-family discovery failures now preserve accepted artifacts instead of publishing incomplete facts.

Validation: full `make` passed, including Go race integration, frontend units/integration, static checks, installation/distribution checks and multi-platform builds. Both configured models approved all four review focuses after corrections. Real Codex extraction matched for one root and eleven native children with registry absent/present, with no capture-state changes.

Native proof covers manual Codex extraction; automatic safeguards and OpenCode use isolated fixtures. Separate extraction cutoffs can change provisional durations. See [contract](CONTRACT.md), [summary](SUMMARY.md) and [retained findings](iterations/01/REVIEW_SUMMARY.json).
