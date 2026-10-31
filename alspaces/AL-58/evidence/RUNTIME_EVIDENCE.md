# AL-58 Runtime Evidence

Captured on 2026-10-29 on Linux x64 from isolated OpenCode runtimes using [mock_openai_server.py](mock_openai_server.py). The server is local, deterministic, uses no provider credentials, returns one `read` tool call and result, supplies 23 prompt and 11 completion tokens per response, and returns a short assistant reply. It reads only the `marker.txt` fixture in the isolated project directory.

## Pinned runtimes

- **v1:** `/home/niltonduarte/.opencode/bin/opencode`, reports `1.19.5`.
- **v2:** `@opencode/cli@2.2.0` with `@opencode/cli-linux-x64@2.2.0`, installed persistently under `/home/niltonduarte/.local/lib/opencode-v2`; `/home/niltonduarte/.local/bin/opencode2` reports `opencode v2.2.0`. The supplied v1 binary was not overwritten.
- **Latest selection:** on 2026-10-29 `npm view @opencode/cli dist-tags --json` returned `latest: 2.2.0`. The registry-resolved version is pinned here for reproducibility; later runs must resolve and review the current dist-tag again rather than treating this value as timeless.

## Isolation and invocation

Each run used unique `HOME`, `XDG_DATA_HOME`, `XDG_CONFIG_HOME`, `XDG_STATE_HOME`, and an absolute `OPENCODE_DB`. The v1 databases were under `/tmp/al58-v1-1.19.5-clean/home/data/`; v2 databases were under `/tmp/al58-v2-2.2.0/home/data/`. No user's default OpenCode database was opened. The retained fixtures replace the generated database path, project path, session IDs, message IDs, and part IDs with placeholders/stable fixture IDs. Raw databases remain outside the repository.

The matching provider configuration samples are [provider-v1.json](provider-v1.json) and [provider-v2.json](provider-v2.json). OpenCode v1 uses a singular `provider` config with `@ai-sdk/openai-compatible`; v2 uses `providers` with `@opencode/ai/providers/openai-compatible`. Each configuration selects `al58local/mock-model`, and each run selects that model explicitly on the CLI.

Run the deterministic provider in one terminal:

```sh
python3 alspaces/AL-58/evidence/mock_openai_server.py --runtime v1
# In another terminal, start the v2 provider instead for a v2 run:
python3 alspaces/AL-58/evidence/mock_openai_server.py --runtime v2
```

For each version, create separate data/config/state directories and a project containing `marker.txt` plus the matching provider JSON copied as `opencode.json`. Set the five isolation variables to those directories and an absolute database file. Then run from that project's directory:

```sh
/home/niltonduarte/.opencode/bin/opencode run --auto --model al58local/mock-model --title 'AL-58 v1.19.5 evidence' 'Read marker.txt and report its contents.'
/home/niltonduarte/.local/bin/opencode2 --standalone run --auto --model al58local/mock-model --title 'AL-58 v2.2.0 evidence' 'Read marker.txt and report its contents.'
```

## Observed compatibility boundary

- v1.19.5 wrote `session`, `message`, and `part` rows. The runtime's `session` table includes model/version and aggregate cost/token fields. `message.data` stores a role plus assistant model, path, timing, cost/token, and finish data; `part.data` stores text, tool state/result, and step events. The `session_message` table exists but had no rows for this run.
- v2.2.0 wrote `session_v2` and `session_message` rows. The fresh database had no legacy `session`, `message`, or `part` tables. `session_v2` contains model/version, aggregate cost/token fields, and v2-specific fork/idle/suspended state columns. `session_message.data` stores `user`, `assistant`, and `idle` variants; assistant content contains typed text or tool items, token/cost data, finish reason, and timing.
- Both versions completed a `read` tool call and stored its successful result. v1 tool arguments use `filePath`; v2 uses `path`. Both runs stored 23 input and 11 output tokens per assistant response. Session totals were 46 input and 22 output tokens because each run made a tool-request response and a final response. Cost was `0` for this custom model. The evidence therefore verifies zero-cost persistence, but does not verify nonzero price calculation.
- The tool-result text and final response are version-produced payloads from the local mock. These fixtures do not establish native-provider pricing, imported v1-history behavior in v2, partial migration, duplicate handling, or behavior for other v1/v2 patch releases.

## Fixture files

- [v1-1.19.5.json](v1-1.19.5.json): sanitized session, message, part rows and runtime schema declarations.
- [v2-2.2.0.json](v2-2.2.0.json): sanitized session_v2 and session_message rows and runtime schema declarations.

The evidence is specific to the two pinned versions above. It is not a claim that every historical or future OpenCode database uses these exact shapes.
