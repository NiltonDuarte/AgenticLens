# refactor(AL-70): honor explicit configuration and ticket identity

AgenticLens now applies explicit configuration and owning-file path origins consistently across CLI, adapters and dashboard. Literal prefix mappings, exact session selectors and natural ordering preserve recorded ticket identity. Git metrics measure committed net branch contribution using independent production/test rules.

Shared configurable summary output and generic artifact items replace fixed destinations/contracts. Live views distinguish unavailable data from measured emptiness, preserve occurrence identity and reject stale polling replies. Retired configuration/CLI features are removed cleanly; portable history, explicit migration/formatting and authorized publication/state recovery remain supported.

Validation: full make passed at `604d6338`, including Go race integration, 218 frontend units, 741 integrations, static checks, installation/distribution/orchestration and five-platform native builds. Both host binaries execute; standalone UI/API preserved 12 source-free fixture ledgers. Reviews and verification closed 23 findings; compact evidence is in alspaces/AL-70/iterations/02/REVIEW_SUMMARY.json.

Limits: Pages prose was inspected but Ruby tests were unavailable locally. Windows evidence is scoped native fixtures. A12 model-placeholder cleanup remains unapproved and unchanged. No hooks were installed or automatic capture enabled.
