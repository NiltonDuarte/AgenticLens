# AL-86: Set up ticket-prefixed OpenSpec changes

## Goal

A developer can set up OpenSpec alongside AgenticLens and create a correctly identified change containing planning artifacts and telemetry in one directory.

## Context

OpenSpec already accepts `al-123-add-login`; no upstream fork is required. AgenticLens setup and operator guidance still assume a separate ticket directory. A deliberate setup/creation workflow must connect canonical `AL-123` identity to the lowercase change name and configure active/archive roots. See [PO-24](../../PO-24/SPEC.md) and its pinned assessment evidence.

## Acceptance criteria

- Extend the adopting-repository setup/operation workflow with an explicit OpenSpec option that uses verified, pinned supported tooling and ordinary OpenSpec change creation. Existing non-OpenSpec adoption remains supported. Refreshing integration preserves unrelated/custom skills, configuration and project files.
- Create a change from an existing explicit configured ticket ID plus description, producing a lowercase ticket-prefixed name. Preserve the canonical ticket spelling separately; invalid/unconfigured prefixes, invalid names, collisions and an existing change for that ticket produce clear guidance without overwriting it.
- Configure the supported active/archive roots and standard proposal/design/tasks artifact paths without a second AgenticLens ticket workspace. A ledger-free known change remains resolvable for first publication; do not fabricate telemetry from its Markdown.
- Demonstrate the configured Codex/OpenCode source intersection and source-backed telemetry in the created change, including exact ownership, explicit conventional type and unavailable-cost/source behavior. OpenSpec support for other assistants must not imply new AgenticLens adapters.
- Setup explains what is project-shared versus machine-local, preserves portable ledger reads on another checkout, and retains existing mandatory setup confirmations. It does not invent shared absolute source/store paths or silently enable unrelated integrations.
- Record the verified OpenSpec CLI/schema version and ensure update instructions preserve the naming/location contract. Use supported project context/rules or custom schemas; do not depend on an unverified upstream hook or unknown metadata field being enforced.

## Implementation notes

This ticket owns generic setup/operator integration, not the AgenticLens repository's persona/governance rewrite or mass backlog migration. Standard artifact browsing suffices; dynamic capability enumeration, phase analytics and separate planning stores are deferred. The developer chooses the smallest supported extension point rather than adding an upstream OpenSpec dependency to the engine runtime.
