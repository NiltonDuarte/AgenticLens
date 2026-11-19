# Native capture release gate

Run from the source revision corresponding to the candidate. The release directory contains the host archive, raw manager and engine, and `SHA256SUMS`. Use the normal `make release-build release-package release-verify` flow for locally packaged artifacts.

```sh
scripts/capture-release-gate.sh --release-dir dist --candidate <exact-v3-rc-tag>
```

The runner requires Python 3.11 or later, Git, tmux, Codex 0.170.0 and OpenCode 1.19.5. Use `--codex` and `--opencode` to select their binary paths. OpenCode uses its normal plugin dependency bootstrap in the isolated home. Codex saves trust through its native TUI after the fixture checks the installed commands. Model responses use a deterministic loopback provider and no hosted credentials.

The fixture adds a table and trigger inside its single private capture database before the root begins. The trigger copies diagnostics inserted by the normal worker before retention removes older rows. The runner reads the copies in sequence so a later commit cannot take the place of the first publication that saw completion. Capture and publication records follow the normal product path.

A successful run prints the two exact session trees, duration checkpoints and candidate digests as JSON. Failed runs preserve their private scratch directory for diagnosis. `--keep-fixture` retains it after success as well. Keep native transcripts and binding messages in that private directory.
