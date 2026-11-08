# AL2-2: Archive workspaces without losing history

## Goal

Operators can compact completed ticket workspaces into archives while continuing to browse and query their historical telemetry.

## Context

Long-lived repositories accumulate completed ticket folders. JSON formatting from AL-76 does not archive those folders. Archiving must respect configured workspace roots and retain portable history, artifacts and ordinary directory-based access.

## Acceptance criteria

* agentic-lens archive supports --before, --workspaces, --output, --keep and --dry-run. Date selection includes only proven completed tickets before the cutoff; missing completion evidence never becomes an inferred date. A terminal invocation without --before offers interactive selection; noninteractive missing selection fails clearly.
* Create .tar.gz packages with a manifest and verified checksums/entries. Publish the verified archive atomically before removing source folders. Failure preserves originals; --keep retains them; --dry-run makes no writes or deletions. Never delete a source changed since the archived snapshot.
* Scanning and ticket/artifact reads discover archives under the configured workspace/archive locations and read telemetry, SUMMARY.md and CONTRACT.md without disk extraction. Unarchived folders take precedence over archived copies of the same ticket.
* Workspace API results identify archived records with is_archived. Dashboard rows show an Archived badge, detail shows archive origin, and All/Active/Archived filtering retains historical queries and metrics.
* Corrupt packages and unsafe archive entries produce diagnostics without losing valid sibling tickets. Bound archive reading and reject path traversal; preserve supported ledger facts and unrelated files.

## Implementation notes

Archive creation/streaming belongs to adapters; eligibility and duplicate precedence belong to the pure core. Completion comes from established ticket data, not folder age. No raw transcript capture or new history-authenticity requirement is included.

Follow [engineering rules](../../../TECHNICAL_GUIDELINES.md) and [regression boundaries](../../../docs/REGRESSION_PREVENTION.md). [PO-19](../../PO-19/SPEC.md) preserves the earlier development reconciliation; historical observations require revalidation, not an assumption of current runtime support.
