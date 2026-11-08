# AL-79 C4 data and failure-handling correction re-review

**Outcome: APPROVED** for the assigned data-integrity and failure-handling scope. No unresolved blocking finding was established. Parent owns the consolidated verdict and delivery gates.

## Identity and revision

- Run: `data-luna-c4`; phase: correction re-review; candidate: `AL79-C4`.
- Requested profile: `gpt-6-luna` / `medium`. This records the requested profile, not independently observed runtime metadata.
- Baseline: `4df45967c801824500ac2ea39a8f52fa9cbdcf66`.
- Frozen `CANDIDATE_C4.diff` SHA-256 independently checked: `9d1ca49a319643feeb13774a28942037cfff9d175884d53a8e8495660a90e265`, matching `CANDIDATE_C4.json`.
- Reviewed the approved `CONTRACT.md`, iteration `PLAN.md`, this focus's prior data report, consolidated correction context, shared review procedure, portable-ledger contract, and relevant C3-to-C4 changes. No sibling detailed reports/personas were loaded.
- Only this report path was written. Product and test sources remained read-only; no delegation, commit, or publication was performed.

## Hypotheses and evidence

- **DATA-01 — strict automatic local limits: no remaining defect found.** The automatic CLI path selects `CaptureAdmission.ResolveCapture`; the manual compatibility path remains separate. The selected strict policy flows to reconciliation's `MaxSessions` and `MaxSourceBytes` budget. Reused the revision-matched targeted automatic-generation test evidence for omitted local numeric limits and conflicting shared limits. This is a targeted guard/consumer check, not a full-suite claim.

- **DATA-02 — journal resource rediscovery and recovery under leases: corrected.** `user_install.go` acquires the discovered resource leases, re-reads pending journals for both root and home, and calls `requireUserLeases` before either recovery write. It also rediscovers legacy resources and builds the migration plan only after recovery, then verifies that the fresh plan's resources are leased. The actual second-installer scheduling test preserves the journal/artifact snapshot while a newly relevant lease is held, then verifies recovery after release. Inspected `recovery-oracle.log`: bypassing the post-lock resource guard fails with `scheduled=true err=<nil>`; the unmodified focused C4 test passed.

- **DATA-03 — explicit child ticket and portable ancestry: no remaining defect found.** `inheritChild` returns without changing a child already explicitly bound (agent or manual selection), allowing parent reconciliation to continue. The child's source-proven `ParentID` remains in its own ledger. Current DTO validation rejects blank or whitespace-padded parent IDs but does not require the parent owner to be in the same ticket ledger; it does not synthesize parent facts. Reused targeted automatic separate-ticket child publication and portable save/reload/read-projection evidence. No new schema, migration, authenticity, or graph policy is needed.

- **DATA-04 — stable retained executable admission: corrected.** Retention checks source type/size/execute bit, copied length and source file identity/size/mtime, rereads the opened inode and compares a confirmation digest, syncs and validates the retained native executable before digest-path admission. Worker retention/launch continues to pass the retained path to later subprocesses. Reused C3 real-process preferred-switch evidence for unchanged worker behavior and C4 targeted truncation, opened-source, restored-mtime, and pending-work tests. I make no claim of a new live rebuild race execution.

- **DATA-05 — external-parent fresh and cached anomaly analysis: corrected.** Domain completeness requires a closed owned source set; an external parent makes fresh group/root coverage incomplete. Complete child analysis remains independently usable, and cached group summaries remain retained when fresh group coverage is unavailable. The formatter/CLI paths report partial parent/group context and avoid a complete clean conclusion while keeping recorded summaries visible. Inspected `coverage-oracle.log`: restoring the prior closed-ticket-only completeness rule fails the fresh and recorded-warning cases, while the recorded-empty control passes; unmodified C4 targeted tests pass. No parent stub, cross-ticket metric import, or ancestry deletion is used.

- **INST-01 — fresh migration plan after interrupted recovery: corrected.** Recovery runs only after re-discovery and lease coverage. `legacyUserChanges` is computed after recovery, then the newly generated change paths are checked against held leases before application. Reused actual durable-prefix/recovery and migration acceptance evidence; the stale pre-recovery-plan hypothesis is refuted by ordering in the production path and the targeted transaction oracle.

## Verification reused and limits

- Reused the assigned `TEST_EVIDENCE.md`, `DISTRIBUTION_EVIDENCE.md`, and `NATIVE_EVIDENCE.md` only for their stated revisions and paths. Focused Go log `/tmp/al79-data-sol-c4-probes/focused.log` shows the domain, capture, and adapter targeted packages passing. The two disposable mutation logs show the specific recovery-resource and external-parent coverage tests fail when their corrected guard is bypassed. These probes are correction validation, not independent broad discovery.
- Broader unchanged owner-refresh conclusions (complete smaller/empty versus partial/unavailable snapshots, idempotence, deleted destinations, and unrelated-owner preservation) carry forward from the prior bounded evidence because C4 does not alter those algorithms; no C4 re-execution is claimed for them.
- Integrated `make`/race verification is pending the parent. Native Windows runtime was not executed locally; cross-compilation is not runtime evidence and the actual hosted Windows stage remains outstanding. Existing normal-trust Codex Linux root/child evidence carries only for the recorded build and unchanged callback/extraction/binding paths; no fresh C4 native session was created.

This verdict covers only the listed C4 data/failure correction hypotheses and does not approve the integrated candidate or authorize delivery.
