# FIX-17: Explicit capture-state recovery for manual generation

## Problem and evidence

The user reports that GEN-21 manual Codex generation fails before source extraction: its configured registry advertises capture metadata version 2 and contains worktrees, but lacks worktree_coordinates. The native root and eleven child sessions are present. The producing registry version is unknown; missing source data is not established. The supplied report is diagnostic evidence, not authorization to alter GEN-21 artifacts.

At baseline 86bea597afe7208cb1836f7cc5747083f7b95f47, backend/cmd/agentic-lens/main.go resolves the selected capture state and calls checkCaptureInstallation before source extraction and selected-state admission. In backend/cmd/agentic-lens/capture.go, checkCaptureInstallation previews the configured registry before previewing the selected override. Consequently an incompatible configured registry prevents recovery through a fresh explicit --capture-dir. backend/internal/adapters/capture/coordination.go deliberately rejects incompatible coordinate layouts; this ticket does not change that compatibility contract.

## Authorized scope and acceptance

The user's explicit preapproval authorizes manual Codex generate with an explicit healthy or fresh --capture-dir despite an incompatible configured registry. Preserve the configured registry unchanged and retain real installed-state isolation checks. Do not infer receipts, source provenance, or a new authenticity requirement for portable ledgers.

1. An explicit independent healthy/fresh state permits manual generation from exact available Codex sessions without requiring the unselected configured registry to pass coordinate-schema validation.
2. Without an independent explicit override, incompatible selected/configured state continues to fail clearly. Selecting the incompatible state explicitly does not make it acceptable.
3. Installed-state ownership, active manifest/pin/template/native/wrapper/source and owned-integration checks remain effective. An explicit state-path drift warning may retain its existing recovery semantics; it must not suppress unrelated installation failures.
4. The old registry is neither reset, migrated nor rewritten. Shared/local project configuration and installed integrations are unchanged.
5. Successfully extracted local sessions remain authoritative only for their exact owners; unavailable sessions and other-harness facts survive generation. No fabricated successful extraction, duplicate cumulative snapshot or new receipt requirement is allowed.
6. Native automatic capture paths retain their current strict health and compatibility failures. This is a manual explicit-state recovery fix, not automatic repair.

## Boundaries

Use Go CLI/adapters only as needed for state-selection and installation validation. Domain functions remain pure; external state stays in the imperative shell. No frontend, ledger schema, registry migration, pricing or metric-formula changes. Zero any/interface{} and zero emojis. Tests use isolated fixtures, never the live GEN-21 workspace or private registry. The parent owns integration, review consolidation, commits, telemetry and PR publication. GEN-21's owner performs its later integration and real root-plus-eleven-child extraction.

Approval source: parent relayed the user's explicit preapproval during iteration 01. Implementation may proceed within these boundaries without another scope gate.
