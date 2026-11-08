# AL-80: Capture OpenCode sessions across worktrees

## Goal

Install the OpenCode integration once in the effective user plugin directory and demonstrate a native root/child session in configured base/worktrees reaching the shared collector and publishing ticket metrics with the exact PO-20 policy.

## Context

The existing OpenCode plugin embeds installing-project/store/state paths. A machine-global plugin must serve actual session worktrees without binding them to the server's cwd or requiring a copied project plugin/manifest. AL-79 supplies the native configuration, routing, binary, binding and publication behavior; this slice adds the OpenCode integration rather than duplicating that logic in TypeScript.

## Acceptance criteria

### One global plugin, actual session scope

* Install additively into the effective OpenCode user plugin/config directory. Preserve unrelated plugins and source configuration, avoid duplicate owned project/global loading, and migrate/remove only unchanged owned legacy content through an explicit operation.
* Resolve the actual event/session worktree through authoritative runtime context, not a fixed projectDir, agent-provided identity or the long-lived server process cwd. No hardcoded source/state/project paths, Python dependency or per-worktree manifest requirement.
* Reuse AL-79's native whole-file configuration/base fallback, file-relative local paths, required local enablement, first-event state creation and read-only inspection. An absent shared file skips; missing/invalid required local settings errors; disabled capture avoids capture-only I/O.
* A known configured/runtime source disagreement errors without changing settings or reading another store. Source-family capability/parse failures remain visible and preserve facts; do not silently switch source harness/family.

### Binding and background publication

* For supported native session events, retain exact runtime session identity and validate actual parentage. Known current-task tickets bind without another user confirmation; missing context asks and stays unassigned. Source-proven same-task children inherit only when not explicitly bound.
* Collect unassigned facts locally without ticket publication, then include them after binding. Notifications request the common worker's collection rather than collecting/replaying event payloads in the plugin or invoking another model.
* Demonstrate periodic and turn/session checkpoint publication, coalescing, bounded retry and unresolved retention using AL-79's implementation. Keep worktree output/state separate and preserve concurrent contributors, other harnesses and known history.
* Shared version and optional development output follow the same selection/fallback/error and stable-running-worker rules as Codex and manual CLI. No plugin-side downloader, local version override or silent fallback policy.

### Native parity and regressions

* Verify a real globally loaded plugin, root binding response/tool, source-proven child and ticket ledger/summary publication on the actual supported OpenCode runtime/interface. Record native version, schema family, source IDs and limits. Fixture events and screenshots of installed files are not sufficient.
* Exercise an ordinary checkout and two worktrees with a different server/callback cwd, custom effective config home, local-file override versus base fallback, relative source path origins, disabled capture and first event. No native session or parent identity may be manufactured through tool arguments.
* Identify supported runtime/plugin interfaces explicitly and verify their source adapters independently; existing v1/v2 extraction support is not proof that both plugin APIs work. Unavailable evidence remains incomplete, not a silent skipped pass or production activation gate.
* Ensure the Codex path from AL-79 and manual CLI still operate with the same shared configuration/metrics semantics. Copy a valid mixed-harness ledger without original private capture/source state and preserve supported recorded facts while refreshing only complete available owners.
* Fixtures use isolated user config/cache/state destinations, zero domain mocks, strict TypeScript and meaningful existing test recipients. No test depends on live PO/ticket workspace contents.

Shared specification: [PO-20 SPEC.md](../../PO-20/SPEC.md), F1-F6; [confirmed discussion](../../PO-20/CAPTURE_VISION_AND_GAPS.md).

## Implementation notes

Related work: AL-63 source-family audit and AL-78 native feedback remain independent. Do not extrapolate historical source-schema support into untested plugin/live support.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [relevant regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. Historical findings require revalidation; this rewrite is not runtime verification or implementation approval.
