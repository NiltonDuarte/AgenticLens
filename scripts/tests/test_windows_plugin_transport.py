"""Execute the shipped plugin transport on an actual Windows Node host.

This is a transport fixture, not an OpenCode/Bun lifecycle support claim. Existing
OpenCode supplies the production runtime; TypeScript compilation is test tooling.
"""
import json
import os
from pathlib import Path
import shutil
import subprocess
import time
import unittest

import test_windows_distribution as windows_boundary
from distribution_fixture import ROOT

IS_WINDOWS = windows_boundary.IS_WINDOWS
NODE = os.environ.get('AL_WINDOWS_NODE', shutil.which('node.exe') if IS_WINDOWS else '')
if not NODE:
    raise SystemExit('Actual Windows Node required; provide AL_WINDOWS_NODE explicitly for WSL transport probes')


class WindowsPluginTransportTest(unittest.TestCase):
    def setUp(self):
        self.fixture = windows_boundary.WindowsDistributionTest('test_real_install_quote_argv_stdin_exit_and_local_overlay')
        self.fixture.setUp()
        self.addCleanup(self.fixture.doCleanups)
        self.fixture.succeeds(self.fixture.install())
        self.root = self.fixture.root
        self.env = dict(self.fixture.env)
        if os.environ.get('AL_WINDOWS_NODE_RUN_AS_NODE') == '1':
            self.env['ELECTRON_RUN_AS_NODE'] = '1'
        self.pid_file = self.root / 'transport-native.pid'
        self.parent_file = self.root / 'transport-powershell.pid'
        self.payload_file = self.root / 'payload.txt'
        self.payload_file.write_bytes('café payload\n'.encode('utf-8'))
        self.env.update(AL_PROBE_PID=windows_boundary.win(self.pid_file), AL_PROBE_PARENT_PID=windows_boundary.win(self.parent_file))
        self.module = self.root / 'transport.mjs'
        source = (ROOT / 'backend/internal/adapters/capture/plugin.ts').read_text().split('const base =', 1)[0]
        source = '\n'.join(line for line in source.splitlines() if '@opencode-ai/plugin' not in line)
        for marker, path in [('__WRAPPER__', self.fixture.dest / 'agentic-lens.ps1'), ('__PROJECT__', self.fixture.repo),
                             ('__STATE__', self.root / 'capture state'), ('__STORE__', self.root / 'source store')]:
            source = source.replace(marker, json.dumps(windows_boundary.win(path)))
        source += '''
import { readFileSync } from 'node:fs';
const started = Date.now();
try {
  const payload = readFileSync(process.argv[3], 'utf8');
  const output = process.argv[4] === 'direct'
    ? await callWindows(process.argv[5], JSON.parse(process.argv[2]), payload)
    : await call(JSON.parse(process.argv[2]), payload);
  console.log(JSON.stringify({ok:true, output, elapsed:Date.now()-started, hostPid:process.pid, node:process.version, platform:process.platform}));
} catch (error) {
  console.log(JSON.stringify({ok:false, error:error instanceof Error ? error.message : String(error), elapsed:Date.now()-started, hostPid:process.pid, node:process.version, platform:process.platform}));
}
'''
        typescript = self.root / 'transport.ts'
        typescript.write_text(source, encoding='utf-8')
        compiler = ROOT / 'web/node_modules/typescript/lib/typescript.js'
        if not compiler.is_file():
            raise RuntimeError('Install repository development dependencies before transport tests: npm ci --prefix web')
        subprocess.run(['node', '-e', "const fs=require('fs'),ts=require(process.argv[1]);fs.writeFileSync(process.argv[3],ts.transpileModule(fs.readFileSync(process.argv[2],'utf8'),{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.ESNext}}).outputText)",
                        str(compiler), str(typescript), str(self.module)], check=True, capture_output=True, timeout=15)
        self.sibling_pid_file = self.root / 'unrelated-sibling.pid'
        sibling_script = self.root / 'sibling.mjs'
        sibling_script.write_text("import{writeFileSync}from'node:fs';writeFileSync(process.argv[2],String(process.pid));console.log('ready');setInterval(()=>{},1000);", encoding='utf-8')
        self.sibling = subprocess.Popen([NODE, windows_boundary.win(sibling_script), windows_boundary.win(self.sibling_pid_file)],
                                        env=self.environment(), stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        self.addCleanup(self.stop_sibling)
        self.assertEqual(self.sibling.stdout.readline().strip(), b'ready')
        self.sibling_pid = int(self.sibling_pid_file.read_text())

    def environment(self, **extra):
        result = dict(self.env, **extra)
        if not IS_WINDOWS:
            forwarded = [key for key in result if key.startswith(('AL_PROBE_', 'AGENTIC_LENS_')) or key in ('ELECTRON_RUN_AS_NODE', 'SystemRoot', 'HOME', 'USERPROFILE', 'CODEX_HOME', 'XDG_CONFIG_HOME', 'XDG_CACHE_HOME', 'XDG_STATE_HOME')]
            result['WSLENV'] = ':'.join(filter(None, [result.get('WSLENV', ''), *forwarded]))
        return result

    def stop_sibling(self):
        if self.sibling_pid_file.exists():
            pid = int(self.sibling_pid_file.read_text())
            subprocess.run([windows_boundary.POWERSHELL, '-NoLogo', '-NoProfile', '-NonInteractive', '-Command',
                            f'Stop-Process -Id {pid} -Force -ErrorAction SilentlyContinue'], capture_output=True, timeout=10)
        if self.sibling.poll() is None:
            try:
                self.sibling.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.sibling.kill(); self.sibling.wait(timeout=5)
        self.sibling.stdout.close(); self.sibling.stderr.close()

    def invoke(self, mode, direct=None, **environment):
        args = ['receive', '--source', 'opencode'] if direct is None else ['capture', 'receive']
        command = [NODE, windows_boundary.win(self.module), json.dumps(args), windows_boundary.win(self.payload_file)]
        if direct is not None:
            command += ['direct', windows_boundary.win(direct)]
        started = time.monotonic()
        result = subprocess.run(command, env=self.environment(AL_PROBE_TRANSPORT_MODE=mode, **environment),
                                input='', text=True, encoding='utf-8', capture_output=True, timeout=8)
        wall_time = time.monotonic() - started
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        records = [json.loads(line) for line in result.stdout.splitlines() if line.startswith('{')]
        self.assertEqual(len(records), 1, result.stdout + result.stderr)
        report = records[0]
        self.assertEqual(report['platform'], 'win32')
        self.assertLess(report['elapsed'], 5700, report)
        self.assertLess(wall_time, 7.5)
        return report, result.stderr

    def assert_owned_tree_stopped(self, parent=True):
        self.assertTrue(self.pid_file.exists(), 'native process was not observed')
        pids = [int(self.pid_file.read_text())]
        if parent:
            self.assertTrue(self.parent_file.exists())
            pids.append(int(self.parent_file.read_text()))
        command = '; '.join(f'if (Get-Process -Id {pid} -ErrorAction SilentlyContinue) {{ throw "owned process {pid} survived" }}' for pid in pids)
        command += f'; if (-not (Get-Process -Id {self.sibling_pid} -ErrorAction SilentlyContinue)) {{ throw "unrelated sibling lost" }}'
        result = subprocess.run([windows_boundary.POWERSHELL, '-NoLogo', '-NoProfile', '-NonInteractive', '-Command', command],
                                capture_output=True, timeout=10)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_normal_delayed_powershell_handoff_preserves_utf8_and_arguments(self):
        report, error = self.invoke('delay')
        self.assertTrue(report['ok'], report)
        self.assertGreater(report['elapsed'], 1500)
        output = json.loads(report['output'])
        self.assertEqual(output['stdin'], 'café payload\n')
        self.assertEqual(output['args'][:3], ['capture', 'receive', '--source'])
        self.assertIn(windows_boundary.win(self.fixture.repo), output['args'])
        self.assertEqual(error, '')

    def test_deadline_rejects_and_kills_powershell_and_native_without_pipe_wait(self):
        report, _ = self.invoke('hang')
        self.assertFalse(report['ok'], report)
        self.assertIn('timed out', report['error'])
        self.assertGreaterEqual(report['elapsed'], 4800)
        self.assert_owned_tree_stopped()
        print(json.dumps({'case':'timeout','node':report['node'],'elapsed_ms':report['elapsed'],
                          'native_pid':int(self.pid_file.read_text()),'powershell_pid':int(self.parent_file.read_text()),
                          'unrelated_sibling_alive':True}))

    def test_stdout_overflow_is_bounded_and_kills_owned_tree(self):
        report, _ = self.invoke('stdout-overflow')
        self.assertFalse(report['ok'], report)
        self.assertIn('output exceeded', report['error'])
        self.assert_owned_tree_stopped()

    def test_stderr_overflow_is_bounded_and_kills_owned_tree(self):
        report, _ = self.invoke('stderr-overflow')
        self.assertFalse(report['ok'], report)
        self.assertIn('diagnostics exceeded', report['error'])
        self.assert_owned_tree_stopped()

    def test_stdin_error_and_spawn_error_do_not_crash_host(self):
        self.payload_file.write_text('p' * (2 * 1024 * 1024), encoding='utf-8')
        report, _ = self.invoke('stdin-close', direct=self.fixture.native())
        self.assertFalse(report['ok'], report)
        self.assertNotIn('timed out', report['error'])
        self.assertLess(report['elapsed'], 3000, report)
        self.assert_owned_tree_stopped(parent=False)
        report, _ = self.invoke('', direct=self.root / 'missing-native.exe')
        self.assertFalse(report['ok'], report)
        self.assertIn('ENOENT', report['error'])

    def test_cleanup_spawn_failure_is_explicit_without_unhandled_error(self):
        report, error = self.invoke('hang', direct=self.fixture.native(), SystemRoot='C:\\missing-system-root')
        self.assertFalse(report['ok'], report)
        self.assertIn('cleanup unavailable', error)
        self.assert_owned_tree_stopped(parent=False)


if __name__ == '__main__':
    unittest.main(verbosity=2)
