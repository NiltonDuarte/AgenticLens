"""Exercise the deployed static builder against isolated Markdown repositories."""

from html.parser import HTMLParser
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


BUILDER = Path(__file__).resolve().parents[1] / "build-tickets-site.py"


class Document(HTMLParser):
    def __init__(self, path):
        super().__init__()
        self.links = []
        self.ids = []
        self.text = []
        self.list_items = []
        self.list_stack = []
        self.feed(path.read_text(encoding="utf-8"))

    def handle_starttag(self, tag, attrs):
        attributes = dict(attrs)
        if tag == "a":
            self.links.append(attributes.get("href"))
        if "id" in attributes:
            self.ids.append(attributes["id"])
        if tag == "li":
            item = {"text": "", "parent": self.list_stack[-1] if self.list_stack else None}
            self.list_items.append(item)
            self.list_stack.append(len(self.list_items) - 1)

    def handle_endtag(self, tag):
        if tag == "li":
            self.list_stack.pop()

    def handle_data(self, data):
        self.text.append(data)
        if self.list_stack:
            self.list_items[self.list_stack[-1]]["text"] += data


class TicketsSiteTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="tickets-site-")
        self.addCleanup(self.temp.cleanup)
        self.project = Path(self.temp.name) / "repository with spaces"
        self.project.mkdir()
        self.output = self.project / "dist" / "tickets-site"

    def fixture(self, workspace="custom tickets", config=True):
        if config:
            (self.project / ".agenticlens-config.yaml").write_text(
                f'workspaces:\n  dir: "{workspace}"\n', encoding="utf-8"
            )
        plan = self.project / workspace / "plan"
        (plan / "to-do").mkdir(parents=True)
        (plan / "done").mkdir()
        (plan / "TICKETS.md").write_text(
            "# Fixture roadmap\n\n## Ready\n\n"
            "- [TASK-Z](to-do/TASK-Z%20sample.md#details)\n\n"
            "## Blocked\n\nTASK-Z blocks TASK-NEXT.\n",
            encoding="utf-8",
        )
        self.ticket = plan / "to-do" / "TASK-Z sample.md"
        self.ticket.write_text(
            "# TASK-Z\n\n## Details\n\n**Fixture requirement**\n\n"
            "[Roadmap](../TICKETS.md#ready)\n\n"
            "[Supporting spec](../../PO-fixture/SPEC.md#context)\n\n"
            "[Sibling](../done/TASK-OLD%20archived.md)\n\n"
            "[Anchor](#details)\n\n"
            "[External](https://example.org/docs?q=value#section)\n",
            encoding="utf-8",
        )
        (plan / "done" / "TASK-OLD archived.md").write_text(
            "# Completed fixture\n", encoding="utf-8"
        )
        supporting = self.project / workspace / "PO-fixture"
        supporting.mkdir()
        (supporting / "SPEC.md").write_text("# Context\n", encoding="utf-8")
        (supporting / "agentic_metrics.json").write_text(
            '{"private_capture":"UNPUBLISHED_CAPTURE_SENTINEL"}', encoding="utf-8"
        )
        return plan

    def build(self, base_path="/FixtureRepo", output=None):
        return subprocess.run(
            [
                sys.executable, str(BUILDER), "--project", str(self.project),
                "--output", str(output or self.output),
                "--repository", "fixture-owner/FixtureRepo",
                "--revision", "0123456789abcdef",
                "--base-path", base_path,
            ],
            text=True, capture_output=True, check=False,
        )

    def assert_build(self, **kwargs):
        result = self.build(**kwargs)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_custom_root_renders_links_and_preserves_sources(self):
        self.fixture()
        originals = {
            path.relative_to(self.project): path.read_bytes()
            for path in self.project.rglob("*") if path.is_file()
        }
        self.assert_build()
        roadmap = Document(self.output / "index.html")
        self.assertIn("Fixture roadmap", " ".join(roadmap.text))
        self.assertIn("Ready", " ".join(roadmap.text))
        self.assertIn("Blocked", " ".join(roadmap.text))
        self.assertIn("/FixtureRepo/to-do/TASK-Z%20sample.html#details", roadmap.links)
        ticket_path = self.output / "to-do" / "TASK-Z sample.html"
        ticket = Document(ticket_path)
        self.assertIn("details", ticket.ids)
        self.assertIn("/FixtureRepo/index.html", ticket.links)
        self.assertIn("/FixtureRepo/index.html#ready", ticket.links)
        self.assertIn("/FixtureRepo/done/TASK-OLD%20archived.html", ticket.links)
        self.assertIn(
            "https://github.com/fixture-owner/FixtureRepo/blob/0123456789abcdef/"
            "custom%20tickets/PO-fixture/SPEC.md#context", ticket.links,
        )
        self.assertIn("#details", ticket.links)
        self.assertIn("https://example.org/docs?q=value#section", ticket.links)
        markup = ticket_path.read_text(encoding="utf-8")
        self.assertIn('name="viewport"', markup)
        self.assertIn("<strong>Fixture requirement</strong>", markup)
        for relative, before in originals.items():
            self.assertEqual((self.project / relative).read_bytes(), before)
        self.assertFalse(list(self.output.rglob("*.md")))
        self.assertFalse(list(self.output.rglob("*.json")))
        for path in self.output.rglob("*"):
            if path.is_file():
                self.assertNotIn(b"UNPUBLISHED_CAPTURE_SENTINEL", path.read_bytes())

    def test_fallback_root_without_config_and_without_workspace_key(self):
        self.fixture(workspace="workspaces", config=False)
        for config in (None, "project_name: Fixture\n", "workspaces: {}\n"):
            with self.subTest(config=config):
                if config is not None:
                    (self.project / ".agenticlens-config.yaml").write_text(config)
                self.assert_build(base_path="")
                roadmap = Document(self.output / "index.html")
                self.assertIn("/to-do/TASK-Z%20sample.html#details", roadmap.links)

    def test_github_heading_fragments_match_punctuation_and_duplicate_ids(self):
        plan = self.fixture()
        (plan / "TICKETS.md").write_text(
            "# Roadmap\n\n[API](to-do/TASK-Z%20sample.md#api--ui)\n\n"
            "[Repeated heading](to-do/TASK-Z%20sample.md#details-1)\n",
            encoding="utf-8",
        )
        self.ticket.write_text(
            "# TASK-Z\n\n[API](#api--ui)\n\n[Repeated heading](#details-1)\n\n"
            "## API & UI\n\nAPI requirement.\n\n"
            "## Details\n\nFirst section.\n\n## Details\n\nSecond section.\n",
            encoding="utf-8",
        )
        self.assert_build()
        roadmap = Document(self.output / "index.html")
        ticket = Document(self.output / "to-do" / "TASK-Z sample.html")
        for fragment in ("api--ui", "details", "details-1"):
            self.assertIn(fragment, ticket.ids)
        for fragment in ("api--ui", "details-1"):
            self.assertIn(f"/FixtureRepo/to-do/TASK-Z%20sample.html#{fragment}", roadmap.links)
            self.assertIn(f"#{fragment}", ticket.links)

    def test_two_space_metadata_bullets_remain_children_of_their_ticket(self):
        plan = self.fixture()
        (plan / "TICKETS.md").write_text(
            "# Roadmap\n\n## Ready\n\n"
            "- **TASK-Z**: Current ticket\n"
            "  - **Status**: Ready\n"
            "  - **File**: [TASK-Z](to-do/TASK-Z%20sample.md)\n\n"
            "- **TASK-OLD**: Previous ticket\n"
            "  - **Status**: Completed\n"
            "  - **File**: [TASK-OLD](done/TASK-OLD%20archived.md)\n",
            encoding="utf-8",
        )
        self.assert_build()
        document = Document(self.output / "index.html")
        parents = [index for index, item in enumerate(document.list_items)
                   if item["parent"] is None]
        self.assertEqual(len(parents), 2, document.list_items)
        for parent, ticket, status in zip(parents, ("TASK-Z", "TASK-OLD"), ("Ready", "Completed")):
            self.assertIn(ticket, document.list_items[parent]["text"])
            children = [" ".join(item["text"].split()) for item in document.list_items
                        if item["parent"] == parent]
            self.assertEqual(children, [f"Status: {status}", f"File: {ticket}"])

    def test_repeated_build_removes_stale_pages(self):
        plan = self.fixture()
        self.assert_build()
        archived = self.output / "done" / "TASK-OLD archived.html"
        self.assertTrue(archived.exists())
        (plan / "done" / "TASK-OLD archived.md").unlink()
        self.ticket.write_text("# TASK-Z\n\nUpdated requirement.\n", encoding="utf-8")
        self.assert_build()
        self.assertFalse(archived.exists())
        ticket = Document(self.output / "to-do" / "TASK-Z sample.html")
        self.assertIn("Updated requirement.", " ".join(ticket.text))

    def test_invalid_inputs_and_output_do_not_destroy_source_or_unowned_files(self):
        plan = self.fixture()
        roadmap = plan / "TICKETS.md"
        original = roadmap.read_bytes()
        for output in (self.project, plan, plan / "generated"):
            with self.subTest(output=output):
                result = self.build(output=output)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(roadmap.read_bytes(), original)
        self.output.mkdir(parents=True)
        unowned = self.output / "user-owned.txt"
        unowned.write_text("retain this file", encoding="utf-8")
        self.assertNotEqual(self.build().returncode, 0)
        self.assertEqual(unowned.read_text(), "retain this file")
        roadmap.unlink()
        missing = self.build(output=self.project / "dist" / "missing-input")
        self.assertNotEqual(missing.returncode, 0)
        self.assertIn("TICKETS.md", missing.stdout + missing.stderr)


if __name__ == "__main__":
    unittest.main()
