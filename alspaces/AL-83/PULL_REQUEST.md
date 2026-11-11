# feat(AL-83): present AgenticLens on GitHub Pages

Replace the minimal welcome page with an English presentation for developers and engineering leads. Explain local telemetry, investigation workflows, three use cases, operating modes, evidence limits and architecture, with labeled illustrative product visuals and release-aware setup links. Preserve the roadmap and single native Jekyll deployment path.

The builder stages only named presentation assets and validates their output bytes. The static page requires no account or local engine and publishes no private session captures. See the [contract](https://github.com/NiltonDuarte/AgenticLens/blob/AL-83/present-agenticlens-on-github-pages/alspaces/AL-83/CONTRACT.md) and [summary](https://github.com/NiltonDuarte/AgenticLens/blob/AL-83/present-agenticlens-on-github-pages/alspaces/AL-83/SUMMARY.md).

Validation: integrated `make` passed; 8 Jekyll tests/335 assertions; project/root builds with 244 valid internal links each; keyboard/responsive checks at four widths and zero axe violations. Four review focuses ran under Sol/high and Luna/medium. Tester found and verified a corrected accessible-name issue; reviewers retained pre-existing output-replacement debt and a nonblocking copy-test suggestion.

Publication remains incomplete until merge and successful main Pages deployment, followed by public verification. The existing local output-replacement filesystem-failure weakness is unchanged; builder maintainers own follow-up. This PR does not authorize merging.
