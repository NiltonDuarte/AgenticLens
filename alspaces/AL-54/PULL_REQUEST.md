# feat(skills): add fleet adoption analytics and comparison filters (AL-54)

Aggregate recorded skill invocations across tickets and show unique skills, the top three skills, and invocation-weighted success in Fleet Overview. Add a Skills Used comparison column, a multi-select skill filter, and skill names in scatter tooltips.

Missing skill facts remain safe and visible without a skill selection. Selected skills match any selected name and compose with the existing filters. No recorded invocations show an unavailable success rate.

Contract: [CONTRACT.md](CONTRACT.md). Implementation and verification record: [SUMMARY.md](SUMMARY.md).
