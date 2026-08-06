package spellcheck

import (
	"regexp"
	"strings"
	"unicode"
)

// rules_style.go holds the opinions: phrases that could be shorter, phrases
// worn smooth by overuse, and the passive voice.
//
// Every rule here is marked SeveritySuggestion, and that mark is what makes
// the file possible at all. The rest of this plugin holds to a standard of
// firing only on text that is wrong whatever the writer meant, because an
// alarming red underline under correct writing teaches people to stop reading
// underlines. None of these meet that standard -- "in order to" is not an
// error, and sometimes it is the better phrasing.
//
// They are shipped anyway because the app underlines suggestions in a
// different colour and lists them on their own page, so the cost of being
// wrong is a quieter mark in a place the reader went looking for opinions.
// That is a different bargain from the one the error rules make, and these
// rules would not be acceptable under the other one.

// SeveritySuggestion marks a rule as an opinion rather than a mistake. Spelled
// out here rather than as a bare string at every rule so a typo in it -- which
// would silently promote a suggestion to an error -- cannot happen.
const SeveritySuggestion = "suggestion"

// wordy pairs a phrase with what it usually shortens to. Held as data rather
// than as rule literals because there are ninety of them and they differ in
// nothing but the two strings; written out longhand, the shape of each rule
// would bury the list itself.
//
// A pair belongs here only if the replacement is a fair paraphrase in most
// contexts. "Protest against" was tried and dropped: it is standard British
// English, not padding.
var wordy = [][2]string{
	// Phrases that say in several words what one word says.
	{"a large number of", "many"},
	{"a majority of", "most"},
	{"a number of", "several"},
	{"are able to", "can"},
	{"as a matter of fact", "in fact"},
	{"at the present time", "now"},
	{"at this point in time", "now"},
	{"by means of", "by"},
	{"come to the conclusion that", "conclude that"},
	{"despite the fact that", "although"},
	{"during the course of", "during"},
	{"for the purpose of", "for"},
	{"for the reason that", "because"},
	{"has the ability to", "can"},
	{"have the ability to", "can"},
	{"in a timely manner", "promptly"},
	{"in an effort to", "to"},
	{"in close proximity to", "near"},
	{"in excess of", "more than"},
	{"in light of the fact that", "because"},
	{"in regard to", "about"},
	{"in spite of the fact that", "although"},
	{"in the absence of", "without"},
	{"in the event that", "if"},
	{"in order to", "to"},
	{"in the near future", "soon"},
	{"is able to", "can"},
	{"it is important to note that", "note that"},
	{"make a decision", "decide"},
	{"on a daily basis", "daily"},
	{"on a regular basis", "regularly"},
	{"on account of", "because of"},
	{"owing to the fact that", "because"},
	{"prior to", "before"},
	{"subsequent to", "after"},
	{"take into consideration", "consider"},
	{"there is no doubt that", "undoubtedly"},
	{"until such time as", "until"},
	{"with regard to", "about"},
	{"with the exception of", "except"},

	// Phrases where one half repeats the other.
	{"absolutely essential", "essential"},
	{"actual fact", "fact"},
	{"added bonus", "bonus"},
	{"advance planning", "planning"},
	{"basic fundamentals", "fundamentals"},
	{"brief summary", "summary"},
	{"close proximity", "proximity"},
	{"collaborate together", "collaborate"},
	{"completely eliminate", "eliminate"},
	{"connect together", "connect"},
	{"each and every", "each"},
	{"end result", "result"},
	{"exact same", "same"},
	{"few in number", "few"},
	{"first and foremost", "first"},
	{"free gift", "gift"},
	{"future plans", "plans"},
	{"general consensus", "consensus"},
	{"join together", "join"},
	{"may possibly", "may"},
	{"merge together", "merge"},
	{"mutual cooperation", "cooperation"},
	{"new innovation", "innovation"},
	{"past experience", "experience"},
	{"past history", "history"},
	{"personal opinion", "opinion"},
	{"plan ahead", "plan"},
	{"postpone until later", "postpone"},
	{"repeat again", "repeat"},
	{"return back", "return"},
	{"revert back", "revert"},
	{"rise up", "rise"},
	{"still remains", "remains"},
	{"sum total", "total"},
	{"unexpected surprise", "surprise"},
	{"usual custom", "custom"},
	{"12 midnight", "midnight"},
	{"12 noon", "noon"},
	{"ATM machine", "ATM"},
	{"PIN number", "PIN"},
	{"LCD display", "LCD"},
}

// cliches are phrases that have been used so often they no longer carry the
// image they were built on. No replacement is offered: there is no single
// rewrite, and proposing one would be a worse cliche than the original.
var cliches = []string{
	"at the end of the day",
	"think outside the box",
	"low-hanging fruit",
	"moving forward",
	"going forward",
	"touch base",
	"circle back",
	"at this juncture",
	"needless to say",
	"last but not least",
	"in this day and age",
	"the bottom line is",
	"win-win",
	"take it to the next level",
	"boots on the ground",
	"push the envelope",
	"when all is said and done",
	"few and far between",
	"tip of the iceberg",
	"par for the course",
	"back to the drawing board",
	"a level playing field",
	"the elephant in the room",
	"raise the bar",
	"move the needle",
}

// participles are past participles distinctive enough that "was <participle>
// by" cannot be anything but the passive. Irregular forms only: an -ed word
// after "was" is far more often an adjective ("was tired", "was interested")
// than a passive verb.
const participles = `written|taken|given|made|sent|held|chosen|brought|sold|` +
	`spoken|built|driven|seen|done|known|paid|met|kept|left|told|taught|` +
	`understood|won|run|read|put|caught|bought|found|felt|meant|thought|` +
	`said|eaten|broken|shown|worn|torn|drawn|grown|thrown`

// styleRules are the opinions, built from the tables above plus the two
// passive-voice checks.
var styleRules = buildStyleRules()

func buildStyleRules() []GrammarRule {
	rules := make([]GrammarRule, 0, len(wordy)+len(cliches)+2)

	for _, pair := range wordy {
		rules = append(rules, GrammarRule{
			Pattern:  `\b` + eitherCase(pair[0]) + `\b`,
			Message:  "Wordy -- try \"" + pair[1] + "\"",
			Suggest:  pair[1],
			Category: "Style",
			Explanation: "\"" + pair[0] + "\" can usually be shortened to \"" +
				pair[1] + "\" without losing anything. Shorter is easier to " +
				"read, though the longer form is sometimes the one you want.",
			Severity: SeveritySuggestion,
		})
	}

	for _, phrase := range cliches {
		rules = append(rules, GrammarRule{
			// Hyphens and other punctuation are escaped; the \b at the end is
			// dropped for phrases ending in a non-word character, where it
			// would never match.
			Pattern:  `\b` + eitherCase(phrase) + wordBoundary(phrase),
			Message:  "Overused phrase",
			Suggest:  "",
			Category: "Style",
			Explanation: "\"" + phrase + "\" is used so often that it no " +
				"longer carries the picture it was built on. Saying it " +
				"plainly usually lands better.",
			Severity: SeveritySuggestion,
		})
	}

	rules = append(rules,
		GrammarRule{
			// Passive with a named agent. The "by" is what makes this
			// certain, and requiring it is a deliberate trade: most passive
			// writing names no agent and goes unflagged here, which is the
			// price of never flagging "I was interested" as passive.
			Pattern:  `\b(is|are|was|were|be|been|being)\s+(` + participles + `)\s+by\b`,
			Message:  "Passive voice",
			Suggest:  "",
			Category: "Style",
			Explanation: "This sentence puts the thing acted on first and " +
				"the actor last. Naming the actor first is usually shorter " +
				"and clearer -- though the passive is right when the actor " +
				"is unknown or beside the point.",
			Severity: SeveritySuggestion,
		},
		GrammarRule{
			// "Is being considered" is unambiguously passive with or without
			// an agent: "being" before a participle admits no adjective
			// reading.
			Pattern:  `\b(is|are|was|were)\s+being\s+(\w+ed|` + participles + `)\b`,
			Message:  "Passive voice",
			Suggest:  "",
			Category: "Style",
			Explanation: "This sentence puts the thing acted on first and " +
				"the actor last. Naming the actor first is usually shorter " +
				"and clearer -- though the passive is right when the actor " +
				"is unknown or beside the point.",
			Severity: SeveritySuggestion,
		},
	)

	return rules
}

// eitherCase escapes a phrase for use in a pattern and lets its first letter
// match in either case, so a phrase that opens a sentence is caught.
//
// These rules matter most at the start of a sentence -- "Due to the fact
// that" and "At the end of the day" are where these phrases live -- and every
// pattern in this plugin is case-sensitive, so without this the commonest
// position for each of them is the one position it never fires in.
//
// Only the first letter: matching the rest either way would flag SHOUTED text
// and acronyms embedded in words.
func eitherCase(phrase string) string {
	quoted := regexp.QuoteMeta(phrase)
	first := rune(phrase[0])
	lower, upper := unicode.ToLower(first), unicode.ToUpper(first)
	if lower == upper {
		return quoted
	}
	return "[" + string(lower) + string(upper) + "]" + quoted[1:]
}

// wordBoundary returns `\b` only where it can match: a phrase ending in a
// non-word character ("win-win" does not, but a hyphenated one might) has no
// word boundary after it, and appending one there makes a rule that can never
// fire -- which nothing else would notice.
func wordBoundary(phrase string) string {
	if phrase == "" {
		return ""
	}
	last := phrase[len(phrase)-1]
	isWord := last == '_' ||
		(last >= 'a' && last <= 'z') ||
		(last >= 'A' && last <= 'Z') ||
		(last >= '0' && last <= '9')
	if isWord {
		return `\b`
	}
	return ""
}

// variantPairs are spellings of one word that are both correct but should not
// be mixed inside a single message. Ordered British first, though the check
// takes no side: it reports that two were used, not that either is wrong.
//
// Pairs whose two spellings mean different things are deliberately absent --
// licence/license and practise/practice differ by part of speech in British
// English, programme/program and enquiry/inquiry differ in sense, and metre
// and meter are separate words. Flagging those would be flagging correct
// writing.
var variantPairs = []string{
	"organise|organize",
	"organised|organized",
	"organisation|organization",
	"realise|realize",
	"recognise|recognize",
	"analyse|analyze",
	"apologise|apologize",
	"summarise|summarize",
	"prioritise|prioritize",
	"customise|customize",
	"optimise|optimize",
	"utilise|utilize",
	"emphasise|emphasize",
	"criticise|criticize",
	"specialise|specialize",
	"minimise|minimize",
	"maximise|maximize",
	"colour|color",
	"colours|colors",
	"favourite|favorite",
	"behaviour|behavior",
	"honour|honor",
	"labour|labor",
	"neighbour|neighbor",
	"centre|center",
	"defence|defense",
	"grey|gray",
	"theatre|theater",
	"fibre|fiber",
	"litre|liter",
	"travelled|traveled",
	"cancelled|canceled",
	"catalogue|catalog",
	"dialogue|dialog",
	"aluminium|aluminum",
	"judgement|judgment",
	"acknowledgement|acknowledgment",
}

// splitVariant is here so the pair format has exactly one reader; the app
// parses the same strings and the two must agree on the separator.
func splitVariant(pair string) (string, string, bool) {
	british, american, found := strings.Cut(pair, "|")
	if !found || british == "" || american == "" {
		return "", "", false
	}
	return british, american, true
}
