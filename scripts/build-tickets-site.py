#!/usr/bin/env python3
"""Render the configured ticket plan as a small GitHub Pages artifact."""

import argparse
import html
import re
import shutil
import tempfile
from pathlib import Path
from urllib.parse import quote, unquote, urlsplit, urlunsplit
from xml.etree.ElementTree import Element

import markdown
import yaml
from github_slugger import GithubSlugger
from markdown.extensions import Extension
from markdown.extensions.toc import TocExtension
from markdown.treeprocessors import Treeprocessor


STYLES = """
:root { color-scheme: light dark; font-family: system-ui, sans-serif; line-height: 1.65; }
body { margin: 0; background: light-dark(#f5f7fa, #111820); color: light-dark(#172330, #e4eaf0); }
header, main, footer { max-width: 56rem; margin: auto; padding: 1.25rem; }
header { border-bottom: 1px solid light-dark(#d6dce3, #384452); }
header nav { display: flex; flex-wrap: wrap; gap: 1.25rem; }
main { overflow-wrap: anywhere; }
a { color: light-dark(#095db5, #8bc4ff); text-underline-offset: .18em; }
a:focus-visible { outline: 3px solid currentColor; outline-offset: 4px; }
h1, h2, h3 { line-height: 1.25; margin-top: 1.6em; }
li { margin-block: .35em; }
pre { overflow-x: auto; padding: 1rem; background: light-dark(#e8edf3, #202c38); }
code { font-size: .9em; }
table { display: block; overflow-x: auto; border-collapse: collapse; }
td, th { padding: .5rem; border: 1px solid light-dark(#cbd3dd, #435162); }
img { max-width: 100%; height: auto; }
footer { font-size: .9rem; }
"""
MARKER = ".tickets-site-output"


class SiteLinks(Treeprocessor):
    def __init__(self, md, source: Path, plan: Path, pages: dict[Path, Path], project: Path,
                 repository: str, revision: str, base_path: str):
        super().__init__(md)
        self.source, self.pages, self.project = source, pages, project
        self.plan = plan
        self.repository, self.revision, self.base_path = repository, revision, base_path

    def run(self, root: Element) -> Element:
        for node in root.iter():
            attribute = "href" if node.tag == "a" else "src" if node.tag == "img" else None
            if attribute is None or attribute not in node.attrib:
                continue
            link = urlsplit(node.attrib[attribute])
            if link.scheme or link.netloc or not link.path:
                continue
            # Repository-root paths and document-relative paths share source semantics.
            target = ((self.project / unquote(link.path).lstrip("/")) if link.path.startswith("/")
                      else self.source.parent / unquote(link.path)).resolve()
            if not target.is_relative_to(self.project):
                raise ValueError(f"Link escapes repository in {self.source}: {link.path}")
            if target in self.pages and node.tag == "a":
                path = self.base_path + "/" + quote(self.pages[target].as_posix())
                node.set(attribute, urlunsplit(("", "", path, link.query, link.fragment)))
            else:
                if target.suffix == ".md" and target.is_relative_to(self.plan):
                    raise ValueError(f"Missing published Markdown target: {target}")
                relative = quote(target.relative_to(self.project).as_posix())
                revision = quote(self.revision, safe="")
                origin = (f"https://raw.githubusercontent.com/{self.repository}/{revision}"
                          if node.tag == "img" else
                          f"https://github.com/{self.repository}/blob/{revision}")
                node.set(attribute, origin + "/" + relative
                         + ("?" + link.query if link.query else "")
                         + ("#" + link.fragment if link.fragment else ""))
        return root


class LinkExtension(Extension):
    def __init__(self, source: Path, plan: Path, pages: dict[Path, Path], project: Path,
                 repository: str, revision: str, base_path: str):
        super().__init__()
        self.arguments = (source, plan, pages, project, repository, revision, base_path)

    def extendMarkdown(self, md):
        processor = SiteLinks(md, *self.arguments)
        md.treeprocessors.register(processor, "site-links", 5)


def build(project: Path, output: Path, repository: str, revision: str, base_path: str) -> int:
    project = project.resolve()
    config_path = project / ".agenticlens-config.yaml"
    config = yaml.safe_load(config_path.read_text(encoding="utf-8")) if config_path.exists() else {}
    if config is None:
        config = {}
    if not isinstance(config, dict):
        raise ValueError("Project configuration must be a YAML mapping")
    workspaces = config.get("workspaces", {})
    if not isinstance(workspaces, dict):
        raise ValueError("workspaces must be a YAML mapping")
    workspace_dir = workspaces.get("dir", "workspaces")
    if not isinstance(workspace_dir, str) or not workspace_dir.strip():
        raise ValueError("workspaces.dir must be a nonempty path")
    plan = (project / workspace_dir / "plan").resolve()
    if not plan.is_relative_to(project):
        raise ValueError("Configured plan must be inside the repository")
    roadmap = plan / "TICKETS.md"
    if not roadmap.is_file():
        raise ValueError(f"Roadmap not found: {roadmap}")
    if output.is_symlink():
        raise ValueError("Output must not be a symlink")
    output = output.resolve()
    if project.is_relative_to(output) or plan.is_relative_to(output) or output.is_relative_to(plan):
        raise ValueError("Output must be separate from project sources and the plan")
    if output.exists() and (not output.is_dir() or
                            (any(output.iterdir()) and not (output / MARKER).is_file())):
        raise ValueError("Refusing to replace an unmanaged output directory")
    pages: dict[Path, Path] = {}
    for source in sorted(plan.rglob("*.md")):
        if source.is_symlink() or not source.resolve().is_relative_to(plan):
            raise ValueError(f"Plan source must not escape through a symlink: {source}")
        target = Path("index.html") if source == roadmap else source.relative_to(plan).with_suffix(".html")
        if target in pages.values():
            raise ValueError(f"Conflicting output page: {target}")
        pages[source.resolve()] = target
    output.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="tickets-site-", dir=output.parent) as temporary:
        staging = Path(temporary) / "site"
        staging.mkdir()
        (staging / MARKER).write_text("Generated by build-tickets-site.py\n", encoding="utf-8")
        (staging / "style.css").write_text(STYLES, encoding="utf-8")
        (staging / ".nojekyll").touch()
        home = html.escape(base_path + "/index.html", quote=True)
        stylesheet = html.escape(base_path + "/style.css", quote=True)
        repository_url = html.escape(f"https://github.com/{repository}", quote=True)
        for source, target in pages.items():
            slugger = GithubSlugger()
            renderer = markdown.Markdown(tab_length=2, extensions=["extra", "sane_lists",
                TocExtension(slugify=lambda value, separator: slugger.slug(value)),
                LinkExtension(source, plan, pages, project, repository, revision, base_path)])
            content = renderer.convert(source.read_text(encoding="utf-8"))
            title = html.escape("AgenticLens Roadmap" if source == roadmap else source.stem)
            document = f"""<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>{title}</title><link rel="stylesheet" href="{stylesheet}"></head>
<body><header><nav aria-label="Main navigation"><a href="{home}">Back to roadmap</a>
<a href="{repository_url}">View repository</a></nav></header>
<main>{content}</main><footer>Published from the repository's ticket Markdown.</footer></body></html>
"""
            destination = staging / target
            destination.parent.mkdir(parents=True, exist_ok=True)
            destination.write_text(document, encoding="utf-8")
        if output.exists():
            shutil.rmtree(output)
        staging.rename(output)
    return len(pages)


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--project", type=Path, default=Path("."))
    parser.add_argument("--output", type=Path, default=Path("dist/tickets-site"))
    parser.add_argument("--repository", required=True, help="GitHub OWNER/REPOSITORY")
    parser.add_argument("--revision", required=True, help="Source commit SHA or ref for supporting links")
    parser.add_argument("--base-path", default="", help="Project-site URL prefix, e.g. /AgenticLens")
    args = parser.parse_args()
    if not re.fullmatch(r"[\w.-]+/[\w.-]+", args.repository):
        parser.error("--repository must be OWNER/REPOSITORY")
    if not args.revision.strip():
        parser.error("--revision must not be empty")
    base_path = args.base_path.rstrip("/")
    if base_path and (not base_path.startswith("/") or base_path.startswith("//") or
                      any(character in base_path for character in "?#")):
        parser.error("--base-path must be an absolute URL path without query or fragment")
    base_path = quote(unquote(base_path), safe="/")
    output = args.output if args.output.is_absolute() else args.project / args.output
    try:
        count = build(args.project, output, args.repository, args.revision, base_path)
    except (OSError, ValueError, yaml.YAMLError) as error:
        parser.exit(1, f"Ticket site build failed: {error}\n")
    print(f"Built {count} Markdown pages in {output}")


if __name__ == "__main__":
    main()
