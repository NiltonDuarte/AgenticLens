"""Exercise shipped shell entrypoints and their native capture command boundary."""

import json
import os
from pathlib import Path
import platform
import shutil
import subprocess
import sys
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
DOWNLOADERS = [ROOT / "scripts/download-agentic-lens.sh", ROOT / ".agents/skills/agentic-lens-setup/scripts/download.sh"]
WRAPPERS = [ROOT / "scripts/agentic-lens-wrapper.sh", ROOT / ".agents/skills/agentic-lens-setup/scripts/agentic-lens-wrapper.sh"]


def native_name():
    operating_system = {"Linux": "linux", "Darwin": "darwin", "Windows": "windows"}[platform.system()]
    arch = {"x86_64": "amd64", "AMD64": "amd64", "arm64": "arm64", "aarch64": "arm64"}[platform.machine()]
    return f"agentic-lens-{operating_system}-{arch}" + (".exe" if operating_system == "windows" else "")


class CaptureDistributionTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="capture distribution ")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.repo = self.root / "adopting repo"
        self.repo.mkdir()
        self.dest = self.repo / "bin space"
        self.store = self.root / "source $(touch INJECTED)"
        self.store.mkdir()
        self.log = self.root / "native-args.json"
        self.release = self.root / "local release"
        self.release.write_text(
            f"#!{sys.executable}\n"
            "import json, os, sys\n"
            "from pathlib import Path\n"
            "if sys.argv[1:] == ['capture', 'protocol']: print(1); sys.exit(0)\n"
            "Path(os.environ['CAPTURE_TEST_LOG']).write_text(json.dumps(sys.argv[1:]))\n"
            "sys.exit(int(os.environ.get('CAPTURE_TEST_EXIT', '0')))\n"
        )
        self.release.chmod(0o755)
        self.env = dict(os.environ, CAPTURE_TEST_LOG=str(self.log), AGENTIC_LENS_NO_UPDATE_CHECK="0")
        # A network invocation is always observable; these tests never need network.
        self.network = self.root / "commands"
        self.network.mkdir()
        self.network_log = self.root / "network-called"
        for command in ["curl", "wget"]:
            path = self.network / command
            path.write_text("#!/bin/sh\nprintf called > \"$CAPTURE_TEST_NETWORK\"\nexit 1\n")
            path.chmod(0o755)
        self.env.update(PATH=str(self.network) + os.pathsep + self.env["PATH"], CAPTURE_TEST_NETWORK=str(self.network_log))

    def run_download(self, script, *options, **env):
        return subprocess.run(["bash", str(script), "--dest", str(self.dest), "--local-bin", str(self.release), *options],
                              cwd=self.repo, env=dict(self.env, **env), text=True, capture_output=True, timeout=10)

    def test_both_downloaders_install_and_delegate_exact_arguments_to_native_handler(self):
        for script in DOWNLOADERS:
            with self.subTest(script=script):
                self.log.unlink(missing_ok=True)
                result = self.run_download(script, "--capture-source", "codex", "--capture-store", str(self.store))
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(json.loads(self.log.read_text()), ["capture", "install", "--source", "codex", "--repo", str(self.repo),
                                 "--state-dir", str(self.repo / ".agenticlens-capture"), "--store", str(self.store)])
                self.assertTrue((self.dest / native_name()).is_file())
                self.assertFalse((self.repo / "INJECTED").exists())
                self.assertFalse(self.network_log.exists())
                self.assertIn("runtime trust, agent binding", result.stdout)

    def test_binary_only_and_invalid_capture_options_never_install_hooks(self):
        for script in DOWNLOADERS:
            with self.subTest(script=script):
                result = self.run_download(script)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertIn("Binary installation only", result.stdout)
                self.assertFalse(self.log.exists())
                for options in [("--capture-source", "codex"), ("--capture-source", "other", "--capture-store", str(self.store)),
                                ("--capture-store", str(self.store)), ("--capture-source", "opencode", "--capture-store", "relative.db"),
                                ("--capture-source",)]:
                    rejected = self.run_download(script, *options)
                    self.assertNotEqual(rejected.returncode, 0, options)
                    self.assertFalse(self.log.exists())

    def test_dry_run_previews_native_capture_and_preserves_destination(self):
        for script in DOWNLOADERS:
            with self.subTest(script=script):
                self.log.unlink(missing_ok=True)
                state = self.root / "custom local state"
                result = self.run_download(script, "--capture-source", "opencode", "--capture-store", str(self.store),
                                           "--capture-state-dir", str(state), "--capture-repo", str(self.repo), "--dry-run")
                self.assertEqual(result.returncode, 0, result.stderr)
                args = json.loads(self.log.read_text())
                self.assertEqual(args[-1], "--dry-run")
                self.assertEqual(args[args.index("--state-dir") + 1], str(state))
                self.assertFalse(self.dest.exists())
                self.assertFalse(state.exists())
                self.assertFalse(self.network_log.exists())

    def test_native_setup_failure_is_reported_without_claiming_lifecycle_success(self):
        for script in DOWNLOADERS:
            with self.subTest(script=script):
                result = self.run_download(script, "--capture-source", "codex", "--capture-store", str(self.store), CAPTURE_TEST_EXIT="17")
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("Capture setup failed", result.stderr)
                self.assertNotIn("Lifecycle configuration installed", result.stdout)
                self.assertFalse((self.dest / native_name()).exists())

    def install_active_fixture(self):
        self.dest.mkdir(exist_ok=True)
        native = self.dest / native_name()
        shutil.copyfile(self.release, native)
        native.chmod(0o755)
        hook = self.repo / ".codex/hooks.json"
        hook.parent.mkdir(exist_ok=True)
        hook.write_text(json.dumps({"hooks": {"SessionStart": [{"hooks": [{"command": str(native), "type": "command"}]}]}}))
        state = self.repo / ".agenticlens-capture"
        state.mkdir(exist_ok=True)
        (state / "install-codex.json").write_text(json.dumps({"active": True, "binary": str(native), "protocol": 1}))
        return native, hook

    def candidate(self, compatible=True, fail_install=False):
        candidate = self.root / "replacement release"
        candidate.write_text(
            f"#!{sys.executable}\nimport sys\n"
            f"if sys.argv[1:] == ['capture', 'protocol']: print({1 if compatible else 2}); sys.exit(0)\n"
            "if sys.argv[1:2] == ['version']: print('v-fixture'); sys.exit(0)\n"
            f"sys.exit({17 if fail_install else 0})\n"
        )
        candidate.chmod(0o755)
        return candidate

    def transport_candidate(self, candidate):
        curl = self.network / "curl"
        curl.write_text(
            f"#!{sys.executable}\nimport os, shutil, sys\n"
            "from pathlib import Path\n"
            "Path(os.environ['CAPTURE_TEST_NETWORK']).write_text(sys.argv[-1])\n"
            "shutil.copyfile(os.environ['CAPTURE_TEST_CANDIDATE'], sys.argv[sys.argv.index('-o') + 1])\n"
        )
        curl.chmod(0o755)
        return dict(self.env, CAPTURE_TEST_CANDIDATE=str(candidate), AGENTIC_LENS_LOCAL_BINARY="")

    def test_capture_upgrade_rejects_incompatible_local_and_downloaded_candidates_before_replacement(self):
        for script in DOWNLOADERS:
            for downloaded in [False, True]:
                for configure in [False, True]:
                    with self.subTest(script=script, downloaded=downloaded, configure=configure):
                        native, hook = self.install_active_fixture()
                        old_binary, old_hook = native.read_bytes(), hook.read_bytes()
                        replacement = self.candidate(compatible=False)
                        args = ["bash", str(script), "--dest", str(self.dest), "--force", "--version", "v-fixture"]
                        if not downloaded:
                            args += ["--local-bin", str(replacement)]
                        if configure:
                            args += ["--capture-source", "codex", "--capture-store", str(self.store)]
                        result = subprocess.run(args, cwd=self.repo, env=self.transport_candidate(replacement), text=True, capture_output=True, timeout=10)
                        self.assertNotEqual(result.returncode, 0)
                        self.assertIn("Candidate capture protocol incompatible", result.stderr)
                        self.assertEqual(native.read_bytes(), old_binary)
                        self.assertEqual(hook.read_bytes(), old_hook)
                        self.assertTrue(json.loads((self.repo / ".agenticlens-capture/install-codex.json").read_text())["active"])
                        self.assertFalse(list(self.dest.glob("*.previous.*")))
                        self.assertFalse(list(self.dest.glob("*.candidate.*")))

    def test_later_capture_setup_failure_restores_previous_native_for_both_replacement_paths(self):
        for script in DOWNLOADERS:
            for downloaded in [False, True]:
                with self.subTest(script=script, downloaded=downloaded):
                    native, hook = self.install_active_fixture()
                    old_binary, old_hook = native.read_bytes(), hook.read_bytes()
                    old_wrapper = self.dest / "agentic-lens"
                    old_wrapper.write_text("#!/bin/sh\necho original-user-wrapper\n")
                    wrapper_bytes = old_wrapper.read_bytes()
                    replacement = self.candidate(fail_install=True)
                    args = ["bash", str(script), "--dest", str(self.dest), "--force", "--version", "v-fixture",
                            "--capture-source", "codex", "--capture-store", str(self.store)]
                    if not downloaded:
                        args += ["--local-bin", str(replacement)]
                    result = subprocess.run(args, cwd=self.repo, env=self.transport_candidate(replacement), text=True, capture_output=True, timeout=10)
                    self.assertNotEqual(result.returncode, 0)
                    self.assertIn("Capture setup failed", result.stderr)
                    self.assertEqual(old_wrapper.read_bytes(), wrapper_bytes)
                    self.assertEqual(native.read_bytes(), old_binary)
                    self.assertEqual(hook.read_bytes(), old_hook)
                    self.assertFalse(list(self.dest.glob("*.previous.*")))
                    self.assertFalse(list(self.dest.glob("*.candidate.*")))

    def test_wrapper_upgrade_preflights_protocol_and_preserves_active_hooks(self):
        for script in WRAPPERS:
            for compatible in [False, True]:
                with self.subTest(script=script, compatible=compatible):
                    native, hook = self.install_active_fixture()
                    wrapper = self.dest / "agentic-lens"
                    shutil.copyfile(script, wrapper)
                    old_binary, old_hook = native.read_bytes(), hook.read_bytes()
                    replacement = self.candidate(compatible=compatible)
                    result = subprocess.run(["sh", str(wrapper), "update", "--version", "v-fixture"], cwd=self.repo,
                                            env=self.transport_candidate(replacement), text=True, capture_output=True, timeout=10)
                    self.assertEqual(result.returncode, 0 if compatible else 1, result.stderr)
                    self.assertEqual(native.read_bytes(), replacement.read_bytes() if compatible else old_binary)
                    self.assertEqual(hook.read_bytes(), old_hook)
                    if not compatible:
                        self.assertIn("Candidate capture protocol incompatible", result.stderr)
                        self.assertNotIn("Successfully updated", result.stdout)
                    self.assertFalse(list(self.dest.glob(".agenticlens-update.*")))

    def test_activation_failure_restores_owned_binary_transaction(self):
        real_mv = shutil.which("mv")
        mv = self.network / "mv"
        mv.write_text(
            f"#!{sys.executable}\nimport os, sys\n"
            "if sys.argv[-1] == os.environ['CAPTURE_TEST_FAIL_TARGET'] and '/previous/' not in sys.argv[-2] and '.previous.' not in sys.argv[-2]: sys.exit(23)\n"
            f"os.execv({real_mv!r}, [{real_mv!r}, *sys.argv[1:]])\n"
        )
        mv.chmod(0o755)
        for script in [*DOWNLOADERS, *WRAPPERS]:
            with self.subTest(script=script):
                native, hook = self.install_active_fixture()
                before, hook_before = native.read_bytes(), hook.read_bytes()
                replacement = self.candidate()
                if script in DOWNLOADERS:
                    args = ["bash", str(script), "--dest", str(self.dest), "--force", "--version", "v-fixture"]
                else:
                    wrapper = self.dest / "agentic-lens"
                    shutil.copyfile(script, wrapper)
                    args = ["sh", str(wrapper), "update", "--version", "v-fixture"]
                env = dict(self.transport_candidate(replacement), CAPTURE_TEST_FAIL_TARGET=str(native))
                result = subprocess.run(args, cwd=self.repo, env=env, text=True, capture_output=True, timeout=10)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(native.read_bytes(), before)
                self.assertEqual(hook.read_bytes(), hook_before)
                self.assertFalse(list(self.dest.glob("*.previous.*")))
                self.assertFalse(list(self.dest.glob(".agenticlens-update.*")))

    def test_capture_aware_upgrade_preserves_unverifiable_foreign_platform_binary(self):
        for script in [*DOWNLOADERS, *WRAPPERS]:
            with self.subTest(script=script):
                native, _ = self.install_active_fixture()
                foreign = self.dest / ("agentic-lens-darwin-arm64" if "darwin" not in native_name() else "agentic-lens-linux-amd64")
                foreign.write_bytes(b"foreign-host-handler")
                foreign.chmod(0o755)
                replacement = self.candidate()
                if script in DOWNLOADERS:
                    args = ["bash", str(script), "--dest", str(self.dest), "--force", "--all-platforms", "--version", "v-fixture"]
                else:
                    wrapper = self.dest / "agentic-lens"
                    shutil.copyfile(script, wrapper)
                    args = ["sh", str(wrapper), "update", "--version", "v-fixture"]
                result = subprocess.run(args, cwd=self.repo, env=self.transport_candidate(replacement), text=True, capture_output=True, timeout=10)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(foreign.read_bytes(), b"foreign-host-handler")
                self.assertIn("native host", result.stdout)
                for other in self.dest.glob("agentic-lens-*"):
                    if other != native:
                        other.unlink()

    def test_download_failure_never_attempts_hook_installation(self):
        for script in DOWNLOADERS:
            with self.subTest(script=script):
                result = subprocess.run(["bash", str(script), "--dest", str(self.dest), "--capture-source", "codex",
                                         "--capture-store", str(self.store), "--version", "v-fixture"],
                                        cwd=self.repo, env=dict(self.env, AGENTIC_LENS_LOCAL_BINARY=""),
                                        text=True, capture_output=True, timeout=10)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("Failed to download", result.stderr)
                self.assertFalse(self.log.exists())
                self.assertFalse((self.repo / ".codex/hooks.json").exists())
                self.assertNotIn("Lifecycle configuration installed", result.stdout)

    def test_advertised_platform_download_plans_use_correct_native_artifact_names(self):
        uname = self.network / "uname"
        uname.write_text('#!/bin/sh\nif [ "$1" = "-s" ]; then printf "%s\\n" "$CAPTURE_TEST_OS"; else printf "%s\\n" "$CAPTURE_TEST_ARCH"; fi\n')
        uname.chmod(0o755)
        for script in DOWNLOADERS:
            for os_name, arch, expected in [("Linux", "x86_64", "linux-amd64"), ("Linux", "aarch64", "linux-arm64"),
                                           ("Darwin", "x86_64", "darwin-amd64"), ("Darwin", "arm64", "darwin-arm64"),
                                           ("MINGW64_NT", "x86_64", "windows-amd64.exe")]:
                with self.subTest(script=script, target=expected):
                    result = self.run_download(script, "--dry-run", CAPTURE_TEST_OS=os_name, CAPTURE_TEST_ARCH=arch)
                    self.assertEqual(result.returncode, 0, result.stderr)
                    self.assertIn("agentic-lens-" + expected, result.stdout)
                    self.assertFalse(self.dest.exists())
                    self.assertFalse(self.network_log.exists())

    def test_a_wrapper_supplied_as_local_native_artifact_fails_without_recursion(self):
        self.release = WRAPPERS[0]
        for script in DOWNLOADERS:
            with self.subTest(script=script):
                result = self.run_download(script, "--capture-source", "codex", "--capture-store", str(self.store))
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("Candidate capture protocol incompatible", result.stderr)
                self.assertFalse((self.repo / ".codex/hooks.json").exists())
                self.assertFalse(self.network_log.exists())

    def test_wrapper_candidate_cannot_probe_through_working_native_and_replace_it(self):
        for script in DOWNLOADERS:
            with self.subTest(script=script):
                native, hook = self.install_active_fixture()
                before, hook_before = native.read_bytes(), hook.read_bytes()
                result = subprocess.run(["bash", str(script), "--dest", str(self.dest), "--local-bin", str(WRAPPERS[0]), "--force"],
                                        cwd=self.repo, env=self.env, text=True, capture_output=True, timeout=10)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("Candidate capture protocol incompatible", result.stderr)
                self.assertEqual(native.read_bytes(), before)
                self.assertEqual(hook.read_bytes(), hook_before)
                self.assertFalse(self.network_log.exists())

    def test_wrapper_capture_is_network_free_with_and_without_native_handler(self):
        self.dest.mkdir()
        for script in WRAPPERS:
            with self.subTest(script=script):
                wrapper = self.dest / "agentic-lens"
                shutil.copyfile(script, wrapper)
                native = self.dest / native_name()
                native.unlink(missing_ok=True)
                missing = subprocess.run(["sh", str(wrapper), "capture", "status"], cwd=self.repo, env=self.env, text=True, capture_output=True)
                self.assertNotEqual(missing.returncode, 0)
                self.assertIn("Native capture handler missing", missing.stderr)
                self.assertFalse(self.network_log.exists())
                shutil.copyfile(self.release, native)
                native.chmod(0o755)
                installed = subprocess.run(["sh", str(wrapper), "capture", "status"], cwd=self.repo, env=self.env, text=True, capture_output=True)
                self.assertEqual(installed.returncode, 0, installed.stderr)
                self.assertEqual(json.loads(self.log.read_text()), ["capture", "status"])
                self.assertFalse(self.network_log.exists())
                self.assertFalse((self.dest / ".last_update_check").exists())


if __name__ == "__main__":
    unittest.main()
