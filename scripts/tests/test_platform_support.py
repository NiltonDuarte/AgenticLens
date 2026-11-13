"""Execute supported distribution boundaries using isolated filesystem fixtures."""
import hashlib
import os
from pathlib import Path
import subprocess
import tarfile
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
TARGETS = ('linux-amd64', 'linux-arm64', 'darwin-amd64', 'darwin-arm64')


class PlatformSupportTest(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix='agenticlens-platform-')
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)

    def test_unsupported_bootstrap_selection_rejects_before_fetch_or_writes(self):
        project = self.root / "owner's project"
        project.mkdir()
        policy = project / '.agenticlens-config.yaml'
        original = b'agenticlens:\n  version: v9.8.7\n'
        policy.write_bytes(original)
        home = self.root / 'home'
        home.mkdir()
        destination = self.root / 'user bin'
        fetch_log = self.root / 'fetch-log'
        commands = self.root / 'commands'
        commands.mkdir()
        for name in ('curl', 'wget'):
            executable = commands / name
            executable.write_text('#!/bin/sh\nprintf attempted >> "$AL_TEST_FETCH_LOG"\nexit 91\n')
            executable.chmod(0o700)
        environment = {key: value for key, value in os.environ.items()
                       if not key.startswith(('AGENTIC_LENS_', 'AL_PROBE_'))}
        environment.update(HOME=str(home), PATH=str(commands) + os.pathsep + os.environ['PATH'],
                           AL_TEST_FETCH_LOG=str(fetch_log))
        for platform in ('windows-amd64', 'windows/arm64', 'freebsd-amd64', 'linux-386'):
            for additional in ([], ['--all-platforms']):
                with self.subTest(platform=platform, flags=additional):
                    result = subprocess.run(
                        ['bash', str(ROOT / '.agents/skills/agentic-lens-setup/scripts/download.sh'),
                         '--project', str(project), '--dest', str(destination),
                         '--manager-version', 'v9.8.7', '--include-host', platform, *additional],
                        env=environment, capture_output=True, text=True, timeout=15)
                    self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
                    self.assertIn('Unsupported platform', result.stderr)
                    self.assertFalse(fetch_log.exists(), 'unsupported selection attempted a fetch')
                    self.assertFalse(destination.exists())
                    self.assertEqual(list(home.iterdir()), [])
                    self.assertEqual(list(project.iterdir()), [policy])
                    self.assertEqual(policy.read_bytes(), original)

    def test_four_platform_tar_packages_have_complete_verifiable_checksums(self):
        distribution = self.root / 'dist'
        binaries = distribution / 'bin'
        binaries.mkdir(parents=True)
        (self.root / 'README.md').write_text('Fixture release readme\n')
        (self.root / 'LICENSE').write_text('Fixture license\n')
        expected_binaries = {}
        for platform in TARGETS:
            for kind in ('agentic-lens', 'agentic-lens-engine'):
                name = f'{kind}-{platform}'
                body = f'fixture bytes for {name}\n'.encode()
                (binaries / name).write_bytes(body)
                expected_binaries['bin/' + name] = body
        stale_files = {
            'bin/agentic-lens-windows-amd64.exe': b'previous Windows manager',
            'bin/agentic-lens-engine-windows-amd64.exe': b'previous Windows engine',
            'bin/agentic-lens-freebsd-amd64': b'unsupported manager',
            'agentic-lens-v9.8.7-windows-amd64.zip': b'previous Windows archive',
            'agentic-lens-v9.8.7-freebsd-amd64.tar.gz': b'unsupported archive',
        }
        for name, body in stale_files.items():
            (distribution / name).write_bytes(body)
        expected_archives = {f'agentic-lens-v9.8.7-{platform}.tar.gz' for platform in TARGETS}
        for execution in ('fresh', 'repeated'):
            with self.subTest(execution=execution):
                result = subprocess.run(
                    ['make', '-f', str(ROOT / 'Makefile'), 'release-package',
                     'VERSION=v9.8.7', 'GIT_COMMIT=fixture', 'BUILD_DATE=fixture',
                     'WEB_DIR=absent-web', 'RELEASE_DIR=dist'],
                    cwd=self.root, capture_output=True, text=True, timeout=30)
                self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
                self.assertEqual({path.name for path in distribution.iterdir()},
                                 expected_archives | {'bin', 'SHA256SUMS'} |
                                 {name for name in stale_files if '/' not in name})
                expected_files = set(expected_binaries) | expected_archives
                entries = {}
                for line in (distribution / 'SHA256SUMS').read_text().splitlines():
                    checksum, name = line.split(maxsplit=1)
                    self.assertNotIn(name, entries, 'duplicate checksum entry')
                    entries[name] = checksum
                    self.assertEqual(hashlib.sha256((distribution / name).read_bytes()).hexdigest(), checksum)
                self.assertEqual(set(entries), expected_files)
                for name, body in stale_files.items():
                    self.assertEqual((distribution / name).read_bytes(), body,
                                     'packaging changed unrelated stale files')
                # Evaluate the configured upload selectors against this same
                # stale distribution tree, matching the release action's paths.
                for workflow in ('release-please.yml', 'release-candidate.yml', 'release-stable.yml'):
                    source = (ROOT / '.github/workflows' / workflow).read_text().splitlines()
                    self.assertEqual(sum(line.strip() == 'files: |' for line in source), 1)
                    start = next(index for index, line in enumerate(source) if line.strip() == 'files: |') + 1
                    patterns = []
                    for line in source[start:]:
                        if not line.strip():
                            break
                        if not line.startswith('            '):
                            break
                        patterns.append(line.strip().replace('${{ env.VERSION }}', 'v9.8.7'))
                    self.assertTrue(patterns, 'release upload selection is empty')
                    selected = set()
                    for pattern in patterns:
                        matches = list(self.root.glob(pattern))
                        self.assertTrue(matches, f'{workflow}: upload pattern has no artifact: {pattern}')
                        selected.update(str(path.relative_to(distribution)) for path in matches)
                    self.assertEqual(selected, expected_files | {'SHA256SUMS'}, workflow)
                for platform in TARGETS:
                    with tarfile.open(distribution / f'agentic-lens-v9.8.7-{platform}.tar.gz') as archive:
                        names = [f'agentic-lens-{platform}', f'agentic-lens-engine-{platform}']
                        self.assertEqual(set(archive.getnames()), set(names) | {'README.md', 'LICENSE'})
                        for name in names:
                            with archive.extractfile(name) as binary:
                                self.assertEqual(binary.read(), expected_binaries['bin/' + name])


if __name__ == '__main__':
    unittest.main()
