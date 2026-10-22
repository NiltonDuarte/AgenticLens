# Implementation Plan: AL-37 Iteration 03 - Honest OpenCode Coupling and Literal Nonce Mandate

## 1. Goal and Iteration Scope

Correct docs/SETUP_MANUAL.md and .opencode/skills/agentic-lens-setup/SKILL.md to be completely honest regarding AgenticLens coupling. AgenticLens is built exclusively for OpenCode; it is not compatible with any other agent harness.

Specific changes required in this iteration:
1. Strip all claims of universal harness compatibility and remove all mentions of Claude Code, Cursor, Aider, GitHub Copilot, and generic agent harnesses.
2. Explicitly document that AgenticLens is tightly coupled to OpenCode internals. State plainly that an active OpenCode installation and its local SQLite database (~/.local/share/opencode/opencode.db) with schema tables session, message, and part are strictly required.
3. Clarify what IS agnostic: development methodology within OpenCode. Teams running OpenCode can use single-agent loops, custom subagents, custom ticket prefixes, custom workspace paths, and custom branching models without being forced into this repository's maintainer personas.
4. Eliminate references to CLAUDE.md in documentation and setup skill instructions, focusing exclusively on OpenCode configuration and instruction files (AGENTS.md, .opencode/agent/*.md).
5. Enforce the literal nonce mandate from FIX-5 across all documented generate CLI commands: remove shell variables (NONCE=$(...) and $NONCE) and instruct agents to supply a directly generated literal token (--nonce <GENERATED_NONCE>) because OpenCode captures tool input strings before subshell expansion.
6. Verify repository health through standard quality gates: make check, make test, and make build, while asserting zero emojis across all deliverables.

---

## 2. Grounded Reference Architecture & File Paths

The plan is grounded against concrete files, adapters, and commands in the AgenticLens repository:

### Documentation Deliverables
- docs/SETUP_MANUAL.md:
  - Primary onboarding and operational manual.
  - Revise Introduction, Section 1, Section 2, Section 3, Section 4, Section 5, Section 6, and Appendix to enforce honest OpenCode coupling, methodology neutrality within OpenCode, and the literal nonce mandate.

### Setup Skill Deliverables
- .opencode/skills/agentic-lens-setup/SKILL.md:
  - Autonomous setup skill manifest for OpenCode.
  - Update description, operational mandates, and Step 4 instructions inspection to remove CLAUDE.md and highlight OpenCode prerequisites.

### OpenCode Imperative Shell Adapters (Grounding References)
- backend/internal/adapters/opencode/client.go:
  - OpenCode SQLite connection manager (DefaultDBPath, OpenDB).
  - Proves hardwired coupling to local SQLite file path.
- backend/internal/adapters/opencode/extractor.go:
  - SQLite query extractor parsing raw message, part, tokens, and tool invocations from OpenCode database tables.
- backend/internal/adapters/opencode/resolver.go:
  - Session resolution logic matching tickets and nonces against OpenCode SQLite rows.
- backend/internal/adapters/live/sqlite.go:
  - Real-time polling service directly querying OpenCode SQLite database for active turns and SSE events.

### CLI Subcommands and Domain Config Grounding
- backend/cmd/agentic-lens/main.go:
  - Subcommands: init, generate, scan, server, version, backfill, migrate.
- backend/internal/domain/config.go:
  - Configuration structure: ProjectConfig, TicketConfig, WorkspacesConfig, DatabaseConfig, LiveConfig, CodeMetricsConfig, RoleMappings, RoleOrder.

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Revise Platform Setup Manual (docs/SETUP_MANUAL.md)
- Task 1.1: Revise Introduction and Section 1 (Quick Start).
  - Update introductory paragraphs to explicitly state that AgenticLens is an observability and developer telemetry platform built exclusively for OpenCode workflows, requiring an active OpenCode installation and its local SQLite database.
  - In Section 1.1 Step 5 (Mandatory User Confirmation Gate), remove CLAUDE.md from the list of instruction files. Restrict inspection targets to AGENTS.md, .opencode/agent/*.md, or other OpenCode-specific instruction files.
- Task 1.2: Rewrite Section 2 (Agnostic Adoption for OpenCode Projects).
  - Retitle Section 2 to "Agnostic Adoption for OpenCode Projects".
  - Eliminate all claims of universal harness compatibility. Remove all references to Claude Code, Cursor, Aider, GitHub Copilot, and generic agent harnesses.
  - State honestly and plainly that AgenticLens is hardwired to OpenCode internals because its data ingestion engine queries OpenCode's SQLite database tables (session, message, part). It cannot function without OpenCode.
  - Clarify what IS agnostic: within OpenCode, AgenticLens is completely agnostic to engineering methodology:
    - Workflow Topology Agnostic: Supports single-agent prompt loops, hierarchical subagent structures, human-in-the-loop flows, or automated pipelines running inside OpenCode.
    - Persona and Role Agnostic: Custom agent persona strings configured in OpenCode can be mapped to canonical roles (via role_mappings in .agenticlens-config.yaml) or tracked as standalone roles, with zero coupling to this repository's internal maintainer personas (agentic-lens-dev, agentic-lens-po, agentic-lens-review).
    - Taxonomy and Workspace Agnostic: Supports arbitrary ticket prefixes (e.g. PAY, SEC, OPS), custom ticket regex patterns, and custom workspace folder locations (e.g. workspaces, alspaces, .tasks).
    - Branching and Quality Agnostic: Operates with any git branching model (trunk-based, feature branch, GitFlow) and any testing/review procedures.
- Task 1.3: Update Section 4 (OpenCode Agent Workflow Integration Guide).
  - In Section 4.2 (Nonce Generation and Ingestion Command), remove the shell script snippet containing NONCE=$(od ...) and --nonce "$NONCE".
  - Replace with the literal nonce command structure: ./bin/agentic-lens generate --nonce <GENERATED_NONCE> ... where the agent generates a literal 8 to 16 character alphanumeric or hex string directly.
  - Explain the technical rationale from FIX-5: OpenCode records tool input arguments verbatim into SQLite before shell variable expansion or subshell execution occurs. If an agent executes a command using $NONCE or $(...), OpenCode stores the literal string "$NONCE" in the database part output, causing database lookups for the expanded token to fail.
  - In Section 4.5 (Standard Repository Agent Guidelines Policy Block), remove CLAUDE.md from the file list and update the embedded policy block to enforce literal nonce generation without shell variables.
- Task 1.4: Audit and Harmonize Section 3, Section 5, Section 6, and Appendix.
  - In Section 3.2 (database block), reinforce that database.path specifies the local OpenCode SQLite database.
  - In Section 5.1 (generate command), clarify that session telemetry is extracted from OpenCode's SQLite database using literal nonces.
  - In Section 6.1 and 6.2 (Troubleshooting), verify that concurrency and cross-platform paths specifically address OpenCode SQLite.
  - In Appendix A.3 and A.4, reaffirm that telemetry facts and live SSE streams are derived from local OpenCode SQLite session traces.
  - Ensure zero references to non-OpenCode harnesses remain in any section.

### Phase 2: Revise Autonomous Setup Skill Manifest (.opencode/skills/agentic-lens-setup/SKILL.md)
- Task 2.1: Update frontmatter, description, and operational mandates.
  - Explicitly state that the setup skill bootstraps AgenticLens in OpenCode repositories, requiring an active OpenCode installation and its SQLite database.
  - In Operational Mandates, remove CLAUDE.md from the Mandatory User Confirmation Gate file inspection list.
- Task 2.2: Update Step 4 (Agent Guidelines Review and Mandatory Confirmation Gate).
  - Remove CLAUDE.md from the list of existing agent instruction files to inspect.
  - In the Proposed Agent Policy Block, verify that the generate command uses --nonce <GENERATED_NONCE> with a literal token generated directly by the agent and explicitly warns against shell variable expansion.

### Phase 3: Repository-Wide Audit and Quality Gates
- Task 3.1: Execute text search audits across docs/SETUP_MANUAL.md and .opencode/skills/agentic-lens-setup/SKILL.md to ensure zero occurrences of Claude Code, Cursor, Aider, GitHub Copilot, and CLAUDE.md.
- Task 3.2: Verify zero occurrences of shell variable substitution ($NONCE or $(...)) in documented generate CLI invocations.
- Task 3.3: Verify zero emojis across all modified files and deliverables.
- Task 3.4: Run quality gates: make check (Go vet, TypeScript check, ESLint, git diff check), make test (Go unit tests with race detection, frontend tests), and make build (single binary build with embedded dashboard).

---

## 4. Test and Verification Commands

The coder and tester subagents will execute the following concrete verification commands in prose:

1. Static Code Analysis and Code Invariant Checks:
   - Run make check.
   - Expected Output: Go vet exits cleanly, npm typecheck exits cleanly, npm lint exits cleanly, and git diff check passes.

2. Automated Unit Test Suite:
   - Run make test.
   - Expected Output: All Go tests in backend/ pass with race detection enabled; all frontend tests in web/ pass.

3. Single Binary Build:
   - Run make build.
   - Expected Output: Multi-platform binaries compile without errors and bin/agentic-lens is generated successfully.

4. Non-OpenCode Harness Name Audit:
   - Run grep searches across docs/SETUP_MANUAL.md and .opencode/skills/agentic-lens-setup/SKILL.md for patterns matching Claude, Cursor, Aider, Copilot, and CLAUDE.md.
   - Expected Output: Zero matches in docs/SETUP_MANUAL.md (excluding any legitimate model names like Claude 3.5 Sonnet if referenced strictly as an LLM model identifier in analytics views) and zero matches in SKILL.md.

5. Literal Nonce Compliance Audit:
   - Run grep searches across docs/SETUP_MANUAL.md and .opencode/skills/agentic-lens-setup/SKILL.md for $NONCE or $(od.
   - Expected Output: Zero matches returned. All commands must use literal token placeholders like --nonce <GENERATED_NONCE>.

6. Zero Emojis Verification:
   - Run grep on modified files to assert zero unicode emoji characters.
   - Expected Output: Zero emoji matches across all deliverables.
