# GEN-26: OpenSpec compatibility assessment

Assess whether the current AgenticLens implementation can accompany OpenSpec and identify the gaps to deeper integration. Deliver an evidence-linked assessment, prioritized gap map and proposed acceptance checks; do not implement application code, tests, schemas, runtime scripts or install OpenSpec into this project.

The user approved this assessment boundary in the session request. Recommendations do not authorize implementation. Compare supported harness telemetry, ticket identity, workspace discovery, artifacts, lifecycle, configuration, persistence and dashboard representation. Separate source-proven behavior, existing test evidence and unverified interoperability. Preserve portable ledgers and truthful missing-source reporting.

Use the configured GEN- prefix and alspaces root. GEN-26 is free in local workspaces, the ticket index and fetched branch refs; GEN-13 already exists and GEN-25 has an active branch. Baseline: 460c511ef96f9e42f2ecbfdba451828b52d3ce91; initial Git status clean. Own only alspaces/GEN-26/**. Delivery is documentation and session telemetry through a PR; no application changes or implementation tickets are created.

Current acceptance boundary: direct integration must use OpenSpec change folders as the configured workspace, with no separate AgenticLens workspace. Assess `openspec/changes` for active and archived changes, standard artifacts and publication. This user clarification supersedes the separate-workspace pilot recommendation; see iterations/02/PLAN.md. No code implementation is authorized.

Iteration 03 additionally assesses explicit active/archive roots, shallow discovery, a shared locator, publication without missing-folder recreation, root-independent stable binding, and ticket-prefixed OpenSpec names. These are architecture recommendations, not frozen application API/configuration contracts or implementation approval; see iterations/03/PLAN.md.
