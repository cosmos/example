# Prerequisites

Before starting the tutorial, make sure you have the following tools installed.

<Warning>
This tutorial is intended for macOS and Linux systems. Other systems may have additional requirements.
</Warning>

## Go

The example chain requires Go 1.25 or higher. 

```bash
go version
# go version go1.25.0 linux/amd64   # Linux
# go version go1.25.0 darwin/arm64  # macOS
```

If Go is not installed, download it from [go.dev/dl](https://go.dev/dl).

### Configure Go Environment Variables

After installing Go, make sure `$GOPATH/bin` is on your `PATH` so installed binaries (like `exampled`) are accessible.

Open your shell config file (`~/.zshrc` on macOS or `~/.bashrc` on Linux) and add:

```bash
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

Then apply the changes:

```bash
source ~/.zshrc    # macOS
source ~/.bashrc   # Linux
```

Verify: `go env GOPATH`

## Make

Make is used to run build and development commands throughout this tutorial.

```bash
make --version
# GNU Make 3.81
```

Make is pre-installed on most Linux and macOS systems. If it is missing:

- **macOS:** `xcode-select --install`
- **Linux (Debian/Ubuntu):** `sudo apt install build-essential`

## Docker

Docker is required to run `make proto-gen`, which generates Go code from the module's proto files using [buf](https://buf.build).

```bash
docker --version
# Docker version 29.2.1
```

Download Docker from [docs.docker.com/get-docker](https://docs.docker.com/get-docker).

Docker must be running before you execute `make proto-gen`.

## Git

```bash
git --version
# git version 2.52.0
```

## Clone the repository

Clone [cosmos/example](https://github.com/cosmos/example) and navigate into it. 

```bash
git clone https://github.com/cosmos/example
cd example
```

The repo has two branches used in this tutorial series:

- `main` — the complete chain with the full `x/counter` module wired in.
- `tutorial/start` — the same chain with the counter module stripped out. Start here if you want to build the module yourself from scratch.

## Repository Layout

After cloning, the repository looks like this:

```text
example/
├── exampled/         # Binary entrypoint (main.go + CLI root command)
├── app.go            # Chain application, module wiring lives here
├── proto/            # Proto definitions for all modules
├── x/                # Module implementations
│   └── counter/      # The example counter module
├── tests/            # E2E and integration tests
├── scripts/          # Local node and proto generation scripts
├── docs/             # This tutorial series
└── Makefile          # Build, test, and dev commands
```

## Where things live

The tutorials in this section will walk you through the most common kinds of chain changes and show you where they usually live in the repo:

- Add or modify a module: `x/<module>/` and `proto/`
- Wire a module into the chain: `app.go`
- Change the binary or CLI: `exampled/`
- Run the chain or tests: `Makefile` targets

---

Next: [Quickstart →](./02-quickstart.md)
