// Command bisonrelay-spellcheck-plugin is a Bison Relay dynamic-wasm
// plugin: it supplies a wordlist + a handful of regex-based writing-style
// rules that Bison Relay's own spellcheck UI (SpellCheckModel, unchanged by
// this plugin) applies to chat/post/comment text as the user types. It's a
// headless plugin -- no nav item, no screens, just the CapabilitySpellcheck
// Data capability -- so it only implements the required `alloc` export plus
// the single optional `get_spellcheck_data` export; there is no fetching or
// per-keystroke work here, the data is static and loaded once. Build with
// build.sh (GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared); see
// wasmhost's package doc in the bisonrelay repo for the ABI this
// implements.
package main

import (
	_ "embed"
	"encoding/json"
	"strings"
	"unsafe"
)

//go:embed words.txt
var wordsTXT string

// grammarRule/spellcheckData mirror wasmhost.GrammarRule/SpellcheckData's
// JSON schema exactly -- this plugin keeps its own copy since it's a
// separate Go module with no dependency on bisonrelay.
type grammarRule struct {
	Pattern string `json:"pattern"`
	Message string `json:"message"`
	Suggest string `json:"suggest"`
}

type spellcheckData struct {
	Words        []string      `json:"words"`
	GrammarRules []grammarRule `json:"grammarRules"`
}

// grammarRules are plain data -- this plugin never compiles or executes
// them itself; Bison Relay's Dart UI does, since Dart's regex engine
// (unlike Go's RE2) supports the backreferences "repeated word" needs.
var grammarRules = []grammarRule{
	{Pattern: `\b(\w+)([ \t]+)\1\b`, Message: "Repeated word", Suggest: "$1"},
	{Pattern: `[ ]{2,}`, Message: "Multiple spaces", Suggest: " "},
	{Pattern: `[ \t]+([,.!?;:])`, Message: "Space before punctuation", Suggest: "$1"},
	{Pattern: `([.!?])([A-Z])`, Message: "Missing space after punctuation", Suggest: "$1 $2"},
}

// parseWords tokenizes words.txt: one lowercase word per line, blank lines
// and "#"-prefixed comment lines ignored.
func parseWords(txt string) []string {
	var words []string
	for _, line := range strings.Split(txt, "\n") {
		line = strings.ToLower(strings.TrimSpace(line))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		words = append(words, line)
	}
	return words
}

// --- ABI plumbing every dynamic-wasm plugin needs ---

// pinned keeps allocated buffers alive: nothing else references them once
// alloc returns, so without this the Go GC could collect the memory the
// host is about to write to (or has just written to) before it's read.
var pinned = map[int32][]byte{}

//go:wasmexport alloc
func alloc(size int32) int32 {
	if size == 0 {
		size = 1
	}
	buf := make([]byte, size)
	// go vet flags this uintptr(unsafe.Pointer(...)) conversion as a
	// possible misuse since the address outlives the expression it's
	// taken in -- but Go's allocator doesn't move objects, and buf is
	// kept alive via pinned (keyed by this same address) for as long as
	// the host needs it, so there's no actual staleness risk here.
	ptr := int32(uintptr(unsafe.Pointer(&buf[0])))
	pinned[ptr] = buf
	return ptr
}

func writeResult(data []byte) uint64 {
	ptr := alloc(int32(len(data)))
	copy(pinned[ptr], data)
	return (uint64(uint32(ptr)) << 32) | uint64(len(data))
}

//go:wasmexport get_spellcheck_data
func getSpellcheckData() uint64 {
	data := spellcheckData{
		Words:        parseWords(wordsTXT),
		GrammarRules: grammarRules,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return writeResult([]byte(`{"words":[],"grammarRules":[]}`))
	}
	return writeResult(b)
}

func main() {}
