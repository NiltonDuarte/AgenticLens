"""Exercise shipped common entrypoints and transactional native capture installation."""
import json
import os
from pathlib import Path
import shutil
import signal
import subprocess
import unittest

from distribution_fixture import DistributionFixture, ROOT, SETUP, build_probe, digest


class CaptureDistributionTest(DistributionFixture):
    def test_same_pin_second_project_force_and_permission_repair_preserve_native_identity(self):
        self.succeeds(self.install_local())
        native = self.native(); admitted = native.stat(); before = native.read_bytes()
        other = self.root / 'same pin project B'; other.mkdir()
        (other / '.agenticlens-config.yaml').write_bytes(self.config.read_bytes())
        for force, repair in [(False, False), (True, False), (True, True)]:
            with self.subTest(force=force, repair=repair):
                if repair: native.chmod(0o600)
                args = ['bash', str(SETUP / 'download.sh'), '--project', str(other), '--dest', str(self.dest),
                        '--local-bin', str(self.probe), '--sha256', digest(self.probe)]
                if force: args.append('--force')
                result = subprocess.run(args, cwd=other, env=self.env, input='', text=True, capture_output=True, timeout=20)
                self.succeeds(result)
                self.assertTrue(os.path.samestat(admitted, native.stat()), 'same release install retired admitted native identity')
                self.assertEqual(native.read_bytes(), before)
                self.assertEqual(self.launch('version', '--short').stdout.strip(), self.pin)
        self.assertEqual(self.requests(), [])

    def test_binary_only_legacy_pin_is_preserved_but_new_capture_capability_is_required(self):
        environment = dict(self.env, AL_PROBE_PROTOCOL='1')
        result = self.succeeds(self.install_local(env=environment))
        self.assertIn('hook protocol 2 and guarded init unavailable', result.stdout)
        self.assertEqual(self.launch('version', '--short', env=environment).stdout.strip(), self.pin)
        before = self.native().read_bytes()
        store = self.root / 'store'; store.mkdir()
        result = self.install_local('--capture-source', 'codex', '--capture-store', store, env=environment)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('required 2', result.stderr)
        self.assertEqual(self.native().read_bytes(), before)
        self.assertFalse((self.repo / '.codex').exists())

    def test_both_installers_route_exact_capture_arguments_through_common_wrapper(self):
        store = self.root / "source $(touch INJECTED)"
        store.mkdir()
        log = self.root / 'native-args.json'
        environment = dict(self.env, AL_PROBE_LOG=str(log))
        for script in [ROOT / 'scripts/download-agentic-lens.sh', SETUP / 'download.sh']:
            with self.subTest(script=script):
                result = self.succeeds(self.install_local('--capture-source', 'codex', '--capture-store', store, script=script, env=environment))
                self.assertEqual(json.loads(log.read_text())['args'], ['capture', 'install', '--source', 'codex', '--store', str(store),
                    '--repo', str(self.repo), '--binary', str(self.native()), '--wrapper', str(self.dest / 'agentic-lens')])
                self.assertIn('Runtime trust, binding and source extraction', result.stdout)
                self.assertFalse((self.repo / 'INJECTED').exists())
                self.assertEqual(self.requests(), [])

    def test_wrapper_preserves_arguments_stdin_stdout_cwd_exit_and_exec_signal(self):
        self.succeeds(self.install_local())
        nested = self.repo / 'nested directory'
        nested.mkdir()
        args = ['capture', 'receive', '--ticket-title', '--repo', '--request', "quote'\"", '--store', 'literal $(touch INJECTED)', '--session-id', '']
        result = self.launch('--al-offline', '--al-project', self.repo, *args, cwd=nested, env=dict(self.env, AL_PROBE_EXIT='23'))
        self.assertEqual(result.returncode, 23, result.stderr)
        self.assertEqual(json.loads(result.stdout), {'version': self.pin, 'args': args, 'stdin': 'payload\n', 'cwd': str(nested)})
        self.assertEqual(result.stderr, '')
        # Native reads stdin until EOF; live pipe holds it without arbitrary sleeps.
        process = subprocess.Popen([str(self.dest / 'agentic-lens'), 'capture', 'receive'], cwd=self.repo, env=self.env,
                                   stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        try:
            if Path('/proc').is_dir():
                for attempt in range(10000):
                    if Path(f'/proc/{process.pid}/exe').resolve() == self.native(): break
                else: self.fail('POSIX wrapper did not exec selected native')
            process.terminate()
            self.assertEqual(process.wait(timeout=5), -signal.SIGTERM)
        finally:
            if process.poll() is None: process.kill(); process.wait()
            process.stdin.close(); process.stdout.close(); process.stderr.close()
        self.assertEqual(self.requests(), [])

    def test_two_projects_and_exact_versions_coexist_without_repository_executables(self):
        self.succeeds(self.install_local())
        original = self.native().read_bytes()
        repo_b = self.root / 'project B'; repo_b.mkdir()
        pin_b = 'v2.0.0-rc.42+abc1234'
        (repo_b / '.agenticlens-config.yaml').write_text(f"agenticlens:\n  version: {pin_b}\n  install_root: '{self.cache}'\n")
        bundle = self.root / 'new compatible setup bundle'; bundle.mkdir()
        for filename in ['download.sh', 'agentic-lens-wrapper.sh', 'agentic-lens-wrapper.ps1']:
            target = bundle / filename
            shutil.copyfile(SETUP / filename, target)
            if filename != 'download.sh':
                target.write_text(target.read_text() + '\n# Compatible common entrypoint update\n')
        probe_b = build_probe(pin_b)
        result = subprocess.run(['bash', str(bundle / 'download.sh'), '--project', str(repo_b), '--dest', str(self.dest),
                                 '--local-bin', str(probe_b), '--sha256', digest(probe_b)],
                                cwd=repo_b, env=self.env, input='', text=True, capture_output=True, timeout=20)
        self.succeeds(result)
        self.assertEqual(json.loads(self.launch(cwd=repo_b).stdout)['version'], pin_b)
        self.assertEqual(json.loads(self.launch().stdout)['version'], self.pin)
        path_result = subprocess.run(['agentic-lens', 'version', '--short'], cwd=repo_b,
                                     env=dict(self.env, PATH=str(self.dest) + os.pathsep + self.env['PATH']),
                                     text=True, capture_output=True, timeout=5)
        self.assertEqual(self.succeeds(path_result).stdout.strip(), pin_b)
        self.assertEqual(self.native().read_bytes(), original)
        self.assertFalse((repo_b / 'bin').exists())
        self.assertFalse((self.repo / 'bin').exists())

    def test_missing_native_never_downloads_updates_or_mutates_cache(self):
        self.succeeds(self.install_local())
        self.native().unlink()
        before = sorted(str(path.relative_to(self.cache)) for path in self.cache.rglob('*'))
        for args in [[], ['--al-offline', 'capture', 'receive'], ['update']]:
            with self.subTest(args=args):
                result = self.launch(*args)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn('explicit', result.stderr)
        self.assertEqual(self.requests(), [])
        self.assertEqual(before, sorted(str(path.relative_to(self.cache)) for path in self.cache.rglob('*')))

    def test_unsupported_duplicate_and_unsafe_yaml_reject_before_native(self):
        self.succeeds(self.install_local())
        for document in ['agenticlens: {version: v9.8.7}\n', 'agenticlens:\n  version: latest\n',
                         'agenticlens:\n  version: v9.8.7\n  version: v9.8.7\n',
                         'agenticlens:\n  version: "v9.8.7+$(touch INJECTED)"\n',
                         '"agenticlens":\n  version: v9.8.7\n',
                         'agenticlens:\n  version: v9.8.7\n---\nproject_name: Other\n',
                         'agenticlens:\n  version: &pin v9.8.7\n', 'agenticlens:\n  version: |\n    v9.8.7\n']:
            with self.subTest(document=document):
                self.config.write_text(document)
                result = self.launch('capture', 'receive')
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(result.stdout, '')
        self.assertFalse((self.repo / 'INJECTED').exists())
        self.assertEqual(self.requests(), [])

    def test_quotes_comments_crlf_local_overlay_and_version_authority(self):
        self.succeeds(self.install_local())
        for quote in ['', "'", '"']:
            self.config.write_text(f"agenticlens: # authority\r\n  version: {quote}{self.pin}{quote} # release\r\n  install_root: /missing\r\n")
            local = self.repo / '.agenticlens-config.local.yaml'
            local.write_text(f"agenticlens:\n  install_root: '{self.cache}'\ntelemetry:\n  capture:\n    state_dir: /local/state\n")
            self.succeeds(self.launch())
        local.write_text(f'agenticlens:\n  version: {self.pin}\n')
        self.assertIn('local version', self.launch().stderr.lower())
        local.unlink(); self.write_config()
        self.assertIn('conflicts', self.launch(env=dict(self.env, AGENTIC_LENS_VERSION='v1.0.0')).stderr)

    def test_nearer_json_git_boundary_and_native_repo_mismatch_cannot_select_parent(self):
        self.succeeds(self.install_local())
        child = self.repo / 'child'; child.mkdir()
        (child / '.agenticlens-config.json').write_text('{"agenticlens":{"version":"v9.8.7"}}')
        self.assertIn('migrate', self.launch(cwd=child).stderr)
        (child / '.agenticlens-config.json').unlink(); (child / '.git').write_text('gitdir: somewhere\n')
        self.assertNotEqual(self.launch(cwd=child).returncode, 0)
        other = self.root / 'other'; other.mkdir()
        (other / '.agenticlens-config.yaml').write_bytes(self.config.read_bytes())
        self.assertIn('--al-project', self.launch('capture', 'status', '--repo', other).stderr)
        self.succeeds(self.launch('--al-project', other, 'capture', 'status', '--repo', other))

    def test_legacy_yml_suffix_ignores_a_directory_at_yaml_filename(self):
        self.succeeds(self.install_local())
        self.config.rename(self.repo / '.agenticlens-config.yml')
        self.config.mkdir()
        self.assertEqual(json.loads(self.succeeds(self.launch()).stdout)['version'], self.pin)

    def test_physical_project_and_cache_symlink_dotdot_resolution(self):
        self.succeeds(self.install_local())
        outside = self.root / 'outside'; outside.mkdir()
        target = outside / 'child'; target.mkdir()
        link = self.repo / 'link'; link.symlink_to(target, target_is_directory=True)
        # link/.. means outside, not the lexical repo parent.
        pin_b = 'v1.2.3'
        (outside / '.agenticlens-config.yaml').write_text(f"agenticlens:\n  version: {pin_b}\n  install_root: '{self.cache}'\n")
        native_b = self.native(pin_b); native_b.parent.mkdir(parents=True)
        shutil.copyfile(build_probe(pin_b), native_b); native_b.chmod(0o700)
        self.assertEqual(json.loads(self.launch('--al-project', str(link) + '/..').stdout)['version'], pin_b)
        self.assertIn('--al-project', self.launch('capture', 'status', '--repo', str(link) + '/..').stderr)
        cache_link = self.root / 'cache-link'; cache_link.symlink_to(self.cache / 'bin', target_is_directory=True)
        self.config.write_text(f"agenticlens:\n  version: {self.pin}\n  install_root: '{cache_link}/..'\n")
        self.succeeds(self.launch())

    def test_nonregular_and_oversized_config_reject_boundedly(self):
        self.succeeds(self.install_local())
        self.config.unlink(); os.mkfifo(self.config)
        self.assertIn('regular file', self.launch().stderr)
        self.config.unlink(); self.config.write_text('#' + 'x' * (1024 * 1024 + 1))
        self.assertIn('1 MiB', self.launch().stderr)

    def test_native_wrapper_recursion_and_bad_candidate_protocol_are_rejected(self):
        self.succeeds(self.install_local())
        before = self.native().read_bytes()
        result = self.install_local('--force', probe=build_probe(protocol='1'))
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(self.native().read_bytes(), before)
        result = self.install_local('--force', probe=ROOT / 'scripts/agentic-lens-wrapper.sh')
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(self.native().read_bytes(), before)
        self.native().write_bytes((ROOT / 'scripts/agentic-lens-wrapper.sh').read_bytes())
        self.assertIn('not a native', self.launch().stderr)

    def test_release_integrity_install_failure_and_capture_failure_restore_working_set(self):
        self.succeeds(self.install_local())
        before = self.native().read_bytes(); wrapper_before = (self.dest / 'agentic-lens').read_bytes()
        for overrides in [{'AL_FETCH_BAD_CHECKSUM': '1'}, {'AL_FETCH_FAIL': self.native().name}]:
            result = self.install('--force', env=dict(self.env, **overrides))
            self.assertNotEqual(result.returncode, 0)
            self.assertEqual(self.native().read_bytes(), before)
            self.assertEqual((self.dest / 'agentic-lens').read_bytes(), wrapper_before)
        store = self.root / 'store'; store.mkdir()
        result = self.install_local('--force', '--capture-source', 'codex', '--capture-store', store,
                                    env=dict(self.env, AL_PROBE_EXIT='17'))
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(self.native().read_bytes(), before)
        self.succeeds(self.launch())

    def test_install_preview_validation_and_concurrent_lock_preserve_owner_files(self):
        for option in ['--version', '--dest', '--project', '--sha256', '--local-bin']:
            self.assertNotEqual(self.install(option).returncode, 0)
        preview = self.succeeds(self.install('--all-platforms', '--dry-run'))
        for artifact in ['linux-amd64', 'linux-arm64', 'darwin-amd64', 'darwin-arm64', 'windows-amd64.exe']:
            self.assertIn(f'/releases/download/{self.pin}/agentic-lens-{artifact}', preview.stdout)
        self.assertIn('agentic-lens-darwin-arm64', self.succeeds(self.install('--include-host', 'm1', '--dry-run')).stdout)
        self.assertFalse(self.dest.exists()); self.assertFalse(self.cache.exists()); self.assertEqual(self.requests(), [])
        self.cache.mkdir(); lock = self.cache / f'.install-{self.pin}.lock'; lock.mkdir()
        self.assertNotEqual(self.install_local().returncode, 0)
        self.assertTrue(lock.exists())
        lock.rmdir(); self.succeeds(self.install_local())
        edited = b'#!/bin/sh\necho owner script\n'; (self.dest / 'agentic-lens').write_bytes(edited)
        self.assertNotEqual(self.install_local('--force').returncode, 0)
        self.assertEqual((self.dest / 'agentic-lens').read_bytes(), edited)


if __name__ == '__main__':
    unittest.main()
