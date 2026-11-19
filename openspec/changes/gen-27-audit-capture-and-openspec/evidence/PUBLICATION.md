# Publication after a change moves

The publication investigator exercised the backend at `b6ab0471` in a disposable Git repository. The source-built engine used a temporary embed placeholder. These probes cover CLI/worker behavior. The dashboard was outside the probe.

**Manual generation, real CLI with synthetic rollouts.** The investigator ran the supported creation helper, `node scripts/openspec-change.mjs GEN-27 "publication probe"`, and supplied a root and a child Codex rollout in an isolated store. Capture was disabled and the publication state directory was separate. The engine call was:

```sh
engine generate --source codex --codex-home "$store" --session-id probe-root \
  --ticket-id GEN-27 --ticket-title 'Publication probe' --ticket-type chore \
  --json-output "$change/agentic_metrics.json" --append "$change/delivery.md"
```

Pinned strict validation and `scripts/openspec.sh archive gen-27-publication-probe --yes --json` passed and synchronized one requirement. Using the returned archive path, the investigator deleted the child rollout and raised the root's input count from 17 to 29. Running the same generation command on the archived ledger saw one available owner. The schema-6 GEN-27/chore ledger retained both Codex owners. The entire child object remained equal, including input 7 and parent `probe-root`. The root updated and authored delivery text survived. No old directory or capture registry appeared. After removing the root rollout too, generation exited 1 with no rollouts. Ledger and summary hashes both remained unchanged.

**Periodic worker, independently of an explicit output path.** A disposable variant of `TestCaptureNativeWorkerPublishesMultipleRootsAndRefreshesWithoutGenerate` configured a second search root and moved the first publication from `records/TEST-100` to `archive/arbitrary-archived-name`. With two registered synthetic roots initially contributing 10 each, the investigator raised one to 50. A periodic pass wrote 60 into the moved ledger. The summary existed there and the old directory stayed absent. The test supplied no destination or new notification after the move. The production worker calls `ResolveTicketLocation(cfg, ticket, "")` on each pass before spawning `generate --capture-auto`.

Existing tests for rename during commit/recovery, lease scope, and deduplicated physical roots also passed:

```sh
go test -count=1 -v ./internal/adapters ./internal/adapters/capture \
  -run '^(TestPublicationTransactionFollowsRenamedDirectory|TestPublicationLeaseFollowsRealRenameAndRejectsOutsideScope|TestWorkspaceDiscoveryDeduplicatesPhysicalScopesAndReadsAbsoluteRoots)$'
```

The unmodified multiple-root worker control passed its publication assertions but failed its five-second terminal idle-shutdown assertion twice, including a standalone rerun with `go test -p 1 -count=1 -v ./cmd/agentic-lens -run '^TestCaptureNativeWorkerPublishesMultipleRootsAndRefreshesWithoutGenerate$'`. The moved-home control completed shutdown. These observations establish a reproducible failure on this host. The cause and production impact remain unresolved.
