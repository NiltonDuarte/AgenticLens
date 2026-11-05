"""Execute release workflow shell against disposable tags; no remote publishing."""
import base64
import copy
from contextlib import contextmanager
import http.server
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import threading
import unittest


WORKFLOW = Path(__file__).resolve().parents[2] / ".github/workflows/release-please.yml"
RC_WORKFLOW = WORKFLOW.with_name("release-candidate.yml")


@contextmanager
def fixture_api():
    response = {"body": {}, "status": 200, "drop": False, "paths": []}

    class Handler(http.server.BaseHTTPRequestHandler):
        def do_GET(self):
            response["paths"].append(self.path)
            if response["drop"]:
                self.close_connection = True
                return
            self.send_response(response["status"])
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps(response["body"]).encode())

        def log_message(self, *args):
            pass

    server = http.server.HTTPServer(("127.0.0.1", 0), Handler)
    thread = threading.Thread(target=server.serve_forever, kwargs={"poll_interval": 0.01}, daemon=True)
    thread.start()
    try:
        yield f"http://127.0.0.1:{server.server_port}", response
    finally:
        server.shutdown()
        server.server_close()
        thread.join()


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


def shell(name, workflow=WORKFLOW):
    lines = steps(workflow)[name]
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
        for name in ["Set up Go", "Set up Node.js 20", "Install web dependencies", "Static Analysis", "Run Test Suite", "Compute Release Metadata", "Cross-Compile Release Binaries", "Upload Release Artifacts"]:
            self.assertGreater(names.index(name), select)
        for name in ["Run Release Please Action", "Update Floating Latest Tag", "Sweep All Stale Release Candidates"]:
            condition = next(line for line in workflow[name] if line.startswith("        if:"))
            self.assertIn("inputs.force_tag == ''", condition)
        self.assertLess(names.index("Verify forced release exists"), names.index("Set up Go"))
        self.assertIn("          go-version-file: backend/go.mod", workflow["Set up Go"])
        self.assertIn("        if: ${{ inputs.force_tag != '' }}", workflow["Verify forced release exists"])
        self.assertIn('        run: gh release view "$RELEASE_TAG" > /dev/null', workflow["Verify forced release exists"])
        self.assertIn("          tag_name: ${{ env.RELEASE_TAG }}", workflow["Upload Release Artifacts"])


class ReleaseCandidateTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.repo = Path(self.temp.name)
        self.git("init", "-q")
        self.git("config", "user.name", "Fixture")
        self.git("config", "user.email", "fixture@example.invalid")
        self.git("commit", "--allow-empty", "-qm", "chore: stable baseline")
        self.git("tag", "-a", "v2.0.0", "-m", "stable release")
        self.git("commit", "--allow-empty", "-qm", "feat: next release")
        self.candidate = self.git("rev-parse", "HEAD")
        self.git("commit", "--allow-empty", "-qm", "chore(main): release 2.1.0")
        self.pr_head = self.git("rev-parse", "HEAD")
        self.git("checkout", "--detach", self.candidate)
        self.branch = "release-please--branches--main--components--agentic-lens"
        self.api = "repos/fixture/repository/"
        self.responses = {
            "pulls": {"value": self.pr_head, "required": [
                "--paginate", "state=open", "base=main", f"head=fixture:{self.branch}",
                ".head.repo.full_name == env.GITHUB_REPOSITORY", "autorelease: pending",
            ]},
            "contents/.release-please-manifest.json": {
                "value": "2.1.0", "required": [f"ref={self.pr_head}", "--jq"],
            },
        }
        fixture_bin = self.repo / "fixture-bin"
        fixture_bin.mkdir()
        gh = fixture_bin / "gh"
        gh.write_text("""#!/usr/bin/env python3
import json, os, sys
from pathlib import Path
args = sys.argv[1:]
with open(os.environ['GH_CALLS'], 'a') as log:
    log.write(json.dumps(args) + '\\n')
endpoint = next((arg for arg in args if arg.startswith('repos/fixture/repository/')), '')
routes = json.loads(Path(os.environ['GH_FIXTURE']).read_text())
response = routes.get(endpoint.removeprefix('repos/fixture/repository/'))
if args[0] != 'api' or response is None:
    sys.exit('Unexpected fixture gh request: ' + repr(args))
if '--method' in args and args[args.index('--method') + 1] != 'GET':
    sys.exit('Fixture permits read-only API calls')
for required in response.get('required', []):
    if not any(required in arg for arg in args):
        sys.exit('Missing required API argument: ' + required)
if response.get('error'):
    sys.exit('fixture GitHub API unavailable')
print(response['value'])
""")
        gh.chmod(0o755)
        self.env = dict(os.environ, PATH=f"{fixture_bin}:{os.environ['PATH']}",
                        GH_TOKEN="fixture-token", GITHUB_REPOSITORY="fixture/repository",
                        GITHUB_ENV=str(self.repo / "github-env"),
                        GITHUB_OUTPUT=str(self.repo / "github-output"),
                        GH_FIXTURE=str(self.repo / "responses.json"),
                        GH_CALLS=str(self.repo / "gh-calls"),
                        GITHUB_RUN_NUMBER="114", EXPECTED_SHA=self.candidate)

    def git(self, *args):
        return subprocess.check_output(["git", *args], cwd=self.repo, text=True, stderr=subprocess.PIPE).strip()

    def read_file(self, name):
        path = self.repo / name
        return path.read_text() if path.exists() else ""

    def run_rc(self, name="Compute Release Candidate Metadata", **env):
        (self.repo / "responses.json").write_text(json.dumps(self.responses))
        (self.repo / "github-output").unlink(missing_ok=True)
        return subprocess.run(["bash", "-c", shell(name, RC_WORKFLOW)], cwd=self.repo,
                              env=dict(self.env, **env), capture_output=True, text=True)

    def resolve(self):
        result = self.run_rc()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.read_file("github-output"), "should_publish=true\n")
        return dict(line.split("=", 1) for line in self.read_file("github-env").splitlines())

    def assert_blocked(self, result, *, skip=False):
        self.assertNotIn("should_publish=true", self.read_file("github-output"))
        if skip:
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("should_publish=false", self.read_file("github-output"))
        else:
            self.assertNotEqual(result.returncode, 0)

    def test_manifest_selection_writes_real_metadata_and_uses_immutable_pr_head(self):
        metadata = self.resolve()
        short_sha = self.git("rev-parse", "--short", self.candidate)
        self.assertEqual(metadata["VERSION"], f"v2.1.0-rc.114+{short_sha}")
        self.assertEqual(metadata["GIT_COMMIT"], short_sha)
        self.assertEqual(metadata["RC_TARGET"], "2.1.0")
        self.assertRegex(metadata["BUILD_DATE"], r"^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ$")
        calls = [json.loads(line) for line in self.read_file("gh-calls").splitlines()]
        content_call = next(args for args in calls if self.api + "contents/.release-please-manifest.json" in args)
        self.assertIn(f"ref={self.pr_head}", content_call)
        for target in ["2.0.1", "3.0.0", "4.0.0", "2.10.0"]:
            with self.subTest(manifest_target=target):
                self.responses["contents/.release-please-manifest.json"]["value"] = target
                selected = self.resolve()
                self.assertEqual(selected["VERSION"], f"v{target}-rc.114+{short_sha}")

    def test_no_release_pr_skips_without_metadata(self):
        self.responses["pulls"]["value"] = ""
        self.assert_blocked(self.run_rc(), skip=True)
        self.assertEqual(self.read_file("github-env"), "")

    def test_real_github_cli_filters_select_expected_raw_api_records(self):
        # Capture the production filters from executed shell; the canned shell
        # fixture cannot establish gojq semantics, so evaluate them with real gh.
        gh = shutil.which("gh")
        self.assertIsNotNone(gh, "GitHub CLI is required for release filter tests")
        self.resolve()
        calls = [json.loads(line) for line in self.read_file("gh-calls").splitlines()]
        filters = {}
        for args in calls:
            route = next(arg.removeprefix(self.api) for arg in args if arg.startswith(self.api))
            filters.setdefault(route, set()).add(args[args.index("--jq") + 1])

        def evaluate(route, body):
            response["body"] = body
            return [subprocess.run(
                [gh, "api", api_url + "/fixture", "--jq", expression],
                env=dict(os.environ, GH_TOKEN="fixture-token", GITHUB_REPOSITORY="fixture/repository"),
                capture_output=True, text=True, timeout=10,
            ) for expression in sorted(filters[route])]

        def expect(route, body, output):
            for selected in evaluate(route, body):
                self.assertEqual(selected.returncode, 0, selected.stderr)
                self.assertEqual(selected.stdout, output, selected.args[-1])

        with fixture_api() as (api_url, response):
            raw_pr = {
                "number": 55, "state": "open", "base": {"ref": "main"},
                "head": {"sha": self.pr_head, "ref": self.branch, "repo": {"full_name": "fixture/repository"}},
                "labels": [{"name": "autorelease: pending"}],
            }
            foreign, other_branch, no_label, other_base = [copy.deepcopy(raw_pr) for _ in range(4)]
            foreign["head"]["repo"]["full_name"] = "another/repository"
            other_branch["head"]["ref"] = "feature-branch"
            no_label["labels"] = []
            other_base["base"]["ref"] = "release"
            invalid_prs = [foreign, other_branch, no_label, other_base]
            expect("pulls", invalid_prs + [raw_pr], self.pr_head + "\n")
            expect("pulls", invalid_prs, "")
            for version, expected in [("2.1.0", "2.1.0\n"), ("2.1.0-rc.1", ""), (2.1, ""), (None, ""), ("2.1.0\n", "")]:
                with self.subTest(manifest_version=version):
                    content = base64.b64encode(json.dumps({".": version}).encode()).decode()
                    expect("contents/.release-please-manifest.json", {"content": content}, expected)
            for malformed in evaluate("contents/.release-please-manifest.json", {"content": "not-base64"}):
                self.assertNotEqual(malformed.returncode, 0)

    def test_api_ambiguity_and_manifest_failures_never_allow_publication(self):
        cases = [
            ("pulls", {"error": True}),
            ("pulls", {"value": f"{self.pr_head}\n{self.pr_head}"}),
            ("pulls", {"value": "invalid-head"}),
            ("contents/.release-please-manifest.json", {"value": "2.1.0-rc.1"}),
            ("contents/.release-please-manifest.json", {"value": "null"}),
            ("contents/.release-please-manifest.json", {"value": "2.1.0\nINJECTED=true"}),
        ]
        for route, response in cases:
            with self.subTest(route=route, response=response):
                previous = self.responses[route]
                self.responses[route] = response
                self.assert_blocked(self.run_rc())
                self.assertEqual(self.read_file("github-env"), "")
                self.responses[route] = previous

    def test_exact_stable_target_guard_handles_http_and_transport_errors(self):
        self.assertIsNotNone(shutil.which("curl"), "curl is required for the exact stable-target test")
        metadata = self.resolve()
        with fixture_api() as (api_url, response):
            for status in [404, 200, 403, 500]:
                with self.subTest(status=status):
                    response["status"] = status
                    result = self.run_rc("Recheck Release Candidate Eligibility", GITHUB_API_URL=api_url, **metadata)
                    self.assertEqual(response["paths"][-1], "/repos/fixture/repository/git/ref/tags/v2.1.0")
                    if status == 404:
                        self.assertEqual(result.returncode, 0, result.stderr)
                        self.assertEqual(self.read_file("github-output"), "should_publish=true\n")
                    else:
                        self.assert_blocked(result, skip=status == 200)
            response["drop"] = True
            self.assert_blocked(self.run_rc("Recheck Release Candidate Eligibility", GITHUB_API_URL=api_url, **metadata))

    def test_candidate_build_and_publication_require_selection_and_final_eligibility(self):
        workflow = steps(RC_WORKFLOW)
        names = list(workflow)
        compute = names.index("Compute Release Candidate Metadata")
        recheck = names.index("Recheck Release Candidate Eligibility")
        for name in ["Install web dependencies", "Static Analysis", "Run Test Suite", "Cross-Compile Release Binaries", "Package Release Archives", "Verify Release Artifacts", "Recheck Release Candidate Eligibility"]:
            self.assertLess(compute, names.index(name))
            self.assertIn("        if: steps.candidate.outputs.should_publish == 'true'", workflow[name])
        self.assertLess(names.index("Verify Release Artifacts"), recheck)
        for name in ["Publish GitHub Pre-release", "Prune Older Release Candidates"]:
            self.assertLess(recheck, names.index(name))
            self.assertIn("        if: steps.eligibility.outputs.should_publish == 'true'", workflow[name])
        self.assertIn('        run: make release-build VERSION="${VERSION}" GIT_COMMIT="${GIT_COMMIT}" BUILD_DATE="${BUILD_DATE}"', workflow["Cross-Compile Release Binaries"])
        self.assertIn("          tag_name: ${{ env.VERSION }}", workflow["Publish GitHub Pre-release"])
        self.assertIn("          target_commitish: ${{ env.EXPECTED_SHA }}", workflow["Publish GitHub Pre-release"])

    def test_automatic_and_manual_events_bind_both_checkouts_to_candidate_commit(self):
        text = RC_WORKFLOW.read_text()
        self.assertIn("  workflow_run:", text)
        self.assertIn("workflows: [Release Please]", text)
        self.assertIn("types: [completed]", text)
        self.assertIn("  workflow_dispatch:", text)
        self.assertNotIn("  push:", text)
        self.assertIn("github.event.workflow_run.conclusion == 'success'", text)
        self.assertIn("github.event.workflow_run.head_sha || github.sha", text)
        self.assertEqual(text.count("ref: ${{ env.EXPECTED_SHA }}"), 2)

    def test_manual_source_detection_forces_build_and_release_commit_skips(self):
        for event, expected in [("workflow_dispatch", "true"), ("workflow_run", "false")]:
            with self.subTest(event=event):
                self.git("checkout", "--detach", self.pr_head)
                result = subprocess.run(["bash", "-c", shell("Check for Source Changes", RC_WORKFLOW)],
                                        cwd=self.repo, env=dict(self.env, EVENT_NAME=event, EXPECTED_SHA=self.pr_head),
                                        capture_output=True, text=True)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertIn(f"should_build={expected}", self.read_file("github-output"))
                (self.repo / "github-output").unlink()


if __name__ == "__main__":
    unittest.main()
