import { configDefaults, defineConfig, mergeConfig } from 'vitest/config';
import viteConfig from './vite.config';
import { unitTestFiles } from './test-suites';

export default mergeConfig(viteConfig, defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    include: configDefaults.include,
    exclude: [...configDefaults.exclude, ...unitTestFiles],
  },
}));
