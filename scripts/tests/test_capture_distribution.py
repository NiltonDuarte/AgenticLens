"""Exercise shipped common entrypoints and transactional native capture installation."""
import json
import os
from pathlib import Path
import shutil
import signal
import select
import subprocess
import unittest

from distribution_fixture import DistributionFixture, ROOT, SETUP, build_probe, build_app, digest


class CaptureDistributionTest(DistributionFixture):
    def install_native_reporting_probe(self):
        application = build_app(self.pin)
        self.succeeds(self.install_local('--force', probe=application))
        development = self.root / 'compatible development probe'; development.mkdir()
        candidate = development / self.native().name
        shutil.copyfile(self.probe, candidate); candidate.chmod(0o700)
        self.config.write_text(self.config.read_text().replace('agenticlens:\n', f"agenticlens:\n  binary_dir: '{development}'\n"))
        return application, candidate

    def snapshot_owned_paths(self, *roots):
        result = {}
        for root in roots:
            if not root.exists():
                result[str(root)] = None
                continue
            for path in [root, *root.rglob('*')]:
                if path.is_symlink():
                    result[str(path)] = ('link', os.readlink(path))
                elif path.is_file():
                    result[str(path)] = ('file', path.stat().st_mode, path.stat().st_size, digest(path))
                else:
                    result[str(path)] = ('directory',)
        return result

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
        self.assertIn('hook protocol 3 and guarded init unavailable', result.stdout)
        self.assertEqual(self.launch('version', '--short', env=environment).stdout.strip(), self.pin)
        before = self.native().read_bytes()
        store = self.root / 'store'; store.mkdir()
        result = self.install_local('--capture-source', 'codex', '--capture-store', store, env=environment)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('required 3', result.stderr)
        self.assertEqual(self.native().read_bytes(), before)
        self.assertFalse((self.repo / '.codex').exists())
        # A different project adopting native capture must not disable this
        # explicitly installed legacy manual CLI boundary in the shared wrapper.
        self.succeeds(self.install_local('--force', probe=build_app(self.pin)))
        legacy = self.root / 'legacy manual project'; legacy.mkdir()
        legacy_pin = 'v8.0.0'
        (legacy / '.agenticlens-config.yaml').write_text(f"agenticlens:\n  version: {legacy_pin}\n  install_root: '{self.cache}'\n")
        legacy_binary = build_probe(legacy_pin, protocol='1')
        self.succeeds(self.install_local('--project', legacy, probe=legacy_binary))
        invoked = self.succeeds(self.launch('version', '--short', cwd=legacy))
        self.assertEqual(invoked.stdout.strip(), legacy_pin)
        self.assertFalse((legacy / '.agenticlens-config.local.yaml').exists())
        self.assertFalse((legacy / '.codex').exists())
        self.assertEqual(self.requests(), [])

    def test_both_installers_route_exact_capture_arguments_through_common_wrapper(self):
        store = self.root / "source $(touch INJECTED)"
        store.mkdir()
        log = self.root / 'native-args.json'
        environment = dict(self.env, AL_PROBE_LOG=str(log))
        for script in [ROOT / 'scripts/download-agentic-lens.sh', SETUP / 'download.sh']:
            with self.subTest(script=script):
                for source in ['codex', 'opencode']:
                    result = self.succeeds(self.install_local('--capture-source', source, '--capture-store', store, script=script, env=environment))
                    expected = ['capture', 'install-user', '--repo', str(self.repo), '--binary', str(self.native())]
                    if source == 'opencode':
                        expected = ['capture', 'install', '--source', source, '--store', str(store),
                            '--repo', str(self.repo), '--binary', str(self.native()), '--wrapper', str(self.dest / 'agentic-lens')]
                    self.assertEqual(json.loads(log.read_text())['args'], expected)
                    self.assertIn('Runtime trust, binding and source extraction', result.stdout)
                    self.assertFalse((self.repo / 'INJECTED').exists())
                    self.assertEqual(self.requests(), [])

    def test_actual_native_user_install_preserves_unrelated_hooks_without_project_enrollment(self):
        home = Path(self.env['CODEX_HOME']); home.mkdir(parents=True)
        unrelated = {'hooks': [{'type': 'command', 'command': 'echo owner hook'}]}
        hooks = home / 'hooks.json'
        hooks.write_text(json.dumps({'hooks': {'SessionStart': [unrelated]}}))
        application = build_app(self.pin)
        self.succeeds(self.install_local('--capture-source', 'codex', '--capture-store', home, probe=application))
        installed = json.loads(hooks.read_text())['hooks']
        self.assertIn(unrelated, installed['SessionStart'])
        self.assertEqual(set(installed), {'SessionStart', 'UserPromptSubmit', 'SubagentStart', 'SubagentStop', 'Stop', 'SessionEnd'})
        for event, groups in installed.items():
            owned = [group for group in groups if group != unrelated]
            self.assertEqual(len(owned), 1, event)
            command = owned[0]['hooks'][0]['command']
            self.assertIn('dispatch', command)
            self.assertIn('codex-hook', command)
            self.assertNotIn(str(self.repo), command)
            self.assertNotIn(str(home), command)
            self.assertNotIn('python', command.lower())
        self.assertFalse((self.repo / '.codex').exists())
        self.assertFalse((self.root / 'capture state').exists())
        self.assertTrue((self.dest / 'agentic-lens.dispatcher').is_file())
        self.assertEqual(self.launch('version', '--short').stdout.strip(), self.pin)
        before = hooks.read_bytes()
        self.succeeds(self.install_local('--capture-source', 'codex', '--capture-store', home, probe=application))
        self.assertEqual(hooks.read_bytes(), before)
        dispatcher = self.cache / ('dispatcher.exe' if os.name == 'nt' else 'dispatcher')
        admitted = dispatcher.stat()
        self.succeeds(self.launch('capture', 'install-user', '--repo', self.repo, '--binary', self.native()))
        self.assertTrue(os.path.samestat(admitted, dispatcher.stat()), 'unchanged native user install replaced active dispatcher')
        self.assertEqual(hooks.read_bytes(), before)
        self.assertEqual(self.requests(), [])

    def test_actual_user_upgrade_keeps_one_dispatcher_authority_and_safe_uninstall(self):
        home = Path(self.env['CODEX_HOME']); home.mkdir(parents=True)
        hooks = home / 'hooks.json'
        foreign = {'hooks': [{'type': 'command', 'command': 'echo unrelated owner hook'}]}
        hooks.write_text(json.dumps({'hooks': {'SessionStart': [foreign]}}))
        first_pin = self.pin
        self.succeeds(self.install_local('--capture-source', 'codex', '--capture-store', home, probe=build_app(first_pin)))
        first_binary = self.native().read_bytes()
        receipt = home / 'agenticlens-user-installation.json'
        dispatcher = self.cache / ('dispatcher.exe' if os.name == 'nt' else 'dispatcher')
        authority = self.cache / 'dispatcher-installation.json'
        self.assertEqual(json.loads(receipt.read_text())['dispatcher_hash'], digest(dispatcher))
        self.write_config('v9.8.8')
        second_application = build_app(self.pin)
        self.succeeds(self.install_local('--capture-source', 'codex', '--capture-store', home, probe=second_application))
        self.assertEqual(self.native(first_pin).read_bytes(), first_binary)
        self.assertEqual(self.succeeds(self.launch('version', '--short')).stdout.strip(), self.pin)
        admitted_receipt = receipt.read_bytes(); admitted_hooks = hooks.read_bytes()
        self.assertEqual(json.loads(admitted_receipt)['dispatcher_hash'], digest(dispatcher))
        self.assertEqual((self.cache / 'dispatcher.sha256').read_text().strip(), digest(dispatcher))
        installed = json.loads(admitted_hooks)['hooks']
        self.assertIn(foreign, installed['SessionStart'])
        self.assertEqual(len(installed['SessionStart']), 2)
        self.assertTrue(all(len(groups) == 1 for event, groups in installed.items() if event != 'SessionStart'))
        self.assertEqual(json.loads(authority.read_text())['hash'], digest(dispatcher))

        # Another effective Codex home shares the same machine dispatcher. Its
        # legitimate update cannot make this home's historical hook receipt a
        # second executable-ownership authority or strand its later removal.
        other_home = self.root / 'second Codex home'; other_home.mkdir()
        other_environment = dict(self.env, CODEX_HOME=str(other_home))
        self.write_config('v9.8.9')
        second_application = build_app(self.pin)
        self.succeeds(self.install_local('--capture-source', 'codex', '--capture-store', other_home,
                                        probe=second_application, env=other_environment))
        self.assertEqual(receipt.read_bytes(), admitted_receipt)
        self.assertEqual(hooks.read_bytes(), admitted_hooks)
        self.assertNotEqual(json.loads(admitted_receipt)['dispatcher_hash'], digest(dispatcher))
        self.assertEqual(json.loads(authority.read_text())['hash'], digest(dispatcher))
        self.assertEqual((self.cache / 'dispatcher.sha256').read_text().strip(), digest(dispatcher))
        other_hooks = other_home / 'hooks.json'; other_hooks_before = other_hooks.read_bytes()
        other_receipt = other_home / 'agenticlens-user-installation.json'; other_receipt_before = other_receipt.read_bytes()
        authority_before = authority.read_bytes()
        dispatcher_bytes = dispatcher.read_bytes()
        customized = dispatcher_bytes + b'owner customization'
        dispatcher.write_bytes(customized)
        rejected = self.install_local('--capture-source', 'codex', '--capture-store', home, probe=second_application)
        self.assertNotEqual(rejected.returncode, 0)
        self.assertEqual(dispatcher.read_bytes(), customized)
        self.assertEqual(authority.read_bytes(), authority_before)
        self.assertEqual(receipt.read_bytes(), admitted_receipt)
        self.assertEqual(hooks.read_bytes(), admitted_hooks)
        dispatcher.write_bytes(dispatcher_bytes)

        changed = json.loads(admitted_hooks)
        owned = next(group for group in changed['hooks']['SessionStart'] if group != foreign)
        owned['hooks'][0]['command'] += ' owner-change'
        hooks.write_text(json.dumps(changed))
        customized_hooks = hooks.read_bytes()
        rejected = self.launch('capture', 'uninstall-user', '--repo', self.repo)
        self.assertNotEqual(rejected.returncode, 0)
        self.assertEqual(hooks.read_bytes(), customized_hooks)
        self.assertEqual(receipt.read_bytes(), admitted_receipt)
        self.assertEqual(dispatcher.read_bytes(), dispatcher_bytes)
        hooks.write_bytes(admitted_hooks)

        self.succeeds(self.launch('capture', 'uninstall-user', '--repo', self.repo))
        self.assertEqual(json.loads(hooks.read_text())['hooks'], {'SessionStart': [foreign]})
        self.assertFalse(json.loads(receipt.read_text()).get('groups'))
        self.assertEqual(json.loads(authority.read_text())['hash'], digest(dispatcher))
        self.assertEqual(other_hooks.read_bytes(), other_hooks_before)
        self.assertEqual(other_receipt.read_bytes(), other_receipt_before)
        self.assertEqual(self.requests(), [])

    def test_actual_installed_native_dispatcher_prefers_development_and_rejects_invalid_output(self):
        self.succeeds(self.install_local(probe=build_app(self.pin)))
        development = self.root / 'development output'; development.mkdir()
        candidate = development / self.native().name
        shutil.copyfile(build_app('v9.8.8-dev.1'), candidate); candidate.chmod(0o700)
        self.config.write_text(self.config.read_text().replace('agenticlens:\n', f"agenticlens:\n  binary_dir: '{development}'\n"))
        self.assertEqual(self.succeeds(self.launch('version', '--short')).stdout.strip(), 'v9.8.8-dev.1')
        candidate.write_bytes(b'invalid native output')
        rejected = self.launch('version', '--short')
        self.assertNotEqual(rejected.returncode, 0)
        self.assertIn('development output', rejected.stderr)
        candidate.unlink()
        self.assertEqual(self.succeeds(self.launch('version', '--short')).stdout.strip(), self.pin)
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
        # Exercise the real installed native dispatcher branch as well as the
        # bootstrap shell branch above. The selected probe exposes native cwd,
        # arguments and a relative file write without implementing selection.
        application, _ = self.install_native_reporting_probe()
        environment = dict(self.env, AL_PROBE_CAPTURE_APPLY=str(application), AL_PROBE_EXIT='23', AL_PROBE_LOG='relative-result.json')
        relative_args = ['fixture-cwd', '--json-output', 'relative/metrics.json', '--append', 'relative/SUMMARY.md']
        payload = 'native payload ü\n'
        for project_args in [[], ['--al-project', self.repo]]:
            with self.subTest(native_project_args=project_args):
                result = self.launch(*project_args, *relative_args, cwd=nested, env=environment, input=payload)
                self.assertEqual(result.returncode, 23, result.stderr)
                expected = {'version': self.pin, 'args': relative_args, 'stdin': payload, 'cwd': str(nested)}
                self.assertEqual(json.loads(result.stdout), expected)
                self.assertEqual(json.loads((nested / 'relative-result.json').read_text()), expected)
                self.assertFalse((self.repo / 'relative-result.json').exists())
                (nested / 'relative-result.json').unlink()
        self.assertEqual(self.requests(), [])

    def test_adopted_native_dispatch_preserves_exec_cancellation_and_signal_exit(self):
        application, candidate = self.install_native_reporting_probe()
        environment = dict(self.env, AL_PROBE_CAPTURE_APPLY=str(application))
        pid_file = self.root / 'selected-command.pid'
        waiting = dict(environment, AL_PROBE_TRANSPORT_MODE='stdin-close', AL_PROBE_READY='1', AL_PROBE_PID=str(pid_file))
        process = subprocess.Popen([str(self.dest / 'agentic-lens'), 'fixture-wait'], cwd=self.repo, env=waiting,
                                   stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        selected_pid = None
        try:
            readable, _, _ = select.select([process.stdout], [], [], 10)
            self.assertTrue(readable, 'selected native command did not signal readiness')
            self.assertEqual(process.stdout.readline(), b'ready\n')
            selected_pid = int(pid_file.read_text())
            self.assertEqual(selected_pid, process.pid, 'native dispatch did not exec the selected command')
            if Path('/proc').is_dir():
                self.assertEqual(Path(f'/proc/{process.pid}/exe').resolve(), candidate)
            process.terminate()
            self.assertEqual(process.wait(timeout=5), -signal.SIGTERM)
            with self.assertRaises(ProcessLookupError):
                os.kill(selected_pid, 0)
        finally:
            if selected_pid is not None and selected_pid != process.pid:
                try:
                    os.kill(selected_pid, signal.SIGKILL)
                except ProcessLookupError:
                    pass
            if process.poll() is None:
                process.kill(); process.wait(timeout=5)
            process.stdin.close(); process.stdout.close(); process.stderr.close()
        signaled = self.launch('fixture-signal', env=dict(environment, AL_PROBE_SELF_SIGNAL=str(signal.SIGTERM)))
        self.assertEqual(signaled.returncode, -signal.SIGTERM, signaled.stderr)
        self.assertEqual(self.requests(), [])

    def test_incompatible_user_dispatcher_is_rejected_without_installation_mutation(self):
        application = build_app(self.pin)
        foreign_system = 'darwin' if __import__('platform').system() != 'Darwin' else 'linux'
        candidates = [Path(shutil.which('true')), build_probe(self.pin, protocol='1'),
                      build_probe(self.pin, operating_system=foreign_system, arch='amd64')]
        home = Path(self.env['CODEX_HOME'])
        for installed in [False, True]:
            if installed:
                self.succeeds(self.install_local('--capture-source', 'codex', '--capture-store', home, probe=application))
            for candidate in candidates:
                with self.subTest(installed=installed, candidate=candidate.name):
                    before = self.snapshot_owned_paths(self.cache, home)
                    result = subprocess.run([str(application), 'capture', 'install-user', '--repo', str(self.repo),
                                             '--binary', str(candidate)], cwd=self.repo, env=self.env,
                                            input='', text=True, capture_output=True, timeout=10)
                    self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
                    self.assertEqual(self.snapshot_owned_paths(self.cache, home), before)
                    if installed:
                        self.assertEqual(self.succeeds(self.launch('version', '--short')).stdout.strip(), self.pin)
        self.assertEqual(self.requests(), [])

    def test_user_install_diagnoses_base_legacy_callbacks_and_migrates_only_owned_groups(self):
        application = build_app(self.pin)
        self.succeeds(self.install_local(probe=application))
        source = Path(self.env['CODEX_HOME']); source.mkdir(parents=True, exist_ok=True)
        (source / 'sessions').mkdir()
        state = self.root / 'legacy capture state'
        self.config.write_text(self.config.read_text() + 'ticket:\n  prefixes: [FX]\nworkspaces:\n  dir: workspaces\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n')
        (self.repo / 'src').mkdir()
        local = self.repo / '.agenticlens-config.local.yaml'
        local.write_text(f"telemetry:\n  default_source: codex\n  sources:\n    codex:\n      home: '{source}'\n  capture:\n    enabled: true\n    state_dir: '{state}'\n")
        for args in [('init', '-q'), ('config', 'user.name', 'Fixture'), ('config', 'user.email', 'fixture@example.invalid'),
                     ('add', '.agenticlens-config.yaml'), ('commit', '-qm', 'fixture base')]:
            subprocess.run(['git', *args], cwd=self.repo, env=self.env, check=True, capture_output=True, timeout=10)
        self.succeeds(self.launch('capture', 'install', '--source', 'codex', '--store', source,
                                 '--repo', self.repo, '--binary', self.native(), '--wrapper', self.dest / 'agentic-lens'))
        project_hooks = self.repo / '.codex/hooks.json'
        anchor = self.repo / '.codex/agenticlens-installation.json'
        manifest = json.loads(anchor.read_text())
        pending = Path(manifest['state_dir']) / 'retained-pending-facts'
        pending.write_bytes(b'preserve existing capture state')
        foreign = {'hooks': [{'type': 'command', 'command': 'echo unrelated project hook'}]}
        document = json.loads(project_hooks.read_text())
        document['hooks']['SessionStart'].append(foreign)
        project_hooks.write_text(json.dumps(document))
        original_hooks = project_hooks.read_bytes()
        worktree = self.root / 'linked worktree'
        subprocess.run(['git', 'worktree', 'add', '-qb', 'fixture-linked', str(worktree)], cwd=self.repo,
                       env=self.env, check=True, capture_output=True, timeout=10)
        roots = [self.cache, source, self.repo / '.codex', state]
        for project in [self.repo, worktree]:
            with self.subTest(project=project):
                before = self.snapshot_owned_paths(*roots)
                refused = self.launch('--al-project', project, 'capture', 'install-user', '--repo', project, '--binary', self.native())
                self.assertNotEqual(refused.returncode, 0, refused.stdout + refused.stderr)
                self.assertIn('legacy', refused.stderr.lower())
                self.assertEqual(self.snapshot_owned_paths(*roots), before)
        changed = json.loads(original_hooks)
        changed['hooks']['SessionStart'][0]['hooks'][0]['command'] += ' owner change'
        project_hooks.write_text(json.dumps(changed))
        before = self.snapshot_owned_paths(*roots)
        refused = self.launch('--al-project', worktree, 'capture', 'install-user', '--repo', worktree,
                              '--binary', self.native(), '--migrate-legacy-repo', self.repo)
        self.assertNotEqual(refused.returncode, 0)
        after = self.snapshot_owned_paths(*roots)
        # Stronger affected-home locking may create its empty coordination
        # anchor on this first explicit attempt; it contains no capture facts.
        # All pre-existing artifacts must still match, and later snapshots
        # include this anchor when testing unknown callback refusal.
        lock = str(self.repo / '.codex/install.lock')
        if lock not in before and lock in after:
            self.assertEqual(after[lock][0], 'file')
            self.assertEqual(after[lock][1] & 0o777, 0o600)
            self.assertEqual(after[lock][2], 0)
            before[lock] = after[lock]
        self.assertEqual(after, before)
        project_hooks.write_bytes(original_hooks)
        # A receipt cannot authorize deleting an additional active integration.
        # Copy the real installed callback, changing only its status text, so
        # this remains executable duplicate capture rather than a marker stub.
        residual = json.loads(original_hooks)
        unknown = json.loads(json.dumps(residual['hooks']['SessionStart'][0]))
        unknown['hooks'][0]['statusMessage'] = 'AgenticLens unknown preserved integration'
        self.assertEqual(unknown['hooks'][0]['command'], residual['hooks']['SessionStart'][0]['hooks'][0]['command'])
        residual['hooks']['SessionStart'].append(unknown)
        project_hooks.write_text(json.dumps(residual))
        for project in [self.repo, worktree]:
            with self.subTest(explicit_migration_with_unknown_callback=project):
                before = self.snapshot_owned_paths(*roots)
                refused = self.launch('--al-project', project, 'capture', 'install-user', '--repo', project,
                                      '--binary', self.native(), '--migrate-legacy-repo', self.repo)
                self.assertNotEqual(refused.returncode, 0, refused.stdout + refused.stderr)
                self.assertRegex(refused.stderr.lower(), 'unknown|unowned|conflict|legacy')
                self.assertEqual(self.snapshot_owned_paths(*roots), before)
                self.assertTrue(json.loads(anchor.read_text())['active'])
                self.assertFalse((source / 'hooks.json').exists(), 'conflicted migration activated global callbacks')
        # Model the user explicitly removing the unknown callback; only then is
        # the unchanged-owned migration permitted. Product code never removes it.
        project_hooks.write_bytes(original_hooks)
        self.succeeds(self.launch('--al-project', worktree, 'capture', 'install-user', '--repo', worktree,
                                 '--binary', self.native(), '--migrate-legacy-repo', self.repo))
        self.assertEqual(json.loads(project_hooks.read_text())['hooks'], {'SessionStart': [foreign]})
        self.assertFalse(json.loads(anchor.read_text())['active'])
        self.assertEqual(pending.read_bytes(), b'preserve existing capture state')
        self.assertEqual(len(json.loads((source / 'hooks.json').read_text())['hooks']), 6)
        self.assertFalse((worktree / '.codex/hooks.json').exists())
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
        for protocol in ['1', '2']:
            with self.subTest(protocol=protocol):
                result = self.install_local('--force', probe=build_probe(protocol=protocol))
                self.assertNotEqual(result.returncode, 0)
                self.assertIn('required 3', result.stderr)
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
