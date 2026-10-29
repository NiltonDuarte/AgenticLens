#!/usr/bin/env bash
set -euo pipefail

runtime=${1:?usage: opencode-compat.sh v1|v2|both}
root=$(cd "$(dirname "$0")/.." && pwd)
v1_bin=${OPENCODE_V1_BIN:-opencode}
v2_bin=${OPENCODE_V2_BIN:-opencode2}

run_one() {
  (
    local version=$1 bin=$2
    local temp port provider_pid=""
    temp=$(mktemp -d "${TMPDIR:-/tmp}/agenticlens-opencode-${version}.XXXXXX")
    cleanup() {
      if [[ -n "$provider_pid" ]]; then kill "$provider_pid" 2>/dev/null || true; fi
      rm -rf "$temp"
    }
    trap cleanup EXIT
    mkdir -p "$temp/home" "$temp/data" "$temp/config" "$temp/state" "$temp/project"
    cp "$root/alspaces/AL-58/evidence/provider-${version}.json" "$temp/project/opencode.json"
    printf '%s\n' 'AgenticLens isolated runtime marker.' > "$temp/project/marker.txt"
    export HOME="$temp/home" XDG_DATA_HOME="$temp/data" XDG_CONFIG_HOME="$temp/config" XDG_STATE_HOME="$temp/state"
    export OPENCODE_DB="$temp/data/opencode.db"
    port=8766; [[ "$version" == v2 ]] && port=8765
    python3 "$root/alspaces/AL-58/evidence/mock_openai_server.py" --runtime "$version" >"$temp/provider.log" 2>&1 &
    provider_pid=$!
    for _ in {1..50}; do if (echo >"/dev/tcp/127.0.0.1/$port") >/dev/null 2>&1; then break; fi; sleep 0.1; done
    cd "$temp/project"
    if [[ "$version" == v1 ]]; then
      "$bin" run --auto --model al58local/mock-model --title 'AgenticLens v1 compatibility' 'Read marker.txt and report its contents.'
    else
      "$bin" --standalone run --auto --model al58local/mock-model --title 'AgenticLens v2 compatibility' 'Read marker.txt and report its contents.'
    fi
  )
}

case "$runtime" in
  v1) run_one v1 "$v1_bin" ;;
  v2) run_one v2 "$v2_bin" ;;
  both) run_one v1 "$v1_bin"; run_one v2 "$v2_bin"; (cd "$root/backend" && go test ./internal/adapters/opencode ./cmd/agentic-lens) ;;
  *) echo "unknown runtime '$runtime'; use v1, v2, or both" >&2; exit 2 ;;
esac
