# AL-96 native capture evidence

The gate installs checksummed Linux amd64 manager and engine artifacts through the supported setup path. A deterministic provider drives Codex and OpenCode. The actual clients create the sessions, execute tools and record source facts. The destination starts empty. All publication comes from installed lifecycle callbacks and the worker. See the [runner](../../../../scripts/capture-release-gate.sh) and [fixture instructions](../../../../scripts/tests/fixtures/capture-release-gate/README.md).

## Defects discovered during verification

- OpenCode 1.19.5 took 15.068 seconds to publish a completed turn on `v3.0.0-rc.0+al96.dev1` using a 15-second refresh and 500 ms debounce. On `dev3` the comparable run took 1.112 seconds. New durable requests now wake the worker after debounce. The normal periodic wait still covers turns without a usable completion callback.
- Codex 0.170.0 on `dev2` published three owners. The native source had a fourth owner, a child of the implementation agent whose lifecycle callback was absent. Recovery now looks for that child's native rollout using a spawn hint from its verified parent. The child's own metadata establishes the edge.

## Passing development package

Both combinations passed on `v3.0.0-rc.0+al96.dev5` (base `a5c8a25e`, working candidate). Both exact ledger trees matched native source, including the root, two direct children and one grandchild. The negative-control session remained absent. The later normal turn preserved owners and counted cumulative usage once.

| Client | Binding to ledger | Completion to commit | Resume completion to commit | Ticket provisional before / after |
| --- | ---: | ---: | ---: | ---: |
| Codex 0.170.0 | 1.165 s | 3.201 s | 1.594 s | 8 / 0 s |
| OpenCode 1.19.5 | 1.761 s | 26.624 s | 1.313 s | 4 / 0 s |

The OpenCode run recorded a transient SQLite lock failure before the slower commit. The worker logged `SQLITE_BUSY` and the pending request reached its next completed pass within the 30-second gate. This run is evidence for the working candidate. OpenCode 2.2.0 standalone remains manual-extraction-only on the existing evidence.

On the next package, `v3.0.0-rc.0+al96.dev6`, both installed clients passed the stricter ticket-total oracle. The rendered, ledger and native measured totals agreed after completion and resume. Codex measured 7 and 8 seconds. OpenCode measured 17 and 20 seconds. Completion committed in 0.963 seconds for Codex and 15.880 seconds for OpenCode, within the 30-second deadline. Both terminal and resumed ticket totals had zero provisional seconds.
