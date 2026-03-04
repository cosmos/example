# Prerequisites

Before starting the tutorial, make sure you have the following tools installed.

## Go

The example chain requires Go 1.22 or higher.

```bash
go version
# go version go1.22.0 darwin/arm64
```

If Go is not installed, download it from [go.dev/dl](https://go.dev/dl).

## Make

Make is used to run build and development commands throughout the tutorial.

```bash
make --version
# GNU Make 3.81
```

Make is pre-installed on most Linux and macOS systems. On macOS, if it is missing, install it with Xcode command line tools:

```bash
xcode-select --install
```

## Docker

Docker is required to run `make proto-gen`, which generates Go code from the module's proto files using [buf](https://buf.build).

```bash
docker --version
# Docker version 24.0.0
```

Download Docker from [docs.docker.com/get-docker](https://docs.docker.com/get-docker).

Docker must be running before you execute `make proto-gen`.

## Git

```bash
git --version
# git version 2.39.0
```

## Clone the repository

```bash
git clone https://github.com/cosmos/example
cd example
```

---

Next: [Quickstart →](./tutorial-01-quickstart.md)
