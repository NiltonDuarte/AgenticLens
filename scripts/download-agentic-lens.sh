#!/usr/bin/env bash
# Contributor installer shares exact project-pin installation semantics.
set -euo pipefail
SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
exec bash "$SCRIPT_DIR/../.agents/skills/agentic-lens-setup/scripts/download.sh" "$@"
