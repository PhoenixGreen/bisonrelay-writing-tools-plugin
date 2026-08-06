#!/usr/bin/env bash
# Builds plugin.wasm and zips it with manifest.json into the layout Bison
# Relay's Settings > Plugins > Import Plugin expects: a single top-level
# folder (named after the plugin id) containing manifest.json.
set -euo pipefail
cd "$(dirname "$0")"

./build.sh

version=$(sed -n 's/.*"version": *"\([^"]*\)".*/\1/p' manifest.json | head -1)
# head -1 matters here: if manifest.json ever grows nested objects with
# their own "id"/"version"-like fields, this pattern would match those too,
# not just the top-level plugin id/version.
id=$(sed -n 's/.*"id": *"\([^"]*\)".*/\1/p' manifest.json | head -1)
outfile="writing-tools-plugin-v${version}.zip"

stagedir=$(mktemp -d)
trap 'rm -rf "$stagedir"' EXIT
mkdir "$stagedir/$id"
cp manifest.json plugin.wasm "$stagedir/$id/"

rm -f "$outfile"
(cd "$stagedir" && zip -rq "$OLDPWD/$outfile" "$id")

echo "Packaged $outfile"
