import { useCallback, useLayoutEffect, useRef, useState } from 'react';
import { fetchWorkspaceLedgers } from '../api/client';
import { reconcileTagCatalog } from '../domain/tags';
import type { AgenticMetricsLedger, WorkspaceDiagnostic } from '../domain/types';

interface WorkspaceLedgersSnapshot {
  readonly ledgers: readonly AgenticMetricsLedger[];
  readonly tagCatalog: readonly AgenticMetricsLedger[];
  readonly diagnostics: readonly WorkspaceDiagnostic[];
}

const EMPTY_SNAPSHOT: WorkspaceLedgersSnapshot = {
  ledgers: [],
  tagCatalog: [],
  diagnostics: [],
};

export function useWorkspaceLedgers(
  selectedTicketTypes: readonly string[],
  selectedPrefixes: readonly string[],
  fallbackError: string,
) {
  const [snapshot, setSnapshot] = useState<WorkspaceLedgersSnapshot | null>(null);
  const [error, setError] = useState<string | null>(null);
  const generation = useRef(0);

  const loadLedgers = useCallback(async () => {
    const request = ++generation.current;
    setError(null);

    const types = selectedTicketTypes.length === 0 || selectedTicketTypes.includes('all')
      ? undefined
      : selectedTicketTypes;
    const prefixes = selectedPrefixes.length === 0 || selectedPrefixes.includes('all')
      ? undefined
      : selectedPrefixes.map((prefix) => prefix.slice('prefix:'.length));
    const filtered = types !== undefined || prefixes !== undefined;

    try {
      const response = filtered
        ? await fetchWorkspaceLedgers('all', undefined, undefined, types, prefixes)
        : await fetchWorkspaceLedgers('all');

      if (request !== generation.current) return;
      setSnapshot((previous) => ({
        ledgers: response.data,
        tagCatalog: filtered ? reconcileTagCatalog(previous?.tagCatalog ?? [], response.data) : response.data,
        diagnostics: response.diagnostics ?? [],
      }));
    } catch (cause: unknown) {
      if (request !== generation.current) return;
      setError(cause instanceof Error ? cause.message : fallbackError);
    }
  }, [selectedTicketTypes, selectedPrefixes, fallbackError]);

  useLayoutEffect(() => {
    const requests = generation;
    void loadLedgers();
    return () => { requests.current++; };
  }, [loadLedgers]);

  return {
    ...(snapshot ?? EMPTY_SNAPSHOT),
    loading: snapshot === null && error === null,
    error,
    loadLedgers,
  };
}
