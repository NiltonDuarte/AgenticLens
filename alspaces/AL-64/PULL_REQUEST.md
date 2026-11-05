# fix(duration): account for active turns and report long provisional activity (AL-64)

Unfinished Codex and OpenCode v1/v2 activity includes elapsed time through the snapshot cutoff. The existing active_duration_seconds total includes an optional positive provisional_seconds portion. Completion refresh replaces snapshots without double counting. Backend readers normalize earlier draft evidence to lean numeric fields; CLI/API/dashboard displays show the recorded total and positive provisional portion.

Explicit generation now reports long provisional activity on stderr, defaulting to three minutes. Feedback identifies the owning source/session/invocation and distinguishes exact source-open intervals from unavailable or retained evidence. Known short intervals do not trigger messages merely because their subtotal crosses the threshold. Configure generation.feedback.provisional_threshold_seconds; zero disables feedback.

Generic native caller/current-interval handoff remains unverified. Production uses qualified F03/F04 feedback; stronger F01/F02/F05 wording is gated without claiming native availability. No automatic continuation, session poke, fabricated completion, frontend feedback or persisted reasoning is added. Optional diagnostic failure leaves successful publication successful.

The branch preserves optional metric omission and shared-session warnings. Stored open intervals remain provisional estimates and can outlive their process. Upgrade readers before expanded v6 scalar writes; older strict readers may reject them. Existing unrelated delegation-retention and legacy-null debt remains documented. No release is authorized.

See the [technical baseline](alspaces/AL-64/CONTRACT.md), [numeric plan](alspaces/AL-64/iterations/03/PLAN.md), [checkpoint plan](alspaces/AL-64/iterations/04/PLAN.md) and [work summary](alspaces/AL-64/SUMMARY.md).
