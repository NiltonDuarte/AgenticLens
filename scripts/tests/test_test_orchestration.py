"""Exercise supported test entrypoints against isolated real Go/Vitest units.

Only expensive boundary commands are marker probes. Unit failures come from real
assertions and the candidate Makefile/npm/configuration are copied unchanged.
"""
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
REAL_GO = shutil.which('go')
REAL_NPM = shutil.which('npm')


DISPATCH = r'''import json, os, pathlib, subprocess, sys
name = pathlib.Path(sys.argv[0]).name
args = sys.argv[1:]
log = pathlib.Path(os.environ['TEST_ORCHESTRATION_LOG'])
def record(event):
    with log.open('a') as stream:
        stream.write(json.dumps({'event': event, 'args': args}) + '\n')
if name == 'go':
    unit = args[:1] == ['test'] and '-race' not in args and './internal/domain/...' in args
    if unit:
        record('unit-go-start')
        result = subprocess.run([os.environ['TEST_ORCHESTRATION_GO'], *args])
        record('unit-go-pass' if result.returncode == 0 else 'unit-go-fail')
        sys.exit(result.returncode)
    if args[:1] == ['test']: record('integration-go')
    if args[:1] == ['build']: record('integration-native-build')
    if args == ['list', './...']: record('integration-discovery')
    sys.exit(subprocess.run([os.environ['TEST_ORCHESTRATION_GO'], *args]).returncode)
if name == 'npm':
    if args[:2] == ['run', 'build']:
        record('integration-assets')
        sys.exit(0)
    if args[:2] in [['run', 'typecheck'], ['run', 'lint']]:
        record('integration-static')
        sys.exit(0)
    if args[:1] == ['ci']:
        record('dependencies-start')
        pathlib.Path('node_modules/.package-lock.json').touch()
        record('dependencies-pass')
        sys.exit(0)
    if args[:2] == ['run', 'test:unit']:
        record('unit-frontend-start')
        result = subprocess.run([os.environ['TEST_ORCHESTRATION_NPM'], *args])
        record('unit-frontend-pass' if result.returncode == 0 else 'unit-frontend-fail')
        sys.exit(result.returncode)
    sys.exit(subprocess.run([os.environ['TEST_ORCHESTRATION_NPM'], *args]).returncode)
if name == 'python3':
    record('integration-python')
    sys.exit(0)
if name == 'cp':
    record('integration-compat-copy')
    sys.exit(73)
raise SystemExit('Unknown orchestration probe: ' + name)
'''


class OrchestrationFixture:
    def __init__(self, destination):
        self.root = destination
        self.log = destination / 'events.jsonl'
        subprocess.run(['git', 'init', '-q', str(destination)], check=True)
        self.env = dict(os.environ,
                        TEST_ORCHESTRATION_LOG=str(self.log),
                        TEST_ORCHESTRATION_GO=REAL_GO,
                        TEST_ORCHESTRATION_NPM=REAL_NPM,
                        NO_COLOR='1')
        for name in ['Makefile', 'scripts/go-test-packages.sh', 'scripts/opencode-compat.sh',
                     'web/package.json', 'web/package-lock.json', 'web/vite.config.ts',
                     'web/vitest.unit.config.ts', 'web/vitest.integration.config.ts',
                     'web/test-suites.ts', 'web/src/test/setup.ts']:
            target = destination / name
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(ROOT / name, target)
        modules = destination / 'web/node_modules'
        modules.mkdir()
        for entry in (ROOT / 'web/node_modules').iterdir():
            if entry.name not in ['.package-lock.json', '.vite', '.vite-temp']:
                (modules / entry.name).symlink_to(entry, target_is_directory=entry.is_dir())
        # A private freshness marker prevents npm ci from replacing linked deps.
        (modules / '.package-lock.json').write_text('{}\n')
        backend = destination / 'backend'
        backend.mkdir()
        (backend / 'go.mod').write_text('module agenticlens\n\ngo 1.26.0\n')
        core = backend / 'internal/domain'
        core.mkdir(parents=True)
        (core / 'add.go').write_text('package domain\nfunc Add(a, b int) int { return a + b }\n')
        migration = core / 'migration'
        migration.mkdir()
        (migration / 'migration.go').write_text('package migration\n')
        similar = backend / 'internal/domainish'
        similar.mkdir()
        (similar / 'similar.go').write_text('package domainish\n')
        boundary = backend / 'internal/boundary'
        boundary.mkdir()
        (boundary / 'boundary_test.go').write_text('package boundary\nimport "testing"\nfunc TestBoundary(t *testing.T) {}\n')
        application = backend / 'cmd/agentic-lens'
        application.mkdir(parents=True)
        (application / 'main.go').write_text('package main\nimport "fmt"\nfunc main() { fmt.Println("fixture-native") }\n')
        domain = destination / 'web/src/domain'
        domain.mkdir(parents=True)
        (domain / 'add.ts').write_text('export const add = (a: number, b: number): number => a + b;\n')
        integration = destination / 'web/src/components'
        integration.mkdir(parents=True)
        (integration / 'boundary.test.ts').write_text('''import { it, expect } from 'vitest';
import { appendFileSync } from 'node:fs';
it('executes the integration fixture', () => {
  const log = process.env.TEST_ORCHESTRATION_LOG;
  if (!log) throw new Error('missing fixture event log');
  appendFileSync(log, JSON.stringify({ event: 'integration-frontend' }) + '\\n');
  expect(2 + 2).toBe(4);
});
''')
        commands = destination / 'commands'
        commands.mkdir()
        for name in ['go', 'npm', 'python3', 'cp']:
            script = commands / name
            script.write_text('#!' + sys.executable + '\n' + DISPATCH)
            script.chmod(0o755)
        self.env['PATH'] = str(commands) + os.pathsep + os.environ['PATH']
        self.set_units()

    def set_units(self, failing=None):
        expected_go = 3 if failing == 'go' else 2
        (self.root / 'backend/internal/domain/add_test.go').write_text(
            'package domain\nimport "testing"\nfunc TestAdd(t *testing.T) { '
            f'if got := Add(1, 1); got != {expected_go} {{ t.Fatalf("fixture unit: got %d", got) }} }}\n')
        expected_frontend = 3 if failing == 'frontend' else 2
        (self.root / 'web/src/domain/add.test.ts').write_text(
            "import { it, expect } from 'vitest';\nimport { add } from './add';\n"
            f"it('adds fixture values', () => expect(add(1, 1)).toBe({expected_frontend}));\n")
        if self.log.exists():
            self.log.unlink()

    def run(self, args, frontend=False):
        return subprocess.run(args, cwd=self.root / 'web' if frontend else self.root,
                              env=self.env, text=True, capture_output=True, timeout=90)

    def events(self):
        return [json.loads(line)['event'] for line in self.log.read_text().splitlines()] if self.log.exists() else []


def jobs_from_text(text):
    """Parse the workflow's bounded jobs/needs subset, rejecting unknown shapes."""
    jobs = {}
    current = None
    collecting_needs = False
    in_jobs = False
    for line in text.splitlines():
        if line == 'jobs:':
            in_jobs = True
            continue
        if not in_jobs:
            continue
        if line.startswith('  ') and not line.startswith('   ') and line.strip().endswith(':'):
            current = line.strip()[:-1]
            jobs[current] = {'needs': [], 'lines': []}
            collecting_needs = False
            continue
        if current is None:
            continue
        jobs[current]['lines'].append(line)
        if line.startswith('    needs:'):
            value = line.split(':', 1)[1].strip()
            if value.startswith('[') and value.endswith(']'):
                jobs[current]['needs'] = [name.strip().strip("'\"") for name in value[1:-1].split(',') if name.strip()]
                collecting_needs = False
            elif not value:
                collecting_needs = True
            elif ' ' not in value:
                jobs[current]['needs'] = [value.strip("'\"")]
                collecting_needs = False
            else:
                raise AssertionError('Unsupported needs shape: ' + line)
        elif collecting_needs:
            if line.startswith('      - '):
                jobs[current]['needs'].append(line.strip()[2:].strip("'\""))
            elif line.strip():
                collecting_needs = False
    return jobs


def require_unit_dependencies(text):
    jobs = jobs_from_text(text)
    units = {'backend-unit', 'frontend-unit'}
    if not units <= jobs.keys():
        raise AssertionError('Both unit jobs must exist')
    for name, command in [('backend-unit', '        run: make test-unit-backend'),
                          ('frontend-unit', '        run: npm run test:unit --prefix web')]:
        if command not in jobs[name]['lines']:
            raise AssertionError('Unit job does not execute its unit suite: ' + name)
    def ancestors(name, visiting):
        if name in visiting:
            raise AssertionError('Cyclic job dependencies')
        if name not in jobs:
            raise AssertionError('Unknown dependency: ' + name)
        result = set(jobs[name]['needs'])
        for dependency in jobs[name]['needs']:
            result |= ancestors(dependency, visiting | {name})
        return result
    for name, job in jobs.items():
        if name in units:
            continue
        if not units <= ancestors(name, set()):
            raise AssertionError('Integration job lacks both unit gates: ' + name)
        if any(line.startswith('    if:') and any(call in line for call in ['always(', 'failure(', 'cancelled('])
               for line in job['lines']):
            raise AssertionError('Integration job bypasses success semantics: ' + name)
    return jobs


class TestOrchestration(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        if not REAL_GO or not REAL_NPM or not (ROOT / 'web/node_modules/vitest').is_dir():
            raise RuntimeError('Go, npm and installed frontend dependencies are required; run make install')

    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix='agenticlens-test-orchestration-')
        self.addCleanup(self.temporary.cleanup)
        self.fixture = OrchestrationFixture(Path(self.temporary.name))

    def assert_blocked(self, result, failed):
        events = self.fixture.events()
        self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertIn('unit-' + failed + '-fail', events, result.stdout + result.stderr)
        self.assertFalse([event for event in events if event.startswith('integration-')],
                         '\n'.join(events) + '\n' + result.stdout + result.stderr)

    def test_each_language_failure_blocks_make_entrypoints_and_parallel_mixed_goals(self):
        goals = [[], ['test-integration'], ['test-backend'], ['test-frontend'], ['backend-check'], ['check'],
                 ['-j8', 'check', 'test', 'web-assets', 'build']]
        for failed, marker in [('go', 'go'), ('frontend', 'frontend')]:
            for goal in goals:
                with self.subTest(failed=failed, goals=goal):
                    self.fixture.set_units(failed)
                    self.assert_blocked(self.fixture.run(['make', *goal]), marker)

    def test_standalone_npm_and_compatibility_commands_apply_their_unit_gates(self):
        for command in [['npm', 'test'], ['npm', 'run', 'test:integration'],
                        ['npm', 'test', '--', 'src/components/boundary.test.ts']]:
            with self.subTest(command=command):
                self.fixture.set_units('frontend')
                self.assert_blocked(self.fixture.run(command, frontend=True), 'frontend')
                # Frontend-only npm remains portable and does not run Go.
                self.fixture.set_units('go')
                result = self.fixture.run(command, frontend=True)
                events = self.fixture.events()
                self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
                self.assertIn('integration-frontend', events)
                self.assertLess(events.index('unit-frontend-pass'), events.index('integration-frontend'))
                self.assertFalse([event for event in events if event.startswith('unit-go-')], events)
        for failed in ['go', 'frontend']:
            with self.subTest(compatibility_failure=failed):
                self.fixture.set_units(failed)
                self.assert_blocked(self.fixture.run(['bash', 'scripts/opencode-compat.sh', 'v1']), failed)
        self.fixture.set_units()
        result = self.fixture.run(['bash', 'scripts/opencode-compat.sh', 'both'])
        events = self.fixture.events()
        # The copy probe deliberately stops before real provider/port startup.
        self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertEqual(events.count('integration-assets'), 1, events)
        self.assertIn('integration-compat-copy', events)
        assets = events.index('integration-assets')
        self.assertIn('unit-go-pass', events[:assets])
        self.assertIn('unit-frontend-pass', events[:assets])
        self.assertLess(assets, events.index('integration-compat-copy'), events)

    def test_successful_parallel_control_reaches_integration_only_after_both_units(self):
        os.utime(self.fixture.root / 'web/node_modules/.package-lock.json', (1, 1))
        result = self.fixture.run(['make', '-j8', 'install', 'check', 'test', 'web-assets', 'build',
                                  'PLATFORMS=linux/amd64'])
        events = self.fixture.events()
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertIn('integration-assets', events)
        self.assertIn('integration-go', events)
        self.assertIn('integration-frontend', events)
        self.assertIn('integration-python', events)
        for index, event in enumerate(events):
            if event.startswith('integration-'):
                self.assertIn('unit-go-pass', events[:index], events)
                self.assertIn('unit-frontend-pass', events[:index], events)
        self.assertEqual(events.count('integration-assets'), 1, events)
        self.assertEqual(events.count('dependencies-start'), 1, events)
        self.assertLess(events.index('dependencies-pass'), events.index('unit-frontend-start'))

    def test_build_only_and_installation_override_do_not_run_tests(self):
        for extra in [[], ['WEB_DIR=absent-web']]:
            with self.subTest(extra=extra):
                self.fixture.set_units('go')
                result = self.fixture.run(['make', 'verify-build', 'PLATFORMS=linux/amd64', *extra])
                self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
                self.assertFalse([event for event in self.fixture.events() if event.startswith('unit-')],
                                 self.fixture.events())

    def test_go_package_partition_is_complete_and_discovery_errors_fail(self):
        helper = self.fixture.root / 'scripts/go-test-packages.sh'
        lanes = {}
        for lane in ['all', 'unit', 'integration']:
            result = subprocess.run([str(helper), lane], cwd=self.fixture.root / 'backend',
                                    text=True, capture_output=True, timeout=30)
            self.assertEqual(result.returncode, 0, result.stderr)
            lanes[lane] = set(result.stdout.splitlines())
            self.assertTrue(lanes[lane])
        discovered = subprocess.check_output(['go', 'list', './...'], cwd=self.fixture.root / 'backend', text=True)
        self.assertEqual(lanes['all'], set(discovered.splitlines()))
        self.assertFalse(lanes['unit'] & lanes['integration'])
        self.assertEqual(lanes['all'], lanes['unit'] | lanes['integration'])
        self.assertEqual(lanes['unit'], {'agenticlens/internal/domain', 'agenticlens/internal/domain/migration'})
        self.assertIn('agenticlens/internal/domainish', lanes['integration'])
        unavailable = self.fixture.root / 'unavailable-go'
        unavailable.mkdir()
        failing = unavailable / 'go'
        failing.write_text('#!/bin/sh\nexit 23\n')
        failing.chmod(0o755)
        result = subprocess.run([str(helper), 'integration'], cwd=self.fixture.root / 'backend',
                                env=dict(os.environ, PATH=str(unavailable)), text=True, capture_output=True)
        self.assertNotEqual(result.returncode, 0)
        failing.write_text('#!/bin/sh\nif [ "$*" = "list -m" ]; then echo agenticlens; fi\n')
        result = subprocess.run([str(helper), 'integration'], cwd=self.fixture.root / 'backend',
                                env=dict(os.environ, PATH=str(unavailable)), text=True, capture_output=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('discovery was empty', result.stderr)

    def test_ci_and_release_graphs_require_units_and_detect_missing_edge(self):
        text = (ROOT / '.github/workflows/ci.yml').read_text()
        jobs = require_unit_dependencies(text)
        self.assertTrue({'windows-entrypoint', 'distribution-check', 'backend-check', 'frontend-check'} <= jobs.keys())
        mutated = text.replace('    needs: [backend-unit, frontend-unit]', '    needs: [frontend-unit]', 1)
        self.assertNotEqual(mutated, text)
        with self.assertRaisesRegex(AssertionError, 'lacks both unit gates'):
            require_unit_dependencies(mutated)
        from test_release_revision import steps
        for filename in ['release-please.yml', 'release-candidate.yml', 'release-stable.yml']:
            with self.subTest(workflow=filename):
                parsed = steps(ROOT / '.github/workflows' / filename)
                ordered = list(parsed)
                self.assertIn('Unit Tests', parsed)
                unit = parsed['Unit Tests']
                self.assertIn('        run: make test-unit', unit)
                for name, lines in parsed.items():
                    if '        run: make check' in lines or '        run: make test' in lines:
                        self.assertLess(ordered.index('Unit Tests'), ordered.index(name))
                        unit_condition = [line for line in unit if line.startswith('        if:')]
                        integration_condition = [line for line in lines if line.startswith('        if:')]
                        self.assertEqual(unit_condition, integration_condition)


if __name__ == '__main__':
    unittest.main()
