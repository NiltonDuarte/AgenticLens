import React, { useState } from 'react';
import type { ActiveTab } from './domain/types';
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
      />
      <main className="main-content">
        {activeTab === 'landing' && (
          <LandingPage
            onNavigate={(tab) => setActiveTab(tab)}
            onSelectTicket={handleSelectTicketFromList}
            onMockChange={setIsMock}
          />
        )}
        {activeTab === 'overview' && (
          <FleetOverview
            onMockChange={setIsMock}
            onSelectTicket={handleSelectTicketFromList}
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
          />
        )}
        {activeTab === 'comparisons' && (
          <TicketComparisons
            onSelectTicket={handleSelectTicketFromList}
            onMockChange={setIsMock}
          />
        )}
        {activeTab === 'live' && (
          <LiveObservability
            onSelectTicket={handleSelectTicketFromList}
            onMockChange={setIsMock}
          />
        )}
      </main>
    </div>
  );
};
