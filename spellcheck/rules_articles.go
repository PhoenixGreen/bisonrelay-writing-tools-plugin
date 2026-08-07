package spellcheck

import (
	"strings"
	"unicode"
)

// rules_articles.go is "a" against "an".
//
// The rule everybody knows -- "an" before a vowel -- is about *sound*, and
// spelling only mostly agrees with it. "A unique opportunity" and "an hour"
// are both correct and both break the rule as usually stated, because
// "unique" opens with a consonant sound and "hour" opens with a vowel one.
//
// Nothing here can hear anything, so the letters decide and the exceptions
// are listed. That makes this the clearest case in the plugin for
// antipatterns: the rule is one line and the interesting part is entirely
// the two lists of words it must not touch. Written as lookaheads they
// would be unreadable, untestable on their own, and invisible to the corpus
// -- which is exactly where a list this long most needs checking.

const explainArticle = "\"An\" goes before a word that starts with a vowel " +
	"sound and \"a\" before one that starts with a consonant sound. It is " +
	"the sound and not the letter, which is why it is \"a unique offer\" " +
	"and \"an hour\"."

// consonantSoundedVowels are the words that open with a vowel letter and a
// consonant sound, so they take "a".
//
// Two families, plus a couple of strays. The "u" words are the ones said
// "yoo" -- unique, uniform, university -- and the trap in them is that "uni"
// is not the test: "unimportant" and "uninteresting" open the same way on
// paper and take "an". So the stems are spelled out rather than shortened.
// The "eu" words are the same sound from the other direction, and "one" and
// "once" open with a "w".
var consonantSoundedVowels = eitherCaseAlternation([]string{
	`uniqu\w*`, `unicorn\w*`, `unicycl\w*`, `unif\w*`, `union\w*`,
	`unit\w*`, `univers\w*`, `use\w*`, `usu\w*`, `usab\w*`, `usag\w*`,
	`usurp\w*`, `util\w*`, `utensil\w*`, `uter\w*`, `utop\w*`, `ubiq\w*`,
	`ukulele`, `ukelele`, `euro\w*`, `europ\w*`, `eulog\w*`, `euph\w*`,
	`eucalypt\w*`, `eunuch`, `one\b`, `one-\w+`, `once\b`, `ewe\b`,
})

// vowelSoundedConsonants are the words that open with a consonant letter and
// a vowel sound, so they take "an": the silent "h" words.
var vowelSoundedConsonants = eitherCaseAlternation([]string{
	`hour\w*`, `honest\w*`, `honour\w*`, `honor\w*`, `heir\w*`,
})

// initialisms are the letters whose *names* open with a vowel -- F is "ef",
// M is "em", X is "ex" -- so "an MBA" and "an X-ray" are right even though
// the letters are consonants.
// The trailing \b is what makes this cover the whole initialism rather
// than its first two letters. An antipattern has to *contain* the match it
// suppresses, and "an MB" does not contain "an MBA" -- which is exactly how
// this first failed.
const initialisms = `[FHLMNRSX][A-Z0-9]*\b`

var articleRules = []GrammarRule{
	{
		Pattern: `\b([Aa])\s+([AaEeIiOoUu]\w*)`,
		Antipatterns: []string{
			`\b[Aa]\s+(` + consonantSoundedVowels + `)`,
		},
		Message:     "Should be \"$1n $2\"",
		Suggest:     "$1n $2",
		Category:    "Grammar",
		Explanation: explainArticle,
	},
	{
		Pattern: `\b([Aa])n\s+([B-DF-HJ-NP-TV-Zb-df-hj-np-tv-z]\w*)`,
		Antipatterns: []string{
			`\b[Aa]n\s+(` + vowelSoundedConsonants + `)`,
			`\b[Aa]n\s+` + initialisms,
		},
		Message:     "Should be \"$1 $2\"",
		Suggest:     "$1 $2",
		Category:    "Grammar",
		Explanation: explainArticle,
	},
}

// eitherCaseAlternation joins stems into one alternation, letting each
// begin in either case.
//
// "European" is capitalised and "euro" is not, and both take "a". Dart's
// regex engine has no inline case-insensitivity -- it is an ECMAScript
// engine, and the flag is on the pattern object rather than in the source
// -- so the alternative is to write every stem twice. Doing it here means
// the lists above stay readable and cannot be half-converted.
func eitherCaseAlternation(stems []string) string {
	out := make([]string, 0, len(stems))
	for _, stem := range stems {
		first := rune(stem[0])
		lower, upper := unicode.ToLower(first), unicode.ToUpper(first)
		if lower == upper {
			out = append(out, stem)
			continue
		}
		out = append(out,
			"["+string(lower)+string(upper)+"]"+stem[1:])
	}
	return strings.Join(out, "|")
}
