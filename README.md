# bisonrelay-writing-tools-plugin

Formerly `bisonrelay-spellcheck-plugin`. The manifest id is still `spellcheck`,
which is what Bison Relay tracks an installed plugin by; changing it would orphan
an existing install for no gain.

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

Rules come in two severities. An **error** is text that is wrong whatever the writer meant --
a misspelling, a missing apostrophe, "have went" -- and every one of them is held to a
standard of never firing on correct writing. A **suggestion** is an opinion: wordiness, a
cliche, the passive voice. Those cannot meet that standard and are not asked to, because
the app underlines them in a different colour and lists them on a separate page, so being
wrong costs a quieter mark somewhere the reader went looking for opinions.

That distinction is what makes the second and third groups below shippable at all. Without
it they would put the same alarming red wave under prose that is perfectly good, and the
reader who learns to ignore that mark ignores it over the misspellings too.

Each rule carries a category and an explanation as well as a message and a replacement.
The message names the problem in the few words a menu row allows; the explanation is a
sentence saying what is wrong and *why*, which is the whole reason the app shows a popup
rather than a list of replacement words. Someone who already knows the rule does not need
to right-click; someone who does not cannot act on "Should be it's" alone.

Each rule is deliberately conservative, because a false positive costs more than a missed
error: an underline under correct writing teaches people to ignore the underlines. That
rules out most of what a word processor offers -- subject/verb agreement, their/there,
its/it's -- none of which a regex can decide without guessing. A filler-word rule was
written and then removed when its own test caught it firing on "he said quite clearly".

**Function words and contractions.** 143 hand-written glosses for the words WordNet structurally cannot hold.
It is organised as four networks -- nouns, verbs, adjectives, adverbs -- and has nothing at
all for "the", "of", "although" or "which". Those are the commonest words in English and so
the ones a reader is likeliest to select, which made the gap look like a broken feature
rather than a boundary of the data. Each says what the word *does* in a sentence, because
that is what these words have instead of a meaning. They replace WordNet's entry where there
is one, since "will" as willpower is not the sense anyone selecting it means.

Contractions are here for the same reason -- both halves of one are function words, so
WordNet holds none of them. The negatives are spelled out rather than left to reduce to
their base verb: "wouldn't" reducing to "would" is true and useless, because it drops the
negation that is the whole reason the word is there.

A typographic apostrophe is folded to a plain one before any lookup. A text field that
substitutes U+2019 as you type -- which macOS does by default -- otherwise makes every
contraction unfindable, since the data is keyed with the plain form.

**Confusable pairs.** 33 rules covering then/than, lose/loose, affect/effect, your/you're,
to/too, whose/who's, their/they're and the past participles after have/has/had. LanguageTool
decides the general case with ngram frequency data -- given "more money then you", it knows
"money than you" is the commoner trigram -- which is gigabytes and out of the question here.
What is possible without any of it is the decidable slice: positions where only one of the
pair can be right whatever the sentence is about. "Than" cannot follow "and"; a possessive
cannot precede a verb; "apart of" is never a phrase.

The your/you're rule is the shape of the work. It began with sixteen words and finished with
four, because almost everything that reads as an adjective after "you're" also reads as
something ownable after "your": your right (to refuse), your correct address, your late
payment, your invited guests, your amazing work. Each of those is now a line in the corpus.

**Wordiness and cliches.** 106 suggestions, from a table of phrase pairs -- "in the event
that" to "if", "each and every" to "each" -- plus 25 phrases worn smooth by overuse, flagged
with no replacement because proposing one would be a worse cliche than the original. Their
first letter matches in either case, since the commonest place for "Due to the fact that" is
the start of a sentence, where a case-sensitive pattern would never fire.

**Passive voice.** Two rules, both requiring proof. One needs a named agent ("was written by
someone else"); the other needs "being" before a participle ("is being considered"). Most
passive writing names no agent and goes unflagged, which is the price of never flagging "I
was interested" -- an `-ed` word after "was" is far more often an adjective than a verb.

**Analysis checks.** Four checks a regex cannot express, because they count: a word used
four or more times in one paragraph, a sentence over thirty words, three consecutive
sentences opening with the same word, and both spellings of a variant pair (organise /
organize) in one message. The plugin declares these rather than implementing them -- it
names a check the app knows how to run and supplies the threshold, the wording and the
explanation. Same division as the regexes, which the plugin also writes and never executes.
An app that does not implement an id ignores it, so a check can ship here first.

**Definitions.** 77,549 words and 111,303 meanings, condensed from WordNet 3.1. Up to
three meanings per word, each labelled with its part of speech, with WordNet's usage
examples dropped -- a menu has room for what the word means and not for a paragraph
demonstrating it.

They are returned alongside the synonyms and kept deliberately separate from them, never
paired up. The two datasets divide a word into different senses, so saying which meaning
went with which group of synonyms would mean guessing, and a wrong guess reads as
confidently wrong rather than merely unhelpful.

**Synonyms.** 62,959 words, condensed from the MyThes/WordNet thesaurus. Senses are kept
apart -- "bank" as a place to keep money is listed separately from its river sense --
because pooling them is what makes a thesaurus give confidently wrong advice. Antonyms are
included and marked.

**Lemmatisation.** Both datasets are keyed by base forms only -- "cat" and not "cats",
"run" and not "running" -- which on its own leaves two thirds of the dictionary with nothing
to look up. Measured against the 121,115-word dictionary, a direct lookup answers for 36% of
it. Reducing the word first takes that to **86%**.

The rules are WordNet's own Morphy: suffix detachments per part of speech, plus the
irregular-form lists that ship with WordNet for the words no rule reaches -- went/go,
children/child, better/good, mice/mouse. Possessives are stripped first, since 27,000 of the
dictionary's entries are one. Every candidate is checked against the data before it is
accepted, so an over-eager rule costs nothing: "sing" does not become "s".

The 16,000 words still unanswered are mostly proper nouns SCOWL carries and WordNet does not
gloss -- Canaveral, Mississauga, Bertha -- plus a tail of rare words WordNet simply lacks.
That is a limit of the source, not of the lookup.

Unlike the dictionary and the rules, none of this is handed to the app: the datasets are
10MB, far too much to push across for a feature used a word at a time, so the plugin keeps it and answers
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

## Languages

Two: **English (UK)** and **English (US)**, chosen in Settings > Plugins. A language
is a different word list rather than a filter over one -- "colour" is in the British list
and "color" in the American, and each is a misspelling in the other -- so switching goes
back to the plugin for a fresh dictionary.

The lists come from [SCOWL](http://wordlist.sourceforge.net), which publishes plain lists
of surface forms split by how common each word is and by which English uses it. Everyone
gets `english-`, and then either `american-` or `british-`, up to frequency tier 60. Both
land at roughly 120,000 words.

An earlier version expanded a hunspell dictionary instead -- roots plus affix flags,
`abandon/LDGS` into the five words that spells. That worked for the American dictionary,
whose affix file has 73 rules, and produced nonsense for the British one, whose 1,362 rules
include productive derivational affixes that hunspell constrains in ways this plugin's
expander did not: `underwearisable`, `overnazismativeness`, `miscorporativismativeness`.
Reimplementing hunspell correctly is a real project and not this one; SCOWL's lists need no
engine at all. The American list changed by 1% in the move, almost all of it obscure, and
gained the accented words that were previously dropped rather than folded -- `melee`,
`emigre`, `confreres`.

WordNet and MyThes are both American, so a British spelling has no definition or synonyms
of its own. Lookups reduce it first, using the same variant table the consistency check is
built from, so "colour" and "organised" still answer.

## Regenerating the data

`words.txt` is committed, so building needs neither the tool nor `data/`; both are kept so
the list is reproducible and auditable rather than an opaque blob.

```sh
go run ./tools/mkwords -locale en-US > words-en-US.txt
go run ./tools/mkwords -locale en-GB > words-en-GB.txt
go run ./tools/mkthesaurus > thesaurus.txt
go run ./tools/mkdefs > definitions.txt
go run ./tools/mkexceptions > exceptions.txt
./build.sh          # recompresses whichever changed, then compiles
```

Only the compressed `.gz` files are committed -- they are what `main.go` embeds, and
keeping the plain text alongside them would be the same data twice, free to drift. To read
one:

```sh
gunzip -c words.txt.gz | grep '^decred$'
```

## Data licences and attribution

The **dictionaries** are SCOWL's own word lists, [SCOWL](http://wordlist.sourceforge.net)
Copyright 2000-2018 Kevin Atkinson. Full text: **`data/LICENSE-SCOWL`**.

The **thesaurus** is the MyThes `th_en_US_v2` data shipped with LibreOffice, derived from
[WordNet](https://wordnet.princeton.edu), WordNet 2.1 Copyright 2005 Princeton University.
Full text: **`data/LICENSE-WORDNET`**.

The **definitions** are from [WordNet](https://wordnet.princeton.edu) 3.1, Copyright 2011
Princeton University. Full text: **`data/LICENSE-WORDNET31`**.

All three permit use, modification and redistribution provided the copyright and permission
notices travel with them. Those notices ship in this repository and must stay with any
redistribution of `words.txt`, `thesaurus.txt`, `definitions.txt`, or a `plugin.wasm` built
from them.

## Building

Requires Go 1.24+ (for `go:wasmexport`) -- no TinyGo, no cgo.

```sh
./build.sh      # -> plugin.wasm
```

## Packaging for import

```sh
./package.sh    # -> writing-tools-plugin-vX.Y.Z.zip (manifest.json + plugin.wasm)
```

## Installing in Bison Relay

Grab `writing-tools-plugin-vX.Y.Z.zip` from the [Releases](../../releases) page (or build it
yourself with `./package.sh`) -- import it as-is, no need to unzip it first.

In Bison Relay: Settings > Plugins > Import Plugin, then select the zip. This plugin ships
with `enabledByDefault: false`, so flip its switch on in the same list after importing.
Disable/remove it from there too.

## Layout

- `main.go` -- the WebAssembly exports, and nothing else: `alloc` (required of every
  dynamic-wasm plugin) and `get_spellcheck_data` (the one capability export this plugin
  implements). See `client/pluginmgr/wasmhost`'s package doc in the main bisonrelay repo
  for the full ABI reference.
- `spellcheck/spellcheck.go` -- the error rules and the wordlist parser.
- `spellcheck/rules_confusions.go` -- the confusable pairs, each in the one position where
  the right word is not a matter of opinion.
- `spellcheck/rules_style.go` -- the opinions: the wordiness table, the cliche list, the two
  passive-voice rules, and the British/American spelling pairs.
- `spellcheck/analysis.go` -- the four declared checks that count rather than match.
- `spellcheck/` -- the plugin's actual content: the grammar rules and the wordlist parser,
  as ordinary Go so they can be tested without building for wasm. Nothing here matches
  anything; the rules are regex *source*, executed by Dart in the app, whose engine
  (unlike Go's RE2) supports the backreferences a rule like "repeated word" needs.
- `thesaurus/` -- the synonym and definition lookup: a binary search over each generated
  file's line offsets, parsing one entry at a time. The two datasets are searched
  independently, so a word in one and not the other still produces an answer.
- `thesaurus/morphy.go` -- reduces a word as typed to the base form the data is keyed by.
- `words.txt.gz`, `common.txt.gz`, `thesaurus.txt.gz`, `definitions.txt.gz`,
  `exceptions.txt.gz` -- the
  generated datasets, embedded compressed via `//go:embed` and decompressed on first use.
  Together they are 11MB of text and 3.6MB compressed, which is the difference between a
  15MB module and a 7MB one. It costs nothing at runtime that was not already paid:
  `//go:embed` puts the data in linear memory either way.
- `data/scowl/` -- SCOWL's curated word lists, one file per variant and frequency tier.
- `data/` -- the other sources: the supplemental vocabulary, the function-word and
  contraction glosses, the MyThes and WordNet data, and every licence.
- `tools/mkwords/` -- builds one language's `words-*.txt` and `common-*.txt` from
  `data/scowl`.
- `tools/mkthesaurus/` -- condenses the MyThes source into `thesaurus.txt`.
- `tools/mkdefs/` -- condenses WordNet's `data.*` files into `definitions.txt`, merging
  `data/function-words.txt` over the top.
- `tools/mkexceptions/` -- merges WordNet's four irregular-inflection lists into
  `exceptions.txt`.
