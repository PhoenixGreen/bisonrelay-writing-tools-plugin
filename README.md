# bisonrelay-spellcheck-plugin

A spell/writing-style-check data plugin for [Bison Relay](https://bisonrelay.org), built
against its dynamic-wasm plugin system (`client/pluginmgr` + `client/pluginmgr/wasmhost` in
the main repo). Supplies a 121,000-word English dictionary and a set of regex-based writing-style rules
that Bison Relay's own spellcheck UI applies to chat/post/comment text as you type (wavy
underline, native suggestion menu with corrections).

This is a **headless** plugin: it contributes no nav item or screens, just the
`spellcheck-data` capability. All matching/highlighting logic lives in Bison Relay itself
(Dart, using Flutter's native `SpellCheckConfiguration` API) -- this plugin only supplies the
data, once, when enabled. It runs as sandboxed WebAssembly with no network or filesystem
access at all; it doesn't need any, since the wordlist and rules are static data compiled
into the module.

## What it checks

**Spelling.** A 121,407-word dictionary: 121,338 surface forms expanded from the
SCOWL-derived hunspell `en_US` dictionary, plus 69 words of Bison Relay, Decred and
crypto vocabulary a general dictionary has never heard of (see `data/extra-words.txt`).

Surface forms matter. A hunspell `.dic` stores roots plus affix flags -- `abandon/LDGS`,
not the five words that spells -- so a checker fed the roots alone flags every inflection
anyone actually types. `tools/mkwords` expands them ahead of time, which keeps this
plugin's runtime a set membership test.

**Writing style.** 30 rules, covering repeated words, runs of spaces, spacing around
punctuation and brackets, uncapitalised "I", and contractions that are wrong in every
context (`cant`, `wont`, `alot`, `could of`).

Each rule is deliberately conservative, because a false positive costs more than a missed
error: an underline under correct writing teaches people to ignore the underlines. That
rules out most of what a word processor offers -- subject/verb agreement, their/there,
its/it's -- none of which a regex can decide without guessing. A filler-word rule was
written and then removed when its own test caught it firing on "he said quite clearly".

## Regenerating the wordlist

`words.txt` is committed, so building needs neither the tool nor `data/`; both are kept so
the list is reproducible and auditable rather than an opaque blob.

```sh
go run ./tools/mkwords > words.txt
```

## Dictionary licence and attribution

The dictionary is derived from [SCOWL](http://wordlist.sourceforge.net), Copyright
2000-2018 Kevin Atkinson, with the affix file derived from Geoff Kuenning's Ispell under
his BSD licence. Both permit use, modification and redistribution provided the copyright
and permission notices travel with them; the full text is in **`data/LICENSE-SCOWL`**,
which ships in this repository and must stay with any redistribution of `words.txt` or a
`plugin.wasm` built from it.

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

Grab `spellcheck-plugin-vX.Y.Z.zip` from the [Releases](../../releases) page (or build it
yourself with `./package.sh`) -- import it as-is, no need to unzip it first.

In Bison Relay: Settings > Plugins > Import Plugin, then select the zip. This plugin ships
with `enabledByDefault: false`, so flip its switch on in the same list after importing.
Disable/remove it from there too.

## Layout

- `main.go` -- the WebAssembly exports, and nothing else: `alloc` (required of every
  dynamic-wasm plugin) and `get_spellcheck_data` (the one capability export this plugin
  implements). See `client/pluginmgr/wasmhost`'s package doc in the main bisonrelay repo
  for the full ABI reference.
- `spellcheck/` -- the plugin's actual content: the grammar rules and the wordlist parser,
  as ordinary Go so they can be tested without building for wasm. Nothing here matches
  anything; the rules are regex *source*, executed by Dart in the app, whose engine
  (unlike Go's RE2) supports the backreferences a rule like "repeated word" needs.
- `words.txt` -- the generated wordlist, embedded via `//go:embed`.
- `data/` -- the sources `words.txt` is generated from: the hunspell `.dic`/`.aff` pair,
  the supplemental vocabulary, and the dictionary licence.
- `tools/mkwords/` -- the affix expander that turns `data/` into `words.txt`.
