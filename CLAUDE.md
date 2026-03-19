# CLAUDE.md — Agent Context for `example` Repo

## What This Repo Is

This is the **Cosmos SDK example application** — a reference implementation showing developers how to build a Cosmos SDK module from scratch and wire it into a chain. It lives at `github.com/cosmos/example`.

The primary audience is Cosmos SDK developers learning module development. The docs are tutorial-style and are published to the **Cosmos SDK documentation site**.

---

## Branch Strategy

| Branch | Purpose |
|---|---|
| `main` | Complete, fully-featured counter module with all bells and whistles |
| `tutorial/start` | Blank template — users follow tutorial docs to build the module from scratch |

**Rule:** `main` always contains the finished module. `tutorial/start` contains the starting scaffolding users begin from. The tutorials in `docs/` walk users from `tutorial/start` → `main`.

---

## Docs Policy

**Critical:** All docs live in `/docs/` and are published to the Cosmos SDK docs site. **Any change to a doc file here must also be reflected on the Cosmos docs site.** Always flag this when making doc changes.

### Doc Files

| File | Content |
|---|---|
| `tutorial-00-prerequisites.md` | Prerequisites: Go, Make, Docker, Git; repo layout overview |
| `tutorial-01-quickstart.md` | Fast path: build, install, run chain, send a tx |
| `tutorial-02-build-a-module.md` | Step-by-step: build a minimal counter module from scratch |
| `tutorial-03-counter-walkthrough.md` | Walk through full module (main branch) — params, fees, auth, errors, telemetry, sim |
| `tutorial-04-run-and-test.md` | Running a local chain, localnet, CLI reference, all test layers |

---

## Module Architecture

The custom module is `x/counter`. The full build pattern is:

```
proto files → make proto-gen → generated types → keeper → msg/query servers → module.go → app.go wiring
```

### Key Layers

- **`proto/example/counter/v1/`** — Protobuf definitions (tx, query, state, genesis)
- **`x/counter/types/`** — Generated + hand-written types (keys, codec, expected keepers)
- **`x/counter/keeper/`** — Keeper (state), MsgServer (txs), QueryServer (queries), telemetry, errors
- **`x/counter/module.go`** — AppModule implementation (genesis, services, simulation, block hooks)
- **`x/counter/autocli.go`** — Auto-generates CLI from proto
- **`app.go`** — Wires counter module into the chain

### Full Module Features (main branch)

- State: `count` (uint64) + `params` (MaxAddValue, AddCost)
- Messages: `MsgAdd`, `MsgUpdateParams` (governance-gated)
- Queries: `Count`, `Params`
- Validation: MaxAddValue limit, overflow protection
- Fees: AddCost charged via bank keeper
- Authority: governance module controls param updates
- Errors: named sentinel errors with codes
- Telemetry: OpenTelemetry Int64Counter metric
- Simulation: simsx weighted operations + randomized genesis
- Tests: unit (keeper, msg_server, query_server), E2E, simulation

---

## Docs Sync System

Docs in `docs/` are kept in sync with the Cosmos docs site repo (`cosmos/docs`) via GitHub Actions + a transform script.

### File mapping

| `example` repo | Docs site |
|---|---|
| `docs/prerequisites.md` | `sdk/next/tutorials/example/prerequisites.mdx` |
| `docs/quickstart.md` | `sdk/next/tutorials/example/quickstart.mdx` |
| `docs/build-a-module.md` | `sdk/next/tutorials/example/build-a-module.mdx` |
| `docs/counter-walkthrough.md` | `sdk/next/tutorials/example/counter-walkthrough.mdx` |
| `docs/run-and-test.md` | `sdk/next/tutorials/example/run-and-test.mdx` |

### Format differences

| | `example` repo | Docs site (Mintlify) |
|---|---|---|
| Title | `# H1 heading` | YAML frontmatter `title:` + `noindex: true` |
| Docs links | `https://docs.cosmos.network/sdk/next/...` | `sdk/next/...` (relative) |
| File extension | `.md` | `.mdx` |

### Transform script

`scripts/docs-sync/transform.py` handles all conversion in both directions:

```bash
# example → docs site
python3 scripts/docs-sync/transform.py --direction to-mintlify --input docs/ --output-dir /path/to/docs-site/sdk/next/tutorials/example/

# docs site → example
python3 scripts/docs-sync/transform.py --direction to-example --input /path/to/docs-site/sdk/next/tutorials/example/ --output-dir docs/

# Run tests
python3 scripts/docs-sync/test_transform.py
```

### GitHub Actions

- **`example` → docs site:** `.github/workflows/docs-sync.yml` — triggers on push to `main` when `docs/**` changes, opens a PR on `cosmos/docs`. Requires secret `DOCS_REPO_TOKEN`.
- **docs site → `example`:** `.github/workflows/docs-sync-to-example.yml` in the docs repo — triggers on push to `main` when `sdk/next/tutorials/example/**` changes, opens a PR on `cosmos/example`. Requires secret `EXAMPLE_REPO_TOKEN`.

### Loop prevention

Both actions check `!contains(github.event.head_commit.message, '[docs-sync]')` and tag sync commits with `[docs-sync]` to prevent infinite loops.

### Required secrets to set up

- In `cosmos/example`: secret `DOCS_REPO_TOKEN` — fine-grained PAT with `contents:write` + `pull-requests:write` on `cosmos/docs`
- In `cosmos/docs`: secret `EXAMPLE_REPO_TOKEN` — fine-grained PAT with `contents:write` + `pull-requests:write` on `cosmos/example`

---

## Generating `tutorial/start` from `main`

The `tutorial/start` branch is **not manually maintained** — it is generated from `main` using:

```bash
bash scripts/create-tutorial-branch.sh
```

This script (lives on `tutorial/start` branch at `scripts/create-tutorial-branch.sh`):
1. Creates the `tutorial/start` branch from `main`
2. Removes all counter module source files (`x/counter/`, proto files, generated `.pb.go`, tests)
3. Adds `.gitkeep` placeholders so `x/counter/` and `proto/example/counter/v1/` exist on clone
4. Strips counter wiring from `app.go` (imports, keeper field, store key, module registration, etc.)
5. Verifies `go build ./...` still passes
6. Commits the result

**Rule:** When `main` changes significantly (new features, restructured files), re-run this script to regenerate `tutorial/start`. Don't hand-edit `tutorial/start`.

---

## App Wiring (app.go)

Counter module integration points in `app.go`:

- `maccPerms` — counter module account entry (for fee collection)
- `ExampleApp.CounterKeeper` field
- Counter store key in store keys
- `NewKeeper(...)` with bank keeper + authority
- `ModuleManager` registration
- Genesis, export, begin/end blocker order lists

---

## Testing

```bash
go test ./x/counter/...         # Unit tests (fast, no chain)
go test -v -run TestE2ETestSuite ./tests/...  # E2E (starts in-process network)
make test-sim-full              # Simulation (randomized txs)
make lint                       # Lint
```

---

## Development Scripts

```bash
make build          # Compile binary
make install        # Install exampled binary
make start          # Build + start single-node chain (Chain ID: demo, accounts: alice/bob, denom: stake)
make proto-gen      # Regenerate proto (requires Docker)
make localnet-init  # Initialize multi-validator localnet
make localnet-start # Start localnet
```

---

## Changelog Policy

**After every change to this repo, update `CHANGELOG.md`.** Add an entry under `## [Unreleased]` describing what changed and why. When changes are committed, move the unreleased items into a dated entry.

---

## Writing Style

- **No em dashes.** Use a period, comma, or colon instead.

---

## Key Constants

- Binary: `exampled`
- Chain ID (dev): `demo`
- Accounts: `alice`, `bob`
- Denom: `stake`
- Module name: `counter`
- Default params: `add_cost = 100stake`, `max_add_value = 100`

---

## Go Dependencies

- `github.com/cosmos/cosmos-sdk v0.54.0-rc.1`
- `cosmossdk.io/core v1.1.0`
- `cosmossdk.io/collections v1.4.0`
- `github.com/cometbft/cometbft v0.39.0-beta.2`
- Go 1.25.7
