# Iteration 01 Plan: Transitioning Agent Guidelines and Skills from Shell-Expanded Nonces to Agent-Generated Literal Nonces

## Objective
Update all repository agent guidelines, orchestrator instructions, and skill manifests to deprecate shell-based variable assignment for telemetry nonces. Replace them with explicit instructions mandating that the agent generate a random nonce directly and embed the literal token in the CLI arguments, ensuring OpenCode records the exact nonce in SQLite prior to execution.

## Grounded Context References
- AGENTS.md: Section 1, Rule 6, lines 29-39.
- .opencode/agent/orchestrator.md: Workflow Execution Loop Step 6, lines 72-76.
- .opencode/skills/agentic-lens-dev/SKILL.md: Phase 3, Step 6, lines 115-121.
- .opencode/skills/agentic-lens-po/SKILL.md: Section 2, Step 4, lines 81-85.
- .opencode/skills/agentic-lens-setup/SKILL.md: Step 3 CLI walkthrough (lines 104-113) and Step 4 Proposed Agent Policy Block (lines 170-181).
- backend/internal/adapters/opencode/resolver.go: ResolveActiveRootSession Priority 1 exact nonce matching against the SQLite part table (p.data).
- scripts/test-setup.sh: Automated verification suite asserting zero emojis and setup skill integrity.

## Phased Implementation Tasks

### Phase 1: Update Repository Operational Guidelines (AGENTS.md)
1. Navigate to Section 1, Rule 6 (Telemetry Ingestion as the Final Step & Conventional Commit Type Classification).
2. Remove the shell command substitution lines assigning NONCE via od and /dev/urandom.
3. Replace the bash snippet with an invocation taking an explicit literal nonce argument.
4. Add clear instructional prose stating that the agent must generate a unique random nonce (an 8 to 16 character alphanumeric or hex string) directly in its own response and embed the literal token into the command argument.
5. Explicitly warn against using shell variables or subshell evaluation ($NONCE or $(...)), explaining that OpenCode persists tool call parameters into SQLite before subshell execution, which causes database searches for the expanded token to fail.

### Phase 2: Update Orchestrator Agent Persona (.opencode/agent/orchestrator.md)
1. Locate Workflow Execution Loop Step 6 (Delivery Sync & Self-Reflexive Telemetry via AgenticLens).
2. Remove the two-line shell snippet containing NONCE=$(od ...) and ./bin/agentic-lens generate --nonce $NONCE.
3. Update the execution instructions to command the orchestrator agent to generate a unique random nonce directly and pass the literal token in the --nonce argument of the generate command.
4. Include the workspace directory placeholder consistent with orchestrator formatting rules.

### Phase 3: Update Developer and Product Owner Skills (.opencode/skills/)
1. In .opencode/skills/agentic-lens-dev/SKILL.md (Phase 3, Step 6):
   - Replace the shell-based nonce generation block with the direct CLI invocation using a literal nonce generated directly by the agent.
   - Clarify in the accompanying text that the agent must generate a fresh random token for each run and avoid shell variable indirection.
2. In .opencode/skills/agentic-lens-po/SKILL.md (Section 2, Step 4):
   - Replace the shell-based nonce generation block with the direct CLI invocation using a literal nonce generated directly by the agent.
   - Preserve all existing PO flags (--ticket-id PO-#, --ticket-type docs, --ticket-title).

### Phase 4: Update Interactive Setup Skill (.opencode/skills/agentic-lens-setup/SKILL.md)
1. In Step 3 (Platform Capability Walkthrough, generate subcommand):
   - Remove the NONCE=$(od ...) assignment and --nonce $NONCE argument.
   - Present the generate command using an example literal nonce string, adding explanatory prose that the invoking agent generates the token directly.
2. In Step 4 (Proposed Agent Policy Block, subsection 3 Delayed Telemetry Ingestion Mandate):
   - Update the template block that gets proposed to users during setup.
   - Replace the shell variable snippet with the literal nonce command structure.
   - Document the requirement that agents generate the token directly so that target repositories adopting AgenticLens also avoid the OpenCode SQLite timing bug.

### Phase 5: Verification, Validation, and Quality Gates
1. Perform a recursive repository search across all non-historical markdown files for any residual references to od -N 4, /dev/urandom, or --nonce $NONCE. Confirm that only literal nonce instructions remain.
2. Execute the setup test suite (scripts/test-setup.sh) to verify that skill manifests remain valid, script execution succeeds, and the zero-emoji invariant is strictly preserved.
3. Run the Go backend test suite (make test or go test ./...) to ensure full repository build and test hygiene.

## Acceptance and Verification Gates
1. Grep verification confirms zero occurrences of od -N 4 or --nonce $NONCE in AGENTS.md and .opencode/.
2. scripts/test-setup.sh passes cleanly with exit code 0.
3. Zero emojis introduced in any modified documentation or manifest files.
