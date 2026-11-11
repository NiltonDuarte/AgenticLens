#!/bin/sh
# AgenticLens advisory pre-commit v1
# wrapper-base64: L2ZpeHR1cmUgb3duZXIncy9iaW4vYWdlbnRpYy1sZW5z
project=$(git rev-parse --show-toplevel 2>/dev/null) || {
  printf '%s\n' 'AgenticLens capture advisory: project unavailable; inspect capture status.' >&2
  exit 0
}
if [ ! -f "$project/.agenticlens-config.yaml" ] &&
   [ ! -f "$project/.agenticlens-config.yml" ]; then
  exit 0
fi
tree_kill=
case "$(uname -s 2>/dev/null)" in
  MINGW*|MSYS*|CYGWIN*)
    tree_kill="${SYSTEMROOT:-$SystemRoot}/System32/taskkill.exe"
    if [ ! -x "$tree_kill" ] || [ ! -r "/proc/$$/winpid" ]; then
      printf '%s\n' 'AgenticLens capture advisory: Windows process-tree cleanup unavailable; inspect capture status explicitly.' >&2
      exit 0
    fi
    ;;
esac
(
  '/fixture owner'"'"'s/bin/agentic-lens' '--al-offline' '--al-project' "$project" 'capture' 'check' '--repo' "$project" </dev/null >/dev/null &
  handler=$!
  handler_winpid=
  if [ -n "$tree_kill" ]; then
    IFS= read -r handler_winpid 2>/dev/null < "/proc/$handler/winpid" || :
    case "$handler_winpid" in ''|*[!0-9]*) handler_winpid= ;; esac
  fi
  (
    sleep 3
    if [ -n "$handler_winpid" ]; then
      current_winpid=
      IFS= read -r current_winpid 2>/dev/null < "/proc/$handler/winpid" || :
      if [ "$current_winpid" = "$handler_winpid" ]; then
        MSYS2_ARG_CONV_EXCL='*' "$tree_kill" /PID "$handler_winpid" /T /F
      fi
    fi
    kill -TERM "$handler" 2>/dev/null
    sleep 1
    kill -KILL "$handler" 2>/dev/null
  ) >/dev/null 2>&1 &
  watchdog=$!
  wait "$handler" 2>/dev/null
  status=$?
  kill "$watchdog" 2>/dev/null
  wait "$watchdog" 2>/dev/null
  if [ "$status" -ne 0 ]; then
    printf '%s\n' 'AgenticLens capture advisory: check unavailable; inspect capture status or refresh user wrapper setup.' >&2
  fi
) 2>&1 | dd bs=1 count=512 2>/dev/null | LC_ALL=C awk 'NR == 1 { sub(/\r$/, ""); print }' >&2
exit 0
