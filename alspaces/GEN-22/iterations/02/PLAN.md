# GEN-22 iteration 02

Source: the user reports that PO tickets obscure their goal and context, permits
separate human/agent sections if needed, and prefers fewer artifacts.

Acceptance delta: PO specs and tickets lead with goal, context and observable
behavior. Necessary technical notes come last in the same document. Remove
mandatory layer inventories and repeated delivery boilerplate; avoid parallel
human/agent specifications. This replaces the remaining layer-based PO guidance
and extends the writing standard to the development-ticket template.

1. Align the PO skill and both templates with the canonical writing standard.
2. Check them against recent PO-20, AL-79 and AL-72 examples for goal visibility,
   duplication and preservation of meaningful constraints.
3. Validate links, metadata and instruction consistency; integrate and verify,
   refresh final telemetry and update the open GEN-22 PR.

Baseline: `305f073c`; initial working tree clean. The documentation coordinator
owns all changes. Prior artifact/evidence protections remain; historical tickets,
application code and generated metrics formatting stay outside scope. Direct
documentation review remains appropriate; no application review or new tests.
