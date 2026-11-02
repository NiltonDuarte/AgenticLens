#!/usr/bin/env bash
# Contributor entrypoint; share release-only installation behavior with the setup skill.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec bash "$SCRIPT_DIR/../.agents/skills/agentic-lens-setup/scripts/download.sh" "$@"
