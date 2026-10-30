# Implementation Plan: AL-46 Iteration 02

## Goal

Address validation findings for the default-collapsed failed tools section without changing its approved behavior or scope.

## Changes

- Correct the keyboard test to send the Space key using the key string recognized by `@testing-library/user-event`, verifying native button activation after Enter.
- Expand the failed tools section in anomaly streak tests before asserting card detail content, so assertions exercise the preserved expanded presentation rather than hidden content.
- Keep implementation limited to the inspector component and its tests; no backend or schema changes.

## Verification

- Run focused `LiveTraceInspector` tests and available web typecheck/lint/test commands.
- Confirm the test assertions validate collapsed/expanded state and anomaly card content in the appropriate state.
