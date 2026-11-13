import { afterEach, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { App } from './App';
import { mockAgenticMetricsLedgers, mockBuildInfo, mockFleetSummary } from './test/fixtures/mockData';

afterEach(() => { cleanup(); vi.restoreAllMocks(); vi.unstubAllGlobals(); });

it('hands fetched literal, case-sensitive policy through the real App to fleet sorting', async () => {
  const base = mockAgenticMetricsLedgers['AL-1']!;
  const ledgers = ['T[1]-10', 't[1]-1', 'T[1]-2'].map((ticket_id) => ({
    ...base, ticket_id, facts: {...base.facts, ticket_id},
  }));
  // Raw App rendering deliberately bypasses the component-fixture configuration provider.
  const config = {
    project_name: 'Wired policy project', workspaces_roots: [{ path: 'tickets', discovery: { max_depth: 1, max_directories: 300 } }], version: 'v1.0.0',
    prefixes: [{prefix: 'T[1]-', description: 'Literal bracketed tickets'}],
    case_insensitive: false, role_order: [], role_mappings: {},
  };
  vi.stubGlobal('fetch', vi.fn(async (input: string | URL | Request) => {
    const route = String(input).split('?')[0];
    const payload: unknown = route === '/api/config' ? config
      : route === '/api/version' ? mockBuildInfo
      : route === '/api/fleet' ? mockFleetSummary
      : route === '/api/workspaces' ? ledgers : undefined;
    if (payload === undefined) throw new Error(`Unexpected fixture route: ${route}`);
    return new Response(JSON.stringify(payload), {status: 200, headers: {'Content-Type': 'application/json'}});
  }));
  const {container} = render(<App />);
  expect(await screen.findByTestId('landing-project-context')).toHaveTextContent('Wired policy project');
  fireEvent.click(screen.getByRole('button', {name: /open fleet overview/i}));
  await screen.findByText('Ticket Workspaces Directory');
  await waitFor(() => expect([...container.querySelectorAll('.ticket-table tbody tr')]
    .map((row) => row.querySelector('td')?.textContent)).toEqual(['T[1]-2', 'T[1]-10', 't[1]-1']));
});

it('uses the fetched shared polling cadence through App and restores the recovered live connection status', async () => {
  vi.stubGlobal('EventSource', undefined);
  const cadence = 333;
  const timer = vi.spyOn(globalThis, 'setInterval');
  const config = {project_name: 'Polling policy project', workspaces_roots: [{ path: 'tickets', discovery: { max_depth: 1, max_directories: 300 } }], version: 'v1.0.0',
    prefixes: [{prefix: 'T-', description: 'Tickets'}], case_insensitive: false, role_order: [], role_mappings: {}, live_poll_interval_ms: cadence};
  const emptyTree = {root_sessions: [], active_anomalies: [], total_active_sessions: 0, last_refreshed_at: '2026-11-11T12:00:00Z'};
  let polls = 0;
  vi.stubGlobal('fetch', vi.fn(async (input: string | URL | Request) => {
    const route = String(input).split('?')[0];
    if (route === '/api/live/sessions') {
      if (++polls === 1) return new Response('', {status: 503});
      return new Response(JSON.stringify(emptyTree), {status: 200});
    }
    const payload: unknown = route === '/api/config' ? config : route === '/api/version' ? mockBuildInfo
      : route === '/api/fleet' ? mockFleetSummary : route === '/api/workspaces' ? [mockAgenticMetricsLedgers['AL-1']!] : undefined;
    if (payload === undefined) throw new Error(`Unexpected fixture route: ${route}`);
    return new Response(JSON.stringify(payload), {status: 200});
  }));
  render(<App />);
  await screen.findByTestId('landing-project-context');
  fireEvent.click(screen.getByRole('button', {name: 'Live Observability'}));
  await screen.findByTestId('live-connection-error-banner');
  expect(timer).toHaveBeenCalledWith(expect.any(Function), cadence);
  await screen.findByText('No Active Sessions');
  expect(screen.getByText('Fallback Polling (REST)')).toBeVisible();
  expect(screen.queryByTestId('live-connection-error-banner')).not.toBeInTheDocument();
  expect(polls).toBeGreaterThanOrEqual(2);
});
