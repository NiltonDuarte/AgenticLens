# Implementation Plan: AL-21 Iteration 02 - Ticket Prefix Enrichment, PO Support, CLI Parity, and Skill Configuration Migration

## 1. Goal and Iteration Scope

Iteration 02 addresses user steering feedback on ticket AL-21 across four critical requirements:
1. First, expand the domain ticket prefix model to include the existing PO prefix alongside AL and GEN within default project configuration.
2. Second, attach descriptive human-readable purpose definitions to ticket prefixes, defining a dedicated TicketPrefix structure and implementing flexible parsing that accepts both rich objects (prefix and description) and plain string scalars in YAML and JSON.
3. Third, prove that executing the CLI application against all existing repository tickets across AL, GEN, and PO families produces 100% identical metric ledgers and summary outputs.
4. Fourth, delete the superseded .agenticlens/config.json and migrate all agent definitions (.opencode/agent/*.md), multi-agent skills (.opencode/skills/*), and top-level instructions (AGENTS.md) to inspect and reference .agenticlens/config.yaml.

This iteration maintains strict Functional Core, Imperative Shell (FCIS) separation: domain models remain pure value-semantic structs in backend/internal/domain/config.go with zero I/O imports, while multi-format deserialization and upward filesystem resolution live in backend/internal/adapters/config/loader.go.

## 2. Grounded Reference Architecture and Symbols

The implementation targets specific symbols and locations across the codebase:

- Pure Functional Domain Core:
  - File: backend/internal/domain/config.go
  - Symbols: TicketPrefix struct, TicketConfig struct with Prefixes slice of TicketPrefix, DefaultProjectConfig function, NormalizeTicketID function.
  - File: backend/internal/domain/config_test.go
  - Symbols: TestDefaultProjectConfig, TestNormalizeTicketID_ExactMatches, TestNormalizeTicketID_CaseInsensitivity, TestNormalizeTicketID_BranchAndTokenPatterns, TestNormalizeTicketID_AdversarialEdgeCases.
- Imperative Shell Configuration Loader:
  - File: backend/internal/adapters/config/loader.go
  - Symbols: rawTicketPrefix type with flexible unmarshaling methods, rawTicketConfig struct, rawProjectConfigDTO struct, mergeConfig function, FindConfigFile function, LoadFromFile function, Load function.
  - File: backend/internal/adapters/config/loader_test.go
  - Symbols: TestLoadFromFile_YAML, TestLoadFromFile_JSON, TestLoad_FlexiblePrefixParsing, TestLoad_DefaultProjectConfigFallback.
- CLI Entrypoint and Workspace Scanner:
  - File: backend/cmd/agentic-lens/main.go
  - Symbols: runGenerate, runBackfill, main, printAnomalyDiagnostics.
  - File: backend/internal/adapters/repository.go
  - Symbols: ResolveWorkspacesDir, ScanWorkspaces, ReadTicketArtifacts.
  - File: backend/cmd/agentic-lens/main_test.go and backend/internal/adapters/repository_test.go
  - Verification target: Existing ticket workspaces under workspaces/ (AL-1 through AL-7, AL-10 through AL-12, AL-21, GEN-1, GEN-2, and PO-1 through PO-4).
- Configuration Artifacts:
  - File: .agenticlens/config.yaml (update with AL, GEN, PO prefix definitions and descriptions).
  - File: .agenticlens/config.json (delete from repository root).
- Repository Agent Instructions and Skills:
  - File: AGENTS.md (update ticket prefix guidance and config path).
  - File: .opencode/agent/orchestrator.md (update kickoff config inspection and prefix expectations).
  - File: .opencode/agent/planner.md (update core references from config.json to config.yaml).
  - File: .opencode/agent/coder.md (update mandatory guidelines from config.json to config.yaml).
  - File: .opencode/agent/tester.md (update core references from config.json to config.yaml).
  - File: .opencode/agent/reviewer.md (audit against config.yaml specifications).
  - File: .opencode/skills/agentic-lens-dev/SKILL.md (update configuration section, parameter definitions, and subagent prompt templates).
  - File: .opencode/skills/agentic-lens-po/SKILL.md (update configuration section and prefix conventions).
  - File: .opencode/skills/agentic-lens-review/SKILL.md (update baseline documents section).

## 3. Phased Step-by-Step Implementation Tasks for Coder

### Phase 1: Pure Domain Model Refactoring (backend/internal/domain/config.go)
- Task 1.1: Define TicketPrefix struct:
  - Declare fields: Prefix as string with json tag "prefix" and yaml tag "prefix"; Description as string with json tag "description" and yaml tag "description".
  - Ensure TicketPrefix adheres strictly to value semantics with no pointers.
- Task 1.2: Refactor TicketConfig struct:
  - Change Prefixes field type from string slice to a slice of TicketPrefix.
  - Preserve CaseInsensitive as boolean and CustomPatterns as string slice.
  - Add pure helper method PrefixNames() on TicketConfig returning a slice of strings containing each non-empty prefix name, facilitating convenient string-based lookups where descriptions are not required.
- Task 1.3: Update DefaultProjectConfig:
  - Populate Ticket.Prefixes with three default TicketPrefix instances:
    1. Prefix "AL", Description "Core AgenticLens development features, bugfixes, and technical initiatives".
    2. Prefix "GEN", Description "General multi-agent engineering tasks and ad-hoc untracked work".
    3. Prefix "PO", Description "Product Owner specifications, requirements definitions, and roadmap breakdowns".
  - Maintain existing defaults for CaseInsensitive (true), Workspaces (dir "workspaces", src_folders "web" and "backend"), Database (path, timeouts, read_only), Live (thresholds and anomalies), CodeMetrics, and RoleMappings.
- Task 1.4: Refactor NormalizeTicketID:
  - Accept raw string and TicketConfig.
  - Extract candidate prefixes from cfg.Prefixes using prefix names.
  - If extracted prefix list is empty, fall back to default prefix names: AL, GEN, and PO.
  - Clean each prefix by trimming whitespace, trailing hyphens, and trailing underscores.
  - Sort candidate prefixes descending by length to ensure longer prefixes take precedence over shorter prefixes sharing a common substring.
  - Construct regular expression matching start-of-string or non-alphanumeric boundary, quoted prefix pattern, hyphen, numeric digits, and ending boundary.
  - Apply case-insensitive flag when cfg.CaseInsensitive is true.
  - Extract matching numeric digits and format canonical string as uppercase prefix hyphen digits (e.g. PO-2, AL-21, GEN-1).
  - If no prefix matches, evaluate CustomPatterns before returning an empty string and false.

### Phase 2: Flexible Prefix Deserialization in Loader Adapter (backend/internal/adapters/config/loader.go)
- Task 2.1: Define flexible raw prefix representation:
  - Declare rawPrefixItem struct containing Prefix string and Description string.
  - Implement UnmarshalYAML on rawPrefixItem:
    - Inspect incoming yaml.Node.
    - If node kind is ScalarNode, assign node value to Prefix and empty string to Description.
    - If node kind is MappingNode, decode into an anonymous struct containing Prefix and Description strings.
    - If node kind is neither, return a descriptive wrapped error indicating unexpected YAML structure.
  - Implement UnmarshalJSON on rawPrefixItem:
    - Inspect incoming byte slice.
    - If byte slice begins with double-quote, unmarshal as a string scalar and set Prefix to the parsed string with empty Description.
    - If byte slice begins with curly brace, unmarshal into an anonymous struct containing prefix and description string fields.
    - If byte slice represents neither, return an error indicating invalid prefix format.
- Task 2.2: Update rawTicketConfig and rawProjectConfigDTO:
  - In rawTicketConfig, declare Prefixes as a slice of rawPrefixItem.
  - Retain CaseInsensitive as pointer to boolean and CustomPatterns as string slice.
  - Retain LegacyTicketPrefix as pointer to string at root level of rawProjectConfigDTO for backward compatibility with older test harnesses.
- Task 2.3: Update mergeConfig logic:
  - If modern Ticket configuration exists and its Prefixes slice contains items, convert each rawPrefixItem into domain.TicketPrefix and assign to cfg.Ticket.Prefixes.
  - If modern prefixes are absent but LegacyTicketPrefix is present and non-empty, convert the legacy string into a single domain.TicketPrefix with empty description.
  - If neither modern nor legacy prefixes are present, retain domain defaults from DefaultProjectConfig (AL, GEN, PO).
- Task 2.4: Ensure robust YAML and JSON file loading:
  - Confirm LoadFromFile and Load functions deserialize both modern structured prefix objects and scalar prefix lists without panics, errors, or data loss.

### Phase 3: Configuration Manifest & Workspace Artifact Migration
- Task 3.1: Update .agenticlens/config.yaml:
  - In ticket.prefixes, write structured items with prefix and description keys for AL, GEN, and PO.
  - Add explanatory comments above the prefixes section describing the role of each prefix and clarifying that both scalar and mapping formats are valid.
  - Retain existing configurations for workspaces, database, live observability, code metrics, and role mappings.
- Task 3.2: Delete .agenticlens/config.json:
  - Remove .agenticlens/config.json from the repository root using filesystem deletion.
  - Verify git status records the deletion of config.json.

### Phase 4: Agent Instructions and Multi-Agent Skills Modernization
- Task 4.1: Update AGENTS.md:
  - In Section 1 (Ticket Tracking Mandate), update Rule 1 to state that missing ticket IDs prompt the user or generate a GEN-# ticket using configured prefixes from .agenticlens/config.yaml (with fallback GEN).
  - Update format examples to cite AL, GEN, and PO explicitly (e.g. AL-42, GEN-42, PO-2).
- Task 4.2: Update .opencode/skills/agentic-lens-dev/SKILL.md:
  - Update heading to reference .agenticlens/config.yaml.
  - Update startup instructions to inspect .agenticlens/config.yaml instead of config.json.
  - Update parameter descriptions to explain ticket.prefixes containing AL, GEN, and PO.
  - In Iteration Lifecycle prompt templates (Phase 1, iteration 01 and iteration > 01), replace Config: .agenticlens/config.json with Config: .agenticlens/config.yaml.
- Task 4.3: Update .opencode/skills/agentic-lens-po/SKILL.md:
  - Update heading and startup instructions to reference .agenticlens/config.yaml.
  - Explicitly document PO prefix alongside AL and GEN.
- Task 4.4: Update .opencode/skills/agentic-lens-review/SKILL.md:
  - In Section 1 (Scope & Hotspot Discovery, item 3), replace reference to .agenticlens/config.json with .agenticlens/config.yaml.
- Task 4.5: Update .opencode/agent/*.md files:
  - orchestrator.md: Change Step 1 kickoff instructions to inspect .agenticlens/config.yaml with fallback ticket prefixes AL, GEN, PO.
  - planner.md: Change Core References to cite .agenticlens/config.yaml.
  - coder.md: Change Mandatory Guidelines to cite .agenticlens/config.yaml.
  - tester.md: Change Core References to cite .agenticlens/config.yaml.
  - reviewer.md: Update references if any exist.

### Phase 5: Verification of CLI Output Parity Across All Existing Tickets
- Task 5.1: Compile fresh CLI binary:
  - Execute make build to compile bin/agentic-lens using updated Go packages.
- Task 5.2: Execute baseline comparison and parity audit:
  - Run bin/agentic-lens --workspaces workspaces --list to verify that all 17 existing ticket workspaces (AL-1, AL-2, AL-3, AL-4, AL-5, AL-6, AL-7, AL-10, AL-11, AL-12, AL-21, GEN-1, GEN-2, PO-1, PO-2, PO-3, PO-4) are discovered without warnings or parse errors.
  - For each ticket, run bin/agentic-lens --workspaces workspaces --ticket-id <TICKET_ID> and compare the resulting markdown table against the corresponding workspaces/<TICKET_ID>/SUMMARY.md metrics table.
  - Verify that token totals, active duration, costs, health statuses, SDLC iteration counts, and anomaly summaries match the existing summary tables with zero divergence.
  - Run bin/agentic-lens backfill --workspaces workspaces --dry-run to confirm all ledgers pass schema validation.

## 4. Adversarial Testing Strategy for Tester

### Pure Domain Unit Tests (backend/internal/domain/config_test.go)
- Test Case 1: Default Configuration Prefix Completeness:
  - Instantiate DefaultProjectConfig.
  - Assert that Ticket.Prefixes contains exactly 3 entries: AL, GEN, and PO.
  - Assert that each entry has both a non-empty Prefix and a non-empty Description string.
  - Assert that PrefixNames() returns a slice containing "AL", "GEN", and "PO".
- Test Case 2: Ticket Normalization with PO Prefix:
  - Test exact match: "PO-1" returns "PO-1" and true. "PO-4" returns "PO-4" and true.
  - Test lowercase normalization: "po-1", "po-2", "Po-3" return canonical "PO-1", "PO-2", "PO-3" with true.
  - Test git branch and token patterns: "spec/po-1-rewrite-core", "feature/po-2-planning", "workspaces/PO-3/agentic_metrics.json" return canonical "PO-1", "PO-2", "PO-3" and true.
- Test Case 3: Ticket Normalization Across All Default Families:
  - In a single table test using DefaultProjectConfig(), verify "al-21" -> "AL-21", "gen-2" -> "GEN-2", "po-4" -> "PO-4".
- Test Case 4: Prefix Length Precedence with Custom Overlaps:
  - Configure TicketConfig with prefixes "PO" and "POSTORE".
  - Normalize "postore-99" and assert it returns "POSTORE-99", confirming that longer prefixes are prioritized over shorter prefix subsets.
- Test Case 5: Empty and Blank Prefix Resilience:
  - Test TicketConfig with empty Prefixes slice; assert it falls back to AL, GEN, and PO defaults.
  - Test TicketConfig with whitespace and punctuation prefixes; assert trailing hyphens and whitespace are sanitized properly.

### Imperative Loader Boundary Tests (backend/internal/adapters/config/loader_test.go)
- Test Case 6: YAML Structured Prefix Objects:
  - Create temporary config.yaml specifying prefixes as a list of mapping objects with prefix and description keys.
  - Call Load and assert that cfg.Ticket.Prefixes contains the structured entries with populated descriptions.
- Test Case 7: YAML Plain String Scalars:
  - Create temporary config.yaml specifying prefixes as a plain list of strings (e.g. "- AL", "- GEN", "- PO").
  - Call Load and assert that cfg.Ticket.Prefixes parses successfully with correct prefix names and empty descriptions.
- Test Case 8: YAML Mixed Prefix Formats:
  - Create temporary config.yaml containing a mixed list where the first item is an object and the second item is a string scalar.
  - Call Load and assert both items deserialize cleanly into TicketPrefix structs without parse error.
- Test Case 9: JSON Structured and String Prefix Compatibility:
  - Create temporary JSON configs testing structured prefix objects and plain string prefix arrays.
  - Call LoadFromFile and assert both formats yield identical prefix names.
- Test Case 10: Legacy Flat JSON Compatibility:
  - Create temporary config.json with root-level "ticket_prefix": "PO".
  - Call Load and assert cfg.Ticket.Prefixes contains one TicketPrefix with Prefix "PO".
- Test Case 11: Priority After config.json Deletion:
  - In a directory containing only config.yaml, assert that config.yaml is loaded and config.json absence produces no error.
- Test Case 12: Invalid Prefix Format Handling:
  - Provide YAML with prefixes as an integer or boolean.
  - Assert LoadFromFile returns a descriptive wrapped error rather than panicking.

### CLI Parity and Repository Scan Regression Tests
- Test Case 13: Repository Workspace Scanning Parity:
  - Execute automated test calling adapters.ScanWorkspaces("workspaces") against the real repository workspaces directory.
  - Assert that report.Errors is empty.
  - Assert that all 17 tickets are present in report.Ledgers.
  - Verify that tickets starting with PO (PO-1, PO-2, PO-3, PO-4) are correctly scanned and have valid derived metrics.
- Test Case 14: Documentation Consistency Static Audit:
  - Execute a test or inspection verifying that zero occurrences of ".agenticlens/config.json" remain in AGENTS.md, .opencode/agent/*.md, and .opencode/skills/*.md.
  - Verify that .agenticlens/config.json does not exist on the filesystem.

## 5. Review Criteria for Reviewer

### Axis 1: Specification and Contract Alignment
- Model Definition: Does backend/internal/domain/config.go define TicketPrefix with Prefix and Description string fields, and does TicketConfig.Prefixes use a slice of TicketPrefix?
- Operational Defaults: Does DefaultProjectConfig() include AL, GEN, and PO prefixes with descriptive purposes?
- Flexible Deserialization: Does backend/internal/adapters/config/loader.go support both structured mapping objects (prefix + description) and plain string scalars in YAML and JSON?
- Config Manifest: Is .agenticlens/config.yaml updated with structured definitions for AL, GEN, and PO with descriptions?
- Artifact Deletion: Has .agenticlens/config.json been cleanly deleted from the repository?
- Skill and Agent Updates: Are AGENTS.md, all three skill documents in .opencode/skills/, and all five agent files in .opencode/agent/ updated to reference .agenticlens/config.yaml?
- CLI Parity: Has running bin/agentic-lens across all existing tickets (AL-*, GEN-*, PO-*) been verified to produce identical results without changes to existing metric ledgers?

### Axis 2: Engineering Standards and Technical Invariants
- FCIS Discipline: Is backend/internal/domain/config.go strictly isolated with zero I/O, network, or external YAML library imports? Does multi-format unmarshaling remain strictly in backend/internal/adapters/config/?
- Strong Typing: Is there zero usage of any or empty interface{} across domain and adapter packages?
- Value Semantics: Are TicketPrefix and TicketConfig immutable value types?
- Error Handling: Are deserialization errors properly wrapped with context using %w?
- Zero Emojis: Are code, comments, config files, agent prompts, and documentation 100% free of emojis?
- Quality Gates: Do make check and make test pass with zero warnings, zero race conditions, and zero skipped tests?
