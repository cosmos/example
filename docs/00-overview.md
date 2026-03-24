# Build a Chain

The Cosmos SDK is a developer-first framework for building custom blockchains. This tutorial series shows you how to build a module from scratch, wire it into a chain, and run it locally, all in minutes.

By the end, you will have:

- A working Cosmos SDK chain running on your machine
- A custom module you built yourself, wired into the chain
- A clear mental model of how modules, keepers, messages, and queries fit together

This series starts from zero; you don't need any prior Cosmos SDK experience to follow along.

## The example repo

All tutorials in this series are based on [cosmos/example](https://github.com/cosmos/example), a reference Cosmos SDK chain built around a custom `x/counter` module.

The repo has two main branches:

- `main`: the complete chain with the full `x/counter` module wired in. This is used in the [Quickstart guide](./02-quickstart.md).
- `tutorial/start`: the same chain without the counter module. The `x/counter` directory and its app wiring are stripped out so you can build the module from scratch by [following the tutorial](./03-build-a-module.md).

If you want to follow along and build the module yourself, start from `tutorial/start`. If you want to browse the finished implementation first, use `main`.

## What's in this series

1. [Prerequisites](./01-prerequisites.md): Install Go, Make, Docker, and Git. Clone the repo and get familiar with the layout.

2. [Quickstart](./02-quickstart.md): Build and run the chain in minutes. Submit a transaction, query the result, and see the counter module in action before you build it yourself.

3. [Build a Module from Scratch](./03-build-a-module.md): Build a minimal counter module step by step: proto definitions, keeper, message server, query server, and app wiring. Start here if you want to understand how a module comes together.

4. [Full Module Walkthrough](./04-counter-walkthrough.md): Walk through the complete `x/counter` implementation on `main`. Covers everything added on top of the minimal module: params, governance-gated authority, validation, fees, sentinel errors, telemetry, AutoCLI, simulation, block hooks, and a full unit test suite.

5. [Run and Test](./05-run-and-test.md): Learn the full development workflow: running a local chain, using the CLI, and working with the three layers of testing: unit tests, end-to-end tests, and simulation.
