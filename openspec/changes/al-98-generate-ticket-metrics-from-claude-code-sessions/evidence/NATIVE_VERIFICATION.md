# AL-98 native verification (tasks 7.3 and 7.4)

The candidate reads real Claude Code 2.1.331 transcripts correctly. Usage, children, roles, delegations, skills and tool counts for this ticket's own session tree match an independent oracle exactly. Repeated runs are idempotent, and the standalone binary serves and renders the resulting ledger. The `claude -p` probe confirmed the `sdk-cli` entrypoint and native compaction. It also exposed one duration and steering defect, caused by local slash-command records (finding F1).

- **Verifier:** tester persona, native-verification run, Claude Code agent `acf98b05392fa1a74` (itself a child of the verified root)
- **Date:** 2026-11-24
- **Tested revision:** `0f106ae7` (clean `git archive` export built outside Git; the concurrent tester's uncommitted test edits are not part of the build)
- **Build:** canonical `npm run build` (tsc, vite, asset sync into `backend/internal/adapters/web/dist`), then `CGO_ENABLED=0 go build ./cmd/agentic-lens` with the Makefile ldflags. The binary reports version `al98-native-0f106ae7`, and the embedded bundle validates `source === "claude"`.
- **Claude Code:** 2.1.331

All ledgers, configuration, publication state, the oracle scripts and the probe fixture repository stayed in a disposable scratch directory. Nothing was written into the repository or into `~/.agenticlens`. No transcript content, prompt or output is recorded here.

## Method

- **Independent oracle.** A disposable Python script reads the JSONL files directly. For each owner it skips `forkedFrom` and `<synthetic>` records and counts each response once, keyed by `requestId` or else `message.id`. It sums input, output, cache read and `output_tokens_details.thinking_tokens`, and takes the largest single-response input + cache read + cache creation as peak context. A child is proven when its meta `toolUseId` matches an `Agent` tool_use in the tree and its records carry the root `sessionId` and its own `agentId`. A separate turn oracle re-derives AL-64 turns, durations and steering. It shares no code with the adapter.
- **Ownership workaround (7.3).** The root's recorded `cwd` is this repository checkout. A first generation therefore has to run from this checkout, and its ledger would land inside the repository. To avoid that:
  - **Run A** read the real `~/.claude` store from a scratch project. It discovered, read and extracted the whole tree, then stopped as designed with `Source ownership conflict: ... source project directory conflicts with configured project`. No ledger was created.
  - **Runs B and C** used complete-line snapshots of the live tree. In each snapshot, only the root transcript's `cwd` prefix was rewritten to the scratch project. A record-level comparison showed the snapshot identical to the source in everything except `cwd`. The oracle always read the unmodified snapshot.

## 7.3 This ticket's root and native subagents

The verified root is `b918d950-b1d1-42d6-8d61-a688c68333da` (entrypoint `t3-code` on every record). It has five proven native children:

- planner `a6bbc56a9508cd15e` (resumed through SendMessage)
- coder `a6908b3b63f013f22` (resumed through SendMessage)
- coder `a1cf292b04ba03662` (resumed through SendMessage)
- tester `af94d85cf2d5b0ae8`, which was still running
- tester `acf98b05392fa1a74`, this verifier, which was still running

Generation printed `Discovered 6 complete Claude Code session(s)` and no child diagnostics. It printed two checkpoint notices for the testers' open turns.

### Per-owner usage: oracle and ledger (snapshot 1)

"Dup recs" counts records that repeat a response's usage and were collapsed. Every row matched exactly, with no conflicts and no missing usage.

| Owner | Role | Responses | Dup recs | Input | Output | Cache read | Reasoning | Peak context |
|---|---|---|---|---|---|---|---|---|
| b918d950 (root) | (root) | 33 | 34 | 78 | 19,476 | 2,792,477 | 2,469 | 121,950 |
| a6bbc56a9508cd15e | planner | 91 | 79 | 182 | 87,438 | 17,109,915 | 32,036 | 314,118 |
| a6908b3b63f013f22 | coder | 88 | 101 | 180 | 117,418 | 18,609,764 | 34,888 | 319,463 |
| a1cf292b04ba03662 | coder | 42 | 38 | 84 | 22,669 | 3,851,918 | 5,193 | 135,298 |
| af94d85cf2d5b0ae8 | tester | 27 | 22 | 54 | 11,986 | 3,286,973 | 6,126 | 177,384 |
| acf98b05392fa1a74 | tester | 20 | 17 | 40 | 13,595 | 1,396,175 | 4,924 | 100,151 |

Other checks against the oracle:

- **Membership and parents.** Session membership equals the oracle's set of six. Each child's `parent_id` is the root, and each role equals its native `agentType`.
- **Models and cost.** There is one invocation row per owner, keyed by the served model `claude-opus-5-5`. Three `<synthetic>` client-error records contributed no usage and no row. Every invocation records `cost_evidence.state: unavailable`.
- **Tools, delegations and skills.** Per-owner tool call counts equal the oracle's. The root records five `Agent` delegations, each with a proven `child_session_id`, and one `Skill` activation (`agentic-lens-dev`).
- **Turns and duration.** The turn oracle agrees with the ledger on every owner. The root has 5 turns, 1,749 s and steering 2: its three task-notification turns do not count as steering. The resumed planner has 2 turns and 1,067 s, and the resumed backend coder has 2 turns and 1,559 s. Each tester has one open turn whose provisional time equals the cutoff minus the turn start.

### Repeat and refresh

- **Unchanged repeat.** Two back-to-back runs on snapshot 1 produced byte-identical ledgers. A third run 89 s later differed only in the cutoff-derived fields: the ledger `completed_at`/`updated_at` and the two testers' `provisional_seconds`/`active_duration_seconds` (229 to 318 and 244 to 332). The provisional time was replaced, not added.
- **Live growth (run C).** A second snapshot was taken later. Each file in snapshot 1 is a byte prefix of the same file in snapshot 2, and only the two tester transcripts grew. Refreshing the same ledger from snapshot 2 kept six sessions. The root, planner and both coders stayed byte-identical. Both testers' new totals equal the oracle on snapshot 2 (tester `acf98b05…`: 54 responses, 108 / 42,184 / 5,603,745 / 14,909; tester `af94d85c…`: 39 responses, 78 / 51,830 / 5,731,931 / 16,744).
- **Backfill.** `backfill --source claude --recompute-db` on snapshot 2 matched run C except in cutoff-derived fields. A second, unchanged backfill left the ledger byte-identical. Backfill then ran from the older snapshot 1, which stands in for a restored, complete transcript. It replaced every owner with the smaller snapshot-1 totals, which equal the oracle. The output reports only `DB updated` and `Incomplete: 0`, with no explicit statement that totals shrank (observation O2). Backfill needs explicit `telemetry.capture.max_sessions` and `max_source_bytes`, as documented.
- **Local setup.** `init --local-only --source claude --claude-home <path>` wrote `default_source: claude` and `sources.claude.home`. It required the explicit `--capture-enabled` choice, as designed.

### Standalone binary read

The binary ran `server` on a free localhost port and was stopped afterwards. `GET /api/workspaces`, `/api/workspaces/AL-98` and `/api/fleet` returned 200. All six sessions came back with `source: "claude"`, and the aggregate totals equal the sum of the oracle rows: input 618, output 272,582, cache read 47,047,222 and reasoning 85,636. `GET /api/workspaces/AL-98/events` returned 200 with `availability: unavailable`, all six owners uncovered and an `unsupported source` diagnostic for each. A headless Chromium driven through the DevTools protocol loaded the embedded dashboard. The fleet view showed the AL-98 row, and the Ticket Deep Dive rendered the Claude Code sessions with unavailable cost and the unsupported-source trace diagnostics. The browser logged no exceptions or console errors.

## 7.4 `claude -p` probe

The probe ran in a scratch Git repository with its own AgenticLens configuration, the default store `~/.claude` and isolated publication state. The parent session's `CLAUDE*` variables were unset, so the probe ran as an independent `sdk-cli` session rather than as a child of this session. It used model alias `haiku`, and Claude Code's own cumulative estimate for the whole probe was about $0.0055.

1. `claude -p --session-id 985a4490-c5bc-4fc7-8ce7-e59ce49e0b9c` asked for one `Agent` call. This produced root `985a4490-c5bc-4fc7-8ce7-e59ce49e0b9c` and a background native child, `aa6eb7a828948a248` (`general-purpose`, served `claude-haiku-5-5`). The child's completion notification started a second root turn (`promptSource: system`).
2. `claude -p --resume 985a4490-… "/compact"` compacted the session.

Observed:

- **Entrypoint.** Every record in the root and child carries `entrypoint: "sdk-cli"`.
- **Compaction.** It kept the same file: the inode was unchanged, and the file grew from 54 to 78 lines by appending, with no new transcript file. It wrote exactly one `system` record with `subtype: "compact_boundary"`, `compactMetadata` and `logicalParentUuid`, followed by an `isCompactSummary` user record. The ledger records `compactions: 1` on the root.
- **Usage.** Both owners match the oracle exactly.
  - Root: 3 responses, input 6, output 308, cache read 50,797, reasoning 149, peak context 21,773.
  - Child: 1 response, input 2, output 4, cache read 0, reasoning 0, peak context 13,301.
  - The child is proven with the root as its parent, and the delegation references it.
- **Generation output.** It printed `Discovered 2 complete Claude Code session(s)` and no diagnostics. Both invocations record unavailable cost.

## Findings

### F1 (defect, medium): local slash-command records start phantom turns and steering

- **Trigger.** Run a local slash command such as `/compact` in a session, observed here through `claude -p --resume <id> "/compact"`. Claude Code then appends two user records with no `isMeta`, `promptSource` or `isCompactSummary` field: a `<command-name>` record and a `<local-command-stdout>` record. Neither gets an assistant response.
- **Mechanism.** `readPrompt` in `transcript_jsonl.go` classifies both records as `entryPrompt`. Each one counts as root steering and starts a turn. The second supersedes the first, and the last stays open forever.
- **Evidence.** The probe root's ledger shows `completed_at: null` and `provisional_seconds` 28, then 31 on a repeat 3 s later. The value grows with every refresh, although the session had ended. `user_iterations` is 3, while the person sent 2 prompts (the initial request and `/compact`), and the stdout record is not a prompt at all.
- **Expected.** The root has no open turn, about 5 s of measured duration and steering of 1 or 2. It is 2 if the typed `/compact` counts as a human prompt.
- **Scope.** The ticket's own `t3-code` root has no such records, so 7.3 is unaffected. Any interactive or `-p` session that ends with a local command is affected, and every local command adds steering.
- **Coverage.** No synthetic fixture covers these records.

### Observations (not defects against the delta spec)

- **O1: compaction usage is not in transcripts.** The compaction summarization request writes no assistant record and no usage field. Only `cost-state` reflects it. The probe's `cost-state` changed by +2,160 input, +600 output and +21,771 cache read across the compaction, so the ledger undercounts compaction calls by that amount. This is a source limit: `cost-state` is correctly excluded by design. It belongs in the capability matrix (6.3) and the cost semantics guidance.
- **O2: shrinking replacements are not called out.** A refresh that replaces snapshots with smaller totals reports only `DB updated`. Reviewers should judge whether this meets the spec's "the replacement is reported".
- **O3: the turn count excludes turns with no served response.** Coder `a1cf292b04ba03662` had two measured turns. The first lasted 0.7 s and ended at a `<synthetic>` client-error record. The ledger records `turns: 1` but includes both turns' time (280 s). This is consistent with the adapter's per-model turn accounting.
- **O4: the root invocation role is empty (`""`), as for a Codex root.** Children carry their native `agentType`.

## Limits

- **Snapshot copies for 7.3.** Runs B and C used snapshot copies whose root `cwd` was rewritten. The real store was exercised end to end up to the ownership check (run A) and was read directly by the probe. A first publication from the unmodified store into this repository's own change directory is left to the final telemetry step (8.3).
- **Live root.** The ticket's root kept running throughout. Equality claims cover frozen snapshots, and growth between snapshots was shown to be append-only.
- **Not observed natively.** Forks (`forkedFrom`), nested subagents, Skill calls from subagents and failed Skill calls were not observed. The `cli`, `sdk-ts`, `sdk-py` and `claude-vscode` entrypoints are NOT RUN.
- **Probe transcripts.** They remain in the default store under the scratch fixture's project folder, outside Git.
