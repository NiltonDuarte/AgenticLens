# OpenCode capture verification

OpenCode 1.19.5 loaded the installed TypeScript plugin and published three independently bound root sessions with their native children. AgenticLens used the shared worker. The v1 SQLite records, private registration, published ledger and generated summary were inspected independently. The native engines and manager were built from the AL-80 candidate over `b6ab0471`. SHA-256 for the concurrent run was `c407c74a06a746806691fb13cc37912cfc020cdb5a5d40e7c5738979e409a941` for the engine and `8f85769938244c8b5866fbc76f22bb63120f949878db0fad02ab7e4bb1e37c01` for the manager.

## Native conditions

A disposable Git checkout had two linked worktrees. Config, cache, data, private state, tools and publication destinations were isolated. The effective global directory held the plugin installed by `capture install-user --source opencode --repo <base> --entrypoint <manager>`. A sibling observer recorded plugin context. A deterministic inference endpoint on loopback returned ordinary OpenAI-compatible tool requests. OpenCode itself executed the tools, created session identities, wrote the SQLite records and dispatched the events. No session ID or parent ID was supplied to the binding or delegation tools.

The long-lived server ran in a fourth directory. Concurrent `opencode run --attach <server> --dir <session-directory> --model <local-model>` calls invoked the root binding tool with the known ticket, title, conventional type and its native request challenge. OpenCode's task tool spawned each child. Each child's source row pointed to its corresponding root, and the registry recorded inherited binding. Both configured destinations contained the ledger and summary. The linked worktree's six owners included an earlier pair before the two concurrent roots.

| Ticket | Session directory | Native root | Source-proven child |
| --- | --- | --- | --- |
| TEST-800 | Linked checkout | ses_e1468aa25ffeyGg0NumPEVMVYH | ses_e14686ad9ffeQ6qLV4rBmDE44H |
| TEST-800 | Same linked checkout | ses_e1468aa42ffetiJdoY2UEIFyAp | ses_e14686d8fffeyGhLgUJz4OLzKM |
| TEST-801 | Nested directory in base checkout | ses_e14689e80ffe83vLabSI4L15bU | ses_e146867edffenlEBHYsyS5kdeO |

The enabled linked checkout had no local file. It used the base file's relative source and private-state paths. The other worktree's complete local file disabled capture. Native creation of its nested session `ses_e146c7435ffeumI3WOMETRazW2` produced no private registration or worktree state there. In the enabled base checkout, `ses_e146c7580ffeMhxYMF6hkYtgvj` produced a locally verified snapshot with an empty ticket and no published owner. Native `opencode db path` confirmed that a relative `OPENCODE_DB` is resolved under the OpenCode data directory.

OpenCode 1.19.5 also loaded two distinct global directories in one server. One observer was installed in `$XDG_CONFIG_HOME/opencode/plugins` and another in `$OPENCODE_CONFIG_DIR/plugins`. A real `POST /session` delivered the same `session.created` event for `ses_e1460b00effeemhReRrWJvKO1k` to both observers. The linked-worktree migration regression checks that after switching to an empty custom config directory, the worker still recognizes the XDG installation and legacy install preserves the retired project plugin state.

A separate OpenCode server received `opencode run --attach <server> --dir <disabled-worktree> --session ses_e1468aa25ffeyGg0NumPEVMVYH`. OpenCode accepted the resume and initialized the requested directory. For this real session, all 11 observed callbacks including system transform, status and idle reached the enabled birth-worktree's plugin instance. Zero reached the disabled instance.

## Interface and portable history boundaries

The rendered 1.19.5 plugin passed `tsc --strict --noEmit --target es2022 --module nodenext --moduleResolution nodenext --skipLibCheck` with the runtime's actual SDK and Node types. A separate native OpenCode 2.2.0 installation produced v2 SQLite session `ses_e146a7e91ffeU7z12D3TO70mXe`. The engine's manual generation produced one OpenCode owner and one invocation from that database. With the same installed plugin in its isolated effective config directory, `opencode2 --standalone plugin list` reported no plugins. Automatic plugin coverage demonstrated here applies to 1.19.5.

A separate pair of disposable repositories exercised copied history. The first added a synthetic Codex rollout to the actual v2 OpenCode ledger. The second received the mixed ledger and its own local Codex source. Manual generation replaced 51 input tokens with 9 on the exact Codex owner. The persisted OpenCode owner remained structurally identical. The second repository configured only its own publication state and Codex source.

The new CLI regressions passed through real manager/engine entry points, covering additive adoption, duplicate detection, edited-plugin preservation, explicit linked-checkout migration, whole-file selection, absent shared config, and source disagreement before registration. Focused regression runs passed for both source schema families, cache/engine selection, Codex hooks, copied and concurrent history, publication budgets, retries, retained workers and manual generation. Hosted inference, provider pricing and additional OpenCode releases were outside these runs.
