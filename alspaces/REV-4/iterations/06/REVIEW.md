# REV-4 iteration 06 documentation review

Documentation-only self-review against baseline `790e14c555d29e819dae51bb26a6849824d7bc2a`. User authorization is recorded in CONTRACT.md. No independent development cycle is needed for this scope; no new subagent was launched. The root contract, scoring catalog, test-removal history, and accepted assertion transfers remain unchanged.

## Evidence translated into guidance

| Observed evidence | Shared guidance and disposition |
|---|---|
| [Removal history](../../REVIEW.md) records local JSON/type-cast cases, fixture-only assertions, copied logic, repeated hashing and duplicate smoke coverage. | Root technical guidance requires a production contract and plausible regression. It distinguishes self-confirming fixtures from actual codecs, JSON tags, runtime validation, and compiler contracts; no blanket ban on serialization tests. |
| [Iteration 04 review](../04/REVIEW.md) records release-verification removal after user-reported CI problems. Workflow/help text did not exercise release success; two missing-input tests did exercise real rejection paths. | Verification must match its claimed boundary. Narrow guards do not prove packaging or publishing. Their accepted deletion reduced narrow coverage; documentation does not claim improved release coverage or automatically authorize replacement CI work. |
| [Iteration 04 review](../04/REVIEW.md) records date-smoke deletion in favor of stronger existing date boundaries and role-order assertions merged into an existing enum test. | Prefer coherent existing coverage while preserving independent failure oracles; identical feature names across layers do not establish redundancy. |
| [Candidate ledger](../05/CANDIDATES.md) records all 47 low-score cases: seven deletions, 35 transfers and five independent retained guards. [Review](../05/REVIEW.md) verifies the transfers and retained scores. | Scores prioritize inspection. Require explicit deletion rationale or meaningful-assertion-to-recipient mapping. Retain independent cases without a coherent equivalent; do not inflate scores or create unrelated large tests to clear a threshold. |
| [Score review](../../SCORE_REVIEW.md) distinguishes focused body/peer inspection from provisional screening; helper and compiler assertions provide real proof despite superficial appearance. | Inspect helpers and compiler directives and state confidence limits. Passing tests alone do not establish equivalent defect detection or mutation equivalence. Numeric scores are optional when requested, with explicit anchors and rubric. |
| [Inventory correction](../../REVIEW.md) restores a multiline Go declaration missed by regex; [iteration 05 verification](../05/TEST_RESULTS.md) distinguishes AST declarations, expanded frontend cases, helpers and unchanged recipients. | Use syntax-aware inventory for suite audits, path/line identity, baseline comparison and defined fingerprint inputs. Report actual execution rather than treating static declarations as runtime case counts. |
| The session's ratio discussion found different whole-tree source and configured counts. Current code_metrics.test_patterns omits web/src/test support paths while prod_patterns matches their TS files. | README and the technical guide require counting/classification assumptions and distinguish whole-tree versus diff-added counts. A ratio does not establish quality. Configuration remains unchanged; no classifier fix or universal target is claimed. |
| Active tester instructions demanded all generic edge cases; the review skill broadly preferred memory fixtures, required microsecond timing, and showed automatic fixture replacement/parallelism examples. | Personas and skill guidance now select missing relevant contracts and preserve real adapter proof. Timing is measured under comparable conditions; parallelism requires independent state. This corrects misleading guidance, not evidence of a measured performance defect in every affected test. |
| The skill-local technical guide was a stale full copy of the root guide. | Replace it with a compatibility pointer to the canonical owner and current command/checklist references. Preserve the path for existing consumers. |

The approach follows [REV-3's completed documentation audit](../../../REV-3/iterations/02/REVIEW.md): evidence stays in ticket artifacts, rules live in shared owners, and consumers link to them. No new ticket-specific references appear in shared guidance. Documentation reduces ambiguity; it neither implements runtime enforcement nor proves historical defects fixed.

## Scope and consistency review

Root TECHNICAL_GUIDELINES.md owns test value, boundary proof, and metric interpretation. The prevention checklist owns the review procedure; CONTRIBUTING.md provides contributor steps. AGENTS.md, README.md, all five development personas, and the development/review skills link or route to those owners. Unrelated architecture checks, role permissions, skill metadata and lifecycle gates remain intact.

Searched active skills, persona guidance, docs, setup and PO specification templates for conflicting test instructions. The setup/templates do not generate the removed test-volume or fixture-replacement requirements, so they need no edit. Existing operator dashboard advice already requires checking ledger coverage before interpreting trends. Historical plans remain historical.

No application/test/schema/runtime/configuration/dependency/generated-asset changes are part of this iteration. No scores or earlier approval decisions change. No new test cases or performance study are required to validate prose. Boundary and inventory recommendations apply when that work is in scope, not as mandatory new matrices for every ticket.

## Verification

Results and integrated baseline are recorded below after execution. Self-review focuses on consistency, meaningful scope, correct references and preservation of existing implementation; it is not an independent re-audit of the entire test suite.

- Documentation checks passed across 18 Markdown paths, including 13 shared guides: all 53 local links/anchors resolve; no new shared ticket references or emojis. Skill/persona frontmatter and permissions are unchanged. Both modified skills pass skill-creator quick_validate.py. git diff --check passed.
- Compared the removed skill-local duplicate against the canonical guide: common architecture/invariant content is retained by the canonical owner, which also has current paths, asset prerequisites and newer fidelity rules. The compatibility pointer avoids a second stale copy.
- Source/config/Makefile diff against the baseline is empty. Setup/PO templates and all historical ticket artifacts are unchanged. The earlier broad fixture/parallelism/microsecond instructions and ratio-quality claim are removed; current scope-based rules remain explicit.
- Self-review accepted the documentation delta. No mutation-equivalence, runtime-policy enforcement, new release assurance or measured performance improvement is claimed. Full integration results follow.
