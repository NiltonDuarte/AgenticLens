# GEN-24: Telemetry-led presentation portal

GEN-24 updates only the public presentation portal and the build/deployment support needed for PR previews. The user approved the clarified scope with “$agentic-lens-dev approved”. Implementation and local review are complete; integrated and hosted verification are in progress; the origin story is excluded. Application and dashboard behavior are outside this ticket.

The requested positioning leads with team telemetry: skills, agents and roles, models, and context-window saturation. Project evolution and ticket comparison explain how teams can assess changes to guidance, specifications and skills using recorded cost, tokens, cache use, code changes, duration and developer steering. Anomalies remain a supporting capability. The presentation will explain why Git-versioned ledgers provide shared history for teammates and engineering managers. The current scope incorporates the [planning correction](iterations/01/PLAN.md#scope-clarification).

## Baseline and discovery

Initial checkout: branch `GEN-24/telemetry-portal-preview-colors`, commit `167c204fc958da34eb82d97129dedd85ac9b8c58`; working tree clean. Configuration selects `alspaces` as the ticket workspace root. No existing GEN-24 workspace or matching plan specification was found.

The presentation is a static Jekyll site under `.github/pages`, deployed by `.github/workflows/tickets-pages.yml`. Platform palette values are defined in `web/src/index.css`; the presentation currently uses a separate cream/teal palette. GitHub API inspection confirms workflow-based Pages at https://niltonduarte.github.io/AgenticLens/ and a `github-pages` deployment environment restricted to `main`.

The proposed preview requires a composed site preserving the main portal at the root and publishing the candidate under `/preview/<PR-number>/`. The initial PR needs a manual workflow dispatch and a dedicated preview deployment environment; the production environment policy stays intact. GitHub's native deploy-pages preview option is currently unavailable publicly. Hosted deployment must be verified before reporting a preview as available.

## Verification and limits

A Ruby 3.3 Docker container with the pinned bundle provides native Jekyll verification because Ruby is absent on the host. A nested-prefix production build passed. Playwright rendered the portal at 1440, 768 and 375 pixels: every image loaded, no horizontal overflow or page errors occurred, and the first Tab reached the skip link. Screenshots were inspected at desktop and mobile widths; private probes/screenshots remain outside Git.

The approved `github-pages-preview` environment was created with a deployment policy permitting only `GEN-24/telemetry-portal-preview-colors`. The production `github-pages` environment remains restricted to `main`. No live preview or pull request exists yet. Focused evidence: Jekyll 8 runs / 453 assertions and Python 9 tests passed, including an actual local Git pending-state check. The actual candidate build shell also passed against baseline source lacking the new composer files (8 runs / 335 assertions and a nested-prefix build). The first full `make` passed. All four review focuses ran on both configured models, with source-observed profiles verified from Codex metadata; C2 rechecks approved the corrections. Integrated and hosted checks remain pending. Telemetry extraction is deferred to the final delivery step.

## Review outcome

C1 found three publication issues: default queue cancellation (behavior, data and architecture reviewers on `gpt-6.1-sol`/high), an older-PR test-file dependency (behavior on `gpt-6.1-sol`/high), and inconsistent retarget eligibility (data on `gpt-6-luna`/medium and `gpt-6.1-sol`/high, with initial dissent about blocking impact). C2 corrects all three. A bootstrap trust concern raised by data on `gpt-6-luna`/medium was refuted against the approved manual reviewed-branch trust boundary; its dissent is retained. Test reviewers on both models verified meaningful fixture assertions, including disposable mutations. Every affected focus was rechecked under both configured profiles and approved C2. Detailed working reports remain private until final consolidation; compact finding attribution will be retained before cleanup.

Latest actionlint 1.7.12 rejects only GitHub’s newly documented `queue` key; a disposable copy omitting solely that key passes. Exact hosted workflow acceptance remains a required delivery check.
