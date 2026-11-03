# AL-60 final-candidate verification

Integrated source/test/CI fingerprint: 3251b53bc2bfdbd9c7e6ea47d5b8e793fc22f4f9e6103ee268d9fbc622a005db against checkpoint 7b00cc1. Source is frozen during final review. Parent owns integrated verification and approval; source-specific focused results below do not replace full make.

## Actual Windows execution

Parent confirmed WSL2 interoperability and executed Windows PE test binaries on the Windows host, rather than using Wine or claiming cross-compilation establishes runtime behavior.

From backend, GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c -o /tmp/al60-corrections/capture-final-windows.test.exe ./internal/adapters/capture passed for the current source. Executed:

`/tmp/al60-corrections/capture-final-windows.test.exe -test.run '^(TestInstallWindowsInspectionOwnsLauncherTree|TestInstallerCompletedRecoveryRetainsExpiredAcceptance)$' -test.v -test.count=1 -test.timeout=60s`

Exit 0. Ownership launcher-exit/cancel passed0.43s; recovery Codex/OpenCode canceled/deadline subcases passed0.38s. Raw result: evidence/WINDOWS_FINAL_RUNTIME.txt. Prior Windows ownership runtime proof is retained in evidence/WINDOWS_RUNTIME.txt. These are parent-executed runtime observations, distinct from coder cross-compilation and hosted CI. Windows race instrumentation and hosted CI were not exercised locally.

## Focused deterministic correction evidence

Retained INSTALL_RED/GREEN, CORRECTIONS_ADAPTER_RED/GREEN, CORRECTIONS_SIGTERM_RED/GREEN and RECOVERY_RED/GREEN logs plus IMPLEMENTATION.md and TESTING.md identify the executed commands and outcomes. All files live in this iteration directory. Historical full gate failures remain in iterations/01/MAKE_INITIAL.log and MAKE_BOUNDED.log; their source and reported external corroboration remain in iterations/01/CAPTURE_DIAGNOSIS.md.

## Integrated gate

APPROVED after all final focus/model reviews and full make passed on057d410253aecd734551995ff1bedb38e94a1f19. Four-CPU affinity, GOMAXPROCS=2 and GOFLAGS=-p=1 bounded concurrency without changing product/test timeout defaults. Verbatim output: evidence/MAKE_VERIFIED.txt. Every Go package passed race checks; all786frontend tests passed; five platform binaries built. Parent verified standalone index/assets against the embed tree and search/no-match strings. Integrated Windows binary ownership and recovery cases passed: evidence/WINDOWS_INTEGRATED.txt. Main added only AL-68workspace prose; later explicit defer transformation is observationally equivalent and carries source approvals with vet/focused guards. Search and AL-55 remain unchanged. Hosted CI and browser geometry are not claimed.
