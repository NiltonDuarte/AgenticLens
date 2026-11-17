# How far each installed harness reached

All probes ran on disposable Git repositories with isolated user configuration, source, publication and capture state. The investigators compiled the manager and engine from main `b6ab0471` using a temporary embed placeholder. Local providers drove the native clients. Session IDs, source parentage, tool calls and hook/worker results below were observed in those clients and AgenticLens. Token counters came from test providers. No paid-inference cost was established.

## Codex 0.169.0

The Codex investigator ran `capture install-user` into the isolated home. An untrusted native run created no capture directory. After approving the six installed hook entries through the native UI, `codex exec` drove a native spawn. The capture registry recorded root `01a1eb9e-e52e-71f0-9223-699e6cba1915`, child `01a1eb9e-fc3d-7390-a694-b3a21fcbef98` and the source parent relationship. The end event reached the registry while both sessions were still unassigned. The investigator answered the root's actual binding request with the native manager, its exact session/store and request, GEN-27, title and type. This proves an operator can answer the challenge. The child inherited the binding from the recorded parent.

The worker could collect owners but its ticket-only lookup found zero ledger candidates in the newly created change. It recorded `publication/failed` with “Resolve an existing ledger or explicitly generate its first destination”. For this fixture the investigator disabled capture, then seeded the selected change with:

```sh
engine generate --source codex --session-id "$root" --ticket-id GEN-27 \
  --ticket-type chore --ticket-title 'Native disposable probe' \
  --json-output "$change/agentic_metrics.json" --append "$change/delivery.md"
```

The first publication contained both owners. After re-enabling capture, pinned strict validation and archive synchronized one spec in the disposable repository. A real `codex exec resume --json "$root" …` triggered the worker. Diagnostics recorded committed ledger **and** delivery under `openspec/changes/archive/2026-11-17-gen-27-native-disposable-probe/`. Root input/output advanced from 84/14 to 126/21 and the old directory stayed absent.

The investigator then moved the child rollout outside the configured store and resumed the root. Its counters advanced to 168/28. The entire recorded child object stayed equal, including parent, model and invocation. Summary feedback counted one refreshed owner, one retained older owner and one failed refresh. Diagnostics recorded the child's failure, history retention and committed publication. Missing child evidence remained visible.

## OpenCode project-local plugin and source adapters

The OpenCode investigator installed the current project integration with `capture install --source opencode --repo . --timeout 8s`. It wrote a project plugin, marker and manifest with installing-project paths. The first v1 run produced native source records but dependency bootstrap failed, the model had no binding tool, and no owners registered. The investigator installed the matching dependency in the disposable project with `npm install --prefix .opencode --ignore-scripts --no-audit --no-fund @opencode-ai/plugin@1.19.5` and reran:

```sh
opencode run --auto --model al58local/mock-model --title 'Plugin retry' \
  'For GEN-970 run a child to read marker.txt.'
```

On 1.19.5, the installed plugin delivered the actual binding challenge. The loopback provider called the exposed `agenticlens_bind` tool with the disposable GEN-970 ticket, title, type and echoed challenge. The runtime supplied the session identity. The agent then used the native task/explore and read tools. SQLite, `capture status --source opencode --repo .` and the ledger agreed on root `ses_e145e6983ffevxog4l56XnLRa4` (explicit_agent_response) and child `ses_e145e3350ffeOW32E1ErHM7TWW` (source_parent). The worker acknowledged generation 33, published both owners and retained three historical owners in the already moved directory. The ledger and delivery metrics updated without a subsequent manual generate or reconcile. No former home reappeared.

On 2.2.0, `opencode2 --standalone run --auto --model al58local/mock-model --title 'Disposable v2 probe' 'Read marker.txt.'` produced a native v2 root and read-tool record. With capture disabled, explicit `generate --source opencode --session-id <root> --ticket-id GEN-971 --ticket-type chore --ticket-title 'Disposable probe' --json-output <path>/agentic_metrics.json --append <path>/delivery.md` published both outputs. Omitting the title on first generation failed. The project-local binding tool was absent from the observed v2 tool inventory.

A separate same-project test manually moved a published directory to `openspec/changes/archive/moved-fixture`, hid the v1 database, and selected a v2 store by `--db-path` and exact v2 root `ses_e145f6039ffecNNDsvWwyEi4Ow`. Generating by ticket GEN-970 rediscovered the moved ledger and summary, added that root, and preserved both unavailable v1 session objects exactly, including their parent edge. A different-project v2 source failed with “source project directory conflicts with configured project” and preserved the prior publication.

**Scope.** Only v1 demonstrated automatic project-plugin capture here. The current plugin's event filter includes session creation, update, status, idle and error. Its bind call transports the runtime context's sessionID; the generated base arguments refer to the installing project. A native v2 reader and a working v1 project plugin establish different capabilities. Global plugin installation, actual event-worktree routing and proof for each supported plugin interface remain [AL-80](../../../../../../alspaces/plan/to-do/AL-80%20capture%20opencode%20sessions%20across%20worktrees.md).
