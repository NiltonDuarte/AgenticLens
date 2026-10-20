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
  const [isMock, setIsMock] = useState(false);
  const [activeTab, setActiveTab] = useState<ActiveTab>('landing');
  const [selectedTicketId, setSelectedTicketId] = useState<string | null>('AL-1');
  const [projectConfig, setProjectConfig] = useState<ProjectConfigResponse | null>(null);

  useEffect(() => {
    let mounted = true;
    void fetchProjectConfig().then((resp) => {
      if (mounted) {
        setProjectConfig(resp.data);
      }
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
        isMock={isMock}
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
            onMockChange={setIsMock}
            projectConfig={projectConfig}
            projectName={projectConfig?.project_name}
          />
        )}
        {activeTab === 'overview' && (
          <FleetOverview
            onMockChange={setIsMock}
            onSelectTicket={handleSelectTicketFromList}
            configuredPrefixes={configuredPrefixes}
          />
        )}
        {activeTab === 'deep-dive' && (
          <TicketDeepDive
            selectedTicketId={selectedTicketId}
            onSelectTicket={setSelectedTicketId}
            onMockChange={setIsMock}
          />
        )}
        {activeTab === 'progression' && (
          <MetricProgression
            onMockChange={setIsMock}
            configuredPrefixes={configuredPrefixes}
          />
        )}
        {activeTab === 'comparisons' && (
          <TicketComparisons
            onSelectTicket={handleSelectTicketFromList}
            onMockChange={setIsMock}
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
