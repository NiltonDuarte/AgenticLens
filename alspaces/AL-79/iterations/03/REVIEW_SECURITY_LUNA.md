# AL-79 iteration 03 security review — gpt-6-luna/medium

**Outcome: APPROVED (scoped security review; no blocking findings).** Requested run identity: iteration 03 SECURITY / gpt-6-luna/medium. Candidate baseline `9c39cf9d4f44bdb968cb92b0212dccf577d10255`; frozen candidate hash `c84c6b274d12ef53544596f8214c1e5501cf6022883276fc0184bf528320e5cd`, matched by `freeze_candidate.py --verify` at start and end. Review covered the approved manager/engine split, installer/update admission and hook boundary, not the whole repository.

## Findings

None. No introduced security defect supported by the approved requirements and inspected paths.

## Hypotheses investigated

- **Untrusted/replaced install targets:** `replaceManager` serializes updates with a no-follow lease, requires the prior executable to match its ownership receipt, stages a journal, verifies each target still matches its recorded before-image, and replaces through same-directory atomic writes. Recovery validates the target and restricts journal paths to the executable and its two metadata siblings. Engine cache admission rejects non-regular files, verifies bounded artifacts and checksums, validates native role/platform/version, and requires `--force` before changing different bytes.
- **Unexpected code execution/network during ordinary routing:** `Run` routes project commands only through local configuration and exact installed/development engine selection; it does not download or edit pins. Network fetches are limited to explicit install/update paths. Release names are exact versions and repo components are validated in the Go downloader. The bootstrap shell script likewise requires an explicit version for first install, verifies the selected artifact against the release checksum list, and invokes the native manager for configuration and installation.
- **Hook payload/project confusion:** manager `codex-hook` bounds input to 64 KiB, requires absolute cwd and nonempty event/session fields, derives project from the shared config found at cwd, validates local source configuration and exact engine selection, and forwards the original body to the engine’s authoritative hook parser. Engine hook processing revalidates Codex store/transcript context before capture. I found no manager bypass of the disabled-policy or malformed-present-config checks.

## Evidence and limits

Static source trace: `backend/internal/adapters/manager/{install.go,storage.go,run.go,lock_unix.go}`, `backend/cmd/agentic-lens-manager/main.go`, `backend/cmd/agentic-lens/capture_user_hook.go`, `backend/internal/adapters/config/entrypoint.go`, the POSIX bootstrap script and `backend/internal/adapters/capture/user_install.go`. Candidate inventory and raw diff were available at `/tmp/al79-native-entrypoint-0ta97kas/{CANDIDATE.json,CANDIDATE.diff}`; relevant newly added manager and hook files were read directly. Reused the recorded `MANAGER_TEST_EVIDENCE.md` result: manager race tests passed with checksum, ownership, interruption and wrong-role probes. I ran no tests. No Windows runtime or live Codex trust exchange was exercised; Windows manager self-update is explicitly deferred by the approved plan. This report does not certify machine installation or integrated `make`.
