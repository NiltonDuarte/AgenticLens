import { formatRoleLabel } from '../domain/formatters';
import { compareRoles, projectedRole } from '../domain/roles';
import { compareText } from '../domain/tickets';
import { createContext, useContext, useMemo } from 'react';
import type { ProjectConfigResponse } from '../domain/types';
import type { TicketRecognitionPolicy } from '../domain/tickets';

export const ProjectConfigContext = createContext<ProjectConfigResponse | null>(null);

export function useProjectConfig(): ProjectConfigResponse | null {
  return useContext(ProjectConfigContext);
}

export function ticketPolicy(config: ProjectConfigResponse | null | undefined): TicketRecognitionPolicy | undefined {
  return config ? {prefixes: config.prefixes.map((entry) => entry.prefix), caseInsensitive: config.case_insensitive} : undefined;
}

/** Source persona labels are projected once; recorded role values remain untouched. */
export function useRolePresentation() {
  const config = useProjectConfig();
  return useMemo(() => {
  const sourceLabel = (role: string): string => formatRoleLabel(projectedRole(role, config?.role_mappings ?? {}));
  const compareProjected = (left: string, right: string): number => compareRoles(left, right, config?.role_order ?? []);
  const compareSource = (left: string, right: string): number => compareProjected(
    projectedRole(left, config?.role_mappings ?? {}), projectedRole(right, config?.role_mappings ?? {})
  ) || compareText(left, right);
  const projectedLabel = formatRoleLabel;
  return {sourceLabel, projectedLabel, compareProjected, compareSource};
  }, [config]);
}
