import { describe, it, expect } from 'vitest';
import type {
  AgenticFactLedger,
  MigrationInfo,
  MigrationStatus,
  EventFact,
  CodeFacts,
} from './types';

describe('Multi-Version Schema Evolution and Migration Types', () => {
  describe('Schema Version 1 Historical Telemetry Ingestion', () => {
    it('deserializes raw historical v1 ledger without schema_version or events', () => {
      const historicalJSON = JSON.stringify({
        ticket_id: 'AL-HISTORICAL',
        ticket_title: 'Unversioned Historical Task',
        ticket_type: 'feature',
        started_at: '2026-10-18T10:00:00Z',
        completed_at: '2026-10-18T11:00:00Z',
        updated_at: '2026-10-18T11:00:00Z',
        code: {
          lines_added: 150,
          lines_deleted: 20,
          test_lines_added: 80,
          test_lines_deleted: 0,
          prod_lines_added: 70,
          prod_lines_deleted: 20,
          files_changed: 5,
          sdlc_iterations: 2,
        },
        invocations: [
          {
            id: 'ses-1',
            parent_id: null,
            role: 'orchestrator',
            model: 'gemini-3.8-flash',
            user: 'dev',
            started_at: '2026-10-18T10:00:00Z',
            completed_at: '2026-10-18T10:30:00Z',
            active_duration_seconds: 1800,
            turns: 8,
            tokens: {
              input: 10000,
              output: 2000,
              cache_read: 5000,
              reasoning: 500,
              peak_context: 40000,
            },
            cost_usd: 0.25,
            compactions: 0,
          },
        ],
        tools: [
          {
            name: 'bash',
            calls: 12,
            errors: 1,
          },
        ],
      });

      const ledger: AgenticFactLedger = JSON.parse(historicalJSON);
      expect(ledger.schema_version).toBeUndefined();
      expect(ledger.migration).toBeUndefined();
      expect(ledger.events).toBeUndefined();
      expect(ledger.ticket_id).toBe('AL-HISTORICAL');
      expect(ledger.code.lines_added).toBe(150);
      expect(ledger.invocations).toHaveLength(1);
      expect(ledger.tools).toHaveLength(1);
    });
  });

  describe('Schema Version 2 Upgraded Telemetry Ledgers', () => {
    it('accurately parses version 2 ledger with migration metadata and empty events array', () => {
      const v2Payload = JSON.stringify({
        schema_version: 2,
        ticket_id: 'AL-V2',
        ticket_title: 'Migrated V2 Task',
        ticket_type: 'feature',
        started_at: '2026-10-18T10:00:00Z',
        completed_at: '2026-10-18T11:00:00Z',
        updated_at: '2026-10-18T11:00:00Z',
        migration: {
          schema_version: 2,
          status: 'completed',
          steps_completed: ['schema_v2_baseline'],
          steps_pending: [],
          last_migrated_at: '2026-10-18T11:00:00Z',
          last_migrated_by: 'universal-migration-engine',
          notes: 'Baseline migration from v1 to v2',
        },
        code: {
          lines_added: 50,
          lines_deleted: 10,
          test_lines_added: 25,
          test_lines_deleted: 0,
          prod_lines_added: 25,
          prod_lines_deleted: 10,
          files_changed: 2,
          sdlc_iterations: 1,
        },
        invocations: [],
        tools: [],
        events: [],
      });

      const ledger: AgenticFactLedger = JSON.parse(v2Payload);
      expect(ledger.schema_version).toBe(2);
      expect(ledger.events).toBeDefined();
      expect(ledger.events).toHaveLength(0);

      const mig = ledger.migration as MigrationInfo;
      expect(mig).toBeDefined();
      expect(mig.schema_version).toBe(2);
      expect(mig.status).toBe('completed');
      expect(mig.steps_completed).toContain('schema_v2_baseline');
      expect(mig.steps_pending).toHaveLength(0);
      expect(mig.last_migrated_by).toBe('universal-migration-engine');
    });

    it('accurately handles partial migration status representation', () => {
      const partialMigration: MigrationInfo = {
        schema_version: 2,
        status: 'partial' as MigrationStatus,
        steps_completed: ['git_code_facts'],
        steps_pending: ['opencode_db'],
        last_migrated_at: '2026-10-18T10:00:00Z',
        last_migrated_by: 'engineer-a',
        notes: 'Pending database resolution',
      };

      expect(partialMigration.status).toBe('partial');
      expect(partialMigration.steps_completed).toEqual(['git_code_facts']);
      expect(partialMigration.steps_pending).toEqual(['opencode_db']);
    });
  });

  describe('Adversarial Boundary Validation', () => {
    it('handles empty code facts defensively', () => {
      const code: CodeFacts = {
        lines_added: 0,
        lines_deleted: 0,
        test_lines_added: 0,
        test_lines_deleted: 0,
        prod_lines_added: 0,
        prod_lines_deleted: 0,
        files_changed: 0,
        sdlc_iterations: 0,
      };

      expect(code.lines_added).toBe(0);
      expect(code.sdlc_iterations).toBe(0);
    });

    it('preserves populated chronological events during schema upgrades', () => {
      const events: readonly EventFact[] = [
        {
          timestamp: '2026-10-18T10:00:01Z',
          session_id: 's1',
          turn: 1,
          role: 'coder',
          event_type: 'turn_message',
          status: 'completed',
          context_tokens: 3000,
        },
        {
          timestamp: '2026-10-18T10:00:05Z',
          session_id: 's1',
          turn: 1,
          role: 'coder',
          event_type: 'tool_call',
          tool_name: 'bash',
          status: 'completed',
          duration_ms: 120,
          context_tokens: 3000,
        },
      ];

      expect(events).toHaveLength(2);
      expect(events[0].event_type).toBe('turn_message');
      expect(events[1].event_type).toBe('tool_call');
      expect(events[1].duration_ms).toBe(120);
    });
  });
});
