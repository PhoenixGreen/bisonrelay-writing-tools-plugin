#!/usr/bin/env bash
# Builds plugin.wasm from this module's Go source. Requires Go 1.24+
# (go:wasmexport); no other toolchain (no TinyGo, no cgo).
set -euo pipefail
cd "$(dirname "$0")"

GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o plugin.wasm .

echo "Built plugin.wasm ($(du -h plugin.wasm | cut -f1))"
