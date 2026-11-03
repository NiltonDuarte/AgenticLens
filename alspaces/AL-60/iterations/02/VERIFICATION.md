# AL-60 final-candidate verification

Current source/test/CI fingerprint: 5e332cad1bfb0d1b25af2d9a8def5bfbdcd8e2b8364c3023131733041d4c8493 against checkpoint 7b00cc1. Source is frozen during final review. Parent owns integrated verification and approval; source-specific focused results below do not replace full make.

## Actual Windows execution

Parent confirmed WSL2 interoperability and executed Windows PE test binaries on the Windows host, rather than using Wine or claiming cross-compilation establishes runtime behavior.

From backend, GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c -o /tmp/al60-corrections/capture-final-windows.test.exe ./internal/adapters/capture passed for the current source. Executed:

`/tmp/al60-corrections/capture-final-windows.test.exe -test.run '^(TestInstallWindowsInspectionOwnsLauncherTree|TestInstallerCompletedRecoveryRetainsExpiredAcceptance)$' -test.v -test.count=1 -test.timeout=60s`

Exit 0. Ownership launcher-exit/cancel passed0.43s; recovery Codex/OpenCode canceled/deadline subcases passed0.38s. Raw result: evidence/WINDOWS_FINAL_RUNTIME.txt. Prior Windows ownership runtime proof is retained in evidence/WINDOWS_RUNTIME.txt. These are parent-executed runtime observations, distinct from coder cross-compilation and hosted CI. Windows race instrumentation and hosted CI were not exercised locally.

## Focused deterministic correction evidence

Retained INSTALL_RED/GREEN, CORRECTIONS_ADAPTER_RED/GREEN, CORRECTIONS_SIGTERM_RED/GREEN and RECOVERY_RED/GREEN logs plus IMPLEMENTATION.md and TESTING.md identify the executed commands and outcomes. All files live in this iteration directory. Historical full gate failures remain in iterations/01/MAKE_INITIAL.log and MAKE_BOUNDED.log; their source and reported external corroboration remain in iterations/01/CAPTURE_DIAGNOSIS.md.

## Integrated gate

Pending final focus/model review completion, implementation commit, origin/main integration and required full make. Standalone embedded assets and AL-55/search preservation will be checked on the integrated candidate. No final approval or publication is asserted yet.
