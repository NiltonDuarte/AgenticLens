from __future__ import annotations

import json
import tempfile
from pathlib import Path
from unittest.mock import patch

from agentic_lens.adapters.repository import resolve_workspaces_dir


def test_resolve_workspaces_dir_override() -> None:
    override = Path("/custom/workspaces")
    assert resolve_workspaces_dir(override) == override


def test_resolve_workspaces_dir_from_config() -> None:
    with tempfile.TemporaryDirectory() as tmpdir:
        tmp_path = Path(tmpdir)
        config_dir = tmp_path / ".agenticlens"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"
        config_file.write_text(
            json.dumps({"workspaces_dir": "custom_ws"}),
            encoding="utf-8",
        )
        custom_ws = tmp_path / "custom_ws"
        custom_ws.mkdir()

        with patch("pathlib.Path.cwd", return_value=tmp_path):
            resolved = resolve_workspaces_dir()
            assert resolved == custom_ws
