# AL-84 iteration 01 architecture review

CHANGES_REQUESTED. Fresh independent diff assessment of baseline `163181428ee9a5a599a2da1e91d3af5e618a9682` to frozen candidate `2b4118736b5bde24d17cc382da7445c8046ce884`, PR144. Reviewer: reviewer-architecture; runtime reports gpt-6-astra, high effort, matching the requested profile. Run identifier: AL-84 iteration01 / astra-architecture (no separate coordinator run identifier supplied).

## Findings

### ARCH-01 — Unavailable duplicate does not prevent implicit destination selection

Introduced; severity **2**. `backend/internal/adapters/workspace_locations.go:413–425,450–452` records ledger read errors as diagnostics without marking identity resolution incomplete. With two otherwise valid schema-6 ledgers recording `Recorded-99`, put a current-format pending indicator beside one. Unselected `ResolveTicketLocation` returns the other with no error, although the distinct duplicate still exists. This violates explicit duplicate selection and permits automatic publication to choose an unintended copy: `capture_worker.go:225` accepts this resolution and passes its destination to generation. Manual generation without output selection uses the same resolver (`main.go:155`). Diagnostic attachment does not stop either caller.

Proof: isolated `TestAstraPendingDuplicate` writes self-contained valid ledgers and a version-2 indicator with a 16-byte hexadecimal transaction reference, without private recovery data. The assertion requiring an unresolved result fails: the healthy `homes/a` is selected and the pending sibling appears only in diagnostics. Explicit healthy selection passes as a control. The guard against budget/unavailable-root incompleteness works, but only covers enumeration; this failure occurs after enumeration. Invalid-ledger exclusion is deliberately supported and must remain separate from temporarily unavailable reads of supported ledgers. Require explicit selection when identity cannot be established because a candidate is pending/unreadable; do not block valid siblings from listing or explicit use.

### ARCH-02 — Discovery and selection disagree on symlink depth

Introduced; severity **2**. `backend/internal/adapters/workspace_locations.go:190–191` queues a symlink target at its traversal depth, while `workspaceDirectoryPolicy` at lines 520–525 derives depth from the canonical physical path. Configure `homes` at depth 1; put a valid ledger in `homes/nested/deep` and an in-root symlink `homes/alias` pointing there. Discovery and unselected resolution accept the ledger through the immediate alias and expose its physical path. Selecting that returned path fails as outside scope. The dashboard always forwards selected rows' physical paths (`web/src/domain/workspaces.ts:16`), so a listed row cannot open detail. Publication scope uses the same physical-depth interpretation and cannot refresh that discovered home either.

Proof: isolated `TestAstraAliasDepth` resolves the ledger without selection, then passes the returned location back to the resolver; the second call fails with “ledger destination is outside configured workspace discovery scope.” Both target and alias are inside the physical root; the outside-root guard does not refute this case, and budget 300 avoids truncation. Align discovery, explicit selection and publication on the approved scope semantics. Reject an out-of-scope home consistently or honor supported alias coverage consistently; do not advertise an unusable home.

## Other hypotheses and verification

- Retired configuration/default fallback: refuted within inspected paths. YAML-node validation requires explicit integer depth/budget and rejects retired/unknown root fields; writers preserve selected values. Operational state is local-only and read resolution does not initialize it.
- Setup/OpenSpec overwrites or accidental runtime Node dependency: no finding. Inspected exact-release bundle preflight, pinned wrapper/creation helper, schemas and lifecycle guidance. Three real isolated OpenSpec workflow tests passed, including custom-file preservation. Approval/archive guidance is explicitly advisory; roadmap completion stays separate.
- FCIS and manager/engine coupling: no finding in scoped diff. Changed domain files add data fields and pure projection only; filesystem/configuration work remains adapters. Manager implementation is unchanged and its entrypoint delegates to the manager adapter; no new engine/web dependency was introduced.
- Pending/portable read boundary: the selected pending ledger is refused and completed fixtures read without sources/private state. ARCH-01 concerns selection among candidates, not accepting the pending ledger's mixed artifacts.

Executed on frozen product source:

- `cd backend && go test ./internal/adapters/config ./internal/adapters -run 'Test(Workspace|PublicationState)' -count=1` — passed both packages.
- `python3 scripts/tests/openspec_workflow_test.py` — 3 passed, 5.984 seconds.
- `cd backend && go test -overlay /tmp/al84-astra-architecture-fg76i2ta/overlay.json ./internal/adapters -run '^TestAstra' -count=1 -v` — both expected guard assertions failed, reproducing the findings. Probe source, overlay and final output remain in `/tmp/al84-astra-architecture-fg76i2ta/`; no shared source/test edits.

Initial hypotheses were formed from requirements/diff before consulting TEST_RESULTS.md. No sibling/prior review report was read. Reused supplied integrated make success at `273d90aa` and final hosted/Pages checks at `2b41187` only as context, not as proof against these cases. No full make rerun. Native Linux probes only; Darwin descriptor code was inspected, but native macOS behavior and physical power-loss recovery were not exercised. This is a scoped architecture verdict, not whole-PR certification. Preserved the parent's tasks.md edit and other reviewers' files. Parent owns fixes, consolidated disposition, telemetry and publication.

PR https://github.com/NiltonDuarte/AgenticLens/pull/144, head 2b4118736b5bde24d17cc382da7445c8046ce884, checks previously passed; scoped architecture review CHANGES_REQUESTED.
