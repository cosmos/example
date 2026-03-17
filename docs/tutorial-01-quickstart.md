# Quickstart

`exampled` is a simple Cosmos SDK chain that shows the core pieces of a working app chain. It includes the basic building-block modules for accounts, bank, staking, distribution, slashing, governance, and more.

This quickstart gets you running `exampled`, submitting a transaction, and querying the result as quickly as possible. It also includes the `x/counter` module, which stores a single counter value, lets you query the current count, and lets you submit `Add` transactions to increment it. In the next tutorials, you'll build a simple version of this module yourself and then walk through the full implementation and its additional features.

## Install the binary

Run the following to compile the `exampled` binary and place it on your `$PATH`.

```bash
make install
```

Verify the install by running:

```bash
exampled version
```

You can also run the following to see all available node CLI commands:

```bash
exampled
```

## Start the chain

Run the following to start a single-node local chain. It handles all setup automatically: initializes the chain data, creates test accounts, and starts the node. Leave it running in this terminal.

```bash
make start
```

## Query the counter

Open a second terminal and query the current count:

```bash
exampled query counter count
```

You should see the following output, which means the counter is starting at `0`:


```text
{}
```

You can also query the module parameters:

```bash
exampled query counter params
```

This shows that the fee to increment the counter is stored as a module parameter. The base coin denomination for the `exampled` chain is `stake`.

```yaml
params:
  add_cost:
    - amount: "100"
      denom: stake
  max_add_value: "100"
```

## Submit an add transaction

Send an `Add` transaction to increment the counter. This charges a fee from the funded `alice` account you are sending the transaction from:

```bash
exampled tx counter add 5 --from alice --chain-id demo --yes
```

## Query the counter again

After submitting the transaction, query the counter again to see the updated module state:

```bash
exampled query counter count
```

You should see the following:

```
count: "5"
```

Congratulations! You just ran a blockchain, submitted a transaction, and queried module state.

## Next steps

In the following tutorials, you will:

1. Build a minimal version of this module from scratch to understand the core pattern
2. Walk through the full `x/counter` module example to see what it adds
3. See how modules are wired into a chain and how to run the full test suite


Next: [Build a Module from Scratch →](./tutorial-02-build-a-module.md)
