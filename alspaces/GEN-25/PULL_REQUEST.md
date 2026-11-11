# chore(GEN-25): remove Windows support

AgenticLens now supports macOS and Linux on amd64 and arm64 only. Remove Windows native implementations, executable/PE handling, PowerShell bootstrap, hook/plugin transports, distribution targets and Windows execution suites. Update active setup/operator/contributor guidance and bundled installer contents.

Unsupported platform requests fail before downloading or changing installation state. Releases package tar.gz archives and select only supported binaries/archives for checksums and uploads, even when stale Windows files remain in the output directory. Exact prior generated advisory hooks can still be refreshed or removed; edited hooks remain protected.

Historical records, portable ledgers, user configuration ownership and supported argument/cancellation/atomicity behavior are preserved. Integrated `make` passes, including Go race suites, frontend checks/tests and all four manager/engine targets. Real release packaging/checksums, standalone host identities and disposable before/after regressions also pass. Native macOS execution and hosted release publication were not exercised.

See the [contract](CONTRACT.md), [plan](iterations/01/PLAN.md) and [summary](SUMMARY.md) for acceptance, review attribution and verification limits.
