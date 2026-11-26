#!/bin/bash
# Transparent make SHELL wrapper: logs start/end/exit of each top-level recipe line.
log="${TIMED_SH_LOG:-/dev/null}"
if [ "${MAKELEVEL:-0}" = "1" ] && [ "$1" = "-c" ]; then
  start=$(date +%s.%N); load=$(cut -d' ' -f1 /proc/loadavg)
  /bin/sh "$@"; rc=$?
  end=$(date +%s.%N)
  cmd=$(printf '%s' "$2" | tr '\n' ' ' | cut -c1-160)
  printf '%s\t%s\t%.1f\t%s\t%s\t%s\n' "$start" "$end" "$(echo "$end - $start" | bc)" "$rc" "$load" "$cmd" >> "$log"
  exit $rc
fi
exec /bin/sh "$@"
