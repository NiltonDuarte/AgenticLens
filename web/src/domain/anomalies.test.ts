import { describe, it, expect } from 'vitest';
import type { EventFact, Anomaly } from '../types/telemetry';
import type { LivePartFact, LiveAnomalyAlert } from './live';
import {
  isToolFailureAnomalyRule,
  matchesAnomaly,
  findMatchingAnomaly,
  extractStreakCount,
  findMatchingLiveAlert,
} from './anomalies';

describe('Domain Anomalies Logic', () => {
  describe('isToolFailureAnomalyRule', () => {
    it('returns true for known tool failure anomaly rule IDs', () => {
      expect(isToolFailureAnomalyRule('TOOL_RETRY_STREAK')).toBe(true);
      expect(isToolFailureAnomalyRule('TOOL_HIGH_FAILURE_RATE')).toBe(true);
      expect(isToolFailureAnomalyRule('ConsecutiveToolErrors')).toBe(true);
      expect(isToolFailureAnomalyRule('ToolThrashing')).toBe(true);
      expect(isToolFailureAnomalyRule('TOOL_TIMEOUT')).toBe(true);
      expect(isToolFailureAnomalyRule('CUSTOM_TOOL_FAILURE_RULE')).toBe(true);
    });

    it('returns false for non-tool anomaly rule IDs', () => {
      expect(isToolFailureAnomalyRule('')).toBe(false);
      expect(isToolFailureAnomalyRule('CONTEXT_SATURATION_SPIKE')).toBe(false);
      expect(isToolFailureAnomalyRule('SUBAGENT_REWORK_LOOP')).toBe(false);
      expect(isToolFailureAnomalyRule('ContextCeilingApproach')).toBe(false);
      expect(isToolFailureAnomalyRule('TokenVelocityRunaway')).toBe(false);
      expect(isToolFailureAnomalyRule('UNKNOWN_RULE')).toBe(false);
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
      title: 'Tool Retry Streak',
      description: 'Repeated bash failures',
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
        title: 'Context Saturation Spike',
        description: 'Large token burst',
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
        title: 'Subagent Rework Loop',
        description: 'Repeated handoffs',
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
        title: 'Spike',
        description: 'tokens',
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
        title: 'Streak',
        description: 'desc',
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
