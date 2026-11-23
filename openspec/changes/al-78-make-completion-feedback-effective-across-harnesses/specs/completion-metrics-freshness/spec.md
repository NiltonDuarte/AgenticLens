## Purpose

Make source-backed ticket metrics queryable soon after an agent commits completed work, with a clear distinction between available provisional facts and facts settled by native completion. [Roadmap ticket](../../../../../alspaces/plan/to-do/AL-78%20make%20completion%20feedback%20effective%20across%20harnesses.md)

## ADDED Requirements

### Requirement: Publish available facts and settle them from native completion

AgenticLens SHALL publish available facts for an exact owner and refresh the owner's snapshot when native completion and its associated source facts become readable.

#### Scenario: Code is committed while the native interval is open

- **WHEN** a Codex or OpenCode agent commits code during an open accounting interval and a supported capture or manual generation pass reads that interval
- **THEN** the resulting snapshot records available duration and usage, with any estimated duration marked provisional. A Git commit, a tool return, or a later model request alone does not establish native completion.

#### Scenario: The same owner completes naturally with automatic capture enabled

- **WHEN** an installed, supported lifecycle integration observes natural completion and the source exposes the facts for the exact owner and interval
- **THEN** the existing automatic capture path publishes the settled snapshot to the selected destination without a metrics-only model turn or a new user prompt. The first eligible publication replaces the provisional contribution for that interval once, includes the usage and turns available at that read, preserves other open intervals' provisional contributions, and retains other harnesses, unavailable owners, and unknown costs.

#### Scenario: Completion, usage, or publication is delayed

- **WHEN** a completion notification arrives before native evidence is readable, or the source or publication pass is incomplete
- **THEN** the workflow retains accepted facts, identifies the pending or failed boundary through existing diagnostics, and uses the supported bounded refresh path. The workflow does not report the interval or unavailable usage as settled solely from the notification.

### Requirement: Give actionable guidance at a manual checkpoint

When an enabled checkpoint qualifies, manual `generate` SHALL describe what is provisional and how to obtain a settled refresh. Current-caller advice SHALL require native evidence for the exact owner and accounting interval.

#### Scenario: An exact native interval reaches the default checkpoint threshold

- **WHEN** manual `generate` publishes a snapshot whose source evidence identifies an open native interval that meets the default 180-second threshold
- **THEN** the checkpoint identifies the source owner and interval and explains that capture or external generation can refresh it after natural completion and source availability. If the exact caller and interval are independently verified, the checkpoint may advise finalizing naturally once work is complete. An unverified or different caller receives qualified guidance that does not ask it to end the owner's turn. Summed short intervals alone do not trigger advice about a long interval.

#### Scenario: The installed workflow has an unsupported boundary

- **WHEN** the selected installed harness lacks a usable automatic capture or in-turn delivery path
- **THEN** the workflow identifies the unsupported boundary. Where automatic refresh is unavailable, it documents explicit source- and session-selected `generate` from outside the completed turn after the source is readable. Manual refresh does not require starting a model turn. The workflow does not force continuation, synthesize native terminal events, or change Git history.

### Requirement: Verify commit-to-publication latency on each supported harness

Delivery evidence SHALL demonstrate the workflow independently on native Codex and OpenCode.

#### Scenario: A long native task commits and returns normally

- **WHEN** the native verification runs a long task that makes a real code commit and settles the same source interval
- **THEN** the evidence records the time from the commit to native completion, from completion to observable source facts, and from that visibility to the first queryable settled publication. For the manual feedback path, evidence distinguishes stderr emission, delivery to model input, and agent response. The result compares baseline and candidate behavior for any changed stage and reports unsupported or unobserved stages explicitly.
