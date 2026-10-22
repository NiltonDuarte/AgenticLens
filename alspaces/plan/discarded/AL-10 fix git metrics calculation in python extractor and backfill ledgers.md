h3. Title: AL-10: Fix Git Metrics Calculation in Python Extractor and Backfill Existing Workspace Ledgers

*STATUS*: Discarded (Superseded by AL-11)

*DISCARD REASONING*

This ticket was discarded in favor of implementing the native Go telemetry extractor directly in AL-11 (`agentic-lens generate`). Rather than applying temporary maintenance to the legacy Python script (`scripts/extract_agentic_metrics.py`) only to immediately replace and deprecate it with the native Go implementation, all identified fixes (multi-language git numstat parsing, hybrid pattern classification, robust git branch/merge-base resolution, and backfilling) are incorporated directly into the scope and acceptance criteria of AL-11.

*TARGET*: AgenticLens (scripts | python | workspaces | telemetry)

*CONTEXT*

The Python extraction script (`scripts/extract_agentic_metrics.py`) currently records 0 lines added and 0 lines deleted across all recent dev ticket workspaces (e.g. AL-1, AL-2). Investigation revealed hardcoded file path prefixes (`src/`, `tests/`, `infra/`, `cli/`) that filter out Go/TypeScript files, hardcoded `src/` vs `tests/` path tests that misclassify Go `*_test.go` and TS `*.test.ts` files, and reliance on `origin/main` git log queries that miss unpushed local commits. This must be fixed immediately so telemetry accurately reflects code delivered, and all existing ledgers must be backfilled.

*GOAL*

Refactor `scripts/extract_agentic_metrics.py` to support multi-language git metrics calculation (Go, TypeScript, Python) with robust branch/commit resolution, verify calculation on dev tickets, and backfill all existing workspace ledgers with accurate code facts.

*VERTICAL SLICE SCOPE*

* Python Script Refactoring:
  - Remove hardcoded path filters (`src/`, `tests/`, `infra/`, `cli/`) in `get_ticket_numstat`.
  - Add optional configuration loader from `.agenticlens/config.json` for `code_metrics` (`include`, `exclude`, `test_patterns`, `prod_patterns`).
  - Update file classification in `compute_code_facts` (convention-over-configuration):
    - When `code_metrics` is specified in `.agenticlens/config.json`, respect custom glob patterns.
    - When omitted, apply universal multi-language defaults:
      - Test files: `*_test.go`, `*.test.ts`, `*.test.tsx`, `*.spec.ts`, `*.spec.tsx`, `test_*.py`, `*_test.py`, `test/**`, `tests/**`, `__tests__/**`.
      - Production files: `.go`, `.ts`, `.tsx`, `.py`, `.js`, `.jsx`, `.html`, `.css`, `Makefile` (excluding test files).
      - Ignore documentation (`*.md`), metadata (`*.json`, `*.lock`, `go.sum`, `package-lock.json`), and `workspaces/**`.
  - Fix git revision range resolution: check active branch diff against merge-base with main/origin/main, check local/remote branches matching ticket, and scan `git log --all --grep=<ticket_id>` across local commits.
* Workspace Backfill:
  - Run backfill across all historical workspaces (`AL-1`, `AL-2`, `GEN-1`, `GEN-2`, `PO-1`, `PO-2`).
  - Update `agentic_metrics.json` and refresh Markdown summary tables in `SUMMARY.md`.
* Quality:
  - Unit test in `tests/` or standalone test script verifying multi-language file categorization and git numstat parsing.

*ACCEPTANCE CRITERIA*

* Running `python3 scripts/extract_agentic_metrics.py --ticket-id AL-1` populates `code.lines_added > 0`, `code.test_lines_added > 0`, and `code.prod_lines_added > 0`.
* Running `python3 scripts/extract_agentic_metrics.py --ticket-id AL-2` populates accurate non-zero code metrics matching the actual git diff.
* All workspace ledgers under `workspaces/` have accurate `code` blocks backfilled.
* `SUMMARY.md` files for AL-1 and AL-2 contain accurate Test-to-Production Code Ratios and Code Delivered metrics.
* Zero regressions in token extraction, cost calculation, or invocation parsing.

*DEFINITION OF DONE*

* Python script updated and tested.
* All existing workspace ledgers backfilled with valid code metrics.
* Quality checks pass.
