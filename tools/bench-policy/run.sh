#!/usr/bin/env bash
#
# Measure what a permission check costs, here and on the commit this branch is
# based on.
#
#   tools/bench-policy/run.sh [base-commit] [-benchtime 20000x]
#
# It checks the base commit out into a temporary worktree, copies the same
# benchmarks in, and runs both. benchstat is used for the comparison when it is
# installed; without it the two raw results are printed.
set -euo pipefail

cd "$(dirname "$0")/../.."
readonly root=$PWD
readonly bench=${BENCH:-BenchmarkHasPermission}
readonly benchtime=${BENCHTIME:-20000x}
readonly count=${COUNT:-6}

base=${1:-}
if [[ -z $base ]]; then
  base=$(git merge-base HEAD upstream/main 2>/dev/null || git rev-parse HEAD~1)
fi
echo "base: $(git rev-parse --short "$base") $(git log -1 --format=%s "$base")"

work=$(mktemp -d)
trap 'git worktree remove --force "$work/base" >/dev/null 2>&1 || true; rm -rf "$work"' EXIT

git worktree add --detach "$work/base" "$base" >/dev/null
cp tools/bench-policy/baseline_test.go "$work/base/src/common/rbac/project/zz_bench_test.go"

echo
echo "== base =="
( cd "$work/base/src" && go test ./common/rbac/project/ -run xxx -bench "$bench" -benchtime "$benchtime" -count "$count" ) | tee "$work/base.txt"

echo
echo "== this branch =="
( cd "$root/src" && go test ./common/rbac/project/ -run xxx -bench "$bench" -benchtime "$benchtime" -count "$count" ) | tee "$work/new.txt"

if command -v benchstat >/dev/null; then
  echo
  echo "== comparison =="
  benchstat "$work/base.txt" "$work/new.txt"
else
  echo
  echo "install benchstat for a comparison: go install golang.org/x/perf/cmd/benchstat@latest"
fi
