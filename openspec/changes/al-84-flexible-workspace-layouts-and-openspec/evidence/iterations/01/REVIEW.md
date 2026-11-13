# AL-84 iteration01 active review

INCOMPLETE pending integrated verification. Eight independent discovery cells completed against baseline `163181428ee9a5a599a2da1e91d3af5e618a9682` plus frozen implementation manifest `13a044baddc800f298f54fba03ee41d4c09092d41c89228afde1a79520a821a6`. Product/test changes were frozen through discovery; report writes were excluded. Runtime source metadata confirms `gpt-6.1-sol`/high for the four native reviewers and `gpt-6-luna`/medium for the four delegated reviewers. Private working run/finding/source records are under `build/al84-review/`.

| Focus | Sol/high | Luna/medium |
| --- | --- | --- |
| Behavior | [Changes requested](reviews/behavior-sol.md) | [Changes requested](reviews/behavior-luna.md) |
| Data | [Changes requested](reviews/data-sol.md) | [Scoped approval](reviews/data-luna.md) |
| Architecture | [Changes requested](reviews/architecture-sol.md) | [Scoped approval](reviews/architecture-luna.md) |
| Tests | [Changes requested](reviews/tests-sol.md) | [Scoped approval](reviews/tests-luna.md) |

## Confirmed corrections

- BEH-01 (6): canonical raw-generation filename admission. Found by behavior reviewer, Sol/high; native generation succeeded but subsequent listing found no ledger. Root computed export is a separate supported artifact.
- DATA-01 (6): final read can return a replacement ticket under the requested ID. Found by data reviewer, Sol/high; two real directory-exchange reproductions confirmed it. This establishes read misattribution, not persisted overwrites.
- ARCH-01 (7): newly created private journal ancestry lacks durable ordering before public writes. Found by architecture reviewer, Sol/high; verified source-level power-loss window. No power-loss experiment is claimed.
- BEH-02 (6): incomplete discovery permits duplicate first publication. Found by behavior reviewer, Luna/medium; parent native probe confirmed budget1 hid an existing AL-903 and first generation created a second ledger with exit0 and a partial warning. Partial listing success with diagnostics is required behavior; it is not itself a defect. Safe implicit selection and explicit existing-home availability must coexist.
- TEST-01 (5): movement coverage misses the complete transaction atomic-replacement boundary. Found by tests reviewer, Sol/high; mutation survived existing tests but failed the added disposable control. Current product passed that control.
- TEST-02 (3): source-error assertions accept unrelated identity diagnostics. Found by tests reviewer, Sol/high; substituted diagnostic survived the negative test.

Coder owns four product corrections; tester owns regression coverage and the two test corrections. After correction, freeze a new candidate and reassess all affected focus/profile cells. Initial scoped approvals do not refute demonstrated findings or pass the corrected candidate. Required integrated `make`, final consolidated approval and publication remain pending.

BEH-03 (6), verification-phase discovery by behavior reviewer Sol/high: backfill/migration local filtering bypasses corrected shared selection semantics. Parent native dry runs confirmed implicit one-match selection despite incomplete enumeration and rejection of an explicitly selected budget-omitted home. Targeted maintenance integration is required. Preliminary rechecks did not issue verdicts when test fixtures changed; the `c0c81b9d...` manifest is superseded and cannot support final approval. Final source freeze and matching rechecks remain pending.

Final673-file recheck: behavior/architecture/tests approved under both profiles; Sol data approved, Luna data requested changes on a residual same-ID directory replacement. The title-count race alone cannot distinguish a replacement already present at resolution from a later switch. Parent is validating the separate resolve/final-open source window and a correlated physical-snapshot oracle. DATA-01 retains both reporting profiles; no final approval/commit/integration occurred. Source freeze reopened for this bounded investigation.

## Read-bound candidate status

Current674-file manifest `41a2ee0b246538aba898442b36c22015511557d4fca7e231504f6f9a3d068720` matches every source path. All seven canonical findings have corrections and actual regression evidence. Both data and behavior profiles approve the correlated physical read-binding fix; Sol architecture and tests approve, with Luna architecture/tests pending. Prior unchanged scopes have explicit carry-forward rationale in each narrow report. The initial uncorrelated same-ID title count is qualified; the deterministic post-resolution replacement and guard-disabled mutant establish the corrected boundary. No candidate commit or integration has occurred; integrated make remains mandatory before final consolidated approval.

## Candidate review gate

All four read-bound focus/profile pairs are APPROVED against `41a2ee0b246538aba898442b36c22015511557d4fca7e231504f6f9a3d068720`, with no unresolved blocking findings. Their `reviews/<focus>-<profile>-read-bound.md` reports retain narrow-delta and unchanged-scope rationale. The parent independently rechecked all674 source hashes. Seven canonical findings retain discovery attribution in private working records and will be compacted only after integrated verification. Candidate commit/integration may proceed; final consolidated approval remains withheld until required make passes.
