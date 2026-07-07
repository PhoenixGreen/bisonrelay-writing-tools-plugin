package main

import "testing"

func TestParseWordsSkipsBlanksAndComments(t *testing.T) {
	in := "# a comment\nHello\n\n  World  \n# another comment\nfoo\n"
	got := parseWords(in)
	want := []string{"hello", "world", "foo"}
	if len(got) != len(want) {
		t.Fatalf("parseWords(%q) = %v, want %v", in, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("parseWords(%q)[%d] = %q, want %q", in, i, got[i], want[i])
		}
	}
}

func TestParseWordsRealFile(t *testing.T) {
	words := parseWords(wordsTXT)
	if len(words) == 0 {
		t.Fatal("parseWords(wordsTXT) returned no words")
	}
	for _, w := range words {
		if w == "" {
			t.Error("parseWords(wordsTXT) returned a blank word")
		}
	}
}

func TestGrammarRulesNonEmpty(t *testing.T) {
	if len(grammarRules) == 0 {
		t.Fatal("grammarRules is empty")
	}
	for i, r := range grammarRules {
		if r.Pattern == "" {
			t.Errorf("grammarRules[%d] has an empty pattern", i)
		}
	}
}
