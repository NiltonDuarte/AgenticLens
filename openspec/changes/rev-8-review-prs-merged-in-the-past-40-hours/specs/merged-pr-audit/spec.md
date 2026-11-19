## ADDED Requirements

### Requirement: Account for merged pull requests in the selected window

The audit SHALL include every pull request whose GitHub merge timestamp falls within the selected 40-hour UTC window.

#### Scenario: Pull requests merged into the reviewed repository
- **WHEN** GitHub reports pull requests merged in the selected interval
- **THEN** the audit lists each pull request and revision, checks its changes against the merged requirements, and reports actionable findings with evidence and remaining verification limits.
