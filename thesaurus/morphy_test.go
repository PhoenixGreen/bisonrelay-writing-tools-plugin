package thesaurus

import "testing"

// morphy_test.go covers the reduction from a word as typed to the form the
// data is keyed by. It is the difference between a lookup answering for 86%
// of the dictionary and answering for 36%, so the cases below are the ones
// that difference is made of.

// wordnet is a stand-in for the real data: only base forms, exactly as
// WordNet and MyThes store them, and sorted, because the index binary-
// searches line offsets rather than reading the file in.
const wordnet = `box|noun:a container with a flat base
bus|noun:a large road vehicle
cat|noun:a small domesticated feline
child|noun:a young human being
foot|noun:the lower extremity of the leg
go|verb:to move from one place to another
good|adj:having desirable qualities
happy|adj:enjoying or showing joy
mouse|noun:a small rodent
payment|noun:the act of paying
run|verb:to move fast on foot
stop|verb:to cease moving
walk|verb:to travel on foot
worry|verb:to feel anxious
`

const irregular = `better|good,well
children|child
feet|foot
mice|mouse
went|go
`

func lookup(t *testing.T, word string) (Entry, bool) {
	t.Helper()
	return NewIndex("", wordnet, irregular).Lookup(word)
}

func TestLookupReducesInflections(t *testing.T) {
	cases := map[string]string{
		// Regular plurals, including the spellings that need their own rule.
		"cats":  "cat",
		"buses": "bus",
		"boxes": "box",
		// -y stems.
		"worries":  "worry",
		"happier":  "happy",
		"happiest": "happy",
		// Participles, including the doubled consonant that is spelling
		// rather than meaning.
		"walking":  "walk",
		"walked":   "walk",
		"running":  "run",
		"stopped":  "stop",
		"stopping": "stop",
		// Irregulars, which no suffix rule reaches. These are the reason the
		// exception list is shipped at all.
		"went":     "go",
		"children": "child",
		"mice":     "mouse",
		"feet":     "foot",
		"better":   "good",
		// Possessives, which are 27,000 of the dictionary's entries.
		"payment's":  "payment",
		"cat's":      "cat",
		"children's": "child",
	}
	for typed, want := range cases {
		entry, ok := lookup(t, typed)
		if !ok {
			t.Errorf("Lookup(%q) found nothing, want the entry for %q", typed, want)
			continue
		}
		if entry.Word != want {
			t.Errorf("Lookup(%q).Word = %q, want %q", typed, entry.Word, want)
		}
	}
}

// The word itself always wins. A rule that fired ahead of a real entry would
// answer for the wrong word while looking as though it had worked.
func TestLookupPrefersTheWordItself(t *testing.T) {
	for _, word := range []string{"cat", "run", "good", "go"} {
		entry, ok := lookup(t, word)
		if !ok || entry.Word != word {
			t.Errorf("Lookup(%q) = %q, %v; want the word itself", word, entry.Word, ok)
		}
	}
}

// A reduction is only accepted if it is actually in the data, so an
// over-eager rule costs nothing.
func TestLookupRejectsGuessesThatMissed(t *testing.T) {
	// "bus" ends in "s" and would reduce to "bu" under the plural rule; the
	// word itself is in the data and wins. "gas" has no entry at all, and
	// neither does "ga", so the guess must not invent one.
	for _, word := range []string{"gas", "sing", "thing", "kencameron"} {
		if entry, ok := lookup(t, word); ok {
			t.Errorf("Lookup(%q) returned %+v, want no match", word, entry)
		}
	}
}

// An index with no exception list still reduces what the suffix rules can
// reach, so an older generated build degrades rather than breaking.
func TestLookupWithoutExceptions(t *testing.T) {
	idx := NewIndex("", wordnet, "")
	if entry, ok := idx.Lookup("cats"); !ok || entry.Word != "cat" {
		t.Errorf(`Lookup("cats") = %q, %v; want "cat"`, entry.Word, ok)
	}
	if _, ok := idx.Lookup("went"); ok {
		t.Error(`Lookup("went") found something without the exception list`)
	}
}
