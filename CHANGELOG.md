# Changelog

All notable changes to this repository are tracked here for agent context.

## [Unreleased]

### Tutorial Verification Pass

Ran `03-build-a-module.md` end to end from a clean `origin/tutorial/start` checkout, twice: once with the code blocks machine-extracted from the doc (to test the code) and once by an agent working only from the prose with no access to the finished module on `main` (to test the instructions). Both reached a working chain. `tx counter add 4` returned `code: 0` and `query counter count` returned `count: "4"`. No blocking failures, no step required debugging. Two doc fixes came out of it:

- `03-build-a-module.md` Step 12: replaced the single-line `code: 0` sample with the real 13-line broadcast response and explained it. The response is the mempool acknowledgement returned before inclusion in a block, so `height: "0"` and `gas_used: "0"` are expected. As written, a reader had good reason to read a successful transaction as a failure and start debugging it. Added a pointer to `exampled query tx <txhash>` for the executed result.
- `05-run-and-test.md` CLI reference: `exampled tx counter update-params --from alice` can never succeed. `MsgUpdateParams` only accepts the gov module address as authority, so signing with a user key always returns `ErrInvalidSigner`, and the command takes no params flags. Replaced it with an "Updating module parameters" section carrying a working `proposal.json`, the real `min_deposit`, the submit and vote commands, and the caveat that the default 48 hour `voting_period` means the change does not land during a normal dev session.

Deliberately not documented, though both were observed and confirmed: the `.gitkeep` files left in `x/counter/` and `proto/example/counter/v1/` on `tutorial/start` (Step 1 calls the directories empty, which is imprecise but harms nothing, and no step needs to remove them); and the proto package/directory mismatch, where the files declare `package example.counter` inside a `v1/` directory. The mismatch is real and load-bearing, since the AutoCLI service strings in Step 9 resolve against the package name rather than the path, and renaming the package makes every `exampled` command panic with `can't find service example.counter.Query: proto: not found`. It is left undocumented because a reader following the tutorial copies the blocks verbatim and never renames anything; the failure only surfaces when adapting the module for another chain, which is out of scope here.

Second pass, verifying the remaining docs against a live chain rather than by reading:

- `02-quickstart.md`: every documented output confirmed byte for byte on a fresh chain. `query counter count` returns `{}`, the params YAML matches, `tx counter add 5` then `query counter count` returns `count: "5"`. The 13-line broadcast response added to `03` Step 12 reproduced identically here, on a separate run and binary.
- `04-counter-walkthrough.md`: all 21 Go snippets checked against `x/counter/`. Every function signature matches the source verbatim. Prose claims in this file are still unverified.
- `05-run-and-test.md` test layers: `go test ./x/counter/...` passes, the targeted `-run TestKeeperTestSuite/TestAddCount` invocation runs all 9 subtests, `TestE2ETestSuite` passes 5/5 in 26s, and `make test-sim-full` passes all 38 seeds in 122s.
- `05-run-and-test.md` localnet: run end to end. All four nodes report `"n_peers":"3"` and advance together, exactly as documented. The `docker exec node0 ... tx counter add 7` command works as written and replicates: `node2` and `node3` both return `count: "7"`. Validator count is 1, confirming the "one validator plus three full nodes" note. `Dockerfile` correctly declares `ARG TARGETOS`/`ARG TARGETARCH` with no defaults, so nodes peer without the emulated-AVX2 handshake failure.
- `05-run-and-test.md` node config: all 9 values in the `app.toml`/`config.toml` tables confirmed against a live `~/.exampleapp`.

Corrections made to this pass's own first draft, both caught by testing claims instead of trusting them:

- The `05` gov section initially told readers to lower `voting_period` in `~/.exampleapp/config/genesis.json` before running `make start`. That does not work: `scripts/local_node.sh` line 7 deletes the entire home directory on every run, so the edit is destroyed. Replaced with the verified sequence (`make start` once, stop, edit genesis, `exampled comet unsafe-reset-all`, then `exampled start` directly). Confirmed end to end: proposal reached `PROPOSAL_STATUS_PASSED` and `query counter params` returned the new `add_cost: 200` / `max_add_value: 50`.
- The same section claimed `MsgUpdateParams` can be reached by choosing `other` in `draft-proposal` and searching by name. Only the verifiable part is now stated: the top-level list contains just `text`, `community-pool-spend`, `software-upgrade`, `cancel-software-upgrade`, and `other`, and `other` opens a scroll-only list of type URLs that typing does not filter. Whether the counter message appears in that list was not confirmed, so the doc no longer asserts it and points at hand-written JSON instead.

Still unverified after this pass: the prose claims (not code) in `04-counter-walkthrough.md`, and how the new `<Note>` and JSON blocks render on the live Mintlify site.

Known issues found and deliberately left alone in this pass: the doc's Go blocks use 4-space indentation, so all seven files a reader creates fail `gofmt -l` (`make lint` passes anyway, 0 issues, since gofmt is not in the golangci config); the seven `app.go` wiring blocks are flush-left and paste misaligned into indented context, and each includes its own marker comment as line 1, so a select-all paste duplicates the marker; `NewKeeper` takes a `cdc codec.Codec` the minimal module never uses, kept for signature parity with `main`; Step 12 gives no guidance to wait for the first block before submitting; `make proto-image-build` cold-builds Docker layers for several minutes with no warning that this is expected.

### Docs Updates for SDK v0.55 Upgrade

- `01-prerequisites.md`: bumped required Go version from 1.25 to 1.26 to match `go.mod` (`go 1.26.5`) after the SDK/CometBFT upgrade. This was the only doc breakage caused by the upgrade itself; no doc references `traceStore`, the `cosmossdk.io/store` → `cosmos-sdk/store/v2` move, or the dropped legacy-subspace args.
- `03-build-a-module.md`: fixed the keeper field snippet in Step 10.2 to match the gofmt'd spacing in `app.go`.

### Tutorial Wiring Fix

- `app.go`: moved the `counter tutorial app wiring 6` and `7` marker comments out of `SetOrderBeginBlockers` / `SetOrderEndBlockers` and into `genesisModuleOrder` / `exportModuleOrder`, where the doc headings ("Genesis Order", "Export Order") always said they were. Previously a reader following `03-build-a-module.md` never added the counter module to the genesis or export lists and hit `panic: all modules must be defined when setting SetOrderInitGenesis, missing: [counter]` at startup. The minimal tutorial module has no block hooks, so it does not belong in the blocker lists; the full module on `main` is still listed there. Verified by regenerating `tutorial/start`, following the tutorial verbatim, and running the chain: `tx counter add 4` returns `code: 0` and `query counter count` returns `count: "4"`.

### Overflow Check Fix

- `x/counter/keeper/keeper.go`: replaced the guard in `AddCount`. It read `if amount >= math.MaxUint64`, which for a `uint64` can only ever be true at exactly `MaxUint64` (staticcheck SA4003) and, more importantly, tested the wrong quantity: overflow occurs at `count + amount`, which Go wraps silently, leaving the counter holding a smaller value with no error. It now checks `amount > math.MaxUint64-count`, written as a subtraction so the comparison cannot overflow. Reachable in practice because `MaxAddValue = 0` disables the per-add cap, letting a single add wrap the counter.
- Moved the `GetCount` read above the fee charge so all validation completes before the user is charged.
- Added two `TestAddCount` cases: an add that would wrap with `MaxAddValue` disabled (must error, state unchanged) and an add landing exactly on `MaxUint64` (must succeed). The first fails against the previous implementation, which wrapped the counter to `0` and returned no error.
- `04-counter-walkthrough.md`: updated the `AddCount` snippet and added a note on why the guard tests the result rather than the input. This also makes the feature table's "overflow check" claim true.

### Error Code Fix

- `x/counter/keeper/errors.go`: changed the registered sentinel error codes from `0, 1, 2` to `2, 3, 4`. Code `0` is the ABCI success code, so a transaction failing with `ErrNumTooLarge` reported `code: 0` and read as a success to every client. Code `1` is reserved for internal errors.
- `04-counter-walkthrough.md`: updated the `errors.go` snippet and corrected the surrounding prose, which claimed codes must be greater than zero while the code registered one at zero.

### Docs Accuracy Fixes

All claims below were checked against a running chain, not against library default constants. Values the application overrides differ from the upstream defaults, so reading the SDK or CometBFT source alone gives the wrong answer.

- `04-counter-walkthrough.md`: updated the `UpdateParams` snippet to match the explicit error check in `msg_server.go`; documented the real `NewKeeper(storeService, cdc, bankKeeper, opts ...Options)` signature and the `WithAuthority` option, which the authority section had never mentioned.
- `04-counter-walkthrough.md`: the Gas section claimed `make start` "leaves `minimum-gas-prices` empty". It is set to `0stake` by `initAppConfig` in `exampled/cmd/commands.go`.
- `02-quickstart.md`: fixed the indentation of the `query counter params` sample output. The doc indented the `add_cost` list item by four spaces; the real YAML output puts the dash at two.
- `05-run-and-test.md`: annotated `minimum-gas-prices` (`0stake`, set by this chain, not the SDK, whose default is empty) and `timeout_commit` (`5s`, the SDK raises CometBFT's own `1s` default in `server/util.go`). Both table values were already correct and are unchanged.
- `05-run-and-test.md`: `api.enable` is `false` by SDK default and only reads `true` because `scripts/local_node.sh` sets it. The table now says so rather than presenting `true` as the stock default.

### Simulation

- `Makefile`: the three `test-sim*` targets passed no simulation flags, so each ran the SDK defaults of 500 blocks and 200 operations per block across the 38 seeds hardcoded in `simsx.Run`. None could finish inside their own timeout: `make test-sim-full` died with `panic: test timed out after 30m0s` and `make test-sim` with `test timed out` after 60m. All three are documented in `05-run-and-test.md` as validation commands, so all three were failing for anyone who ran them.
- Added overridable `SIM_NUM_BLOCKS` (50), `SIM_BLOCK_SIZE` (100), and `SIM_TIMEOUT` (30m). Measured after the change: `make test-sim-determinism` passes in 551s and `make test-sim-full` in 663s.
- `05-run-and-test.md`: documented the ~10 minute runtime per target, the 38-seed sweep, and how to override the parameters for a deeper simulation.

### Localnet

- `Dockerfile`: bumped the build stage from `golang:1.25-alpine` to `golang:1.26-alpine`. The SDK upgrade raised `go.mod` to `go 1.26.5` without touching the Dockerfile, so `make localnet-init` failed outright with `go: go.mod requires go >= 1.26.5 (running go 1.25.12)`. Verified by a clean image build.
- `05-run-and-test.md`: rewrote the localnet section. It is one validator plus three full nodes, not four validators, because `init.sh` creates a single gentx. Added the Docker prerequisite, the 16 host ports required, the multi-minute cold start, the fact that `localnet-logs` follows and needs Ctrl+C, and that `localnet-clean` deletes without confirming. Also documented that the chain ID is `example-localnet` with a single `validator` key, not `demo` with `alice`/`bob`.
- `Dockerfile`: removed the `=linux` and `=amd64` defaults from `ARG TARGETOS` / `ARG TARGETARCH`. This is the root cause of the long-standing localnet P2P failure. The hardcoded default beat BuildKit's platform value, so on an arm64 host `make build-docker` cross-compiled an **amd64 binary into an arm64 image**, which Docker then ran under x86 emulation.

  The emulator advertises AVX2 and BMI2, so `golang.org/x/crypto/chacha20poly1305` dispatched to its AVX2 assembly path, which the emulator mis-executes. An RFC 8439 known-answer test fails under emulation and passes natively, and `Seal` and `Open` go wrong differently, so even two identical emulated binaries cannot talk to each other. The AVX2 bulk path engages above 320 bytes and CometBFT's P2P frame is a fixed 1028 bytes, which is why every handshake failed, in both directions, 100% of the time, with `chacha20poly1305: message authentication failed`. X25519 is unaffected, which is why TCP, protobuf framing, ed25519, and single-node block production all worked and masked the problem.

  Verified end to end: all four nodes reach `n_peers=3` and advance in lockstep, zero `chacha20poly1305` errors, and a counter transaction submitted on `node0` replicates to `node2`.

Remaining localnet hygiene, not required for it to work and confirmed to predate this upgrade:

- `scripts/localnet/init.sh` sets `persistent_peers` before running the genesis commands, and `collect-gentxs` then rewrites `node0`'s `config.toml` and blanks the field. Harmless in practice: `node0` still reaches three peers via inbound dials.
- `addr_book_strict` is left at `true`. It does not in fact reject the compose network's `192.168.10.0/25` addresses.

### Lint

- `make lint` is documented in `05-run-and-test.md` as a validation command but failed on the repo's own code. The `staticcheck` SA4003 finding was the ineffective overflow guard, fixed above. The three remaining `errcheck` findings are now handled with explicit discards: `fmt.Fprintln` in `exampled/main.go`, `s.conn.Close` in `tests/counter_test.go`, and the deferred `os.RemoveAll` in `tests/test_helpers.go`. `make lint` now reports `0 issues`.

### Build

- `Makefile`: added `-ldflags` injecting `version.Name`, `AppName`, `Version`, and `Commit` into `build` and `install`. Previously `make install` was a bare `go install`, so `exampled version` printed an empty line despite `02-quickstart.md` telling readers to run it to verify the install. It now prints the `git describe` version.

### Docs Sync Fixes

- Fixed `git diff --quiet` in both sync workflows — it missed untracked files (new tutorial pages would silently skip sync). Replaced with `git status --porcelain <dir>` which catches new, modified, and deleted files.
- Moved `transform.py` into `cosmos/docs` repo (`scripts/docs-sync/transform.py`) so the workflow no longer copies and executes an unverified script from an external repo; changes to the script now go through PR review on the docs repo.

### Tutorial Branch Automation

- Added `scripts/create-tutorial-branch.sh` to `main` (previously only on `tutorial/start`); updated to use `git checkout -B` for existing branch and cross-platform `perl -i` instead of `sed -i ''`
- Added `.github/workflows/update-tutorial-branch.yml` — regenerates `tutorial/start` automatically on every push to `main`; skips `[tutorial-sync]` and `[docs-sync]` commits to prevent loops
- Updated `CLAUDE.md` to document the automation

### Docs Changes

- Renamed tutorial docs to `NN-name.md` format (`01-prerequisites.md` through `05-run-and-test.md`), added `00-overview.md` intro page
- `02-quickstart.md`: rewrote opening paragraph, added Mintlify Note callout linking to prerequisites
- `03-build-a-module.md`: replaced blockquote prerequisite notice with Mintlify Note callout; added links to SDK concept docs throughout (modules, transactions, encoding, keeper, app.go, etc.)
- `04-counter-walkthrough.md`: added anchor links in feature comparison table; added Gas section covering `minimum-gas-prices` in `app.toml`; linked repo in branch switch instruction; removed stale to-do comment
- `05-run-and-test.md`: added Node Configuration section covering `app.toml` and `config.toml` with key settings tables
- `01-prerequisites.md`: updated Go version output to show both Linux and macOS variants; updated Make install instructions to cover both platforms
- Verified `make install` builds successfully on Linux via Docker `golang:1.25` container

### CLAUDE.md Changes

- Added Changelog Policy section requiring changelog updates after every change

---

## [2026-03-18] — Docs sync system + file renames

Branch: `feat/docs-sync` (in `cosmos/example`), `feat/example-tutorial-sync` (in `cosmos/docs`)

### example repo changes
- Renamed all 5 tutorial docs to drop the `tutorial-NN-` prefix:
  - `tutorial-00-prerequisites.md` → `prerequisites.md`
  - `tutorial-01-quickstart.md` → `quickstart.md`
  - `tutorial-02-build-a-module.md` → `build-a-module.md`
  - `tutorial-03-counter-walkthrough.md` → `counter-walkthrough.md`
  - `tutorial-04-run-and-test.md` → `run-and-test.md`
- Updated all internal cross-links between tutorial files to use new names
- Added `scripts/docs-sync/transform.py` — bidirectional transform between `.md` and Mintlify `.mdx` format (H1 ↔ frontmatter, absolute ↔ relative links, `.md` ↔ `.mdx` extensions)
- Added `scripts/docs-sync/test_transform.py` — 25 unit + round-trip tests (all passing)
- Added `.github/workflows/docs-sync.yml` — GitHub Action that auto-opens PRs on `cosmos/docs` when `docs/**` changes on `main`

### cosmos/docs repo changes
- Fixed typo: `prerequisits.mdx` → `prerequisites.mdx`
- Added 5 new `.mdx` files in `sdk/next/tutorials/example/` (transformed from `docs/*.md`)
- Added `Build a Chain` group to `docs.json` navigation under `sdk/next` How-to Guides
- Added `.github/workflows/docs-sync-to-example.yml` — reverse sync action that opens PRs on `cosmos/example` when tutorial `.mdx` files change

### Setup required (one-time)
- Add secret `DOCS_REPO_TOKEN` to `cosmos/example` (PAT: `contents:write` + `pull-requests:write` on `cosmos/docs`)
- Add secret `EXAMPLE_REPO_TOKEN` to `cosmos/docs` (PAT: `contents:write` + `pull-requests:write` on `cosmos/example`)

---

## [2026-03-18] — Initial setup

- Added `CLAUDE.md` to document repo purpose, branch strategy, docs policy, and module architecture for future agents.
- Added `CHANGELOG.md` (this file) to track changes over time.

---

## Format

Each entry should follow:

```
## [YYYY-MM-DD] — Short description

- What changed and why
- Docs site impact (if any)
- Branch(es) affected
```
