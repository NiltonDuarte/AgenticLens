# Project presentation and ticket roadmap

Native Jekyll publishes the AgenticLens presentation at `https://niltonduarte.github.io/AgenticLens/` and links to the separately selected `alspaces/plan/TICKETS.md` roadmap at `/AgenticLens/roadmap/`, with ticket pages beneath that route. The English presentation explains the workflow, product boundaries, practical investigations and local architecture; setup links lead to maintained release-aware instructions. Edit the source Markdown normally; the **Ticket roadmap pages** workflow validates pull requests and publishes merged `main` changes. Manual deployment also requires selecting `main`.

Only plan Markdown and the trusted presentation scaffold and explicitly listed CSS/SVG assets enter Jekyll's source directory. Established optional-front-matter and relative-links plugins render the existing files without editing them. The GFM parser preserves heading fragments and nested ticket metadata. A small hook keeps ticket destinations beneath `roadmap/` and points supporting files outside the plan to their matching GitHub source commit. Ticket front matter cannot overwrite the presentation or its trusted assets. Assets are copied individually from `.github/pages/presentation/`; the builder verifies the exact output list and bytes. New assets require an explicit builder allowlist change. The visuals are labeled qualitative illustrations, not session screenshots or measured results. Workspace telemetry, private source sessions and repository screenshots are excluded. The public site uses no JavaScript, account, local engine, analytics service or remote font.

## Build locally

With Ruby 3.3 and Bundler installed, run from the repository root:

```bash
export BUNDLE_GEMFILE="$PWD/.github/pages/Gemfile"
bundle install
bundle exec ruby .github/pages/tests/site_test.rb
bundle exec ruby .github/pages/build.rb --roadmap alspaces/plan --repository NiltonDuarte/AgenticLens --revision "$(git rev-parse HEAD)" --base-path /AgenticLens
```

Build at both `/AgenticLens` and an empty base path to verify project and root hosting. CSS and SVG assets live beneath the matching `presentation/` route; all navigation uses normal links and stable section fragments.

The same commands can run in the Ruby environment provided by `ghcr.io/actions/jekyll-build-pages:v1.0.13`. Use the native Jekyll CLI through this entrypoint, rather than the image's default GitHub Pages wrapper: the wrapper enables safe mode, which disables the supporting-link hook.

The staging entrypoint selects the roadmap from `roadmap_path` in `.github/pages/_config.yml`; this repository explicitly chooses `alspaces/plan`. `--roadmap PATH` overrides that choice. Both paths resolve relative to `--project` and must remain inside the repository. A missing selection fails explicitly. Roadmap selection is independent of `workspaces.roots`; the builder never infers it from ledger search roots. It then invokes native `jekyll build`. The artifact is `dist/tickets-site/`; `index.html` is the presentation and `roadmap/index.html` is the ticket roadmap. Inspect the HTML or use a suitable static host, since direct file opening does not resolve absolute site links. Omit `--base-path` when serving at a web host's root. Generated output is ignored by Git and replaced on subsequent successful builds.

## Publishing

Repository **Settings > Pages** must select **GitHub Actions** as the build source. The deployment job alone receives Pages write and identity permissions. Environment protection rules may require approving the `github-pages` deployment. A successful local build or PR check does not establish that the site is live; the new route becomes available after the merged workflow successfully deploys.

After deployment, open the public root, a presentation asset, each section anchor and `/AgenticLens/roadmap/` (including a ticket link). Verify the changed content, not only an HTTP success response or an older deployment. Publication remains incomplete until the changed main revision deploys and those public checks pass.
