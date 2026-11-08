# AL2-8: Show configured reasoning effort across views

## Goal

Operators can distinguish configured reasoning effort for model runs in historical and live views without inferring settings from token counts.

## Context

Historical OpenCode observations reported variant values in session-model and message JSON. These locations and sample counts are not current schema guarantees. Reasoning token metrics already exist; they do not establish the configured reasoning setting.

## Acceptance criteria

* Validate setting locations separately for supported OpenCode v1/v2 families, including top-level/nested examples where applicable. Record version/field evidence and honestly identify unsupported, absent or unavailable settings.
* Preserve native values, including values beyond low/medium/high, with exact invocation/session ownership. Missing/default stays unspecified; never infer configured effort from reasoning-token totals or another owner’s setting.
* Persistence and workspace/fleet projections retain effort facts and expose them without changing existing model, token or cost totals. Unavailable optional hydration and cross-machine reads preserve supported recorded facts.
* Deep Dive and Fleet model breakdowns distinguish reasoning effort; Model Profile/session timelines and Live Trace message headers show known settings. Missing/default values avoid empty badges and remain understandable. Mixed settings cannot be presented as one unqualified dominant setting.
* Fleet users can group or filter models by effort while retaining unknown values and recoverable empty states. Aggregations distinguish model/effort cohorts deterministically without duplicating usage.

## Implementation notes

Do not prescribe retired fact types or a new broad source abstraction. Resolve native session defaults versus message overrides using actual source evidence in the developer contract. Existing live capability limits still apply; this ticket does not introduce a new harness live service.

Follow [engineering rules](../../../TECHNICAL_GUIDELINES.md) and [regression boundaries](../../../docs/REGRESSION_PREVENTION.md). [PO-19](../../PO-19/SPEC.md) preserves the earlier development reconciliation; historical observations require revalidation, not an assumption of current runtime support.
