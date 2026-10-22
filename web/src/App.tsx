import React, { useEffect, useState, useMemo } from 'react';
import type { ActiveTab, ProjectConfigResponse } from './domain/types';
import { fetchProjectConfig } from './api/client';
import { Navbar } from './components/Navbar';
import { LandingPage } from './views/LandingPage';
import { FleetOverview } from './views/FleetOverview';
import { TicketDeepDive } from './views/TicketDeepDive';
import { MetricProgression } from './views/MetricProgression';
import { TicketComparisons } from './views/TicketComparisons';
import { LiveObservability } from './views/LiveObservability';

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState<ActiveTab>('landing');
  const [selectedTicketId, setSelectedTicketId] = useState<string | null>(null);
  const [projectConfig, setProjectConfig] = useState<ProjectConfigResponse | null>(null);

  useEffect(() => {
    let mounted = true;
    void fetchProjectConfig().then((resp) => {
      if (mounted) {
        setProjectConfig(resp.data);
      }
    }).catch(() => {
      // Ignore network failure on startup; keep default state
    });
    return () => {
      mounted = false;
    };
  }, []);

  const configuredPrefixes = useMemo(() => {
    if (!projectConfig?.prefixes) return [];
    return projectConfig.prefixes.map((p) => p.prefix);
  }, [projectConfig]);

  const handleSelectTicketFromList = (ticketId: string) => {
    setSelectedTicketId(ticketId);
    setActiveTab('deep-dive');
  };

  return (
    <div className="app-container">
      <Navbar
        activeTab={activeTab}
        onSelectTab={setActiveTab}
        projectConfig={projectConfig ?? undefined}
        projectName={projectConfig?.project_name}
      />
      <main className="main-content">
        {activeTab === 'landing' && (
          <LandingPage
            onNavigate={(tab) => setActiveTab(tab)}
            onSelectTicket={handleSelectTicketFromList}
            projectConfig={projectConfig}
            projectName={projectConfig?.project_name}
          />
        )}
        {activeTab === 'overview' && (
          <FleetOverview
            onSelectTicket={handleSelectTicketFromList}
            configuredPrefixes={configuredPrefixes}
          />
        )}
        {activeTab === 'deep-dive' && (
          <TicketDeepDive
            selectedTicketId={selectedTicketId}
            onSelectTicket={setSelectedTicketId}
          />
        )}
        {activeTab === 'progression' && (
          <MetricProgression
            configuredPrefixes={configuredPrefixes}
          />
        )}
        {activeTab === 'comparisons' && (
          <TicketComparisons
            onSelectTicket={handleSelectTicketFromList}
            configuredPrefixes={configuredPrefixes}
          />
        )}
        {activeTab === 'live' && (
          <LiveObservability
            onSelectTicket={handleSelectTicketFromList}
          />
        )}
      </main>
    </div>
  );
};
