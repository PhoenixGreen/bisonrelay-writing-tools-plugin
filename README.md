# bisonrelay-spellcheck-plugin

A spell/writing-style-check data plugin for [Bison Relay](https://bisonrelay.org), built
against its dynamic-wasm plugin system (`client/pluginmgr` + `client/pluginmgr/wasmhost` in
the main repo). Supplies a wordlist and a handful of regex-based writing-style rules that
Bison Relay's own spellcheck UI applies to chat/post/comment text as you type (wavy
underline, native suggestion menu).

This is a **headless** plugin: it contributes no nav item or screens, just the
`spellcheck-data` capability. All matching/highlighting logic lives in Bison Relay itself
(Dart, using Flutter's native `SpellCheckConfiguration` API) -- this plugin only supplies the
data, once, when enabled. It runs as sandboxed WebAssembly with no network or filesystem
access at all; it doesn't need any, since the wordlist and rules are static data compiled
into the module.

## Building

Requires Go 1.24+ (for `go:wasmexport`) -- no TinyGo, no cgo.

```sh
./build.sh      # -> plugin.wasm
```

## Packaging for import

```sh
./package.sh    # -> spellcheck-plugin-vX.Y.Z.zip (manifest.json + plugin.wasm)
```

## Installing in Bison Relay

Settings > Plugins > Import Plugin, then select the generated zip. Enable/disable and
remove it from the same screen.

## Layout

- `main.go` -- the plugin's WebAssembly exports (the ABI boundary): `alloc` (required by
  every dynamic-wasm plugin) and `get_spellcheck_data` (the one capability export this
  plugin implements). See `client/pluginmgr/wasmhost`'s package doc in the main bisonrelay
  repo for the full ABI reference.
- `words.txt` -- the wordlist, embedded into the compiled module via `//go:embed`. Swap in a
  larger/different wordlist here without touching any code.
