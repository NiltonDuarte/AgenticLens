import { describe, expect, it } from 'vitest';
import { filterLedgersByTags, matchesTags, tagCategory, tagChoices, tagColors, reconcileTagCatalog } from './tags';
import { mockAgenticMetricsLedgers } from '../test/fixtures/mockData';

describe('ticket tag cohorts', () => {
  it('ANDs keys on one ticket, ORs values/Missing, and retains whole ticket metrics', () => {
    const base = mockAgenticMetricsLedgers['AL-1']!;
    const ledgers = [{ ...base, facts: { ...base.facts, tags: { team: 'core' } } }, { ...base, facts: { ...base.facts, tags: { environment: 'prod' } } }, { ...base, facts: { ...base.facts, tags: { team: 'core', environment: 'prod' } } }, base];
    const filters = [{ key: 'team', values: ['core', 'other'], missing: false }, { key: 'environment', values: ['prod'], missing: false }];
    expect(filterLedgersByTags(ledgers, filters)).toEqual([ledgers[2]]);
    expect(filterLedgersByTags(ledgers, filters)[0]).toBe(ledgers[2]);
    expect(filterLedgersByTags(ledgers, [])).toBe(ledgers);
    expect(filterLedgersByTags(ledgers, [{ key: 'team', values: ['core'], missing: true }])).toHaveLength(4);
  });
  it('keeps reserved display words and prototype keys as literal values', () => {
    const tags = JSON.parse('{"__proto__":"Missing","constructor":"Mixed","all":"ALL"}') as Record<string, string>;
    expect(matchesTags(tags, [{ key: '__proto__', values: ['Missing'], missing: false }])).toBe(true);
    expect(matchesTags({}, [{ key: 'constructor', values: [], missing: true }])).toBe(true);
    expect(tagChoices([tags])).toEqual([{ key: '__proto__', values: ['Missing'] }, { key: 'all', values: ['ALL'] }, { key: 'constructor', values: ['Mixed'] }]);
    expect(tagCategory(tags, '__proto__').id).not.toBe(tagCategory({}, '__proto__').id);
    const colors = tagColors([tagCategory(tags, '__proto__').id, 'missing', tagCategory({ team: 'Aa' }, 'team').id, tagCategory({ team: 'BB' }, 'team').id]);
    expect(new Set(colors.values()).size).toBe(4);
    expect(tagColors([...colors.keys()].reverse())).toEqual(colors);
  });
  it('updates catalog tickets and deduplicates anonymous tag representatives without touching metrics', () => {
    const base = mockAgenticMetricsLedgers['AL-1']!;
    const old = { ...base, facts: { ...base.facts, tags: { team: 'old' } } };
    const updated = { ...base, facts: { ...base.facts, tags: JSON.parse('{"__proto__":"new"}') as Record<string, string> } };
    const outside = { ...base, ticket_id: 'constructor', facts: { ...base.facts, tags: { team: 'outside' } } };
    const anonymous = { ...base, ticket_id: null, facts: { ...base.facts, tags: { team: 'anonymous' } } };
    const missing = { ...base, ticket_id: null };
    const incoming = [updated, anonymous, missing];
    const catalog = reconcileTagCatalog([old, outside, anonymous], incoming);
    expect(catalog).toEqual([updated, outside, anonymous, missing]);
    expect(reconcileTagCatalog(catalog, incoming)).toEqual(catalog);
    expect(old.facts.tags).toEqual({ team: 'old' });
    expect(catalog[0]).toBe(updated);
  });

});


it('retains independently selected tag facts for two physical locations of one ticket', () => {
  const base = mockAgenticMetricsLedgers['AL-1']!;
  const first = {...base, workspace_path: '/records/first', facts: {...base.facts, tags: {team: 'one'}}};
  const second = {...base, workspace_path: '/records/second', facts: {...base.facts, tags: {team: 'two'}}};
  const updated = {...second, facts: {...second.facts, tags: {team: 'updated'}}};
  expect(reconcileTagCatalog([first, second], [updated])).toEqual([first, updated]);
});
