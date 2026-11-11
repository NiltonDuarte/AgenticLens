"""Shared real-native fixtures for versioned release/entrypoint boundary tests."""
import hashlib
import json
import os
from pathlib import Path
import platform
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
SETUP = ROOT / '.agents/skills/agentic-lens-setup/scripts'
PROBE = ROOT / 'scripts/tests/fixtures/entrypoint_probe.go'
COMPILED = tempfile.TemporaryDirectory(prefix='agenticlens native fixtures ')


def build_probe(version='v9.8.7', operating_system=None, arch=None, protocol='3'):
    operating_system = operating_system or {'Linux': 'linux', 'Darwin': 'darwin'}[platform.system()]
    arch = arch or {'x86_64': 'amd64', 'amd64': 'amd64', 'aarch64': 'arm64', 'arm64': 'arm64'}[platform.machine().lower()]
    target = Path(COMPILED.name) / (version + '-p' + protocol + '-' + operating_system + '-' + arch)
    if not target.exists():
        subprocess.run(['go', 'build', '-ldflags', '-X main.version=' + version + ' -X main.protocol=' + protocol, '-o', str(target), str(PROBE)],
                       cwd=ROOT, env=dict(os.environ, GOOS=operating_system, GOARCH=arch), check=True, capture_output=True, timeout=120)
    return target


def digest(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def build_app(version='v9.8.7', operating_system=None, arch=None):
    """Build actual application fixtures only after the canonical web asset build."""
    if not (ROOT / 'backend/internal/adapters/web/dist/index.html').is_file():
        raise RuntimeError('Actual application fixtures require canonical assets: npm run build --prefix web')
    operating_system = operating_system or {'Linux': 'linux', 'Darwin': 'darwin'}[platform.system()]
    arch = arch or {'x86_64': 'amd64', 'amd64': 'amd64', 'aarch64': 'arm64', 'arm64': 'arm64'}[platform.machine().lower()]
    target = Path(COMPILED.name) / ('application-' + version + '-' + operating_system + '-' + arch)
    if not target.exists():
        subprocess.run(['go', 'build', '-ldflags', '-X agenticlens/internal/domain.Version=' + version, '-o', str(target), './cmd/agentic-lens'],
                       cwd=ROOT / 'backend', env=dict(os.environ, GOOS=operating_system, GOARCH=arch, CGO_ENABLED='0'),
                       check=True, capture_output=True, timeout=180)
    return target


def build_manager(version='v9.8.7', operating_system=None, arch=None):
    operating_system = operating_system or {'Linux': 'linux', 'Darwin': 'darwin'}[platform.system()]
    arch = arch or {'x86_64': 'amd64', 'amd64': 'amd64', 'aarch64': 'arm64', 'arm64': 'arm64'}[platform.machine().lower()]
    target = Path(COMPILED.name) / ('manager-' + version + '-' + operating_system + '-' + arch)
    if not target.exists():
        subprocess.run(['go', 'build', '-ldflags', '-X agenticlens/internal/domain.Version=' + version, '-o', str(target), './cmd/agentic-lens-manager'],
                       cwd=ROOT / 'backend', env=dict(os.environ, GOOS=operating_system, GOARCH=arch, CGO_ENABLED='0'),
                       check=True, capture_output=True, timeout=180)
    return target


def shared_policy(pin, cache, workspace='workspaces', name='Fixture'):
    """Explicit approved current policy; never rely on runtime numerical presets."""
    body = (ROOT / 'scripts/tests/fixtures/project-policy.yaml').read_text()
    body = body.replace('project_name: Fixture', 'project_name: ' + json.dumps(name))
    body = body.replace('version: v9.8.7', 'version: ' + pin + '\n  install_root: ' + json.dumps(str(cache)))
    return body.replace('dir: records', 'dir: ' + json.dumps(str(workspace)))


def local_policy(state, source, cache=None):
    install = 'agenticlens:\n  install_root: ' + json.dumps(str(cache)) + '\n' if cache else ''
    return install + ('telemetry:\n  default_source: codex\n  sources:\n    codex:\n      home: ' + json.dumps(str(source)) +
                      '\n  capture:\n    enabled: false\n    state_dir: ' + json.dumps(str(state)) +
                      '\n    max_sessions: 19\n    max_source_bytes: 16777217\n')


class DistributionFixture(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        # Compile once before filesystem cases so build failures stay bounded.
        cls.fixture_probe = build_probe()
        cls.fixture_application = build_app()
        cls.fixture_manager = build_manager()

    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix='agenticlens distribution ')
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.repo = self.root / 'adopting project'
        self.repo.mkdir()
        self.home = self.root / 'isolated user home'; self.home.mkdir()
        self.cache = self.root / 'native cache'
        self.dest = self.root / 'user PATH bin'
        self.config = self.repo / '.agenticlens-config.yaml'
        self.pin = 'v9.8.7'
        self.write_config()
        self.write_local_policy()
        self.probe = self.fixture_probe
        self.application = self.fixture_application
        self.manager = self.fixture_manager
        self.network_log = self.root / 'fetch-log'
        commands = self.root / 'commands'
        commands.mkdir()
        curl = commands / 'curl'
        curl.write_text('#!' + __import__('sys').executable + '\n' + '''
import hashlib
import json, os, pathlib, shutil, sys
args = sys.argv[1:]
url = args[-1]
with open(os.environ['AL_FETCH_LOG'], 'a') as log: log.write(url + '\\n')
target = pathlib.Path(args[args.index('-o')+1])
if os.environ.get('AL_FETCH_FAIL', 'UNMATCHABLE') in url:
    target.write_text('partial transfer'); sys.exit(23)
if '/.agents/skills/' in url:
    shutil.copyfile(pathlib.Path(os.environ['AL_FIXTURE_SOURCE']) / '.agents/skills' / url.split('/.agents/skills/', 1)[1], target)
elif url.endswith('/SHA256SUMS'):
    release = pathlib.Path(os.environ['AL_FIXTURE_RELEASE'])
    checksum = hashlib.sha256(release.read_bytes()).hexdigest()
    if os.environ.get('AL_FETCH_BAD_CHECKSUM') == '1': checksum = '0' * 64
    target.write_text(checksum + '  bin/' + os.environ['AL_FIXTURE_NAME'] + '\\n')
else: shutil.copyfile(os.environ['AL_FIXTURE_RELEASE'], target)
''')
        curl.chmod(0o755)
        self.env = {key: value for key, value in os.environ.items() if not key.startswith(('AGENTIC_LENS_', 'AL_PROBE_'))}
        self.env.update(HOME=str(self.home), USERPROFILE=str(self.home), CODEX_HOME=str(self.home / '.codex'),
                        XDG_CONFIG_HOME=str(self.home / '.config'), XDG_CACHE_HOME=str(self.home / '.cache'),
                        XDG_STATE_HOME=str(self.home / '.state'), XDG_DATA_HOME=str(self.home / '.local/share'), PATH=str(commands) + os.pathsep + os.environ['PATH'], AL_FETCH_LOG=str(self.network_log),
                        AL_FIXTURE_SOURCE=str(ROOT), AL_FIXTURE_RELEASE=str(self.application), AL_FIXTURE_NAME=self.native().name)

    def write_config(self, pin=None):
        self.pin = pin or self.pin
        self.config.write_text(shared_policy(self.pin, self.cache))

    def write_local_policy(self):
        path = self.repo / '.agenticlens-config.local.yaml'
        path.write_text(local_policy(self.root / 'capture state', self.home / '.codex'))
        return path

    def approved_settings(self, name='Fixture'):
        path = self.root / 'approved shared.yaml'
        path.write_text(shared_policy(self.pin, self.cache, name=name))
        return path

    def native(self, pin=None):
        operating_system = {'Linux': 'linux', 'Darwin': 'darwin'}[platform.system()]
        arch = {'x86_64': 'amd64', 'amd64': 'amd64', 'aarch64': 'arm64', 'arm64': 'arm64'}[platform.machine().lower()]
        name = f'agentic-lens-engine-{operating_system}-{arch}'
        return self.cache / 'bin' / (pin or self.pin) / f'{operating_system}-{arch}' / name

    def install(self, *arguments, script=None, env=None):
        return subprocess.run(['bash', str(script or SETUP / 'download.sh'), '--project', str(self.repo), '--dest', str(self.dest), *map(str, arguments)],
                              cwd=self.repo, env=env or self.env, input='', capture_output=True, text=True, timeout=30)

    def install_local(self, *arguments, script=None, probe=None, env=None):
        probe = probe or self.application
        return self.install('--local-manager-bin', self.manager, '--manager-sha256', digest(self.manager),
                            '--local-bin', probe, '--sha256', digest(probe), *arguments, script=script, env=env)

    def launch(self, *arguments, cwd=None, env=None, input='payload\n'):
        return subprocess.run([str(self.dest / 'agentic-lens'), *map(str, arguments)], cwd=cwd or self.repo,
                              env=env or self.env, input=input, capture_output=True, text=True, timeout=10)

    def succeeds(self, result):
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        return result

    def requests(self):
        return self.network_log.read_text().splitlines() if self.network_log.exists() else []
