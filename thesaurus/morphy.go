package thesaurus

import "strings"

// morphy.go reduces a word as typed to the form the datasets are keyed by.
//
// This is the difference between a thesaurus that answers and one that mostly
// does not. WordNet and MyThes both store base forms only -- "cat" and not
// "cats", "run" and not "running", "happy" and not "happier" -- so of the
// 121,000 words the dictionary knows, only 77,000 have an entry to look up
// directly. The other 44,000 are almost all inflections of a word that is
// there, plus 27,000 possessives, and reducing them is the whole job.
//
// The rules are WordNet's own (its "Morphy" component): a table of suffix
// detachments per part of speech, plus an exception list for the forms no
// rule reaches -- went/go, children/child, better/good. The exception list
// ships with WordNet and is generated into exceptions.txt by tools/mkdefs.
//
// Every candidate is checked against the data before it is accepted, so a
// wrong guess costs nothing: "sing" does not become "s" because "s" has no
// entry, and "axes" is tried as both "axe" and "ax".

// detachment is one suffix rule: strip Suffix, append Replace.
type detachment struct{ suffix, replace string }

// Ordered longest-suffix-first within each group so "buses" is tried as
// "bus" before "buse", and "flies" as "fly" before "flie".
var (
	nounRules = []detachment{
		{"ches", "ch"}, {"shes", "sh"}, {"sses", "ss"},
		{"ses", "s"}, {"xes", "x"}, {"zes", "z"},
		{"men", "man"}, {"ies", "y"}, {"es", "e"}, {"es", ""}, {"s", ""},
	}
	verbRules = []detachment{
		{"ies", "y"}, {"ing", "e"}, {"ing", ""},
		{"es", "e"}, {"es", ""}, {"ed", "e"}, {"ed", ""}, {"s", ""},
	}
	adjRules = []detachment{
		{"iest", "y"}, {"ier", "y"},
		{"est", "e"}, {"est", ""}, {"er", "e"}, {"er", ""},
	}
	// Adverbs have no inflections in English; the -ly rule is here because
	// people look up "quickly" and want "quick", which WordNet holds as the
	// adjective. Offered last, so a real adverb entry always wins.
	advRules = []detachment{{"ly", ""}}
)

// candidates returns the forms [word] might reduce to, best first.
//
// Order matters and is not arbitrary: the word itself first, so a real entry
// is never passed over for a guess; then the exception list, which is
// authoritative where it applies; then the suffix rules, whose earlier
// entries are the more specific ones.
//
// This yields possibilities, not answers. The caller tries each against the
// data and takes the first that exists, which is what makes an over-eager
// rule harmless.
func (idx *Index) candidates(word string) []string {
	seen := map[string]bool{word: true}
	out := []string{word}
	add := func(w string) {
		// Two characters is the shortest English word, and a rule that
		// produces something shorter has eaten the word rather than reduced
		// it.
		if len(w) < 2 || seen[w] {
			return
		}
		seen[w] = true
		out = append(out, w)
	}

	// A possessive is not an inflection of the word, it is the word plus a
	// clitic -- and 27,000 of the dictionary's entries are one. Stripped
	// first, and what remains goes through everything else, so "children's"
	// reaches "child".
	base := word
	if strings.HasSuffix(base, "'s") || strings.HasSuffix(base, "’s") {
		base = base[:len(base)-2]
		if strings.HasSuffix(word, "’s") {
			base = strings.TrimSuffix(word, "’s")
		}
		add(base)
	}

	for _, form := range idx.exceptionsFor(base) {
		add(form)
	}

	for _, rules := range [][]detachment{nounRules, verbRules, adjRules, advRules} {
		for _, r := range rules {
			if !strings.HasSuffix(base, r.suffix) {
				continue
			}
			stem := base[:len(base)-len(r.suffix)]
			if stem == "" {
				continue
			}
			add(stem + r.replace)
			// A doubled final consonant before -ing/-ed is spelling, not
			// meaning: "running" is "run", "stopped" is "stop". Only after
			// the plain form, which is right for "billing" and "falling".
			if r.replace == "" && len(stem) > 2 &&
				stem[len(stem)-1] == stem[len(stem)-2] &&
				!isVowel(stem[len(stem)-1]) {
				add(stem[:len(stem)-1])
			}
		}
	}
	return out
}

func isVowel(c byte) bool {
	switch c {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}

// exceptionsFor returns the base forms WordNet records for an irregular
// word, or nothing.
//
// Read from the same sorted-line index the datasets use, so the exception
// list costs one offset per entry rather than a resident map of six thousand
// strings inside the wasm instance.
func (idx *Index) exceptionsFor(word string) []string {
	line, ok := idx.exceptions.find(word)
	if !ok {
		return nil
	}
	_, rest, found := strings.Cut(line, "|")
	if !found {
		return nil
	}
	return strings.Split(rest, ",")
}
