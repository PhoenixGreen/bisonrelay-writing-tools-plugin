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

	// The split-compound rules. Each of these is the reading that keeps a
	// pair of words apart, and each one cost a rule its guard.
	"My self-esteem took a knock.",
	"Your self-esteem matters too.",
	"It gave him self-confidence.",
	"We watched it self-destruct.",
	"Her self-control is remarkable.",
	"I like my true self better.",
	"Every body in the room turned around.",
	"That would be cause for concern.",
	"Interest is charged on the whole amount, be cause for celebration or not.",
	"We parked in side streets all week.",
	"It came with out-of-date firmware.",
	"We shipped it with out of date documentation.",
	"Send it to me to review before Friday.",
	"There is a part of this I still do not follow.",

	// Punctuation. Numbers are the trap that runs through all of it.
	"It cost 1,000 DCR at 3:00 on 3.5 percent.",
	"The file is at v1.2.3, not v1.2.4.",
	"The loop runs 0..10 in Rust.",

	// a/an. The first line is every family the "a" rule must not touch, the
	// second the same for "an", and both are correct writing that the rule
	// as usually stated -- "an before a vowel" -- would break.
	"It was a unique opportunity, a one-off, and a euro-denominated unit.",
	"A university course on a European utility is a useful thing once a year.",
	"We waited an hour for an honest answer from an heir.",
	"She has an MBA, an SMS alert and an X-ray booked.",
	"An unimportant detail took an uninteresting hour.",
	"That is an odd effect on an old idea.",
	"The plan is to close the door and to narrow the gap.",
	"No one knows the answer yet.",
	"No longer a problem.",
	"Yes and no, depending.",
	"However you do it, the result is the same.",
	"Similarly designed products failed the same way.",
	"For example sentences, see the appendix.",
	"Wait for it... there it is.",
	"She said, \"that works for me\".",
	"Therefore, we should wait.",
	"Meanwhile, the others kept going.",
	"I went but he stayed.",
}

// TestRulesDoNotFireOnCorrectText is the false-positive guard, and it is the
// most valuable test here. A wavy underline under correct writing is worse
// than a missed error, because it teaches people to stop reading the
// underlines at all.
//
// It covers the error rules. The suggestions are held to a different
// standard on purpose -- see the note at the top of rules_style.go.
//
// Rules using backreferences are skipped: they are written in Dart's dialect
// and Go's RE2 cannot compile them by design. Those are covered by the app's
// own tests, where they actually run.
func TestRulesDoNotFireOnCorrectText(t *testing.T) {
	skipped := 0
	for _, r := range Rules {
		// Errors only. A suggestion is an opinion about writing that is not
		// wrong -- that is what the severity means -- so holding one to this
		// standard would mean no sentence in the corpus below could contain
		// a cliche, a wordy phrase or a missing optional comma. The corpus
		// is meant to be ordinary writing, and ordinary writing contains all
		// three.
		if r.Severity == SeveritySuggestion {
			continue
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			// RE2 rejects it. That is expected for the constructs Dart has
			// and Go does not -- but only for those, so the reason is
			// checked rather than counted.
			if !usesDartOnlySyntax(r.Pattern) {
				t.Errorf("rule %q does not compile and uses nothing Dart-only: %v",
					r.Message, err)
			}
			skipped++
			continue
		}
		// The antipatterns are part of the rule, so they are part of the
		// test. Without this a rule that moved its guard out of the pattern
		// -- which is the point of having them -- would read as a new false
		// positive on text it has always handled correctly.
		var exceptions []*regexp.Regexp
		for _, source := range r.Antipatterns {
			anti, err := regexp.Compile(source)
			if err != nil {
				t.Errorf("rule %q: antipattern %q does not compile: %v",
					r.Message, source, err)
				continue
			}
			exceptions = append(exceptions, anti)
		}

		for _, text := range correctText {
			for _, at := range re.FindAllStringIndex(text, -1) {
				if suppressed(exceptions, text, at) {
					continue
				}
				t.Errorf("rule %q fired on correct text %q (matched %q)",
					r.Message, text, text[at[0]:at[1]])
			}
		}
	}
	// Every skip is accounted for above rather than counted here. A count
	// was a magic number that had to be raised each time a lookaround rule
	// was added, and raising it is exactly the moment nobody checks what
	// else slipped past -- a pattern with a genuine typo in it would have
	// been waved through as one more expected skip.
	t.Logf("%d of %d rules are beyond RE2 and are covered app-side, where "+
		"they run", skipped, len(Rules))
}

// suppressed reports whether an antipattern covers the match at [start,end).
//
// Contained, not merely overlapping: an antipattern describes a longer
// reading that the match is part of -- "my self" inside "my self-esteem" --
// so a pattern that happens to clip the edge of one is not that reading.
func suppressed(exceptions []*regexp.Regexp, text string, at []int) bool {
	for _, anti := range exceptions {
		for _, e := range anti.FindAllStringIndex(text, -1) {
			if e[0] <= at[0] && e[1] >= at[1] {
				return true
			}
		}
	}
	return false
}

// dartOnlySyntax is the constructs Dart's regex engine has and Go's RE2 does
// not, by design: RE2 guarantees linear time and neither can be done in it.
var dartOnlySyntax = regexp.MustCompile(`\(\?<?[=!]|\\[1-9]`)

func usesDartOnlySyntax(pattern string) bool {
	return dartOnlySyntax.MatchString(pattern)
}

// TestRulesCatchTheirOwnMistake is gone, and TestRuleExamples replaced it.
//
// It paired rules with text by *message*, in a map, and messages are not
// unique: eight rules say `Should be "$1 effect"` and a map keeps one of
// them. So it silently covered a fraction of what it appeared to, and the
// fraction shrank every time a rule was added. Examples live on the rule
// now, so every rule is covered exactly once and by construction.

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

// TestAntipatternsAreReachable: every antipattern must match one of its own
// rule's Leaves examples.
//
// An exception that matches nothing is one nobody notices has stopped
// working: the rule simply gets noisier, and the only evidence is a false
// positive somebody has to report. This is the check a negative lookahead
// could never have -- glued onto the end of a pattern, there is nothing to
// test on its own.
//
// Against the rule's own examples rather than the shared corpus, which is
// where this started. The corpus answers a different question -- does any
// rule fire on ordinary writing -- and making it carry a sentence for every
// exception in the plugin turned it into a list whose reason for each line
// lived somewhere else. It also subsumes the separate "every antipattern has
// a reason" check: a rule with antipatterns and no Leaves now fails here, by
// name.
func TestAntipatternsAreReachable(t *testing.T) {
	for i, r := range Rules {
		for _, source := range r.Antipatterns {
			anti, err := regexp.Compile(source)
			if err != nil {
				t.Errorf("Rules[%d] (%q): antipattern %q does not compile: %v",
					i, r.Message, source, err)
				continue
			}
			var matched bool
			for _, text := range r.Leaves {
				if anti.MatchString(text) {
					matched = true
					break
				}
			}
			if !matched {
				t.Errorf("Rules[%d] (%q): antipattern %q matches none of the "+
					"rule's own Leaves examples -- add the reading it is for, "+
					"or the exception is dead", i, r.Message, source)
			}
		}
	}
}

func TestRuleExamples(t *testing.T) {
	for i, r := range Rules {
		if len(r.Flags) == 0 {
			t.Errorf("Rules[%d] (%q) carries no example of what it catches",
				i, r.Message)
		}

		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			// One of the handful Dart has constructs for and RE2 does not.
			// Its examples cannot run here; those rules are exercised
			// individually in the app's own tests, where they do run.
			if !usesDartOnlySyntax(r.Pattern) {
				t.Errorf("Rules[%d] (%q) does not compile: %v", i, r.Message, err)
			}
			continue
		}

		var exceptions []*regexp.Regexp
		for _, source := range r.Antipatterns {
			anti, err := regexp.Compile(source)
			if err != nil {
				t.Errorf("Rules[%d] (%q): antipattern %q does not compile: %v",
					i, r.Message, source, err)
				continue
			}
			exceptions = append(exceptions, anti)
		}

		// fires reports whether the rule, exceptions and all, has anything
		// to say about text -- which is what the app will do with it.
		fires := func(text string) bool {
			for _, at := range re.FindAllStringIndex(text, -1) {
				if !suppressed(exceptions, text, at) {
					return true
				}
			}
			return false
		}

		for _, text := range r.Flags {
			if !fires(text) {
				t.Errorf("Rules[%d] (%q) does not catch %q, which it is for",
					i, r.Message, text)
			}
		}
		for _, text := range r.Leaves {
			if fires(text) {
				t.Errorf("Rules[%d] (%q) fires on %q, which it must not",
					i, r.Message, text)
			}
		}
	}
}

// expandSuggestion fills a rule's Suggest template from a match, the way the
// app does: $1..$9 are the capture groups and $U1..$U9 are the same groups
// with the first letter capitalised.
func expandSuggestion(template, text string, at []int) string {
	var out strings.Builder
	for i := 0; i < len(template); i++ {
		if template[i] != '$' || i+1 >= len(template) {
			out.WriteByte(template[i])
			continue
		}
		upper := template[i+1] == 'U'
		digit := i + 1
		if upper {
			digit = i + 2
		}
		if digit >= len(template) || template[digit] < '1' || template[digit] > '9' {
			out.WriteByte(template[i])
			continue
		}
		n := int(template[digit] - '0')
		group := ""
		if 2*n+1 < len(at) && at[2*n] >= 0 {
			group = text[at[2*n]:at[2*n+1]]
		}
		if upper && group != "" {
			group = strings.ToUpper(group[:1]) + group[1:]
		}
		out.WriteString(group)
		i = digit
	}
	return out.String()
}

// TestRuleFixesResolveTheRule applies each rule's own suggestion to each of
// its examples and checks the rule then has nothing left to say.
//
// A fix that does not fix is the most embarrassing thing a checker can do,
// and it is invisible to every other test here: the rule fires on the wrong
// text, which is all the examples ask of it, and then offers a replacement
// that is just as wrong. Two of these shipped. "a affect" was corrected to
// "a effect" because the rule carried the determiner through untouched, and
// a blanket case-folding once left the weekday rules asking for a capital on
// "Tuesday", which already had one.
//
// The rules that legitimately have no fix -- the cliches and the passive,
// where the replacement depends on what the sentence is trying to say -- are
// skipped rather than exempted, because they carry no suggestion at all.
func TestRuleFixesResolveTheRule(t *testing.T) {
	for i, r := range Rules {
		if r.Suggest == "" {
			continue
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			continue // Dart-only; covered app-side.
		}
		var exceptions []*regexp.Regexp
		for _, source := range r.Antipatterns {
			if anti, err := regexp.Compile(source); err == nil {
				exceptions = append(exceptions, anti)
			}
		}

		for _, text := range r.Flags {
			at := re.FindStringSubmatchIndex(text)
			if at == nil || suppressed(exceptions, text, at[:2]) {
				continue // TestRuleExamples already reports this.
			}
			replacement := expandSuggestion(r.Suggest, text, at)
			fixed := text[:at[0]] + replacement + text[at[1]:]
			if fixed == text {
				t.Errorf("Rules[%d] (%q): the fix for %q changes nothing",
					i, r.Message, text)
				continue
			}
			// Only what was replaced is the fix's business. An example is
			// allowed to carry a second fault of the same kind -- "she
			// speaks french and german" does -- and the rule firing again
			// further along is it working, not failing.
			from, to := at[0], at[0]+len(replacement)
			for _, still := range re.FindAllStringIndex(fixed, -1) {
				if still[0] < to && still[1] > from &&
					!suppressed(exceptions, fixed, still) {
					t.Errorf("Rules[%d] (%q): fixing %q gives %q, which the "+
						"rule still fires on", i, r.Message, text, fixed)
					break
				}
			}
		}
	}
}
