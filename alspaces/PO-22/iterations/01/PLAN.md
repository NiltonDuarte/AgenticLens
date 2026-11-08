# PO-22 iteration 01: Active backlog review

## Scope and source

User request: “review all the existing tickets with the new version of PO skill. Rewrite them as needed.” Apply the checked-in agentic-lens-po skill and its current templates. PO-22 was unused across workspaces and the registry. Initial Git status was clean at `731de946d7d4f1d3715a9d647d54889b8871a39c`; initial branch was `PO-#/review-existing-tickets`, renamed to `docs/PO-22-review-existing-tickets`. No subagents or application implementation are required for this planning work.

## Acceptance boundaries

Review all 27 active tickets and registry entries. Rewrite around human outcomes; retain requirements, finding attribution, evidence pointers and dependencies. Preserve accepted consolidation, unrelated/historical numbering and completed artifacts. Eliminate duplicated workflow checklists and concrete code recipes. Do not mark historical findings fixed or claim native-runtime verification. No application/test/runtime/contract edits or developer workspace allocation.

The new skill supersedes the old presentation structure, not accepted product policy. Preserve portable history, measured-zero/unknown separation, source-family limits, exact-owner replacement and legitimate occurrence multiplicity. Newly explicit draft defaults are identified in SPEC.md; development approval is not implied.

## Work and verification

1. Read configuration, current templates, all active tickets, registry, relevant canonical rules and prior PO-19/20/21 context.
2. Rewrite every active ticket and generate a concise registry from its actual status/dependencies; preserve existing IDs and accepted scopes.
3. Check unique IDs, title/file consistency, template structure, local links, matching index/spec status, reciprocal dependency edges and cycles. Compare old/new acceptance to retain scoped outcomes; inspect the changed-file boundary.
4. Commit task-owned planning deliverables; merge origin/main and run required `make`. Record actual failures/skips rather than treating documentation checks as application proof.
5. Generate final Codex telemetry with a literal nonce and docs classification, inspect root/source/session coverage, commit, push and open the PO planning PR.

Selected regression boundaries: configured workspace root; ticket relocation independent of application/tests; portable ledger policy; explicit unavailable evidence; historical versus current source claims. Verification results and delivery limits belong in SUMMARY.md. This is author verification, not an independent code review.

Document verification passed against baseline `731de946`: 27 stable IDs/titles; required section order; valid local links; unchanged status/prerequisite fields; reciprocal acyclic dependency graph; finding fields and all legacy correction/capture acceptance bullets retained. AL2 requirements were reviewed manually against their original scope and acceptance. `origin/main` remained `731de946`; merging after planning commit `2f5a9920` was a no-op. Only AL-78 wording and PO artifacts changed afterward; application/test/runtime inputs to the integrated gate are unchanged.

Integrated `make` completed with exit 0: static checks, Go race/integration lanes, frontend 190 unit + 678 integration tests, release/install/capture/orchestration checks, asset sync and multi-platform builds passed. No gate stage skipped. npm advisory/bundle and React act warnings were observed, not remediated. Application inputs remained unchanged; no native capture proof is claimed.
