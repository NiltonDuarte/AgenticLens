import { cpSync, rmSync } from 'node:fs';

const source = new URL('../dist/', import.meta.url);
const destination = new URL('../../backend/internal/adapters/web/dist/', import.meta.url);

rmSync(destination, { recursive: true, force: true });
cpSync(source, destination, { recursive: true });
