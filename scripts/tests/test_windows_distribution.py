"""Execute shipped native PowerShell installers/entrypoints against real PE fixtures.

Run on Windows, or explicitly provide AL_WINDOWS_POWERSHELL from WSL. This suite
fails rather than silently substituting text assertions when Windows is absent.
"""
import json
import os
from pathlib import Path
import platform
import shutil
import stat
import subprocess
import time
import unittest
import uuid

from distribution_fixture import ROOT, SETUP, build_probe, build_app, digest

IS_WINDOWS = platform.system() == 'Windows'
POWERSHELL = os.environ.get('AL_WINDOWS_POWERSHELL', 'powershell.exe' if IS_WINDOWS else '')
if not POWERSHELL:
    raise SystemExit('Real Windows PowerShell required; run this distribution gate on Windows or explicitly via WSL interop')


def win(path):
    if IS_WINDOWS:
        return str(path)
    return subprocess.check_output(['wslpath', '-w', str(path)], text=True).strip()


def local(path):
    if IS_WINDOWS:
        return Path(path)
    return Path(subprocess.check_output(['wslpath', '-u', path], text=True).strip())


class WindowsDistributionTest(unittest.TestCase):
    def setUp(self):
        name = 'al68-iter04-bootstrap-' + uuid.uuid4().hex
        code = "$p = [IO.Path]::Combine([IO.Path]::GetTempPath(), '" + name + "'); [IO.Directory]::CreateDirectory($p) | Out-Null; [Console]::WriteLine($p)"
        created = subprocess.check_output([POWERSHELL, '-NoLogo', '-NoProfile', '-NonInteractive', '-Command', code], text=True).strip()
        self.root = local(created)
        self.addCleanup(self.remove_fixture_tree)
        self.repo = self.root / 'project A'; self.repo.mkdir()
        self.home = self.root / 'isolated native home'; self.home.mkdir()
        self.cache = self.root / 'native cache ü'; self.dest = self.root / 'user PATH bin'
        self.scripts = self.root / 'setup scripts'; self.scripts.mkdir()
        for filename in ['download.ps1', 'agentic-lens-wrapper.ps1']:
            shutil.copyfile(SETUP / filename, self.scripts / filename)
        self.pin = 'v9.8.7'
        self.config = self.repo / '.agenticlens-config.yaml'
        self.write_config()
        self.probe = self.root / 'native release.exe'
        shutil.copyfile(build_probe(self.pin, 'windows', 'amd64'), self.probe)
        self.env = {key: value for key, value in os.environ.items() if not key.startswith(('AL_PROBE_', 'AGENTIC_LENS_'))}
        self.env.update(HOME=win(self.home), USERPROFILE=win(self.home), CODEX_HOME=win(self.home / '.codex'),
                        XDG_CONFIG_HOME=win(self.home / '.config'), XDG_CACHE_HOME=win(self.home / '.cache'),
                        XDG_STATE_HOME=win(self.home / '.state'))

    def remove_fixture_tree(self):
        root = self.root.resolve()

        def remove_readonly(function, name, exception):
            error = exception[1]
            candidate = Path(name)
            if not isinstance(error, PermissionError) or candidate.is_symlink() or not candidate.is_file():
                raise error
            if not candidate.resolve().is_relative_to(root):
                raise error
            metadata = candidate.stat()
            readonly = getattr(metadata, 'st_file_attributes', 0) & getattr(stat, 'FILE_ATTRIBUTE_READONLY', 1)
            if not readonly and metadata.st_mode & stat.S_IWRITE:
                raise error
            # Retained fixture executables are intentionally read-only. Unlock
            # only this proven owned file and retry once; a loaded image or any
            # other permission failure remains a visible test error.
            candidate.chmod(metadata.st_mode | stat.S_IWRITE)
            function(name)

        shutil.rmtree(self.root, onerror=remove_readonly)

    def write_config(self, form=None):
        self.config.write_text(f"agenticlens: # authority\n  version: {form or self.pin} # exact release\n  install_root: '{win(self.cache)}'\n", encoding='utf-8')

    def native(self):
        return self.cache / 'bin' / self.pin / 'windows-amd64' / 'agentic-lens-windows-amd64.exe'

    def call(self, script, args, env=None, input='payload\n'):
        environment = dict(env or self.env)
        if not IS_WINDOWS:
            forwarded = [key for key in environment if key.startswith(('AL_PROBE_', 'AGENTIC_LENS_')) or key in ('HOME', 'USERPROFILE', 'CODEX_HOME', 'XDG_CONFIG_HOME', 'XDG_CACHE_HOME', 'XDG_STATE_HOME')]
            environment['WSLENV'] = ':'.join(filter(None, [environment.get('WSLENV', ''), *forwarded]))
        # Text-mode pipes translate LF to CRLF in native Windows Python before
        # the wrapper receives stdin. Send exact UTF-8 bytes and decode only output.
        result = subprocess.run([POWERSHELL, '-NoLogo', '-NoProfile', '-NonInteractive', '-File', win(script), *args],
                                env=environment, input=input.encode('utf-8'), capture_output=True, timeout=20)
        return subprocess.CompletedProcess(result.args, result.returncode,
                                           result.stdout.decode('utf-8'), result.stderr.decode('utf-8'))

    def install(self, *extra, env=None):
        return self.call(self.scripts / 'download.ps1', ['-Project', win(self.repo), '-Dest', win(self.dest),
            '-LocalBin', win(self.probe), '-Sha256', digest(self.probe), *extra], env=env)

    def launch(self, *args, env=None, input='payload\n'):
        return self.call(self.dest / 'agentic-lens.ps1', ['--al-offline', '--al-project', win(self.repo), *args], env=env, input=input)

    def succeeds(self, result):
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        return result

    def use_actual_application(self):
        application = self.root / 'actual application.exe'
        shutil.copyfile(build_app(self.pin, 'windows', 'amd64'), application)
        shutil.copyfile(application, self.probe)
        self.config.write_text(f"project_name: Fixture\nagenticlens:\n  version: {self.pin}\n  install_root: '{win(self.cache)}'\nticket:\n  prefixes: [FX]\nworkspaces:\n  dir: workspaces\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n", encoding='utf-8')
        self.write_local_policy()
        return application

    def write_local_policy(self):
        local_config = self.repo / '.agenticlens-config.local.yaml'
        local_config.write_text(f"agenticlens:\n  install_root: '{win(self.cache)}'\ntelemetry:\n  capture:\n    enabled: false\n    state_dir: '{win(self.root / 'capture state')}'\n  default_source: codex\n  sources:\n    codex:\n      home: '{win(self.home / '.codex')}'\n", encoding='utf-8')
        return local_config

    def install_native_reporting_probe(self):
        application = self.use_actual_application()
        self.succeeds(self.install('-Force'))
        development = self.root / 'compatible development probe'; development.mkdir()
        candidate = development / self.native().name
        shutil.copyfile(build_probe(self.pin, 'windows', 'amd64'), candidate)
        self.config.write_text(self.config.read_text().replace('agenticlens:\n', f"agenticlens:\n  binary_dir: '{win(development)}'\n"), encoding='utf-8')
        return application

    def test_same_pin_second_project_and_force_retain_native_file_identity(self):
        self.succeeds(self.install())
        native = self.native(); admitted = native.stat(); original = native.read_bytes()
        other = self.root / 'same pin project B'; other.mkdir()
        (other / '.agenticlens-config.yaml').write_bytes(self.config.read_bytes())
        for force in [False, True]:
            arguments = ['-Project', win(other), '-Dest', win(self.dest), '-LocalBin', win(self.probe), '-Sha256', digest(self.probe)]
            if force: arguments.append('-Force')
            self.succeeds(self.call(self.scripts / 'download.ps1', arguments))
            self.assertTrue(os.path.samestat(admitted, native.stat()), 'identical install changed admitted Windows file identity')
            self.assertEqual(native.read_bytes(), original)
            self.assertEqual(self.launch('version', '--short').stdout.strip(), self.pin)

    def test_printed_completion_changes_only_selected_project_from_different_scaffold(self):
        self.use_actual_application()
        self.config.unlink()
        # Windows UserProfile is independent of HOME: pin the native cache locally.
        selected_local = self.repo / '.agenticlens-config.local.yaml'
        self.write_local_policy()
        first = self.succeeds(self.install('-InitConfig', '-Version', self.pin))
        printed = next(line.split(': ', 1)[1] for line in first.stdout.splitlines() if line.startswith('Complete unchanged owned scaffold: '))
        command = printed.replace('--project-name NAME', '--project-name Selected').replace('--src-folders DIRS', '--src-folders src').replace('--ticket-prefixes PREFIXES', '--ticket-prefixes FX')
        caller = self.root / 'caller project'; caller.mkdir()
        caller_shared = caller / '.agenticlens-config.yaml'; caller_shared.write_bytes(self.config.read_bytes())
        caller_local = caller / '.agenticlens-config.local.yaml'; caller_local.write_text('# caller local\n')
        before = caller_shared.read_bytes(), caller_local.read_bytes(), selected_local.read_bytes()
        quote = lambda value: "'" + value.replace("'", "''") + "'"
        execute = 'Set-Location -LiteralPath ' + quote(win(caller)) + '; ' + command
        execute += '; if ((Get-Location).ProviderPath -ne ' + quote(win(caller)) + ") { throw 'Caller location changed' }; exit $LASTEXITCODE"
        result = subprocess.run([POWERSHELL, '-NoLogo', '-NoProfile', '-NonInteractive', '-Command', execute],
                                env=self.env, input='', text=True, encoding='utf-8', capture_output=True, timeout=20)
        self.succeeds(result)
        self.assertIn('project_name: Selected', self.config.read_text())
        self.assertEqual((caller_shared.read_bytes(), caller_local.read_bytes(), selected_local.read_bytes()), before)

    def test_actual_capture_install_drains_large_result_and_retains_handlers_after_uncertain_apply(self):
        application = self.use_actual_application()
        store = self.root / 'source store'; store.mkdir()
        installed = self.succeeds(self.install('-CaptureSource', 'codex', '-CaptureStore', win(store)))
        decoder = json.JSONDecoder(); result, _ = decoder.raw_decode(installed.stdout)
        self.assertEqual(result['manifest']['source'], 'codex')
        hooks = self.home / '.codex/hooks.json'; previous_hooks = hooks.read_bytes()
        self.assertEqual(len(json.loads(previous_hooks)['hooks']), 6)
        self.assertFalse((self.repo / '.codex/hooks.json').exists())
        dispatcher = self.cache / 'dispatcher.exe'; admitted = dispatcher.stat()
        self.succeeds(self.launch('capture', 'install-user', '--repo', win(self.repo), '--binary', win(self.native())))
        self.assertTrue(os.path.samestat(admitted, dispatcher.stat()), 'unchanged install rewrote running dispatcher')
        self.assertEqual(hooks.read_bytes(), previous_hooks)
        # Real application apply finishes before the proxy injects result failure.
        shutil.copyfile(build_probe(self.pin, 'windows', 'amd64'), self.probe)
        pid_file = self.root / 'owned-apply.pid'
        environment = dict(self.env, AL_PROBE_CAPTURE_APPLY=win(application), AL_PROBE_PID=win(pid_file))
        for mode in ['both-streams', 'exit', 'overflow', 'hang']:
            with self.subTest(mode=mode):
                result = self.install('-Force', '-CaptureSource', 'codex', '-CaptureStore', win(store),
                                      env=dict(environment, AL_PROBE_CAPTURE_RESULT=mode))
                if mode == 'both-streams':
                    self.succeeds(result)
                    self.assertGreater(len(result.stderr.encode('utf-8')), 8192)
                else:
                    self.assertNotEqual(result.returncode, 0)
                    self.assertIn('validated native/wrapper retained', result.stderr)
                    self.assertTrue(any(self.cache.glob('.install-stage.*')), 'recoverable prior backups were removed')
                self.assertTrue(self.native().is_file())
                self.assertTrue((self.dest / 'agentic-lens.ps1').is_file())
                self.assertEqual(hooks.read_bytes(), previous_hooks)
                self.succeeds(self.launch('capture', 'status', '--repo', win(self.repo), env=environment))
                if mode == 'hang':
                    pid = int(pid_file.read_text())
                    stopped = subprocess.run([POWERSHELL, '-NoLogo', '-NoProfile', '-NonInteractive', '-Command',
                                              f'if (Get-Process -Id {pid} -ErrorAction SilentlyContinue) {{ exit 1 }}; exit 0'], capture_output=True)
                    self.assertEqual(stopped.returncode, 0, 'owned capture apply process survived timeout')

    def test_native_user_upgrade_keeps_receipt_consistent_and_preserves_foreign_hooks(self):
        self.use_actual_application()
        home = self.home / '.codex'; home.mkdir(parents=True)
        hooks = home / 'hooks.json'
        foreign = {'hooks': [{'type': 'command', 'command': 'echo owner hook'}]}
        hooks.write_text(json.dumps({'hooks': {'SessionStart': [foreign]}}))
        self.succeeds(self.install('-CaptureSource', 'codex', '-CaptureStore', win(home)))
        first_native = self.native(); first_bytes = first_native.read_bytes()
        self.pin = 'v9.8.8'
        self.use_actual_application()
        self.succeeds(self.install('-CaptureSource', 'codex', '-CaptureStore', win(home)))
        self.assertEqual(first_native.read_bytes(), first_bytes)
        self.assertEqual(self.succeeds(self.launch('version', '--short')).stdout.strip(), self.pin)
        dispatcher = self.cache / 'dispatcher.exe'
        receipt = home / 'agenticlens-user-installation.json'
        authority = self.cache / 'dispatcher-installation.json'
        self.assertEqual(json.loads(authority.read_text())['hash'], digest(dispatcher))
        self.assertEqual((self.cache / 'dispatcher.exe.sha256').read_text().strip(), digest(dispatcher))
        active = local(json.loads(authority.read_text())['active_native'])
        self.assertEqual(json.loads(authority.read_text())['active_hash'], digest(active))
        self.assertEqual(active.read_bytes(), self.native().read_bytes())
        self.assertIn(foreign, json.loads(hooks.read_text())['hooks']['SessionStart'])
        before_hooks = hooks.read_bytes(); before_receipt = receipt.read_bytes()
        before_dispatcher = dispatcher.read_bytes()
        dispatcher.write_bytes(before_dispatcher + b'owner customization')
        refused = self.install('-CaptureSource', 'codex', '-CaptureStore', win(home))
        self.assertNotEqual(refused.returncode, 0)
        self.assertEqual(dispatcher.read_bytes(), before_dispatcher + b'owner customization')
        self.assertEqual(receipt.read_bytes(), before_receipt)
        self.assertEqual(hooks.read_bytes(), before_hooks)
        dispatcher.write_bytes(before_dispatcher)
        active_bytes = active.read_bytes(); authority_bytes = authority.read_bytes()
        active_mode = active.stat().st_mode
        active.chmod(active_mode | stat.S_IWRITE)
        try:
            active.write_bytes(active_bytes + b'owner active-build customization')
            refused = self.install('-CaptureSource', 'codex', '-CaptureStore', win(home))
            self.assertNotEqual(refused.returncode, 0)
            self.assertEqual(active.read_bytes(), active_bytes + b'owner active-build customization')
            self.assertEqual(authority.read_bytes(), authority_bytes)
            self.assertEqual(hooks.read_bytes(), before_hooks)
            self.assertEqual(receipt.read_bytes(), before_receipt)
        finally:
            active.write_bytes(active_bytes)
            active.chmod(active_mode)
        # Repair of a missing stable launcher writes the selected B bytes, so
        # both stable and active identities must remain usable and coherent.
        dispatcher.unlink()
        self.succeeds(self.install('-CaptureSource', 'codex', '-CaptureStore', win(home)))
        repaired = json.loads(authority.read_text())
        self.assertEqual(repaired['hash'], digest(dispatcher))
        self.assertEqual((self.cache / 'dispatcher.exe.sha256').read_text().strip(), digest(dispatcher))
        self.assertEqual(json.loads(receipt.read_text())['dispatcher_hash'], digest(dispatcher))
        self.assertEqual(repaired['active_hash'], digest(local(repaired['active_native'])))
        self.assertEqual(dispatcher.read_bytes(), self.native().read_bytes())
        self.assertEqual(self.succeeds(self.launch('version', '--short')).stdout.strip(), self.pin)
        self.assertEqual(hooks.read_bytes(), before_hooks)
        self.succeeds(self.launch('capture', 'uninstall-user', '--repo', win(self.repo)))
        self.assertEqual(json.loads(hooks.read_text())['hooks'], {'SessionStart': [foreign]})
        self.assertFalse(json.loads(receipt.read_text()).get('groups'))
        self.assertEqual(json.loads(authority.read_text())['hash'], digest(dispatcher))

    def test_rendered_native_user_binding_command_executes_in_powershell(self):
        renamed = self.root / "project O'Brien"
        self.repo.rename(renamed); self.repo = renamed
        self.config = self.repo / '.agenticlens-config.yaml'
        self.use_actual_application()
        old_cache = win(self.cache)
        self.cache = self.root / "native cache O'Brien"
        local_config = self.repo / '.agenticlens-config.local.yaml'
        cache_scalar = '"' + win(self.cache).replace('\\', '/') + '"'
        for configured in [self.config, local_config]:
            contents = configured.read_text()
            old_scalar = "'" + old_cache + "'"
            self.assertIn(old_scalar, contents)
            configured.write_text(contents.replace(old_scalar, cache_scalar), encoding='utf-8')
        self.succeeds(self.install())
        selected = self.native()
        (self.repo / 'src').mkdir()
        enabled = local_config.read_text().replace('enabled: false', 'enabled: true')
        enabled = enabled.replace('    state_dir:', '    refresh_ms: 100\n    debounce_ms: 0\n    idle_ms: 100\n    max_lifetime_ms: 1000\n    pass_timeout_ms: 100\n    state_dir:')
        local_config.write_text(enabled, encoding='utf-8')
        state = self.root / 'capture state'
        quote = lambda value: "'" + value.replace("'", "''") + "'"
        cleanup = self.root / 'stop owned fixture workers.ps1'
        prefix = win(self.root) + '\\'
        cleanup.write_text("$ErrorActionPreference = 'Stop'\nGet-CimInstance Win32_Process | Where-Object { $_.ExecutablePath -and $_.ExecutablePath.StartsWith(" + quote(prefix) + ", [StringComparison]::OrdinalIgnoreCase) } | ForEach-Object { Stop-Process -Id $_.ProcessId -Force -ErrorAction SilentlyContinue }\n", encoding='utf-8-sig')
        self.addCleanup(lambda: self.succeeds(self.call(cleanup, [], input='')))
        receiver = self.root / 'invoke native receiver.ps1'
        receiver.write_text('& ' + quote(win(selected)) + ' @args\nexit $LASTEXITCODE\n', encoding='utf-8-sig')
        source = self.home / '.codex'; (source / 'sessions').mkdir(parents=True, exist_ok=True)
        session = 'rendered-bind-root'
        transcript = source / 'sessions' / 'binding-root.jsonl'
        transcript.write_text(json.dumps({'timestamp': '2026-01-01T00:00:00Z', 'type': 'session_meta',
                                         'payload': {'id': session, 'cwd': win(self.repo), 'timestamp': '2026-01-01T00:00:00Z'}}) + '\n', encoding='utf-8')
        arguments = ['dispatch', 'codex-hook']
        payload = {'session_id': session, 'cwd': win(self.repo), 'hook_event_name': 'SessionStart', 'transcript_path': win(transcript)}
        received = self.succeeds(self.call(receiver, arguments, input=json.dumps(payload)))
        context = json.loads(received.stdout)['hookSpecificOutput']['additionalContext']
        command = context.split('respond now by executing: ', 1)[1].split('. If no ticket', 1)[0]
        title = "Rendered O'Brien binding"
        command = command.replace('<EXPLICIT_TICKET_ID>', quote('FX-901')).replace('<EXPLICIT_TICKET_TITLE>', quote(title))
        command = command.replace('[--ticket-type <EXPLICIT_TYPE>]', "--ticket-type 'test'")
        rendered_project = context.split(' Use --repo ', 1)[1].split(' on that binding command', 1)[0]
        command += ' --repo ' + rendered_project
        binding = self.root / 'execute rendered binding.ps1'
        binding.write_text(command + '\nexit $LASTEXITCODE\n', encoding='utf-8-sig')
        bound = self.succeeds(self.call(binding, [], input=''))
        registration = json.loads(bound.stdout)
        self.assertEqual(registration['ticket_id'], 'FX-901')
        self.assertEqual(registration['ticket_title'], title)
        self.assertEqual(registration['identity']['session_id'], session)
        payload['hook_event_name'] = 'SessionEnd'
        self.succeeds(self.call(receiver, arguments, input=json.dumps(payload)))
        local_config.write_text(enabled.replace('enabled: true', 'enabled: false'), encoding='utf-8')

    def test_actual_native_dispatch_preserves_nested_cwd_relative_io_stdin_and_exit(self):
        application = self.install_native_reporting_probe()
        nested = self.repo / 'nested cwd'; nested.mkdir()
        quote = lambda value: "'" + value.replace("'", "''") + "'"
        driver = self.root / 'invoke from nested.ps1'
        driver.write_text('Set-Location -LiteralPath ' + quote(win(nested)) + '\n& ' + quote(win(self.dest / 'agentic-lens.ps1')) + ' @args\nexit $LASTEXITCODE\n', encoding='utf-8-sig')
        environment = dict(self.env, AL_PROBE_CAPTURE_APPLY=win(application), AL_PROBE_LOG='relative-result.json', AL_PROBE_EXIT='23')
        relative_args = ['fixture-cwd', '', "O'Brien", 'embedded"quote', 'tail\\', '--session-id', '',
                         '--json-output', 'relative/metrics.json', '--append', 'relative/SUMMARY.md']
        for project_args in [[], ['--al-project', win(self.repo)]]:
            for payload in ['native payload ü\n', 'first ü\nsecond\r\nthird\rlast', '']:
                with self.subTest(project_args=project_args, payload=payload):
                    result = self.call(driver, [*project_args, *relative_args], env=environment, input=payload)
                    self.assertEqual(result.returncode, 23, result.stderr)
                    expected = {'version': self.pin, 'args': relative_args, 'stdin': payload, 'cwd': win(nested)}
                    self.assertEqual(json.loads(result.stdout), expected)
                    self.assertEqual(json.loads((nested / 'relative-result.json').read_text(encoding='utf-8')), expected)
                    self.assertFalse((self.repo / 'relative-result.json').exists())
                    (nested / 'relative-result.json').unlink()

    def test_real_install_quote_argv_stdin_exit_and_local_overlay(self):
        result = self.succeeds(self.install())
        self.assertIn('No profile changed', result.stdout)
        self.assertTrue(self.native().is_file())
        literal = lambda value: "'" + value.replace("'", "''") + "'"
        path_command = "$env:PATH = " + literal(win(self.dest) + ';') + " + $env:PATH; $selected = Get-Command agentic-lens -ErrorAction Stop; if ($selected.Path -ne " + literal(win(self.dest / 'agentic-lens.ps1')) + ") { throw 'Unexpected PATH entrypoint' }; & agentic-lens --al-project " + literal(win(self.repo)) + " version --short"
        path_result = subprocess.run([POWERSHELL, '-NoLogo', '-NoProfile', '-NonInteractive', '-Command', path_command],
                                     env=self.env, input='', text=True, encoding='utf-8', capture_output=True, timeout=10)
        self.assertEqual(self.succeeds(path_result).stdout.strip(), self.pin)
        arguments = ['capture', 'receive', '--ticket-title', '--repo', '--request', "quote'\"", '--store', 'literal$()', '--session-id', '', '--wrapper', 'tail\\']
        for form, payload in [(self.pin, 'payload\n'), ('"' + self.pin + '"', 'payload ü\r\n'),
                              ("'" + self.pin + "'", 'first ü\nsecond\r\nthird\rlast'), (self.pin, '')]:
            with self.subTest(form=form, payload=payload):
                self.write_config(form)
                result = self.launch(*arguments, env=dict(self.env, AL_PROBE_EXIT='23'), input=payload)
                self.assertEqual(result.returncode, 23, result.stderr)
                output = json.loads(result.stdout)
                self.assertEqual(output['args'], arguments)
                self.assertEqual(output['stdin'], payload)
                self.assertEqual(output['version'], self.pin)
                self.assertEqual(result.stderr, '')
        self.write_config()
        plain = self.config.read_bytes()
        self.config.write_bytes(b'\xef\xbb\xbf' + plain)
        self.succeeds(self.launch())
        self.config.write_bytes(b'\xef\xbb\xbf\xef\xbb\xbf' + plain)
        self.assertNotEqual(self.launch().returncode, 0)
        self.config.write_text(f"agenticlens:\n  version: {self.pin}\n  install_root: 'C:\\missing'\n")
        override = self.repo / '.agenticlens-config.local.yaml'
        override.write_text(f"agenticlens:\n  install_root: '{win(self.cache)}'\n")
        self.succeeds(self.launch())
        override.write_text(f'agenticlens:\n  version: {self.pin}\n')
        self.assertNotEqual(self.launch().returncode, 0)
        override.unlink(); override.mkdir()
        self.assertIn('regular file', self.launch().stderr)

    def test_failed_candidate_checksum_protocol_and_wrapper_edit_preserve_install(self):
        self.succeeds(self.install())
        before = self.native().read_bytes(); wrapper = self.dest / 'agentic-lens.ps1'; wrapper_before = wrapper.read_bytes()
        original_probe = self.probe
        for protocol in ['1', '2']:
            with self.subTest(protocol=protocol):
                candidate = self.root / f'native protocol {protocol}.exe'
                shutil.copyfile(build_probe(self.pin, 'windows', 'amd64', protocol=protocol), candidate)
                self.probe = candidate
                try:
                    result = self.install('-Force')
                finally:
                    self.probe = original_probe
                self.assertNotEqual(result.returncode, 0)
                self.assertIn('required 3', result.stderr)
                self.assertEqual(self.native().read_bytes(), before)
                self.assertEqual(wrapper.read_bytes(), wrapper_before)
        invalid = self.root / 'invalid native input.bin'; invalid.write_bytes(b'not native')
        self.probe = invalid
        try:
            self.assertNotEqual(self.install('-Force').returncode, 0)
        finally:
            self.probe = original_probe
        self.assertEqual(self.native().read_bytes(), before)
        wrapper.write_text('owner custom script\n')
        self.assertNotEqual(self.install('-Force').returncode, 0)
        self.assertEqual(wrapper.read_text(), 'owner custom script\n')

    def test_checksum_without_utility_autoload_preserves_verified_install(self):
        # Keep the real filesystem cmdlets, but leave Get-FileHash unavailable.
        # This process-local control does not alter PSModulePath or user profiles.
        literal = lambda value: "'" + value.replace("'", "''") + "'"
        driver = self.root / 'checksum driver.ps1'
        driver.write_text("$ErrorActionPreference = 'Stop'\n"
                          "$PSModuleAutoLoadingPreference = 'None'\n"
                          "Import-Module Microsoft.PowerShell.Management\n"
                          "if (Get-Command Get-FileHash -ErrorAction SilentlyContinue) { throw 'Hash cmdlet must be unavailable' }\n"
                          "& " + literal(win(self.scripts / 'download.ps1')) + " @args\n"
                          "exit $LASTEXITCODE\n", encoding='utf-8-sig')
        original = self.probe
        self.probe = self.root / 'native release ü [literal].exe'
        original.rename(self.probe)
        arguments = ['-Project', win(self.repo), '-Dest', win(self.dest), '-LocalBin', win(self.probe)]
        expected = digest(self.probe)
        self.succeeds(self.call(driver, [*arguments, '-Sha256', expected.upper()]))
        native = self.native(); admitted = native.stat(); before = native.read_bytes()
        wrapper = self.dest / 'agentic-lens.ps1'; wrapper_before = wrapper.read_bytes()
        sidecar = self.dest / 'agentic-lens.ps1.sha256'
        self.assertEqual(sidecar.read_text(), digest(self.scripts / 'agentic-lens-wrapper.ps1'))
        self.assertRegex(sidecar.read_text(), r'^[0-9a-f]{64}$')
        # Replacing a checksum failure with unchecked acceptance must fail here.
        rejected = self.call(driver, [*arguments, '-Sha256', '0' * 64, '-Force'])
        self.assertNotEqual(rejected.returncode, 0)
        self.assertIn('matching explicit expected -Sha256', rejected.stderr)
        missing_arguments = [*arguments[:-1], win(self.root / 'missing ü [literal].exe'), '-Sha256', expected]
        missing = self.call(driver, missing_arguments)
        self.assertNotEqual(missing.returncode, 0)
        self.assertEqual(native.read_bytes(), before)
        self.assertTrue(os.path.samestat(admitted, native.stat()))
        self.assertEqual(wrapper.read_bytes(), wrapper_before)
        self.assertEqual(sidecar.read_text(), digest(self.scripts / 'agentic-lens-wrapper.ps1'))
        self.assertFalse(list(self.cache.glob('.install-stage.*')))
        self.assertFalse(list(self.cache.glob('.install-*.lock')))
        self.assertFalse((self.dest / '.agenticlens-wrapper-install.lock').exists())
        self.succeeds(self.call(driver, [*arguments, '-Sha256', expected, '-Force']))
        self.assertTrue(os.path.samestat(admitted, native.stat()))
        self.assertEqual(self.succeeds(self.launch('version', '--short')).stdout.strip(), self.pin)

    def test_hash_stream_closes_for_literal_paths_and_open_failures(self):
        # Load the shipped function through PowerShell's parser, without running
        # the installer, so an undisposed handle cannot hide behind process exit.
        fixture = self.root / 'hash bytes ü [literal].bin'
        fixture.write_bytes(bytes(range(256)) * 8192)
        literal = lambda value: "'" + value.replace("'", "''") + "'"
        driver = self.root / 'hash stream driver.ps1'
        driver.write_text("$ErrorActionPreference = 'Stop'\n"
                          "$PSModuleAutoLoadingPreference = 'None'\n"
                          "$tokens = $null; $errors = $null\n"
                          "$ast = [Management.Automation.Language.Parser]::ParseFile(" + literal(win(self.scripts / 'download.ps1')) + ", [ref]$tokens, [ref]$errors)\n"
                          "if ($errors.Count) { throw 'Installer parse failed' }\n"
                          "$definition = $ast.Find({ param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -ceq 'Hash-File' }, $true)\n"
                          "if ($null -eq $definition) { throw 'Missing shipped hash function' }\n"
                          ". ([ScriptBlock]::Create($definition.Extent.Text))\n"
                          "$path = " + literal(win(fixture)) + "\n"
                          "$value = Hash-File $path\n"
                          "if ($value -isnot [string] -or $value -cnotmatch '^[0-9a-f]{64}$') { throw 'Unexpected hash output' }\n"
                          "[Console]::WriteLine($value)\n"
                          "[void]$ExecutionContext.SessionState.Path.SetLocation(" + literal(win(self.root)) + ")\n"
                          "try { $relative = Hash-File " + literal(fixture.name) + " } catch { throw 'Relative PowerShell path hash failed' }\n"
                          "if ($relative -cne $value) { throw 'Relative PowerShell path hash differs' }\n"
                          "$exclusive = [IO.File]::Open($path, [IO.FileMode]::Open, [IO.FileAccess]::ReadWrite, [IO.FileShare]::None)\n"
                          "try {\n"
                          "  $failed = $false\n"
                          "  try { Hash-File $path } catch { $failed = $true }\n"
                          "  if (-not $failed) { throw 'Locked file hash was accepted' }\n"
                          "} finally { $exclusive.Dispose() }\n"
                          "$exclusive = [IO.File]::Open($path, [IO.FileMode]::Open, [IO.FileAccess]::ReadWrite, [IO.FileShare]::None)\n"
                          "$exclusive.Dispose()\n"
                          "$missing = $path + '.missing'; $failed = $false\n"
                          "try { Hash-File $missing } catch { $failed = $true }\n"
                          "if (-not $failed -or [IO.File]::Exists($missing)) { throw 'Missing file hash was accepted or created' }\n",
                          encoding='utf-8-sig')
        result = self.succeeds(self.call(driver, []))
        self.assertEqual(result.stdout.strip(), digest(fixture))
        self.assertEqual(result.stderr, '')

    def test_legacy_binary_only_install_reports_missing_new_capability(self):
        shutil.copyfile(build_probe(self.pin, 'windows', 'amd64', protocol='none'), self.probe)
        result = self.succeeds(self.install())
        self.assertIn('hook protocol 3 and guarded init unavailable', result.stdout)
        self.assertEqual(self.launch('version', '--short').stdout.strip(), self.pin)
        store = self.root / 'source store'; store.mkdir()
        result = self.install('-CaptureSource', 'codex', '-CaptureStore', win(store))
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('required 3', result.stderr)
        self.assertFalse((self.repo / '.codex').exists())

    def test_duplicate_unsafe_json_mismatch_and_missing_native_reject_without_dispatch(self):
        self.succeeds(self.install())
        for document in ['agenticlens: {version: v9.8.7}\n', 'agenticlens:\n  version: \u00a0v9.8.7\u00a0\n', 'agenticlens:\n  version: v9.8.7\n  version: v9.8.7\n',
                         'agenticlens:\n  version: "v9.8.7+$(touch INJECTED)"\n']:
            self.config.write_text(document)
            result = self.launch('capture', 'receive'); self.assertEqual(result.returncode, 2); self.assertEqual(result.stdout, '')
        self.write_config()
        other = self.root / 'project B'; other.mkdir(); (other / '.agenticlens-config.yaml').write_bytes(self.config.read_bytes())
        self.assertIn('--al-project', self.launch('capture', 'status', '--repo', win(other)).stderr)
        self.config.rename(self.repo / '.agenticlens-config.json')
        self.assertIn('migrate', self.launch().stderr)
        (self.repo / '.agenticlens-config.json').unlink(); self.write_config(); self.native().unlink()
        self.assertIn('Missing pinned native', self.launch().stderr)
        self.assertFalse(self.native().exists())

    def test_offline_advisory_timeout_kills_actual_native_child(self):
        self.succeeds(self.install())
        pid_file = self.root / 'native.pid'
        for adopted in [False, True]:
            with self.subTest(native_dispatcher_adopted=adopted):
                environment = dict(self.env, AL_PROBE_HANG='1', AL_PROBE_PID=win(pid_file))
                if adopted:
                    application = self.install_native_reporting_probe()
                    environment['AL_PROBE_CAPTURE_APPLY'] = win(application)
                started = time.monotonic()
                result = self.launch('capture', 'check', '--repo', win(self.repo), env=environment)
                elapsed = time.monotonic() - started
                self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
                self.assertIn('timed out', result.stderr)
                self.assertLess(elapsed, 4)
                pid = int(pid_file.read_text())
                probe = subprocess.run([POWERSHELL, '-NoLogo', '-NoProfile', '-NonInteractive', '-Command',
                                        f'if (Get-Process -Id {pid} -ErrorAction SilentlyContinue) {{ exit 1 }}; exit 0'], capture_output=True)
                self.assertEqual(probe.returncode, 0, 'native child survived wrapper timeout')
                pid_file.unlink()
                print(f'Windows native check PID {pid} terminated; wrapper returned {result.returncode} after {elapsed:.3f}s')

    def test_windows_junction_dotdot_selects_physical_project_pin(self):
        self.succeeds(self.install())
        physical = self.root / 'physical project'; physical.mkdir()
        child = physical / 'child'; child.mkdir()
        other_pin = 'v1.2.3'
        (physical / '.agenticlens-config.yaml').write_text(f"agenticlens:\n  version: {other_pin}\n  install_root: '{win(self.cache)}'\n")
        native = self.cache / 'bin' / other_pin / 'windows-amd64' / 'agentic-lens-windows-amd64.exe'
        native.parent.mkdir(parents=True)
        shutil.copyfile(build_probe(other_pin, 'windows', 'amd64'), native)
        link = self.repo / 'junction'
        quote = lambda value: "'" + value.replace("'", "''") + "'"
        command = 'New-Item -ItemType Junction -Path ' + quote(win(link)) + ' -Target ' + quote(win(child)) + ' | Out-Null'
        subprocess.run([POWERSHELL, '-NoLogo', '-NoProfile', '-NonInteractive', '-Command', command], check=True, capture_output=True)
        result = self.succeeds(self.call(self.dest / 'agentic-lens.ps1', ['--al-project', win(link) + '\\..']))
        self.assertEqual(json.loads(result.stdout)['version'], other_pin)
        result = self.launch('capture', 'status', '--repo', win(link) + '\\..')
        self.assertIn('--al-project', result.stderr)


if __name__ == '__main__':
    unittest.main(verbosity=2)
