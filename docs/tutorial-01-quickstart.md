# Quickstart

This page gets the example chain running in a few minutes so you can see the counter module working before building it yourself.

## Install the binary

```bash
make install
```

This compiles the `exampled` binary and places it on your `$PATH`.

Verify:

```bash
exampled version
```

## Start the chain

```bash
make start
```

This starts a single-node local chain. It handles all setup automatically: initializes the chain data, creates test accounts, and starts the node. Leave it running in this terminal.

## Query the counter

Open a second terminal and query the current count:

```bash
exampled query counter count
```

```
count: "0"
```

Query the module parameters:

```bash
exampled query counter params
```

```yaml
params:
  add_cost:
    - amount: "100"
      denom: stake
  max_add_value: "100"
```

## Submit an add transaction

The `alice` account is funded at chain start. Send an add transaction:

```bash
exampled tx counter add 5 --from alice --chain-id demo --yes
```

## Query the counter again

```bash
exampled query counter count
```

```
count: "5"
```

The counter incremented by 5.

## What you just ran

The chain is running the `x/counter` module — the production counter module in this repository. In the following tutorials you will:

1. Build a minimal version of this module from scratch to understand the core pattern
2. Walk through the production `x/counter` to see what it adds
3. See how modules are wired into a chain and how to run the full test suite

---

Next: [Build a Module from Scratch →](./tutorial-02-build-a-module.md)
