# Native capture release gate

Run from the source revision corresponding to the candidate. The release directory contains the host archive, raw manager and engine, and `SHA256SUMS`. Use the normal `make release-build release-package release-verify` flow for locally packaged artifacts.

```sh
scripts/capture-release-gate.sh --release-dir dist --candidate <exact-v3-rc-tag>
```

The runner requires Python 3.11 or later, Git, tmux, Codex 0.170.0 and OpenCode 1.19.5. Use `--codex` and `--opencode` to select their binary paths. OpenCode uses its normal plugin dependency bootstrap in the isolated home. Codex saves trust through its native TUI after the fixture checks the installed commands. Model responses use a deterministic loopback provider and no hosted credentials.

The root performs and completes native work without a ticket or destination. The runner checks its provisional and settled facts in the same Git-visible pending file. It then resumes the root to bind the ticket, indicate a still-unavailable directory, and create that directory in a later turn. The provider holds the repair turn open while the runner checks the exact first ledger and summary against native activity and the earlier pending facts. The runner checks pending cleanup before releasing that turn, then verifies its settled publication and a subsequent refresh. It compares every Git index entry and blob ID across these stages, including a deliberately staged unrelated file.

The default gate also discards and restores the same capture before binding and after its first canonical publication. Codex submits the native control challenge under workspace-write, and OpenCode calls its installed native control tool. In each cycle, the root resumes and runs a real verification tool while the capture is discarded. The runner waits for the worker to acknowledge that native wake, checks that the pending file stays absent and checks the canonical ledger and summary bytes during the bound cycle. After restore, it compares the same capture and its complete owner tree, tools, usage and duration against the native sources, including work completed during discard.

The gate uses the configured pending directory and drives all publication through the installed callbacks. It allows 60 seconds for first publication once the native destination repair begins, and 30 seconds for settlement after native completion.

The fixture adds a table and trigger inside its single private capture database before the root begins. The trigger copies diagnostics inserted by the normal worker before retention removes older rows. The runner reads the copies in sequence to tie pending snapshots and canonical ledgers to the first passes that saw native completion.

Use `--prepared-destination` to reproduce the earlier gate for a historical candidate that required operator preparation.

A successful run prints the two exact session trees, duration checkpoints and candidate digests as JSON. Failed runs preserve their private scratch directory for diagnosis. `--keep-fixture` retains it after success as well. Keep native transcripts and binding messages in that private directory.
