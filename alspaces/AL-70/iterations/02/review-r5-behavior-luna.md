# AL-70 R5 behavior review — Luna

**Outcome: APPROVED (scoped behavior review).** Candidate AL70-R5; base `f63e3db6b2d4baa8e22452e4ec94eff566867721`; entry fingerprint `53216836c57b58aea99790ea61b6556e92a7bdc485bcda24907b4545b975da57` matched the supplied manifest. Requested reviewer identity: Codex `gpt-6-luna`, medium effort; this records the requested profile, not source-observed runtime metadata.

## Scope and findings

Reviewed generated anomaly role text in `web/src/domain/anomalies.ts`, its raw source-trace path through `TicketDeepDive` and `TraceWaterfall`, recorded-summary selection, `AnomalyDrawer`, and the configured role presentation context. No blocking behavior defect found.

Raw and hydrated source anomalies are created with `roleOrigin: 'source'` and passed to the waterfall, which formats their generated title/description, Inspect accessible name, and tooltip through `sourceLabel`. The drawer receives that ownership and applies the same mapping to role headings and generated text. The recorded anomaly summary control formats through `projectedLabel` and stores `roleOrigin: 'projected'`; the drawer therefore keeps recorded controls projected. The mapper only transforms display strings: `formatAnomaly(anomaly)` still returns raw source-role wording, used by debug prompt generation, and does not mutate anomaly context. Missing required rework-loop roles take the existing generic incomplete-context fallback. Direct drawer callers supply the required origin; the view defaults event-only drawer use to `source`.

## Evidence and verification

- Frontend focused suite: 218 unit tests and 173 affected integration tests passed, including the shared native partial-wire fixture for partial and unavailable source responses, one occurrence, projected detail, and preserved raw context/prompt bytes.
- Typecheck, lint, and production Vite build passed. Build emitted the existing large-chunk advisory; current assets were generated.
- Disposable old-formatter mutation: three raw-display guards failed as expected, while the projected-display control passed. This supports the new oracle's ability to distinguish the correction.
- Standalone smoke result `/tmp/al70-r5-standalone-tjq42onb/RESULT.json`: current embedded SPA served for 12 shared-only synthetic ledgers; recorded ledger bytes remained unchanged; invalid and occupied ports exited with failure; default port behavior was observed. This is bounded installation evidence, not live-source or browser coverage.
- The entry manifest fingerprint matched with no file mismatches. The report is outside its fingerprinted input set; parent should verify the same fingerprint at handoff.

R4 data/test scopes and their unchanged rationale are carried forward per the brief; this report makes no new claim about those scopes. The parent still owns integration and `make` verification after integration. This approval applies only to the stated display behavior and inspected candidate.
