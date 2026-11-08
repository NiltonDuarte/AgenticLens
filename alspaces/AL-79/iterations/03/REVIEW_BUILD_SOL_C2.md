# AL-79 iteration 03 C2 build/distribution correction review

APPROVED within this focus. New findings: 0; highest new severity: 0/10. Run `i03_build_sol_c2`; requested identity `codex/gpt-6.1-sol/high` (runtime model/usage metadata unavailable). Baseline `9c39cf9d4f44bdb968cb92b0212dccf577d10255`; frozen 65-path C2 candidate `d3a52d7385dae492e3c224f76b890234cc673987db543ee0cffefae58b910afe`. `freeze_candidate.py --verify` matched before and after investigation. Product, tests and guidance remained read-only.

The C1 severity-6 custom-installation self-update finding is fixed. `manager/install.go` now resolves the running executable with `os.Executable()` and `filesystem.PhysicalPath` when `--dest` is omitted. The real native regression installs at a custom path, invokes a symlink, updates from version A to B, verifies changed bytes/identity at the physical installation and checks that no default-home installation appears; project pin and engine bytes remain unchanged. My C1 real-binary failure is the before evidence; this C2 regression passed directly.

I also traced explicit cache roots through physical resolution before installation/listing, foreign-local refusal before artifact/cache creation, and checksum-sidecar ownership checks before journaling. Orphan/edited sidecars are preserved. Published foreign artifacts retain exact platform-named download/checksum validation; interface/version validation is explicitly deferred until target-host execution, as the approved C2 boundary specifies. No broader foreign-local admission is claimed.

Executed diagnostic checks passed:

- Focused `go test -count=1 ./internal/adapters/manager -run ... -v`: custom/symlink self-update, foreign-local refusal, ownership/sidecar/lease controls and exact-artifact checksum/HTTP cases; 4.058s.
- Separate native physical-cache-root subtest: install and versions both preserve literal symlink/parent traversal semantics; 3.624s.
- Production plugin Node VM fixture: direct native argv/payload and bounded Windows tree-cleanup construction passed. This is Linux construction evidence, not Windows execution.

Reviewed corrected CONTRIBUTING/ARCHITECTURE commands, native bootstrap, manager maintenance and Windows consumers. Guidance now explains both artifact families and directly invokes the source engine. The earlier claim that the manager cannot forward `server` is refuted: ordinary commands pass through to the selected engine. C2 changes do not alter Makefile artifact names, release archives/checksum recipes, release upload globs, or manager dependency design, so I reuse C1 Linux packaging/checksum/embedded-UI and Linux/Windows compilation evidence. C2 installation/distribution and manager race/mutation results are additionally recorded in TEST_EVIDENCE.md and MANAGER_TEST_EVIDENCE.md.

Full `make`, main integration and actual Linux installation remain parent-owned delivery gates. Complete Windows archives were not reproduced because C1 lacked `zip`; no remote release or native Windows runtime certification was performed. The explicit Windows loaded-manager replacement refusal remains the accepted limitation. These limits do not establish release-wide approval.
