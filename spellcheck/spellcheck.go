// Package spellcheck holds this plugin's actual content -- the wordlist and
// the writing-style rules -- separately from the WebAssembly plumbing in
// main.go, so it can be built and tested as ordinary Go.
//
// Nothing here matches anything. The plugin's whole job is to hand Bison
// Relay a body of data once, when it is enabled; the matching, the wavy
// underline and the suggestion menu all live in the app. That split is why
// the rules below are regex *source*, never compiled here: they are executed
// by Dart, whose engine (unlike Go's RE2) supports the backreferences a rule
// like "repeated word" needs.
package spellcheck

import "strings"

// GrammarRule is one regex-based writing check. It mirrors the
// spellcheck-data capability's JSON schema exactly; this plugin is a
// separate Go module and deliberately takes no dependency on Bison Relay to
// get it.
type GrammarRule struct {
	// Pattern is a regular expression in Dart's dialect.
	Pattern string `json:"pattern"`
	// Message names the problem, for the suggestion menu.
	Message string `json:"message"`
	// Suggest is the replacement, which may reference Pattern's capture
	// groups as $1, $2 and so on. Empty means the rule only flags the text
	// and proposes nothing -- correct when there is no single right fix.
	Suggest string `json:"suggest"`
}

// Data is the whole payload the capability returns.
type Data struct {
	Words []string `json:"words"`

	// CommonWords is a subset of Words ordered most-common-first, used by the
	// app to rank corrections. Edit distance cannot rank on its own: "teh" is
	// one typo from "the", "tech", "meh" and "te" alike, and without knowing
	// which of those people write, the intended word is as likely to be
	// missing from the handful shown as not.
	CommonWords []string `json:"commonWords,omitempty"`

	GrammarRules []GrammarRule `json:"grammarRules"`
}

// domainTLDs are the endings that mark a dot as part of a web address rather
// than the end of a sentence. Not exhaustive, and cannot be: it is a
// heuristic standing in for "is this a hostname", which no regex can decide.
const domainTLDs = "com|org|net|edu|gov|mil|int|io|dev|app|xyz|info|biz|tv|" +
	"ai|gg|rs|sh|ly|cc|onion|uk|de|fr|ru|au|ca|us|jp|cn|nz|za|eu|ch|nl|se|" +
	"tech|site|online|store|blog|news|wiki|link|page|pro|me|co|gl|fm"

// Rules are the writing checks, in the order they are applied.
//
// Every rule here is deliberately conservative, because a false positive
// costs far more than a missed error: a wavy underline under correct writing
// trains people to ignore the feature entirely. So each one fires only on
// text that is wrong regardless of context, and anything needing to know
// what a sentence *means* is left out. That rules out the whole family of
// checks people expect from a word processor -- subject/verb agreement,
// their/there, its/it's, a/an before an acronym -- none of which a regex can
// decide without guessing.
var Rules = []GrammarRule{
	// --- doubled input: almost always a slip of the hands ---
	{
		Pattern: `\b(\w+)([ \t]+)\1\b`,
		Message: "Repeated word",
		Suggest: "$1",
	},
	{
		Pattern: `[ ]{2,}`,
		Message: "Multiple spaces",
		Suggest: " ",
	},
	{
		Pattern: `([,;:])\1+`,
		Message: "Repeated punctuation",
		Suggest: "$1",
	},

	// --- spacing around punctuation ---
	{
		Pattern: `[ \t]+([,.!?;:])`,
		Message: "Space before punctuation",
		Suggest: "$1",
	},
	{
		// Fires whatever the case of the next letter, and fixes both faults
		// at once: "(if you are free).what's" becomes ". What's". Restricted
		// to a capital, this missed exactly that -- a run-together sentence
		// whose next word was also left lower case fell between this rule and
		// the capitalisation one, since each assumed the other's condition.
		//
		// Two guards keep it off text that is not a sentence boundary. The
		// lookbehind wants two word characters or a closing bracket before
		// the stop, which excludes initialisms ("e.g.", "U.S.A."). The
		// lookahead excludes a domain, either directly ("example.com") or
		// through further labels ("news.ycombinator.com").
		//
		// The domain list deliberately omits ccTLDs that are ordinary English
		// words -- it, is, at, in, so, be, no -- because a sentence starts
		// with one of those far more often than a chat mentions example.it,
		// and a missed flag costs less than a wrong one.
		Pattern: `(?<=\w\w|[)\]"'])([.!?])(?!(?:` + domainTLDs +
			`)\b|[\w-]*\.(?:` + domainTLDs + `)\b)([A-Za-z])`,
		Message: "Missing space after punctuation",
		Suggest: "$1 $U2",
	},
	{
		Pattern: `([,;:])([A-Za-z])`,
		Message: "Missing space after punctuation",
		Suggest: "$1 $2",
	},
	{
		Pattern: `\(\s+`,
		Message: "Space inside bracket",
		Suggest: "(",
	},
	{
		Pattern: `\s+\)`,
		Message: "Space inside bracket",
		Suggest: ")",
	},

	// --- capitalisation with one unambiguous answer ---
	{
		// The pronoun is the only single letter that is always capitalised,
		// which is what makes this safe to correct outright.
		//
		// A dot on either side means an initialism -- "i.e." -- and not the
		// pronoun at all. That costs the rule a genuine "i" ending a
		// sentence, which is a sentence hardly anyone writes.
		Pattern: `(?<!\.)\bi\b(?!\.)`,
		Message: "\"I\" is capitalised",
		Suggest: "I",
	},
	{
		Pattern: `\bi'(m|ve|ll|d)\b`,
		Message: "\"I\" is capitalised",
		Suggest: "I'$1",
	},

	// --- contractions people reliably mistype ---
	// Each of these is wrong in every context, unlike its/it's, which is
	// exactly why those two are absent.
	{
		Pattern: `\byour welcome\b`,
		Message: "Should be \"you're welcome\"",
		Suggest: "you're welcome",
	},
	{
		Pattern: `\b(could|should|would|must|might)\s+of\b`,
		Message: "Should be \"$1 have\"",
		Suggest: "$1 have",
	},
	{
		Pattern: `\balot\b`,
		Message: "\"a lot\" is two words",
		Suggest: "a lot",
	},
	{
		Pattern: `\beveryday (I|you|we|they|he|she|it)\b`,
		Message: "\"every day\" is two words as an adverb",
		Suggest: "every day $1",
	},
	{
		Pattern: `\bcant\b`,
		Message: "Missing apostrophe",
		Suggest: "can't",
	},
	{
		Pattern: `\bwont\b`,
		Message: "Missing apostrophe",
		Suggest: "won't",
	},
	{
		Pattern: `\bdont\b`,
		Message: "Missing apostrophe",
		Suggest: "don't",
	},
	{
		Pattern: `\bdoesnt\b`,
		Message: "Missing apostrophe",
		Suggest: "doesn't",
	},
	{
		Pattern: `\bisnt\b`,
		Message: "Missing apostrophe",
		Suggest: "isn't",
	},
	{
		Pattern: `\bwasnt\b`,
		Message: "Missing apostrophe",
		Suggest: "wasn't",
	},
	{
		Pattern: `\bwouldnt\b`,
		Message: "Missing apostrophe",
		Suggest: "wouldn't",
	},
	{
		Pattern: `\bcouldnt\b`,
		Message: "Missing apostrophe",
		Suggest: "couldn't",
	},
	{
		Pattern: `\bshouldnt\b`,
		Message: "Missing apostrophe",
		Suggest: "shouldn't",
	},
	{
		Pattern: `\bthats\b`,
		Message: "Missing apostrophe",
		Suggest: "that's",
	},
	{
		Pattern: `\bwhats\b`,
		Message: "Missing apostrophe",
		Suggest: "what's",
	},
	{
		Pattern: `\blets\s+(go|see|say|try|talk|do|get|make)\b`,
		Message: "Missing apostrophe",
		Suggest: "let's $1",
	},

	{
		// The preceding punctuation is matched in a lookbehind rather than a
		// capture, so the flagged span is the letter alone. Written as a
		// capture, the span began at the previous sentence's full stop, which
		// underlined text that was not wrong and -- because a right-click
		// selects the word under the pointer -- meant clicking the offending
		// letter often missed the span entirely while clicking the full stop
		// beside it worked.
		//
		// $U1 upper-cases the captured letter; a literal template cannot,
		// since the letter to capitalise is whatever was typed.
		//
		// This is the noisiest rule here by some way: plenty of people write
		// chat entirely in lower case on purpose, and it flags the first word
		// of every sentence they send. It is included because a missing
		// capital is a real error and the fix is unambiguous, but it is the
		// first rule to drop if the underlines become wallpaper.
		Pattern: `(?<=^|[.!?]\s)([a-z])`,
		Message: "Sentence should start with a capital",
		Suggest: "$U1",
	},

	// --- confusions with a decidable answer ---
	// their/there/they're normally needs to know what a sentence means, which
	// is why the general case is absent. These are the positions where it
	// does not: "their" is a possessive, so a pronoun or a verb cannot follow
	// it, whatever the sentence is about.
	{
		Pattern: `\btheir\s+(is|are|was|were|will|would|has|have|had)\b`,
		Message: "Should be \"there $1\"",
		Suggest: "there $1",
	},
	{
		Pattern: `\btheir\s+(he|she|it|we|they|you)\b`,
		Message: "Should be \"there $1\"",
		Suggest: "there $1",
	},
	{
		// "there own" is never right: only the possessive can precede it.
		Pattern: `\bthere\s+(own|self|selves)\b`,
		Message: "Should be \"their $1\"",
		Suggest: "their $1",
	},
	{
		Pattern: `\bthere\s+(is|are|was|were)\s+own\b`,
		Message: "Should be \"their own\"",
		Suggest: "their own",
	},

	// --- style: flagged, but with no replacement proposed ---
	// Suggest is empty because there is no single correct rewrite; the
	// point is to draw the eye, not to make the edit.
	{
		// Captured and back-referenced so the fix keeps whichever mark was
		// used: "!!!" collapses to "!", "???" to "?". A run of mixed marks
		// deliberately does not match -- there is no single right answer for
		// "!?!".
		Pattern: `([!?])\1{2,}`,
		Message: "Excessive punctuation",
		Suggest: "$1",
	},
	// A "filler word" rule (very, really, quite, actually...) was tried here
	// and removed: it fired on "he said quite clearly", which is correct
	// writing. Whether an intensifier is filler depends on the sentence, so
	// the rule cannot help but underline prose that is fine -- the exact
	// failure the note at the top of this list warns about.
	{
		Pattern: `\b(in order to)\b`,
		Message: "Wordy -- \"to\" usually suffices",
		Suggest: "to",
	},
	{
		Pattern: `\b(due to the fact that)\b`,
		Message: "Wordy",
		Suggest: "because",
	},
	{
		Pattern: `\b(at this point in time)\b`,
		Message: "Wordy",
		Suggest: "now",
	},
}

// ParseWords tokenizes a generated wordlist: one lowercase word per line,
// with blank lines and "#" comments ignored.
func ParseWords(txt string) []string {
	lines := strings.Split(txt, "\n")
	words := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.ToLower(strings.TrimSpace(line))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		words = append(words, line)
	}
	return words
}
