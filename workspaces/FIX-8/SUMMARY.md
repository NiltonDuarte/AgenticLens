# FIX-8 Work Summary

## Adapter

Updated Codex backfill recomputation to discover each stored root tree once and replace its full refreshed session set only after discovery, parsing, transcript combination, and domain grouping succeed. Failed refreshes retain the stored tree; unrelated Codex and other-source sessions remain unchanged.

## Tests

Added coverage for root-plus-child replacement, single-session replacement, discovery and parse failure preservation, no duplicate IDs, and preservation of unrelated Codex and OpenCode sessions.

## Verification

- `go test ./internal/adapters/backfill` — passed.
- `go test ./...` — passed.
- `go test -race ./...` — passed.
- `go vet ./...` — passed.
- `git diff --check` — passed.
- Independent review — APPROVED. Review noted no direct test of combine/grouping failure branches; those failures occur before the staged replacement mutates stored sessions.
