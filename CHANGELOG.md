# Changelog

All notable changes to this repository are tracked here for agent context.

## [Unreleased]

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
