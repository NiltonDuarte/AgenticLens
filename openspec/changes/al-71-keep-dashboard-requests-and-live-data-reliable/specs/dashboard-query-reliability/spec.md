## Purpose

Keep dashboard results aligned with the active selection and make historical Progression windows reproducible. See the [AL-71 proposal](../../proposal.md).

## ADDED Requirements

### Requirement: The current dashboard request owns its result

Fleet, Comparisons and Progression SHALL apply data, related diagnostics, request-derived filter choices, error and loading transitions only for the current request. A retry or refresh SHALL supersede an older request for the same selection. Fleet summary and workspace rows SHALL come from the same accepted request.

#### Scenario: Superseded request settles first

- **WHEN** request A settles successfully or fails while its replacement B is pending
- **THEN** A leaves B's loading, data, diagnostics, filter choices and error state untouched

#### Scenario: Superseded request settles last

- **WHEN** B succeeds or fails and A subsequently succeeds or fails
- **THEN** B's outcome remains authoritative and A contributes no rows, options, diagnostics or errors

#### Scenario: The view is left while a request is pending

- **WHEN** a view is unmounted before its request settles
- **THEN** the response cannot change that view's state

### Requirement: Progression presets use the latest ledger in the selected results

Progression SHALL calculate its preset range in UTC from the latest usable ledger reference in the current type, prefix and tag selection. Presets SHALL use the same rule with and without filters and SHALL expose the effective date range when one can be calculated. Custom ranges SHALL keep their explicit boundaries and unbounded sides. Lifetime SHALL remain unbounded.

#### Scenario: An all-matching filter on an old history

- **WHEN** the server clock is well beyond the newest ledger and a type or prefix filter matches every ledger in the current results
- **THEN** toggling that filter leaves the preset range, chart buckets and totals unchanged

#### Scenario: The latest ledger leaves the selection

- **WHEN** a filter removes the ledger that anchored the preset and leaves older dated ledgers
- **THEN** the range shifts to the newest usable reference among the remaining ledgers and the chart reflects that range

#### Scenario: Individual invocation dates span a boundary

- **WHEN** a selected ledger contains invocations on both sides of the effective range
- **THEN** the chart counts only invocations within the range and places delivered code by recorded ticket completion

#### Scenario: No usable anchor

- **WHEN** the selection is empty or has no usable ledger reference for a preset
- **THEN** Progression keeps the filter controls usable and does not show a fabricated effective range
