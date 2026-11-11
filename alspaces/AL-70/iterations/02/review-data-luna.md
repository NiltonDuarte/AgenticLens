# AL-70 Data Review — DISCOVERY

## Scope and identity

Candidate `AL70-R1`, base `f63e3db6b2d4baa8e22452e4ec94eff566867721`; checkout remains at that commit with the supplied AL-70 working-tree changes. Assigned focus: data integrity and failure handling. Requested execution: Codex `gpt-6-luna`, medium; actual runtime identity was not observable in source evidence. No sibling review conclusions were consulted.

## Findings

None substantiated in the assigned focus.

## Consequential hypotheses investigated

- **Missing/invalid policy or source, source failures, and empty refresh must not erase facts.** `runGenerate` validates the selected source/policies before output, loads the baseline under publication ownership, and fails on source open/query/extraction and fresh Git errors before publication. Retained recorded owners are added after refresh decisions; only explicit fresh complete owners replace same-owner facts. Empty source extraction cannot create an authoritative empty transaction. Backfill marks budget exhaustion, missing sources, incomplete trees, and per-owner failures incomplete, retaining affected owners; complete exact-owner snapshots replace even when older or smaller. Other owners and repeated row occurrences survive. Reviewed `main.go`, backfill `engine.go`/`recovery.go`/`source_budget.go`, and owner replacement paths.
- **Mixed stores and portability must not manufacture ownership or require private receipts.** Generation retains canonical ledger history when local sources are absent and overlays successful exact owner refreshes. Backfill separates OpenCode, Codex, and other owners, protects cross-source ID collisions, and gates completion on required recovery. No private capture registration is required to preserve portable facts. These paths are consistent with the frozen portable-ledger and refresh requirements.
- **Migration/read and transaction behavior must not silently clean data or expose partial output.** Ordinary ledger ingestion validates raw current-schema bytes before in-memory normalization; A11 former metadata cleanup is confined to explicit backfill preparation. Explicit current cleanup validates original duplicate-member input and raw facts before writing. Backfill and generation use the bounded prior-good publication journal, restore all authorized siblings on write failure, and leave the marker when recovery cannot complete. Relocation remains an explicit, scoped operation with protocol/layout checks. No auto-migrate or auto-repair ordinary read path was found.
- **Summary and disabled metrics must not publish misleading partial transactions.** Generation resolves its effective summary before collection and includes ledger plus requested summary in the same journaled publication set. Backfill resolves configured summary destinations and uses the same rollback protocol. Optional metric omission preserves stored categories; unknown/unavailable cost and absent versus empty data remain represented distinctly in the reviewed DTO/domain paths.

## Evidence and limits

Read-only source trace covered `backend/cmd/agentic-lens/main.go`, `capture.go`, `backfill/{engine,recovery,source_budget,current_ledger_preparation,publication}.go`, `adapters/{publication_set,repository,generation_policy,fact_ledger_dto}.go`, capture relocation/store paths, and domain read normalization. Reused supplied raw evidence: source-boundary Git/OpenCode/Codex/backfill run (`/tmp/al70-source-boundaries-8_015izt/final-test.log`, exit 0, recorded inputs); targeted tester and CLI logs; and fresh-binary loopback evidence (`/tmp/al70-standalone-6t_y516j/RESULT.json`) showing valid fixtures/read bytes unchanged for absent/invalid optional local policy. These are checkpoint results, not a newly executed integrated gate. Full `make` remains pending with parent. No real developer-source fixtures were supplied. I ran no tests or product mutations.

**Scoped outcome: APPROVED.** No blocking data-integrity finding or material focus-specific evidence gap was identified. This does not certify other review focuses or the pending integrated gate.
