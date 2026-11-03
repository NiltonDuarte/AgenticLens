// @vitest-environment node

import { execFileSync } from 'node:child_process';
import { copyFileSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { expect, test } from 'vitest';

test('synchronizes the complete embedded asset tree independently of the caller directory', () => {
  const root = mkdtempSync(join(tmpdir(), 'agenticlens-assets ü '));
  try {
    const scripts = join(root, 'web', 'scripts');
    const source = join(root, 'web', 'dist');
    const destination = join(root, 'backend', 'internal', 'adapters', 'web', 'dist');
    mkdirSync(scripts, { recursive: true });
    mkdirSync(join(source, 'assets'), { recursive: true });
    mkdirSync(join(destination, 'assets'), { recursive: true });
    copyFileSync(fileURLToPath(new URL('../../scripts/sync-assets.mjs', import.meta.url)), join(scripts, 'sync-assets.mjs'));
    writeFileSync(join(source, 'index.html'), '<html>current build</html>');
    writeFileSync(join(source, 'assets', 'current.js'), 'current bundle');
    writeFileSync(join(source, '.metadata'), 'current metadata');
    writeFileSync(join(destination, 'assets', 'stale.js'), 'stale bundle');
    writeFileSync(join(destination, '.stale'), 'stale metadata');

    const synchronize = () => execFileSync(process.execPath, [join(scripts, 'sync-assets.mjs')], { cwd: root });
    synchronize();
    expect(readdirSync(destination).sort()).toEqual(['.metadata', 'assets', 'index.html']);
    expect(readdirSync(join(destination, 'assets'))).toEqual(['current.js']);
    expect(readFileSync(join(destination, 'index.html'), 'utf8')).toBe('<html>current build</html>');
    expect(readFileSync(join(destination, 'assets', 'current.js'), 'utf8')).toBe('current bundle');
    expect(readFileSync(join(destination, '.metadata'), 'utf8')).toBe('current metadata');

    rmSync(join(source, 'assets', 'current.js'));
    writeFileSync(join(source, 'assets', 'next.js'), 'next bundle');
    synchronize();
    expect(readdirSync(join(destination, 'assets'))).toEqual(['next.js']);
    expect(readFileSync(join(destination, 'assets', 'next.js'), 'utf8')).toBe('next bundle');
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});
