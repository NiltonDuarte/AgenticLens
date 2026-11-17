# chore(GEN-27): verify capture and the OpenSpec development workflow

## Pull request

This audit exercised pinned OpenSpec 2.0.0 and native Codex 0.169.0 and OpenCode 1.19.5/2.2.0 against fetched main `b6ab0471`. It records the supported CLI sequence, native event-to-ledger evidence, and what happens when source data or a former change path disappears. The audit proposes guidance corrections. No application or runtime change was made.

- **First publication is the operational gap.** An installed, trusted Codex hook recorded a real root and child. Both were bound to the correct ticket, yet the worker could not publish into the newly created OpenSpec change until an explicit generation supplied its first ledger path. The setup manual describes automatic publication without explaining this condition. The setup/operator and workflow owners should document capture pending before a first ledger and include the required human-readable title in first-generation examples. Repository policy reserves final telemetry for delivery. If first-ticket live publication is a product requirement, authorize a separate design for selecting the destination before the final extraction.
- **The CLI can carry the development lifecycle.** Use the creation helper, then pinned status and per-artifact instructions. Read apply instructions and obtain explicit scope approval before coding. Run strict validation alongside application verification and consolidated review. Read archive instructions, invoke the pinned archive command and use the returned location for subsequent operations. The generated propose and archive skills route around repository-specific creation checks and the supported archive operation. The workflow and setup owners should reconcile those entry points.
- **Published history followed the ticket.** Native Codex resume and the v1 OpenCode plugin published to moved homes. Unavailable owners survived, including exact parentage. OpenCode v2 manual extraction succeeded. Global plugin and cross-worktree implementation remain AL-80. A worker-control test twice missed its terminal shutdown deadline after its publication assertions passed. Its cause needs separate investigation by the capture-worker owner.

## Evidence

- [Observed CLI stages, prior delivery evidence and guidance findings](evidence/iterations/01/WORKFLOW.md)
- [Native event, binding, worker and publication results](evidence/iterations/01/NATIVE_CAPTURE.md)
- [Moved-ledger and unavailable-source controls](evidence/iterations/01/PUBLICATION.md)

The native investigators used loopback providers with synthetic usage. They exercised actual client runtimes and source stores. The publication investigator separately used synthetic rollouts with a real writer. These bounded checks do not establish provider pricing, global OpenCode callbacks or every runtime/profile combination. The coordinator consolidated the three investigations as audit evidence. There is no application-change review verdict.
