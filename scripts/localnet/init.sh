#!/bin/bash
set -euo pipefail

NUM_NODES=4
CHAIN_ID="example-localnet"
STAKE_AMOUNT="1000000000stake"
LOCALNET_DIR="./build/localnet"
DOCKER_IMAGE="example-node"

echo "==> Cleaning up previous localnet..."
rm -rf "$LOCALNET_DIR"
mkdir -p "$LOCALNET_DIR"

echo "==> Initializing $NUM_NODES nodes..."
for i in $(seq 0 $((NUM_NODES - 1))); do
    NODE_DIR="$LOCALNET_DIR/node$i"
    mkdir -p "$NODE_DIR"
    
    docker run --rm -v "$PWD/$NODE_DIR:/root/.exampled" $DOCKER_IMAGE \
        exampled init "node$i" --chain-id "$CHAIN_ID" --home /root/.exampled
    
    docker run --rm -v "$PWD/$NODE_DIR:/root/.exampled" $DOCKER_IMAGE \
        exampled keys add validator --keyring-backend test --home /root/.exampled
done

echo "==> Fetching validator addresses..."
declare -a ACCOUNTS
for i in $(seq 0 $((NUM_NODES - 1))); do
    NODE_DIR="$LOCALNET_DIR/node$i"
    ACCOUNTS[$i]=$(docker run --rm -v "$PWD/$NODE_DIR:/root/.exampled" $DOCKER_IMAGE \
        exampled keys show validator -a --keyring-backend test --home /root/.exampled | tr -d '\r')
    echo "Node $i address: ${ACCOUNTS[$i]}"
done

echo "==> Fetching node IDs..."
declare -a NODE_IDS
for i in $(seq 0 $((NUM_NODES - 1))); do
    NODE_DIR="$LOCALNET_DIR/node$i"
    NODE_IDS[$i]=$(docker run --rm -v "$PWD/$NODE_DIR:/root/.exampled" $DOCKER_IMAGE \
        exampled comet show-node-id --home /root/.exampled | tr -d '\r')
    echo "Node $i ID: ${NODE_IDS[$i]}"
done

echo "==> Setting up persistent peers..."
for i in $(seq 0 $((NUM_NODES - 1))); do
    PEERS=""
    for j in $(seq 0 $((NUM_NODES - 1))); do
        if [ "$i" -ne "$j" ]; then
            IP="192.168.10.$((j + 2))"
            PEERS+="${NODE_IDS[$j]}@$IP:26656,"
        fi
    done
    PEERS=${PEERS%,}
    
    CONFIG_FILE="$LOCALNET_DIR/node$i/config/config.toml"
    echo "Node $i peers: $PEERS"
    sed -i.bak "s|^persistent_peers *=.*|persistent_peers = \"$PEERS\"|" "$CONFIG_FILE"
done

echo "==> Adding genesis accounts on node0..."
NODE0_DIR="$LOCALNET_DIR/node0"
for i in $(seq 0 $((NUM_NODES - 1))); do
    docker run --rm -v "$PWD/$NODE0_DIR:/root/.exampled" $DOCKER_IMAGE \
        exampled genesis add-genesis-account "${ACCOUNTS[$i]}" "$STAKE_AMOUNT" --home /root/.exampled
done

echo "==> Creating gentx for node0..."
docker run --rm -v "$PWD/$NODE0_DIR:/root/.exampled" $DOCKER_IMAGE \
    exampled genesis gentx validator 100000000stake --chain-id "$CHAIN_ID" --keyring-backend test --home /root/.exampled

echo "==> Collecting gentxs..."
docker run --rm -v "$PWD/$NODE0_DIR:/root/.exampled" $DOCKER_IMAGE \
    exampled genesis collect-gentxs --home /root/.exampled

echo "==> Distributing genesis.json to all nodes..."
for i in $(seq 1 $((NUM_NODES - 1))); do
    cp "$NODE0_DIR/config/genesis.json" "$LOCALNET_DIR/node$i/config/genesis.json"
done

echo "==> Enabling API and gRPC on all nodes..."
for i in $(seq 0 $((NUM_NODES - 1))); do
    APP_FILE="$LOCALNET_DIR/node$i/config/app.toml"
    sed -i.bak 's|^enable = false|enable = true|' "$APP_FILE"
    sed -i.bak 's|^address = "tcp://localhost:1317"|address = "tcp://0.0.0.0:1317"|' "$APP_FILE"
    sed -i.bak 's|^address = "localhost:9090"|address = "0.0.0.0:9090"|' "$APP_FILE"
done

echo "==> Configuring RPC to listen on all interfaces..."
for i in $(seq 0 $((NUM_NODES - 1))); do
    CONFIG_FILE="$LOCALNET_DIR/node$i/config/config.toml"
    sed -i.bak 's|^laddr = "tcp://127.0.0.1:26657"|laddr = "tcp://0.0.0.0:26657"|' "$CONFIG_FILE"
done

echo "==> Cleanup backup files..."
find "$LOCALNET_DIR" -name "*.bak" -delete

echo "==> Localnet initialization complete!"
echo ""
echo "Start the network with: make localnet-start"
echo "Stop the network with:  make localnet-stop"
echo ""
echo "Node 0 RPC:  http://localhost:26657"
echo "Node 0 API:  http://localhost:1317"
echo "Node 0 gRPC: localhost:9090"
