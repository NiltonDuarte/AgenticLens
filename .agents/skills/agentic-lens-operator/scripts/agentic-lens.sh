#!/bin/sh
# Compatibility launcher for the stable user-level common entrypoint.
set -eu
wrapper="$HOME/.local/bin/agentic-lens"
if [ ! -x "$wrapper" ]; then
 printf 'Error: common wrapper missing at %s; explicitly run .agents/skills/agentic-lens-setup/scripts/download.sh --project PROJECT.\n' "$wrapper" >&2
 printf 'If setup scripts are missing, restore the complete setup/operator skill bundle at the selected release; see https://github.com/NiltonDuarte/AgenticLens/blob/main/.agents/skills/agentic-lens-setup/SKILL.md\n' >&2
 exit 1
fi
exec "$wrapper" "$@"
