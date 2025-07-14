#!/bin/bash
# Save as ~/bin/nc or somewhere in your PATH
args=()
for arg in "$@"; do
    if [[ "$arg" == "-c" ]]; then
        args+=("-C")
    else
        args+=("$arg")
    fi
done
exec /usr/bin/nc "${args[@]}"