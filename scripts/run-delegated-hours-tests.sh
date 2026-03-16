#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$repo_root"

# Run delegated-hours focused tests in coin package
GO111MODULE=on go test ./src/coin -run 'TestVerifyDelegatedConstraints'
