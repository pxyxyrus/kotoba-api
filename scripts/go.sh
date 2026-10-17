#!/bin/sh
# Keep downloaded toolchains and caches inside this workspace.
set -eu
KOTOBA_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
export GOPATH="$KOTOBA_ROOT/work/go"
export GOMODCACHE="$KOTOBA_ROOT/work/go/pkg/mod"
export GOCACHE="$KOTOBA_ROOT/work/go/build-cache"
export GOBIN="$KOTOBA_ROOT/work/go/bin"
export STATICCHECK_CACHE="$KOTOBA_ROOT/work/go/staticcheck-cache"
export GOTOOLCHAIN=auto
mkdir -p "$GOCACHE" "$GOBIN"
cd "$KOTOBA_ROOT"
exec go "$@"
