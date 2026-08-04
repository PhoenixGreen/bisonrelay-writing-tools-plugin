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
	if skipped > 2 {
		t.Errorf("%d rules could not be compiled by RE2; expected only the "+
			"backreference ones", skipped)
	}
}

// TestRulesCatchTheirOwnMistake pairs each compilable rule with text it is
// supposed to flag, so a pattern edited into uselessness is noticed.
func TestRulesCatchTheirOwnMistake(t *testing.T) {
	// Only rules RE2 can compile appear here; "Repeated word" and "Repeated
	// punctuation" use backreferences and are exercised app-side instead.
	cases := map[string]string{
		"Multiple spaces":                 "hello  world",
		"Space before punctuation":        "hello , world",
		"Missing space after punctuation": "Stop.Now",
		"Excessive punctuation":           "really!!!",
		"Space inside bracket":            "( hello)",
		"\"a lot\" is two words":          "thanks alot",
		"Missing apostrophe":              "i cant do it",
		"Wordy":                           "due to the fact that it rained",
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
