h3. Title: AL-50: Audit and Streamline Debug Prompt Generation and Sanitization Pipeline

*TARGET*: AgenticLens (presentation / web / utils / tests / review)

*CONTEXT*
 
Branch `fix/FIX-3-centralize-live-anomalies-debug-identifiers` introduced `web/src/utils/debugPrompt.ts` (420+ LOC) to generate structured markdown debug prompts across 4 operational layers (Tool Execution, Execution Issue, Message Turn, and Session Overview) for instant copy-pasting into an AI debugging assistant.

However, the implementation suffers from two major design defects:
1. **Excessive and Destructive Sanitization Pipeline**: The module applies heavyweight, multi-pass regexes (`stripInternalIdentifiers`, `cleanSeverityTiers`, `sanitizeTextForPrompt`) and a cascade of fragile grammatical repair regexes (cleaning orphaned commas, coordinating conjunctions, and dashes) across raw payloads. Applying this pipeline to raw tool inputs, tool outputs, compiler error messages, and agent deliberation irreversibly corrupts legitimate technical diagnostic data (e.g. git commands lose branch arguments, test names containing ticket prefixes get stripped, stack traces lose file paths, and syntax errors are mangled).
2. **Verbose and Redundant Prompt Messages**: The generated debug prompts contain repetitive boilerplate sentences (e.g. repeating "This data was extracted from the OpenCode database." across all layers), overly conversational provenance phrasing, and repetitive multi-line "### What to check" checklists with leading rhetorical questions.

*GOAL*
 
Audit and streamline `web/src/utils/debugPrompt.ts`:
1. Eliminate destructive regex sanitization on raw technical payloads (tool inputs, tool outputs, compiler errors, deliberation, and message bodies) so that code fences preserve exact, unmutated runtime artifacts.
2. Remove fragile regex grammar repairs and restrict any metadata sanitization strictly to top-level UI summary titles, or replace it with a lightweight, deterministic transform.
3. Review and refine prompt messages to be concise, direct, factual, and optimized for AI assistant ingestion.

*VERTICAL SLICE SCOPE*
 
* Domain Core / Utils:
  - Refactor `web/src/utils/debugPrompt.ts` to dramatically reduce complexity and size (targeting under 200 LOC, down from 420+ LOC).
  - Guarantee raw technical payloads (`input`, `output`, `errorMessage`, `reasoningContent`, `messageContent`) bypass destructive text replacement and are inserted verbatim into defensive Markdown code fences (`formatCodeBlock`).
  - Simplify helper routines (`sanitizeTextForPrompt`, `translateRuleToPlainEnglish`), removing fragile punctuation and conjunction stripping regexes.
  - Refactor prompt templates across all 4 layers (Tool Execution, Execution Issue, Message Turn, Session Overview) to use direct, concise diagnostic wording and sharp, non-redundant checklists.
* Boundary Adapter: Zero backend/API changes; leverages existing live event data.
* Presentation: Verify that `CopyPromptButton` in `LiveTraceInspector.tsx` and anomaly cards receive clean, faithful Markdown output.
* Quality: Update and expand `web/src/utils/debugPrompt.test.ts` to assert that code blocks and payloads containing ticket IDs (e.g. `AL-42`), branch names (`fix/FIX-3`), and severity tags remain unmutated inside code fences, while prompt metadata remains clean and properly formatted.

*ACCEPTANCE CRITERIA*
 
* Tool inputs, outputs, error messages, and message contents are preserved verbatim inside fenced code blocks without destructive string replacement or regex mangling.
* Fragile regex-based grammar mending (orphan commas, trailing conjunctions) is eliminated.
* Code size of `web/src/utils/debugPrompt.ts` is reduced by at least 40% (under 250 LOC).
* Repetitive boilerplate phrases (e.g. "This data was extracted from the OpenCode database.") are removed or condensed into clean, concise provenance headers.
* Checklists are concise, actionable, and free of rhetorical filler.
* Unit tests in `web/src/utils/debugPrompt.test.ts` pass with 100% assertion coverage, including test cases explicitly verifying that technical payloads containing tickets, branches, and keywords are not corrupted.
* `npm run typecheck`, `npm run lint`, and `npm test` pass with zero errors.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
