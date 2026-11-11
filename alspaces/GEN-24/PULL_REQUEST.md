# chore(GEN-24): align presentation with telemetry and add PR previews

The presentation previously led with anomaly investigation and used a separate cream/teal palette. It now matches the platform’s navy, blue and cyan styling and leads with skills, agents/roles, models and context telemetry. Project evolution, ticket comparisons, developer steering and Git-shared team history explain how teams can assess changes to guidance and development practices. Anomalies remain a supporting capability.

Same-repository PRs targeting main receive `/AgenticLens/preview/<PR-number>/` previews while the production presentation and roadmap continue to build from main. Isolated read-only candidate builds feed a serialized static publisher; durable state retains other open previews and an interrupted-publication guard prevents losing them after a failed transaction. This first PR uses a manual bootstrap; automatic lifecycle handling starts after the workflow is merged.

Validation: native Jekyll and composition/Git fixtures, desktop/tablet/mobile rendering, full repository `make`, and independent reviews. Hosted preview verification is pending. Current actionlint rejects GitHub’s newly documented `queue` key; the remaining workflow validates. Hosted execution must verify acceptance before delivery.

Scope, evidence and remaining limits: [contract](https://github.com/NiltonDuarte/AgenticLens/blob/GEN-24/telemetry-portal-preview-colors/alspaces/GEN-24/CONTRACT.md), [summary](https://github.com/NiltonDuarte/AgenticLens/blob/GEN-24/telemetry-portal-preview-colors/alspaces/GEN-24/SUMMARY.md), [verification](https://github.com/NiltonDuarte/AgenticLens/blob/GEN-24/telemetry-portal-preview-colors/alspaces/GEN-24/iterations/01/TEST_REPORT.md).
