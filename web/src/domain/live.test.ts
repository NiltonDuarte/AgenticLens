// @vitest-environment node
import { describe, it, expect } from 'vitest';
import type {
  LiveAnomalyAlert,
  LiveAnomalyRule,
  LiveAnomalySeverity,
  LiveMessageFact,
  LivePartFact,
  LiveSessionNode,
  LiveSessionTree,
  LiveTraceView,
} from './live';

describe('Live Domain Models and Serialization Contracts', () => {
  describe('LiveAnomalyAlert Specifications', () => {
    it('deserializes alert across all severity tiers and canonical rule names', () => {
      const severities: readonly LiveAnomalySeverity[] = ['WARNING', 'HIGH', 'CRITICAL'];
      const rules: readonly LiveAnomalyRule[] = [
        'ConsecutiveToolErrors',
        'ContextCeilingApproach',
        'TokenVelocityRunaway',
        'ToolThrashing',
      ];

      for (let i = 0; i < rules.length; i++) {
        const rule = rules[i];
        const severity = severities[i % severities.length];
        const payload = JSON.stringify({
          session_id: `ses_alert_${i}`,
          rule_name: rule,
          severity,
          message: `Rule ${rule} triggered with severity ${severity}`,
          trigger_event: `event_${i}`,
          timestamp: '2026-10-20T12:00:00Z',
        });

        const alert: LiveAnomalyAlert = JSON.parse(payload) as LiveAnomalyAlert;
        expect(alert.session_id).toBe(`ses_alert_${i}`);
        expect(alert.rule_name).toBe(rule);
        expect(alert.severity).toBe(severity);
        expect(alert.message).toContain(rule);
        expect(alert.trigger_event).toBe(`event_${i}`);
        expect(alert.timestamp).toBe('2026-10-20T12:00:00Z');
      }
    });

    it('supports custom rule identifier strings', () => {
      const payload = JSON.stringify({
        session_id: 'ses_custom_rule',
        rule_name: 'CustomHeuristicRule',
        severity: 'WARNING',
        message: 'Custom rule alert triggered',
        trigger_event: 'custom_hook_01',
        timestamp: '2026-10-20T12:05:00Z',
      });

      const alert: LiveAnomalyAlert = JSON.parse(payload) as LiveAnomalyAlert;
      expect(alert.rule_name).toBe('CustomHeuristicRule');
      expect(alert.severity).toBe('WARNING');
    });

    it('handles hostile trigger_event strings including path traversal and SQL injection patterns', () => {
      const hostilePayloads = [
        '../../../../etc/passwd',
        "'; DROP TABLE sessions; --",
        'tool:bash:rm -rf / --no-preserve-root',
        '<script>alert("xss")</script>',
        'A'.repeat(512),
      ];

      for (const hostileTrigger of hostilePayloads) {
        const payload = JSON.stringify({
          session_id: 'ses_hostile_01',
          rule_name: 'ToolThrashing',
          severity: 'CRITICAL',
          message: 'Hostile trigger event contained',
          trigger_event: hostileTrigger,
          timestamp: '2026-10-20T12:00:00Z',
        });

        const alert: LiveAnomalyAlert = JSON.parse(payload) as LiveAnomalyAlert;
        expect(alert.trigger_event).toBe(hostileTrigger);
        expect(alert.severity).toBe('CRITICAL');
      }
    });
  });

  describe('LivePartFact Specifications', () => {
    it('deserializes complete tool execution part with input and output snippets', () => {
      const payload = JSON.stringify({
        id: 'part_tool_01',
        type: 'tool',
        tool_name: 'bash',
        input_snippet: 'go test -v ./...',
        output_snippet: 'PASS\nok backend 0.12s',
        status: 'completed',
        is_error: false,
        error_message: null,
        severity: 'none',
      });

      const part: LivePartFact = JSON.parse(payload) as LivePartFact;
      expect(part.id).toBe('part_tool_01');
      expect(part.type).toBe('tool');
      expect(part.tool_name).toBe('bash');
      expect(part.input_snippet).toBe('go test -v ./...');
      expect(part.output_snippet).toBe('PASS\nok backend 0.12s');
      expect(part.status).toBe('completed');
      expect(part.is_error).toBe(false);
      expect(part.error_message).toBeNull();
      expect(part.severity).toBe('none');
    });

    it('deserializes failing tool part with error message and high severity', () => {
      const payload = JSON.stringify({
        id: 'part_fail_01',
        type: 'tool',
        tool_name: 'readFile',
        input_snippet: 'nonexistent.txt',
        output_snippet: 'no such file or directory',
        status: 'error',
        is_error: true,
        error_message: 'no such file or directory',
        severity: 'high',
      });

      const part: LivePartFact = JSON.parse(payload) as LivePartFact;
      expect(part.is_error).toBe(true);
      expect(part.error_message).toBe('no such file or directory');
      expect(part.output_snippet).toBe('no such file or directory');
      expect(part.severity).toBe('high');
    });

    it('handles optional fields omitted or null gracefully', () => {
      const payload = JSON.stringify({
        id: 'part_reasoning_01',
        type: 'reasoning',
        status: 'completed',
        is_error: false,
        severity: 'none',
      });

      const part: LivePartFact = JSON.parse(payload) as LivePartFact;
      expect(part.id).toBe('part_reasoning_01');
      expect(part.type).toBe('reasoning');
      expect(part.tool_name).toBeUndefined();
      expect(part.input_snippet).toBeUndefined();
      expect(part.output_snippet).toBeUndefined();
      expect(part.error_message).toBeUndefined();
    });

    it('deserializes failing tool part where error_message is null', () => {
      const payload = JSON.stringify({
        id: 'part_fail_no_err_msg',
        type: 'tool',
        tool_name: 'git_checkout',
        input_snippet: 'git checkout nonexistent',
        output_snippet: '',
        status: 'error',
        is_error: true,
        error_message: null,
        severity: 'medium',
      });

      const part: LivePartFact = JSON.parse(payload) as LivePartFact;
      expect(part.id).toBe('part_fail_no_err_msg');
      expect(part.is_error).toBe(true);
      expect(part.error_message).toBeNull();
      expect(part.output_snippet).toBe('');
    });

    it('deserializes varied non-tool part types such as text, step_finish, and compaction', () => {
      const types = ['text', 'reasoning', 'step_finish', 'compaction'];
      for (const partType of types) {
        const payload = JSON.stringify({
          id: `part_${partType}_01`,
          type: partType,
          status: 'completed',
          is_error: false,
          severity: 'none',
        });

        const part: LivePartFact = JSON.parse(payload) as LivePartFact;
        expect(part.type).toBe(partType);
        expect(part.is_error).toBe(false);
        expect(part.tool_name).toBeUndefined();
      }
    });
  });

  describe('LiveMessageFact Specifications', () => {
    it('deserializes message fact with full token breakdown and reasoning content', () => {
      const payload = JSON.stringify({
        id: 'msg_01',
        role: 'assistant',
        content: 'Executing verification step',
        timestamp: '2026-10-20T12:10:00Z',
        reasoning_content: 'Checking pre-conditions before running tool',
        context_tokens: 185000,
        input_tokens: 150000,
        output_tokens: 1200,
        cache_read_tokens: 35000,
        reasoning_tokens: 800,
        total_tokens: 186200,
        parts: [
          {
            id: 'p_01',
            type: 'tool',
            tool_name: 'bash',
            input_snippet: 'make check',
            output_snippet: 'ok',
            status: 'completed',
            is_error: false,
            severity: 'none',
          },
        ],
      });

      const msg: LiveMessageFact = JSON.parse(payload) as LiveMessageFact;
      expect(msg.id).toBe('msg_01');
      expect(msg.role).toBe('assistant');
      expect(msg.content).toBe('Executing verification step');
      expect(msg.reasoning_content).toBe('Checking pre-conditions before running tool');
      expect(msg.context_tokens).toBe(185000);
      expect(msg.input_tokens).toBe(150000);
      expect(msg.output_tokens).toBe(1200);
      expect(msg.cache_read_tokens).toBe(35000);
      expect(msg.reasoning_tokens).toBe(800);
      expect(msg.total_tokens).toBe(186200);
      expect(msg.parts).toHaveLength(1);
      expect(msg.parts[0].output_snippet).toBe('ok');
    });

    it('deserializes message fact when reasoning_content is omitted or null', () => {
      const payloadWithoutReasoning = JSON.stringify({
        id: 'msg_02',
        role: 'user',
        content: 'Please verify the build status',
        timestamp: '2026-10-20T12:11:00Z',
        context_tokens: 0,
        input_tokens: 25,
        output_tokens: 0,
        cache_read_tokens: 0,
        reasoning_tokens: 0,
        total_tokens: 25,
        parts: [],
      });

      const msgWithout: LiveMessageFact = JSON.parse(payloadWithoutReasoning) as LiveMessageFact;
      expect(msgWithout.reasoning_content).toBeUndefined();
      expect(msgWithout.parts).toHaveLength(0);

      const payloadWithNull = JSON.stringify({
        id: 'msg_03',
        role: 'system',
        content: 'System initialized',
        timestamp: '2026-10-20T12:00:00Z',
        reasoning_content: null,
        context_tokens: 100,
        input_tokens: 100,
        output_tokens: 0,
        cache_read_tokens: 0,
        reasoning_tokens: 0,
        total_tokens: 100,
        parts: [],
      });

      const msgWithNull: LiveMessageFact = JSON.parse(payloadWithNull) as LiveMessageFact;
      expect(msgWithNull.reasoning_content).toBeNull();
    });

    it('deserializes message fact with zero tokens across all token fields without failure', () => {
      const payloadZeroTokens = JSON.stringify({
        id: 'msg_zero_tokens',
        role: 'system',
        content: 'ping',
        timestamp: '2026-10-20T12:00:00Z',
        context_tokens: 0,
        input_tokens: 0,
        output_tokens: 0,
        cache_read_tokens: 0,
        reasoning_tokens: 0,
        total_tokens: 0,
        parts: [],
      });

      const msg: LiveMessageFact = JSON.parse(payloadZeroTokens) as LiveMessageFact;
      expect(msg.context_tokens).toBe(0);
      expect(msg.input_tokens).toBe(0);
      expect(msg.output_tokens).toBe(0);
      expect(msg.cache_read_tokens).toBe(0);
      expect(msg.reasoning_tokens).toBe(0);
      expect(msg.total_tokens).toBe(0);
    });

    it('deserializes message fact with saturated context window exceeding healthy ceiling', () => {
      const payloadSurge = JSON.stringify({
        id: 'msg_surge_turn',
        role: 'assistant',
        content: 'Long generation turn',
        timestamp: '2026-10-20T12:15:00Z',
        reasoning_content: 'Deep planning thought',
        context_tokens: 215000,
        input_tokens: 180000,
        output_tokens: 3500,
        cache_read_tokens: 35000,
        reasoning_tokens: 1200,
        total_tokens: 218500,
        parts: [],
      });

      const msg: LiveMessageFact = JSON.parse(payloadSurge) as LiveMessageFact;
      expect(msg.context_tokens).toBe(215000);
      expect(msg.total_tokens).toBe(218500);
      expect(msg.context_tokens).toBeGreaterThan(200000);
    });

    it('deserializes message fact containing multiple mixed tool and non-tool parts', () => {
      const payload = JSON.stringify({
        id: 'msg_mixed_parts',
        role: 'assistant',
        content: 'Running multiple steps',
        timestamp: '2026-10-20T12:20:00Z',
        context_tokens: 5000,
        input_tokens: 4000,
        output_tokens: 1000,
        cache_read_tokens: 0,
        reasoning_tokens: 200,
        total_tokens: 5000,
        parts: [
          {
            id: 'p1',
            type: 'reasoning',
            status: 'completed',
            is_error: false,
            severity: 'none',
          },
          {
            id: 'p2',
            type: 'tool',
            tool_name: 'read_file',
            input_snippet: 'CONTRACT.md',
            output_snippet: '# Contract',
            status: 'completed',
            is_error: false,
            severity: 'none',
          },
          {
            id: 'p3',
            type: 'tool',
            tool_name: 'bash',
            input_snippet: 'invalid_cmd',
            output_snippet: 'command not found',
            status: 'error',
            is_error: true,
            error_message: 'command not found',
            severity: 'high',
          },
        ],
      });

      const msg: LiveMessageFact = JSON.parse(payload) as LiveMessageFact;
      expect(msg.parts).toHaveLength(3);
      expect(msg.parts[0].type).toBe('reasoning');
      expect(msg.parts[1].tool_name).toBe('read_file');
      expect(msg.parts[2].is_error).toBe(true);
    });
  });

  describe('LiveSessionNode Specifications', () => {
    it('deserializes root session node with null or omitted parent_id', () => {
      const payloadOmitted = JSON.stringify({
        id: 'ses_root_10',
        agent: 'orchestrator',
        role: 'orchestrator',
        model: 'gemini-3.8-flash',
        title: 'Task AL-19 implementation',
        time_created: '2026-10-20T10:00:00Z',
        time_updated: '2026-10-20T12:00:00Z',
        status: 'running',
        guessed_ticket_id: 'AL-19',
        ticket_confidence: 1.0,
        sub_sessions: [],
        anomalies: [],
      });

      const nodeOmitted: LiveSessionNode = JSON.parse(payloadOmitted) as LiveSessionNode;
      expect(nodeOmitted.id).toBe('ses_root_10');
      expect(nodeOmitted.parent_id).toBeUndefined();
      expect(nodeOmitted.guessed_ticket_id).toBe('AL-19');
      expect(nodeOmitted.ticket_confidence).toBe(1.0);
      expect(nodeOmitted.sub_sessions).toHaveLength(0);
      expect(nodeOmitted.anomalies).toHaveLength(0);

      const payloadNull = JSON.stringify({
        id: 'ses_root_11',
        parent_id: null,
        agent: 'orchestrator',
        role: 'orchestrator',
        model: 'gemini-3.8-flash',
        title: 'Task AL-19 root',
        time_created: '2026-10-20T10:00:00Z',
        time_updated: '2026-10-20T12:00:00Z',
        status: 'running',
        ticket_confidence: 0.0,
        sub_sessions: [],
        anomalies: [],
      });

      const nodeNull: LiveSessionNode = JSON.parse(payloadNull) as LiveSessionNode;
      expect(nodeNull.parent_id).toBeNull();
      expect(nodeNull.guessed_ticket_id).toBeUndefined();
    });

    it('deserializes child sub-session node with populated parent_id and attached anomalies', () => {
      const alert: LiveAnomalyAlert = {
        session_id: 'ses_child_20',
        rule_name: 'ConsecutiveToolErrors',
        severity: 'HIGH',
        message: '3 consecutive tool execution failures',
        trigger_event: 'part_fail_03',
        timestamp: '2026-10-20T12:15:00Z',
      };

      const payload = JSON.stringify({
        id: 'ses_child_20',
        parent_id: 'ses_root_10',
        agent: 'coder',
        role: 'coder',
        model: 'gemini-3.8-flash',
        title: 'Code implementation subagent',
        time_created: '2026-10-20T10:05:00Z',
        time_updated: '2026-10-20T12:15:00Z',
        status: 'running',
        ticket_confidence: 0.0,
        sub_sessions: [],
        anomalies: [alert],
      });

      const childNode: LiveSessionNode = JSON.parse(payload) as LiveSessionNode;
      expect(childNode.id).toBe('ses_child_20');
      expect(childNode.parent_id).toBe('ses_root_10');
      expect(childNode.anomalies).toHaveLength(1);
      expect(childNode.anomalies[0].rule_name).toBe('ConsecutiveToolErrors');
      expect(childNode.anomalies[0].severity).toBe('HIGH');
    });

    it('deserializes deeply nested 3-tier session hierarchy (root -> child -> grandchild)', () => {
      const payload = JSON.stringify({
        id: 'ses_root',
        parent_id: null,
        agent: 'orchestrator',
        role: 'orchestrator',
        model: 'gemini-3.8-flash',
        title: 'Root Orchestrator',
        time_created: '2026-10-20T10:00:00Z',
        time_updated: '2026-10-20T12:00:00Z',
        status: 'running',
        guessed_ticket_id: 'GEN-42',
        ticket_confidence: 0.85,
        anomalies: [],
        sub_sessions: [
          {
            id: 'ses_child',
            parent_id: 'ses_root',
            agent: 'coder',
            role: 'coder',
            model: 'gemini-3.8-flash',
            title: 'Coder Subagent',
            time_created: '2026-10-20T10:05:00Z',
            time_updated: '2026-10-20T11:55:00Z',
            status: 'running',
            ticket_confidence: 0.0,
            anomalies: [],
            sub_sessions: [
              {
                id: 'ses_grandchild',
                parent_id: 'ses_child',
                agent: 'tester',
                role: 'tester',
                model: 'gemini-3.8-flash',
                title: 'Nested Test Runner',
                time_created: '2026-10-20T10:10:00Z',
                time_updated: '2026-10-20T11:50:00Z',
                status: 'completed',
                ticket_confidence: 0.0,
                anomalies: [
                  {
                    session_id: 'ses_grandchild',
                    rule_name: 'ContextCeilingApproach',
                    severity: 'WARNING',
                    message: 'Context reached 165k tokens (82.5%)',
                    trigger_event: 'msg_99',
                    timestamp: '2026-10-20T11:50:00Z',
                  },
                ],
                sub_sessions: [],
              },
            ],
          },
        ],
      });

      const root: LiveSessionNode = JSON.parse(payload) as LiveSessionNode;
      expect(root.sub_sessions).toHaveLength(1);
      const child = root.sub_sessions[0];
      expect(child.parent_id).toBe('ses_root');
      expect(child.sub_sessions).toHaveLength(1);
      const grandchild = child.sub_sessions[0];
      expect(grandchild.parent_id).toBe('ses_child');
      expect(grandchild.anomalies).toHaveLength(1);
      expect(grandchild.anomalies[0].severity).toBe('WARNING');
    });
  });

  describe('LiveSessionTree Hierarchy Specifications', () => {
    it('deserializes nested multi-level tree hierarchy and aggregated active anomalies', () => {
      const anomaly1: LiveAnomalyAlert = {
        session_id: 'ses_sub_01',
        rule_name: 'ToolThrashing',
        severity: 'WARNING',
        message: 'Tool thrashing detected on tool readFile',
        trigger_event: 'readFile',
        timestamp: '2026-10-20T12:20:00Z',
      };

      const anomaly2: LiveAnomalyAlert = {
        session_id: 'ses_sub_02',
        rule_name: 'ContextCeilingApproach',
        severity: 'CRITICAL',
        message: 'Context window utilization reached 205k tokens (102.5%)',
        trigger_event: 'msg_surge_01',
        timestamp: '2026-10-20T12:22:00Z',
      };

      const payload = JSON.stringify({
        root_sessions: [
          {
            id: 'ses_root_01',
            parent_id: null,
            agent: 'orchestrator',
            role: 'orchestrator',
            model: 'gemini-3.8-flash',
            title: 'Primary Orchestrator',
            time_created: '2026-10-20T10:00:00Z',
            time_updated: '2026-10-20T12:22:00Z',
            status: 'running',
            guessed_ticket_id: 'AL-19',
            ticket_confidence: 1.0,
            anomalies: [],
            sub_sessions: [
              {
                id: 'ses_sub_01',
                parent_id: 'ses_root_01',
                agent: 'coder',
                role: 'coder',
                model: 'gemini-3.8-flash',
                title: 'Coder subagent',
                time_created: '2026-10-20T10:10:00Z',
                time_updated: '2026-10-20T12:20:00Z',
                status: 'running',
                ticket_confidence: 0.0,
                anomalies: [anomaly1],
                sub_sessions: [],
              },
              {
                id: 'ses_sub_02',
                parent_id: 'ses_root_01',
                agent: 'tester',
                role: 'tester',
                model: 'gemini-3.8-flash',
                title: 'Tester subagent',
                time_created: '2026-10-20T10:15:00Z',
                time_updated: '2026-10-20T12:22:00Z',
                status: 'running',
                ticket_confidence: 0.0,
                anomalies: [anomaly2],
                sub_sessions: [],
              },
            ],
          },
        ],
        last_refreshed_at: '2026-10-20T12:22:01Z',
        total_active_sessions: 3,
        active_anomalies: [anomaly2, anomaly1],
      });

      const tree: LiveSessionTree = JSON.parse(payload) as LiveSessionTree;
      expect(tree.total_active_sessions).toBe(3);
      expect(tree.last_refreshed_at).toBe('2026-10-20T12:22:01Z');
      expect(tree.root_sessions).toHaveLength(1);
      expect(tree.root_sessions[0].sub_sessions).toHaveLength(2);
      expect(tree.root_sessions[0].sub_sessions[0].parent_id).toBe('ses_root_01');
      expect(tree.root_sessions[0].sub_sessions[1].parent_id).toBe('ses_root_01');
      expect(tree.active_anomalies).toHaveLength(2);
      expect(tree.active_anomalies[0].severity).toBe('CRITICAL');
      expect(tree.active_anomalies[1].severity).toBe('WARNING');
    });

    it('deserializes empty session tree safely', () => {
      const payload = JSON.stringify({
        root_sessions: [],
        last_refreshed_at: '2026-10-20T12:00:00Z',
        total_active_sessions: 0,
        active_anomalies: [],
      });

      const tree: LiveSessionTree = JSON.parse(payload) as LiveSessionTree;
      expect(tree.root_sessions).toHaveLength(0);
      expect(tree.total_active_sessions).toBe(0);
      expect(tree.active_anomalies).toHaveLength(0);
    });

    it('deserializes session tree with multiple concurrent root sessions and distinct ticket formats', () => {
      const payload = JSON.stringify({
        root_sessions: [
          {
            id: 'ses_root_al',
            parent_id: null,
            agent: 'orchestrator',
            role: 'orchestrator',
            model: 'gemini-3.8-flash',
            title: 'AL Feature Orchestration',
            time_created: '2026-10-20T10:00:00Z',
            time_updated: '2026-10-20T12:00:00Z',
            status: 'running',
            guessed_ticket_id: 'AL-19',
            ticket_confidence: 1.0,
            sub_sessions: [],
            anomalies: [],
          },
          {
            id: 'ses_root_po',
            parent_id: null,
            agent: 'agentic-lens-po',
            role: 'po',
            model: 'gemini-3.8-flash',
            title: 'Product Owner Ticket Inception',
            time_created: '2026-10-20T10:30:00Z',
            time_updated: '2026-10-20T12:00:00Z',
            status: 'running',
            guessed_ticket_id: 'PO-2',
            ticket_confidence: 0.9,
            sub_sessions: [],
            anomalies: [],
          },
        ],
        last_refreshed_at: '2026-10-20T12:00:00Z',
        total_active_sessions: 2,
        active_anomalies: [],
      });

      const tree: LiveSessionTree = JSON.parse(payload) as LiveSessionTree;
      expect(tree.root_sessions).toHaveLength(2);
      expect(tree.total_active_sessions).toBe(2);
      expect(tree.root_sessions[0].guessed_ticket_id).toBe('AL-19');
      expect(tree.root_sessions[1].guessed_ticket_id).toBe('PO-2');
    });
  });

  describe('LiveTraceView Specifications', () => {
    it('deserializes trace view with message history, tool statistics, and anomalies', () => {
      const alert: LiveAnomalyAlert = {
        session_id: 'ses_trace_01',
        rule_name: 'TokenVelocityRunaway',
        severity: 'CRITICAL',
        message: 'Abnormally high token velocity of 65000 tokens consumed in turn 4',
        trigger_event: 'msg_04',
        timestamp: '2026-10-20T12:30:00Z',
      };

      const payload = JSON.stringify({
        session_id: 'ses_trace_01',
        messages: [
          {
            id: 'msg_01',
            role: 'assistant',
            content: 'Turn 1 execution',
            timestamp: '2026-10-20T12:25:00Z',
            reasoning_content: 'Step 1 thought',
            context_tokens: 5000,
            input_tokens: 4500,
            output_tokens: 500,
            cache_read_tokens: 0,
            reasoning_tokens: 150,
            total_tokens: 5000,
            parts: [],
          },
        ],
        total_tool_calls: 8,
        failed_tool_calls: 3,
        recent_errors: [
          {
            id: 'part_err_01',
            type: 'tool',
            tool_name: 'bash',
            input_snippet: 'cargo build',
            output_snippet: 'cargo: command not found',
            status: 'error',
            is_error: true,
            error_message: 'cargo: command not found',
            severity: 'medium',
          },
        ],
        anomalies: [alert],
      });

      const trace: LiveTraceView = JSON.parse(payload) as LiveTraceView;
      expect(trace.session_id).toBe('ses_trace_01');
      expect(trace.total_tool_calls).toBe(8);
      expect(trace.failed_tool_calls).toBe(3);
      expect(trace.messages).toHaveLength(1);
      expect(trace.recent_errors).toHaveLength(1);
      expect(trace.recent_errors[0].output_snippet).toBe('cargo: command not found');
      expect(trace.anomalies).toHaveLength(1);
      expect(trace.anomalies[0].rule_name).toBe('TokenVelocityRunaway');
    });

    it('deserializes empty trace view safely', () => {
      const payload = JSON.stringify({
        session_id: 'ses_empty_trace',
        messages: [],
        total_tool_calls: 0,
        failed_tool_calls: 0,
        recent_errors: [],
        anomalies: [],
      });

      const trace: LiveTraceView = JSON.parse(payload) as LiveTraceView;
      expect(trace.session_id).toBe('ses_empty_trace');
      expect(trace.messages).toHaveLength(0);
      expect(trace.total_tool_calls).toBe(0);
      expect(trace.failed_tool_calls).toBe(0);
      expect(trace.recent_errors).toHaveLength(0);
      expect(trace.anomalies).toHaveLength(0);
    });

    it('deserializes trace with 100% tool failure rate and multiple recent errors', () => {
      const payload = JSON.stringify({
        session_id: 'ses_all_failed',
        messages: [],
        total_tool_calls: 5,
        failed_tool_calls: 5,
        recent_errors: [
          {
            id: 'err_5',
            type: 'tool',
            tool_name: 'bash',
            status: 'error',
            is_error: true,
            error_message: 'segmentation fault (core dumped)',
            severity: 'critical',
          },
          {
            id: 'err_4',
            type: 'tool',
            tool_name: 'bash',
            status: 'error',
            is_error: true,
            error_message: 'process killed: OOM',
            severity: 'critical',
          },
        ],
        anomalies: [
          {
            session_id: 'ses_all_failed',
            rule_name: 'ConsecutiveToolErrors',
            severity: 'CRITICAL',
            message: 'Tool execution error streak reached 5 consecutive failures without recovery',
            trigger_event: 'err_5',
            timestamp: '2026-10-20T12:00:00Z',
          },
        ],
      });

      const trace: LiveTraceView = JSON.parse(payload) as LiveTraceView;
      expect(trace.total_tool_calls).toBe(5);
      expect(trace.failed_tool_calls).toBe(5);
      expect(trace.recent_errors).toHaveLength(2);
      expect(trace.recent_errors[0].severity).toBe('critical');
      expect(trace.anomalies).toHaveLength(1);
      expect(trace.anomalies[0].severity).toBe('CRITICAL');
    });
  });
});
