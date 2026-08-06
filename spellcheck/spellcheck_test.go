package spellcheck

import (
	"regexp"
	"strings"
	"testing"
)

func TestParseWordsSkipsBlanksAndComments(t *testing.T) {
	in := "# a comment\nHello\n\n  World  \n# another comment\nfoo\n"
	got := ParseWords(in)
	want := []string{"hello", "world", "foo"}
	if len(got) != len(want) {
		t.Fatalf("ParseWords(%q) = %v, want %v", in, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ParseWords(%q)[%d] = %q, want %q", in, i, got[i], want[i])
		}
	}
}

// countGroups counts the capturing groups in a pattern: an unescaped "("
// that does not open a non-capturing "(?...)" construct.
func countGroups(pattern string) int {
	n := 0
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '\\':
			i++ // skip whatever is escaped
		case '[':
			// Character classes may contain a literal "(".
			for i < len(pattern) && pattern[i] != ']' {
				if pattern[i] == '\\' {
					i++
				}
				i++
			}
		case '(':
			if i+1 < len(pattern) && pattern[i+1] == '?' {
				continue
			}
			n++
		}
	}
	return n
}

// TestRulesSuggestReferencesExist is the check that matters most for a rule
// file edited by hand: a "$2" in a rule with one group produces a suggestion
// with a literal "$2" in it, and nothing at build or load time notices.
func TestRulesSuggestReferencesExist(t *testing.T) {
	ref := regexp.MustCompile(`\$(\d+)`)
	for i, r := range Rules {
		groups := countGroups(r.Pattern)
		for _, m := range ref.FindAllStringSubmatch(r.Suggest, -1) {
			var n int
			for _, c := range m[1] {
				n = n*10 + int(c-'0')
			}
			if n > groups {
				t.Errorf("Rules[%d] (%q): Suggest references $%d but Pattern has %d groups",
					i, r.Message, n, groups)
			}
		}
	}
}

func TestRulesAreWellFormed(t *testing.T) {
	if len(Rules) == 0 {
		t.Fatal("Rules is empty")
	}
	seen := map[string]bool{}
	for i, r := range Rules {
		if r.Pattern == "" {
			t.Errorf("Rules[%d] has an empty pattern", i)
		}
		if r.Message == "" {
			t.Errorf("Rules[%d] (%q) has no message", i, r.Pattern)
		}
		if seen[r.Pattern] {
			t.Errorf("Rules[%d]: duplicate pattern %q", i, r.Pattern)
		}
		seen[r.Pattern] = true
	}
}

// correctText is ordinary, correct writing of the kind these rules will be
// run against constantly. Nothing here should be flagged.
var correctText = []string{
	"I sent the payment yesterday and it cleared this morning.",
	"Let's meet at 3pm. That works for me.",
	"The channel is open, so the invoice should route fine.",
	"Have a look at example.com when you get a chance.",
	"It cost 3.5 DCR, which isn't bad at all.",
	"He said (quite clearly) that he wouldn't be joining us.",
	"Can't wait -- that's excellent news!",
	"What's the plan for the release? I'd like to help.",
	"We should have tested it first; now we know.",
	"I'm going to the shop, and then I'll head home.",
	"Fine. Next question, then.",
	"see docs.rs and news.ycombinator.com for details",
	// its/it's used correctly, which the new rules must not touch.
	"The channel lost its funding, and it's closed now.",
	"Its balance is low, but its owner has more.",
	"It's going to take a while.",
	// Possessives before nouns that look like the words above: each of these
	// is correct and an over-eager rule would flag it.
	"Its true value is hard to judge.",
	"Its cold storage keeps the keys offline.",
	"Its early adopters were patient.",
	"Its best feature is the relay.",
	"Its interesting parts are undocumented.",
	"Its going rate is higher than that.",
	"Its freezing point is well below zero.",
	"Its only purpose is to relay messages.",
	"e.g. this one, i.e. that one",
	// their/there used correctly, which the new confusion rules must not
	// touch: a possessive before a noun, and an existential before a verb.
	"Their wallet is empty, so there is nothing to send.",
	"There are two of them, and their channels are both open.",
	"They brought their own hardware to the meetup.",
	"I told them there would be a delay.",

	// then/than used correctly, in both directions.
	"I would rather walk than drive.",
	"Other than that, it went fine.",
	"We synced the wallet and then sent the payment.",
	"Back then the fees were higher.",

	// lose/loose. The adjective is the whole reason the rule needs an object
	// after it before it fires.
	"The connector is loose, so check it before you plug in.",
	"He handed over a loose collection of notes.",
	"I did not want to lose the channel.",

	// affect/effect. "Effect" as a verb is real; so is "affect" as a noun in
	// clinical writing, though that one is rarer than the typo.
	"The upgrade will effect change across the network.",
	"The fee had no effect on routing.",
	"Latency does affect the user experience.",

	// your/you're. Every one of these was flagged by an earlier draft of the
	// rule, and every one of them is correct.
	"It is your right to refuse.",
	"Please confirm your correct address.",
	"We received your late payment yesterday.",
	"Your next appointment is on Tuesday.",
	"Thanks for your amazing work on the relay.",
	"Your invited guests can bring one other person.",
	"Your mistaken belief is understandable.",

	// to/too. The bare "to" before an adjective is fine; only the leading
	// verb makes it wrong.
	"I spoke to many people about it.",
	"We went to great lengths to fix it.",
	"That is to be expected at this stage.",

	// in principal. "Principal" is also an adjective.
	"The bank operates in principal cities only.",
	"Interest is charged on principal amounts over 1 DCR.",
	"I agree in principle with the proposal.",

	// have + participle, done properly.
	"They have gone quiet since the release.",
	"We have run the migration twice.",
	"She has read the specification.",

	// Words the confusion rules cover, used the right way round.
	"Whose keys are these?",
	"Who's going to review it?",
	"Please advise us on the best approach.",
	"Thanks for the advice about routing.",
	"Bear with me while I sync.",
	"That was a part of the original design.",
}

// TestRulesDoNotFireOnCorrectText is the false-positive guard. A wavy
// underline under correct writing is worse than a missed error, because it
// teaches people to stop reading the underlines at all.
//
// Rules using backreferences are skipped: they are written in Dart's dialect
// and Go's RE2 cannot compile them by design. Those are covered by the app's
// own tests, where they actually run.
func TestRulesDoNotFireOnCorrectText(t *testing.T) {
	skipped := 0
	for _, r := range Rules {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			skipped++
			continue
		}
		for _, text := range correctText {
			if m := re.FindString(text); m != "" {
				t.Errorf("rule %q fired on correct text %q (matched %q)",
					r.Message, text, m)
			}
		}
	}
	// Eight rules are beyond RE2 by design: three use backreferences
	// (repeated word, repeated punctuation, excessive punctuation) and five
	// use lookarounds (sentence capital, missing space, the "I" pronoun, "me
	// to" at the end of a sentence, and "in principal" at the end of a
	// clause). They are written in Dart's dialect and are covered by the
	// app's tests, where they actually run.
	if skipped > 8 {
		t.Errorf("%d rules could not be compiled by RE2; expected only the "+
			"backreference and lookaround ones", skipped)
	}
}

// TestRulesCatchTheirOwnMistake pairs each compilable rule with text it is
// supposed to flag, so a pattern edited into uselessness is noticed.
func TestRulesCatchTheirOwnMistake(t *testing.T) {
	// Only rules RE2 can compile appear here; "Repeated word" and "Repeated
	// punctuation" use backreferences and are exercised app-side instead.
	cases := map[string]string{
		"Multiple spaces":          "hello  world",
		"Space before punctuation": "hello , world",
		"Space inside bracket":     "( hello)",
		"\"a lot\" is two words":   "thanks alot",
		"Missing apostrophe":       "i cant do it",
		// Keyed by the rule's message template, not the expanded text: a
		// message may reference the pattern's capture groups.
		"Should be \"there $1\"": "their is a problem",
		"Should be \"their $1\"": "they brought there own",
		"Should be \"$1t's $2\"": "its going to rain",
		"Should be \"$1ts $2\"":  "it's own fault",
	}
	for message, text := range cases {
		var fired bool
		for _, r := range Rules {
			if r.Message != message && !strings.HasPrefix(r.Message, message) {
				continue
			}
			re, err := regexp.Compile(r.Pattern)
			if err != nil {
				continue
			}
			if re.MatchString(text) {
				fired = true
				break
			}
		}
		if !fired {
			t.Errorf("no rule with message %q flagged %q", message, text)
		}
	}
}

// categories is the closed set the app groups rules by. Adding one is a
// deliberate act -- the popup heading it appears under is written to match --
// so a typo in an existing name must not quietly create a new group.
var categories = map[string]bool{
	"Spacing":        true,
	"Punctuation":    true,
	"Capitalization": true,
	"Grammar":        true,
	"Confused words": true,
	"Style":          true,
	"Repetition":     true,
	"Readability":    true,
	"Consistency":    true,
}

// TestRulesAreExplained guards the thing a new rule is likeliest to be
// missing. A rule without an explanation still works -- the app falls back to
// the message alone -- so nothing breaks visibly, and the gap survives.
func TestRulesAreExplained(t *testing.T) {
	for i, r := range Rules {
		if !categories[r.Category] {
			t.Errorf("Rules[%d] (%q): category %q is not one of the known groups",
				i, r.Message, r.Category)
		}
		if r.Explanation == "" {
			t.Errorf("Rules[%d] (%q) has no explanation", i, r.Message)
			continue
		}
		if !strings.HasSuffix(r.Explanation, ".") {
			t.Errorf("Rules[%d] (%q): explanation is not a sentence: %q",
				i, r.Message, r.Explanation)
		}
		// An explanation no longer than the message is repeating it rather
		// than explaining it, which is worse than none: it takes up the space
		// where the reader expected to learn something.
		if len(r.Explanation) <= len(r.Message) {
			t.Errorf("Rules[%d] (%q): explanation adds nothing: %q",
				i, r.Message, r.Explanation)
		}
	}
}

// Rules sharing a message must at least agree on which group they belong to.
//
// Only the category, not the explanation. Explanations were once required to
// match as well, on the reasoning that two rules saying the same thing should
// say it the same way -- but the generated style rules quote the phrase they
// caught ("\"in regard to\" can usually be shortened to \"about\""), which is
// more use than a shared sentence would be. A category that disagrees is
// still a mistake: it decides which page the issue is listed on.
func TestSharedMessagesShareTheirCategory(t *testing.T) {
	seen := map[string]string{}
	for i, r := range Rules {
		if was, ok := seen[r.Message]; ok && was != r.Category {
			t.Errorf("Rules[%d] (%q) is in category %q, but an earlier rule "+
				"with the same message is in %q", i, r.Message, r.Category, was)
		}
		seen[r.Message] = r.Category
	}
}
