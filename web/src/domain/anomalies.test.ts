import { describe, it, expect } from 'vitest';
import type { EventFact, Anomaly } from '../types/telemetry';
import type { LivePartFact, LiveAnomalyAlert } from './live';
import {
  isToolFailureAnomalyRule,
  matchesAnomaly,
  findMatchingAnomaly,
  extractStreakCount,
  findMatchingLiveAlert,
  formatAnomaly,
  type AnomalyRule,
  type ToolFailureAnomalyRule,
} from './anomalies';

describe('Domain Anomalies Logic', () => {
  describe('formatAnomaly', () => {
    it.each([
      [{ rule_id: 'TOOL_RETRY_STREAK', severity: 'WARNING', context: { tool_name: 'bash', count: 3 } }, 'Tool retry streak: bash (3 consecutive errors)', "Tool 'bash' encountered 3 consecutive execution failures."],
      [{ rule_id: 'TOOL_HIGH_FAILURE_RATE', severity: 'WARNING', context: { tool_name: 'git', count: 2, total_calls: 6, failure_rate_pct: 33.3, threshold: '>20%' } }, 'High tool failure rate: git (33.3%)', "Tool 'git' failed 2 out of 6 calls (33.3% failure rate, threshold: >20%)."],
      [{ rule_id: 'CONTEXT_SATURATION_SPIKE', severity: 'WARNING', context: { token_delta: 45000, context_tokens: 165000 } }, 'Context saturation spike (+45k tokens)', 'Context tokens surged by 45k to 165k tokens.'],
      [{ rule_id: 'CONTEXT_SATURATION_SPIKE', severity: 'CRITICAL', context: { context_tokens: 165000, threshold: '160000' } }, 'Context ceiling breached (165k tokens)', 'Context tokens reached 165k, breaching the 160k critical ceiling.'],
      [{ rule_id: 'SUBAGENT_REWORK_LOOP', severity: 'WARNING', context: { role: 'coder', secondary_role: 'tester', count: 4 } }, 'Subagent rework loop: coder <-> tester (4 cycles)', "Subagent roles 'coder' and 'tester' alternated for 4 consecutive cycles."],
    ] as const)('formats rule context', (anomaly, title, description) => {
      expect(formatAnomaly(anomaly)).toEqual({ title, description });
    });

    it('uses a safe fallback for incomplete context', () => {
      expect(formatAnomaly({ rule_id: 'TOOL_RETRY_STREAK', severity: 'WARNING', context: {} })).toEqual({
        title: 'Tool retry streak',
        description: 'tool retry streak detected; some diagnostic context is unavailable.',
      });
    });
  });
  describe('isToolFailureAnomalyRule', () => {
    it('returns true for all four tool failure anomaly rule IDs', () => {
      const canonicalFailures: readonly string[] = [
        'TOOL_RETRY_STREAK',
        'TOOL_HIGH_FAILURE_RATE',
        'ConsecutiveToolErrors',
        'ToolThrashing',
      ];

      for (const rule of canonicalFailures) {
        const candidate: AnomalyRule | string = rule;
        if (isToolFailureAnomalyRule(candidate)) {
          const narrowed: ToolFailureAnomalyRule = candidate;
          expect(narrowed).toBe(rule);
        } else {
          throw new Error(`Expected candidate ${rule} to be narrowed to ToolFailureAnomalyRule`);
        }
      }
    });

    it('returns false for non-tool anomaly rule IDs', () => {
      expect(isToolFailureAnomalyRule('CONTEXT_SATURATION_SPIKE')).toBe(false);
      expect(isToolFailureAnomalyRule('SUBAGENT_REWORK_LOOP')).toBe(false);
      expect(isToolFailureAnomalyRule('ContextCeilingApproach')).toBe(false);
      expect(isToolFailureAnomalyRule('TokenVelocityRunaway')).toBe(false);
    });

    it('returns false for adversarial inputs including empty strings, random strings, prefix spoofing, substrings, and lowercased names', () => {
      // Empty and whitespace strings
      expect(isToolFailureAnomalyRule('')).toBe(false);
      expect(isToolFailureAnomalyRule('   ')).toBe(false);
      expect(isToolFailureAnomalyRule('\t\n')).toBe(false);

      // Random strings and identifiers
      expect(isToolFailureAnomalyRule('UNKNOWN_RULE')).toBe(false);
      expect(isToolFailureAnomalyRule('foobar')).toBe(false);
      expect(isToolFailureAnomalyRule('12345')).toBe(false);
      expect(isToolFailureAnomalyRule('!@#$%^&*()')).toBe(false);
      expect(isToolFailureAnomalyRule('true')).toBe(false);
      expect(isToolFailureAnomalyRule('null')).toBe(false);
      expect(isToolFailureAnomalyRule('undefined')).toBe(false);

      // Prefix spoofing (previously matched by startsWith('TOOL_'))
      expect(isToolFailureAnomalyRule('TOOL_')).toBe(false);
      expect(isToolFailureAnomalyRule('TOOL_SUCCESS')).toBe(false);
      expect(isToolFailureAnomalyRule('TOOL_INFO')).toBe(false);
      expect(isToolFailureAnomalyRule('TOOL_FAILURE_FOO')).toBe(false);
      expect(isToolFailureAnomalyRule('TOOL_ANYTHING')).toBe(false);
      expect(isToolFailureAnomalyRule('TOOL_TIMEOUT')).toBe(false);
      expect(isToolFailureAnomalyRule('TOOL_RETRY')).toBe(false);
      expect(isToolFailureAnomalyRule('TOOL_HIGH_FAILURE')).toBe(false);
      expect(isToolFailureAnomalyRule('CUSTOM_TOOL_FAILURE_RULE')).toBe(false);
      expect(isToolFailureAnomalyRule('ConsecutiveTool')).toBe(false);
      expect(isToolFailureAnomalyRule('Tool')).toBe(false);

      // Substrings of valid rules
      expect(isToolFailureAnomalyRule('RETRY_STREAK')).toBe(false);
      expect(isToolFailureAnomalyRule('HIGH_FAILURE_RATE')).toBe(false);
      expect(isToolFailureAnomalyRule('Errors')).toBe(false);
      expect(isToolFailureAnomalyRule('Thrashing')).toBe(false);

      // Lowercased or mixed-case rule names
      expect(isToolFailureAnomalyRule('tool_retry_streak')).toBe(false);
      expect(isToolFailureAnomalyRule('tool_high_failure_rate')).toBe(false);
      expect(isToolFailureAnomalyRule('consecutivetoolerrors')).toBe(false);
      expect(isToolFailureAnomalyRule('toolthrashing')).toBe(false);
      expect(isToolFailureAnomalyRule('Tool_Retry_Streak')).toBe(false);

      // Suffix spoofing and extensions
      expect(isToolFailureAnomalyRule('TOOL_RETRY_STREAK_EXTRA')).toBe(false);
      expect(isToolFailureAnomalyRule('ToolThrashing2')).toBe(false);
      expect(isToolFailureAnomalyRule(' TOOL_RETRY_STREAK ')).toBe(false);
      expect(isToolFailureAnomalyRule('ConsecutiveToolErrors\n')).toBe(false);
    });

    it('strictly enforces LiveAnomalyAlert.rule_name type safety', () => {
      const validAlert: LiveAnomalyAlert = {
        session_id: 'ses-1',
        rule_name: 'ConsecutiveToolErrors',
        severity: 'WARNING',
        message: 'consecutive errors',
        trigger_event: 'p1',
        timestamp: '2026-10-21T10:00:00Z',
      };
      expect(validAlert.rule_name).toBe('ConsecutiveToolErrors');

      // Compile-time type check: arbitrary string rejected
      const invalidAlertString: LiveAnomalyAlert = {
        session_id: 'ses-1',
        // @ts-expect-error - rule_name strictly accepts LiveAnomalyRule, not arbitrary string
        rule_name: 'ArbitraryRuleString',
        severity: 'WARNING',
        message: 'arbitrary',
        trigger_event: 'p1',
        timestamp: '2026-10-21T10:00:00Z',
      };
      expect(invalidAlertString.session_id).toBe('ses-1');

      // Compile-time type check: domain AnomalyRuleID rejected in LiveAnomalyAlert
      const invalidAlertDomainRule: LiveAnomalyAlert = {
        session_id: 'ses-1',
        // @ts-expect-error - rule_name strictly accepts LiveAnomalyRule, not AnomalyRuleID
        rule_name: 'TOOL_RETRY_STREAK',
        severity: 'WARNING',
        message: 'domain rule in live alert',
        trigger_event: 'p1',
        timestamp: '2026-10-21T10:00:00Z',
      };
      expect(invalidAlertDomainRule.session_id).toBe('ses-1');
    });
  });

  describe('matchesAnomaly and findMatchingAnomaly', () => {
    const baseEvent: EventFact = {
      timestamp: '2026-10-21T10:00:00Z',
      session_id: 'ses-1',
      turn: 2,
      role: 'coder',
      event_type: 'tool_call',
      tool_name: 'bash',
      status: 'error',
      context_tokens: 5000,
    };

    const streakAnomaly: Anomaly = {
      rule_id: 'TOOL_RETRY_STREAK',
      severity: 'WARNING',
      context: {
        session_id: 'ses-1',
        tool_name: 'bash',
        turn: 2,
        count: 3,
      },
    };

    it('matches an error event with matching session, tool name, and turn', () => {
      expect(matchesAnomaly(baseEvent, streakAnomaly)).toBe(true);
      expect(findMatchingAnomaly(baseEvent, [streakAnomaly])).toEqual(streakAnomaly);
    });

    it('matches case-insensitively on tool name', () => {
      const upperToolEvent: EventFact = {
        ...baseEvent,
        tool_name: 'BASH',
      };
      expect(matchesAnomaly(upperToolEvent, streakAnomaly)).toBe(true);
    });

    it('uses exact evidence refs for persisted anomalies instead of matching every same-tool failure', () => {
      const evidencedAnomaly: Anomaly = {
        ...streakAnomaly,
        evidence: [{ session_id: 'ses-1', event_id: 'failed-call-2' }],
      };
      const referencedEvent: EventFact = { ...baseEvent, event_id: 'failed-call-2' };
      const unrelatedFailure: EventFact = { ...baseEvent, event_id: 'failed-call-9' };

      expect(matchesAnomaly(referencedEvent, evidencedAnomaly)).toBe(true);
      expect(matchesAnomaly(unrelatedFailure, evidencedAnomaly)).toBe(false);
      expect(findMatchingAnomaly(unrelatedFailure, [evidencedAnomaly])).toBeUndefined();
    });

    it('matches a nested event without session_id when its evidence ID has one owner', () => {
      const evidencedAnomaly: Anomaly = {
        ...streakAnomaly,
        evidence: [{ session_id: 'ses-1', event_id: 'failed-call-2' }],
      };
      const nestedEvent: Omit<EventFact, 'session_id'> = {
        event_id: 'failed-call-2', timestamp: baseEvent.timestamp, turn: baseEvent.turn,
        role: baseEvent.role, event_type: baseEvent.event_type, tool_name: baseEvent.tool_name,
        status: baseEvent.status, context_tokens: baseEvent.context_tokens,
      };
      expect(matchesAnomaly(nestedEvent, evidencedAnomaly)).toBe(true);
    });

    it('rejects a nested event when the same evidence ID appears under multiple sessions', () => {
      const evidencedAnomaly: Anomaly = {
        ...streakAnomaly,
        evidence: [
          { session_id: 'ses-1', event_id: 'call-1' },
          { session_id: 'ses-2', event_id: 'call-1' },
        ],
      };
      const nestedEvent: Omit<EventFact, 'session_id'> = {
        event_id: 'call-1', timestamp: baseEvent.timestamp, turn: baseEvent.turn,
        role: baseEvent.role, event_type: baseEvent.event_type, tool_name: baseEvent.tool_name,
        status: baseEvent.status, context_tokens: baseEvent.context_tokens,
      };
      expect(matchesAnomaly(nestedEvent, evidencedAnomaly)).toBe(false);
    });

    it('does NOT match tool failure anomaly if event status is completed', () => {
      const completedEvent: EventFact = {
        ...baseEvent,
        status: 'completed',
      };
      expect(matchesAnomaly(completedEvent, streakAnomaly)).toBe(false);
      expect(findMatchingAnomaly(completedEvent, [streakAnomaly])).toBeUndefined();
    });

    it('does NOT match tool failure anomaly if event status is non-error (e.g. pending or null)', () => {
      const otherStatusEvent: EventFact = {
        ...baseEvent,
        status: 'pending',
      };
      expect(matchesAnomaly(otherStatusEvent, streakAnomaly)).toBe(false);
    });

    it('allows non-tool failure anomaly to match completed event on matching turn and session', () => {
      const saturationAnomaly: Anomaly = {
        rule_id: 'CONTEXT_SATURATION_SPIKE',
        severity: 'WARNING',
        context: {
          session_id: 'ses-1',
          turn: 2,
        },
      };

      const completedEvent: EventFact = {
        ...baseEvent,
        status: 'completed',
      };
      expect(matchesAnomaly(completedEvent, saturationAnomaly)).toBe(true);
    });

    it('does NOT match if session_id differs', () => {
      const otherSessionEvent: EventFact = {
        ...baseEvent,
        session_id: 'ses-different',
      };
      expect(matchesAnomaly(otherSessionEvent, streakAnomaly)).toBe(false);
    });

    it('does NOT match if turn differs when context specifies turn', () => {
      const differentTurnEvent: EventFact = {
        ...baseEvent,
        turn: 5,
      };
      expect(matchesAnomaly(differentTurnEvent, streakAnomaly)).toBe(false);
    });

    it('matches tool call across turns if context omits turn', () => {
      const turnFreeAnomaly: Anomaly = {
        ...streakAnomaly,
        context: {
          session_id: 'ses-1',
          tool_name: 'bash',
        },
      };
      const anyTurnEvent: EventFact = {
        ...baseEvent,
        turn: 99,
      };
      expect(matchesAnomaly(anyTurnEvent, turnFreeAnomaly)).toBe(true);
    });

    it('matches role-based anomaly across turn', () => {
      const reworkAnomaly: Anomaly = {
        rule_id: 'SUBAGENT_REWORK_LOOP',
        severity: 'WARNING',
        context: {
          session_id: 'ses-1',
          role: 'coder',
          turn: 2,
        },
      };

      expect(matchesAnomaly(baseEvent, reworkAnomaly)).toBe(true);
    });

    it('findMatchingAnomaly picks the first matching anomaly among candidates', () => {
      const secondAnomaly: Anomaly = {
        ...streakAnomaly,
        rule_id: 'TOOL_HIGH_FAILURE_RATE',
      };
      const result = findMatchingAnomaly(baseEvent, [streakAnomaly, secondAnomaly]);
      expect(result).toEqual(streakAnomaly);
    });

    it('findMatchingAnomaly returns undefined on empty or undefined anomaly slice', () => {
      expect(findMatchingAnomaly(baseEvent, [])).toBeUndefined();
      expect(findMatchingAnomaly(baseEvent, undefined)).toBeUndefined();
    });
  });

  describe('extractStreakCount', () => {
    it('returns undefined for nil or missing count', () => {
      expect(extractStreakCount(undefined)).toBeUndefined();
      expect(extractStreakCount(null)).toBeUndefined();
      expect(extractStreakCount({
        rule_id: 'CONTEXT_SATURATION_SPIKE',
        severity: 'WARNING',
        context: {},
      })).toBeUndefined();
    });

    it('correctly preserves count 0 without dropping it as falsy', () => {
      const zeroCountAlert: LiveAnomalyAlert = {
        session_id: 'ses-1',
        rule_name: 'ConsecutiveToolErrors',
        severity: 'WARNING',
        message: 'zero count',
        trigger_event: 'p0',
        timestamp: '2026-10-21T10:00:00Z',
        count: 0,
      };
      expect(extractStreakCount(zeroCountAlert)).toBe(0);
    });

    it('extracts count from Anomaly context', () => {
      const anomaly: Anomaly = {
        rule_id: 'TOOL_RETRY_STREAK',
        severity: 'WARNING',
        context: { count: 4 },
      };
      expect(extractStreakCount(anomaly)).toBe(4);
    });

    it('extracts count from LiveAnomalyAlert', () => {
      const alert: LiveAnomalyAlert = {
        session_id: 'ses-1',
        rule_name: 'ConsecutiveToolErrors',
        severity: 'HIGH',
        message: '3 consecutive errors',
        trigger_event: 'p1',
        timestamp: '2026-10-21T10:00:00Z',
        count: 3,
      };
      expect(extractStreakCount(alert)).toBe(3);
    });
  });

  describe('findMatchingLiveAlert', () => {
    const errorPart: LivePartFact = {
      id: 'part-fail-99',
      type: 'tool',
      tool_name: 'bash',
      input_snippet: 'cargo build',
      output_snippet: 'error TS2322',
      status: 'error',
      is_error: true,
      severity: 'high',
    };

    const alertByPartId: LiveAnomalyAlert = {
      session_id: 'ses-1',
      rule_name: 'ConsecutiveToolErrors',
      severity: 'CRITICAL',
      message: 'Tool execution error streak reached 4 consecutive failures on bash',
      trigger_event: 'part-fail-99: error TS2322',
      timestamp: '2026-10-21T10:00:00Z',
      count: 4,
    };

    const alertByToolName: LiveAnomalyAlert = {
      session_id: 'ses-1',
      rule_name: 'ToolThrashing',
      severity: 'HIGH',
      message: 'Tool thrashing detected on tool bash with 3 failing invocations',
      trigger_event: 'bash: p1, p2, p3',
      timestamp: '2026-10-21T10:00:00Z',
      count: 3,
    };

    const nonToolAlert: LiveAnomalyAlert = {
      session_id: 'ses-1',
      rule_name: 'ContextCeilingApproach',
      severity: 'WARNING',
      message: 'Context ceiling approached: bash invoked in large context',
      trigger_event: 'tokens: 95%',
      timestamp: '2026-10-21T10:00:00Z',
    };

    it('finds alert by direct part ID in trigger_event', () => {
      const matched = findMatchingLiveAlert(errorPart, [alertByPartId]);
      expect(matched).toEqual(alertByPartId);
    });

    it('finds alert by tool failure rule and tool name match', () => {
      const otherPart: LivePartFact = {
        ...errorPart,
        id: 'different-part-id',
      };
      const matched = findMatchingLiveAlert(otherPart, [alertByToolName]);
      expect(matched).toEqual(alertByToolName);
    });

    it('does NOT match non-tool failure alerts even if tool name appears in message', () => {
      const otherPart: LivePartFact = {
        ...errorPart,
        id: 'different-part-id',
      };
      const matched = findMatchingLiveAlert(otherPart, [nonToolAlert]);
      expect(matched).toBeUndefined();
    });

    it('returns undefined if no alerts match', () => {
      const readPart: LivePartFact = {
        ...errorPart,
        tool_name: 'read',
        id: 'read-part-1',
      };
      const matched = findMatchingLiveAlert(readPart, [alertByToolName]);
      expect(matched).toBeUndefined();
    });

    it('returns undefined when alerts list is undefined or empty', () => {
      expect(findMatchingLiveAlert(errorPart, undefined)).toBeUndefined();
      expect(findMatchingLiveAlert(errorPart, [])).toBeUndefined();
    });
  });
});
