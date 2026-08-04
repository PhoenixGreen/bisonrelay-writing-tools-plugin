// Command mkwords generates the plugin's wordlist: it expands the
// SCOWL-derived hunspell dictionary in data/ from roots plus affix flags
// into the flat list of surface forms the checker actually compares against,
// then merges data/extra-words.txt on top.
//
// The expansion is why this tool exists. A hunspell .dic stores "abandon/
// LDGS", not the five words that spells; a checker fed the roots alone
// flags every inflection a user actually types ("running", "walked") as
// misspelled. Shipping surface forms keeps the plugin's runtime trivial --
// a set membership test -- at the cost of a larger, generated file.
//
// Run from the repo root:
//
//	go run ./tools/mkwords > words.txt
//
// words.txt is committed, so building the plugin needs neither this tool nor
// the data/ sources; both are kept so the list is reproducible and auditable
// rather than an opaque blob.
package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

const (
	dicPath   = "data/index.dic"
	affPath   = "data/index.aff"
	extraPath = "data/extra-words.txt"
)

// rule is one affix transformation: strip these characters off the end (or
// front), add these, and only where the word matches cond.
type rule struct {
	strip string
	add   string
	cond  *regexp.Regexp
}

// affix is one flag's worth of rules. cross records whether hunspell allows
// this affix to combine with one from the other side of the word.
type affix struct {
	prefix bool
	cross  bool
	rules  []rule
}

func parseAff(path string) (map[rune]*affix, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	affixes := map[rune]*affix{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 4 || (fields[0] != "SFX" && fields[0] != "PFX") {
			continue
		}
		flag := []rune(fields[1])[0]
		isPfx := fields[0] == "PFX"

		// A flag's block opens with "SFX <flag> <cross> <count>", where
		// cross is Y or N; every later line with the same flag is a rule.
		if fields[2] == "Y" || fields[2] == "N" {
			affixes[flag] = &affix{prefix: isPfx, cross: fields[2] == "Y"}
			continue
		}
		a := affixes[flag]
		if a == nil {
			continue
		}

		strip, add, cond := fields[2], fields[3], "."
		if len(fields) >= 5 {
			cond = fields[4]
		}
		if strip == "0" {
			strip = ""
		}
		if add == "0" {
			add = ""
		}
		// The condition is anchored to whichever end the affix attaches to.
		pattern := cond + "$"
		if isPfx {
			pattern = "^" + cond
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("flag %c: condition %q: %w", flag, cond, err)
		}
		a.rules = append(a.rules, rule{strip: strip, add: add, cond: re})
	}
	return affixes, sc.Err()
}

// apply returns every form a's rules produce from word.
func apply(word string, a *affix) []string {
	var out []string
	for _, r := range a.rules {
		if !r.cond.MatchString(word) {
			continue
		}
		if a.prefix {
			if !strings.HasPrefix(word, r.strip) {
				continue
			}
			out = append(out, r.add+word[len(r.strip):])
		} else {
			if !strings.HasSuffix(word, r.strip) {
				continue
			}
			out = append(out, word[:len(word)-len(r.strip)]+r.add)
		}
	}
	return out
}

type wordSet map[string]bool

// add normalizes and keeps a word, dropping anything the checker could never
// match anyway: empty lines, and entries carrying digits (the dictionary has
// a few, e.g. "0th", which only exist to spell ordinals).
func (s wordSet) add(w string) {
	w = strings.ToLower(strings.TrimSpace(w))
	if w == "" || strings.ContainsAny(w, "0123456789") {
		return
	}
	s[w] = true
}

func expandDic(path string, affixes map[rune]*affix, words wordSet) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	sc.Scan() // the first line is the entry count, not a word

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		word, flags := line, ""
		if i := strings.Index(line, "/"); i >= 0 {
			word, flags = line[:i], line[i+1:]
		}
		words.add(word)

		var suffixed []string
		for _, fl := range flags {
			a := affixes[fl]
			if a == nil {
				continue
			}
			forms := apply(word, a)
			for _, w := range forms {
				words.add(w)
			}
			if !a.prefix {
				suffixed = append(suffixed, forms...)
			}
		}

		// Cross-product: "unthinkable" needs a prefix applied to an
		// already-suffixed form, which hunspell permits only when both
		// sides declare it.
		for _, fl := range flags {
			a := affixes[fl]
			if a == nil || !a.prefix || !a.cross {
				continue
			}
			for _, s := range suffixed {
				for _, w := range apply(s, a) {
					words.add(w)
				}
			}
		}
	}
	return sc.Err()
}

func readExtra(path string, words wordSet) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		words.add(line)
	}
	return sc.Err()
}

func main() {
	affixes, err := parseAff(affPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mkwords:", affPath+":", err)
		os.Exit(1)
	}

	words := wordSet{}
	if err := expandDic(dicPath, affixes, words); err != nil {
		fmt.Fprintln(os.Stderr, "mkwords:", dicPath+":", err)
		os.Exit(1)
	}
	fromDic := len(words)
	if err := readExtra(extraPath, words); err != nil {
		fmt.Fprintln(os.Stderr, "mkwords:", extraPath+":", err)
		os.Exit(1)
	}

	out := make([]string, 0, len(words))
	for w := range words {
		out = append(out, w)
	}
	sort.Strings(out)

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	fmt.Fprintln(w, "# Generated by tools/mkwords -- do not edit.")
	fmt.Fprintf(w, "# %d words: %d expanded from %s, %d added from %s.\n",
		len(out), fromDic, dicPath, len(words)-fromDic, extraPath)
	fmt.Fprintln(w, "# Dictionary licence and attribution: data/LICENSE-SCOWL.")
	for _, word := range out {
		fmt.Fprintln(w, word)
	}
	fmt.Fprintln(os.Stderr, "mkwords:", len(out), "words")
}
