# chore(GEN-13): consolidate and shorten project documentation

Repeated workflow explanations and operational references made the project docs harder to scan. Shorten the README and setup prose, consolidate telemetry source/classification/recovery guidance in the setup manual, and link contributor/build rules to their canonical owners.

Preserve setup alternatives, commands, configuration semantics, compatibility limitations, technical requirements, screenshots, and historical records. No application or test changes.

Validation: 33 local Markdown links/anchors and whitespace checks passed. Integrated `make` passed (Go vet/race tests, TypeScript typecheck/ESLint, 712 frontend tests, asset embedding, and binary builds). Telemetry availability is recorded in SUMMARY.md.
