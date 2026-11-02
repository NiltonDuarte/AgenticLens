#!/usr/bin/env bash
# Contributor entrypoint; share release-only installation behavior with the setup skill.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Contributor checkouts have no installation marker; preserve their explicit
# entrypoint default without letting an installed setup skill guess a release.
if [[ -z "${AGENTIC_LENS_VERSION:-}" && ! -f "$SCRIPT_DIR/../.agents/skills/agentic-lens-setup/RELEASE_VERSION" ]]; then
    export AGENTIC_LENS_VERSION=latest
fi
exec bash "$SCRIPT_DIR/../.agents/skills/agentic-lens-setup/scripts/download.sh" "$@"
