# AL-79 iteration 03 architecture correction review

**APPROVED, scoped; 0 new findings.** Run `AL-79-i03-architecture-luna-c2`; requested profile `codex/gpt-6-luna/medium` (runtime metadata unavailable). Base `9c39cf9d4f44bdb968cb92b0212dccf577d10255`; C1 `c84c6b274d12ef53544596f8214c1e5501cf6022883276fc0184bf528320e5cd` to frozen 65-path C2 `d3a52d7385dae492e3c224f76b890234cc673987db543ee0cffefae58b910afe`. `freeze_candidate.py --verify` matched at review start and end. Scope: manager/engine responsibilities and dependency direction, corrected invocation and physical path consumers, and active architecture guidance. Product/test/docs paths were read-only; only this report was written.

## Findings and corrections

No new blocking architecture defect found. The two C1 architecture findings are corrected in C2:

- **ARCH-SOL-01 (severity 6):** Windows OpenCode and binding-prompt consumers now invoke the installed native manager directly. `plugin.ts` retains bounded output/deadline handling and process-tree cleanup; `hooks.go` renders PowerShell's native `& 'path.exe'` invocation with quoted project and payload arguments. The distribution/plugin VM and binding tests recorded in `TEST_EVIDENCE.md` cover command construction and corrected-oracle mutations. This is not Windows runtime evidence.
- **ARCH-SOL-02 (severity 4):** self-update now resolves the running installation's physical executable path when `--dest` is omitted. The C2 manager fixture covers a custom installation invoked through a symlink, including physical destination replacement and no default-home installation. Existing owned-file and transaction checks remain in the manager adapter.

The C2 paths for orphan/edited checksum sidecars, runtime cache identity, and explicit cache-root symlink-plus-`..` resolution are supported by the manager test evidence and its isolated mutation controls. Foreign local executables are refused; exact named/checksum-verified foreign published artifacts retain the approved target-runtime admission deferral. These remain bounded adapter responsibilities.

## Boundary assessment and evidence

`go list -deps ./cmd/agentic-lens-manager` confirms the manager closure includes its adapter and domain but not capture, SQLite, or web implementations. The engine closure has no manager adapter or manager command. The manager selects configuration through the config adapter, validates and executes the chosen engine, and forwards the original bounded Codex callback to `capture user-hook`; source/session authority and persistence remain engine-owned. The engine's reserved protocol response and manager-side refusal prevent recursive selection. Retained-state inspection is an explicit pass-through exception whose identity and read-only semantics are validated by the engine; ordinary capture still requires a live selected project. Worker retention continues to invoke the admitted immutable engine directly.

C1 FCIS and portable-ledger assessments carry forward: domain/config policy and source/ledger/publishing paths relevant to those guarantees are unchanged; filesystem, process, network, installation, and persistence operations remain outside the domain. The split adds no local receipt or machine identity prerequisite for Git-shared history. `ARCHITECTURE.md`, `CONTRIBUTING.md`, and `TECHNICAL_GUIDELINES.md` now describe direct engine commands separately from supported public manager forwarding, with distinct manager identity and engine pin semantics. No manager/server forwarding defect was found.

Reused C2 evidence: `go test -race ./internal/adapters/manager -count=1` (10.663s), distribution tests (23 passed), installation tests (10 passed), focused capture advisory/binding race checks, retained-state engine integration, and `git diff --check`; all are recorded as passing in `MANAGER_TEST_EVIDENCE.md` / `TEST_EVIDENCE.md`. The dependency listings and source traces above were performed in this review. No broad suite was repeated.

Integrated `make`, actual Linux installation, and the live home-hook/trust exchange remain parent-owned/pending. Windows runtime and loaded-manager self-update remain uncertified as allowed by the approved scope; no Windows CI gate is required. This scoped architecture approval does not certify whole-ticket delivery or publication.
