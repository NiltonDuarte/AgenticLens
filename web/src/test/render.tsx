import React from 'react';
import { render as renderComponent, type RenderOptions } from '@testing-library/react';
import { ProjectConfigContext } from '../contexts/ProjectConfigContext';
import { mockProjectConfig } from './fixtures/mockData';

export * from '@testing-library/react';

/** Component fixtures receive the same explicit project policy as the application. */
export function render(ui: React.ReactNode, options?: RenderOptions) {
  return renderComponent(ui, {
    wrapper: ({children}) => <ProjectConfigContext.Provider value={mockProjectConfig}>{children}</ProjectConfigContext.Provider>,
    ...options,
  });
}
