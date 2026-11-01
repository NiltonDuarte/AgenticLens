# AL-68 iteration 01 review corrections

These are corrections within approved iteration 01, not a new scope or approval request. The initial product remains frozen at CANDIDATE.json until all discovery runs complete. FINDINGS.json and individual reports retain discovery evidence. Parent owns integration, telemetry and publication.

## Backend coder ownership

Own capture domain/adapters/commands and necessary main.go integration, preserving unrelated changes. Correct CAP-01 through CAP-07, CAP-10 and CAP-11 using one coherent ownership and snapshot publication contract. Do not solve stale publication by selecting a maximum token count: snapshots have multiple fields and optional/unknown facts. Separate successful authoritative replacements from retained fallback data; explicit fresh extraction and newer persisted facts must survive failed/partial capture refresh. Prevent publication ordering races with recoverable publication ownership plus baseline reload/precedence at the protected boundary. A report digest alone is not ordering protection.

Validate explicit selector owners against known exact bindings before publication; a legacy selector is not authorization to steal another ticket's session. Preserve authoritative source/store distinctions and reject canonical identity collisions.

Retain unresolved Codex child-start evidence and recover source-proven children within bounded work even when stop never arrives. Do not equate agent_id with source ID without proof. Keep hook-to-agent binding on actual supported protocols, without guesses.

Charge consumed or reserved bytes even when parsing fails; bound queued child work as well. A native receive timeout must bound stdin handoff itself, not only database calls. Render embedded templates in a single pass that cannot reinterpret replacement values.

Define truthful cutoff semantics and per-owner checkpoints; the reported boundary cannot precede reads included as if they were captured at that earlier instant. Keep ticket publication scoped to proven ticket owners. Unbound registrations remain discoverable through local diagnostics, but unrelated unassigned work must not consume a ticket's full extraction budget or publish its session identities into that ticket. Proven inherited children need an exact, bounded route into the ticket scope.

## Distribution coder ownership

Own both downloaders, both wrappers, existing distribution tests, setup/operator documentation and docs/SETUP_MANUAL.md. Correct CAP-08 and CAP-09. Stage replacement handlers, validate capture protocol and compatibility when hooks are installed, preserve the old binary on preflight failure and roll back owned state when later setup fails. Cover local-binary, release-download and wrapper-update paths; an incompatible update must never leave active hooks pointing at an unusable handler. Keep direct Unix and PowerShell setup guidance consistent by linking canonical capture/binding/trust/removal instructions. No third-party skill/persona changes or trust bypass.

## Tester ownership

Own capture domain/adapter/CLI regression tests and TEST_REPORT.md updates, coordinate with both coders. Reproduce meaningful report triggers before fixing and retain exact outcome oracles: unequal registry/canonical snapshots; successful explicit extraction despite failed registry; two overlapping publishers; known other-ticket explicit selection; child start without stop; multiple failed reads under one byte budget; placeholder-bearing paths; blocked stdin; replacement with incompatible binaries; unrelated unassigned sessions ahead of a bound owner; truthful cutoff. Convert existing successful mid-write rollback probes to a coherent durable regression where practical. Extend existing tests instead of mirroring implementation.

## Verification and review

After corrections, freeze a new candidate with exact content fingerprint and map fixes to affected focus/model cells. Re-run affected checks, use both configured profiles for affected re-reviews, retain old reports. Real Codex trust/callback/agent binding smoke remains pending a human trust action; no code correction or fixture success substitutes for it. Refresh real OpenCode smoke if interaction/source changes invalidate that evidence. Required integrated make follows implementation commit and main integration. Final telemetry is delayed to finalization, and no PR may be represented as approved while blockers or material evidence gaps remain.

## Focused remainder after candidate 02 re-review

Source: REVIEW-data-sol-02.md and behavior/architecture Sol-02 reports on unchanged fingerprint 17fdae...51cbc. Preserve all accepted correction guarantees. CAP-01 must compare real source checkpoints across successful explicit and registered generation, including an unchanged old source reread after a newer canonical publication; no fieldwise token maxima. CAP-03 must compare actual source-store identity through symlink aliases for new and existing registrations. CAP-12 requires durable/fair child-discovery continuation: skip/advance beyond already known OpenCode children before or with a page limit, and advance past unavailable/mismatched Codex metadata queue entries without dropping unresolved evidence. Every bounded retry must eventually visit later eligible work.

These remain review corrections in iteration 01, without a new acceptance boundary or approval request. Remaining candidate-02 review cells were superseded before dispatch; final affected reviews must cover every selected focus/model on the next frozen candidate.

## Scheduler fairness remainder after candidate 03

CAP01 and CAP03 remaining counterexamples are now verified fixed. CAP12 still blocks through two source-grounded schedules in REVIEW-data-sol-03.md and REVIEW-architecture-sol-03.md. Fix fairness across stop recovery, start discovery, OpenCode child discovery and ordinary owner refresh under BOTH session and byte limits. Failed/large work cannot permanently occupy first priority; the next pass must give other eligible work a chance with a usable budget. Persist phase/queue priority or an equivalent general continuation protocol; retain unresolved evidence. Do not patch only one budget value or rely on increasing limits. No new scope or approval required.
