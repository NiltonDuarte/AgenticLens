h3. Title: AL-80: Deliver global OpenCode capture using the shared worktree runtime

*TARGET*: AgenticLens (Domain Integration / OpenCode Source Adapter / TypeScript Plugin / Native CLI / Worker / Setup Distribution / Tests)

*Status*: Blocked
*Work type*: feat
*Priority*: P1
*Blocked by*: AL-79
*Blocks*: None
*Origin*: PO-20, reclassified from GEN-22 after the user closed the capture design review.
*Specification*: [PO-20 SPEC.md](../../PO-20/SPEC.md), F1-F6; [confirmed discussion](../../PO-20/CAPTURE_VISION_AND_GAPS.md).
*Related*: AL-63 source-family audit and AL-78 native feedback remain independent. Do not extrapolate historical source-schema support into untested plugin/live support.

*CONTEXT*

The existing OpenCode plugin embeds installing-project/store/state paths. A machine-global plugin must serve actual session worktrees without binding them to the server's cwd or requiring a copied project plugin/manifest. AL-79 supplies the native configuration, routing, binary, binding and publication behavior; this slice adds the OpenCode integration rather than duplicating that logic in TypeScript.

*GOAL*

Install the OpenCode integration once in the effective user plugin directory and demonstrate a native root/child session in configured base/worktrees reaching the shared collector and publishing ticket metrics with the exact PO-20 policy.

*VERTICAL SLICE SCOPE*

* Domain Core: connect native OpenCode identities, explicit binding and proven-parent relationships to the shared policies from AL-79; preserve pure core decisions and owner/snapshot semantics.
* Boundary Adapters: global OpenCode loading, per-session context/source extraction, SDK/plugin binding surface, selected source families, bounded native handoff and worker integration.
* Presentation: thin strictly typed TypeScript plugin gives the agent exact-session binding requests/tools; CLI/setup/status provides actionable installation/source/configuration and publication outcomes. No new dashboard feature is required.
* Quality: actual global plugin discovery and native root/child evidence, plus isolated TypeScript/process/Go adapter regression coverage and cross-harness portable history checks.

*ACCEPTANCE CRITERIA*

h4. One global plugin, actual session scope

* Install additively into the effective OpenCode user plugin/config directory. Preserve unrelated plugins and source configuration, avoid duplicate owned project/global loading, and migrate/remove only unchanged owned legacy content through an explicit operation.
* Resolve the actual event/session worktree through authoritative runtime context, not a fixed projectDir, agent-provided identity or the long-lived server process cwd. No hardcoded source/state/project paths, Python dependency or per-worktree manifest requirement.
* Reuse AL-79's native whole-file configuration/base fallback, file-relative local paths, required local enablement, first-event state creation and read-only inspection. An absent shared file skips; missing/invalid required local settings errors; disabled capture avoids capture-only I/O.
* A known configured/runtime source disagreement errors without changing settings or reading another store. Source-family capability/parse failures remain visible and preserve facts; do not silently switch source harness/family.

h4. Binding and background publication

* For supported native session events, retain exact runtime session identity and validate actual parentage. Known current-task tickets bind without another user confirmation; missing context asks and stays unassigned. Source-proven same-task children inherit only when not explicitly bound.
* Collect unassigned facts locally without ticket publication, then include them after binding. Notifications request the common worker's collection rather than collecting/replaying event payloads in the plugin or invoking another model.
* Demonstrate periodic and turn/session checkpoint publication, coalescing, bounded retry and unresolved retention using AL-79's implementation. Keep worktree output/state separate and preserve concurrent contributors, other harnesses and known history.
* Shared version and optional development output follow the same selection/fallback/error and stable-running-worker rules as Codex and manual CLI. No plugin-side downloader, local version override or silent fallback policy.

h4. Native parity and regressions

* Verify a real globally loaded plugin, root binding response/tool, source-proven child and ticket ledger/summary publication on the actual supported OpenCode runtime/interface. Record native version, schema family, source IDs and limits. Fixture events and screenshots of installed files are not sufficient.
* Exercise an ordinary checkout and two worktrees with a different server/callback cwd, custom effective config home, local-file override versus base fallback, relative source path origins, disabled capture and first event. No native session or parent identity may be manufactured through tool arguments.
* Identify supported runtime/plugin interfaces explicitly and verify their source adapters independently; existing v1/v2 extraction support is not proof that both plugin APIs work. Unavailable evidence remains incomplete, not a silent skipped pass or production activation gate.
* Ensure the Codex path from AL-79 and manual CLI still operate with the same shared configuration/metrics semantics. Copy a valid mixed-harness ledger without original private capture/source state and preserve supported recorded facts while refreshing only complete available owners.
* Fixtures use isolated user config/cache/state destinations, zero domain mocks, strict TypeScript and meaningful existing test recipients. No test depends on live PO/ticket workspace contents.

*DEFINITION OF DONE*

* Start in a separate agentic-lens-dev session after AL-79 and an approved developer contract. Do not recreate common policy logic or absorb unrelated AL-63/AL-78 scope.
* Complete global OpenCode root/child-to-publication behavior and preserve Codex parity; update setup/operator documentation and packaged copies at their canonical sources. Describe tested interfaces and remaining limits accurately.
* Pass relevant Go/TypeScript/native integration checks, revision-specific focused independent review and required make. Apply frontend/embedded build rules if affected; keep supported UI/API consumers functional.
* Publish implementation PR and source-grounded session telemetry. Development verification stays in tests/evidence, not a production certification step. No model loop, automatic runtime download, implicit worktree recreation or automatic state deletion is introduced.
