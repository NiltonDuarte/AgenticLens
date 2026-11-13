#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
CLI="$ROOT/tools/openspec/node_modules/@fission-ai/openspec/bin/openspec.js"
if [[ ! -f "$CLI" ]]; then
  printf '%s\n' 'Install the pinned development tool with: npm ci --prefix tools/openspec' >&2
  exit 1
fi
export OPENSPEC_TELEMETRY=0
exec node "$CLI" "$@"
