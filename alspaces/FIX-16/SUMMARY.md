# FIX-16: Restore release candidate publishing

## Baseline and status

Initial working tree clean at dc44619ae15586fe1194f872743033f523635c92 on FIX-16/restore-release-publishing. Configured workspace root: alspaces. No existing FIX-16 workspace or ticket index entry was found. Planning is in progress; implementation awaits scope approval under the development skill.

## Investigation

GitHub Release Candidate run [41733288460](https://github.com/NiltonDuarte/AgenticLens/actions/runs/41733288460) computed v3.0.0-rc.122+dc44619 successfully, then failed Static Analysis. The installation source-build test compared that inherited VERSION to Git's default v2.0.0-rc.115+65e0eeb-7-gdc44619. make honors the explicit VERSION as designed. Failure occurs before artifact publication. The preceding run 41720631449 also failed Static Analysis; its detailed logs were unavailable through the attempted CLI queries.

RC metadata is exported before checks; stable metadata is computed after checks. Preserve GEN-19 target selection, immutable source binding, stable-tag eligibility, checks and publication behavior.

## Verification and limits

Planning correction: the review feedback requested original wording and confirmed GitHub evidence rather than a hypothesis. Iteration 01 now omits the copied request, links the exact run and revision, and records the failed assertion and skipped publication steps. The implementation scope and acceptance criteria are unchanged; scope approval remains pending. No new iteration is needed for this clarification.

Read-only remote run metadata and exact failure log inspected. Attempted local focused reproduction with the recorded RC environment stopped earlier because generated embedded frontend assets and frontend dependencies are absent in this checkout. This is an environment prerequisite failure, not local confirmation of the version mismatch. After approval, install dependencies, build canonical assets and reproduce before fixing.

Explicit scope approval received after the planning correction. Coder dispatched for the existing installation test; no workflow changes are planned. Telemetry deferred until the final delivery step.
