"""Installed skill bundle, exact-pin repair and user-level migration boundaries."""
import json
import os
from pathlib import Path
import shutil
import shlex
import subprocess
import unittest

from distribution_fixture import DistributionFixture, ROOT, SETUP, digest, build_app, shared_policy

INSTALL = SETUP / 'install.sh'
BUNDLE_SETUP = Path('.agents/skills/agentic-lens-setup')
BUNDLE_OPERATOR = Path('.agents/skills/agentic-lens-operator')


class InstallationTest(DistributionFixture):
    def test_source_build_verifies_new_native_without_changing_selected_release_cache(self):
        self.succeeds(self.install_local())
        cached = self.native(); before = cached.stat(); cached_bytes = cached.read_bytes()
        destination = self.root / 'source build bin'
        host_os = subprocess.check_output(['go', 'env', 'GOHOSTOS'], text=True).strip()
        host_arch = subprocess.check_output(['go', 'env', 'GOHOSTARCH'], text=True).strip()
        environment = dict(self.env, GOCACHE=subprocess.check_output(['go', 'env', 'GOCACHE'], text=True).strip(),
                           GOMODCACHE=subprocess.check_output(['go', 'env', 'GOMODCACHE'], text=True).strip())
        for key in ('VERSION', 'GIT_COMMIT', 'BUILD_DATE'):
            environment.pop(key, None)
        native = destination / f'agentic-lens-engine-{host_os}-{host_arch}'
        source_version = subprocess.check_output(['git', 'describe', '--tags', '--always', '--dirty'], cwd=ROOT, text=True).strip()
        candidate = {'VERSION': 'v9.8.8-rc.1+abcdef0', 'GIT_COMMIT': 'abcdef0', 'BUILD_DATE': '2026-01-02T03:04:05Z'}
        for name, metadata, expected in [('default', {}, source_version), ('candidate', candidate, candidate['VERSION'])]:
            with self.subTest(metadata=name):
                build_environment = dict(environment, **metadata)
                result = subprocess.run(['make', 'verify-build', 'BIN_DIR=' + str(destination),
                                         'WEB_DIR=' + str(self.root / 'no web dependency scan'), 'PLATFORMS=' + host_os + '/' + host_arch],
                                        cwd=ROOT, env=build_environment, text=True, capture_output=True, timeout=180)
                self.succeeds(result)
                actual = subprocess.check_output([str(native), 'version', '--short'], env=build_environment, text=True).strip()
                self.assertEqual(actual, expected)
                self.assertNotEqual(actual, self.pin)
                self.assertIn(expected, result.stdout)
                if metadata:
                    info = json.loads(subprocess.check_output([str(native), 'version', '--json'], env=build_environment, text=True))
                    self.assertEqual(info, {'version': expected, 'commit': metadata['GIT_COMMIT'], 'build_date': metadata['BUILD_DATE']})
                self.assertTrue(os.path.samestat(before, cached.stat()))
                self.assertEqual(cached.read_bytes(), cached_bytes)
                self.assertEqual(self.launch('version', '--short').stdout.strip(), self.pin)
                self.assertFalse((destination / 'agentic-lens').exists(), 'source build copied an unqualified installed command')

    def test_printed_completion_scopes_both_files_to_selected_project_from_another_scaffold(self):
        self.config.unlink()
        application = build_app()
        initial = self.succeeds(self.install_local('--init-config', '--version', self.pin, probe=application))
        completion = next(line.strip() for line in initial.stdout.splitlines() if line.startswith('  (cd '))
        caller = self.root / 'caller project'; caller.mkdir()
        caller_shared = caller / '.agenticlens-config.yaml'; caller_shared.write_bytes(self.config.read_bytes())
        caller_local = caller / '.agenticlens-config.local.yaml'; caller_local.write_text('# preserve caller local\n')
        before = (caller_shared.read_bytes(), caller_local.read_bytes())
        selected_local = self.repo / '.agenticlens-config.local.yaml'
        selected_local.write_text(f"# selected local settings\ntelemetry:\n  capture:\n    enabled: false\n    state_dir: '{self.root / 'capture state'}'\n  default_source: codex\n  sources:\n    codex:\n      home: '{self.home / '.codex'}'\n")
        local_before = selected_local.read_bytes()
        approved = self.approved_settings('Selected')
        approved.write_text(shared_policy(self.pin, self.home / '.agenticlens', name='Selected'))
        candidate = self.root / 'approved local.yaml'
        candidate.write_bytes(local_before)
        command = completion.replace('APPROVED_SHARED.yaml', shlex.quote(str(approved))).replace('--workspaces DIR', '--workspaces workspaces').replace('APPROVED_LOCAL.yaml', shlex.quote(str(candidate)))
        scaffold_before = self.config.read_bytes()
        selected_local.unlink()
        candidate.unlink()
        missing_choices = subprocess.run(['bash', '-c', command], cwd=caller, env=self.env, text=True, capture_output=True, timeout=15)
        self.assertNotEqual(missing_choices.returncode, 0)
        self.assertIn(str(candidate), missing_choices.stderr)
        self.assertEqual(self.config.read_bytes(), scaffold_before)
        self.assertFalse(selected_local.exists())
        selected_local.write_bytes(local_before)
        candidate.write_bytes(local_before)
        result = subprocess.run(['bash', '-c', command], cwd=caller, env=self.env, text=True, capture_output=True, timeout=15)
        self.succeeds(result)
        self.assertIn('project_name: Selected', self.config.read_text())
        self.assertEqual(selected_local.read_bytes(), local_before)
        self.assertEqual((caller_shared.read_bytes(), caller_local.read_bytes()), before)
        completed = self.config.read_bytes()
        self.assertNotIn('state_dir:', completed.decode())
        selected_local.write_text(''.join(line for line in local_before.decode().splitlines(keepends=True)
                                          if not line.lstrip().startswith('state_dir:')))
        missing = self.launch('capture', 'status', '--repo', self.repo, '--state-dir', self.root / 'CLI rescue')
        self.assertNotEqual(missing.returncode, 0)
        self.assertIn('telemetry.capture.state_dir', missing.stderr)
        self.assertFalse((self.root / 'CLI rescue').exists())
        self.assertFalse((self.home / '.agenticlens/state').exists())
        self.config.write_bytes(completed)
        selected_local.write_bytes(local_before)

    def test_operator_launcher_missing_installation_and_default_user_entrypoint(self):
        launcher = ROOT / '.agents/skills/agentic-lens-operator/scripts/agentic-lens.sh'
        def launch():
            return subprocess.run(['sh', str(launcher), 'version', '--short'],
                                  cwd=self.repo, env=self.env, input='', text=True, capture_output=True, timeout=10)
        missing = launch()
        self.assertNotEqual(missing.returncode, 0)
        self.assertIn('restore the complete setup/operator', missing.stderr)
        self.assertIn('https://github.com/', missing.stderr)
        self.assertEqual(missing.stdout, '')
        self.succeeds(self.install_local('--dest', self.home / '.local/bin'))
        self.assertEqual(self.succeeds(launch()).stdout.strip(), self.pin)
        self.assertEqual(self.requests(), [])

    def git(self, *arguments, cwd=None):
        return subprocess.check_output(['git', *map(str, arguments)], cwd=cwd or self.repo, text=True, stderr=subprocess.PIPE).strip()

    def bundle(self, *arguments, remote=False):
        options = [] if remote else ['--source-dir', str(ROOT)]
        return subprocess.run(['bash', str(INSTALL), '--version', self.pin, '--dest', str(self.repo), *options, *map(str, arguments)],
                              cwd=self.repo, env=self.env, text=True, capture_output=True, timeout=30)

    def test_complete_skill_bundle_and_repair_preserve_owner_git_policy_and_config(self):
        self.git('init', '-q'); self.git('config', 'user.name', 'Fixture'); self.git('config', 'user.email', 'fixture@example.invalid')
        ignore = self.repo / '.gitignore'; ignore.write_text('# owner policy\n')
        before = self.config.read_bytes(); ignored = ignore.read_bytes()
        self.succeeds(self.bundle())
        for relative in ['scripts/install.sh', 'scripts/download.sh', 'SKILL.md']:
            self.assertTrue((self.repo / BUNDLE_SETUP / relative).is_file(), relative)
        self.assertEqual((self.repo / BUNDLE_SETUP / 'RELEASE_VERSION').read_text(), self.pin + '\n')
        self.succeeds(self.install_local(script=self.repo / BUNDLE_SETUP / 'scripts/download.sh'))
        self.assertEqual(self.config.read_bytes(), before); self.assertEqual(ignore.read_bytes(), ignored)
        self.assertEqual(self.git('config', '--local', '--get-regexp', r'^user\.'), 'user.name Fixture\nuser.email fixture@example.invalid')
        self.assertFalse((self.repo / BUNDLE_OPERATOR / 'bin').exists())
        self.assertEqual(self.requests(), [])

    def test_pinned_remote_bundle_and_checksum_repair_never_resolve_latest(self):
        self.succeeds(self.bundle(remote=True))
        self.assertTrue(self.requests())
        self.assertTrue(all('/' + self.pin + '/' in url for url in self.requests()))
        self.network_log.unlink()
        self.succeeds(self.install_local(script=self.repo / BUNDLE_SETUP / 'scripts/download.sh'))
        self.assertEqual(self.requests(), [])
        self.succeeds(self.launch('version', '--short'))

    def test_bundle_conflict_preflight_preserves_every_existing_file(self):
        conflict = self.repo / BUNDLE_OPERATOR / 'references/dashboard.md'
        conflict.parent.mkdir(parents=True); conflict.write_text('owner customized guidance\n')
        result = self.bundle()
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(conflict.read_text(), 'owner customized guidance\n')
        self.assertFalse((self.repo / BUNDLE_SETUP / 'scripts/download.sh').exists())
        self.assertFalse((self.repo / BUNDLE_SETUP / 'RELEASE_VERSION').exists())

    def test_worktree_uses_shared_pin_without_release_marker_or_repository_binary(self):
        (self.repo / '.agenticlens-config.local.yaml').unlink()
        self.git('init', '-q'); self.git('config', 'user.name', 'Fixture'); self.git('config', 'user.email', 'fixture@example.invalid')
        self.succeeds(self.bundle())
        marker = self.repo / BUNDLE_SETUP / 'RELEASE_VERSION'; marker.unlink()
        self.git('add', '.'); self.git('commit', '-qm', 'fixture baseline')
        worktree = self.root / 'worktree without marker'
        self.git('worktree', 'add', '-qb', 'fixture-worktree', worktree)
        self.succeeds(self.install_local(script=worktree / BUNDLE_SETUP / 'scripts/download.sh'))
        result = self.succeeds(self.launch('--al-project', worktree, 'version', '--short', cwd=worktree))
        self.assertEqual(result.stdout.strip(), self.pin)
        self.assertEqual(self.git('status', '--porcelain', cwd=worktree), '')
        self.assertFalse((worktree / BUNDLE_OPERATOR / 'bin').exists())
        self.assertFalse((worktree / BUNDLE_SETUP / 'RELEASE_VERSION').exists())
        shared_before = self.config.read_bytes()
        self.succeeds(self.install_local('--force', probe=build_app()))
        refused = self.launch('--al-project', worktree, 'init', '--local-only', '--capture-enabled', 'true',
                              '--source', 'codex', '--codex-home', self.home / '.codex',
                              '--state-dir', self.root / 'explicit state', cwd=worktree)
        self.assertNotEqual(refused.returncode, 0)
        self.assertIn('linked-worktree setup', refused.stderr)
        self.assertEqual(self.config.read_bytes(), shared_before)
        self.assertFalse((self.repo / '.agenticlens-config.local.yaml').exists())
        self.assertFalse((worktree / '.agenticlens-config.local.yaml').exists())
        self.assertFalse((self.root / 'explicit state').exists())

    def test_initial_pin_is_explicit_and_failure_rolls_back_only_created_configuration(self):
        (self.repo / '.agenticlens-config.local.yaml').unlink()
        self.config.unlink()
        self.assertNotEqual(self.install_local().returncode, 0)
        self.assertFalse(self.cache.exists())
        self.succeeds(self.install('--init-config', '--version', self.pin, '--dry-run'))
        self.assertFalse(self.config.exists()); self.assertFalse(self.cache.exists())
        result = self.install('--init-config', '--version', self.pin, env=dict(self.env, AL_FETCH_FAIL='SHA256SUMS'))
        self.assertNotEqual(result.returncode, 0); self.assertFalse(self.config.exists())
        self.succeeds(self.install_local('--init-config', '--version', self.pin))
        self.assertIn('version: ' + self.pin, self.config.read_text())
        self.assertNotIn('state_dir:', self.config.read_text())
        self.assertFalse((self.repo / '.agenticlens-config.local.yaml').exists(), 'binary bootstrap guessed local capture choices')

    def test_invalid_or_conflicting_versions_and_missing_values_fail_before_fetch(self):
        for version in ['latest', '../v9.8.7', 'v9.8.7/rc', 'v9.8.7 rc', 'v9.8.7+$(touch INJECTED)', 'v1.0.0']:
            for action in ([lambda: self.install_local('--version', version)] if version == 'v1.0.0' else [lambda: self.install_local('--version', version), lambda: self.bundle('--version', version)]):
                with self.subTest(version=version):
                    self.assertNotEqual(action().returncode, 0)
        self.assertFalse((self.repo / 'INJECTED').exists()); self.assertEqual(self.requests(), [])
        self.assertFalse(self.dest.exists())

    def test_explicit_repair_restores_native_permission_and_preserves_legacy_repository_files(self):
        legacy = self.repo / 'bin'; legacy.mkdir(); (legacy / 'agentic-lens').write_text('owner wrapper\n')
        self.succeeds(self.install_local())
        contents = self.native().read_bytes(); self.native().chmod(0o600)
        self.assertNotEqual(self.launch('version', '--short').returncode, 0)
        self.succeeds(self.install_local())
        self.assertEqual(self.native().read_bytes(), contents)
        self.succeeds(self.launch('version', '--short'))
        self.assertEqual((legacy / 'agentic-lens').read_text(), 'owner wrapper\n')


if __name__ == '__main__':
    unittest.main()
