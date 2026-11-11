#!/bin/sh
# Convenience invocation of the installed native manager.
set -eu
manager="$HOME/.local/bin/agentic-lens"
if [ ! -x "$manager" ]; then
 printf 'Error: native manager missing at %s; explicitly run .agents/skills/agentic-lens-setup/scripts/download.sh --project PROJECT.\n' "$manager" >&2
 printf 'If setup scripts are missing, restore the complete setup/operator skill bundle at the selected release; see https://github.com/NiltonDuarte/AgenticLens/blob/main/.agents/skills/agentic-lens-setup/SKILL.md\n' >&2
 exit 1
fi
exec "$manager" "$@"
