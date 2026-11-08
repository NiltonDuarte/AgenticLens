# Ticket roadmap website

Native Jekyll publishes the configured workspace's `plan/TICKETS.md` at `https://niltonduarte.github.io/AgenticLens/roadmap/`, with ticket pages beneath that route. The site does not create a homepage at `/AgenticLens/`. Edit the source Markdown normally; the **Ticket roadmap pages** workflow validates pull requests and publishes merged `main` changes. Manual deployment also requires selecting `main`.

Only plan Markdown and the shared presentation scaffold enter Jekyll's source directory. Established optional-front-matter and relative-links plugins render the existing files without editing them. The GFM parser preserves heading fragments and nested ticket metadata. A small post-render hook points supporting files outside the plan to their matching GitHub source commit. Workspace telemetry and source sessions are excluded.

## Build locally

With Ruby 3.3 and Bundler installed, run from the repository root:

```bash
export BUNDLE_GEMFILE="$PWD/.github/pages/Gemfile"
bundle install
bundle exec ruby .github/pages/tests/site_test.rb
bundle exec ruby .github/pages/build.rb --repository NiltonDuarte/AgenticLens --revision "$(git rev-parse HEAD)" --base-path /AgenticLens
```

The same commands can run in the Ruby environment provided by `ghcr.io/actions/jekyll-build-pages:v1.0.13`. Use the native Jekyll CLI through this entrypoint, rather than the image's default GitHub Pages wrapper: the wrapper enables safe mode, which disables the supporting-link hook.

The staging entrypoint reads `workspaces.dir` from `.agenticlens-config.yaml`, falling back to `workspaces`, then invokes native `jekyll build`. The artifact is `dist/tickets-site/`; its homepage is `roadmap/index.html`. Inspect the HTML or use a suitable static host, since direct file opening does not resolve absolute site links. Omit `--base-path` when serving at a web host's root. Generated output is ignored by Git and replaced on subsequent successful builds.

## Publishing

Repository **Settings > Pages** must select **GitHub Actions** as the build source. The deployment job alone receives Pages write and identity permissions. Environment protection rules may require approving the `github-pages` deployment. A successful local build or PR check does not establish that the site is live; the new route becomes available after the merged workflow successfully deploys.
