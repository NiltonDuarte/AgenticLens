# AL-79 iteration 03 architecture correction review

**APPROVED, scoped; 0 new findings.** Run `AL-79-i03-architecture-sol-c2` (`/root/i03_arch_sol_c2`), correction phase. Base `9c39cf9d4f44bdb968cb92b0212dccf577d10255`; C1 `c84c6b274d12ef53544596f8214c1e5501cf6022883276fc0184bf528320e5cd` to frozen 65-path C2 `d3a52d7385dae492e3c224f76b890234cc673987db543ee0cffefae58b910afe`. The supplied freeze verifier matched before and after investigation. Requested profile: `codex/gpt-6.1-sol/high`; actual runtime model/effort metadata was unavailable.

The approved PLAN03 steering supersedes the original installation design: independent native manager and project-pinned engine, no older-installation compatibility, no runtime POSIX wrapper or engine dispatch. This bounded re-review inspected corrected paths and relevant unchanged dependencies. The assigned C1 architecture reports were consulted; no sibling C2 conclusions were loaded. Product, tests and guidance remained read-only.

## Findings resolved

- **ARCH-SOL-01, severity 6/10: fixed.** `capture/plugin.ts` now calls `callWindows(wrapper, nativeArgs, payload)`, directly spawning the installed native manager. Its deadline, buffer limits and PID-based process-tree cleanup remain. `capture/hooks.go` renders native manager binding commands as PowerShell `& '...exe'`, retaining literal quoting and project selection. The retired Git advisory transport likewise invokes the native executable directly. The production plugin VM and native binding-prompt test passed independently in this review. These establish command construction on Linux, not Windows execution.
- **ARCH-SOL-02, severity 4/10: fixed.** `manager/install.go` defaults self-update to `os.Executable()` and resolves the physical destination through the existing filesystem helper. The owned-file transaction still admits the matching receipt and rejects unknown/modified targets. Independently running the real custom-installation/symlink self-update subtest passed: the invoked physical installation advances, the project pin/cache remain intact, and no default-home installation appears. Discovery credit remains with the C1 reviewer; correction ownership remains with the implementation engineer.

## Architectural assessment

The retained-state exception in `manager/run.go` applies only to `capture inspect-state` with explicit registry and project name. The engine's unchanged inspection handler enforces its complete argument contract and reads the explicit registry/recorded identity without opening normal capture state. Ordinary cross-project capture retains the live-project guard. This preserves engine ownership rather than introducing manager registry logic.

Explicit cache roots and self-update destinations share `filesystem.PhysicalPath`; symlink traversal precedes `..`. Sidecar ownership is checked before journal creation. Foreign local artifacts are refused before cache creation; exact named/checksum-verified foreign published artifacts explicitly defer runtime admission to their target host. These corrections remain adapter responsibilities.

Fresh `go list -deps` for both executables confirms the manager excludes engine CLI, capture, SQLite and web, while the engine excludes the manager adapter. C1 FCIS, one-way callback/source validation, immutable-worker and portable-ledger assessments carry forward: relevant source parsers, domain, publication and worker core are unchanged, and corrected selection does not add ledger authenticity prerequisites. Updated architecture/contributor guidance accurately separates direct engine commands from public manager forwarding; no server-forwarding defect is asserted.

## Verification and limits

Executed: both dependency listings; `go test ./internal/adapters/capture -run '^TestNativeBindingPromptUsesExecutableShellSyntaxAndLiteralArguments$' -count=1`; production plugin Node VM; `go test ./internal/adapters/manager -run '^TestNativeArtifactAdmissionAndStandaloneCompanion$/^self-update_replaces_the_invoked_custom_installation$' -count=1 -v` (4.020s). All passed. Reused revision-bounded C2 tester evidence for remaining correction races, mutation oracles, retained-registry inspection and physical cache roots; no broad suite was duplicated.

No unresolved architecture evidence gap within this scope. Full integrated `make` and actual Linux installation remain parent-owned and pending in the supplied handoff. Windows runtime/loaded-manager replacement and a fresh live hook/trust exchange are uncertified; the approved scope requires no Windows CI gate. Actual home hooks/trust were untouched. This approval does not certify whole-ticket delivery or authorize publication.
