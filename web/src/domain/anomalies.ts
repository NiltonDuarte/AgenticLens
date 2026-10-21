import type { EventFact, Anomaly } from '../types/telemetry';
import type { LivePartFact, LiveAnomalyAlert } from './live';

/**
 * Checks whether an anomaly rule identifier matches tool failure anomaly patterns
 * (retries, high failure rates, consecutive errors, thrashing).
 */
export function isToolFailureAnomalyRule(ruleId: string): boolean {
  return (
    ruleId === 'TOOL_RETRY_STREAK' ||
    ruleId === 'TOOL_HIGH_FAILURE_RATE' ||
    ruleId === 'ConsecutiveToolErrors' ||
    ruleId === 'ToolThrashing' ||
    ruleId.startsWith('TOOL_') ||
    ruleId.includes('TOOL_FAILURE')
  );
}

/**
 * Evaluates whether an EventFact matches a detected Anomaly.
 * Enforces strict session identity, error status for tool failure rules, and turn/tool matching.
 */
export function matchesAnomaly(event: EventFact, anomaly: Anomaly): boolean {
  const ctx = anomaly.context;

  // Strict Session ID Guard: if both context and event specify session_id, they must match
  if (ctx.session_id && event.session_id && ctx.session_id !== event.session_id) {
    return false;
  }

  // Tool Failure Anomaly Status Criterion:
  // If anomaly is a tool failure rule, only match if the event is an error
  if (isToolFailureAnomalyRule(anomaly.rule_id)) {
    if (event.status !== 'error') {
      return false;
    }
  }

  // 1. Tool call match: tool_name matches
  if (
    ctx.tool_name &&
    event.tool_name &&
    ctx.tool_name.toLowerCase() === event.tool_name.toLowerCase()
  ) {
    if (ctx.turn !== undefined && event.turn !== undefined) {
      return ctx.turn === event.turn;
    }
    return true;
  }

  // 2. Turn and session match (e.g. context saturation spike)
  if (ctx.turn !== undefined && ctx.turn === event.turn) {
    if (ctx.session_id && event.session_id) {
      return ctx.session_id === event.session_id;
    }
    return true;
  }

  // 3. Role match
  if (
    ctx.role &&
    event.role &&
    (ctx.role.toLowerCase() === event.role.toLowerCase() ||
      ctx.secondary_role?.toLowerCase() === event.role.toLowerCase())
  ) {
    if (ctx.turn !== undefined) {
      return ctx.turn === event.turn;
    }
  }

  return false;
}

/**
 * Finds the first matching Anomaly for an EventFact from an optional list of anomalies.
 */
export function findMatchingAnomaly(
  event: EventFact,
  anomalies?: readonly Anomaly[]
): Anomaly | undefined {
  if (!anomalies || anomalies.length === 0) {
    return undefined;
  }

  return anomalies.find((a) => matchesAnomaly(event, a));
}

/**
 * Extracts numeric streak or failure count from an Anomaly or LiveAnomalyAlert.
 */
export function extractStreakCount(
  anomaly?: Anomaly | LiveAnomalyAlert | null | undefined
): number | undefined {
  if (!anomaly) {
    return undefined;
  }

  if ('count' in anomaly && typeof anomaly.count === 'number') {
    return anomaly.count;
  }

  if ('context' in anomaly && anomaly.context && typeof anomaly.context.count === 'number') {
    return anomaly.context.count;
  }

  return undefined;
}

/**
 * Locates an active live anomaly alert matching a live part fact by tool name, trigger event, or part id.
 */
export function findMatchingLiveAlert(
  part: LivePartFact,
  alerts?: readonly LiveAnomalyAlert[]
): LiveAnomalyAlert | undefined {
  if (!alerts || alerts.length === 0) {
    return undefined;
  }

  return alerts.find((alert) => {
    // 1. Direct trigger event match (part id in trigger event)
    if (alert.trigger_event && part.id && alert.trigger_event.includes(part.id)) {
      return true;
    }

    // 2. Tool failure alert matching tool name
    if (isToolFailureAnomalyRule(alert.rule_name)) {
      if (
        part.tool_name &&
        (alert.message.toLowerCase().includes(part.tool_name.toLowerCase()) ||
          (alert.trigger_event &&
            alert.trigger_event.toLowerCase().includes(part.tool_name.toLowerCase())))
      ) {
        return true;
      }
    }

    return false;
  });
}
