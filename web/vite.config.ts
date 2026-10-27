/// <reference types="vitest" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

function resolveBackendUrl(): string {
  const argv = process.argv;
  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i] ?? '';
    if (arg === '--backend-url' && i + 1 < argv.length) {
      return argv[i + 1] ?? 'http://localhost:8085';
    }
    if (arg.startsWith('--backend-url=')) {
      return arg.slice('--backend-url='.length) || 'http://localhost:8085';
    }
    if (arg === '--backend-port' && i + 1 < argv.length) {
      return `http://localhost:${argv[i + 1]}`;
    }
    if (arg.startsWith('--backend-port=')) {
      return `http://localhost:${arg.slice('--backend-port='.length)}`;
    }
  }

  if (process.env.BACKEND_URL) return process.env.BACKEND_URL;
  const port = process.env.BACKEND_PORT || process.env.VITE_BACKEND_PORT || process.env.PORT || '8085';
  return `http://localhost:${port}`;
}

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: resolveBackendUrl(),
        changeOrigin: true,
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
  },
});
