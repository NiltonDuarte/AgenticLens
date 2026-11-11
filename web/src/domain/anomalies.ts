import type { EventFact, Anomaly, AnomalyRuleID } from '../types/telemetry';
import type { LivePartFact, LiveAnomalyAlert, LiveAnomalyRule } from './live';
import { formatTokensK } from './formatters';

export interface AnomalyDisplayText {
  readonly title: string;
  readonly description: string;
}

/** Derives presentation strings from the canonical rule and structured context. */
export function formatAnomaly(anomaly: Anomaly, roleLabel: (role: string) => string = (role) => role): AnomalyDisplayText {
  const context = anomaly.context;
  const tool = context.tool_name?.trim() || 'tool';
  switch (anomaly.rule_id) {
    case 'TOOL_RETRY_STREAK':
      if (context.tool_name && context.count !== undefined) {
        return { title: `Tool retry streak: ${tool} (${context.count} consecutive errors)`, description: `Tool '${tool}' encountered ${context.count} consecutive execution failures.` };
      }
      break;
    case 'TOOL_HIGH_FAILURE_RATE':
      if (context.tool_name && context.count !== undefined && context.total_calls !== undefined && context.failure_rate_pct !== undefined) {
        return {
          title: `High tool failure rate: ${tool} (${context.failure_rate_pct.toFixed(1)}%)`,
          description: `Tool '${tool}' failed ${context.count} out of ${context.total_calls} calls (${context.failure_rate_pct.toFixed(1)}% failure rate, threshold: ${context.threshold || 'configured threshold'}).`,
        };
      }
      break;
    case 'CONTEXT_SATURATION_SPIKE':
      if (context.context_tokens !== undefined && context.token_delta === undefined && context.threshold) {
        const configuredCeiling = Number(context.threshold);
        const ceiling = Number.isFinite(configuredCeiling) && configuredCeiling > 0
          ? formatTokensK(configuredCeiling) : context.threshold;
        return { title: `Context ceiling breached (${formatTokensK(context.context_tokens)} tokens)`, description: `Context tokens reached ${formatTokensK(context.context_tokens)}, breaching the ${ceiling} ${anomaly.severity.toLowerCase()} ceiling.` };
      }
      if (context.token_delta !== undefined && context.context_tokens !== undefined) {
        return { title: `Context saturation spike (+${formatTokensK(context.token_delta)} tokens)`, description: `Context tokens surged by ${formatTokensK(context.token_delta)} to ${formatTokensK(context.context_tokens)} tokens.` };
      }
      break;
    case 'SUBAGENT_REWORK_LOOP':
      if (context.role && context.secondary_role && context.count !== undefined) {
        return {
          title: `Subagent rework loop: ${roleLabel(context.role)} <-> ${roleLabel(context.secondary_role)} (${context.count} cycles)`,
          description: `Subagent roles '${roleLabel(context.role)}' and '${roleLabel(context.secondary_role)}' alternated for ${context.count} consecutive cycles.`,
        };
      }
      break;
  }
  const name = anomaly.rule_id.toLowerCase().replaceAll('_', ' ') || 'execution anomaly';
  return { title: name.charAt(0).toUpperCase() + name.slice(1), description: `${name} detected; some diagnostic context is unavailable.` };
}

/**
 * Union of domain and live anomaly rule identifiers.
 */
export type AnomalyRule = AnomalyRuleID | LiveAnomalyRule;

/**
 * Anomaly rules specifically attributed to tool call failures and thrashing.
 */
export type ToolFailureAnomalyRule =
  | 'TOOL_RETRY_STREAK'
  | 'TOOL_HIGH_FAILURE_RATE'
  | 'ConsecutiveToolErrors'
  | 'ToolHighFailureRate'
  | 'ToolThrashing';

type AnomalyMatchEvent = Omit<EventFact, 'session_id'> & { readonly session_id?: string | null | undefined };

/**
 * Checks whether an anomaly rule identifier matches tool failure anomaly patterns
 * (retries, high failure rates, consecutive errors, thrashing).
 */
export function isToolFailureAnomalyRule(
  ruleId: AnomalyRule | string
): ruleId is ToolFailureAnomalyRule {
  return (
    ruleId === 'TOOL_RETRY_STREAK' ||
    ruleId === 'TOOL_HIGH_FAILURE_RATE' ||
    ruleId === 'ConsecutiveToolErrors' ||
    ruleId === 'ToolHighFailureRate' ||
    ruleId === 'ToolThrashing'
  );
}

/**
 * Evaluates whether an EventFact matches a detected Anomaly.
 * Enforces strict session identity, error status for tool failure rules, and turn/tool matching.
 */
export function matchesAnomaly(event: AnomalyMatchEvent, anomaly: Anomaly): boolean {
  const ctx = anomaly.context;

  // Persisted evidence refs identify the exact source records. When present,
  // do not broaden attribution by tool, turn, or role.
  if (anomaly.evidence && anomaly.evidence.length > 0) {
    if (!event.event_id) return false;
    const matchingRefs = anomaly.evidence.filter((ref) => ref.event_id === event.event_id);
    if (event.session_id) {
      return matchingRefs.some((ref) => ref.session_id === event.session_id);
    }
    // Nested session event rows omit session_id. Accept an ID only when its
    // evidence refs identify one unambiguous owning session.
    return new Set(matchingRefs.map((ref) => ref.session_id)).size === 1;
  }

  // A fallback cannot establish an explicitly scoped session from a missing ID.
  if (ctx.session_id && ctx.session_id !== event.session_id) {
    return false;
  }
  if (ctx.tool_name && (!event.tool_name || ctx.tool_name.toLowerCase() !== event.tool_name.toLowerCase())) {
    return false;
  }
  if (ctx.turn !== undefined && ctx.turn !== event.turn) {
    return false;
  }
  if ((ctx.role || ctx.secondary_role) &&
    ctx.role?.toLowerCase() !== event.role.toLowerCase() &&
    ctx.secondary_role?.toLowerCase() !== event.role.toLowerCase()) {
    return false;
  }

  // Tool Failure Anomaly Status Criterion:
  // If anomaly is a tool failure rule, only match if the event is an error
  if (isToolFailureAnomalyRule(anomaly.rule_id)) {
    if (event.status !== 'error' && event.status !== 'failed') {
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

/** Find streak evidence independently of the display ordering of other rules. */
export function findMatchingStreak(
  event: EventFact,
  anomalies?: readonly Anomaly[]
): number | undefined {
  const streak = anomalies?.find((anomaly) => extractStreakCount(anomaly) !== undefined && matchesAnomaly(event, anomaly));
  return extractStreakCount(streak);
}

/** Rebind a selected summary only when its rule and occurrence context identify one counterpart. */
export function rebindAnomaly(selected: Anomaly | null, anomalies: readonly Anomaly[]): Anomaly | null {
  if (!selected) return null;
  if (anomalies.includes(selected)) return selected;
  const matches = anomalies.filter((anomaly) => anomaly.rule_id === selected.rule_id &&
    anomaly.context.session_id === selected.context.session_id &&
    anomaly.context.turn === selected.context.turn &&
    anomaly.context.tool_name === selected.context.tool_name &&
    anomaly.context.role === selected.context.role &&
    anomaly.context.secondary_role === selected.context.secondary_role);
  return matches.length === 1 ? matches[0] ?? null : null;
}

/**
 * Only rules that measure consecutive failures can supply a streak count.
 */
export function extractStreakCount(
  anomaly?: Anomaly | LiveAnomalyAlert | null | undefined
): number | undefined {
  if (!anomaly) {
    return undefined;
  }
  if ('rule_name' in anomaly && anomaly.rule_name === 'ConsecutiveToolErrors' && typeof anomaly.count === 'number') {
    return anomaly.count;
  }
  if ('rule_id' in anomaly && anomaly.rule_id === 'TOOL_RETRY_STREAK' && typeof anomaly.context.count === 'number') {
    return anomaly.context.count;
  }

  return undefined;
}

/**
 * Matches structured identities first. Legacy triggers use the exact emitted
 * part/evidence or tool/part-list grammar, never message/output substrings.
 */
export function findMatchingLiveAlert(
  part: LivePartFact,
  alerts?: readonly LiveAnomalyAlert[],
  sessionId?: string
): LiveAnomalyAlert | undefined {
  if (!alerts || alerts.length === 0) {
    return undefined;
  }

  return alerts.find((alert) => {
    if (sessionId !== undefined && alert.session_id !== sessionId) return false;
    if (!isToolFailureAnomalyRule(alert.rule_name) || !part.is_error) return false;
    if (alert.tool_name && alert.tool_name.toLowerCase() !== part.tool_name?.toLowerCase()) return false;
    if (alert.part_id) {
      return alert.part_id === part.id;
    }
    if (alert.rule_name === 'ToolHighFailureRate') {
      return Boolean(alert.tool_name && part.tool_name);
    }
    const separator = alert.trigger_event.indexOf(': ');
    if (alert.rule_name === 'ConsecutiveToolErrors') {
      const partId = separator < 0 ? alert.trigger_event : alert.trigger_event.slice(0, separator);
      return part.id.length > 0 && partId === part.id;
    }
    if (alert.rule_name !== 'ToolThrashing' || separator < 0) return false;
    const toolName = alert.trigger_event.slice(0, separator);
    if (toolName.toLowerCase() !== part.tool_name?.toLowerCase()) return false;
    const partIds = alert.trigger_event.slice(separator + 2).split(', ');
    return part.id.length > 0 && partIds.includes(part.id);
  });
}
