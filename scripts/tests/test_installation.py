"""Exercise installed production scripts in isolated adopting repositories/worktrees."""
import json
import os
from pathlib import Path
import platform
import shlex
import shutil
import subprocess
import sys
import tempfile
import unittest


SOURCE = Path(__file__).resolve().parents[2]
SKILLS = Path(".agents/skills")
SETUP = SKILLS / "agentic-lens-setup"
OPERATOR = SKILLS / "agentic-lens-operator"
INSTALL = SETUP / "scripts/install.sh"
DOWNLOAD = SETUP / "scripts/download.sh"
LAUNCH = OPERATOR / "scripts/agentic-lens.sh"
FILES = [
    SETUP / "SKILL.md", INSTALL, DOWNLOAD,
    SETUP / "scripts/agentic-lens-wrapper.sh", SETUP / "RELEASE_VERSION",
    OPERATOR / "SKILL.md", OPERATOR / "references/cli.md",
    OPERATOR / "references/configuration.md", OPERATOR / "references/dashboard.md",
    LAUNCH,
]


class InstallationTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="agenticlens install ")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.repo = self.root / "adopting repository"
        self.repo.mkdir()
        self.git("init", "-q")
        self.git("config", "user.name", "Installation Fixture")
        self.git("config", "user.email", "fixture@example.invalid")
        self.config = self.repo / ".agenticlens-config.yaml"
        self.config.write_text("project_name: Fixture\nworkspaces:\n  dir: custom-artifacts\n")
        self.git("add", ".agenticlens-config.yaml")
        self.git("commit", "-qm", "fixture baseline")
        self.transport = self.root / "controlled transport"
        self.transport.mkdir()
        transport_script = "#!" + sys.executable + "\n" + r'''
import json, os, pathlib, shutil, sys
args = sys.argv[1:]
url = args[-1]
with open(os.environ["FETCH_LOG"], "a") as stream:
    stream.write(json.dumps({"url": url, "args": args}) + "\n")
if "-w" in args:
    if os.environ.get("FETCH_MODE") == "online":
        print("https://github.com/NiltonDuarte/AgenticLens/releases/tag/v9.8.7", end="")
        sys.exit(0)
    sys.exit(23)
flag = "-o" if "-o" in args else "-O"
target = pathlib.Path(args[args.index(flag) + 1])
if os.environ.get("FETCH_MODE") != "online" or os.environ.get("FETCH_FAIL", "NEVER") in url:
    target.write_text("partial transfer")
    sys.exit(23)
marker = "/.agents/skills/"
if marker in url:
    shutil.copyfile(pathlib.Path(os.environ["FIXTURE_SOURCE"]) / ".agents/skills" / url.split(marker, 1)[1], target)
else:
    shutil.copyfile(os.environ["FIXTURE_BINARY"], target)
'''
        for name in ["curl", "wget"]:
            path = self.transport / name
            path.write_text(transport_script)
            path.chmod(0o755)
        self.binary = self.root / "prebuilt fixture"
        self.binary.write_text("#!" + sys.executable + "\n" + '''
import json, os, sys
if sys.argv[1:] == ["version", "--short"]:
    print("v9.8.7")
elif sys.argv[1:] == ["version", "--json"]:
    print(json.dumps({"version": "v9.8.7", "commit": "fixture", "build_date": "2026-01-01T00:00:00Z"}))
else:
    print(json.dumps({"cwd": os.getcwd(), "args": sys.argv[1:]}))
sys.exit(int(os.environ.get("FIXTURE_EXIT", "0")))
''')
        self.binary.chmod(0o755)
        self.log = self.root / "transport.jsonl"
        self.env = {key: value for key, value in os.environ.items()
                    if not key.startswith("AGENTIC_LENS_")}
        self.env.update(PATH=str(self.transport) + os.pathsep + os.environ["PATH"],
                        FETCH_LOG=str(self.log), FETCH_MODE="offline",
                        FIXTURE_SOURCE=str(SOURCE), FIXTURE_BINARY=str(self.binary),
                        AGENTIC_LENS_NO_UPDATE_CHECK="1")
        os_name = {"Linux": "linux", "Darwin": "darwin"}[platform.system()]
        arch = {"x86_64": "amd64", "amd64": "amd64", "arm64": "arm64",
                "aarch64": "arm64"}[platform.machine().lower()]
        self.native_name = f"agentic-lens-{os_name}-{arch}"

    def run_script(self, path, *args, cwd=None, **env):
        return subprocess.run(["bash", str(path), *map(str, args)],
                              cwd=cwd or self.repo, env=dict(self.env, **env),
                              capture_output=True, text=True, timeout=20)

    def git(self, *args, cwd=None):
        return subprocess.check_output(["git", *args], cwd=cwd or self.repo,
                                       stderr=subprocess.PIPE, text=True).strip()

    def succeeds(self, result):
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        return result

    def install(self, repo=None):
        repo = repo or self.repo
        self.succeeds(self.run_script(SOURCE / INSTALL, "--source-dir", SOURCE,
                                     "--version", "v9.8.7", "--dest", repo))

    def repair(self, repo=None, *args, **env):
        repo = repo or self.repo
        return self.run_script(repo / DOWNLOAD, *args, cwd=repo, **env)

    def native(self, repo=None):
        return (repo or self.repo) / OPERATOR / "bin" / self.native_name

    def requests(self):
        return [json.loads(line) for line in self.log.read_text().splitlines()] if self.log.exists() else []

    def test_install_and_repair_preserve_owner_git_policy_and_configuration(self):
        config_before = self.config.read_bytes()
        index_before = (self.repo / ".git/index").read_bytes()
        self.install()
        for relative in FILES:
            self.assertTrue((self.repo / relative).is_file(), str(relative))
        self.assertEqual((self.repo / SETUP / "RELEASE_VERSION").read_text(), "v9.8.7\n")
        self.succeeds(self.repair(None, "--local-bin", self.binary))
        self.assertFalse(list(self.repo.rglob(".gitignore")))
        self.assertEqual(index_before, (self.repo / ".git/index").read_bytes())
        self.assertEqual(config_before, self.config.read_bytes())
        self.assertFalse((self.repo / ".agenticlens-config.local.yaml").exists())
        self.assertIn(str(OPERATOR / "bin" / self.native_name), self.git("ls-files", "--others", "--exclude-standard"))
        self.assertEqual(self.requests(), [])

    def test_fresh_worktree_repairs_locally_and_preserves_cwd_arguments_and_exit(self):
        policy = ".agents/skills/agentic-lens-operator/bin/\n.agenticlens-config.local.yaml\n"
        (self.repo / ".gitignore").write_text(policy)
        self.install()
        self.succeeds(self.repair(None, "--local-bin", self.binary))
        original_binary = self.native().read_bytes()
        self.git("add", ".agents", ".gitignore")
        self.git("commit", "-qm", "owner chooses tracked skills and ignored binaries")
        worktree = self.root / "second worktree with spaces"
        self.git("worktree", "add", "-qb", "fixture-worktree", str(worktree))
        self.assertTrue((worktree / INSTALL).is_file())
        self.assertFalse(self.native(worktree).exists())
        missing = self.run_script(worktree / LAUNCH, "--help", cwd=worktree)
        self.assertNotEqual(missing.returncode, 0)
        self.assertIn("download.sh", missing.stderr)
        self.assertEqual(self.requests(), [])
        self.succeeds(self.repair(worktree, "--local-bin", self.binary))
        args = ["generate", "--source", "codex", "--codex-home", "home with spaces", "literal;$argument"]
        result = self.run_script(worktree / LAUNCH, *args, cwd=worktree, FIXTURE_EXIT="37")
        self.assertEqual(result.returncode, 37, result.stderr)
        self.assertEqual(json.loads(result.stdout), {"cwd": str(worktree), "args": args})
        self.assertEqual(self.native().read_bytes(), original_binary)
        self.assertEqual((worktree / ".gitignore").read_text(), policy)
        self.assertEqual(self.git("status", "--porcelain", cwd=worktree), "")
        self.assertFalse((worktree / ".agenticlens-config.local.yaml").exists())
        self.assertEqual((worktree / ".agenticlens-config.yaml").read_bytes(), self.config.read_bytes())
        self.assertEqual(self.requests(), [])

    def test_missing_setup_assets_offer_bootstrap_without_network(self):
        self.install()
        self.git("add", str(LAUNCH), str(OPERATOR / "SKILL.md"))
        self.git("commit", "-qm", "owner chooses only launcher and operator instructions")
        worktree = self.root / "worktree without setup"
        self.git("worktree", "add", "-qb", "without-setup", str(worktree))
        self.assertFalse((worktree / DOWNLOAD).exists())
        result = self.run_script(worktree / LAUNCH, "version", cwd=worktree)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("bootstrap", result.stderr.lower())
        self.assertIn("https://", result.stderr)
        self.assertEqual(self.requests(), [])
        self.succeeds(self.run_script(SOURCE / INSTALL, "--dest", worktree,
                                     "--version", "v9.8.7", FETCH_MODE="online"))
        self.assertTrue((worktree / DOWNLOAD).is_file())

    def test_pinned_remote_install_and_repair_use_one_release(self):
        self.succeeds(self.run_script(SOURCE / INSTALL, "--dest", self.repo,
                                     "--version", "v9.8.7", FETCH_MODE="online"))
        self.succeeds(self.repair(None, FETCH_MODE="online"))
        requests = self.requests()
        self.assertEqual(len(requests), 10)
        self.assertTrue(all("/v9.8.7/" in request["url"] for request in requests))
        self.assertIn("/releases/download/v9.8.7/" + self.native_name, requests[-1]["url"])
        self.assertFalse(any("/main/" in request["url"] or "/latest/" in request["url"] for request in requests))
        self.succeeds(self.run_script(self.repo / LAUNCH, "version", "--json"))

    def test_latest_is_resolved_once_then_pinned_for_binary_repair(self):
        self.succeeds(self.run_script(SOURCE / INSTALL, "--dest", self.repo,
                                     FETCH_MODE="online"))
        self.succeeds(self.repair(None, FETCH_MODE="online"))
        urls = [request["url"] for request in self.requests()]
        self.assertEqual(urls[0], "https://github.com/NiltonDuarte/AgenticLens/releases/latest")
        self.assertEqual(sum("/releases/latest" in url for url in urls), 1)
        self.assertTrue(all("/v9.8.7/" in url for url in urls[1:]))

    def test_failed_download_preserves_existing_binary_and_installed_offline_launch(self):
        self.install()
        missing = self.repair()
        self.assertNotEqual(missing.returncode, 0)
        self.assertIn(self.native_name, missing.stderr)
        self.assertIn("Retry: bash", missing.stderr)
        self.assertIn("download.sh", missing.stderr)
        self.assertFalse(self.native().exists())
        self.assertNotIn("Successfully installed", missing.stdout)
        self.succeeds(self.repair(None, "--local-bin", self.binary))
        before = self.native().read_bytes()
        failed = self.repair(None, "--force", "--version", "v9.8.8")
        self.assertNotEqual(failed.returncode, 0)
        self.assertEqual(self.native().read_bytes(), before)
        self.assertFalse(list(self.native().parent.glob("*.tmp.*")))
        request_count = len(self.requests())
        result = self.succeeds(self.run_script(self.repo / LAUNCH, "offline"))
        self.assertEqual(json.loads(result.stdout)["args"], ["offline"])
        self.assertEqual(len(self.requests()), request_count)

    def test_conflicting_skill_preflight_does_not_partially_install_or_overwrite(self):
        self.install()
        edited = self.repo / SETUP / "SKILL.md"
        edited.write_text("Owner-authored local skill changes\n")
        absent = self.repo / LAUNCH
        absent.unlink()
        result = self.run_script(SOURCE / INSTALL, "--dest", self.repo,
                                 "--source-dir", SOURCE, "--version", "v9.8.7")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("differs", result.stderr)
        self.assertEqual(edited.read_text(), "Owner-authored local skill changes\n")
        self.assertFalse(absent.exists())

    def test_custom_destination_legacy_wrapper_and_explicit_replacement(self):
        self.install()
        destination = self.repo / "custom executable directory"
        self.succeeds(self.repair(None, "--dest", destination, "--local-bin", self.binary))
        legacy = destination / "agentic-lens"
        result = self.succeeds(self.run_script(legacy, "legacy", "argument with spaces"))
        self.assertEqual(json.loads(result.stdout), {"cwd": str(self.repo), "args": ["legacy", "argument with spaces"]})
        launcher = self.succeeds(self.run_script(self.repo / LAUNCH, "custom", AGENTIC_LENS_BIN_DIR=str(destination)))
        self.assertEqual(json.loads(launcher.stdout)["args"], ["custom"])
        alternate = self.root / "alternate prebuilt"
        alternate.write_text("#!/bin/sh\necho replacement\n")
        alternate.chmod(0o755)
        self.succeeds(self.repair(None, "--dest", destination, "--local-bin", alternate))
        self.assertEqual((destination / self.native_name).read_bytes(), self.binary.read_bytes())
        self.succeeds(self.repair(None, "--dest", destination, "--local-bin", alternate, "--force"))
        self.assertEqual((destination / self.native_name).read_bytes(), alternate.read_bytes())
        self.assertFalse(self.native().exists())
        self.assertFalse(list(self.repo.rglob(".gitignore")))
        self.assertEqual(self.requests(), [])

    def test_recommended_repair_restores_native_execute_permission_without_replacement(self):
        self.install()
        self.succeeds(self.repair(None, "--local-bin", self.binary))
        before = self.native().read_bytes()
        for local_override in [False, True]:
            with self.subTest(local_override=local_override):
                self.native().chmod(0o644)
                denied = self.run_script(self.repo / LAUNCH, "permission-check")
                self.assertNotEqual(denied.returncode, 0)
                self.assertIn("not executable", denied.stderr)
                command = next(line.strip() for line in denied.stderr.splitlines()
                               if line.startswith("  bash "))
                args = shlex.split(command)
                if local_override:
                    args.extend(["--local-bin", str(self.binary)])
                repaired = subprocess.run(args, cwd=self.repo, env=self.env,
                                          capture_output=True, text=True, timeout=20)
                self.succeeds(repaired)
                self.assertTrue(os.access(self.native(), os.X_OK), repaired.stdout + repaired.stderr)
                self.assertEqual(self.native().read_bytes(), before)
                launched = self.succeeds(self.run_script(self.repo / LAUNCH, "permission-check"))
                self.assertEqual(json.loads(launched.stdout)["args"], ["permission-check"])
                self.assertEqual(self.requests(), [])

    def test_worktree_without_release_marker_requires_explicit_version_before_fetch(self):
        policy = str(SETUP / "RELEASE_VERSION") + "\n" + str(OPERATOR / "bin") + "/\n"
        (self.repo / ".gitignore").write_text(policy)
        self.install()
        self.git("add", ".agents", ".gitignore")
        self.git("commit", "-qm", "owner omits generated release marker and binaries")
        worktree = self.root / "worktree without release marker"
        self.git("worktree", "add", "-qb", "without-release-marker", str(worktree))
        self.assertTrue((worktree / DOWNLOAD).is_file())
        self.assertFalse((worktree / SETUP / "RELEASE_VERSION").exists())
        for args in [[], ["--dry-run"]]:
            with self.subTest(args=args):
                unpinned = self.repair(worktree, *args, FETCH_MODE="online")
                self.assertNotEqual(unpinned.returncode, 0, unpinned.stdout)
                self.assertIn("--version", unpinned.stderr)
                self.assertEqual(self.requests(), [])
                self.assertFalse(self.native(worktree).parent.exists())
        offline_destination = self.root / "unversioned offline prebuilt"
        self.succeeds(self.repair(worktree, "--dest", offline_destination,
                                 "--local-bin", self.binary))
        self.assertTrue((offline_destination / self.native_name).is_file())
        self.assertEqual(self.requests(), [])
        self.succeeds(self.repair(worktree, "--version", "v9.8.7", FETCH_MODE="online"))
        self.assertEqual([request["url"] for request in self.requests()], [
            "https://github.com/NiltonDuarte/AgenticLens/releases/download/v9.8.7/" + self.native_name
        ])
        self.assertEqual(self.git("status", "--porcelain", cwd=worktree), "")
        self.assertFalse((worktree / SETUP / "RELEASE_VERSION").exists())
        self.assertEqual((worktree / ".agenticlens-config.yaml").read_bytes(), self.config.read_bytes())

    def test_contributor_entrypoint_preserves_explicit_latest_default_without_marker(self):
        destination = self.repo / "contributor binary directory"
        source = self.root / "contributor source without installed marker"
        for relative in [DOWNLOAD, SETUP / "scripts/agentic-lens-wrapper.sh",
                         Path("scripts/download-agentic-lens.sh")]:
            target = source / relative
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(SOURCE / relative, target)
        entrypoint = source / "scripts/download-agentic-lens.sh"
        dry_run = self.succeeds(self.run_script(entrypoint, "--dest", destination, "--dry-run"))
        self.assertIn("Version: latest", dry_run.stdout)
        self.assertEqual(self.requests(), [])
        self.succeeds(self.run_script(entrypoint, "--dest", destination, FETCH_MODE="online"))
        self.assertEqual([request["url"] for request in self.requests()], [
            "https://github.com/NiltonDuarte/AgenticLens/releases/latest/download/" + self.native_name
        ])
        self.assertTrue((destination / self.native_name).is_file())

    def test_dry_run_platform_artifacts_and_missing_option_values_do_not_write(self):
        self.install()
        result = self.succeeds(self.repair(None, "--all-platforms", "--dry-run"))
        for artifact in ["linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64", "windows-amd64.exe"]:
            self.assertIn("/releases/download/v9.8.7/agentic-lens-" + artifact, result.stdout)
        self.assertFalse((self.repo / OPERATOR / "bin").exists())
        for option in ["--version", "--dest", "--local-bin", "--include-host"]:
            with self.subTest(option=option):
                result = self.repair(None, option)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("requires a value", result.stderr)
        self.assertEqual(self.requests(), [])


if __name__ == "__main__":
    unittest.main()
