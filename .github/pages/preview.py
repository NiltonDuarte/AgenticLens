"""Compose static PR previews without executing code from a build artifact."""
from __future__ import annotations

import argparse
import base64
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tempfile
import urllib.request

ASSETS = {"style.css", "investigation.svg", "analytics.svg", "history.svg", "architecture.svg"}
SHA = re.compile(r"[0-9a-f]{40}\Z")
STATE = ".portal-state.json"


def require_sha(value: str) -> str:
    if not isinstance(value, str) or not SHA.fullmatch(value):
        raise ValueError("Expected a full lowercase commit SHA")
    return value


def validate_site(root: Path) -> None:
    if root.is_symlink() or not root.is_dir():
        raise ValueError(f"Not a regular site directory: {root}")
    required = {"index.html", "roadmap/index.html"} | {f"presentation/{name}" for name in ASSETS}
    found = set()
    for path in root.rglob("*"):
        relative = path.relative_to(root).as_posix()
        if path.is_symlink():
            raise ValueError(f"Symlink in site: {relative}")
        if path.is_dir():
            if relative not in {"presentation", "roadmap"} and not relative.startswith("roadmap/"):
                raise ValueError(f"Unexpected site directory: {relative}")
            continue
        if not path.is_file():
            raise ValueError(f"Not a regular site file: {relative}")
        allowed = relative in required or relative == ".tickets-site-output" or (
            relative.startswith("roadmap/") and path.suffix in {".html", ".css"}
            and not any(part.startswith(".") for part in path.relative_to(root).parts)
        )
        if not allowed:
            raise ValueError(f"Unexpected site file: {relative}")
        found.add(relative)
    if not required <= found:
        raise ValueError(f"Missing required site files: {sorted(required - found)}")


def read_state(previous: Path, repository: str) -> dict:
    if previous.is_symlink() or not previous.is_dir():
        raise ValueError("Previous publication is not a regular directory")
    # Validate every entry before copying, including directories outside known subtrees.
    for entry in previous.rglob("*"):
        if entry.is_symlink() or not (entry.is_file() or entry.is_dir()):
            raise ValueError("Unsafe previous publication entry")
    state = json.loads((previous / STATE).read_text())
    if state.get("version") != 1 or state.get("repository") != repository:
        raise ValueError("Publication state has unrelated repository or schema")
    require_sha(state["main_revision"])
    if not isinstance(state.get("previews"), dict):
        raise ValueError("Invalid preview state")
    for number, record in state["previews"].items():
        if not re.fullmatch(r"[1-9][0-9]*", number):
            raise ValueError("Invalid preview number")
        require_sha(record["revision"])
        validate_site(previous / "preview" / number)
    expected = {"index.html", "presentation", "roadmap", STATE, ".tickets-site-output"}
    if state["previews"]:
        expected.add("preview")
    if {entry.name for entry in previous.iterdir()} - expected:
        raise ValueError("Unexpected publication root entry")
    preview_root = previous / "preview"
    if preview_root.exists() and {p.name for p in preview_root.iterdir()} != set(state["previews"]):
        raise ValueError("Preview directories do not match state")
    return state


def compose(production: Path, previous: Path | None, candidate: Path | None, output: Path,
            *, repository: str, main_revision: str, pull_requests: list[dict],
            preview_number: int | None = None, preview_revision: str | None = None) -> None:
    require_sha(main_revision)
    if not re.fullmatch(r"[\w.-]+/[\w.-]+", repository):
        raise ValueError("Invalid repository")
    validate_site(production)
    state = read_state(previous, repository) if previous is not None else {"previews": {}}
    eligible = {}
    for pr in pull_requests:
        number = pr["number"]
        if type(number) is not int or number <= 0:
            raise ValueError("Invalid PR number")
        if (pr["state"] == "open" and pr["base"]["ref"] == "main" and pr["head"]["repo"]
                and pr["head"]["repo"]["full_name"] == repository):
            eligible[str(number)] = require_sha(pr["head"]["sha"])
    if candidate is not None:
        if type(preview_number) is not int or preview_number <= 0:
            raise ValueError("Candidate requires a positive preview number")
        if eligible.get(str(preview_number)) != require_sha(preview_revision):
            raise ValueError("Stale, closed or unrelated preview candidate")
        validate_site(candidate)
    elif preview_number is not None or preview_revision is not None:
        raise ValueError("Preview identity requires a candidate")
    if output.is_symlink() or any(parent.is_symlink() for parent in output.parents):
        raise ValueError("Output must not traverse a symlink")
    if output.exists() and (not output.is_dir() or not (output / STATE).is_file()):
        raise ValueError("Refusing to replace unmanaged output")
    inputs = [production, previous, candidate]
    for source in inputs:
        if source is not None and (output.resolve() == source.resolve() or output.resolve() in source.resolve().parents or source.resolve() in output.resolve().parents):
            raise ValueError("Output and inputs must be separate")
    output.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="portal-compose-", dir=output.parent) as temporary:
        staged = Path(temporary) / "site"
        shutil.copytree(production, staged)
        previews = {}
        for number, record in state["previews"].items():
            if number in eligible:
                shutil.copytree(previous / "preview" / number, staged / "preview" / number)
                previews[number] = record
        if candidate is not None:
            target = staged / "preview" / str(preview_number)
            if target.exists():
                shutil.rmtree(target)
            shutil.copytree(candidate, target)
            previews[str(preview_number)] = {"revision": preview_revision}
        (staged / STATE).write_text(json.dumps({"version": 1, "repository": repository,
            "main_revision": main_revision, "previews": previews}, indent=2) + "\n")
        if output.exists():
            shutil.rmtree(output)
        staged.rename(output)


def api(path: str) -> object:
    request = urllib.request.Request("https://api.github.com/" + path, headers={
        "Authorization": "Bearer " + os.environ["GH_TOKEN"],
        "Accept": "application/vnd.github+json", "X-GitHub-Api-Version": "2022-11-28"})
    with urllib.request.urlopen(request, timeout=30) as response:
        return json.load(response)


def open_prs(repository: str) -> list[dict]:
    result = []
    page = 1
    while True:
        batch = api(f"repos/{repository}/pulls?state=open&base=main&per_page=100&page={page}")
        result.extend(batch)
        if len(batch) < 100:
            return result
        page += 1


def resolve() -> None:
    repository = os.environ["GITHUB_REPOSITORY"]
    event = json.loads(Path(os.environ["GITHUB_EVENT_PATH"]).read_text())
    event_name = os.environ["GITHUB_EVENT_NAME"]
    main = api(f"repos/{repository}/commits/main")["sha"]
    number = None
    revision = ""
    if event_name == "pull_request_target" and event["action"] != "closed":
        number = event["number"]
    elif event_name == "workflow_dispatch":
        value = event.get("inputs", {}).get("preview_pr", "")
        if value:
            if not re.fullmatch(r"[1-9][0-9]*", value):
                raise ValueError("preview_pr must be a positive PR number")
            number = int(value)
        elif os.environ["GITHUB_REF"] != "refs/heads/main":
            raise ValueError("A feature-ref dispatch requires preview_pr")
    if number is not None:
        pr = api(f"repos/{repository}/pulls/{number}")
        if pr["state"] != "open" or not pr["head"]["repo"] or pr["head"]["repo"]["full_name"] != repository or pr["base"]["ref"] != "main":
            raise ValueError("Only open same-repository PRs targeting main have previews")
        revision = require_sha(pr["head"]["sha"])
        if event_name == "workflow_dispatch" and os.environ["GITHUB_REF"] != "refs/heads/main" and os.environ["GITHUB_SHA"] != revision:
            raise ValueError("Bootstrap dispatch must use the PR's current head")
    with open(os.environ["GITHUB_OUTPUT"], "a") as output:
        output.write(f"main_revision={require_sha(main)}\npreview_number={number or ''}\npreview_revision={revision}\n")


def prepare(arguments: argparse.Namespace) -> None:
    repository = os.environ["GITHUB_REPOSITORY"]
    if api(f"repos/{repository}/commits/main")["sha"] != arguments.main_revision:
        raise ValueError("Main advanced during build; rerun publication")
    compose(Path(arguments.production), Path(arguments.previous) if Path(arguments.previous).exists() else None,
        Path(arguments.candidate) if arguments.preview_number else None, Path(arguments.output),
        repository=repository, main_revision=arguments.main_revision, pull_requests=open_prs(repository),
        preview_number=int(arguments.preview_number) if arguments.preview_number else None,
        preview_revision=arguments.preview_revision or None)



def assert_no_pending(directory: Path) -> None:
    if (directory / ".pending-publication.json").exists():
        raise RuntimeError("Unfinished Pages publication: recover the recorded run before publishing again; see .github/pages/README.md")


def validate_loaded_state(directory: Path, repository: str) -> None:
    assert_no_pending(directory)
    read_state(directory / "site", repository)


def state_repository(command: str, directory: Path, site: Path | None = None) -> None:
    """Use a durable static branch; never run code or hooks from that branch."""
    repository = os.environ["GITHUB_REPOSITORY"]
    environment = os.environ.copy()
    authorization = base64.b64encode(("x-access-token:" + os.environ["GH_TOKEN"]).encode()).decode()
    environment.update({"GIT_CONFIG_COUNT": "2", "GIT_CONFIG_KEY_0": "http.https://github.com/.extraheader",
        "GIT_CONFIG_VALUE_0": "AUTHORIZATION: basic " + authorization,
        "GIT_CONFIG_KEY_1": "core.hooksPath", "GIT_CONFIG_VALUE_1": "/dev/null"})

    def git(*arguments: str, check: bool = True) -> subprocess.CompletedProcess:
        result = subprocess.run(["git", *arguments], cwd=directory, env=environment,
            text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if check and result.returncode:
            # Do not print subprocess arguments or credential-bearing config.
            raise RuntimeError("Static publication Git operation failed: " + result.stderr.replace(os.environ["GH_TOKEN"], "[redacted]"))
        return result

    if command == "load-state":
        directory.mkdir(parents=True, exist_ok=False)
        git("init", "--initial-branch=portal-pages-state")
        git("remote", "add", "origin", f"https://github.com/{repository}.git")
        exists = git("ls-remote", "--exit-code", "origin", "refs/heads/portal-pages-state", check=False)
        if exists.returncode == 0:
            git("fetch", "--depth=1", "origin", "portal-pages-state")
            git("reset", "--hard", "FETCH_HEAD")
            validate_loaded_state(directory, repository)
        elif exists.returncode != 2:
            raise RuntimeError("Cannot determine whether publication state exists")
    else:
        git("config", "user.name", "github-actions[bot]")
        git("config", "user.email", "41898282+github-actions[bot]@users.noreply.github.com")
        if command == "begin-state":
            marker = directory / ".pending-publication.json"
            if marker.exists():
                raise RuntimeError("A publication is already pending")
            marker.write_text(json.dumps({"run_id": os.environ["GITHUB_RUN_ID"],
                "repository": repository, "recovery_artifact": "portal-publication-recovery"}) + "\n")
            git("add", ".pending-publication.json")
            git("commit", "-m", "chore: record pending portal publication")
            git("push", "origin", "HEAD:refs/heads/portal-pages-state")
            return
        if site is None:
            raise ValueError("Saving state requires the composed site")
        read_state(site, repository)
        destination = directory / "site"
        if destination.exists():
            shutil.rmtree(destination)
        shutil.copytree(site, destination)
        marker = directory / ".pending-publication.json"
        if not marker.is_file():
            raise RuntimeError("No pending publication to finalize")
        pending = json.loads(marker.read_text())
        if pending.get("repository") != repository or pending.get("run_id") != os.environ["GITHUB_RUN_ID"]:
            raise RuntimeError("Pending publication belongs to another run")
        marker.unlink()
        git("add", "--all", "site", ".pending-publication.json")
        if git("diff", "--cached", "--quiet", check=False).returncode == 0:
            return
        git("commit", "-m", "chore: retain published portal and PR previews")
        # Retry transient network failures; an actual concurrent update stays a failure.
        for attempt in range(3):
            result = git("push", "origin", "HEAD:refs/heads/portal-pages-state", check=False)
            if result.returncode == 0:
                return
        raise RuntimeError("Pages deployed but durable state push failed; recover the pending transaction before another publication")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)
    commands.add_parser("resolve")
    for command in ("load-state", "begin-state", "save-state"):
        state_parser = commands.add_parser(command)
        state_parser.add_argument("--directory", required=True)
        state_parser.add_argument("--site")
    preparation = commands.add_parser("prepare")
    for name in ("production", "previous", "candidate", "output", "main-revision", "preview-number", "preview-revision"):
        preparation.add_argument("--" + name, required=name not in {"preview-number", "preview-revision"}, default="")
    args = parser.parse_args()
    if args.command == "resolve":
        resolve()
    elif args.command == "prepare":
        prepare(args)
    else:
        state_repository(args.command, Path(args.directory), Path(args.site) if args.site else None)
