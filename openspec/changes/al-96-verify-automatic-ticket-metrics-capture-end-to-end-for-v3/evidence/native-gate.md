# AL-96 native capture evidence

The gate installs checksummed Linux amd64 manager and engine artifacts through the supported setup path. A deterministic provider drives Codex and OpenCode. The actual clients create the sessions, execute tools and record source facts. The destination starts empty. All publication comes from installed lifecycle callbacks and the worker. See the [runner](../../../../scripts/capture-release-gate.sh) and [fixture instructions](../../../../scripts/tests/fixtures/capture-release-gate/README.md).

The isolated registry has a fixture-only trigger that copies actual worker diagnostic inserts into an ordered observation table. The gate reads that table without modifying the source, registration, or published artifacts. A later commit cannot substitute for the first commit that observed native completion when the runtime compacts its diagnostic heads.

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

On `v3.0.0-rc.0+al96.dev8`, the complete ordered-history gate passed for both clients. The first completion-aware commit followed native completion by 1.839 seconds in Codex and 2.062 seconds in OpenCode. Each fixture retained 48 actual ticket diagnostics in the observer while normal capture compacted its records to two. The tightened gate checks that the held root accounts for provisional duration at the native cutoff, that finished descendants contribute measured time, and that the first qualifying commit and resumed turn agree across source, ledger and summary.

## Integrated product package

The [checksummed native result](native-result.json) records the passing Linux amd64 package `v3.0.0-rc.0+al96.02ff2608`, product revision `02ff2608` and runner revision `32fd1347`. Both ledgers contained the expected root, two direct children and a native grandchild exactly once. The unrelated owner stayed absent and resume added measured duration without duplicating previous owners or usage.

| Client | Binding to first commit | Completion to first eligible commit | Resume completion to commit | Ticket provisional before / after |
| --- | ---: | ---: | ---: | ---: |
| Codex 0.170.0 | 0.983 s | 1.125 s | 1.226 s | 4 / 0 s |
| OpenCode 1.19.5 | 2.472 s | 1.272 s | 1.052 s | 16 / 0 s |

## Capture before ticket binding

The revised scenario begins without a ticket or destination. It checks a held interval in a Git-visible, unstaged snapshot, completion in the same snapshot, delayed binding, then a native destination request. The fixture starts with an unavailable destination to verify that pending facts survive the failed attempt. Once the directory exists, the worker commits the ledger and removes the pending snapshot. The gate checks exact source facts through promotion and a later resume.

In `v3.0.0-rc.0+al96.pending1`, Codex captured the complete unbound tree but exceeded the 30-second completion deadline. Rediscovery put an already-accepted child back into a pending state and the worker backed off for roughly a minute after the completion request. Making repeated child inheritance idempotent and recognizing refreshed unassigned owners restored timely settlement. The corrected development package `v3.0.0-rc.0+al96.pending2` passed both clients with the original deadline.
