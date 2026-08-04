#!/usr/bin/env bash
# Builds plugin.wasm from this module's Go source. Requires Go 1.24+
# (go:wasmexport); no other toolchain (no TinyGo, no cgo).
set -euo pipefail
cd "$(dirname "$0")"

# The datasets are embedded compressed; see main.go. Only the .gz files are
# committed, so this regenerates them when a fresh .txt has been produced by
# tools/mkwords or tools/mkthesaurus.
for name in words common thesaurus; do
  if [ -f "$name.txt" ] && [ "$name.txt" -nt "$name.txt.gz" ]; then
    echo "Recompressing $name.txt"
    gzip -9 -c "$name.txt" > "$name.txt.gz"
  fi
done

GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o plugin.wasm .

echo "Built plugin.wasm ($(du -h plugin.wasm | cut -f1))"
