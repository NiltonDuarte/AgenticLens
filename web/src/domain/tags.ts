import type { AgenticMetricsLedger } from './types';

export type TicketTags = Readonly<Record<string, string>>;
export interface TagFilter {
  readonly key: string;
  readonly values: readonly string[];
  readonly missing: boolean;
}
export interface TagChoice {
  readonly key: string;
  readonly values: readonly string[];
}
export function tagValue(tags: TicketTags | undefined, key: string): string | undefined {
  return tags && Object.hasOwn(tags, key) ? tags[key] : undefined;
}
export function matchesTags(tags: TicketTags | undefined, filters: readonly TagFilter[]): boolean {
  return filters.every((filter) => {
    const value = tagValue(tags, filter.key);
    return value === undefined ? filter.missing : filter.values.includes(value);
  });
}
export function filterLedgersByTags(ledgers: readonly AgenticMetricsLedger[], filters: readonly TagFilter[]): readonly AgenticMetricsLedger[] {
  return filters.length === 0 ? ledgers : ledgers.filter((ledger) => matchesTags(ledger.facts.tags, filters));
}
export function tagChoices(tags: readonly (TicketTags | undefined)[]): readonly TagChoice[] {
  const keys = new Map<string, Set<string>>();
  for (const entry of tags) for (const [key, value] of Object.entries(entry ?? {})) {
    const values = keys.get(key) ?? new Set<string>();
    values.add(value);
    keys.set(key, values);
  }
  return [...keys].sort(([a], [b]) => a < b ? -1 : a > b ? 1 : 0).map(([key, values]) => ({ key, values: [...values].sort() }));
}
export function tagCategory(tags: TicketTags | undefined, key: string): { readonly id: string; readonly label: string } {
  const value = tagValue(tags, key);
  return value === undefined ? { id: 'missing', label: 'Missing' } : { id: JSON.stringify(['value', value]), label: `Value: ${value}` };
}
export function tagColors(ids: readonly string[]): ReadonlyMap<string, string> {
  const assigned = new Map<string, string>([['missing', '#8793a3']]);
  const values = [...new Set(ids)].filter((id) => id !== 'missing').sort();
  for (const [index, id] of values.entries()) assigned.set(id, `hsl(${(index * 137.508) % 360} 65% 58%)`);
  return assigned;
}

/** Refresh observed tickets without erasing tag choices outside a filtered cohort. */
export function reconcileTagCatalog(previous: readonly AgenticMetricsLedger[], incoming: readonly AgenticMetricsLedger[]): readonly AgenticMetricsLedger[] {
  const catalog = new Map<string, AgenticMetricsLedger>();
  for (const ledger of [...previous, ...incoming]) {
    // Anonymous entries have no stable ticket identity. Retain distinct tag-map
    // representatives only; their numerical metrics are never aggregated here.
    const key = ledger.ticket_id
      ? JSON.stringify(['ticket', ledger.ticket_id])
      : JSON.stringify(['anonymous-tags', Object.entries(ledger.facts.tags ?? {}).sort(([a], [b]) => a < b ? -1 : a > b ? 1 : 0)]);
    catalog.set(key, ledger);
  }
  return [...catalog.values()];
}
