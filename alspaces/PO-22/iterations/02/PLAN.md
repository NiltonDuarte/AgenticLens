# PO-22 iteration 02: Remove duplicated ticket metadata

## Steering delta

Source: the user asks to remove Status, Work type, Blocked by, Blocks, Origin and Priority lines, use TICKETS.md for status/dependencies, avoid default priorities and correct the Markdown formatting.

Changed acceptance: active tickets use Markdown headings and contain no standard metadata block. The roadmap exclusively owns status/dependency fields; priority is optional and index-only, never assigned by default. The canonical PO skill and ticket template must enforce this so the removed fields do not return. This supersedes iteration 01's preservation of status/prerequisite fields and the Jira-style title convention. Earlier outcome titles, filename slugs, requirements, evidence attribution, dependency edges and successful application verification remain valid.

## Evidence and boundaries

Clean baseline: `f0debcf3`; PR #129 remains open. The current PO template explicitly requested Status and Blocked by, and skill discovery step 4 required preserving status/dependency fields. Work type, Blocks, Origin and Priority were not required by that template; they were accumulated draft conventions. The earlier Jira-style template is present at `452a8b5a`; current guidance was introduced at `731de946`. This establishes where the conventions are written, not who invented each field.

Update only the 27 active ticket specs, canonical PO skill/template and PO-22 artifacts. Leave the roadmap graph and historical records unchanged. Keep substantive context and finding/evidence links; remove metadata labels rather than meaningful provenance. No new priorities or replacement work-type field. Application/tests/runtime remain outside scope.

## Verification

Compare every active ticket with the baseline using only the approved heading conversions, metadata removal and reference-label formatting; assert no other content changes. Check unique IDs, filename/title alignment, valid local links, unchanged registry and reciprocal acyclic dependencies. Scan active skills/personas/setup guidance for conflicting field requirements and Jira-style ticket instructions. Reuse the earlier full make on unchanged application/build/test inputs; document this as documentation verification, not a new runtime test or independent review. Integrate current main, refresh final Codex telemetry, commit and update PR #129.

Format verification passed (`python3 /tmp/po22_verify_format.py`): all 27 ticket bodies match precisely the allowed transformations from `f0debcf3`, retaining requirements and evidence. Registry is byte-identical, title/ID references and graph are valid, and local links resolve. `git diff --check` passed. Search of active skills/personas/setup docs found no remaining old field requirements or Jira-style ticket instruction. `origin/main` remains `731de946`; application/test/build inputs remain unchanged, so the earlier full make is reused.
