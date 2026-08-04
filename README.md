# bisonrelay-spellcheck-plugin

A spell/writing-style-check data plugin for [Bison Relay](https://bisonrelay.org), built
against its dynamic-wasm plugin system (`client/pluginmgr` + `client/pluginmgr/wasmhost` in
the main repo). Supplies a 121,000-word English dictionary and a set of regex-based writing-style rules
that Bison Relay's own spellcheck UI applies to chat/post/comment text as you type (wavy
underline, native suggestion menu with corrections).

This is a **headless** plugin: it contributes no nav item or screens, just the
`spellcheck-data` and `thesaurus` capabilities. All matching/highlighting logic lives in Bison Relay itself
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

**Synonyms.** 62,959 words, condensed from the MyThes/WordNet thesaurus. Senses are kept
apart -- "bank" as a place to keep money is listed separately from its river sense --
because pooling them is what makes a thesaurus give confidently wrong advice. Antonyms are
included and marked.

Unlike the two datasets above, this one is never handed to the app: at 3.5MB it is far too
much to push across for a feature used a word at a time, so the plugin keeps it and answers
lookups. It is not fully parsed either -- the generated file is sorted, so a lookup binary-
searches line offsets and parses only the line it lands on, keeping the resident cost inside
the wasm instance to a few hundred KB rather than tens of megabytes.

Synonym ranking follows how safely a word substitutes. MyThes marks its cross-references,
and they do not all mean the same thing: `(generic term)` is a hypernym and never a
replacement (offering "city" for "Hague" changes what the sentence says), while
`(similar term)` and `(related term)` usually are -- and for many adjectives are the *only*
ones, since the direct synonym list is empty. A first version dropped all three and left
"happy" with felicitous, glad and well-chosen while discarding cheerful, contented and
blissful.

## Regenerating the data

`words.txt` is committed, so building needs neither the tool nor `data/`; both are kept so
the list is reproducible and auditable rather than an opaque blob.

```sh
go run ./tools/mkwords > words.txt
go run ./tools/mkthesaurus > thesaurus.txt
./build.sh          # recompresses whichever changed, then compiles
```

Only the compressed `.gz` files are committed -- they are what `main.go` embeds, and
keeping the plain text alongside them would be the same data twice, free to drift. To read
one:

```sh
gunzip -c words.txt.gz | grep '^decred$'
```

## Data licences and attribution

The **dictionary** is derived from [SCOWL](http://wordlist.sourceforge.net), Copyright
2000-2018 Kevin Atkinson, with the affix file derived from Geoff Kuenning's Ispell under
his BSD licence. Full text: **`data/LICENSE-SCOWL`**.

The **thesaurus** is the MyThes `th_en_US_v2` data shipped with LibreOffice, derived from
[WordNet](https://wordnet.princeton.edu), WordNet 2.1 Copyright 2005 Princeton University.
Full text: **`data/LICENSE-WORDNET`**.

Both permit use, modification and redistribution provided the copyright and permission
notices travel with them. Those notices ship in this repository and must stay with any
redistribution of `words.txt`, `thesaurus.txt`, or a `plugin.wasm` built from them.

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
- `thesaurus/` -- the synonym lookup: a binary search over the generated file's line
  offsets, parsing one entry at a time.
- `words.txt.gz`, `thesaurus.txt.gz` -- the generated datasets, embedded compressed via
  `//go:embed` and decompressed on first use. Together they are 4.6MB of text and 1.5MB
  compressed, which is the difference between a 7.6MB module and a 4.7MB one. It costs
  nothing at runtime that was not already paid: `//go:embed` puts the data in linear memory
  either way.
- `data/` -- the sources `words.txt` is generated from: the hunspell `.dic`/`.aff` pair,
  the supplemental vocabulary, and the dictionary licence.
- `tools/mkwords/` -- the affix expander that turns `data/` into `words.txt`.
- `tools/mkthesaurus/` -- condenses the MyThes source into `thesaurus.txt`.
