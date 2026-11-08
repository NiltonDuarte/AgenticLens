# GEN-22 iteration 02: Product ticket for metric tags

Source: latest user request invoking agentic-lens-po on 2026-11-08.

Changed acceptance: additive key-value metric tags, shared and local config plus CLI; CLI > local > shared; dashboard filters default off; comparison identification by tag value; untagged points remain valid and visible.

Superseded decision: iteration 01's unapproved user_tag attribution override and generate-only UI scope are obsolete. CONTRACT.md remains the historical unapproved baseline and grants no implementation authority. SPEC.md and planned AL-79 govern this product scope.

Preserved guarantees: existing user attribution, portable session history, unavailable-owner preservation, typed FCIS architecture, no application changes in this planning session. No prior implementation approval or verification exists to carry forward.

Workflow: inspect PO templates and actual configuration/metric/comparison boundaries; author SPEC.md and one complete development ticket; register AL-79 as Ready; validate planning links and diff; finalize artifacts and actual session telemetry; publish documentation PR. No development agents or code changes.

Ticket ownership: retain active GEN-22 rather than switching to a new PO ticket, honoring repository single-ticket exclusivity. The PO skill's usual PO-# workspace convention is adapted to this existing session; GEN-22 owns only planning artifacts and registry edits, and AL-79 has no developer workspace. This is not development of a second ticket.

Verification impact: documentation/link consistency now; AL-79 requires config/CLI/persistence/API/frontend and portable-history tests during a separate development session. No product tests are claimed here.
