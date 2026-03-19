#!/usr/bin/env bash
# Creates (or recreates) the tutorial/start branch by stripping all counter
# module files and reverting counter wiring from app.go.
#
# Run from the repo root on the main branch:
#   bash scripts/create-tutorial-branch.sh
#
# Then push:
#   git push -f origin tutorial/start

set -euo pipefail

BRANCH="tutorial/start"

echo "==> Checking working tree is clean..."
if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "ERROR: working tree has uncommitted changes. Commit or stash first."
  exit 1
fi

echo "==> Creating/resetting branch $BRANCH from current HEAD..."
git checkout -B "$BRANCH"

echo "==> Removing proto files..."
git rm -f \
  proto/example/counter/v1/state.proto \
  proto/example/counter/v1/tx.proto \
  proto/example/counter/v1/query.proto \
  proto/example/counter/v1/genesis.proto

echo "==> Removing generated pb.go files..."
git rm -f \
  x/counter/types/state.pb.go \
  x/counter/types/tx.pb.go \
  x/counter/types/query.pb.go \
  x/counter/types/query.pb.gw.go \
  x/counter/types/genesis.pb.go

echo "==> Removing hand-written types..."
git rm -f \
  x/counter/types/keys.go \
  x/counter/types/codec.go \
  x/counter/types/expected_keepers.go

echo "==> Removing keeper files..."
git rm -f \
  x/counter/keeper/errors.go \
  x/counter/keeper/telemetry.go \
  x/counter/keeper/keeper.go \
  x/counter/keeper/msg_server.go \
  x/counter/keeper/query_server.go \
  x/counter/keeper/keeper_test.go \
  x/counter/keeper/msg_server_test.go \
  x/counter/keeper/query_server_test.go

echo "==> Removing simulation files..."
git rm -f \
  x/counter/simulation/genesis.go \
  x/counter/simulation/msg_factory.go

echo "==> Removing module files..."
git rm -f \
  x/counter/module.go \
  x/counter/autocli.go

echo "==> Removing E2E test file..."
git rm -f \
  tests/counter_test.go

echo "==> Adding .gitkeep placeholders so tutorial directories exist on clone..."
mkdir -p x/counter
touch x/counter/.gitkeep
git add x/counter/.gitkeep

mkdir -p proto/example/counter/v1
touch proto/example/counter/v1/.gitkeep
git add proto/example/counter/v1/.gitkeep

echo "==> Stripping counter wiring from app.go..."
# perl -i is cross-platform (macOS + Linux)
perl -i -ne 'print unless /counter "github\.com\/cosmos\/example\/x\/counter"$/' app.go
perl -i -ne 'print unless /counterkeeper "github\.com\/cosmos\/example\/x\/counter\/keeper"$/' app.go
perl -i -ne 'print unless /countertypes "github\.com\/cosmos\/example\/x\/counter\/types"$/' app.go
perl -i -ne 'print unless /countertypes\.ModuleName:\s*nil/' app.go
perl -i -ne 'print unless /CounterKeeper\s*\*counterkeeper\.Keeper/' app.go
perl -i -ne 'print unless /countertypes\.StoreKey/' app.go
perl -i -ne 'print unless /app\.CounterKeeper = counterkeeper\.NewKeeper/' app.go
perl -i -ne 'print unless /counter\.NewAppModule\(appCodec, app\.CounterKeeper\)/' app.go
perl -i -ne 'print unless /countertypes\.ModuleName,/' app.go

git add app.go
git add docs/ scripts/create-tutorial-branch.sh

echo ""
echo "==> Verifying strip is complete..."
echo ""

ERRORS=0

# Ignore tutorial placeholder comments; check for actual counter code
if grep -v '// counter tutorial app wiring' app.go | grep -q "counter"; then
  echo "FAIL: app.go still contains counter code references:"
  grep -v '// counter tutorial app wiring' app.go | grep -n "counter"
  ERRORS=$((ERRORS + 1))
else
  echo "PASS: app.go has no counter code references"
fi

REMAINING=$(git ls-files x/counter/ proto/example/counter/ tests/counter_test.go 2>/dev/null \
  | grep -v '\.gitkeep' || true)
if [ -n "$REMAINING" ]; then
  echo "FAIL: these counter files still exist in the index:"
  echo "$REMAINING"
  ERRORS=$((ERRORS + 1))
else
  echo "PASS: all counter source files removed from index"
fi

if grep -rq 'counter.*"github.com/cosmos/example' tests/ 2>/dev/null; then
  echo "FAIL: tests/ still contains counter imports:"
  grep -rn 'counter.*"github.com/cosmos/example' tests/
  ERRORS=$((ERRORS + 1))
else
  echo "PASS: tests/ has no counter imports"
fi

echo "==> Running go build ./... to verify app compiles..."
if go build ./... 2>&1; then
  echo "PASS: go build ./... succeeded"
else
  echo "FAIL: go build ./... failed"
  ERRORS=$((ERRORS + 1))
fi

echo ""
if [ "$ERRORS" -eq 0 ]; then
  echo "==> All checks passed. Committing..."
  git commit -m "tutorial/start: strip counter module for tutorial [tutorial-sync]

Generated from cosmos/example main. Removes all counter module source
files and reverts app.go counter wiring. Adds .gitkeep placeholders so
tutorial directories exist on fresh clone. Docs are included so users
have the tutorial content on this branch."
  echo ""
  echo "Branch $BRANCH is ready. Push with:"
  echo "  git push -f origin $BRANCH"
else
  echo "==> $ERRORS check(s) failed."
  echo "    To abort: git checkout main && git branch -D $BRANCH"
  exit 1
fi
