# refactor(AL-70): make configuration and ticket identity explicit

Implementation draft; validation and delivery remain pending.

AgenticLens currently substitutes source paths, configuration values and ticket identities inconsistently across CLI, adapters and dashboard consumers. The approved change uses explicit policy and owning-file path origins, literal ticket prefixes with consistent filtering/sorting, exact direct session selectors and committed net production/test contribution metrics. Shared summary policy and generic artifact API/UI align manual capture and dashboard behavior.

Retired configuration/CLI features are removed cleanly. Preserve portable recorded history, explicit data migration/formatting, exact interrupted-publication recovery, recognized source schemas and capture-state safeguards. Former generation metadata cleanup remains only in explicit migrate/backfill preparation.

Acceptance: [contract](CONTRACT.md) and [confirmed inventory](LEGACY_AND_FALLBACKS.md). Execution/verification: [iteration 02](iterations/02/PLAN.md). Current results and limits: [summary](SUMMARY.md).

Main was merged before implementation at f63e3db6. Replace this draft's pending status with actual verification and review evidence before publication.
