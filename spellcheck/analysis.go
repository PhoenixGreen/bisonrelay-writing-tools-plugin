package spellcheck

// analysis.go declares the checks that count instead of matching.
//
// A regex can say "this text contains X". It cannot say "this word appears
// four times in this paragraph", "this sentence runs to fifty words" or
// "these two spellings of one word are both in use" -- each needs to look at
// the whole message and keep a tally, which is not what a pattern does.
//
// So these are declared rather than written: the plugin names a check the app
// knows how to run, and supplies everything else about it. That keeps the
// division this plugin is built on -- the provider decides what is checked,
// the app owns every mechanism -- rather than moving logic into the app under
// the plugin's name. It is the same arrangement the regex rules already have,
// where the plugin writes patterns it never executes.
//
// An app that does not implement one of these ids ignores it, so a check can
// ship here before the app that runs it.

// The check ids the app implements. Named constants because a typo in one is
// silently ignored by design -- the whole point of ignoring unknown ids -- so
// nothing at runtime would report it.
const (
	// CheckRepeatedWord fires when one word is used Threshold or more times
	// in a single paragraph. $1 is the word, $2 the count.
	CheckRepeatedWord = "repeated-word-in-paragraph"

	// CheckLongSentence fires on a sentence of Threshold or more words. $2 is
	// the length.
	CheckLongSentence = "long-sentence"

	// CheckRepeatedOpener fires when Threshold consecutive sentences begin
	// with the same word. $1 is the word, $2 the run length.
	CheckRepeatedOpener = "repeated-sentence-opener"

	// CheckSpellingVariants fires when both spellings of a Values pair appear
	// in one message. $1 is the spelling flagged, $2 the other one.
	CheckSpellingVariants = "spelling-variant-inconsistency"
)

// AnalysisCheck is one declared check. It mirrors the spellcheck-data
// capability's JSON schema exactly; this plugin is a separate Go module and
// deliberately takes no dependency on Bison Relay to get it.
type AnalysisCheck struct {
	ID          string   `json:"id"`
	Threshold   int      `json:"threshold,omitempty"`
	Message     string   `json:"message"`
	Category    string   `json:"category,omitempty"`
	Explanation string   `json:"explanation,omitempty"`
	Severity    string   `json:"severity,omitempty"`
	Values      []string `json:"values,omitempty"`
}

// AnalysisChecks are the counting checks this plugin asks for, with the
// thresholds it thinks are right.
//
// The thresholds are the whole argument in each case, and each is set below
// where a tool would usually put it. A check that fires on ordinary writing
// is worse than no check: the reader learns to skip that page.
var AnalysisChecks = []AnalysisCheck{
	{
		ID: CheckRepeatedWord,
		// Four, not three. Three uses of a word in a paragraph is often just
		// what the paragraph is about -- a paragraph on payments will say
		// "payment" three times and be right to. By four it reads as a rut.
		Threshold:   4,
		Message:     "\"$1\" used $2 times in this paragraph",
		Category:    "Repetition",
		Severity:    SeveritySuggestion,
		Explanation: "Repeating a word in a short space makes writing feel flat. A synonym, a pronoun, or dropping the word altogether usually reads better -- though a technical term is often best repeated rather than varied for its own sake.",
	},
	{
		ID: CheckLongSentence,
		// Thirty rather than the forty most tools use. Forty catches only
		// sentences already past saving; thirty catches the ones still worth
		// splitting, which is the point of saying anything.
		Threshold:   30,
		Message:     "Long sentence -- $2 words",
		Category:    "Readability",
		Severity:    SeveritySuggestion,
		Explanation: "A sentence this long asks the reader to hold a lot at once. Splitting it at a natural break usually makes both halves easier to follow.",
	},
	{
		ID:          CheckRepeatedOpener,
		Threshold:   3,
		Message:     "$2 sentences in a row start with \"$1\"",
		Category:    "Repetition",
		Severity:    SeveritySuggestion,
		Explanation: "Consecutive sentences opening the same way give writing a monotonous rhythm. Varying how a sentence starts is usually enough to fix it.",
	},
	{
		ID:          CheckSpellingVariants,
		Threshold:   0,
		Message:     "\"$1\" and \"$2\" are both used here",
		Category:    "Consistency",
		Severity:    SeveritySuggestion,
		Explanation: "Both spellings are correct, but mixing them in one message looks careless. Pick whichever you prefer and use it throughout.",
		Values:      VariantPairs,
	},
}
