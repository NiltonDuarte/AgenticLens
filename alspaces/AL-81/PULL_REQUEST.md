# fix: generate manual metrics without capture registry access

Manual `generate` currently opens capture state and can be blocked by machine-local bindings even when authoritative source transcripts and the requested output ledger are available. This change separates manual extraction from automatic capture admission.

Manual generation uses project, metrics and selected harness source configuration without reading or initializing the capture registry. It retains source/project validation, portable owner refresh, shared publication serialization and interruption recovery. `--capture-dir` remains accepted for compatibility and has no manual state effect. Automatic worker generation retains capture configuration and registry ownership checks; explicit session/nonce selectors cannot bypass registered-owner selection.

Validation is in progress. Final results, real Codex root/child parity evidence and any limits will be recorded before publication.

See [contract](CONTRACT.md) and [summary](SUMMARY.md).
