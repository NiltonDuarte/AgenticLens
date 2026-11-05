#!/bin/sh
# Run from the backend module. Discovery failure must not become an empty pass.
set -eu
lane=${1:?usage: go-test-packages.sh unit|integration|all}
case "$lane" in unit|integration|all) ;; *) echo "Unknown Go test lane: $lane" >&2; exit 2 ;; esac
module=$(go list -m) || exit 1
[ -n "$module" ] || { echo 'Go module discovery was empty' >&2; exit 1; }
if [ "$lane" = unit ]; then
  packages=$(go list ./internal/domain/...) || exit 1
else
  packages=$(go list ./...) || exit 1
fi
selected=''
for package in $packages; do
  is_unit=0
  case "$package" in "$module/internal/domain"|"$module/internal/domain/"*) is_unit=1 ;; esac
  if [ "$lane" = all ] || { [ "$lane" = unit ] && [ "$is_unit" -eq 1 ]; } || { [ "$lane" = integration ] && [ "$is_unit" -eq 0 ]; }; then
    selected="$selected$package
"
  fi
done
[ -n "$selected" ] || { echo "Go $lane package discovery was empty" >&2; exit 1; }
printf '%s' "$selected"
