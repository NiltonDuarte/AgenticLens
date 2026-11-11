# AL-83 verification

The corrected AL-83 candidate passes the native site suite, production builds, route/resource checks and targeted browser checks. The tested scope is the uncommitted site diff against `59f7ddeeed4ad4b95283acb2d0fd445317536057`. Its source fingerprint is SHA-256 `c084d8e7465ddd25e540fd85d39cc9e0e06345f2e13f57313a5ec91d588b81df`, computed from the builder, homepage, configuration, supporting-link plugin, site tests and five presentation assets in deterministic path order. The exact path list and digest input description remain in `/tmp/al83-browser/final-tested-source.json`.

## Native and route checks

The fixture suite passed **8 tests and 335 assertions, with zero failures, errors or skips**. New coverage replaces obsolete welcome-copy assertions and checks local presentation assets and fragments at both project and host roots, semantic landmarks, heading order, natural keyboard order, captioned SVG alternatives, self-contained illustrations, and hostile ticket permalinks targeting presentation assets. Existing source-isolation, workspace fallback, roadmap links, repeated build and failed-replacement cases remain intact.

Ruby is provided by the existing pinned Docker image. Its missing lockfile dependencies were installed into external `/tmp/al83-jekyll-gems` using `bundle install`; no repository dependencies changed. Final execution used:

```bash
docker run --rm -v "$PWD:/repo" -v /tmp/al83-jekyll-gems:/gems -w /repo \
  -e BUNDLE_GEMFILE=/repo/.github/pages/Gemfile -e BUNDLE_PATH=/gems \
  --entrypoint sh ghcr.io/actions/jekyll-build-pages:v1.0.13 \
  -c 'bundle exec ruby .github/pages/tests/site_test.rb && bundle exec ruby .github/pages/build.rb --repository NiltonDuarte/AgenticLens --revision 59f7ddeeed4ad4b95283acb2d0fd445317536057 --output dist/al83-project --base-path /AgenticLens && bundle exec ruby .github/pages/build.rb --repository NiltonDuarte/AgenticLens --revision 59f7ddeeed4ad4b95283acb2d0fd445317536057 --output dist/al83-root'
```

Both production builds passed. An independent HTML parser scanned each complete artifact: **103 pages and 244 local page/resource/fragment links per build**, with zero missing targets or project-path errors. Raw native output is `/tmp/al83-final-native.log`; route-check scripts/results are under `/tmp/al83-browser`.

## Browser checks and corrected finding

Headless Chromium through Playwright loaded the generated artifact using intercepted local responses, without a server or external page services. At **320, 375, 768 and 1440px**, final checks passed document width without horizontal overflow, SVG loading, visible skip-link focus and transfer to the main landmark, five deep-link destinations, and absence of failed resources. Axe WCAG 2/2.1 A/AA checks reported **zero violations at every width**, including contrast checks. A separate JavaScript-disabled context retained content and visuals. Hero viewport captures at 320 and 1440px were also inspected. This is targeted browser evidence, not a complete manual assistive-technology audit.

**TEST-01 — confirmed and fixed:** The tester independently discovered a WCAG label-content-name-mismatch on the header brand's explicit `aria-label="AgenticLens home"`. Its visible decorative `al` and brand text conflicted with that explicit label. Initial probes reproduced the same issue at all four widths, with no other axe violations. The coder removed the redundant label and retained the decorative monogram's `aria-hidden`; the complete final browser pass confirms the finding is resolved. This discovery belongs to the tester, separately from later review confirmation. No product changes were made by the tester.

Browser tooling, probe scripts, initial/final JSON results and captures remain external under `/tmp/al83-browser`. The final executable probe was `node /tmp/al83-browser/check.cjs "$PWD/dist/al83-project"`; output is `final-results.json` and `final-errors.log` is empty. Browser packages and raw captures are not repository deliverables.

The integrated `make` gate and independent reviews belong to the parent. Local checks do not verify AL-83 publication; public URL verification remains incomplete until an authorized merge and successful Pages deployment.
