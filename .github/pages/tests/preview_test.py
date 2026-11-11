"""Exercise static preview composition with self-contained published-site fixtures."""

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


class PreviewCompositionTest(unittest.TestCase):
    repository = "fixture-owner/FixtureRepo"
    main_revision = "a" * 40
    revision_24 = "b" * 40
    revision_25 = "c" * 40

    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix="portal preview ")
        self.root = Path(self.temporary.name)
        self.production = self.site("production", "production")
        self.candidate = self.site("candidate", "candidate 24")
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

    def pull_request(self, number, revision, state="open", repository=None, base="main"):
        return {"number": number, "state": state, "base": {"ref": base}, "head": {
            "sha": revision,
            "repo": {"full_name": repository or self.repository},
        }}

    def compose(self, *, production=None, previous=None, candidate=None,
                output=None, pulls=None, number=None, revision=None):
        return preview.compose(
            production or self.production, previous, candidate,
            output or self.output, repository=self.repository,
            main_revision=self.main_revision,
            pull_requests=[] if pulls is None else pulls,
            preview_number=number, preview_revision=revision,
        )

    @staticmethod
    def snapshot(directory):
        return {path.relative_to(directory).as_posix(): path.read_bytes()
                for path in directory.rglob("*") if path.is_file()}

    def seed(self):
        self.compose(candidate=self.candidate,
                     pulls=[self.pull_request(24, self.revision_24)],
                     number=24, revision=self.revision_24)
        return self.output

    def test_initial_preview_retains_production_and_separate_preview_assets(self):
        before = self.snapshot(self.production)
        self.seed()
        published = self.snapshot(self.output)
        for name, contents in before.items():
            self.assertEqual(contents, published[name], name)
        candidate_files = self.snapshot(self.candidate)
        for name, contents in candidate_files.items():
            self.assertEqual(contents, published[f"preview/24/{name}"], name)
        state = json.loads((self.output / ".portal-state.json").read_text())
        self.assertEqual(self.repository, state["repository"])
        self.assertEqual(self.main_revision, state["main_revision"])
        self.assertEqual(self.revision_24, state["previews"]["24"]["revision"])
        self.assertEqual(before, self.snapshot(self.production), "Composition must not edit its inputs")

    def test_update_preserves_other_preview_and_replaces_old_matching_files(self):
        first = self.seed()
        candidate25 = self.site("candidate25", "candidate 25")
        second = self.root / "second"
        pulls = [self.pull_request(24, self.revision_24),
                 self.pull_request(25, self.revision_25)]
        self.compose(previous=first, candidate=candidate25, output=second,
                     pulls=pulls, number=25, revision=self.revision_25)
        preserved25 = self.snapshot(second / "preview" / "25")
        (self.candidate / "roadmap" / "to-do" / "TASK-Z.html").unlink()
        (self.candidate / "index.html").write_text("<html>updated candidate 24</html>")
        next_revision = "d" * 40
        pulls[0] = self.pull_request(24, next_revision)
        third = self.root / "third"
        self.compose(previous=second, candidate=self.candidate, output=third,
                     pulls=pulls, number=24, revision=next_revision)
        self.assertEqual(preserved25, self.snapshot(third / "preview" / "25"))
        self.assertIn("updated candidate 24", (third / "preview" / "24" / "index.html").read_text())
        self.assertFalse((third / "preview" / "24" / "roadmap" / "to-do" / "TASK-Z.html").exists())
        self.assertEqual(self.snapshot(self.production)["index.html"], (third / "index.html").read_bytes())

    def test_open_changed_head_retains_last_success_until_close_prunes(self):
        first = self.seed()
        second = self.root / "retained"
        self.compose(previous=first, output=second,
                     pulls=[self.pull_request(24, "d" * 40)])
        self.assertEqual(self.snapshot(first / "preview" / "24"),
                         self.snapshot(second / "preview" / "24"))
        third = self.root / "closed"
        self.compose(previous=second, output=third,
                     pulls=[self.pull_request(24, "d" * 40, state="closed")])
        self.assertFalse((third / "preview" / "24").exists())
        state = json.loads((third / ".portal-state.json").read_text())
        self.assertEqual({}, state["previews"])
        self.assertEqual((self.production / "index.html").read_bytes(), (third / "index.html").read_bytes())

    def test_rejected_candidate_preserves_previous_published_tree(self):
        self.seed()
        before = self.snapshot(self.output)
        cases = [
            ([self.pull_request(24, "d" * 40)], 24, self.revision_24),
            ([self.pull_request(24, self.revision_24, state="closed")], 24, self.revision_24),
            ([self.pull_request(24, self.revision_24, repository="external/fork")], 24, self.revision_24),
            ([self.pull_request(25, self.revision_24)], 24, self.revision_24),
            ([self.pull_request(24, self.revision_24)], "../24", self.revision_24),
        ]
        for pulls, number, revision in cases:
            with self.subTest(number=number, pulls=pulls):
                with self.assertRaises((ValueError, RuntimeError, TypeError)):
                    self.compose(candidate=self.candidate, pulls=pulls,
                                 number=number, revision=revision)
                self.assertEqual(before, self.snapshot(self.output))

    def test_retargeted_open_pr_is_removed_and_cannot_replace_existing_publication(self):
        self.seed()
        before = self.snapshot(self.output)
        retargeted = [self.pull_request(24, self.revision_24, base="release")]
        with self.assertRaises((ValueError, RuntimeError)):
            self.compose(candidate=self.candidate, pulls=retargeted,
                         number=24, revision=self.revision_24)
        self.assertEqual(before, self.snapshot(self.output))
        next_output = self.root / "retargeted"
        self.compose(previous=self.output, output=next_output, pulls=retargeted)
        self.assertFalse((next_output / "preview" / "24").exists())
        state = json.loads((next_output / ".portal-state.json").read_text())
        self.assertEqual({}, state["previews"])
        self.assertEqual((self.production / "index.html").read_bytes(), (next_output / "index.html").read_bytes())

    def test_candidate_symlink_or_unexpected_payload_is_rejected_before_replacement(self):
        self.seed()
        before = self.snapshot(self.output)
        outside = self.root / "private-ledger.json"
        outside.write_text("PRIVATE_CAPTURE_MUST_NOT_BE_PUBLISHED")
        unsafe = self.candidate / "roadmap" / "private.html"
        unsafe.symlink_to(outside)
        for target in (unsafe, self.candidate / "agentic_metrics.json"):
            if target != unsafe:
                unsafe.unlink()
                target.write_text(outside.read_text())
            with self.subTest(target=target.name):
                with self.assertRaises((ValueError, RuntimeError)):
                    self.compose(candidate=self.candidate,
                                 pulls=[self.pull_request(24, self.revision_24)],
                                 number=24, revision=self.revision_24)
                self.assertEqual(before, self.snapshot(self.output))
        self.assertEqual("PRIVATE_CAPTURE_MUST_NOT_BE_PUBLISHED", outside.read_text())

    def test_previous_metadata_cannot_escape_preview_directory_or_change_repository(self):
        self.seed()
        state_path = self.output / ".portal-state.json"
        original = json.loads(state_path.read_text())
        for changed in (
            {**original, "repository": "other/repository"},
            {**original, "previews": {"../production": {"revision": self.revision_24}}},
            {**original, "previews": {"24": {"revision": "not-a-revision"}}},
        ):
            state_path.write_text(json.dumps(changed))
            before = self.snapshot(self.output)
            next_output = self.root / "rejected"
            with self.subTest(state=changed):
                with self.assertRaises((ValueError, RuntimeError)):
                    self.compose(previous=self.output, output=next_output,
                                 pulls=[self.pull_request(24, self.revision_24)])
                self.assertEqual(before, self.snapshot(self.output))
                self.assertFalse(next_output.exists())

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
