#!/usr/bin/env bash
set -euo pipefail
root=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd -P)
exec python3 -B "$root/scripts/tests/fixtures/capture-release-gate/gate.py" "$@"
