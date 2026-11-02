# chore(roles): preserve source session personas (GEN-17)

Preserve source-provided personas for root and child sessions instead of synthesizing RootAgent or UnassignedSubAgent. Resolve missing session metadata from unambiguous assistant persona evidence, including Codex root agent paths. Keep ancestry in the existing parent relationships.

Display missing personas as Unknown while retaining them in role metrics, summaries and trace filters. Cache live fallback evidence with source revisions and keep historical ledgers readable when source enrichment is unavailable.

Scope and acceptance: [CONTRACT.md](CONTRACT.md).
Approved scope extension: [iteration 02](iterations/02/PLAN.md).
Work status and evidence: [SUMMARY.md](SUMMARY.md).
