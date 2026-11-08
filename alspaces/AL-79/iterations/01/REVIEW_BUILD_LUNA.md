# AL-79 build/distribution review — gpt-6-luna / medium

## Scope and identity

- Requested profile: `gpt-6-luna`, medium. Runtime identity: `gpt-6-luna`, medium. This report is the `build-luna-01` first-pass discovery. No separate dispatch metadata or cost data was supplied, so neither is claimed.
- Reviewed base `4df45967c801824500ac2ea39a8f52fa9cbdcf66` and the checked-out candidate diff excluding `alspaces/`; its SHA-256 is `84aa66679b8b2b643ccdfd3c17148b77a369ea383a546eee372397778218a7de`, matching the requested C2 fingerprint. Candidate metadata is excluded as specified.
- Requirements: AL-79 `CONTRACT.md`, `iterations/01/PLAN.md`, AL-79 ticket specification, PO-20 F1-F6, `TECHNICAL_GUIDELINES.md`, and relevant installation/distribution and portability guidance in `docs/REGRESSION_PREVENTION.md` and the setup skill.
- Focused paths: `config/entrypoint.go`, `config/native.go`, `cmd/agentic-lens/dispatch.go`, worker fence/launch paths, capture executable retention and user installation/activation, both source and shipped wrapper copies, plus relevant callers and tests. Product and test files were read-only.

## Findings

No substantiated introduced build/distribution correctness finding in the Linux paths reviewed.

No pre-existing defect or optional maintenance suggestion is reported in this scoped pass.

## Hypotheses investigated

- **Development output selection and fallback:** `ResolveEntrypoint` requires the shared exact version pin. `NativePath` distinguishes an absent host-named development executable (select exact cache) from an existing but invalid one (error), and validates native format, capture/dispatch protocols, and build identity before selecting development output. The CLI and automatic dispatch both call this resolver. `dispatch_test.go`, `native_test.go`, and distribution tests cover absent, valid-different-identity, invalid-present, missing-pin, and local-override cases. No selection defect found.
- **Cache identity and provisioning:** Cached paths include the exact pin and host tuple; dispatch validates the selected release identity against the pin. The reviewed execution paths contain no download/latest/pin mutation. Provisioning is explicit installer behavior; tests check no request is made during ordinary selection. No fallback or identity defect found.
- **Worker lifetime across rebuild:** Admission copies executable bytes to a content-addressed retained path; the fence checks that retained artifact and operational source/state policy rather than retiring merely because preferred selection changes. The Linux process regression pauses a real worker, changes preferred/install selection, replaces mutable bytes, and checks the later child uses the retained artifact. This is strong evidence for the Linux path; it does not establish Windows behavior.
- **Windows stable activation and installer ownership:** The stable dispatcher remains in place while `dispatcher-installation.json` points to a hash-checked retained native build. `ActiveDispatcher` constrains the active path to the installation root, checks both hashes and validates the native header. User hook groups and machine dispatcher ownership are separately recorded; installation is serialized and transactionally applied. Windows distribution fixtures exercise installer/source behavior, but the supplied evidence explicitly says no Windows native runtime was available. Static inspection and cross-compilation cannot establish loaded-image replacement, child process lifetime, or activation under real Windows execution.
- **Wrapper parity and process forwarding:** Source wrappers and their setup-shipped copies have matching dispatcher routing and bootstrap updates. Native dispatch passes argument vectors and streams to the selected process; automatic callbacks set the validated project cwd while ordinary manual calls preserve caller cwd. POSIX distribution probes cover stdin/stdout/stderr, arguments, relative I/O, exit status and signal forwarding. Windows PowerShell tests cover its quoting/stream/cwd/exit behavior at the script boundary; they do not prove the Go Windows stable activation runtime.
- **Install/update/remove/recovery:** User installation adds only owned groups, preserves unrelated hooks, checks existing dispatcher ownership, and uses recovery/transaction logic. Distribution and adapter tests exercise idempotence, upgrade, modified-owned refusal, foreign-hook preservation, uninstall and interruption recovery. No Linux ownership or transaction regression found. I did not independently run these suites in this review.

## Verification and limits

Reused scoped evidence from `iterations/01/TEST_EVIDENCE.md`, `DISTRIBUTION_EVIDENCE.md`, and `NATIVE_EVIDENCE.md`; inspected the corresponding implementations and test oracles. I ran no tests or probes. The evidence records targeted Go and POSIX checks and a genuine normal-trust Codex 0.167.0 root/child publication on Linux; those native observations support shipped-path integration but are not Windows evidence. Full integrated `make` and race verification remain parent-owned and pending. Windows native runtime remains unavailable; cross-compilation/source evidence only. I also did not independently rerun the integrated build or verify every generated release artifact.

## Scoped verdict

**INCOMPLETE** — no blocking defect found in the reviewed Linux selection, worker-retention, installer, or process-forwarding paths. A defensible build/distribution approval across the required Windows stable activation and build-lifetime boundary needs an actual Windows runtime check (or equivalent accepted platform-lane result). Keep that limitation explicit; source inspection and cross-compilation are not runtime proof.
