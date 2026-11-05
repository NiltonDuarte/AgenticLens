# AL-64 iteration 04 implementation

Production candidate frozen on 2026-11-06 against baseline `d8a7c14cfd9f0ee626fedc33c70f92bd928d9192`; integration commit and final review remain parent-owned. Approval is recorded in PLAN.md. Initial status contained only parent-owned SUMMARY.md and iteration 04 PLAN.md changes. The coder preserved parent, investigator and independent tester changes and made no test edits, commits, staging changes or telemetry calls.

## Delivered behavior

Five approved human diagnostic templates are implemented, with eligible stderr emission after successful explicit ledger/summary publication and diagnostic acknowledgement. Automatic lifecycle generate does not emit checkpoint feedback. The installed command has no verified native caller/current-interval transport, so its production selection uses F03 for exact source-open intervals and F04 for qualified invocation subtotals. F01/F02/F05 have pure selection and rendering support, with positive caller and interval proof required; no production transport claims that support.

The shared `generation.feedback.provisional_threshold_seconds` setting defaults to 180, accepts zero to disable and positive int64 seconds to override, and rejects negative, fractional, quoted, null and unrepresentable values. Existing discovery and local override precedence remain in force. The loader, defaults, YAML writer/default template and generate help include the setting. Bootstrap scaffolds omit it and use the loader default. The default template explicitly initializes the existing enabled metric defaults alongside the feedback threshold.

The selector receives explicit cutoff, immutable session facts, transient source/owner/invocation evidence and independently supplied caller proof. It evaluates exact elapsed milliseconds before invocation rounding; 179999ms is below threshold and 180000ms qualifies. Individually known short intervals do not fall through to F04. It never pools invocation, model or owner rows. Output is deterministic by source, session, role, model, row and native start/ID. Legitimate invocation rows remain distinct; duplicate interval observations within a row emit once, ambiguous row/native attribution is qualified, and F05 is appended at most once after F01.

Source evidence is collected before model grouping and carried through explicit extraction and successful capture source reads. Cached/retained owners have no fresh native evidence. Evidence includes native timing separately from positive current source-open proof, allowing known short timing to suppress feedback without pretending a historical interval remains open. Already completed or superseded source records after the numeric cutoff disqualify source-open wording without changing the established accounting at that cutoff. Missing OpenCode lifecycle/timing keeps evidence incomplete, permitting qualified F04 alongside a valid current exact interval. Codex generated accounting IDs such as turn-N are never presented as native interval IDs; absent native IDs qualify fallback only. Existing Codex source warning failures remain failures.

No ledger/invocation schema, persistence DTO, API DTO, summary renderer or frontend changed. SourceEvents and Codex Transcript explicitly exclude checkpoint evidence from JSON. ConfigHandler continues to use its existing explicit sanitized ProjectConfigDTO, with no generation/feedback fields. Source arithmetic, tokens, cost state and numeric rounding remain unchanged. Optional diagnostic write failure leaves successful generation successful. There is no polling, scheduling, process wakeup, source mutation, completion fabrication or persisted dedup state.

## Owned paths

- `backend/internal/domain/checkpoint.go`: transient evidence, caller proof, typed cases and pure selection.
- `backend/internal/domain/generation.go`, `config.go`: feedback configuration and defaults.
- `backend/internal/domain/models.go`, `events.go`: request-local source evidence and complete-owner replacement propagation.
- `backend/internal/adapters/codex/rollout.go`: native timing/identity collection before grouping, historical lifecycle distinction, transient combination.
- `backend/internal/adapters/opencode/extractor.go`, `schema_v1_session_extraction.go`, `schema_v2_session_extraction.go`, `capture_snapshot.go`: native message timing/current-source qualification and request-local evidence propagation.
- `backend/internal/adapters/capture/reconcile.go`: fresh Codex evidence propagation; retained snapshots stay unqualified.
- `backend/internal/adapters/config/generation.go`, `loader.go`, `writer.go`: strict shared parsing, validation, defaults and YAML templates.
- `backend/cmd/agentic-lens/checkpoint_feedback.go`, `main.go`: exact escaped single-line copy, post-publication stderr emission and help.
- `docs/SETUP_MANUAL.md`: configuration and supported feedback semantics.
- This IMPLEMENTATION.md only within the ticket workspace.

## Verification performed by coder

Environment: Linux shared checkout; source candidate is the working diff against the baseline above. No test files were modified by the coder. Independent tester evidence and the parent's integrated make gate are recorded separately.

- `bin/agentic-lens-linux-amd64 generate --help`: inspected existing built command successfully; it had extraction root/nonce flags and no feedback setting.
- `go run ./cmd/agentic-lens generate --help` from backend: successful fresh build/run; displays the default/disable configuration and extraction-root identity limitation.
- `go test ./internal/domain ./internal/adapters/config ./internal/adapters/codex ./internal/adapters/opencode ./cmd/agentic-lens -run 'Test(Checkpoint|GenerationFeedback|GenerateOmissionPreservesActiveDurationAndDescendants|OpenCodeCheckpoint|RolloutActiveDuration|SchemaV1ActiveDuration|SchemaV2ActiveDuration)' -count=1` from backend: PASS for all five packages after source/current-evidence qualification and synthetic-ID corrections. This covers selector/config/copy/CLI checks plus affected source cutoff regressions available at execution time.
- Earlier focused passes also succeeded for existing generation defaults and source duration boundaries. One intermediate CLI fixture run failed before generation because its required ticket title was absent; the independent tester corrected that fixture and the subsequent focused command passed.
- `git diff --check`: PASS. No frontend changes or asset build is required for this backend-only slice. Parent owns the full integration make gate and frozen-revision review.

## Native capability limits

See the independently owned [RUNTIME_CAPABILITIES.md](RUNTIME_CAPABILITIES.md). Installed Codex 0.166.0 and OpenCode 1.19.5 do not expose a verified generic-generate current-caller/native-current-interval handoff in the inspected transport. Environment IDs, extraction roots, nonces, paths, receipts and model matches are not promoted to proof. Source-open evidence establishes source observations, not executing-process liveness. Advertised continuation surfaces have not been verified through a real natural-completion followed by next-turn telemetry refresh, and no such native cycle was executed or claimed. These limits do not block F03/F04 fallback support.

## CP-REV-01 correction and renewed production freeze

The confirmed review finding (severity 4/10) was independently discovered by the behavior/integration reviewer and data/failure reviewer, both gpt-6.1-sol / high, and source-verified by the parent. A retained v2 running-tool item could override native session idle for numeric activity and incorrectly lend that same state to new source-open feedback. The numeric behavior predates this iteration and remains outside this correction.

The correction changes only `backend/internal/adapters/opencode/schema_v2_session_extraction.go`. An optional bulk read retains native idle timing for checkpoint qualification. NULL remains known active; a positive integer idle timestamp strictly before the current message does not terminate that later interval. An idle at or after the message start disqualifies source-open wording, including after the numeric cutoff. Older schemas without the optional column retain existing behavior. Missing, malformed or failed optional lookup evidence qualifies feedback instead of failing otherwise supported extraction. Source running/DurationInterval arithmetic, tokens and stored facts are untouched. Known short interval timing remains available, preserving suppression before current-source qualification.

Production is re-frozen after this correction; the only other coder-owned update in this correction is this IMPLEMENTATION.md. No tests, staging, commits, telemetry or publication were performed by the coder. The independent tester owns durable boundary regressions and perturbation evidence; parent owns bounded reviewer rechecks and integrated make.

Coder verification: from backend, `go test ./internal/adapters/opencode -run 'Test(OpenCode.*Checkpoint|SchemaV2ActiveDurationSnapshotAndCompletion)' -count=1` PASS, including the new idle-boundary table and existing numeric cutoff/completion coverage; `git diff --check` PASS. This is deterministic adapter evidence, not an authentic native completion/next-turn runtime claim.
