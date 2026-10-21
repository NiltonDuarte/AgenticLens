import { describe, it, expect } from 'vitest';
import {
  generateToolExecutionPrompt,
  generateExecutionIssuePrompt,
  generateMessageTurnPrompt,
  generateSessionOverviewPrompt,
  stripInternalIdentifiers,
  translateRuleToPlainEnglish,
  sanitizeTextForPrompt,
} from './debugPrompt';

describe('debugPrompt Utility Module', () => {
  describe('stripInternalIdentifiers', () => {
    it('strips ticket identifiers such as AL-20, FIX-3, GEN-42, PO-2', () => {
      const input = 'Resolving issue on ticket AL-20 and fixing bug in FIX-3 for GEN-42 and PO-2.';
      const output = stripInternalIdentifiers(input);
      expect(output).not.toContain('AL-20');
      expect(output).not.toContain('FIX-3');
      expect(output).not.toContain('GEN-42');
      expect(output).not.toContain('PO-2');
      expect(output).not.toContain('ticket');
    });

    it('strips git branch references such as feat/..., fix/..., and refs/heads/...', () => {
      const input = 'Switched to branch feat/observability-dash and merged fix/fix-3-clean into main.';
      const output = stripInternalIdentifiers(input);
      expect(output).not.toContain('feat/observability-dash');
      expect(output).not.toContain('fix/fix-3-clean');
      expect(output).not.toContain('branch');
    });

    it('returns empty string for empty or nullish text', () => {
      expect(stripInternalIdentifiers('')).toBe('');
    });
  });

  describe('translateRuleToPlainEnglish', () => {
    it('translates TOOL_RETRY_STREAK and ConsecutiveToolErrors with and without tool', () => {
      expect(translateRuleToPlainEnglish('TOOL_RETRY_STREAK', 'bash')).toBe(
        'Consecutive tool failures on `bash`'
      );
      expect(translateRuleToPlainEnglish('ConsecutiveToolErrors', 'read')).toBe(
        'Consecutive tool failures on `read`'
      );
      expect(translateRuleToPlainEnglish('TOOL_RETRY_STREAK')).toBe(
        'Consecutive tool failure streak'
      );
      expect(translateRuleToPlainEnglish('ConsecutiveToolErrors')).toBe(
        'Consecutive tool failure streak'
      );
    });

    it('translates TOOL_HIGH_FAILURE_RATE with and without tool', () => {
      expect(translateRuleToPlainEnglish('TOOL_HIGH_FAILURE_RATE', 'curl')).toBe(
        'High failure rate on `curl`'
      );
      expect(translateRuleToPlainEnglish('TOOL_HIGH_FAILURE_RATE')).toBe(
        'High tool failure rate'
      );
    });

    it('translates CONTEXT_SATURATION_SPIKE and ContextCeilingApproach', () => {
      expect(translateRuleToPlainEnglish('CONTEXT_SATURATION_SPIKE')).toBe(
        'Context window saturation approach'
      );
      expect(translateRuleToPlainEnglish('ContextCeilingApproach')).toBe(
        'Context window saturation approach'
      );
    });

    it('translates SUBAGENT_REWORK_LOOP and ToolThrashing with and without tool', () => {
      expect(translateRuleToPlainEnglish('SUBAGENT_REWORK_LOOP', 'edit')).toBe(
        'Repeated tool thrashing on `edit`'
      );
      expect(translateRuleToPlainEnglish('ToolThrashing', 'bash')).toBe(
        'Repeated tool thrashing on `bash`'
      );
      expect(translateRuleToPlainEnglish('SUBAGENT_REWORK_LOOP')).toBe(
        'Repeated tool thrashing'
      );
    });

    it('translates TokenVelocityRunaway', () => {
      expect(translateRuleToPlainEnglish('TokenVelocityRunaway')).toBe(
        'Rapid token consumption velocity'
      );
    });

    it('falls back to cleaned name for unknown rule IDs', () => {
      expect(translateRuleToPlainEnglish('UNKNOWN_ANOMALY_RULE')).toBe(
        'UNKNOWN ANOMALY RULE'
      );
    });
  });

  describe('sanitizeTextForPrompt', () => {
    it('strips internal severity tags and brackets', () => {
      const input = '[CRITICAL] Database connection failed [WARNING] [HIGH] SeverityHigh severity: high';
      const output = sanitizeTextForPrompt(input);
      expect(output).not.toContain('[CRITICAL]');
      expect(output).not.toContain('[WARNING]');
      expect(output).not.toContain('[HIGH]');
      expect(output).not.toContain('SeverityHigh');
      expect(output).not.toContain('severity: high');
    });

    it('translates embedded rule identifiers and strips ticket IDs in a single pass', () => {
      const input = '[HIGH] TOOL_RETRY_STREAK detected in ticket FIX-3 branch fix/fix-3-loop.';
      const output = sanitizeTextForPrompt(input);
      expect(output).toContain('Consecutive tool failures');
      expect(output).not.toContain('TOOL_RETRY_STREAK');
      expect(output).not.toContain('FIX-3');
      expect(output).not.toContain('fix/fix-3-loop');
      expect(output).not.toContain('[HIGH]');
    });

    it('handles empty string gracefully', () => {
      expect(sanitizeTextForPrompt('')).toBe('');
    });
  });

  describe('generateToolExecutionPrompt (Layer 1)', () => {
    it('generates prompt for tool execution failure with consecutive errors', () => {
      const prompt = generateToolExecutionPrompt({
        toolName: 'bash',
        turn: 4,
        sessionId: 'ses-12345',
        messageId: 'msg-67890',
        partId: 'part-abcde',
        status: 'error',
        consecutiveErrors: 4,
        input: 'npm test',
        output: 'FAIL src/app.test.ts',
      });

      expect(prompt).toContain('# Tool Execution Failure: `bash`');
      expect(prompt).toContain(
        'During turn 4 of session `ses-12345` (message `msg-67890`, part `part-abcde`), the agent executed `bash` and it failed with 4 consecutive failures on this tool.'
      );
      expect(prompt).toContain('### Input');
      expect(prompt).toContain('npm test');
      expect(prompt).toContain('### Output');
      expect(prompt).toContain('FAIL src/app.test.ts');
      expect(prompt).toContain('### What to check');
      expect(prompt).toContain('- Why did this tool call fail or produce unexpected output?');
      expect(prompt).toContain(
        '- Is this a genuine environment issue, invalid arguments from the agent, or a misclassified success?'
      );
      expect(prompt).toContain(
        '- What exact action or command should the agent execute next to recover?'
      );
    });

    it('generates prompt for completed tool execution with success outcome', () => {
      const prompt = generateToolExecutionPrompt({
        toolName: 'read',
        turn: 2,
        sessionId: 'ses-12345',
        messageId: 'msg-67890',
        partId: 'part-abcde',
        status: 'completed',
        input: 'README.md',
        output: 'Contents of file',
      });

      expect(prompt).toContain('# Tool Execution: `read`');
      expect(prompt).toContain(
        'During turn 2 of session `ses-12345` (message `msg-67890`, part `part-abcde`), the agent executed `read` and it completed successfully.'
      );
      expect(prompt).toContain('README.md');
      expect(prompt).toContain('Contents of file');
    });

    it('generates prompt for error status without consecutive errors streak', () => {
      const prompt = generateToolExecutionPrompt({
        toolName: 'glob',
        turn: 1,
        sessionId: 'ses-12345',
        status: 'error',
        consecutiveErrors: 1,
        input: '**/*.ts',
        errorMessage: 'path does not exist',
      });

      expect(prompt).toContain('# Tool Execution Failure: `glob`');
      expect(prompt).toContain(
        'During turn 1 of session `ses-12345`, the agent executed `glob` and it failed with an execution error.'
      );
      expect(prompt).toContain('path does not exist');
    });

    it('handles provenance variations: only messageId, only partId, or neither', () => {
      const promptMsgOnly = generateToolExecutionPrompt({
        toolName: 'bash',
        sessionId: 'ses-1',
        messageId: 'msg-99',
        status: 'completed',
      });
      expect(promptMsgOnly).toContain('In session `ses-1` (message `msg-99`),');

      const promptPartOnly = generateToolExecutionPrompt({
        toolName: 'bash',
        sessionId: 'ses-1',
        partId: 'part-88',
        status: 'completed',
      });
      expect(promptPartOnly).toContain('In session `ses-1` (part `part-88`),');

      const promptNeither = generateToolExecutionPrompt({
        toolName: 'bash',
        sessionId: 'ses-1',
        status: 'completed',
      });
      expect(promptNeither).toContain('In session `ses-1`, the agent executed');
    });

    it('renders fallback text when input and output are empty', () => {
      const prompt = generateToolExecutionPrompt({
        toolName: 'read',
        turn: 1,
        sessionId: 'ses-1',
        status: 'completed',
        input: null,
        output: '',
      });

      expect(prompt).toContain('(No input payload recorded)');
      expect(prompt).toContain('(No output recorded)');
    });

    it('defensively handles content containing code fence delimiters', () => {
      const prompt = generateToolExecutionPrompt({
        toolName: 'bash',
        turn: 1,
        sessionId: 'ses-1',
        status: 'error',
        input: 'echo "```markdown\ntest\n```"',
      });

      expect(prompt).toContain('````');
    });
  });

  describe('generateExecutionIssuePrompt (Layer 2)', () => {
    it('generates prompt with full context: turn, role, model, partId, and errorDetails', () => {
      const prompt = generateExecutionIssuePrompt({
        summary: 'Consecutive failures on `bash`',
        turn: 5,
        sessionId: 'ses-alpha',
        role: 'coder',
        model: 'gemini-3.8-flash',
        description: 'Command failed repeatedly due to exit status 2',
        partId: 'part-err-99',
        errorDetails: 'gcc: error: no such file or directory',
      });

      expect(prompt).toContain('# Agent Execution Issue: Consecutive failures on `bash`');
      expect(prompt).toContain(
        'During turn 5 of session `ses-alpha` (agent `coder` running `gemini-3.8-flash`), command failed repeatedly due to exit status 2. The last failure was triggered by part `part-err-99`.'
      );
      expect(prompt).toContain('### Error Details');
      expect(prompt).toContain('gcc: error: no such file or directory');
      expect(prompt).toContain('### What to check');
      expect(prompt).toContain(
        '- What is causing this failure pattern (such as an infinite retry loop, incorrect assumptions about the filesystem, or misinterpreting file content as an error)?'
      );
      expect(prompt).toContain('- How should the agent break out of this loop and recover?');
    });

    it('handles omitted turn, model, or partId gracefully', () => {
      const prompt = generateExecutionIssuePrompt({
        summary: 'Repeated tool thrashing',
        sessionId: 'ses-beta',
        role: 'tester',
        description: 'Rapid alternating edits without verification',
      });

      expect(prompt).toContain(
        'In session `ses-beta` (agent `tester`), rapid alternating edits without verification.'
      );
      expect(prompt).not.toContain('The last failure was triggered by part');
      expect(prompt).toContain('(No additional error details recorded)');
    });

    it('sanitizes summary, description, and error details', () => {
      const prompt = generateExecutionIssuePrompt({
        summary: '[CRITICAL] TOOL_RETRY_STREAK on ticket FIX-3',
        turn: 2,
        sessionId: 'ses-gamma',
        description: 'Failed on branch feat/fix-loop for ticket AL-20 with SeverityHigh',
        errorDetails: 'Build error on branch fix/al20-tests in ticket GEN-42',
      });

      expect(prompt).not.toContain('[CRITICAL]');
      expect(prompt).not.toContain('FIX-3');
      expect(prompt).not.toContain('AL-20');
      expect(prompt).not.toContain('GEN-42');
      expect(prompt).not.toContain('feat/fix-loop');
      expect(prompt).not.toContain('fix/al20-tests');
      expect(prompt).not.toContain('SeverityHigh');
      expect(prompt).toContain('Consecutive tool failures');
    });
  });

  describe('generateMessageTurnPrompt (Layer 3)', () => {
    it('generates prompt with model, deliberation, message body, and executed tools', () => {
      const prompt = generateMessageTurnPrompt({
        turn: 3,
        role: 'assistant',
        model: 'gemini-3.8-flash',
        sessionId: 'ses-100',
        messageId: 'msg-200',
        reasoningContent: 'I need to check the files first.\nThen edit them.',
        messageContent: 'I will now inspect the source directory and update the config.',
        toolsExecuted: [
          { toolName: 'glob', status: 'completed', partId: 'part-g1' },
          { toolName: 'edit', status: 'error', partId: 'part-e2' },
        ],
      });

      expect(prompt).toContain('# Turn 3 Review (`assistant` / `gemini-3.8-flash`)');
      expect(prompt).toContain(
        'In turn 3 of session `ses-100` (message `msg-200`), the agent deliberated and executed tools.'
      );
      expect(prompt).toContain('### Model Deliberation');
      expect(prompt).toContain('> I need to check the files first.');
      expect(prompt).toContain('> Then edit them.');
      expect(prompt).toContain('### Message');
      expect(prompt).toContain(
        'I will now inspect the source directory and update the config.'
      );
      expect(prompt).toContain('### Tools Executed');
      expect(prompt).toContain('- `glob` (`part-g1`): completed');
      expect(prompt).toContain('- `edit` (`part-e2`): error');
      expect(prompt).toContain('### What to check');
      expect(prompt).toContain(
        "- Compare the agent's stated plan with the actual tool outcomes in this turn."
      );
    });

    it('handles turn without model or messageId', () => {
      const prompt = generateMessageTurnPrompt({
        turn: 1,
        role: 'user',
        sessionId: 'ses-user',
        messageContent: 'Please start implementation.',
      });

      expect(prompt).toContain('# Turn 1 Review (`user`)');
      expect(prompt).toContain('In turn 1 of session `ses-user`, the agent deliberated and executed tools.');
      expect(prompt).toContain('(No deliberation content recorded)');
      expect(prompt).toContain('Please start implementation.');
      expect(prompt).toContain('(No tools executed in this turn)');
    });

    it('formats tool items without partId', () => {
      const prompt = generateMessageTurnPrompt({
        turn: 1,
        role: 'assistant',
        sessionId: 'ses-tool-test',
        toolsExecuted: [
          { toolName: 'bash', status: 'completed' },
        ],
      });

      expect(prompt).toContain('- `bash`: completed');
    });
  });

  describe('generateSessionOverviewPrompt (Layer 4)', () => {
    it('generates prompt with role, model, parent session, turn and tool counts, and flagged issues', () => {
      const prompt = generateSessionOverviewPrompt({
        sessionId: 'ses-sub-1',
        parentId: 'ses-root-0',
        role: 'coder',
        model: 'gemini-3.8-flash',
        status: 'running',
        turnCount: 8,
        totalToolCalls: 15,
        failedToolCalls: 3,
        flaggedIssues: [
          'Consecutive tool failures on `bash`',
          'Context window saturation approach',
        ],
      });

      expect(prompt).toContain('# Session Overview: `ses-sub-1`');
      expect(prompt).toContain(
        'Session `ses-sub-1` is a `coder` subagent running `gemini-3.8-flash` (parent session: `ses-root-0`). The session is currently **running** after 8 turns and 15 tool executions (3 failed).'
      );
      expect(prompt).toContain('### Flagged Issues');
      expect(prompt).toContain('- Consecutive tool failures on `bash`');
      expect(prompt).toContain('- Context window saturation approach');
      expect(prompt).toContain('### What to check');
      expect(prompt).toContain('- Why did this session stall or accumulate repeated tool failures?');
      expect(prompt).toContain(
        '- What guidance or state reset is needed to unblock and resume the agent?'
      );
    });

    it('handles session without parentId and with zero flagged issues', () => {
      const prompt = generateSessionOverviewPrompt({
        sessionId: 'ses-clean-root',
        role: 'orchestrator',
        model: 'gemini-3.8-flash',
        status: 'completed',
        turnCount: 4,
        totalToolCalls: 6,
        failedToolCalls: 0,
        flaggedIssues: [],
      });

      expect(prompt).toContain(
        'Session `ses-clean-root` is a `orchestrator` subagent running `gemini-3.8-flash`. The session is currently **completed** after 4 turns and 6 tool executions (0 failed).'
      );
      expect(prompt).toContain('No flagged issues or anomalies recorded for this session.');
      expect(prompt).not.toContain('parent session:');
    });

    it('sanitizes flagged issues removing ticket and branch references', () => {
      const prompt = generateSessionOverviewPrompt({
        sessionId: 'ses-sanitized',
        role: 'tester',
        model: 'gemini-3.8-flash',
        status: 'error',
        turnCount: 2,
        totalToolCalls: 4,
        failedToolCalls: 4,
        flaggedIssues: [
          '[CRITICAL] Failure in ticket AL-20 on branch fix/al20-run',
        ],
      });

      expect(prompt).not.toContain('[CRITICAL]');
      expect(prompt).not.toContain('AL-20');
      expect(prompt).not.toContain('fix/al20-run');
    });
  });

  describe('Adversarial Edge Cases & Boundaries', () => {
    describe('Adversarial Sanitization', () => {
      it('cleanly strips multiple ticket IDs without leaving orphaned commas or conjunctions', () => {
        const input = 'FIX-3, AL-42, GEN-10 and PO-5';
        expect(sanitizeTextForPrompt(input)).toBe('');
      });

      it('cleanly strips multiple ticket IDs embedded in descriptive sentences', () => {
        const input = 'Fixing failures across tickets FIX-3, AL-42, GEN-10, and PO-5 in the core subsystem.';
        const output = sanitizeTextForPrompt(input);
        expect(output).not.toContain('FIX-3');
        expect(output).not.toContain('AL-42');
        expect(output).not.toContain('GEN-10');
        expect(output).not.toContain('PO-5');
        expect(output).toContain('Fixing failures across in the core subsystem.');
      });

      it('cleanly strips branch names like feat/fix-something-branch-123 and refs/heads patterns', () => {
        const input = 'Deployed from branch feat/fix-something-branch-123 to refs/heads/fix/something-else.';
        const output = sanitizeTextForPrompt(input);
        expect(output).not.toContain('feat/fix-something-branch-123');
        expect(output).not.toContain('fix/something-else');
        expect(output).not.toContain('refs/heads');
      });

      it('cleanly strips multiple severity tags without leaving commas or brackets', () => {
        const input = '[CRITICAL], [HIGH], [WARNING]';
        expect(sanitizeTextForPrompt(input)).toBe('');
      });

      it('translates multiple rule IDs while stripping tickets, branches, and severity tags in a single pass', () => {
        const input =
          '[CRITICAL] TOOL_RETRY_STREAK detected in ticket FIX-3 on branch feat/fix-something-branch-123 accompanied by [HIGH] CONTEXT_SATURATION_SPIKE.';
        const output = sanitizeTextForPrompt(input);
        expect(output).not.toContain('TOOL_RETRY_STREAK');
        expect(output).not.toContain('CONTEXT_SATURATION_SPIKE');
        expect(output).not.toContain('FIX-3');
        expect(output).not.toContain('feat/fix-something-branch-123');
        expect(output).not.toContain('[CRITICAL]');
        expect(output).not.toContain('[HIGH]');
        expect(output).toContain('Consecutive tool failures');
        expect(output).toContain('Context window saturation approach');
      });

      it('handles null, undefined, empty string, and whitespace gracefully', () => {
        expect(sanitizeTextForPrompt('')).toBe('');
        expect(sanitizeTextForPrompt('   \t\n  ')).toBe('');
        expect(stripInternalIdentifiers('')).toBe('');
        expect(stripInternalIdentifiers('   ')).toBe('');
      });
    });

    describe('Adversarial Layer 1: generateToolExecutionPrompt', () => {
      it('handles null/empty inputs, outputs, messages, and error messages', () => {
        const prompt = generateToolExecutionPrompt({
          toolName: '',
          sessionId: 'ses-edge',
          status: 'completed',
          input: null,
          output: undefined,
          errorMessage: null,
        });

        expect(prompt).toContain('# Tool Execution: `unknown_tool`');
        expect(prompt).toContain('In session `ses-edge`, the agent executed `unknown_tool` and it completed successfully.');
        expect(prompt).toContain('(No input payload recorded)');
        expect(prompt).toContain('(No output recorded)');
      });

      it('uses errorMessage when output is absent or empty', () => {
        const prompt = generateToolExecutionPrompt({
          toolName: 'bash',
          sessionId: 'ses-edge',
          status: 'error',
          input: 'ls /missing',
          output: '',
          errorMessage: 'ls: cannot access /missing: No such file or directory',
        });

        expect(prompt).toContain('# Tool Execution Failure: `bash`');
        expect(prompt).toContain('ls: cannot access /missing: No such file or directory');
      });

      it('handles turn 0 correctly without falling back to omitting turn', () => {
        const prompt = generateToolExecutionPrompt({
          toolName: 'read',
          turn: 0,
          sessionId: 'ses-turn0',
          status: 'completed',
          input: 'config.json',
        });

        expect(prompt).toContain('During turn 0 of session `ses-turn0`');
      });

      it('handles consecutive failures edge cases: 0, 1, and values greater than 1', () => {
        // consecutiveErrors = 0 with error status
        const prompt0 = generateToolExecutionPrompt({
          toolName: 'bash',
          sessionId: 'ses-c0',
          status: 'error',
          consecutiveErrors: 0,
        });
        expect(prompt0).toContain('failed with an execution error');
        expect(prompt0).not.toContain('0 consecutive failures');

        // consecutiveErrors = 1 with error status
        const prompt1 = generateToolExecutionPrompt({
          toolName: 'bash',
          sessionId: 'ses-c1',
          status: 'error',
          consecutiveErrors: 1,
        });
        expect(prompt1).toContain('failed with an execution error');
        expect(prompt1).not.toContain('1 consecutive failures');

        // consecutiveErrors = 7 with error status
        const prompt7 = generateToolExecutionPrompt({
          toolName: 'bash',
          sessionId: 'ses-c7',
          status: 'error',
          consecutiveErrors: 7,
        });
        expect(prompt7).toContain('failed with 7 consecutive failures on this tool');

        // consecutiveErrors > 1 with completed status (should ignore consecutiveErrors for success)
        const promptSuccess = generateToolExecutionPrompt({
          toolName: 'bash',
          sessionId: 'ses-ok',
          status: 'completed',
          consecutiveErrors: 3,
        });
        expect(promptSuccess).toContain('completed successfully');
      });

      it('defensively handles deeply nested markdown fences within input and output', () => {
        const nestedInput = '```typescript\nconst x = "```";\nconsole.log(x);\n```';
        const nestedOutput = '````bash\necho "nested 4 ticks"\n````';

        const prompt = generateToolExecutionPrompt({
          toolName: 'edit',
          turn: 1,
          sessionId: 'ses-fences',
          status: 'completed',
          input: nestedInput,
          output: nestedOutput,
        });

        // 3 backticks in input must be wrapped in at least 4 backticks
        expect(prompt).toContain('````\n```typescript');
        // 4 backticks in output must be wrapped in at least 5 backticks
        expect(prompt).toContain('`````\n````bash');
      });

      it('sanitizes multiple ticket IDs and branches embedded within tool input and output', () => {
        const prompt = generateToolExecutionPrompt({
          toolName: 'bash',
          sessionId: 'ses-sec',
          status: 'error',
          input: 'git checkout feat/fix-something-branch-123 && echo "ticket FIX-3, AL-42, GEN-10 and PO-5"',
          output: '[CRITICAL] Error on branch fix/sub-loop: tickets FIX-3 and AL-42 failed',
        });

        expect(prompt).not.toContain('feat/fix-something-branch-123');
        expect(prompt).not.toContain('fix/sub-loop');
        expect(prompt).not.toContain('FIX-3');
        expect(prompt).not.toContain('AL-42');
        expect(prompt).not.toContain('GEN-10');
        expect(prompt).not.toContain('PO-5');
        expect(prompt).not.toContain('[CRITICAL]');
      });
    });

    describe('Adversarial Layer 2: generateExecutionIssuePrompt', () => {
      it('falls back to generic title when summary consists entirely of ticket IDs or tags', () => {
        const prompt = generateExecutionIssuePrompt({
          summary: 'FIX-3, AL-42, GEN-10 and PO-5',
          sessionId: 'ses-sum-strip',
          description: 'A tool failure occurred.',
        });

        expect(prompt).toContain('# Agent Execution Issue: Operational Execution Issue');
      });

      it('handles turn 0 and omitted role/model/partId', () => {
        const prompt = generateExecutionIssuePrompt({
          summary: 'Tool failure',
          turn: 0,
          sessionId: 'ses-turn0',
          description: '',
          partId: null,
          errorDetails: null,
        });

        expect(prompt).toContain('During turn 0 of session `ses-turn0`, an execution issue occurred.');
        expect(prompt).not.toContain('(agent');
        expect(prompt).not.toContain('The last failure was triggered by part');
        expect(prompt).toContain('(No additional error details recorded)');
      });

      it('defensively wraps markdown code fences inside errorDetails', () => {
        const prompt = generateExecutionIssuePrompt({
          summary: 'Compilation error',
          sessionId: 'ses-err-fence',
          description: 'Compilation aborted',
          errorDetails: '```go\nfunc main() { panic("fail") }\n```',
        });

        expect(prompt).toContain('````\n```go\nfunc main() { panic("fail") }\n```\n````');
      });
    });

    describe('Adversarial Layer 3: generateMessageTurnPrompt', () => {
      it('handles null/empty deliberations, reasoning tokens, and text body', () => {
        const prompt = generateMessageTurnPrompt({
          turn: 0,
          role: '',
          sessionId: 'ses-empty-turn',
          reasoningContent: null,
          messageContent: '   ',
          toolsExecuted: [],
        });

        expect(prompt).toContain('# Turn 0 Review (`assistant`)');
        expect(prompt).toContain('In turn 0 of session `ses-empty-turn`, the agent deliberated and executed tools.');
        expect(prompt).toContain('(No deliberation content recorded)');
        expect(prompt).toContain('(No text message body recorded)');
        expect(prompt).toContain('(No tools executed in this turn)');
      });

      it('formats multiline deliberation reasoning with blockquote markdown markers', () => {
        const prompt = generateMessageTurnPrompt({
          turn: 1,
          role: 'coder',
          model: 'gemini-3.8-flash',
          sessionId: 'ses-delib',
          reasoningContent: 'First check inputs.\nSecond run tests.\nThird submit.',
        });

        expect(prompt).toContain('> First check inputs.');
        expect(prompt).toContain('> Second run tests.');
        expect(prompt).toContain('> Third submit.');
      });

      it('handles toolsExecuted entries with missing fields or status', () => {
        const prompt = generateMessageTurnPrompt({
          turn: 1,
          role: 'assistant',
          sessionId: 'ses-tools-edge',
          toolsExecuted: [
            { toolName: '', status: '', partId: null },
            { toolName: 'bash', status: 'error', partId: '   ' },
          ],
        });

        expect(prompt).toContain('- `unknown_tool`: unknown');
        expect(prompt).toContain('- `bash`: error');
      });
    });

    describe('Adversarial Layer 4: generateSessionOverviewPrompt', () => {
      it('handles missing parentId (null, empty, whitespace) without parent session note', () => {
        const promptNull = generateSessionOverviewPrompt({
          sessionId: 'ses-root-null',
          parentId: null,
          role: 'coder',
          model: 'gemini-3.8-flash',
          status: 'running',
          turnCount: 0,
          totalToolCalls: 0,
          failedToolCalls: 0,
        });
        expect(promptNull).not.toContain('parent session:');

        const promptEmpty = generateSessionOverviewPrompt({
          sessionId: 'ses-root-empty',
          parentId: '   ',
          role: 'coder',
          model: 'gemini-3.8-flash',
          status: 'running',
          turnCount: 0,
          totalToolCalls: 0,
          failedToolCalls: 0,
        });
        expect(promptEmpty).not.toContain('parent session:');
      });

      it('handles turnCount 0, totalToolCalls 0, and failedToolCalls 0 correctly', () => {
        const prompt = generateSessionOverviewPrompt({
          sessionId: 'ses-zero-counts',
          role: 'planner',
          model: 'gemini-3.8-flash',
          status: 'initialized',
          turnCount: 0,
          totalToolCalls: 0,
          failedToolCalls: 0,
        });

        expect(prompt).toContain(
          'Session `ses-zero-counts` is a `planner` subagent running `gemini-3.8-flash`. The session is currently **initialized** after 0 turns and 0 tool executions (0 failed).'
        );
      });

      it('falls back to no flagged issues when flaggedIssues contains only stripped tickets/tags', () => {
        const prompt = generateSessionOverviewPrompt({
          sessionId: 'ses-issues-strip',
          role: 'tester',
          model: 'gemini-3.8-flash',
          status: 'completed',
          turnCount: 1,
          totalToolCalls: 1,
          failedToolCalls: 0,
          flaggedIssues: ['FIX-3, AL-42, GEN-10 and PO-5', '[CRITICAL], [HIGH], [WARNING]'],
        });

        expect(prompt).toContain('No flagged issues or anomalies recorded for this session.');
      });
    });
  });
});
