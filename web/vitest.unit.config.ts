import { defineConfig, mergeConfig } from 'vitest/config';
import viteConfig from './vite.config';
import { unitTestFiles } from './test-suites';

export default mergeConfig(viteConfig, defineConfig({
  test: {
    globals: true,
    environment: 'node',
    include: unitTestFiles,
    setupFiles: [],
  },
}));
