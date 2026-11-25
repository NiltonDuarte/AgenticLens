# FIX-21: Keep delegation child links across capture refresh

## Roadmap ticket

FIX-21 is an ad hoc fix with no roadmap specification or index entry.

## Why

Automatic capture can delete accepted `child_session_id` links when a pass re-reads a parent but not all of its children.

- **How links are lost.** A Codex `spawn_agent` result often names only the child task. Its child is resolved only from transcripts loaded in the same pass. When the byte or session budget leaves a child out, the link comes back empty. The refreshed parent then replaces the accepted one, and nothing reports the loss.
- **The AL-71 incident.** 7 of 9 links disappeared. A full manual generate restored them.
- **Two causes in the same area made the loss likely and persistent:**
  - Identifying unrelated rollouts consumed almost all of the 512 MiB budget. The AL-71 family is about 50 MB, but the budget went on reading headers across a 1,900-file store.
  - A rollout moved to `archived_sessions/` while its row was unassigned. That made every worker pass fail until the worker's lifetime ran out.

## What changes

- **Link retention.** In automatic capture, a refreshed Codex parent keeps an accepted child link that the pass could not re-prove.
  - The link stays on the delegation that matches by agent type, description and invocation time.
  - Fresh child metadata that contradicts a link still clears it.
  - A different child replaces a link only when its parent is confirmed.
  - Retained links count as unresolved child evidence, so the pass stays visibly incomplete.
- **Discovery within budget.**
  - Capture skips child candidates whose UUIDv7 creation instant is earlier than their parent's, and skips exact-lookup files whose name carries a different session ID. Neither is read.
  - Each header is read at most once per pass.
  - Headers draw on a reserved quarter of the budget, and the actual bytes read are charged. Any unused share goes to the pass's final extraction.
  - Identifying unrelated files can no longer starve ticket-owner extraction.
- **Archived unassigned rollouts.**
  - An unassigned row whose rollout moved to `archived_sessions/` is found again, but only when its own metadata proves the same session in the same store, with no ambiguity.
  - Otherwise the row is recorded as unavailable without failing the worker pass.
- **Unchanged.** Manual `generate`, explicit backfill and OpenCode behavior.

## Capabilities

- Added `capture-child-link-retention`: [spec](specs/capture-child-link-retention/spec.md).
- Added `capture-discovery-budget`: [spec](specs/capture-discovery-budget/spec.md).
- Added `capture-archived-unassigned-recovery`: [spec](specs/capture-archived-unassigned-recovery/spec.md).

## Impact

The capture worker and `generate --capture-auto` keep accepted links. A ticket family is extracted within the configured budget even on a large store. A moved unassigned rollout no longer stalls the worker.

These guarantees are preserved:

- portable ledgers
- other-harness facts
- stable ledger ordering
- the FIX-20 equal-facts guarantee
- ticket attribution, which still requires binding

The details are in the [design](design.md).
