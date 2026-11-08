# AL-79 iteration03 security discovery

CHANGES_REQUESTED — scoped artifact replacement, invocation and hook trust review. Run `i03_security_sol`, independent discovery; requested identity `codex/gpt-6.1-sol/high` (runtime model metadata unavailable). Base `9c39cf9d4f44bdb968cb92b0212dccf577d10255`; frozen 58-path candidate SHA-256 `c84c6b274d12ef53544596f8214c1e5501cf6022883276fc0184bf528320e5cd`. Candidate verification matched at start and end. Requirements: iteration03 plan, preserved contract and technical guidelines. New untracked manager/hook files were read directly; no sibling reports were consulted.

## Finding

**I03-SEC-SOL-01 — severity 5/10: foreign-platform installs bypass executable role and version admission.** In `backend/internal/adapters/manager/install.go:126`, `ValidateNativeExecutable` and exact-version comparison run only for the executing host. The foreign branch checks format/architecture, writes the cache leaf and reports `Installed engine` at line 163. The newly exposed `--platform` plus `--local-bin` combination therefore accepts a separate manager as an engine, or accepts bytes from a different version. A checksum establishes byte integrity, not the required executable role/version.

Actual Linux probes compiled the candidate manager for amd64 and arm64 with version `v9.4.2`. From an isolated unconfigured directory and temporary home, the amd64 manager successfully ran `install v9.4.2 --platform linux-arm64 --local-bin <arm64-manager> --sha256 <correct-digest> --install-root <temporary-cache>`. The installed engine leaf exactly equals the manager bytes. Repeating with requested `v9.4.3` also succeeded. `go version -m` identifies the supplied artifact as `agenticlens/cmd/agentic-lens-manager`, arm64, with the `v9.4.2` linker setting. Results: `/tmp/al79-security-sol-efavklcq/ROLE_ONLY_RESULT.json` and `RESULTS.json`.

This violates preserved staged role/version admission and creates a broken foreign cache under a false success result. Host-side selection would later reject these bytes; no arbitrary-code or trust bypass is claimed. Preserve foreign provisioning while enforcing identity before claiming successful installation; if identity cannot be established for a particular local foreign artifact, reject that operation explicitly. Add a behavioral rejection oracle for this branch.

## Assessed boundaries and evidence

Source tracing found explicit maintenance is the only manager download path. Local checksum verification copies into a private stage before admission. Manager replacement rejects modified/unowned primary executables; bounded regular-file reads, journal path allowlisting, before-image checks and kernel leases refute straightforward replacement/recovery escapes. Hook routing preserves original payload bytes; engine validation still checks runtime source, rollout store and authoritative session context. Hook installation preserves unrelated groups and reports trust as unverified; no implicit trust mutation was found.

Executed selected manager recovery, ownership/lease, HTTP checksum, local-checksum and disabled-hook tests with `go test -race ... -count=1`: PASS, 1.706s. Native hook probes rejected oversized payload, relative cwd and missing session identity; an unconfigured project returned `{}`. Probe files are outside the checkout and actual home/hooks were untouched.

Existing low-impact behavior: a fresh manager installation overwrites an orphan `target.sha256` regular file without an ownership check (`storage.go:92`); reproduced in `RESULTS.json`. The baseline downloader already replaced this sidecar, so this is recorded as existing debt, severity 2/10, rather than a new blocker.

Full `make`, real machine installation/live hook trust and Windows runtime verification remain parent-owned. This review does not certify those boundaries or whole-repository correctness.
