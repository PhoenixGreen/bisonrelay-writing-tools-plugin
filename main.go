// Command bisonrelay-spellcheck-plugin is a Bison Relay dynamic-wasm plugin
// providing two capabilities:
//
//   - spellcheck-data: a dictionary and a set of writing-style rules, handed
//     over once when the plugin is enabled. Every match after that happens
//     in the app; there is no per-keystroke work here.
//   - thesaurus: synonyms for one word, answered on demand. This data stays
//     in the plugin -- at 3.5MB it is far too much to push across for a
//     feature used a word at a time.
//
// It is headless -- no nav item, no screens -- so beyond the `alloc` export
// every plugin must provide, it implements only those two capabilities'
// exports.
//
// That also means the plugin never sees a word anyone types. It has no
// network and no filesystem access, and needs neither.
//
// This file is only the WebAssembly plumbing; the content lives in package
// spellcheck. Build with build.sh (GOOS=wasip1 GOARCH=wasm go build
// -buildmode=c-shared); see wasmhost's package doc in the bisonrelay repo
// for the ABI it implements.
package main

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"io"

	"unsafe"

	"github.com/PhoenixGreen/bisonrelay-spellcheck-plugin/spellcheck"
	"github.com/PhoenixGreen/bisonrelay-spellcheck-plugin/thesaurus"
)

// The two datasets are embedded compressed. Together they are 4.6MB of text,
// which is most of the built module; gzipped they are 1.5MB, and the module
// drops from 7.6MB to well under half that.
//
// This costs nothing at runtime that was not already being paid: //go:embed
// puts the data in the module's linear memory either way, so the only
// difference is one decompression the first time each is used.
//
//go:embed words.txt.gz
var wordsGZ []byte

//go:embed thesaurus.txt.gz
var thesaurusGZ []byte

// gunzip decompresses an embedded dataset. A failure here means the module
// was built wrong rather than that anything went wrong at runtime, so the
// caller degrades to having no data rather than trying to recover.
func gunzip(b []byte) string {
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return ""
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		return ""
	}
	return string(out)
}

// thesaurusIndex is built on first use rather than at startup: a plugin
// whose thesaurus is never consulted should not pay to scan 3.5MB, and the
// host loads this module during client startup.
var thesaurusIndex *thesaurus.Index

func lookupIndex() *thesaurus.Index {
	if thesaurusIndex == nil {
		thesaurusIndex = thesaurus.NewIndex(gunzip(thesaurusGZ))
	}
	return thesaurusIndex
}

// --- ABI plumbing every dynamic-wasm plugin needs ---

// pinned keeps allocated buffers alive: nothing else references them once
// alloc returns, so without this the Go GC could collect the memory the host
// is about to write to (or has just written to) before it is read.
var pinned = map[int32][]byte{}

//go:wasmexport alloc
func alloc(size int32) int32 {
	if size == 0 {
		size = 1
	}
	buf := make([]byte, size)
	// go vet flags this uintptr(unsafe.Pointer(...)) conversion as a
	// possible misuse since the address outlives the expression it is taken
	// in -- but Go's allocator doesn't move objects, and buf is kept alive
	// via pinned (keyed by this same address) for as long as the host needs
	// it, so there is no actual staleness risk here.
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
	b, err := json.Marshal(spellcheck.Data{
		Words:        spellcheck.ParseWords(gunzip(wordsGZ)),
		GrammarRules: spellcheck.Rules,
	})
	if err != nil {
		// Returning zero bytes is how this ABI says "I could not answer";
		// the host logs it and carries on without this provider's data.
		return 0
	}
	return writeResult(b)
}

//go:wasmexport lookup_synonyms
func lookupSynonyms(word string) uint64 {
	entry, ok := lookupIndex().Lookup(word)
	if !ok {
		// Zero bytes is this ABI's "I could not answer", which is the honest
		// result for a name, a typo, or a word the data doesn't cover.
		return 0
	}
	b, err := json.Marshal(entry)
	if err != nil {
		return 0
	}
	return writeResult(b)
}

func main() {}
