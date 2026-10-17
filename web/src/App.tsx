import React, { useState } from 'react';
import { Navbar } from './components/Navbar';
import { FleetOverview } from './views/FleetOverview';

export const App: React.FC = () => {
  const [isMock, setIsMock] = useState(false);
  const [activeTab, setActiveTab] = useState<'overview' | 'detail'>('overview');

  return (
    <div className="app-container">
      <Navbar
        isMock={isMock}
        activeTab={activeTab}
        onSelectTab={setActiveTab}
      />
      <main className="main-content">
        {activeTab === 'overview' && (
          <FleetOverview
            onMockChange={setIsMock}
            onSelectTicket={(_ticketId) => {
              // Deep dive view is slated for AL-4
            }}
          />
        )}
      </main>
    </div>
  );
};
