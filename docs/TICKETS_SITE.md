# Ticket roadmap website

The GitHub Pages site renders the configured workspace's `plan/TICKETS.md` as its homepage and the plan's Markdown files as readable pages. Edit those source files normally; the dedicated **Ticket roadmap pages** workflow validates pull requests and publishes changes after they merge into `main`. Manual deployment also requires selecting `main`.

The site contains only rendered plan Markdown and its stylesheet. Supporting documents outside the plan link to the matching GitHub source revision. It does not publish workspace telemetry or change the AgenticLens dashboard.

## Build locally

From the repository root, create an isolated Python environment and build the same artifact used by Actions:

```bash
python3 -m venv .venv
.venv/bin/python -m pip install -r scripts/tickets-site-requirements.txt
.venv/bin/python -m unittest discover -s scripts/tests -p test_tickets_site.py
.venv/bin/python scripts/build-tickets-site.py --repository NiltonDuarte/AgenticLens --revision "$(git rev-parse HEAD)" --base-path /AgenticLens
```

The command supplies the current commit SHA for immutable supporting-document links. The output is `dist/tickets-site/`, which is ignored by Git. Inspect the generated HTML or serve it with a suitable static host; direct file opening does not resolve its absolute site links. Omit `--base-path` when serving the artifact at a web host's root; the GitHub project URL uses `/AgenticLens`. The builder reads `workspaces.dir` from `.agenticlens-config.yaml`, with `workspaces` as the fallback, and never edits the source Markdown. Repeated builds replace only a directory marked as this builder's generated output.

## Enable publishing

In repository **Settings > Pages**, select **GitHub Actions** as the build source. After the workflow is merged and successfully deploys, the default URL is `https://NiltonDuarte.github.io/AgenticLens/`. Repository environment protection rules may require approving the `github-pages` deployment. A successful local build or PR check does not establish that the site has been deployed.
