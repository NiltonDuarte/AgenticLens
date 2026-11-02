# GEN-17: Preserve root session roles

## Status

Implementation in progress under the explicitly approved iteration 02 delta. The user approved applying Unknown to missing roles for both roots and children; this supersedes the proposed child fallback in the frozen baseline contract and iteration 01.

## Allocation and baseline

- User requested the next generic ticket and preservation of root session roles, asking whether the RootAgent label should move to a new field or be removed.
- Configuration: workspace root alspaces; GEN tickets use conventional type chore.
- Checked existing workspaces, ticket index, local branches and origin branch references. GEN-1 through GEN-15 are allocated, including GEN-14 on origin. GEN-16 is available.
- Initial working tree clean at 99fbcd7be0d0ed05639e4a0cb3be8744a3d5371e. Starting branch GEN-13/root-session-role was a generated name, not an available ticket identifier.
- Allocated branch chore/GEN-17-preserve-root-session-role. No matching specification or ticket index entry; this is an ad hoc ticket.

## Proposed direction

Preserve source-provided persona roles independently of session ancestry. Derive root presentation from existing parent relationships where they establish root status; avoid redundant persisted root flags. The contract and iteration plan define missing-role and legacy-data handling for approval.

## Workflow

The planner authored CONTRACT.md and iterations/01 and 02 plans. A coder owns application implementation; a separate tester and independent focused reviewers follow. The parent owns integration, remaining artifacts, telemetry and publication. User approval source: "apply this recommendation" after the recommendation to use Unknown for missing personas in both roots and children, with ancestry separate. No further scope confirmation is required. Telemetry extraction is deferred until the final delivery step; no ledger has been generated yet.

Planning verification: parent reviewed both artifacts and ran git diff --check successfully. Local bin/agentic-lens is absent (generate --help could not run); build the canonical binary before delivery telemetry. No implementation, review approval, integration checks, commit or PR is claimed at this checkpoint.

## Ticket reassignment and integration checkpoint

User steering: "GEN-16 was taken. change the workspace to GEN-17 and merge main into the branch". This is an identifier correction for the same approved work, not a second task. All task-owned artifacts moved from alspaces/GEN-16 to alspaces/GEN-17; application and test files were not changed for relocation. Historical reviewer findings retain GEN16 identifiers and reviewed fingerprints as provenance. Active artifact paths and branch references now point to GEN-17. Original allocation evidence above describes the earlier local view; the user's correction supersedes that assignment. GEN-17 workspace/index/local and remote references were checked before relocation.

The latest origin/main contains a separate GEN-16 documentation ticket (8f0fd4f, PR95). Preserve those incoming artifacts. Reviews paused for integration; no consolidated approval or PR delivery is claimed. Implementation will be checkpoint-committed with open review evidence, then main merged, as explicitly requested. Telemetry remains deferred until final delivery.

Open correction: data/failure re-review (gpt-6.1-sol/high) independently reproduced a new partial metadata-read failure: source trace lookup can succeed while session.agent query fails; matched persisted event-only session personas then become blank through authoritative reconciliation. Recorded as D-SOL-R1, with failing overlay and passing generic/healthy-source controls. This still blocks delivery. Behavior, architecture and test Sol re-reviews approved the prior two corrections at fingerprint 7b720b47; Luna correction re-reviews have not run yet.
