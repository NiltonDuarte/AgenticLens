# GEN-26 iteration 03: Multiple roots and ticket-prefixed changes

Source: user asked whether multiple workspace roots and immediate-child scanning could cover active/archive directories, prevent recreation, preserve binding, and permit ticket-prefixed OpenSpec names. This expands the design assessment, without authorizing implementation. Preserve direct storage in OpenSpec, exact session ownership, portable history and the no-code boundary.

Assess a generic multiple-root locator shared by discovery, detail, artifacts and publication; separate stable ticket identity from mutable folder paths. Explain that configured roots alone do not prevent recreation, define missing/ambiguous-target behavior, and establish OpenSpec naming capabilities from the pinned source. Keep configuration syntax and collision/write policies as recommendations pending a development contract; do not silently change global scan behavior.

Supersedes iteration 02's use of a full change slug as the only proposed identity: that remains a currently supported primitive, while the ticket-prefixed design can preserve an external canonical ticket ID separately from the folder name. Existing make evidence remains valid: all changes are ticket documentation/telemetry. Verify document links and diff, refresh final Codex telemetry, check existing PR state, then publish the assessment update.
