# MGT-2 iteration 02: approved shared capture repair

Source: explicit user question-tool answer on 2026-11-03, “Authorize AL-60 shared capture repair (Recommended)”.

Changed acceptance: AL-60 may repair a deterministically confirmed shared capture defect before the remaining merges. It owns installer optional/required deadline separation, precise failure diagnostics, disposable deterministic harness fixtures, and separate worker-success versus deliberate-expiry coverage. Preserve pending publication, isolation and ownership guarantees; do not blanket-inflate global timeouts. Manifest variability remains unconfirmed until reproduced. Capture scope is not released to other owners.

Superseded decision: iteration 01's read-only-only capture hold and pending approval. Preserved: reviewed search behavior at 7a50a5e, existing review/test evidence for unchanged code, all failed full-gate evidence, release/human-only exclusions, external management merge ownership, and exact-head review/integration gates.

Current AL-60 checkpoint: 7b00cc1ac7fb5cd8afa2d93f7c65931c8ab2fa30, corroborated against remote feat/AL-60-search-tickets. It adds only artifacts to reviewed search implementation. No product PR yet.

1. Resume existing AL-60 owner in its next scope iteration; confirm the defect, implement and review the minimal repair, and pass the full integrated gate.
2. AL-61 keeps timestamp evidence and waits for the actual repaired AL-60 merge notification before one integration and gate. AL-76 continues writer-race correction/re-review. AL-62 continues its own review. No duplicate capture work.
3. Keep order AL-60, AL-61, AL-76, AL-62 unless concrete dependencies/readiness require change; serialize expensive full verification around upstream integrations.
4. Before each merge refresh PR inventory, expected head/base, reviews, discussion and actual checks. Record the real merge SHA before notifying later owners.
5. Maintain direct owner handoffs and management PR watches; retain manager-only telemetry at checkpoint finalization.

Verification impact: the old search reviews remain valid for unchanged search code; they do not approve new capture code. Capture repair requires new relevant review and regression evidence. Existing baseline make failure is retained and not re-run for manager-only documentation changes. No product merge has occurred.

Status correction from AL-62 owner read: the earlier suspected growing-snapshot data loss was a fixture error, refuted because the source did not prove a newer snapshot. The preservation guard remains intact. Reviews now examine skill attribution and persisted timeline evidence as open hypotheses, not established defects.

AL-76 local checkpoint: 02dcb11c2a757074a9cd0826692b11ecc3fa4817 (documentation successor); product reviewed at c32c8937d8897179a6772c6aa75300eafe94351f, final tests at 0dedb33110bafe2fc290ec99b86d3635b07c81b5. Owner reports all eight focus/model reviews and affected re-reviews approved. Severity 8 writer race corrected through canonical ownership in migration/backfill/computed export; two missing test assertions corrected. Focused race checks, binary build and disposable native CLI probes pass. Full make has not started under the deliberate deferral. Notification to integrate remains contingent on verified AL-60 then AL-61 merges. Local approval does not establish merge readiness. No new scope or superseded verification; no extra test run for management-only records.
