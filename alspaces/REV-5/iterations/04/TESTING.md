# REV-5 iteration 04 independent testing

Verdict: PASS for ADP-01 and D3 changed boundaries. Baseline 88a4a1b plus current implementation and test diff. No implementation, release or publication performed by tester. Other audit findings remain open.

Inspected contract, plan, implementation report, tester persona, technical testing policy and regression checklist. Existing real-parser cases reach persistence with malformed middle JSON, malformed final complete record, unterminated partial write and invalid token usage in a discovered child. They require retaining the whole root/child tree, partial diagnostics, independent healthy-tree refresh and unrelated-harness retention; repaired source retries successfully and valid final records without newline are accepted.

Strengthened the same recovery table to compare retained DTOs completely against persisted baseline (including provenance), reject unknown/duplicate session IDs, and compare complete recovered session snapshots on unchanged retry. Extended the existing release fixture with retargeted-tag rejection and rejection of a missing forced tag despite valid action output. No new production API or test-only production seam.

Verification:
- `go test -race ./internal/adapters/backfill ./internal/adapters/codex`: passed. Backfill executed after test changes; unchanged Codex package reused Go cache.
- `make test-release-workflow`: six tests passed. Actual workflow shell executes against isolated local commits and annotated/lightweight tags; asserts selected checkout and metadata identity, normal output, forced precedence, invalid inputs and unchanged HEAD on rejection. Workflow ordering/side-effect guards are structural checks, not a GitHub expression interpreter.
- Disposable Go overlay removing only the warning guard: the four real-parser persistence cases all failed with missing partial diagnostics, proving the original bug is detected. Production source was never edited by this probe; temporary files removed.
- `git diff --check`: passed.

Limits: no hosted Actions run, real release lookup/upload, tag promotion or release deletion. The forced-release existence command and action conditions were inspected structurally. Full integrated make and standalone build verification are parent-owned. Fixtures are self-contained, with no dependency on ticket artifacts. No silent skips or domain mocks.

Tested file hashes:
- `backend/internal/adapters/backfill/recovery.go`: `3247c491b97c15e838a3b0302db178a6aab68ebe8739006cbb5b586fabd64745`
- `backend/internal/adapters/backfill/engine_test.go`: `51995277cbaad184aca8fc8f947526ad17a5c67aa765c260e92113961d62780f`
- `backend/internal/adapters/codex/rollout.go`: `fc53e5ba7686ae502c1792f4fb2cdb76c3705d46af56672732bc2d8725391832`
- `.github/workflows/release-please.yml`: `4fb4b06a9697dd7d499f874b62e4856d9417263d77d2fb02b68fb78d96341152`
- `scripts/tests/test_release_revision.py`: `c792cc05580435baf02b7bb56bff6e917aad4f555b388879861db48f67bee6cb`
- `Makefile`: `e6c75e58ff671c710e8264e23aba26eb0a995cc09ee067999b8d6ddd17f4e45c`
