# C5 behavior and integration review

**Run:** AL79 I03 C5; requested profile `codex/gpt-6-luna/medium` (runtime identity not independently measured). **Outcome: APPROVED** for the assigned corrected installation/adoption behavior only.

## Scope and evidence

Reviewed the frozen C5 candidate against base `3a62a726e029adb7c92fae26d8fe687339a21440`; `/tmp/al79-native-entrypoint-0ta97kas/verify_c5.py` matched candidate digest `6e7390d021da02c37ebe5c4213a6afa908bb164e57e8afef87b7f900be7955fb` both before and after review. Read the eight correction-scope files from `CANDIDATE_C5.json`, including untracked coordinator tests, and traced `InstallUserCodex` through resource discovery, lease acquisition, validation, recovery and post-recovery migration planning. Also traced `UserCodexInstalled` to automatic worker eligibility and the native user-binding CLI path.

The new shared record contains only owner/coordinator/checksum identity; rollback bytes remain in the private owner journal. A reference enrolls resources only when it matches that journal. Under the installer, fresh resource discovery is checked against all held leases before validation or recovery. Commit is private-journal removal, so stale references cannot roll back committed writes. A missing owner/journal does not manufacture recovery authority; ordinary repository installation checks pending legacy records and refuses without explicit migration. The corrected receipt reader accepts the writer's protocol 2 and recognizes active installation state.

Executed with race detection from `backend/`:

- Focused capture recovery/coordinator, busy-resource, private-journal corruption and install/adoption tests: passed (5.579s).
- Native CLI inactive-legacy adoption and acknowledged-worker continuity cases: passed (5.524s).
- `git diff --check` on the frozen correction scope: passed.

No introduced behavior defect found in this focus; no additional finding or discovery credit claimed. The assigned evidence does not certify actual home hooks/trust, Windows runtime, power-loss durability, full `make`, or unchanged manager/engine routing, collector and ledger paths. Those remain parent-owned integration or previously bounded evidence. This review is scoped to the C5 correction, not whole-ticket approval.
