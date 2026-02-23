#!/usr/bin/env bash

EXAMPLED_BIN=${EXAMPLED_BIN:=$(which exampled 2>/dev/null)}

if [ -z "$EXAMPLED_BIN" ]; then echo "EXAMPLED_BIN is not set. Make sure to run make install before"; exit 1; fi
echo "using $EXAMPLED_BIN"
if [ -d "$($EXAMPLED_BIN config home)" ]; then rm -rv $($EXAMPLED_BIN config home); fi
$EXAMPLED_BIN config set client chain-id demo
$EXAMPLED_BIN config set client keyring-backend test
$EXAMPLED_BIN config set client keyring-default-keyname alice
$EXAMPLED_BIN config set app api.enable true
$EXAMPLED_BIN keys add alice
$EXAMPLED_BIN keys add bob
$EXAMPLED_BIN init test --chain-id demo
$EXAMPLED_BIN genesis add-genesis-account alice 5000000000stake --keyring-backend test
$EXAMPLED_BIN genesis add-genesis-account bob 5000000000stake --keyring-backend test
$EXAMPLED_BIN genesis gentx alice 1000000stake --chain-id demo
$EXAMPLED_BIN genesis collect-gentxs