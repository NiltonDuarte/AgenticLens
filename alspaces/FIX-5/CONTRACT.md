# Technical Contract: Ticket FIX-5

## Problem Statement
AgenticLens relies on a multi-tiered session resolution algorithm in backend/internal/adapters/opencode/resolver.go (ResolveActiveRootSession) to bind telemetry ingestion to the root OpenCode session executing a given ticket. Priority 1 performs an exact match by querying the OpenCode SQLite database for a unique execution token stored in the part table:
SELECT DISTINCT s.id FROM part p JOIN session s ON ... WHERE s.parent_id IS NULL AND p.data LIKE ?

Current operational guidelines across AGENTS.md, .opencode/agent/orchestrator.md, and .opencode/skills/ prescribe a shell-based nonce generation pattern:
NONCE=$(od -N 4 -tx1 /dev/urandom | head -1 | tr -d ' ')
./bin/agentic-lens generate --nonce $NONCE ...

This pattern fails due to the persistence lifecycle of OpenCode:
1. OpenCode records tool call inputs into SQLite before subshell execution begins.
2. When the tool call payload contains bash variable assignment and expansion (NONCE=$(...) and --nonce $NONCE), the part.data column stores the unexpanded literal string "$NONCE".
3. The subshell process subsequently initializes, evaluates the command substitution, and passes the resolved hex token to the agentic-lens generate CLI process.
4. ResolveActiveRootSession queries p.data LIKE '%<RESOLVED_TOKEN>%'. Because the database record holds the literal string "$NONCE", the query returns sql.ErrNoRows.
5. Priority 1 exact nonce resolution misses completely. The CLI falls back to Priority 2 (explicit session ID) or Priority 3 (ticket ID in session title or workspace path). In multi-agent environments, parallel runs, or sessions with divergent titles, Priority 3 either misattributes telemetry to foreign sessions or aborts execution with ErrSessionNotFound.

The resolution requires removing all shell-based nonce generation routines and dynamic bash variable expansions from agent guidelines, orchestrator instructions, and skill definitions. Instead, the agent must be instructed to generate a unique random nonce directly (by any means, such as an 8 to 16 character random hex or alphanumeric token) and embed the literal token directly into the command arguments. Because the tool call input itself will contain the concrete token, OpenCode will write the literal nonce into the SQLite part table before the command executes, restoring deterministic Priority 1 session resolution.

## Core Architectural Invariants

### 1. Functional Core, Imperative Shell (FCIS) Boundaries
- Pure Domain (backend/internal/domain/): Domain models, metric calculations, and anomaly evaluations remain completely decoupled from operating system shells, random number generation, and database persistence. No domain changes are required.
- Imperative Shell (backend/internal/adapters/opencode/resolver.go): The backend resolver logic is correct as designed. ResolveActiveRootSession continues to enforce Priority 1 exact matching against part.data.
- Agent Interaction Boundary (AGENTS.md, .opencode/): Operational documents define the interaction contract between LLM agents and the CLI binary. Prescribing literal argument embedding ensures data contracts are satisfied before tool execution commences.

### 2. OpenCode Persistence Timing Invariant
- Tool execution parameters are frozen in OpenCode SQLite storage at dispatch time, not execution termination time.
- Dynamic shell evaluation constructs ($VAR, $(...), `...`) must never be used for arguments intended for database lookup.
- The executing agent is the authoritative source of generation for session identification tokens.

### 3. Zero Emojis Mandate
- No emojis permitted across modified agent instructions, skills, markdown documentation, or workspace artifacts.

## Scope Boundaries

### In-Scope Files
- AGENTS.md: Section 1, Rule 6 (Telemetry Ingestion as the Final Step).
- .opencode/agent/orchestrator.md: Workflow Execution Loop Step 6 (Delivery Sync & Self-Reflexive Telemetry via AgenticLens).
- .opencode/skills/agentic-lens-dev/SKILL.md: Phase 3, Step 6 (Extract Telemetry & Render Metrics).
- .opencode/skills/agentic-lens-po/SKILL.md: Section 2, Step 4 (Generate Jira / Development Tickets).
- .opencode/skills/agentic-lens-setup/SKILL.md: Step 3 (CLI Subcommands - generate) and Step 4 (Proposed Agent Policy Block - 3. Delayed Telemetry Ingestion Mandate).

### Out-of-Scope Files
- backend/internal/adapters/opencode/resolver.go: The SQL resolution logic is correct and remains untouched.
- backend/cmd/agentic-lens/main.go: Flag parsing for --nonce remains unchanged.
- Historical workspace ledgers and fixtures (alspaces/, testdata/): Past completed tickets and test fixtures remain immutable historical records.

## Acceptance Criteria
1. Elimination of Shell Nonce Generation: Zero occurrences of NONCE=$(od -N 4 -tx1 /dev/urandom | head -1 | tr -d ' ') and zero occurrences of --nonce $NONCE across AGENTS.md, .opencode/agent/orchestrator.md, .opencode/skills/agentic-lens-dev/SKILL.md, .opencode/skills/agentic-lens-po/SKILL.md, and .opencode/skills/agentic-lens-setup/SKILL.md.
2. Explicit Agent Generation Directive: Every target file explicitly directs the agent to generate a unique random nonce directly (e.g. an 8 to 16 character random hex or alphanumeric token) and interpolate the literal string directly into the command invocation.
3. Architectural Rationale Documented: Instructions explicitly explain that shell variables must be avoided because OpenCode writes tool inputs to SQLite before subshell expansion occurs.
4. Clean Validation: Automated setup checks in scripts/test-setup.sh and repository test suites pass with zero regressions and zero emojis.
