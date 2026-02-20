#!/usr/bin/env sh
set -eu

ID=${ID:-0}
LOG=${LOG:-exampled.log}

export EXAMPLED_HOME="/data/node${ID}"

if [ ! -d "$EXAMPLED_HOME" ]; then
    echo "Home directory $EXAMPLED_HOME does not exist"
    exit 1
fi

exec exampled --home "$EXAMPLED_HOME" "$@" 2>&1 | tee "$EXAMPLED_HOME/$LOG"
