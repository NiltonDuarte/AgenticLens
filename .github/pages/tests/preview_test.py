"""Exercise static main publication with self-contained published-site fixtures."""

import importlib.util
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


MODULE_PATH = Path(__file__).resolve().parents[1] / "preview.py"
spec = importlib.util.spec_from_file_location("portal_preview", MODULE_PATH)
preview = importlib.util.module_from_spec(spec)
spec.loader.exec_module(preview)


class PublicationCompositionTest(unittest.TestCase):
    repository = "fixture-owner/FixtureRepo"
    main_revision = "a" * 40
    revision_24 = "b" * 40
    revision_25 = "c" * 40

    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix="portal publication ")
        self.root = Path(self.temporary.name)
        self.production = self.site("production", "production")
        self.output = self.root / "published"

    def tearDown(self):
        self.temporary.cleanup()

    def site(self, directory, label):
        site = self.root / directory
        (site / "presentation").mkdir(parents=True)
        (site / "roadmap" / "to-do").mkdir(parents=True)
        (site / "index.html").write_text(f"<html><body>{label}</body></html>")
        for name in ("style.css", "investigation.svg", "analytics.svg", "history.svg", "architecture.svg"):
            (site / "presentation" / name).write_text(f"{label} {name}")
        (site / "roadmap" / "index.html").write_text(f"<html>{label} roadmap</html>")
        (site / "roadmap" / "style.css").write_text(f"{label} roadmap stylesheet")
        (site / "roadmap" / "to-do" / "TASK-Z.html").write_text(f"<html>{label} ticket</html>")
        (site / ".tickets-site-output").write_text("Generated ticket website\n")
        return site

    def compose(self, *, production=None, previous=None, output=None):
        return preview.compose(
            production or self.production, previous, output or self.output,
            repository=self.repository, main_revision=self.main_revision,
        )

    @staticmethod
    def snapshot(directory):
        return {path.relative_to(directory).as_posix(): path.read_bytes()
                for path in directory.rglob("*") if path.is_file()}

    def seed(self):
        self.compose()
        return self.output

    def legacy_publication_with_previews(self):
        """A retained publication from before previews were retired."""
        legacy = self.root / "legacy"
        shutil.copytree(self.production, legacy)
        for number, revision in (("24", self.revision_24), ("25", self.revision_25)):
            self.site(f"legacy-preview-{number}", f"preview {number}")
            shutil.copytree(self.root / f"legacy-preview-{number}", legacy / "preview" / number)
        (legacy / ".portal-state.json").write_text(json.dumps({
            "version": 1, "repository": self.repository, "main_revision": "d" * 40,
            "previews": {"24": {"revision": self.revision_24}, "25": {"revision": self.revision_25}},
        }))
        return legacy

    def test_initial_publication_contains_only_production(self):
        before = self.snapshot(self.production)
        self.seed()
        published = self.snapshot(self.output)
        self.assertEqual(set(before) | {".portal-state.json"}, set(published))
        for name, contents in before.items():
            self.assertEqual(contents, published[name], name)
        state = json.loads((self.output / ".portal-state.json").read_text())
        self.assertEqual({"version": 1, "repository": self.repository,
                          "main_revision": self.main_revision, "previews": {}}, state)
        self.assertEqual(before, self.snapshot(self.production), "Composition must not edit its inputs")

    def test_main_publication_drops_retained_previews_and_keeps_production(self):
        legacy = self.legacy_publication_with_previews()
        legacy_before = self.snapshot(legacy)
        (self.production / "index.html").write_text("<html>new main</html>")
        output = self.root / "after-retirement"
        self.compose(previous=legacy, output=output)
        self.assertFalse((output / "preview").exists())
        state = json.loads((output / ".portal-state.json").read_text())
        self.assertEqual({}, state["previews"])
        self.assertEqual(self.main_revision, state["main_revision"])
        self.assertEqual("<html>new main</html>", (output / "index.html").read_text())
        expected = self.snapshot(self.production)
        published = self.snapshot(output)
        published.pop(".portal-state.json")
        self.assertEqual(expected, published)
        self.assertEqual(legacy_before, self.snapshot(legacy), "Composition must not edit retained state")
        # The composed state is itself a valid previous publication.
        self.compose(previous=output, output=self.root / "next")

    def test_previous_metadata_cannot_escape_preview_directory_or_change_repository(self):
        legacy = self.legacy_publication_with_previews()
        state_path = legacy / ".portal-state.json"
        original = json.loads(state_path.read_text())
        for changed in (
            {**original, "repository": "other/repository"},
            {**original, "previews": {"../production": {"revision": self.revision_24}}},
            {**original, "previews": {"24": {"revision": "not-a-revision"}}},
            {**original, "previews": {"24": {"revision": self.revision_24}}},
        ):
            state_path.write_text(json.dumps(changed))
            before = self.snapshot(legacy)
            next_output = self.root / "rejected"
            with self.subTest(state=changed):
                with self.assertRaises((ValueError, RuntimeError)):
                    self.compose(previous=legacy, output=next_output)
                self.assertEqual(before, self.snapshot(legacy))
                self.assertFalse(next_output.exists())

    def test_unsafe_previous_preview_is_rejected_before_replacement(self):
        legacy = self.legacy_publication_with_previews()
        self.seed()
        before = self.snapshot(self.output)
        outside = self.root / "private-ledger.json"
        outside.write_text("PRIVATE_CAPTURE_MUST_NOT_BE_PUBLISHED")
        (legacy / "preview" / "24" / "roadmap" / "private.html").symlink_to(outside)
        with self.assertRaises((ValueError, RuntimeError)):
            self.compose(previous=legacy)
        self.assertEqual(before, self.snapshot(self.output))
        self.assertEqual("PRIVATE_CAPTURE_MUST_NOT_BE_PUBLISHED", outside.read_text())

    def test_unexpected_production_payload_is_rejected_before_replacement(self):
        self.seed()
        before = self.snapshot(self.output)
        (self.production / "agentic_metrics.json").write_text("PRIVATE_CAPTURE_MUST_NOT_BE_PUBLISHED")
        next_output = self.root / "rejected"
        with self.assertRaises((ValueError, RuntimeError)):
            self.compose(previous=self.output, output=next_output)
        self.assertEqual(before, self.snapshot(self.output))
        self.assertFalse(next_output.exists())

    def test_resolve_refuses_a_ref_other_than_main(self):
        output = self.root / "github-output"
        environment = {**os.environ, "GITHUB_REF": "refs/heads/feature",
                       "GITHUB_REPOSITORY": self.repository, "GITHUB_OUTPUT": str(output),
                       "GH_TOKEN": "fixture-token-not-a-secret"}
        result = subprocess.run(["python3", str(MODULE_PATH), "resolve"], env=environment,
                                capture_output=True, text=True)
        self.assertNotEqual(0, result.returncode)
        self.assertIn("Pages publishes only from main", result.stderr)
        self.assertFalse(output.exists())

    def test_output_cannot_replace_inputs_unmanaged_files_or_symlink_targets(self):
        unowned = self.root / "unmanaged"
        unowned.mkdir()
        (unowned / "keep.txt").write_text("user-owned contents")
        linked = self.root / "linked-output"
        linked.symlink_to(unowned, target_is_directory=True)
        before = self.snapshot(self.production)
        for output in (self.production, self.production / "nested", unowned, linked):
            with self.subTest(output=str(output)):
                with self.assertRaises((ValueError, RuntimeError)):
                    self.compose(output=output)
                self.assertEqual(before, self.snapshot(self.production))
                self.assertEqual("user-owned contents", (unowned / "keep.txt").read_text())
                self.assertTrue(linked.is_symlink())

    def test_state_load_accepts_clean_git_history_and_blocks_pending_publication(self):
        self.seed()
        remote = self.root / "publication.git"
        working = self.root / "publication source"
        working.mkdir()
        subprocess.run(["git", "init", "--bare", str(remote)], check=True, capture_output=True)

        def git(*arguments):
            return subprocess.run(["git", *arguments], cwd=working,
                                  check=True, capture_output=True, text=True)

        git("init", "--initial-branch=portal-pages-state")
        git("config", "user.name", "Fixture Author")
        git("config", "user.email", "fixture@example.invalid")
        git("remote", "add", "origin", remote.as_uri())
        shutil.copytree(self.output, working / "site")
        git("add", "site")
        git("commit", "-m", "chore: fixture publication")
        git("push", "origin", "HEAD:refs/heads/portal-pages-state")
        configuration = self.root / "gitconfig"
        subprocess.run(["git", "config", "--file", str(configuration),
                        f"url.{remote.as_uri()}.insteadOf",
                        f"https://github.com/{self.repository}.git"],
                       check=True, capture_output=True)
        environment = {**os.environ, "GIT_CONFIG_GLOBAL": str(configuration),
                       "GIT_CONFIG_NOSYSTEM": "1", "GITHUB_REPOSITORY": self.repository,
                       "GH_TOKEN": "fixture-token-not-a-secret"}

        def load(directory):
            return subprocess.run(["python3", str(MODULE_PATH), "load-state",
                                   "--directory", str(directory)], env=environment,
                                  capture_output=True, text=True)

        clean = self.root / "clean loaded state"
        result = load(clean)
        self.assertEqual(0, result.returncode, result.stdout + result.stderr)
        self.assertEqual(self.snapshot(self.output), self.snapshot(clean / "site"))
        marker = working / ".pending-publication.json"
        marker.write_text(json.dumps({"repository": self.repository, "run_id": "123"}))
        git("add", marker.name)
        git("commit", "-m", "chore: fixture interrupted publication")
        git("push", "origin", "HEAD:refs/heads/portal-pages-state")
        interrupted = self.root / "interrupted loaded state"
        result = load(interrupted)
        self.assertNotEqual(0, result.returncode)
        self.assertIn("Unfinished Pages publication", result.stderr)
        self.assertEqual(marker.read_bytes(), (interrupted / marker.name).read_bytes())
        self.assertEqual(self.snapshot(self.output), self.snapshot(interrupted / "site"))


if __name__ == "__main__":
    unittest.main()
