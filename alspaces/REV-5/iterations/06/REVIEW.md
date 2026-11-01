# REV-5 iteration 06 independent instruction review

Outcome: APPROVED for the instruction-consistency scope below. This is a child review outcome, not the parent consolidated delivery approval.

## Revision and scope

Baseline: a44000d. Reviewed the uncommitted tracked guidance diff plus new shared context, four focused personas, four Codex registrations and the parallel-review reference. The 17 instruction files have SHA256 content fingerprint `1c73139de979156f597fbc4316b1e6e7819a7ab4e3d1755d831be6e5159f11d4`. Fingerprint inputs are sorted changed tracked and untracked paths under `.agents/` and `.codex/`, plus changed `AGENTS.md` and `docs/REGRESSION_PREVENTION.md`; hash each UTF-8 relative path, NUL, file bytes, NUL. Workspace evidence is excluded because independent reports are being written concurrently.

Read the approved CONTRACT.md and iteration PLAN.md, applicable engineering/regression guidance and existing dispatch/configuration dependencies. Initial status contained the parent-owned guidance changes and new instruction/evidence paths. Only this explicitly assigned report was written. FORWARD_TEST.md and the other exercise conclusions were not read. The personas are review subjects here, not sibling reviewers participating in this instruction audit.

## Findings

No blocking introduced defect was substantiated.

Optional clarity improvement: `.agents/skills/agentic-lens-dev/SKILL.md:91` and `.agents/agent/orchestrator.md:55` still describe consolidated APPROVED as the transition to Phase 3, while final approval requires Phase 3 integrated checks. The explicit preceding dev-skill paragraph and canonical parallel protocol lines 63-69 correctly authorize commit/integration/checks before final approval, so the complete instructions resolve the ordering. Aligning the short loop bullets with that readiness/publication distinction would make them less ambiguous in isolation. This is a suggestion, not an unresolved delivery blocker.

Existing unrelated guidance debt remains outside this change: persona/dev-skill delivery snippets use `git add -A`, while AGENTS.md requires staging task-owned paths; persona kickoff also advances iterations on any resumed ticket. These are unchanged by this diff and overridden by the repository policy. Parent owns any later guidance reconciliation; they do not undermine the newly defined review contract.

## Hypotheses and disposition

- Role/context leakage: each focused Markdown persona loads the dedicated common context and its own investigation lens. The common context leads to the investigation skill without requiring the coordinator matrix. Registrations point to the matching persona and common file. The protocol explicitly limits instructional isolation claims; filesystem/model independence is not promised. No required sibling loading found.
- Missing role availability: local TOML declarations use the existing repository registration shape and their persona pointers resolve. The running session does not expose the four new names. The protocol expressly requires a documented generic-reviewer fallback with the exact persona, so unavailable live registration does not silently lose focus coverage.
- Capacity starvation/omitted focus: protocol lines 7 and 25 require four applicable core focuses for substantive changes, record justified exceptions, reserve parent capacity and queue remaining work in waves. Risk specialists require concrete questions. No fixed four-child concurrency assumption or model-switch authorization found.
- Evidence collision or permission expansion: protocol lines 22 and 38-43 give children unique evidence paths and reserve consolidation/publication for the parent. Shared context retains return-in-response when writes are denied. All five reviewer permission blocks exactly match the baseline generic reviewer block; Codex declarations add no model, sandbox or permission override. Product/test writes remain prohibited.
- Stale approvals or majority dismissal: protocol lines 49-59 retain evidence/disagreement, require verified dispositions, preserve reports by revision and require dependency-based carry-forward or affected re-review. A single confirmed blocker prevents delivery; incomplete required coverage cannot pass.
- Verification ordering: protocol lines 63-69 explicitly permit integrated verification before final approval and invalidate affected evidence after integration. The dev skill states the same sequencing. The shorthand loop bullets merit the optional clarification above, but the canonical gate is executable.
- New persona telemetry names: the unchanged `RoleMappings.ResolveRole` preserves an unmapped normalized role instead of dropping it. Dedicated names are not yet aliases of reviewer in project configuration; that is supported raw-role behavior, not demonstrated data loss or a required change to this documentation-only scope.

## Verification and limits

Executed: `git diff --check a44000d` passed; Python TOML parsing/name/path checks passed for changed/new registrations; 36 relative Markdown link targets across the 17 instruction files exist; reviewer permission-block comparison passed. The initial whole-frontmatter comparison rejected intentionally different persona descriptions; the corrected comparison isolated permissions and passed. The links check validates target files, not every fragment anchor.

Read-only searches covered active `.agents/`, `.codex/`, AGENTS.md, CONTRIBUTING.md and docs dispatch/review references. A probe for `.codex/config.toml` found no such file; existing agents use the same per-agent TOML layout, and no live registration success is claimed.

No application/runtime files changed in this iteration. No runtime suite, remote action, harness restart, native role invocation or effectiveness comparison was executed by this reviewer. This report establishes static consistency and the stated boundary checks, not production defect recall. The independent forward exercise, final integrated verification, telemetry and consolidated delivery verdict remain parent-owned.

## Parent reconciliation

Applied the optional phase-order clarification: focus readiness permits commit/integration/checks, while final consolidated APPROVED permits remaining publication only after checks pass. Clarified shared context loading to include its linked common procedure. No permission or product behavior changed. FORWARD_TEST.md independently exercises scheduling, low-risk exceptions and blocked delivery; it is not runtime role discovery or measured effectiveness evidence. Parent final validators/link checks cover the completed wording.
