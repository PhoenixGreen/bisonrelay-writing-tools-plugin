package thesaurus

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

// TestCoverageOverDictionary is the measurement the lemmatiser exists for:
// how much of the dictionary a lookup can actually answer for.
//
// Skipped unless the generated files are present, so it does not fail a
// checkout that has only the compressed ones.
//
// It asks for the wordlist by its real name. It used to look for
// "../words.txt", which stopped existing when the list was split per
// language -- so this skipped on every run and reported nothing, which is
// how the sense-ordering fault in tools/mkdefs went unnoticed: the one
// measurement that would have shown it had quietly turned itself off. A
// skipped test that looks like a passing one is worse than no test.
func TestCoverageOverDictionary(t *testing.T) {
	idx := loadIndex(t)
	words := loadWords(t, "../words-en-GB.txt")

	var answered, definitions, synonyms, examples int
	for _, w := range words {
		entry, ok := idx.Lookup(w)
		if ok {
			answered++
		}
		if len(entry.Definitions) > 0 {
			definitions++
		}
		if len(entry.Senses) > 0 {
			synonyms++
		}
		for _, d := range entry.Definitions {
			if d.Example != "" {
				examples++
				break
			}
		}
	}
	pct := func(n int) int { return n * 100 / len(words) }
	t.Logf("dictionary %d words: answered %d (%d%%), definitions %d (%d%%), "+
		"synonyms %d (%d%%), with an example %d (%d%%)",
		len(words), answered, pct(answered), definitions, pct(definitions),
		synonyms, pct(synonyms), examples, pct(examples))

	// Floors rather than exact figures: the point is to catch the data
	// silently getting worse -- a generator change that drops half the
	// senses, or a wordlist that grows past what WordNet covers -- without
	// failing every time a dataset is legitimately updated.
	if got := pct(definitions); got < 80 {
		t.Errorf("definition coverage fell to %d%%, floor is 80%%", got)
	}
	if got := pct(synonyms); got < 65 {
		t.Errorf("synonym coverage fell to %d%%, floor is 65%%", got)
	}
	if got := pct(examples); got < 25 {
		t.Errorf("example coverage fell to %d%%, floor is 25%%", got)
	}
}

// TestCommonSensesComeFirst pins the ordering fault this data had.
//
// Senses used to arrive in data-file order, which is byte offset and means
// nothing to a reader: "bank" led with a flight manoeuvre, "happy" with
// "well expressed and to the point". Both are WordNet's least common sense
// for the word. tools/mkdefs now walks index.<pos>, which is WordNet's own
// frequency rank.
func TestCommonSensesComeFirst(t *testing.T) {
	idx := loadIndex(t)
	for _, tc := range []struct{ word, want string }{
		{"bank", "sloping land"},
		{"happy", "joy or pleasure"},
		{"set", "group of things"},
		{"book", "written work"},
	} {
		entry, ok := idx.Lookup(tc.word)
		if !ok || len(entry.Definitions) == 0 {
			t.Errorf("%q: no definition", tc.word)
			continue
		}
		if got := entry.Definitions[0].Text; !strings.Contains(got, tc.want) {
			t.Errorf("%q leads with %q, expected something containing %q",
				tc.word, got, tc.want)
		}
	}
}

// A definition must not stop at the first semicolon. The gloss for bank's
// flight-manoeuvre sense is "a flight maneuver; aircraft tips laterally
// about its longitudinal axis (especially in turning)", and cutting at the
// semicolon left "a flight maneuver", which says nothing at all.
func TestDefinitionsAreNotTruncated(t *testing.T) {
	idx := loadIndex(t)
	entry, ok := idx.Lookup("bank")
	if !ok {
		t.Fatal("no entry for bank")
	}
	for _, d := range entry.Definitions {
		if d.Text == "a flight maneuver" {
			t.Error("bank's flight sense is still truncated at the semicolon")
		}
	}
}

func loadIndex(t *testing.T) *Index {
	t.Helper()
	read := func(name string) string {
		b, err := os.ReadFile(name)
		if err != nil {
			t.Skipf("generated data not present: %v", err)
		}
		return string(b)
	}
	return NewIndex(read("../thesaurus.txt"), read("../definitions.txt"),
		read("../exceptions.txt"))
}

func loadWords(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Skipf("wordlist not present: %v", err)
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		w := strings.TrimSpace(sc.Text())
		if w != "" && !strings.HasPrefix(w, "#") {
			out = append(out, w)
		}
	}
	return out
}
