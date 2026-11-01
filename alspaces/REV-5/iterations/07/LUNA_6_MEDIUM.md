# Architecture review: Luna 6 medium

- Run: `ARCH-luna_6_medium-01` (discovery)
- Requested identity: `gpt-6-luna`, reasoning effort `medium` (provider instance `codex` per policy)
- Runtime identity: requested by coordinator; this report has no independent dispatch/source-session metadata to verify the actual model or effort. No usage or cost claim.
- Candidate: `0040edf70316ed3313669b173c2c51054d10aa1cc2e7c658b50d5bfbf1f4db79`
- Baseline: `f9bd054`
- Scope: iteration 07 plan and ticket contract; model policy and attribution docs; aligned reviewer/orchestrator documentation. Architecture focus limited to policy semantics, dispatch mappings, attribution, deduplication and gate behavior.

## Verdict: APPROVED (scoped)

No blocking architecture or correctness finding in the reviewed documentation. The policy separates selected persona from model profile, expands each applicable focus across the configured profiles, and specifies capacity waves without dropping cells. It makes requested, acknowledged and source-observed identity distinct; preserves failed, no-finding, refuted and duplicate-discovery evidence; and keeps verification credit separate from discovery. The gate requires complete persona/profile cells and lets a confirmed blocker from either profile block delivery.

## Hypotheses and evidence

- **Configuration could silently weaken the explicit two-model requirement.** `default_profiles` contains both named profiles. Overrides replace defaults, but the policy explicitly disallows applying an override that conflicts with current user instructions; a selection is required to remain nonempty and reference defined profiles. A coordinator applying this contract cannot silently drop one configured model under the current user requirement.
- **Capacity or model unavailability could be recorded as completion or substituted silently.** The policy marks queued/unavailable cells incomplete, requires explicit dispatch settings, and rejects silent inherited/default settings or substitution. The parallel protocol requires all persona/model cells for the approval gate.
- **Deduplication could erase independent discovery credit or distort verifier efficacy.** Canonical findings retain all discovering run IDs and per-report observations; duplicate confirmed claims are not false positives. Verifiers do not receive discovery credit merely for confirming supplied findings.
- **Identity/cost could be overstated from prompt text.** The contract separates requested settings from dispatch acknowledgment and source observation, leaves unavailable observations null with a reason, and permits cost only when source-backed. This run has no independent host metadata, so requested identity is recorded separately above.

## Checks and limits

- Read required architecture persona, shared review context, canonical review procedure, regression checklist, technical guidelines, iteration 07 plan, root contract, and scoped policy/attribution plus aligned documentation diff against `f9bd054`.
- `git diff --check` on the scoped changed paths passed.
- Static documentation/configuration review only. No runtime scheduler or application code is in scope; YAML parser behavior and actual cross-provider dispatch/source-session metadata were not exercised. The plan's `CONTRACT.md` path does not exist; the approved iteration 07 contract is recorded in `alspaces/REV-5/CONTRACT.md` and was reviewed.
