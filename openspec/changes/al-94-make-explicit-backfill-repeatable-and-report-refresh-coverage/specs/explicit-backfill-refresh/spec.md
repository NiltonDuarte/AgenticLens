## Purpose

Operators can repeat an explicit source refresh safely and identify which recorded sessions still need original evidence. See the [AL-94 scope](../../proposal.md).

## ADDED Requirements

### Requirement: Repeat exact-source backfill without redundant ledger writes

Backfill with an explicit harness and `--recompute-db` SHALL verify exact recorded sessions even for completed ledgers. It SHALL preserve equivalent ledger bytes, `updated_at`, and modification time across repeated runs. Genuine changes to persisted facts or migration state SHALL use the existing safe publication path.

#### Scenario: Equivalent complete or partial repeat

- **WHEN** a repeat against unchanged source evidence produces equivalent persisted facts and migration state
- **THEN** the ledger bytes and modification time stay identical, including its timestamps and any pending steps

#### Scenario: Source changes with incomplete owner coverage

- **WHEN** original evidence changes a selected owner's facts while a recorded child remains unavailable and another harness has stored facts
- **THEN** the available owner's genuine change is published, the unavailable child and unrelated harness facts survive, and an equivalent repeat leaves the ledger unchanged

#### Scenario: Combined source and Git refresh

- **WHEN** the operator requests both source and Git recomputation with a valid ticket and code ref
- **THEN** a genuine Git fact change is published even if the source facts are equivalent, and an equivalent repeat leaves the ledger unchanged

#### Scenario: A pending step becomes resolvable

- **WHEN** the original evidence satisfies a pending migration step under the existing migration requirements
- **THEN** the migration state advances and an equivalent repeat leaves the resulting ledger unchanged

#### Scenario: Stored-fact maintenance

- **WHEN** the operator runs ordinary backfill or uses Force without explicit source recomputation
- **THEN** ordinary selection remains as before and Force retains its existing stored-fact maintenance behavior

### Requirement: Compare requested summaries independently

Backfill SHALL compare the final rendered summary with existing content and SHALL update only requested summaries whose contents would change. It SHALL preserve authored text and use the existing publication safeguards.

#### Scenario: Identical or unrequested summary

- **WHEN** the requested summary already matches the rendered content or the operator has not requested a summary update
- **THEN** the summary file stays untouched, regardless of whether the ledger changes

#### Scenario: Missing or stale summary with equivalent ledger

- **WHEN** summary maintenance is requested and the summary is missing or its managed metrics are stale
- **THEN** backfill creates or repairs the summary, preserves any authored text, and leaves the equivalent ledger bytes and modification time intact

### Requirement: Report verified and retained sessions

Backfill SHALL report coverage per ticket and selected harness independently of migration status. Each recorded owner SHALL count once as refreshed with changed facts, verified unchanged, or retained. Retained owners SHALL be identified by session ID and reason. Newly discovered sessions SHALL be counted separately, and unselected harnesses SHALL remain outside the denominator.

#### Scenario: Partial refresh on a completed ledger

- **WHEN** a completed ledger has a readable parent, an unavailable child, and facts from an unselected harness
- **THEN** the report distinguishes the parent's verification from the retained child and names the reason for retention, reports incomplete coverage alongside completed migration status, and excludes the unselected harness from the owner totals

#### Scenario: Empty versus retained collection

- **WHEN** a complete source read verifies an empty collection, or a read lacks the evidence needed to replace a stored skill/task snapshot
- **THEN** the report distinguishes verified emptiness from retained or unavailable collection evidence, even when session facts were refreshed

#### Scenario: All selected owners are unavailable

- **WHEN** the selected store opens but no recorded owner can be verified
- **THEN** the report identifies the retained owner IDs and supported reasons such as a missing session or unsupported source layout, preserves their facts, and keeps the existing exit behavior

#### Scenario: Batch read budget exhausted

- **WHEN** the shared source-read budget prevents a batch from verifying some owners
- **THEN** the batch totals reconcile with per-ticket and per-harness results, the report identifies budget-limited owner IDs, and the operator can retry the affected exact ticket and harness, using an explicit workspace if required to disambiguate the ledger

### Requirement: Report publication independently from processing

Backfill SHALL preserve the existing meaning of `Updated` and report committed ledger and summary writes separately. A missing selected store SHALL still fail the command and identify the requested work blocked before publication.

#### Scenario: Summary-only repair and unchanged repeat

- **WHEN** an explicit refresh repairs a summary without changing the ledger and then repeats the same inputs
- **THEN** the first run reports one summary write and zero ledger writes, the repeat reports zero writes to both, and the existing successful-processing count retains its meaning

#### Scenario: Selected store missing

- **WHEN** an explicitly selected store is missing, including when both stores were requested and only one is available
- **THEN** the command preserves the existing failure exit behavior, identifies the unavailable store and other requested work that could not run, reports zero publication, and leaves accepted files intact
