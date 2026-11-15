#!/usr/bin/env sh
set -eu
cd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
if [ "${1:-}" = "--test" ]; then
    shift
    exec go test ./... "$@"
fi
exec go run ./src "$@"
