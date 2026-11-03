"""Execute release workflow shell against disposable tags; no remote publishing."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest


WORKFLOW = Path(__file__).resolve().parents[2] / ".github/workflows/release-please.yml"
RC_WORKFLOW = WORKFLOW.with_name("release-candidate.yml")
RC_RESOLVER = WORKFLOW.parents[2] / "scripts/release/rc-version.cjs"


def steps(workflow=WORKFLOW):
    # Read this workflow's named steps and literal run blocks without a YAML
    # dependency. Fail if the relevant block format changes instead of skipping.
    result = {}
    current = None
    for line in workflow.read_text().splitlines():
        if line.startswith("      - name: "):
            current = line.removeprefix("      - name: ")
            result[current] = []
        elif current:
            result[current].append(line)
    return result


def shell(name):
    lines = steps()[name]
    start = lines.index("        run: |") + 1
    block = []
    for line in lines[start:]:
        if line and not line.startswith("          "):
            break
        block.append(line[10:])
    if not block:
        raise AssertionError(f"Missing shell block: {name}")
    return "\n".join(block)


class ReleaseRevisionTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.repo = Path(self.temp.name)
        self.git("init", "-q")
        self.git("config", "user.name", "Fixture")
        self.git("config", "user.email", "fixture@example.invalid")
        self.git("commit", "--allow-empty", "-qm", "older")
        self.older = self.git("rev-parse", "HEAD")
        self.git("tag", "-a", "v1.0.0", "-m", "annotated older tag")
        self.git("commit", "--allow-empty", "-qm", "newer")
        self.newer = self.git("rev-parse", "HEAD")
        self.git("tag", "v2.0.0")
        self.env = dict(os.environ, GITHUB_ENV=str(self.repo / "github-env"))

    def git(self, *args):
        return subprocess.check_output(["git", *args], cwd=self.repo, text=True, stderr=subprocess.PIPE).strip()

    def run_step(self, name, **env):
        return subprocess.run(["bash", "-c", shell(name)], cwd=self.repo,
                              env=dict(self.env, **env), capture_output=True, text=True)

    def select(self, force, action):
        result = self.run_step("Select release revision", FORCE_TAG=force, ACTION_TAG=action)
        self.assertEqual(result.returncode, 0, result.stderr)
        return dict(line.split("=", 1) for line in (self.repo / "github-env").read_text().splitlines())

    def test_forced_annotated_tag_wins_over_new_release_output(self):
        env = self.select("v1.0.0", "v2.0.0")
        self.assertEqual(self.git("rev-parse", "HEAD"), self.older)
        self.assertEqual(env["RELEASE_COMMIT"], self.older)
        self.assertEqual(env["RELEASE_TAG"], "v1.0.0")
        result = self.run_step("Compute Release Metadata", **env)
        self.assertEqual(result.returncode, 0, result.stderr)
        metadata = dict(line.split("=", 1) for line in (self.repo / "github-env").read_text().splitlines())
        self.assertEqual(metadata["VERSION"], "v1.0.0")
        self.assertEqual(metadata["GIT_COMMIT"], self.git("rev-parse", "--short", self.older))
        self.git("checkout", "--detach", self.newer)
        self.assertNotEqual(self.run_step("Compute Release Metadata", **env).returncode, 0)

    def test_retargeted_tag_is_rejected_before_metadata(self):
        env = self.select("v1.0.0", "")
        self.git("tag", "-f", "v1.0.0", self.newer)
        self.assertNotEqual(self.run_step("Compute Release Metadata", **env).returncode, 0)

    def test_missing_forced_tag_does_not_fall_back_to_action_tag(self):
        result = self.run_step("Select release revision", FORCE_TAG="missing", ACTION_TAG="v2.0.0")
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse((self.repo / "github-env").exists())
        self.assertEqual(self.git("rev-parse", "HEAD"), self.newer)

    def test_normal_release_uses_action_tag(self):
        self.git("checkout", "--detach", self.older)
        self.select("", "v2.0.0")
        self.assertEqual(self.git("rev-parse", "HEAD"), self.newer)

    def test_invalid_selection_fails_without_metadata_or_checkout(self):
        for value in ["", "missing", "HEAD", self.older, "bad\nname", "$(touch INJECTED)"]:
            with self.subTest(value=value):
                result = self.run_step("Select release revision", FORCE_TAG=value, ACTION_TAG="")
                self.assertNotEqual(result.returncode, 0)
                self.assertFalse((self.repo / "github-env").exists())
                self.assertFalse((self.repo / "INJECTED").exists())
                self.assertEqual(self.git("rev-parse", "HEAD"), self.newer)

    def test_workflow_order_and_forced_side_effect_guards(self):
        workflow = steps()
        names = list(workflow)
        select = names.index("Select release revision")
        for name in ["Set up Go 1.23", "Set up Node.js 20", "Install web dependencies", "Static Analysis", "Run Test Suite", "Compute Release Metadata", "Cross-Compile Release Binaries", "Upload Release Artifacts"]:
            self.assertGreater(names.index(name), select)
        for name in ["Run Release Please Action", "Update Floating Latest Tag", "Sweep All Stale Release Candidates"]:
            condition = next(line for line in workflow[name] if line.startswith("        if:"))
            self.assertIn("inputs.force_tag == ''", condition)
        self.assertLess(names.index("Verify forced release exists"), names.index("Set up Go 1.23"))
        self.assertIn("        if: ${{ inputs.force_tag != '' }}", workflow["Verify forced release exists"])
        self.assertIn('        run: gh release view "$RELEASE_TAG" > /dev/null', workflow["Verify forced release exists"])
        self.assertIn("          tag_name: ${{ env.RELEASE_TAG }}", workflow["Upload Release Artifacts"])

    def test_candidate_build_and_publication_require_resolver_and_final_eligibility(self):
        workflow = steps(RC_WORKFLOW)
        names = list(workflow)
        compute = names.index("Compute Release Candidate Metadata")
        recheck = names.index("Recheck Release Candidate Eligibility")
        self.assertLess(names.index("Install release version tooling"), compute)
        self.assertIn("        run: node scripts/release/rc-version.cjs resolve", workflow["Compute Release Candidate Metadata"])
        self.assertIn("        run: node scripts/release/rc-version.cjs verify", workflow["Recheck Release Candidate Eligibility"])
        for name in ["Install web dependencies", "Static Analysis", "Run Test Suite", "Cross-Compile Release Binaries", "Package Release Archives", "Verify Release Artifacts", "Recheck Release Candidate Eligibility"]:
            self.assertLess(compute, names.index(name))
            self.assertIn("        if: steps.candidate.outputs.should_publish == 'true'", workflow[name])
        self.assertLess(names.index("Verify Release Artifacts"), recheck)
        for name in ["Publish GitHub Pre-release", "Prune Older Release Candidates"]:
            self.assertLess(recheck, names.index(name))
            self.assertIn("        if: steps.eligibility.outputs.should_publish == 'true'", workflow[name])
        self.assertIn('        run: make release-build VERSION="${VERSION}" GIT_COMMIT="${GIT_COMMIT}" BUILD_DATE="${BUILD_DATE}"', workflow["Cross-Compile Release Binaries"])
        self.assertIn("          tag_name: ${{ env.VERSION }}", workflow["Publish GitHub Pre-release"])

    def test_candidate_cli_failure_leaves_no_publish_permission_or_metadata(self):
        for expected, token, error in [
            (self.older, "fixture-token", "Checkout does not match"),
            (self.newer, "", "GH_TOKEN is required"),
        ]:
            with self.subTest(error=error):
                output = self.repo / "github-output"
                result = subprocess.run(
                    ["node", str(RC_RESOLVER), "resolve"], cwd=self.repo,
                    env=dict(self.env, EXPECTED_SHA=expected, GH_TOKEN=token,
                             GITHUB_REPOSITORY="fixture/repository", GITHUB_OUTPUT=str(output)),
                    capture_output=True, text=True,
                )
                self.assertNotEqual(result.returncode, 0)
                self.assertIn(error, result.stderr)
                self.assertFalse(output.exists())
                self.assertFalse((self.repo / "github-env").exists())


if __name__ == "__main__":
    unittest.main()
