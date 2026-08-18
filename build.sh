#!/usr/bin/env bash
# Builds plugin.wasm from this module's Go source. Requires Go 1.24+
# (go:wasmexport); no other toolchain (no TinyGo, no cgo).
set -euo pipefail
cd "$(dirname "$0")"

# The datasets are embedded compressed; see main.go. Only the .gz files are
# committed, so this regenerates them when a fresh .txt has been produced by
# tools/mkwords or tools/mkthesaurus.
for name in words-en-US common-en-US words-en-GB common-en-GB \
            thesaurus definitions exceptions; do
  if [ -f "$name.txt" ] && [ "$name.txt" -nt "$name.txt.gz" ]; then
    echo "Recompressing $name.txt"
    gzip -9 -c "$name.txt" > "$name.txt.gz"
  fi
done

GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o plugin.wasm .

echo "Built plugin.wasm ($(du -h plugin.wasm | cut -f1))"

# Bison Relay ships this plugin inside the client rather than importing it, as
# client/pluginmgr/builtin/writingtools.wasm.gz -- so a build that stops here
# reaches nothing. The app unpacks that gz into its data directory on first
# run and re-checks it by size, and the copy in this repository is not
# consulted by anything at all.
#
# This is worth automating because the failure is silent in the worst way: the
# build succeeds, the tests pass, and the running app quietly goes on using
# whatever rules were embedded the last time somebody remembered. A rule that
# does not fire looks exactly like a rule that was never written.
#
# Only when the checkout is where it is expected, and only the bytes -- the
# client is Go, so the app still has to be rebuilt (bruig/build_desktop.sh)
# before a new rule reaches a running copy.
bruig_builtin="../bisonrelay/client/pluginmgr/builtin"
if [ -d "$bruig_builtin" ]; then
  gzip -9 -c plugin.wasm > "$bruig_builtin/writingtools.wasm.gz"
  cp manifest.json "$bruig_builtin/writingtools.manifest.json"
  echo "Synced into $bruig_builtin -- rebuild bruig for it to take effect"
else
  echo "No bruig checkout at $bruig_builtin; built-in copy NOT updated"
fi
