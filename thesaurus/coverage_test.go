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
func TestCoverageOverDictionary(t *testing.T) {
	idx := loadIndex(t)
	words := loadWords(t, "../words.txt")

	var answered, definitions, synonyms int
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
	}
	pct := func(n int) int { return n * 100 / len(words) }
	t.Logf("dictionary %d words: answered %d (%d%%), definitions %d (%d%%), synonyms %d (%d%%)",
		len(words), answered, pct(answered), definitions, pct(definitions),
		synonyms, pct(synonyms))
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
